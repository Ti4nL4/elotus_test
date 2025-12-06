package user

import (
	"errors"
	"time"
)

// User represents a registered user
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// Repository defines the interface for user data access
type Repository interface {
	CreateUser(username, hashedPassword string) (*User, error)
	GetUserByUsername(username string) (*User, bool)
	GetUserByID(id int64) (*User, bool)
}

// Errors
var (
	ErrUserExists   = errors.New("username already exists")
	ErrUserNotFound = errors.New("user not found")
)
