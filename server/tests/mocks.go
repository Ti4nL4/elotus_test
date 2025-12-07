package tests

import (
	"sync"
	"time"

	"elotus_test/server/models/upload"
	"elotus_test/server/models/user"
)

type MockUserRepository struct {
	mu              sync.RWMutex
	users           map[int64]*user.User
	byName          map[string]*user.User
	nextID          int64
	CreateUserError error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[int64]*user.User),
		byName: make(map[string]*user.User),
		nextID: 1,
	}
}

func (r *MockUserRepository) CreateUser(username, hashedPassword string) (*user.User, error) {
	if r.CreateUserError != nil {
		return nil, r.CreateUserError
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byName[username]; exists {
		return nil, user.ErrUserExists
	}

	u := &user.User{
		ID:        r.nextID,
		Username:  username,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}
	r.nextID++

	r.users[u.ID] = u
	r.byName[username] = u

	return u, nil
}

func (r *MockUserRepository) GetUserByUsername(username string) (*user.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.byName[username]
	return u, exists
}

func (r *MockUserRepository) GetUserByID(id int64) (*user.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.users[id]
	return u, exists
}

func (r *MockUserRepository) UpdateLastLogin(userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if u, exists := r.users[userID]; exists {
		now := time.Now()
		u.LastLoginAt = &now
	}
	return nil
}

func (r *MockUserRepository) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users = make(map[int64]*user.User)
	r.byName = make(map[string]*user.User)
	r.nextID = 1
	r.CreateUserError = nil
}

func (r *MockUserRepository) AddUser(u *user.User) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[u.ID] = u
	r.byName[u.Username] = u
	if u.ID >= r.nextID {
		r.nextID = u.ID + 1
	}
}

type MockUploadRepository struct {
	mu          sync.RWMutex
	uploads     map[int64]*upload.FileUpload
	nextID      int64
	CreateError error
	GetError    error
}

func NewMockUploadRepository() *MockUploadRepository {
	return &MockUploadRepository{
		uploads: make(map[int64]*upload.FileUpload),
		nextID:  1,
	}
}

func (r *MockUploadRepository) CreateFileUpload(u *upload.FileUpload) (*upload.FileUpload, error) {
	if r.CreateError != nil {
		return nil, r.CreateError
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	u.ID = r.nextID
	u.CreatedAt = time.Now()
	r.nextID++

	stored := *u
	r.uploads[u.ID] = &stored

	return u, nil
}

func (r *MockUploadRepository) GetFileUploadByID(id int64) (*upload.FileUpload, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.uploads[id]
	if !exists {
		return nil, false
	}

	result := *u
	return &result, true
}

func (r *MockUploadRepository) GetFileUploadsByUserID(userID int64) ([]*upload.FileUpload, error) {
	if r.GetError != nil {
		return nil, r.GetError
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*upload.FileUpload
	for _, u := range r.uploads {
		if u.UserID == userID {
			copied := *u
			result = append(result, &copied)
		}
	}

	return result, nil
}

func (r *MockUploadRepository) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.uploads = make(map[int64]*upload.FileUpload)
	r.nextID = 1
	r.CreateError = nil
	r.GetError = nil
}

func (r *MockUploadRepository) AddUpload(u *upload.FileUpload) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if u.ID == 0 {
		u.ID = r.nextID
		r.nextID++
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}

	stored := *u
	r.uploads[u.ID] = &stored

	if u.ID >= r.nextID {
		r.nextID = u.ID + 1
	}
}
