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
}

// NewHandler creates a new Handler
func NewHandler(db *bsql.DB, uploadRepo Repository) *Handler {
	return &Handler{
		db:         db,
		uploadRepo: uploadRepo,
	}
}

// Upload handles image file upload - POST /upload
func (h *Handler) Upload(c echo.Context) error {
	return h.uploadImageCore(c, "", 0)
}

// UploadWithLimit handles image file upload with custom max size
func (h *Handler) UploadWithLimit(c echo.Context, category string, maxFileSize int64) error {
	return h.uploadImageCore(c, category, maxFileSize)
}

func (h *Handler) uploadImageCore(c echo.Context, category string, maxFileSize int64) error {
	// Get user claims from context (set by JWT middleware)
	claims := c.Get("user").(*auth.TokenClaims)

	// Default max file size is 8MB
	if maxFileSize == 0 {
		maxFileSize = MaxFileSize
	}

	name := "image"
	fileType := "image"

	fileHeader, err := c.FormFile(name)
	validateErr := validateUploadFile(fileHeader, fileType, err)
	if validateErr != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"success": false,
			"error":   validateErr.Error(),
		})
	}

	// Check file size
	if fileHeader.Size > maxFileSize {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"success": false,
			"error":   ErrFileTooLarge.Error(),
			"data": echo.Map{
				"maxSize":    maxFileSize,
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
	tail := tokens[len(tokens)-1]

	fileTypeFolder := fmt.Sprintf("%ss", fileType) // "images"

	// Collect HTTP metadata
	req := c.Request()
	clientIP := c.RealIP()
	userAgent := req.UserAgent()
	requestHost := req.Host
	requestURI := req.RequestURI

	// Save file and get paths (use project's tmp folder)
	relativePath, absolutePath, err := h.saveMediaFile(claims.UserID, file, "tmp", fileTypeFolder, tail)
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
	// Get user claims from context (set by JWT middleware)
	claims := c.Get("user").(*auth.TokenClaims)

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

// UploadForm serves a simple HTML form for file upload testing
func (h *Handler) UploadForm(c echo.Context) error {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>File Upload</title>
</head>
<body>
    <h1>Upload Image</h1>
    <form action="/upload" method="POST" enctype="multipart/form-data">
        <p>
            <label>Authorization Token:</label><br>
            <input type="text" id="token" name="token" placeholder="Enter your JWT token" style="width: 400px;">
        </p>
        <p>
            <label>Select Image:</label><br>
            <input type="file" name="image" accept="image/*">
        </p>
        <p>
            <button type="submit">Upload</button>
        </p>
    </form>
    <p><small>Note: Max file size is 8MB. Only image files are accepted.</small></p>
    
    <script>
    document.querySelector('form').addEventListener('submit', function(e) {
        e.preventDefault();
        
        var token = document.getElementById('token').value;
        var formData = new FormData();
        var fileInput = document.querySelector('input[type="file"]');
        formData.append('image', fileInput.files[0]);
        
        fetch('/upload', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + token
            },
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            alert(JSON.stringify(data, null, 2));
        })
        .catch(error => {
            alert('Error: ' + error);
        });
    });
    </script>
</body>
</html>`
	return c.HTML(http.StatusOK, html)
}
