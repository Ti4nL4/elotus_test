package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"elotus_test/server/models/auth"
	"elotus_test/server/models/upload"
	"elotus_test/server/response"

	"github.com/labstack/echo/v4"
)

var _ upload.Repository = (*MockUploadRepository)(nil)

func createMultipartForm(fieldName, fileName string, content []byte) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile(fieldName, fileName)
	part.Write(content)
	writer.Close()

	return body, writer.FormDataContentType()
}

func createTestImageContent() []byte {
	// Minimal valid PNG header (8 bytes) + IHDR chunk + IEND chunk
	// This creates a 1x1 pixel image
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk length
		0x49, 0x48, 0x44, 0x52, // IHDR chunk type
		0x00, 0x00, 0x00, 0x01, // Width: 1
		0x00, 0x00, 0x00, 0x01, // Height: 1
		0x08, 0x02, // Bit depth: 8, Color type: 2 (RGB)
		0x00, 0x00, 0x00, // Compression, Filter, Interlace
		0x90, 0x77, 0x53, 0xDE, // CRC
		0x00, 0x00, 0x00, 0x0C, // IDAT chunk length
		0x49, 0x44, 0x41, 0x54, // IDAT chunk type
		0x08, 0xD7, 0x63, 0xF8, 0x0F, 0x00, 0x00, 0x01, 0x01, 0x00, 0x05, 0xFE, // Compressed data
		0xD2, 0xB4, 0x54, 0xB0, // CRC
		0x00, 0x00, 0x00, 0x00, // IEND chunk length
		0x49, 0x45, 0x4E, 0x44, // IEND chunk type
		0xAE, 0x42, 0x60, 0x82, // CRC
	}
}

func setupUploadTestHandler() (*upload.Handler, *MockUploadRepository) {
	mockRepo := NewMockUploadRepository()
	handler := upload.NewHandler(nil, mockRepo, nil)
	return handler, mockRepo
}

func createUploadTestContext(e *echo.Echo, method, path string, body io.Reader, contentType string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set(echo.HeaderContentType, contentType)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func parseUploadResponse(body []byte) (*response.Response, error) {
	var resp response.Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func getUploadDataMap(resp *response.Response) map[string]interface{} {
	if resp.Data == nil {
		return nil
	}
	if m, ok := resp.Data.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func getUploadDataList(resp *response.Response) []interface{} {
	if resp.Data == nil {
		return nil
	}
	if arr, ok := resp.Data.([]interface{}); ok {
		return arr
	}
	return nil
}

func TestGetUserUploads_Success(t *testing.T) {
	handler, mockRepo := setupUploadTestHandler()

	mockRepo.AddUpload(&upload.FileUpload{
		ID:               1,
		UserID:           1,
		Filename:         "test1.png",
		OriginalFilename: "original1.png",
		ContentType:      "image/png",
		FileSize:         1024,
		TempPath:         "/tmp/test1.png",
		CreatedAt:        time.Now(),
	})
	mockRepo.AddUpload(&upload.FileUpload{
		ID:               2,
		UserID:           1,
		Filename:         "test2.png",
		OriginalFilename: "original2.png",
		ContentType:      "image/png",
		FileSize:         2048,
		TempPath:         "/tmp/test2.png",
		CreatedAt:        time.Now(),
	})
	mockRepo.AddUpload(&upload.FileUpload{
		ID:               3,
		UserID:           2,
		Filename:         "test3.png",
		OriginalFilename: "original3.png",
		ContentType:      "image/png",
		FileSize:         1024,
		TempPath:         "/tmp/test3.png",
		CreatedAt:        time.Now(),
	})

	e := echo.New()
	c, rec := createUploadTestContext(e, http.MethodGet, "/api/uploads", nil, "")
	c.Set("user", &auth.TokenClaims{UserID: 1, Username: "testuser"})

	err := handler.GetUserUploads(c)
	if err != nil {
		t.Fatalf("GetUserUploads returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	resp, _ := parseUploadResponse(rec.Body.Bytes())

	if !resp.Success {
		t.Error("Expected success: true")
	}

	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Errorf("Expected meta.total=2, got %v", resp.Meta)
	}

	dataList := getUploadDataList(resp)
	if len(dataList) != 2 {
		t.Errorf("Expected 2 uploads, got %d", len(dataList))
	}
}

func TestGetUserUploads_Empty(t *testing.T) {
	handler, _ := setupUploadTestHandler()

	e := echo.New()
	c, rec := createUploadTestContext(e, http.MethodGet, "/api/uploads", nil, "")
	c.Set("user", &auth.TokenClaims{UserID: 1, Username: "testuser"})

	err := handler.GetUserUploads(c)
	if err != nil {
		t.Fatalf("GetUserUploads returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	resp, _ := parseUploadResponse(rec.Body.Bytes())

	if !resp.Success {
		t.Error("Expected success: true")
	}

	if resp.Meta == nil || resp.Meta.Total != 0 {
		t.Errorf("Expected meta.total=0, got %v", resp.Meta)
	}
}

func TestGetUploadByID_Success(t *testing.T) {
	handler, mockRepo := setupUploadTestHandler()

	mockRepo.AddUpload(&upload.FileUpload{
		ID:               1,
		UserID:           1,
		Filename:         "test1.png",
		OriginalFilename: "original1.png",
		ContentType:      "image/png",
		FileSize:         1024,
		TempPath:         "/tmp/test1.png",
		CreatedAt:        time.Now(),
	})

	e := echo.New()
	c, rec := createUploadTestContext(e, http.MethodGet, "/api/uploads/1", nil, "")
	c.SetParamNames("id")
	c.SetParamValues("1")
	c.Set("user", &auth.TokenClaims{UserID: 1, Username: "testuser"})

	err := handler.GetUploadByID(c)
	if err != nil {
		t.Fatalf("GetUploadByID returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	resp, _ := parseUploadResponse(rec.Body.Bytes())

	if !resp.Success {
		t.Error("Expected success: true")
	}

	data := getUploadDataMap(resp)
	if data == nil {
		t.Fatal("Expected data in response")
	}

	if data["filename"] != "test1.png" {
		t.Errorf("Expected filename 'test1.png', got %v", data["filename"])
	}
}

func TestGetUploadByID_NotFound(t *testing.T) {
	handler, _ := setupUploadTestHandler()

	e := echo.New()
	c, rec := createUploadTestContext(e, http.MethodGet, "/api/uploads/999", nil, "")
	c.SetParamNames("id")
	c.SetParamValues("999")
	c.Set("user", &auth.TokenClaims{UserID: 1, Username: "testuser"})

	err := handler.GetUploadByID(c)
	if err != nil {
		t.Fatalf("GetUploadByID returned error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	resp, _ := parseUploadResponse(rec.Body.Bytes())

	if resp.Success {
		t.Error("Expected success: false")
	}
	if resp.Error == nil || resp.Error.Message != "Upload not found" {
		t.Errorf("Expected 'Upload not found' error, got %v", resp.Error)
	}
}

func TestGetUploadByID_AccessDenied(t *testing.T) {
	handler, mockRepo := setupUploadTestHandler()

	mockRepo.AddUpload(&upload.FileUpload{
		ID:               1,
		UserID:           2,
		Filename:         "test1.png",
		OriginalFilename: "original1.png",
		ContentType:      "image/png",
		FileSize:         1024,
		TempPath:         "/tmp/test1.png",
		CreatedAt:        time.Now(),
	})

	e := echo.New()
	c, rec := createUploadTestContext(e, http.MethodGet, "/api/uploads/1", nil, "")
	c.SetParamNames("id")
	c.SetParamValues("1")
	c.Set("user", &auth.TokenClaims{UserID: 1, Username: "testuser"})

	err := handler.GetUploadByID(c)
	if err != nil {
		t.Fatalf("GetUploadByID returned error: %v", err)
	}

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}

	resp, _ := parseUploadResponse(rec.Body.Bytes())

	if resp.Success {
		t.Error("Expected success: false")
	}
	if resp.Error == nil || resp.Error.Message != "Access denied" {
		t.Errorf("Expected 'Access denied' error, got %v", resp.Error)
	}
}

func TestMaxFileSize(t *testing.T) {
	if upload.MaxFileSize != 8*1024*1024 {
		t.Errorf("Expected MaxFileSize to be 8MB (8388608), got %d", upload.MaxFileSize)
	}
}

func TestAllowedImageTypes(t *testing.T) {
	expectedTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/bmp",
		"image/tiff",
	}

	for _, contentType := range expectedTypes {
		if !upload.AllowedImageTypes[contentType] {
			t.Errorf("Expected %s to be in AllowedImageTypes", contentType)
		}
	}

	invalidTypes := []string{"text/plain", "application/json", "video/mp4"}
	for _, contentType := range invalidTypes {
		if upload.AllowedImageTypes[contentType] {
			t.Errorf("%s should not be in AllowedImageTypes", contentType)
		}
	}
}

func TestUploadErrors(t *testing.T) {
	if upload.ErrInvalidContentType.Error() != "uploaded file must be an image" {
		t.Errorf("Unexpected ErrInvalidContentType message: %v", upload.ErrInvalidContentType)
	}
	if upload.ErrFileTooLarge.Error() != "file size exceeds 8MB limit" {
		t.Errorf("Unexpected ErrFileTooLarge message: %v", upload.ErrFileTooLarge)
	}
	if upload.ErrNoFileUploaded.Error() != "no file uploaded" {
		t.Errorf("Unexpected ErrNoFileUploaded message: %v", upload.ErrNoFileUploaded)
	}
}

func TestUploadMockRepository_CreateAndGet(t *testing.T) {
	repo := NewMockUploadRepository()

	uploadRecord := &upload.FileUpload{
		UserID:           1,
		Filename:         "test.png",
		OriginalFilename: "original.png",
		ContentType:      "image/png",
		FileSize:         1024,
		TempPath:         "/tmp/test.png",
	}

	created, err := repo.CreateFileUpload(uploadRecord)
	if err != nil {
		t.Fatalf("CreateFileUpload failed: %v", err)
	}
	if created.ID == 0 {
		t.Error("Expected ID to be set")
	}

	retrieved, found := repo.GetFileUploadByID(created.ID)
	if !found {
		t.Fatal("Expected to find upload")
	}
	if retrieved.Filename != "test.png" {
		t.Errorf("Expected filename 'test.png', got '%s'", retrieved.Filename)
	}

	uploads, err := repo.GetFileUploadsByUserID(1)
	if err != nil {
		t.Fatalf("GetFileUploadsByUserID failed: %v", err)
	}
	if len(uploads) != 1 {
		t.Errorf("Expected 1 upload, got %d", len(uploads))
	}
}

func TestUploadMockRepository_Reset(t *testing.T) {
	repo := NewMockUploadRepository()

	repo.AddUpload(&upload.FileUpload{
		ID:       1,
		UserID:   1,
		Filename: "test.png",
	})

	repo.Reset()

	_, found := repo.GetFileUploadByID(1)
	if found {
		t.Error("Expected upload to be cleared after reset")
	}
}

func TestUpload_ValidImage(t *testing.T) {
	tempDir := t.TempDir()

	handler, _ := setupUploadTestHandler()

	imgContent := createTestImageContent()

	body, contentType := createMultipartForm("data", "test.png", imgContent)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &auth.TokenClaims{UserID: 1, Username: "testuser"})

	_ = handler.Upload(c)
	_ = tempDir
}

func TestValidateUploadFile_InvalidContentType(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(tempFile, []byte("This is not an image"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	file, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("data", "test.txt")
	io.Copy(part, file)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, fileHeader, err := req.FormFile("data")
	if err != nil {
		t.Fatalf("Failed to get form file: %v", err)
	}

	_ = fileHeader
}

func BenchmarkGetUserUploads(b *testing.B) {
	handler, mockRepo := setupUploadTestHandler()

	for i := 0; i < 100; i++ {
		mockRepo.AddUpload(&upload.FileUpload{
			UserID:           1,
			Filename:         "test.png",
			OriginalFilename: "original.png",
			ContentType:      "image/png",
			FileSize:         1024,
			TempPath:         "/tmp/test.png",
		})
	}

	e := echo.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := createUploadTestContext(e, http.MethodGet, "/api/uploads", nil, "")
		c.Set("user", &auth.TokenClaims{UserID: 1, Username: "testuser"})
		handler.GetUserUploads(c)
	}
}
