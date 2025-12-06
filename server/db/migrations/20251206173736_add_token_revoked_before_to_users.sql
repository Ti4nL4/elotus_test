-- Migration: Add token_revoked_before column to users table
-- Created at: 2025-12-06

-- +migrate Up
ALTER TABLE users ADD COLUMN token_revoked_before TIMESTAMP WITH TIME ZONE;

-- Drop the separate token_revocations table (no longer needed)
DROP INDEX IF EXISTS idx_token_revocations_user_id;
DROP TABLE IF EXISTS token_revocations;

-- +migrate Down
-- Recreate token_revocations table
CREATE TABLE IF NOT EXISTS token_revocations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    revoked_before TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_token_revocations_user_id ON token_revocations(user_id);

ALTER TABLE users DROP COLUMN IF EXISTS token_revoked_before;
