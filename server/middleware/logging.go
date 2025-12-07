package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

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

					c.Error(echo.NewHTTPError(500, "Internal Server Error"))
				}
			}()
			return next(c)
		}
	}
}
