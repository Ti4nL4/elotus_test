package auth

import (
	"time"

	"elotus_test/server/bredis"
	"elotus_test/server/bsql"
	"elotus_test/server/models/user"
	"elotus_test/server/response"
	"elotus_test/server/validation"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db         *bsql.DB
	userRepo   user.Repository
	jwtService *JWTService
	redis      *bredis.Client
}

func NewHandler(db *bsql.DB, userRepo user.Repository, jwtService *JWTService, redis *bredis.Client) *Handler {
	return &Handler{
		db:         db,
		userRepo:   userRepo,
		jwtService: jwtService,
		redis:      redis,
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RevokeRequest struct {
	RevokeBeforeTime *time.Time `json:"revoke_before_time,omitempty"`
}

const (
	loginRateLimitMax    = 5
	loginRateLimitWindow = 15 * time.Minute
)

func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if valid, msg := validation.ValidateUsername(req.Username); !valid {
		return response.ValidationError(c, msg)
	}

	if valid, msg := validation.ValidatePassword(req.Password); !valid {
		return response.ValidationError(c, msg)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.InternalError(c, "Failed to process password")
	}

	u, err := h.userRepo.CreateUser(req.Username, string(hashedPassword))
	if err != nil {
		if err == user.ErrUserExists {
			return response.Conflict(c, "Username already exists")
		}
		return response.InternalError(c, "Failed to create user")
	}

	return response.Created(c, echo.Map{
		"message": "User registered successfully",
		"user": echo.Map{
			"id":         u.ID,
			"username":   u.Username,
			"created_at": u.CreatedAt,
		},
	})
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if req.Username == "" || req.Password == "" {
		return response.ValidationError(c, "Username and password are required")
	}

	if h.redis != nil {
		result := h.redis.CheckRateLimit("login:user:"+req.Username, loginRateLimitMax, loginRateLimitWindow)
		if !result.Allowed {
			return response.TooManyRequests(c, "Too many login attempts for this account", result.RetryAfter.Seconds())
		}
	}

	u, exists := h.userRepo.GetUserByUsername(req.Username)
	if !exists {
		return response.Unauthorized(c, "Invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return response.Unauthorized(c, "Invalid username or password")
	}

	token, expiresAt, err := h.jwtService.GenerateToken(u.ID, u.Username)
	if err != nil {
		return response.InternalError(c, "Failed to generate token")
	}

	if h.redis != nil {
		h.redis.ResetRateLimit("login:user:" + req.Username)
	}

	_ = h.userRepo.UpdateLastLogin(u.ID)

	return response.Success(c, echo.Map{
		"token":      token,
		"expires_at": expiresAt,
	})
}

func (h *Handler) RevokeToken(c echo.Context) error {
	claims := c.Get("user").(*TokenClaims)

	var req RevokeRequest
	if err := c.Bind(&req); err != nil {
		if err := h.jwtService.RevokeUserTokens(claims.UserID); err != nil {
			return response.InternalError(c, "Failed to revoke tokens")
		}
		return response.Success(c, echo.Map{
			"message": "All tokens have been revoked",
		})
	}

	if req.RevokeBeforeTime != nil {
		if err := h.jwtService.RevokeUserTokensBefore(claims.UserID, *req.RevokeBeforeTime); err != nil {
			return response.InternalError(c, "Failed to revoke tokens")
		}
		return response.Success(c, echo.Map{
			"message": "Tokens issued before " + req.RevokeBeforeTime.Format(time.RFC3339) + " have been revoked",
		})
	}

	if err := h.jwtService.RevokeUserTokens(claims.UserID); err != nil {
		return response.InternalError(c, "Failed to revoke tokens")
	}
	return response.Success(c, echo.Map{
		"message": "All tokens have been revoked",
	})
}

func (h *Handler) Protected(c echo.Context) error {
	claims := c.Get("user").(*TokenClaims)

	return response.Success(c, echo.Map{
		"message":  "Welcome to the protected endpoint!",
		"user_id":  claims.UserID,
		"username": claims.Username,
	})
}

func (h *Handler) HealthCheck(c echo.Context) error {
	health := echo.Map{
		"status": "UP",
		"db":     "OK",
		"redis":  "OK",
	}

	overallStatus := "UP"

	if h.db != nil {
		if err := h.db.Ping(); err != nil {
			health["db"] = "DOWN"
			health["db_error"] = err.Error()
			overallStatus = "DEGRADED"
		}
	} else {
		health["db"] = "NOT_CONFIGURED"
		overallStatus = "DEGRADED"
	}

	if h.redis != nil {
		ctx := c.Request().Context()
		if err := h.redis.Ping(ctx).Err(); err != nil {
			health["redis"] = "DOWN"
			health["redis_error"] = err.Error()
			if overallStatus == "UP" {
				overallStatus = "DEGRADED"
			}
		}
	} else {
		health["redis"] = "DISABLED"
	}

	health["status"] = overallStatus
	health["timestamp"] = time.Now().Format(time.RFC3339)

	return response.Success(c, health)
}
