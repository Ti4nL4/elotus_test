-- Migration: Create token revocations table
-- Created at: 2024-12-06

-- +migrate Up
CREATE TABLE IF NOT EXISTS token_revocations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    revoked_before TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_token_revocations_user_id ON token_revocations(user_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_token_revocations_user_id;
DROP TABLE IF EXISTS token_revocations;

