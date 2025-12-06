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

	"github.com/labstack/echo/v4"
)

// timeNow is a helper for getting current time (can be mocked for testing)
var timeNow = time.Now

// Handler handles file upload requests
type Handler struct {
	db         *bsql.DB
	uploadRepo Repository
	redis      *bredis.Client
}

// NewHandler creates a new Handler
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

// Upload handles image file upload - POST /upload
func (h *Handler) Upload(c echo.Context) error {
	// Get user claims from context (set by JWT middleware)
	claims := c.Get("user").(*auth.TokenClaims)

	fileHeader, err := c.FormFile("image")
	validateErr := validateUploadFile(fileHeader, "image", err)
	if validateErr != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"success": false,
			"error":   validateErr.Error(),
		})
	}

	// Check file size (max 8MB)
	if fileHeader.Size > MaxFileSize {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"success": false,
			"error":   ErrFileTooLarge.Error(),
			"data": echo.Map{
				"maxSize":    MaxFileSize,
				"actualSize": fileHeader.Size,
			},
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"success": false,
			"error":   "Failed to open uploaded file",
		})
	}
	defer file.Close()

	// Get file extension from original filename
	tokens := strings.Split(fileHeader.Filename, ".")
	ext := tokens[len(tokens)-1]

	// Collect HTTP metadata
	req := c.Request()
	clientIP := c.RealIP()
	userAgent := req.UserAgent()
	requestHost := req.Host
	requestURI := req.RequestURI

	// Save file and get paths (use project's tmp folder)
	relativePath, absolutePath, err := h.saveMediaFile(claims.UserID, file, "tmp", "images", ext)
	if err != nil {
		log.Printf("[Upload] saveMediaFile error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"success": false,
			"error":   "Failed to save file: " + err.Error(),
		})
	}

	// Create file upload record for database
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

	// Save metadata to database
	savedUpload, err := h.uploadRepo.CreateFileUpload(uploadRecord)
	if err != nil {
		// Clean up file on database error
		os.Remove(absolutePath)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"success": false,
			"error":   "Failed to save file metadata",
		})
	}

	// Invalidate uploads cache for this user
	if h.redis != nil {
		_ = h.redis.Delete(h.cacheKey(claims.UserID))
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success":      true,
		"absolute_url": absolutePath,
		"relative_url": relativePath,
		"data": echo.Map{
			"file_id":           savedUpload.ID,
			"filename":          savedUpload.Filename,
			"original_filename": savedUpload.OriginalFilename,
			"content_type":      savedUpload.ContentType,
			"file_size":         savedUpload.FileSize,
			"uploaded_at":       savedUpload.CreatedAt,
		},
	})
}

// saveMediaFile saves the uploaded file to the specified path
func (h *Handler) saveMediaFile(userID int64, file multipart.File, savePath, fileTypeFolder, tail string) (relativeUrl, absoluteUrl string, err error) {
	// Generate unique filename
	fileName := fmt.Sprintf("%d_%s.%s", userID, randSeq(20), tail)

	// Resolve to absolute path within project
	baseFolder := cmd.ResolvePath(savePath)
	fileFolder := filepath.Join(baseFolder, fileTypeFolder)

	log.Printf("[Upload] Base folder: %s", baseFolder)
	log.Printf("[Upload] File folder: %s", fileFolder)

	// Create folder if not exists
	if err := os.MkdirAll(fileFolder, 0755); err != nil {
		log.Printf("[Upload] Error creating folder: %v", err)
		return "", "", fmt.Errorf("failed to create folder: %w", err)
	}

	filePath := filepath.Join(fileFolder, fileName)
	log.Printf("[Upload] File path: %s", filePath)

	// Create destination file
	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("[Upload] Error creating file: %v", err)
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy file content
	written, err := io.Copy(out, file)
	if err != nil {
		log.Printf("[Upload] Error writing file: %v", err)
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}
	log.Printf("[Upload] Written %d bytes to %s", written, filePath)

	// Relative URL for web access (e.g., /media/images/filename.jpg)
	relativeUrl = fmt.Sprintf("/media/%s/%s", strings.Trim(fileTypeFolder, "/"), fileName)
	absoluteUrl = filePath

	return relativeUrl, absoluteUrl, nil
}

// validateUploadFile validates the uploaded file
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

	// Read first 512 bytes to detect content type
	buff := make([]byte, 512)
	n, err := file.Read(buff)
	if err != nil && err != io.EOF {
		return errors.New("failed to read file content")
	}

	if fileType == "image" {
		contentType := http.DetectContentType(buff[:n])
		if !strings.Contains(contentType, fileType) {
			return ErrInvalidContentType
		}
	}

	// Seek back to beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return errors.New("failed to process file")
	}

	return nil
}

// detectContentType detects the content type of the uploaded file
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

// randSeq generates a random string of specified length using time-based randomness
func randSeq(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	seed := uint64(timeNow().UnixNano())
	for i := range b {
		seed = seed*1103515245 + 12345 // Linear congruential generator
		b[i] = letters[seed%uint64(len(letters))]
	}
	return string(b)
}

// GetUserUploads returns all uploads for the authenticated user
func (h *Handler) GetUserUploads(c echo.Context) error {
	claims := c.Get("user").(*auth.TokenClaims)
	cacheKey := h.cacheKey(claims.UserID)

	// Try cache first
	if h.redis != nil {
		var cached []echo.Map
		if h.redis.Get(cacheKey, &cached) == nil {
			return c.JSON(http.StatusOK, echo.Map{
				"success": true,
				"data":    cached,
				"total":   len(cached),
				"cached":  true,
			})
		}
	}

	// Cache miss - query DB
	uploads, err := h.uploadRepo.GetFileUploadsByUserID(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"success": false,
			"error":   "Failed to get uploads",
		})
	}

	// Convert to response format
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

	// Cache for 30 seconds
	if h.redis != nil {
		_ = h.redis.Set(cacheKey, uploadList, 30*time.Minute)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"data":    uploadList,
		"total":   len(uploadList),
	})
}

// GetUploadByID returns a specific upload by ID
func (h *Handler) GetUploadByID(c echo.Context) error {
	// Get user claims from context
	claims := c.Get("user").(*auth.TokenClaims)

	// Get ID from path parameter
	idStr := c.Param("id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	upload, found := h.uploadRepo.GetFileUploadByID(id)
	if !found {
		return c.JSON(http.StatusNotFound, echo.Map{
			"success": false,
			"error":   "Upload not found",
		})
	}

	// Check if upload belongs to user
	if upload.UserID != claims.UserID {
		return c.JSON(http.StatusForbidden, echo.Map{
			"success": false,
			"error":   "Access denied",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"data": echo.Map{
			"id":                upload.ID,
			"filename":          upload.Filename,
			"original_filename": upload.OriginalFilename,
			"content_type":      upload.ContentType,
			"file_size":         upload.FileSize,
			"file_path":         upload.TempPath,
			"created_at":        upload.CreatedAt,
		},
	})
}
