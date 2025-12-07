package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"elotus_test/server/middleware"

	"github.com/labstack/echo/v4"
)

// mockClaims for testing
type mockClaims struct {
	UserID   int64
	Username string
}

func TestJWTMiddleware_Success(t *testing.T) {
	validateFn := func(token string) (interface{}, error) {
		if token == "valid-token" {
			return &mockClaims{UserID: 1, Username: "testuser"}, nil
		}
		return nil, errors.New("invalid token")
	}

	mw := middleware.JWTMiddleware(validateFn)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		claims := c.Get("user").(*mockClaims)
		return c.JSON(http.StatusOK, echo.Map{
			"user_id":  claims.UserID,
			"username": claims.Username,
		})
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got: %v", response["username"])
	}
}

func TestJWTMiddleware_MissingAuthHeader(t *testing.T) {
	validateFn := func(token string) (interface{}, error) {
		return nil, nil
	}

	mw := middleware.JWTMiddleware(validateFn)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// No Authorization header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["error"] != "Authorization header required" {
		t.Errorf("Expected auth header error, got: %v", response["error"])
	}
}

func TestJWTMiddleware_InvalidFormat_NoBearer(t *testing.T) {
	validateFn := func(token string) (interface{}, error) {
		return nil, nil
	}

	mw := middleware.JWTMiddleware(validateFn)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["error"] != "Invalid authorization header format" {
		t.Errorf("Expected format error, got: %v", response["error"])
	}
}

func TestJWTMiddleware_InvalidFormat_OnlyBearer(t *testing.T) {
	validateFn := func(token string) (interface{}, error) {
		return nil, nil
	}

	mw := middleware.JWTMiddleware(validateFn)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	validateFn := func(token string) (interface{}, error) {
		return nil, errors.New("token expired")
	}

	mw := middleware.JWTMiddleware(validateFn)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	errMsg, ok := response["error"].(string)
	if !ok || errMsg == "" {
		t.Error("Expected error message in response")
	}
}

func TestJWTMiddleware_BearerCaseInsensitive(t *testing.T) {
	validateFn := func(token string) (interface{}, error) {
		if token == "valid-token" {
			return &mockClaims{UserID: 1, Username: "testuser"}, nil
		}
		return nil, errors.New("invalid token")
	}

	mw := middleware.JWTMiddleware(validateFn)

	testCases := []string{
		"bearer valid-token",
		"BEARER valid-token",
		"Bearer valid-token",
	}

	for _, authHeader := range testCases {
		t.Run(authHeader, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", authHeader)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := mw(func(c echo.Context) error {
				return c.JSON(http.StatusOK, nil)
			})

			err := handler(c)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status %d for '%s', got %d", http.StatusOK, authHeader, rec.Code)
			}
		})
	}
}

func TestRateLimitByIP_NoRedis(t *testing.T) {
	// When redis is nil, middleware should pass through
	mw := middleware.RateLimitByIP(nil, 10, 0)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"success": true})
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["success"] != true {
		t.Error("Expected success response")
	}
}

