package middleware

import (
	"net/http"
	"strconv"
	"time"

	"elotus_test/server/bredis"

	"github.com/labstack/echo/v4"
)

// RateLimitByIP creates a rate limit middleware by IP address
func RateLimitByIP(redis *bredis.Client, limit int64, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip if no redis
			if redis == nil {
				return next(c)
			}

			key := "ip:" + c.RealIP()
			result := redis.CheckRateLimit(key, limit, window)

			// Set rate limit headers
			c.Response().Header().Set("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))

			if !result.Allowed {
				c.Response().Header().Set("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
				return c.JSON(http.StatusTooManyRequests, echo.Map{
					"error":       "Too many requests",
					"retry_after": result.RetryAfter.Seconds(),
				})
			}

			return next(c)
		}
	}
}
