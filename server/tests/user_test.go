package tests

import (
	"testing"
	"time"

	"elotus_test/server/models/user"
)

var _ user.Repository = (*MockUserRepository)(nil)

func TestMockRepository_CreateUser_Success(t *testing.T) {
	repo := NewMockUserRepository()

	u, err := repo.CreateUser("testuser", "hashedpassword123")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if u.ID == 0 {
		t.Error("Expected user ID to be set")
	}
	if u.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", u.Username)
	}
	if u.Password != "hashedpassword123" {
		t.Errorf("Expected password 'hashedpassword123', got '%s'", u.Password)
	}
	if u.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestMockRepository_CreateUser_Duplicate(t *testing.T) {
	repo := NewMockUserRepository()

	_, err := repo.CreateUser("testuser", "password1")
	if err != nil {
		t.Fatalf("First CreateUser failed: %v", err)
	}

	_, err = repo.CreateUser("testuser", "password2")
	if err != user.ErrUserExists {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}
}

func TestMockRepository_CreateUser_WithError(t *testing.T) {
	repo := NewMockUserRepository()
	repo.CreateUserError = user.ErrUserNotFound

	_, err := repo.CreateUser("testuser", "password")
	if err != user.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestMockRepository_GetUserByUsername_Found(t *testing.T) {
	repo := NewMockUserRepository()

	repo.CreateUser("testuser", "password")

	u, found := repo.GetUserByUsername("testuser")
	if !found {
		t.Fatal("Expected to find user")
	}
	if u.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", u.Username)
	}
}

func TestMockRepository_GetUserByUsername_NotFound(t *testing.T) {
	repo := NewMockUserRepository()

	u, found := repo.GetUserByUsername("nonexistent")
	if found {
		t.Error("Expected user not to be found")
	}
	if u != nil {
		t.Error("Expected nil user")
	}
}

func TestMockRepository_GetUserByID_Found(t *testing.T) {
	repo := NewMockUserRepository()

	created, _ := repo.CreateUser("testuser", "password")

	u, found := repo.GetUserByID(created.ID)
	if !found {
		t.Fatal("Expected to find user")
	}
	if u.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, u.ID)
	}
}

func TestMockRepository_GetUserByID_NotFound(t *testing.T) {
	repo := NewMockUserRepository()

	u, found := repo.GetUserByID(999)
	if found {
		t.Error("Expected user not to be found")
	}
	if u != nil {
		t.Error("Expected nil user")
	}
}

func TestMockRepository_UpdateLastLogin(t *testing.T) {
	repo := NewMockUserRepository()

	created, _ := repo.CreateUser("testuser", "password")

	err := repo.UpdateLastLogin(created.ID)
	if err != nil {
		t.Fatalf("UpdateLastLogin failed: %v", err)
	}

	u, _ := repo.GetUserByID(created.ID)
	if u.LastLoginAt == nil {
		t.Error("Expected LastLoginAt to be set")
	}
}

func TestMockRepository_UpdateLastLogin_NonExistentUser(t *testing.T) {
	repo := NewMockUserRepository()

	err := repo.UpdateLastLogin(999)
	if err != nil {
		t.Errorf("Expected no error for non-existent user, got %v", err)
	}
}

func TestMockRepository_AddUser(t *testing.T) {
	repo := NewMockUserRepository()

	u := &user.User{
		ID:        100,
		Username:  "manualuser",
		Password:  "password",
		CreatedAt: time.Now(),
	}
	repo.AddUser(u)

	retrieved, found := repo.GetUserByID(100)
	if !found {
		t.Fatal("Expected to find manually added user")
	}
	if retrieved.Username != "manualuser" {
		t.Errorf("Expected username 'manualuser', got '%s'", retrieved.Username)
	}

	_, found = repo.GetUserByUsername("manualuser")
	if !found {
		t.Fatal("Expected to find manually added user by username")
	}
}

func TestMockRepository_Reset(t *testing.T) {
	repo := NewMockUserRepository()

	repo.CreateUser("testuser", "password")

	repo.Reset()

	_, found := repo.GetUserByUsername("testuser")
	if found {
		t.Error("Expected user to be cleared after reset")
	}
}

func TestMockRepository_ConcurrentAccess(t *testing.T) {
	repo := NewMockUserRepository()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			username := "user" + string(rune('a'+id))
			repo.CreateUser(username, "password")
			repo.GetUserByUsername(username)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMockRepository_NextID_AutoIncrement(t *testing.T) {
	repo := NewMockUserRepository()

	user1, _ := repo.CreateUser("user1", "password")
	user2, _ := repo.CreateUser("user2", "password")
	user3, _ := repo.CreateUser("user3", "password")

	if user1.ID >= user2.ID || user2.ID >= user3.ID {
		t.Errorf("Expected incrementing IDs, got %d, %d, %d", user1.ID, user2.ID, user3.ID)
	}
}

func TestMockRepository_AddUser_UpdatesNextID(t *testing.T) {
	repo := NewMockUserRepository()

	repo.AddUser(&user.User{
		ID:        100,
		Username:  "user100",
		Password:  "password",
		CreatedAt: time.Now(),
	})

	newUser, _ := repo.CreateUser("newuser", "password")
	if newUser.ID <= 100 {
		t.Errorf("Expected ID > 100, got %d", newUser.ID)
	}
}

func TestUserErrors(t *testing.T) {
	if user.ErrUserExists.Error() != "username already exists" {
		t.Errorf("Unexpected ErrUserExists message: %v", user.ErrUserExists)
	}
	if user.ErrUserNotFound.Error() != "user not found" {
		t.Errorf("Unexpected ErrUserNotFound message: %v", user.ErrUserNotFound)
	}
}

func TestUserStruct(t *testing.T) {
	now := time.Now()
	lastLogin := now.Add(-time.Hour)
	lastRevoked := now.Add(-2 * time.Hour)

	u := user.User{
		ID:                 1,
		Username:           "testuser",
		Password:           "hashedpassword",
		CreatedAt:          now,
		LastLoginAt:        &lastLogin,
		LastRevokedTokenAt: &lastRevoked,
	}

	if u.ID != 1 {
		t.Errorf("Expected ID 1, got %d", u.ID)
	}
	if u.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", u.Username)
	}
	if u.LastLoginAt == nil || !u.LastLoginAt.Equal(lastLogin) {
		t.Error("LastLoginAt not set correctly")
	}
	if u.LastRevokedTokenAt == nil || !u.LastRevokedTokenAt.Equal(lastRevoked) {
		t.Error("LastRevokedTokenAt not set correctly")
	}
}

func BenchmarkMockRepository_CreateUser(b *testing.B) {
	repo := NewMockUserRepository()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		username := "user" + string(rune(i%26+'a'))
		repo.CreateUser(username+string(rune(i/26)), "password")
	}
}

func BenchmarkMockRepository_GetUserByUsername(b *testing.B) {
	repo := NewMockUserRepository()

	for i := 0; i < 100; i++ {
		username := "user" + string(rune(i%26+'a')) + string(rune(i/26+'a'))
		repo.CreateUser(username, "password")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetUserByUsername("usera")
	}
}

func BenchmarkMockRepository_GetUserByID(b *testing.B) {
	repo := NewMockUserRepository()

	for i := 0; i < 100; i++ {
		username := "user" + string(rune(i%26+'a')) + string(rune(i/26+'a'))
		repo.CreateUser(username, "password")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetUserByID(50)
	}
}
