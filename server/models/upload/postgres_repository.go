package upload

import (
	"database/sql"
	"time"

	"elotus_test/server/bsql"
)

// PostgresRepository implements the Repository interface for PostgreSQL
type PostgresRepository struct {
	db *bsql.DB
}

// NewPostgresRepository creates a new PostgresRepository
func NewPostgresRepository(db *bsql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateFileUpload inserts a new file upload record into the database
func (r *PostgresRepository) CreateFileUpload(upload *FileUpload) (*FileUpload, error) {
	query := `
		INSERT INTO file_uploads (
			user_id, filename, original_filename, content_type, file_size, 
			temp_path, client_ip, user_agent, request_host, request_uri, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at`

	now := time.Now()
	err := r.db.QueryRow(
		query,
		upload.UserID,
		upload.Filename,
		upload.OriginalFilename,
		upload.ContentType,
		upload.FileSize,
		upload.TempPath,
		upload.ClientIP,
		upload.UserAgent,
		upload.RequestHost,
		upload.RequestURI,
		now,
	).Scan(&upload.ID, &upload.CreatedAt)

	if err != nil {
		return nil, err
	}

	return upload, nil
}

// GetFileUploadByID retrieves a file upload by its ID
func (r *PostgresRepository) GetFileUploadByID(id int64) (*FileUpload, bool) {
	query := `
		SELECT id, user_id, filename, original_filename, content_type, file_size,
			   temp_path, client_ip, user_agent, request_host, request_uri, created_at
		FROM file_uploads
		WHERE id = $1`

	upload := &FileUpload{}
	var clientIP, userAgent, requestHost, requestURI sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&upload.ID,
		&upload.UserID,
		&upload.Filename,
		&upload.OriginalFilename,
		&upload.ContentType,
		&upload.FileSize,
		&upload.TempPath,
		&clientIP,
		&userAgent,
		&requestHost,
		&requestURI,
		&upload.CreatedAt,
	)

	if err != nil {
		return nil, false
	}

	upload.ClientIP = clientIP.String
	upload.UserAgent = userAgent.String
	upload.RequestHost = requestHost.String
	upload.RequestURI = requestURI.String

	return upload, true
}

// GetFileUploadsByUserID retrieves all file uploads for a specific user
func (r *PostgresRepository) GetFileUploadsByUserID(userID int64) ([]*FileUpload, error) {
	query := `
		SELECT id, user_id, filename, original_filename, content_type, file_size,
			   temp_path, client_ip, user_agent, request_host, request_uri, created_at
		FROM file_uploads
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploads []*FileUpload
	for rows.Next() {
		upload := &FileUpload{}
		var clientIP, userAgent, requestHost, requestURI sql.NullString

		err := rows.Scan(
			&upload.ID,
			&upload.UserID,
			&upload.Filename,
			&upload.OriginalFilename,
			&upload.ContentType,
			&upload.FileSize,
			&upload.TempPath,
			&clientIP,
			&userAgent,
			&requestHost,
			&requestURI,
			&upload.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		upload.ClientIP = clientIP.String
		upload.UserAgent = userAgent.String
		upload.RequestHost = requestHost.String
		upload.RequestURI = requestURI.String

		uploads = append(uploads, upload)
	}

	return uploads, nil
}
