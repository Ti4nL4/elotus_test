package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ValidateTokenFunc is the function signature for token validation
type ValidateTokenFunc func(tokenString string) (claims interface{}, err error)

// JWTMiddleware creates an Echo middleware that validates JWT tokens
func JWTMiddleware(validateFn ValidateTokenFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Authorization header required"})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid authorization header format"})
			}

			claims, err := validateFn(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid or expired token: " + err.Error()})
			}

			c.Set("user", claims)
			return next(c)
		}
	}
}
