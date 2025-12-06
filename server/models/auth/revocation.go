package auth

import (
	"database/sql"
	"time"

	"elotus_test/server/bsql"
	"elotus_test/server/env"
)

// TokenRevocationStore manages token revocation using the users table
type TokenRevocationStore struct {
	db *bsql.DB
}

// NewTokenRevocationStore creates a new TokenRevocationStore
func NewTokenRevocationStore(db *bsql.DB) *TokenRevocationStore {
	return &TokenRevocationStore{db: db}
}

// RevokeUserTokensBefore revokes all tokens for a user issued before the given time
func (s *TokenRevocationStore) RevokeUserTokensBefore(userID int64, before time.Time) error {
	_, err := s.db.Exec(
		`UPDATE users SET token_revoked_before = $1 
		 WHERE id = $2 AND (token_revoked_before IS NULL OR token_revoked_before < $1)`,
		before, userID,
	)
	return err
}

// RevokeAllUserTokens revokes all current tokens for a user
func (s *TokenRevocationStore) RevokeAllUserTokens(userID int64) error {
	return s.RevokeUserTokensBefore(userID, time.Now())
}

// IsTokenRevoked checks if a token is revoked based on its issued time
// Revocation expires after env.E.GetRevokeDuration()
func (s *TokenRevocationStore) IsTokenRevoked(userID int64, issuedAt time.Time) bool {
	var revokedBefore sql.NullTime
	err := s.db.QueryRow(
		"SELECT token_revoked_before FROM users WHERE id = $1",
		userID,
	).Scan(&revokedBefore)

	if err != nil || !revokedBefore.Valid {
		return false
	}

	// Check if revocation has expired
	if time.Since(revokedBefore.Time) > env.E.GetRevokeDuration() {
		return false
	}

	return issuedAt.Before(revokedBefore.Time)
}

// GetRevocationTime returns the revocation time for a user (if any)
func (s *TokenRevocationStore) GetRevocationTime(userID int64) (*time.Time, error) {
	var revokedBefore sql.NullTime
	err := s.db.QueryRow(
		"SELECT token_revoked_before FROM users WHERE id = $1",
		userID,
	).Scan(&revokedBefore)

	if err != nil || !revokedBefore.Valid {
		return nil, nil
	}

	return &revokedBefore.Time, nil
}

// ClearUserRevocation clears revocation for a user (allows all tokens again)
func (s *TokenRevocationStore) ClearUserRevocation(userID int64) error {
	_, err := s.db.Exec(
		"UPDATE users SET token_revoked_before = NULL WHERE id = $1",
		userID,
	)
	return err
}
