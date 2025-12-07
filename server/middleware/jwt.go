package middleware

import (
	"strings"

	"elotus_test/server/response"

	"github.com/labstack/echo/v4"
)

type ValidateTokenFunc func(tokenString string) (claims interface{}, err error)

func JWTMiddleware(validateFn ValidateTokenFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return response.Unauthorized(c, "Authorization header required")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return response.Unauthorized(c, "Invalid authorization header format")
			}

			claims, err := validateFn(parts[1])
			if err != nil {
				return response.Unauthorized(c, "Invalid or expired token: "+err.Error())
			}

			c.Set("user", claims)
			return next(c)
		}
	}
}
