package middleware

import (
	"strconv"
	"time"

	"elotus_test/server/bredis"
	"elotus_test/server/response"

	"github.com/labstack/echo/v4"
)

func RateLimitByIP(redis *bredis.Client, limit int64, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if redis == nil {
				return next(c)
			}

			key := "ip:" + c.RealIP()
			result := redis.CheckRateLimit(key, limit, window)

			c.Response().Header().Set("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))

			if !result.Allowed {
				c.Response().Header().Set("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
				return response.TooManyRequests(c, "Too many requests", result.RetryAfter.Seconds())
			}

			return next(c)
		}
	}
}
