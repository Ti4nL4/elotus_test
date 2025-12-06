package auth

import (
	"net/http"
	"time"

	"elotus_test/server/bredis"
	"elotus_test/server/models/user"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Handler handles authentication-related requests
type Handler struct {
	userRepo   user.Repository
	jwtService *JWTService
	redis      *bredis.Client
}

// NewHandler creates a new Handler
func NewHandler(userRepo user.Repository, jwtService *JWTService, redis *bredis.Client) *Handler {
	return &Handler{
		userRepo:   userRepo,
		jwtService: jwtService,
		redis:      redis,
	}
}

// RegisterRequest represents the request body for registration
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents the request body for login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RevokeRequest represents the request body for token revocation
type RevokeRequest struct {
	RevokeBeforeTime *time.Time `json:"revoke_before_time,omitempty"`
}

// Rate limit config
const (
	loginRateLimitMax    = 5
	loginRateLimitWindow = 15 * time.Minute
)

// Register handles user registration
func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Username and password are required"})
	}

	if len(req.Password) < 6 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Password must be at least 6 characters"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to process password"})
	}

	u, err := h.userRepo.CreateUser(req.Username, string(hashedPassword))
	if err != nil {
		if err == user.ErrUserExists {
			return c.JSON(http.StatusConflict, echo.Map{"error": "Username already exists"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"message": "User registered successfully",
		"user": echo.Map{
			"id":         u.ID,
			"username":   u.Username,
			"created_at": u.CreatedAt,
		},
	})
}

// Login handles user login
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Username and password are required"})
	}

	// Check rate limit by username (IP rate limit is handled by middleware)
	if h.redis != nil {
		result := h.redis.CheckRateLimit("login:user:"+req.Username, loginRateLimitMax, loginRateLimitWindow)
		if !result.Allowed {
			return c.JSON(http.StatusTooManyRequests, echo.Map{
				"error":       "Too many login attempts for this account.",
				"retry_after": result.RetryAfter.Seconds(),
			})
		}
	}

	u, exists := h.userRepo.GetUserByUsername(req.Username)
	if !exists {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
	}

	token, expiresAt, err := h.jwtService.GenerateToken(u.ID, u.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to generate token"})
	}

	// Reset rate limit on success
	if h.redis != nil {
		h.redis.ResetRateLimit("login:user:" + req.Username)
	}

	// Update last login time
	_ = h.userRepo.UpdateLastLogin(u.ID)

	return c.JSON(http.StatusOK, LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

// RevokeToken handles token revocation by time
func (h *Handler) RevokeToken(c echo.Context) error {
	claims := c.Get("user").(*TokenClaims)

	var req RevokeRequest
	if err := c.Bind(&req); err != nil {
		// If no body, revoke all current tokens for the user
		if err := h.jwtService.RevokeUserTokens(claims.UserID); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to revoke tokens"})
		}
		return c.JSON(http.StatusOK, echo.Map{
			"message": "All tokens have been revoked",
		})
	}

	if req.RevokeBeforeTime != nil {
		if err := h.jwtService.RevokeUserTokensBefore(claims.UserID, *req.RevokeBeforeTime); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to revoke tokens"})
		}
		return c.JSON(http.StatusOK, echo.Map{
			"message": "Tokens issued before " + req.RevokeBeforeTime.Format(time.RFC3339) + " have been revoked",
		})
	}

	if err := h.jwtService.RevokeUserTokens(claims.UserID); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to revoke tokens"})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "All tokens have been revoked",
	})
}

// Protected is a sample protected endpoint
func (h *Handler) Protected(c echo.Context) error {
	claims := c.Get("user").(*TokenClaims)

	return c.JSON(http.StatusOK, echo.Map{
		"message":  "Welcome to the protected endpoint!",
		"user_id":  claims.UserID,
		"username": claims.Username,
	})
}

// HealthCheck handler
func (h *Handler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
}
