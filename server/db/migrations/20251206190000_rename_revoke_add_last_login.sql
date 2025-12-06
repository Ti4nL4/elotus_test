-- Migration: Rename token_revoked_before to last_revoked_token_at and add last_login_at
-- Created at: 2025-12-06

-- +migrate Up
ALTER TABLE users RENAME COLUMN token_revoked_before TO last_revoked_token_at;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_users_last_login_at ON users(last_login_at);

-- +migrate Down
DROP INDEX IF EXISTS idx_users_last_login_at;
ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE users RENAME COLUMN last_revoked_token_at TO token_revoked_before;

