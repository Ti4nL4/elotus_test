package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config holds JWT configuration
type Config struct {
	SecretKey     []byte
	TokenDuration time.Duration
}

// DefaultConfig returns default JWT configuration
func DefaultConfig() *Config {
	return &Config{
		SecretKey:     []byte("your-256-bit-secret-key-here-change-in-production"),
		TokenDuration: 24 * time.Hour,
	}
}

// TokenClaims represents the JWT claims
type TokenClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTService handles JWT operations
type JWTService struct {
	config          *Config
	revocationStore *TokenRevocationStore
}

// NewJWTService creates a new JWT service
func NewJWTService(config *Config, revocationStore *TokenRevocationStore) *JWTService {
	if config == nil {
		config = DefaultConfig()
	}
	return &JWTService{
		config:          config,
		revocationStore: revocationStore,
	}
}

// GenerateToken generates a new JWT token for a user
func (s *JWTService) GenerateToken(userID int64, username string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.config.TokenDuration)

	claims := &TokenClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "elotus-auth",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.config.SecretKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.config.SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token is revoked
	if s.revocationStore != nil {
		issuedAt := claims.IssuedAt.Time
		if s.revocationStore.IsTokenRevoked(claims.UserID, issuedAt) {
			return nil, errors.New("token has been revoked")
		}
	}

	return claims, nil
}

// RevokeUserTokens revokes all tokens for a specific user
func (s *JWTService) RevokeUserTokens(userID int64) {
	if s.revocationStore != nil {
		s.revocationStore.RevokeAllUserTokens(userID)
	}
}

// RevokeUserTokensBefore revokes all tokens for a user issued before the given time
func (s *JWTService) RevokeUserTokensBefore(userID int64, before time.Time) {
	if s.revocationStore != nil {
		s.revocationStore.RevokeUserTokensBefore(userID, before)
	}
}
