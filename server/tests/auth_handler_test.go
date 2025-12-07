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
	"elotus_test/server/response"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

var _ user.Repository = (*MockUserRepository)(nil)

func setupAuthTestHandler() (*auth.Handler, *MockUserRepository, *auth.JWTService) {
	userRepo := NewMockUserRepository()
	jwtService := auth.NewJWTService(&auth.Config{
		SecretKey:     []byte("test-secret-key-for-testing-only"),
		TokenDuration: time.Hour,
	}, nil)

	handler := auth.NewHandler(nil, userRepo, jwtService, nil)
	return handler, userRepo, jwtService
}

func parseResponse(body []byte) (*response.Response, error) {
	var resp response.Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func getDataMap(resp *response.Response) map[string]interface{} {
	if resp.Data == nil {
		return nil
	}
	if m, ok := resp.Data.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func getErrorMessage(resp *response.Response) string {
	if resp.Error == nil {
		return ""
	}
	return resp.Error.Message
}

func TestRegister_Success(t *testing.T) {
	handler, _, _ := setupAuthTestHandler()

	e := echo.New()
	reqBody := `{"username": "testuser", "password": "Password123"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Register(c)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	resp, err := parseResponse(rec.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false. Error: %v", getErrorMessage(resp))
	}

	data := getDataMap(resp)
	if data == nil {
		t.Fatal("Expected data in response")
	}

	if data["message"] != "User registered successfully" {
		t.Errorf("Expected success message, got: %v", data["message"])
	}

	userData, ok := data["user"].(map[string]interface{})
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
	reqBody := `{"username": "", "password": "Password123"}`
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

	resp, _ := parseResponse(rec.Body.Bytes())
	if resp.Success {
		t.Error("Expected success=false for validation error")
	}
	errMsg := getErrorMessage(resp)
	if errMsg != "Username is required" {
		t.Errorf("Expected validation error, got: %v", errMsg)
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

	resp, _ := parseResponse(rec.Body.Bytes())
	if resp.Success {
		t.Error("Expected success=false for validation error")
	}
	errMsg := getErrorMessage(resp)
	if errMsg != "Password must be at least 8 characters" {
		t.Errorf("Expected password length error, got: %v", errMsg)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	handler, userRepo, _ := setupAuthTestHandler()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
	userRepo.AddUser(&user.User{
		ID:        1,
		Username:  "testuser",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	})

	e := echo.New()
	reqBody := `{"username": "testuser", "password": "Password123"}`
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

	resp, _ := parseResponse(rec.Body.Bytes())
	if resp.Success {
		t.Error("Expected success=false for conflict error")
	}
	errMsg := getErrorMessage(resp)
	if errMsg != "Username already exists" {
		t.Errorf("Expected duplicate error, got: %v", errMsg)
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

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
	userRepo.AddUser(&user.User{
		ID:        1,
		Username:  "testuser",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	})

	e := echo.New()
	reqBody := `{"username": "testuser", "password": "Password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Login(c)
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	resp, err := parseResponse(rec.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false. Error: %v", getErrorMessage(resp))
	}

	data := getDataMap(resp)
	if data == nil {
		t.Fatal("Expected data in response")
	}

	if data["token"] == nil || data["token"] == "" {
		t.Error("Expected token in response")
	}
}

func TestLogin_InvalidUsername(t *testing.T) {
	handler, _, _ := setupAuthTestHandler()

	e := echo.New()
	reqBody := `{"username": "nonexistent", "password": "Password123"}`
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

	resp, _ := parseResponse(rec.Body.Bytes())
	if resp.Success {
		t.Error("Expected success=false for unauthorized error")
	}
	errMsg := getErrorMessage(resp)
	if errMsg != "Invalid username or password" {
		t.Errorf("Expected invalid credentials error, got: %v", errMsg)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	handler, userRepo, _ := setupAuthTestHandler()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
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

	resp, _ := parseResponse(rec.Body.Bytes())
	if resp.Success {
		t.Error("Expected success=false for unauthorized error")
	}
	errMsg := getErrorMessage(resp)
	if errMsg != "Invalid username or password" {
		t.Errorf("Expected invalid credentials error, got: %v", errMsg)
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

	resp, _ := parseResponse(rec.Body.Bytes())
	if resp.Success {
		t.Error("Expected success=false for validation error")
	}
	errMsg := getErrorMessage(resp)
	if errMsg != "Username and password are required" {
		t.Errorf("Expected validation error, got: %v", errMsg)
	}
}

func TestProtected_Success(t *testing.T) {
	handler, _, jwtService := setupAuthTestHandler()

	token, _, err := jwtService.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", claims)

	err = handler.Protected(c)
	if err != nil {
		t.Fatalf("Protected returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	resp, _ := parseResponse(rec.Body.Bytes())
	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}
	data := getDataMap(resp)
	if data["message"] != "Welcome to the protected endpoint!" {
		t.Errorf("Expected welcome message, got: %v", data["message"])
	}
	if data["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got: %v", data["username"])
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

	resp, _ := parseResponse(rec.Body.Bytes())
	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}
	data := getDataMap(resp)

	// With no DB and no Redis, status should be DEGRADED
	if data["status"] != "DEGRADED" {
		t.Errorf("Expected status 'DEGRADED' (no DB/Redis in test), got: %v", data["status"])
	}
	if data["db"] != "NOT_CONFIGURED" {
		t.Errorf("Expected db 'NOT_CONFIGURED', got: %v", data["db"])
	}
	if data["redis"] != "DISABLED" {
		t.Errorf("Expected redis 'DISABLED', got: %v", data["redis"])
	}
	if data["timestamp"] == nil {
		t.Error("Expected timestamp to be present")
	}
}

func TestRevokeToken_RevokeAll(t *testing.T) {
	handler, _, jwtService := setupAuthTestHandler()

	token, _, err := jwtService.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

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

	resp, _ := parseResponse(rec.Body.Bytes())
	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}
	data := getDataMap(resp)
	if data["message"] != "All tokens have been revoked" {
		t.Errorf("Expected revoke message, got: %v", data["message"])
	}
}
