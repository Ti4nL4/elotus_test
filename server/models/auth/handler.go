package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"elotus_test/server/models/user"

	"golang.org/x/crypto/bcrypt"
)

// Handler handles authentication-related requests
type Handler struct {
	userRepo   user.Repository
	jwtService *JWTService
}

// NewHandler creates a new Handler
func NewHandler(userRepo user.Repository, jwtService *JWTService) *Handler {
	return &Handler{
		userRepo:   userRepo,
		jwtService: jwtService,
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

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	if len(req.Password) < 6 {
		respondWithError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	u, err := h.userRepo.CreateUser(req.Username, string(hashedPassword))
	if err != nil {
		if err == user.ErrUserExists {
			respondWithError(w, http.StatusConflict, "Username already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id":         u.ID,
			"username":   u.Username,
			"created_at": u.CreatedAt,
		},
	})
}

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	u, exists := h.userRepo.GetUserByUsername(req.Username)
	if !exists {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	token, expiresAt, err := h.jwtService.GenerateToken(u.ID, u.Username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondWithJSON(w, http.StatusOK, LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

// RevokeToken handles token revocation by time
func (h *Handler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims, ok := r.Context().Value(UserContextKey).(*TokenClaims)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req RevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jwtService.RevokeUserTokens(claims.UserID)
		respondWithJSON(w, http.StatusOK, map[string]string{
			"message": "All tokens have been revoked",
		})
		return
	}

	if req.RevokeBeforeTime != nil {
		h.jwtService.RevokeUserTokensBefore(claims.UserID, *req.RevokeBeforeTime)
		respondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Tokens issued before " + req.RevokeBeforeTime.Format(time.RFC3339) + " have been revoked",
		})
		return
	}

	h.jwtService.RevokeUserTokens(claims.UserID)
	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "All tokens have been revoked",
	})
}

// Protected is a sample protected endpoint
func (h *Handler) Protected(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserContextKey).(*TokenClaims)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Welcome to the protected endpoint!",
		"user_id":  claims.UserID,
		"username": claims.Username,
	})
}

// HealthCheck handler
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
