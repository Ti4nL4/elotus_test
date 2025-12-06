package user

import (
	"sync"
	"time"
)

// MemoryRepository is an in-memory user storage
type MemoryRepository struct {
	sync.RWMutex
	users     map[string]*User
	idCounter int64
}

// NewMemoryRepository creates a new MemoryRepository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:     make(map[string]*User),
		idCounter: 0,
	}
}

// CreateUser adds a new user to the store
func (r *MemoryRepository) CreateUser(username, hashedPassword string) (*User, error) {
	r.Lock()
	defer r.Unlock()

	if _, exists := r.users[username]; exists {
		return nil, ErrUserExists
	}

	r.idCounter++
	user := &User{
		ID:        r.idCounter,
		Username:  username,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}
	r.users[username] = user
	return user, nil
}

// GetUserByUsername retrieves a user by username
func (r *MemoryRepository) GetUserByUsername(username string) (*User, bool) {
	r.RLock()
	defer r.RUnlock()
	user, exists := r.users[username]
	return user, exists
}

// GetUserByID retrieves a user by ID
func (r *MemoryRepository) GetUserByID(id int64) (*User, bool) {
	r.RLock()
	defer r.RUnlock()
	for _, user := range r.users {
		if user.ID == id {
			return user, true
		}
	}
	return nil, false
}
