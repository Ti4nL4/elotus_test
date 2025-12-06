package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// ContextKey is a custom type for context keys
type ContextKey string

const (
	UserContextKey ContextKey = "user"
)

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]

			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "Invalid or expired token: "+err.Error())
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthMiddlewareFunc is a helper for using with http.HandlerFunc
func AuthMiddlewareFunc(jwtService *JWTService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token: "+err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
