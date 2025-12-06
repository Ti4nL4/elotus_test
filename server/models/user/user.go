package user

import (
	"errors"
	"time"
)

// User represents a registered user
type User struct {
	ID                 int64      `json:"id"`
	Username           string     `json:"username"`
	Password           string     `json:"-"`
	CreatedAt          time.Time  `json:"created_at"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	LastRevokedTokenAt *time.Time `json:"last_revoked_token_at,omitempty"`
}

// Repository defines the interface for user data access
type Repository interface {
	CreateUser(username, hashedPassword string) (*User, error)
	GetUserByUsername(username string) (*User, bool)
	GetUserByID(id int64) (*User, bool)
	UpdateLastLogin(userID int64) error
}

// Errors
var (
	ErrUserExists   = errors.New("username already exists")
	ErrUserNotFound = errors.New("user not found")
)
