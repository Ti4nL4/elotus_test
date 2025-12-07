package auth

import (
	"database/sql"
	"fmt"
	"time"

	"elotus_test/server/bredis"
	"elotus_test/server/bsql"
	"elotus_test/server/env"
)

type TokenRevocationStore struct {
	db    *bsql.DB
	redis *bredis.Client
}

func NewTokenRevocationStore(db *bsql.DB, redis *bredis.Client) *TokenRevocationStore {
	return &TokenRevocationStore{db: db, redis: redis}
}

func (s *TokenRevocationStore) cacheKey(userID int64) string {
	return fmt.Sprintf("revoke:%d", userID)
}

func (s *TokenRevocationStore) RevokeUserTokensBefore(userID int64, before time.Time) error {
	_, err := s.db.Exec(
		`UPDATE users SET last_revoked_token_at = $1 
		 WHERE id = $2 AND (last_revoked_token_at IS NULL OR last_revoked_token_at < $1)`,
		before, userID,
	)
	if err != nil {
		return err
	}

	if s.redis != nil {
		_ = s.redis.Delete(s.cacheKey(userID))
	}

	return nil
}

func (s *TokenRevocationStore) RevokeAllUserTokens(userID int64) error {
	return s.RevokeUserTokensBefore(userID, time.Now())
}

// IsTokenRevoked checks if a token is revoked.
// A token is revoked if it was issued BEFORE the user's last_revoked_token_at timestamp.
func (s *TokenRevocationStore) IsTokenRevoked(userID int64, issuedAt time.Time) bool {
	cacheKey := s.cacheKey(userID)

	if s.redis != nil {
		var cachedTime time.Time
		if err := s.redis.Get(cacheKey, &cachedTime); err == nil {
			return issuedAt.Before(cachedTime)
		}
	}

	var lastRevokedAt sql.NullTime
	err := s.db.QueryRow(
		"SELECT last_revoked_token_at FROM users WHERE id = $1",
		userID,
	).Scan(&lastRevokedAt)

	if err != nil || !lastRevokedAt.Valid {
		return false
	}

	if s.redis != nil {
		_ = s.redis.Set(cacheKey, lastRevokedAt.Time, env.E.GetRevokeDuration())
	}

	return issuedAt.Before(lastRevokedAt.Time)
}
