-- Migration: Create file_uploads table
-- Created at: 2025-12-06

-- +migrate Up
CREATE TABLE IF NOT EXISTS file_uploads (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    temp_path VARCHAR(500) NOT NULL,
    
    -- HTTP metadata
    client_ip VARCHAR(45),
    user_agent TEXT,
    request_host VARCHAR(255),
    request_uri VARCHAR(500),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_file_uploads_user_id ON file_uploads(user_id);
CREATE INDEX IF NOT EXISTS idx_file_uploads_created_at ON file_uploads(created_at);

-- +migrate Down
DROP INDEX IF EXISTS idx_file_uploads_created_at;
DROP INDEX IF EXISTS idx_file_uploads_user_id;
DROP TABLE IF EXISTS file_uploads;

