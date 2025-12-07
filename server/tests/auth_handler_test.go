package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"elotus_test/server/models/auth"
	"elotus_test/server/models/user"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Ensure MockUserRepository implements user.Repository
var _ user.Repository = (*MockUserRepository)(nil)

// setupTestHandler creates a test handler with mock dependencies
func setupAuthTestHandler() (*auth.Handler, *MockUserRepository, *auth.JWTService) {
	userRepo := NewMockUserRepository()
	jwtService := auth.NewJWTService(&auth.Config{
		SecretKey:     []byte("test-secret-key-for-testing-only"),
		TokenDuration: time.Hour,
	}, nil)

	handler := auth.NewHandler(userRepo, jwtService, nil)
	return handler, userRepo, jwtService
}

func TestRegister_Success(t *testing.T) {
	handler, _, _ := setupAuthTestHandler()

	e := echo.New()
	reqBody := `{"username": "testuser", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Register(c)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["message"] != "User registered successfully" {
		t.Errorf("Expected success message, got: %v", response["message"])
	}

	userData, ok := response["user"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected user data in response")
	}
	if userData["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got: %v", userData["username"])
	}
}

func TestRegister_EmptyUsername(t *testing.T) {
	handler, _, _ := setupAuthTestHandler()

	e := echo.New()
	reqBody := `{"username": "", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Register(c)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["error"] != "Username and password are required" {
		t.Errorf("Expected validation error, got: %v", response["error"])
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	handler, _, _ := setupAuthTestHandler()

	e := echo.New()
	reqBody := `{"username": "testuser", "password": "123"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Register(c)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["error"] != "Password must be at least 6 characters" {
		t.Errorf("Expected password length error, got: %v", response["error"])
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	handler, userRepo, _ := setupAuthTestHandler()

	// First, create a user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userRepo.AddUser(&user.User{
		ID:        1,
		Username:  "testuser",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	})

	e := echo.New()
	reqBody := `{"username": "testuser", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Register(c)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["error"] != "Username already exists" {
		t.Errorf("Expected duplicate error, got: %v", response["error"])
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	handler, _, _ := setupAuthTestHandler()

	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Register(c)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	handler, userRepo, _ := setupAuthTestHandler()

	// Create a user first
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userRepo.AddUser(&user.User{
		ID:        1,
		Username:  "testuser",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	})

	e := echo.New()
	reqBody := `{"username": "testuser", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Login(c)
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response auth.LoginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token in response")
	}

	if response.ExpiresAt.Before(time.Now()) {
		t.Error("Token should expire in the future")
	}
}

func TestLogin_InvalidUsername(t *testing.T) {
	handler, _, _ := setupAuthTestHandler()

	e := echo.New()
	reqBody := `{"username": "nonexistent", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Login(c)
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["error"] != "Invalid username or password" {
		t.Errorf("Expected invalid credentials error, got: %v", response["error"])
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	handler, userRepo, _ := setupAuthTestHandler()

	// Create a user first
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userRepo.AddUser(&user.User{
		ID:        1,
		Username:  "testuser",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	})

	e := echo.New()
	reqBody := `{"username": "testuser", "password": "wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Login(c)
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["error"] != "Invalid username or password" {
		t.Errorf("Expected invalid credentials error, got: %v", response["error"])
	}
}

func TestLogin_EmptyCredentials(t *testing.T) {
	handler, _, _ := setupAuthTestHandler()

	e := echo.New()
	reqBody := `{"username": "", "password": ""}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Login(c)
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["error"] != "Username and password are required" {
		t.Errorf("Expected validation error, got: %v", response["error"])
	}
}

func TestProtected_Success(t *testing.T) {
	handler, _, jwtService := setupAuthTestHandler()

	// Generate a valid token
	token, _, err := jwtService.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate token to get claims
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", claims) // Simulate middleware setting claims

	err = handler.Protected(c)
	if err != nil {
		t.Fatalf("Protected returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["message"] != "Welcome to the protected endpoint!" {
		t.Errorf("Expected welcome message, got: %v", response["message"])
	}
	if response["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got: %v", response["username"])
	}
}

func TestHealthCheck(t *testing.T) {
	handler, _, _ := setupAuthTestHandler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HealthCheck(c)
	if err != nil {
		t.Fatalf("HealthCheck returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got: %v", response["status"])
	}
}

func TestRevokeToken_RevokeAll(t *testing.T) {
	handler, _, jwtService := setupAuthTestHandler()

	// Generate a valid token
	token, _, err := jwtService.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate token to get claims
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/revoke", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", claims)

	err = handler.RevokeToken(c)
	if err != nil {
		t.Fatalf("RevokeToken returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["message"] != "All tokens have been revoked" {
		t.Errorf("Expected revoke message, got: %v", response["message"])
	}
}
