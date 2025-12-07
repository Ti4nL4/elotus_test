package upload

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"elotus_test/server/bredis"
	"elotus_test/server/bsql"
	"elotus_test/server/cmd"
	"elotus_test/server/models/auth"
	"elotus_test/server/response"

	"github.com/labstack/echo/v4"
)

var timeNow = time.Now

type Handler struct {
	db         *bsql.DB
	uploadRepo Repository
	redis      *bredis.Client
}

func NewHandler(db *bsql.DB, uploadRepo Repository, redis *bredis.Client) *Handler {
	return &Handler{
		db:         db,
		uploadRepo: uploadRepo,
		redis:      redis,
	}
}

func (h *Handler) cacheKey(userID int64) string {
	return fmt.Sprintf("uploads:%d", userID)
}

func (h *Handler) Upload(c echo.Context) error {
	claims := c.Get("user").(*auth.TokenClaims)

	fileHeader, err := c.FormFile("data")
	validateErr := validateUploadFile(fileHeader, "data", err)
	if validateErr != nil {
		return response.ValidationError(c, validateErr.Error())
	}

	if fileHeader.Size > MaxFileSize {
		return response.BadRequest(c, fmt.Sprintf("%s (max: %d bytes, actual: %d bytes)",
			ErrFileTooLarge.Error(), MaxFileSize, fileHeader.Size))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return response.InternalError(c, "Failed to open uploaded file")
	}
	defer file.Close()

	// Extract extension from filename, fallback to content-type detection
	ext := strings.TrimPrefix(filepath.Ext(fileHeader.Filename), ".")
	if ext == "" {
		contentType := detectContentType(fileHeader)
		switch contentType {
		case "image/jpeg":
			ext = "jpg"
		case "image/png":
			ext = "png"
		case "image/gif":
			ext = "gif"
		case "image/webp":
			ext = "webp"
		case "image/bmp":
			ext = "bmp"
		default:
			ext = "bin"
		}
	}

	req := c.Request()
	clientIP := c.RealIP()
	userAgent := req.UserAgent()
	requestHost := req.Host
	requestURI := req.RequestURI

	relativePath, absolutePath, err := h.saveMediaFile(claims.UserID, file, "tmp", "images", ext)
	if err != nil {
		log.Printf("[Upload] saveMediaFile error: %v", err)
		return response.InternalError(c, "Failed to save file: "+err.Error())
	}

	uploadRecord := &FileUpload{
		UserID:           claims.UserID,
		Filename:         filepath.Base(absolutePath),
		OriginalFilename: fileHeader.Filename,
		ContentType:      detectContentType(fileHeader),
		FileSize:         fileHeader.Size,
		TempPath:         absolutePath,
		ClientIP:         clientIP,
		UserAgent:        userAgent,
		RequestHost:      requestHost,
		RequestURI:       requestURI,
	}

	savedUpload, err := h.uploadRepo.CreateFileUpload(uploadRecord)
	if err != nil {
		os.Remove(absolutePath)
		return response.InternalError(c, "Failed to save file metadata")
	}

	if h.redis != nil {
		_ = h.redis.Delete(h.cacheKey(claims.UserID))
	}

	return response.Success(c, echo.Map{
		"file_id":           savedUpload.ID,
		"filename":          savedUpload.Filename,
		"original_filename": savedUpload.OriginalFilename,
		"content_type":      savedUpload.ContentType,
		"file_size":         savedUpload.FileSize,
		"temp_path":         absolutePath,
		"relative_url":      relativePath,
		"uploaded_at":       savedUpload.CreatedAt,
	})
}

func (h *Handler) saveMediaFile(userID int64, file multipart.File, savePath, fileTypeFolder, tail string) (relativeUrl, absoluteUrl string, err error) {
	fileName := fmt.Sprintf("%d_%s.%s", userID, randSeq(20), tail)
	baseFolder := cmd.ResolvePath(savePath)
	fileFolder := filepath.Join(baseFolder, fileTypeFolder)

	log.Printf("[Upload] Base folder: %s", baseFolder)
	log.Printf("[Upload] File folder: %s", fileFolder)

	if err := os.MkdirAll(fileFolder, 0755); err != nil {
		log.Printf("[Upload] Error creating folder: %v", err)
		return "", "", fmt.Errorf("failed to create folder: %w", err)
	}

	filePath := filepath.Join(fileFolder, fileName)
	log.Printf("[Upload] File path: %s", filePath)

	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("[Upload] Error creating file: %v", err)
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, file)
	if err != nil {
		log.Printf("[Upload] Error writing file: %v", err)
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}
	log.Printf("[Upload] Written %d bytes to %s", written, filePath)

	relativeUrl = fmt.Sprintf("/media/%s/%s", strings.Trim(fileTypeFolder, "/"), fileName)
	absoluteUrl = filePath

	return relativeUrl, absoluteUrl, nil
}

func validateUploadFile(fileHeader *multipart.FileHeader, fileType string, err error) error {
	if err != nil {
		if err == http.ErrMissingFile {
			return ErrNoFileUploaded
		}
		return errors.New("failed to get uploaded file")
	}

	if fileHeader == nil || fileHeader.Size == 0 {
		return ErrNoFileUploaded
	}

	file, err := fileHeader.Open()
	if err != nil {
		return errors.New("failed to open file")
	}
	defer file.Close()

	// Read first 512 bytes to detect content type via magic bytes
	buff := make([]byte, 512)
	n, err := file.Read(buff)
	if err != nil && err != io.EOF {
		return errors.New("failed to read file content")
	}

	if fileType == "data" || fileType == "image" {
		contentType := http.DetectContentType(buff[:n])
		if !strings.HasPrefix(contentType, "image/") {
			return ErrInvalidContentType
		}
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return errors.New("failed to process file")
	}

	return nil
}

func detectContentType(fileHeader *multipart.FileHeader) string {
	file, err := fileHeader.Open()
	if err != nil {
		return "application/octet-stream"
	}
	defer file.Close()

	buff := make([]byte, 512)
	n, err := file.Read(buff)
	if err != nil && err != io.EOF {
		return "application/octet-stream"
	}

	return http.DetectContentType(buff[:n])
}

// randSeq generates a random string using Linear Congruential Generator
func randSeq(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	seed := uint64(timeNow().UnixNano())
	for i := range b {
		seed = seed*1103515245 + 12345
		b[i] = letters[seed%uint64(len(letters))]
	}
	return string(b)
}

func (h *Handler) GetUserUploads(c echo.Context) error {
	claims := c.Get("user").(*auth.TokenClaims)
	cacheKey := h.cacheKey(claims.UserID)

	if h.redis != nil {
		var cached []echo.Map
		if h.redis.Get(cacheKey, &cached) == nil {
			return response.SuccessWithMeta(c, cached, &response.Meta{
				Total:  len(cached),
				Cached: true,
			})
		}
	}

	uploads, err := h.uploadRepo.GetFileUploadsByUserID(claims.UserID)
	if err != nil {
		return response.InternalError(c, "Failed to get uploads")
	}

	uploadList := make([]echo.Map, 0, len(uploads))
	for _, upload := range uploads {
		uploadList = append(uploadList, echo.Map{
			"id":                upload.ID,
			"filename":          upload.Filename,
			"original_filename": upload.OriginalFilename,
			"content_type":      upload.ContentType,
			"file_size":         upload.FileSize,
			"file_path":         upload.TempPath,
			"created_at":        upload.CreatedAt,
		})
	}

	if h.redis != nil {
		_ = h.redis.Set(cacheKey, uploadList, 30*time.Minute)
	}

	return response.SuccessWithMeta(c, uploadList, &response.Meta{
		Total: len(uploadList),
	})
}

func (h *Handler) GetUploadByID(c echo.Context) error {
	claims := c.Get("user").(*auth.TokenClaims)

	idStr := c.Param("id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	upload, found := h.uploadRepo.GetFileUploadByID(id)
	if !found {
		return response.NotFound(c, "Upload not found")
	}

	if upload.UserID != claims.UserID {
		return response.Forbidden(c, "Access denied")
	}

	return response.Success(c, echo.Map{
		"id":                upload.ID,
		"filename":          upload.Filename,
		"original_filename": upload.OriginalFilename,
		"content_type":      upload.ContentType,
		"file_size":         upload.FileSize,
		"file_path":         upload.TempPath,
		"created_at":        upload.CreatedAt,
	})
}
