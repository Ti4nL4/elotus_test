package models

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"elotus_test/server/cmd"
	"elotus_test/server/env"
	"elotus_test/server/logger"
	custommiddleware "elotus_test/server/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (m *Models) SetupRoutes() {
	e := echo.New()
	e.HideBanner = true
	m.echo = e

	e.Use(custommiddleware.RequestLoggerWithSkipper(func(c echo.Context) bool {
		path := c.Request().URL.Path
		return strings.HasPrefix(path, "/media/") ||
			strings.HasSuffix(path, ".css") ||
			strings.HasSuffix(path, ".js") ||
			strings.HasSuffix(path, ".html")
	}))
	e.Use(custommiddleware.RecoverWithLogger())
	e.Use(middleware.CORS())

	authRateLimit := custommiddleware.RateLimitByIP(m.bredisClient, 10, time.Minute)

	jwtMiddleware := custommiddleware.JWTMiddleware(func(token string) (interface{}, error) {
		return m.jwtService.ValidateToken(token)
	})

	e.GET("/health", m.authHandler.HealthCheck)
	e.POST("/register", m.authHandler.Register, authRateLimit)
	e.POST("/login", m.authHandler.Login, authRateLimit)

	e.GET("/config.js", configHandler)

	mediaPath := cmd.ResolvePath("tmp")
	e.Static("/media", mediaPath)

	e.POST("/upload", m.uploadHandler.Upload, jwtMiddleware)

	protected := e.Group("/api")
	protected.Use(jwtMiddleware)
	{
		protected.POST("/revoke", m.authHandler.RevokeToken)
		protected.GET("/protected", m.authHandler.Protected)
		protected.POST("/upload", m.uploadHandler.Upload)
		protected.GET("/uploads", m.uploadHandler.GetUserUploads)
		protected.GET("/uploads/:id", m.uploadHandler.GetUploadByID)
	}

	htmlPath := cmd.ResolvePath("html")
	e.Static("/", htmlPath)

	serverAddr := ":" + env.E.GetServerPort()
	logger.Infof("Server starting on %s...", serverAddr)
	logger.Info("Available endpoints:")
	logger.Info("  GET  /              - Web UI for testing")
	logger.Info("  POST /register      - Register a new user")
	logger.Info("  POST /login         - Login and get JWT token")
	logger.Info("  POST /upload        - Upload image (requires auth, field: 'data')")
	logger.Info("  POST /api/revoke    - Revoke tokens (requires auth)")
	logger.Info("  GET  /api/protected - Protected endpoint (requires auth)")
	logger.Info("  POST /api/upload    - Upload image file (requires auth, max 8MB)")
	logger.Info("  GET  /api/uploads   - Get all uploads for user (requires auth)")
	logger.Info("  GET  /api/uploads/:id - Get specific upload (requires auth)")
	logger.Info("  GET  /health        - Health check")

	go func() {
		if err := e.Start(serverAddr); err != nil && err.Error() != "http: Server closed" {
			logger.Errorf("Server stopped: %v", err)
		}
	}()
}

func configHandler(c echo.Context) error {
	apiBase := env.E.GetAPIBaseURL()

	js := fmt.Sprintf(`const CONFIG = {
    API_BASE: "%s",
    SERVER_NAME: "%s",
    ENVIRONMENT: "%s"
};
`, apiBase, env.E.ServerName, env.E.Environment)

	return c.Blob(http.StatusOK, "application/javascript", []byte(js))
}
