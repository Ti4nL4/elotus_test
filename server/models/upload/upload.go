package upload

import (
	"errors"
	"time"
)

type FileUpload struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"user_id"`
	Filename         string    `json:"filename"`
	OriginalFilename string    `json:"original_filename"`
	ContentType      string    `json:"content_type"`
	FileSize         int64     `json:"file_size"`
	TempPath         string    `json:"temp_path"`
	ClientIP         string    `json:"client_ip"`
	UserAgent        string    `json:"user_agent"`
	RequestHost      string    `json:"request_host"`
	RequestURI       string    `json:"request_uri"`
	CreatedAt        time.Time `json:"created_at"`
}

type Repository interface {
	CreateFileUpload(upload *FileUpload) (*FileUpload, error)
	GetFileUploadByID(id int64) (*FileUpload, bool)
	GetFileUploadsByUserID(userID int64) ([]*FileUpload, error)
}

var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/bmp":  true,
	"image/tiff": true,
}

const MaxFileSize = 8 * 1024 * 1024

var (
	ErrInvalidContentType = errors.New("uploaded file must be an image")
	ErrFileTooLarge       = errors.New("file size exceeds 8MB limit")
	ErrNoFileUploaded     = errors.New("no file uploaded")
)
