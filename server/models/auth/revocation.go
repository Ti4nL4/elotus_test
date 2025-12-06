package auth

import (
	"sync"
	"time"
)

// TokenRevocationStore manages token revocation by time
type TokenRevocationStore struct {
	sync.RWMutex
	revokedBefore       map[int64]time.Time
	globalRevokedBefore time.Time
}

// NewTokenRevocationStore creates a new TokenRevocationStore
func NewTokenRevocationStore() *TokenRevocationStore {
	return &TokenRevocationStore{
		revokedBefore: make(map[int64]time.Time),
	}
}

// RevokeUserTokensBefore revokes all tokens for a user issued before the given time
func (s *TokenRevocationStore) RevokeUserTokensBefore(userID int64, before time.Time) {
	s.Lock()
	defer s.Unlock()
	if existing, ok := s.revokedBefore[userID]; !ok || before.After(existing) {
		s.revokedBefore[userID] = before
	}
}

// RevokeAllUserTokens revokes all current tokens for a user
func (s *TokenRevocationStore) RevokeAllUserTokens(userID int64) {
	s.RevokeUserTokensBefore(userID, time.Now())
}

// RevokeGlobalTokensBefore revokes all tokens issued before the given time
func (s *TokenRevocationStore) RevokeGlobalTokensBefore(before time.Time) {
	s.Lock()
	defer s.Unlock()
	if before.After(s.globalRevokedBefore) {
		s.globalRevokedBefore = before
	}
}

// IsTokenRevoked checks if a token is revoked based on its issued time
func (s *TokenRevocationStore) IsTokenRevoked(userID int64, issuedAt time.Time) bool {
	s.RLock()
	defer s.RUnlock()

	if !s.globalRevokedBefore.IsZero() && issuedAt.Before(s.globalRevokedBefore) {
		return true
	}

	if revokedBefore, ok := s.revokedBefore[userID]; ok {
		return issuedAt.Before(revokedBefore)
	}

	return false
}
