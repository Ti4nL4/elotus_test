package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// RequestLogger is a middleware that logs HTTP requests using zerolog
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Calculate latency
			latency := time.Since(start)

			// Get request info
			req := c.Request()
			res := c.Response()

			// Build log event
			event := log.Info()
			if res.Status >= 500 {
				event = log.Error()
			} else if res.Status >= 400 {
				event = log.Warn()
			}

			// Log the request
			event.
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Str("query", req.URL.RawQuery).
				Int("status", res.Status).
				Dur("latency", latency).
				Str("ip", c.RealIP()).
				Str("user_agent", req.UserAgent()).
				Int64("bytes_out", res.Size).
				Msg("HTTP Request")

			return err
		}
	}
}

// RequestLoggerWithSkipper is a request logger with custom skipper
func RequestLoggerWithSkipper(skipper func(c echo.Context) bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper != nil && skipper(c) {
				return next(c)
			}

			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			req := c.Request()
			res := c.Response()

			event := log.Info()
			if res.Status >= 500 {
				event = log.Error()
			} else if res.Status >= 400 {
				event = log.Warn()
			}

			event.
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", res.Status).
				Dur("latency", latency).
				Str("ip", c.RealIP()).
				Msg("HTTP Request")

			return err
		}
	}
}

// RecoverWithLogger is a recovery middleware that logs panics
func RecoverWithLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					req := c.Request()
					log.Error().
						Interface("panic", r).
						Str("method", req.Method).
						Str("path", req.URL.Path).
						Str("ip", c.RealIP()).
						Msg("Panic recovered")

					// Return 500 error
					c.Error(echo.NewHTTPError(500, "Internal Server Error"))
				}
			}()
			return next(c)
		}
	}
}
