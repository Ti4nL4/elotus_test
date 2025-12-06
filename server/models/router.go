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

// SetupRoutes configures and starts the HTTP server
func (m *Models) SetupRoutes() {
	e := echo.New()
	e.HideBanner = true

	// Global middleware - use custom zerolog middleware
	e.Use(custommiddleware.RequestLoggerWithSkipper(func(c echo.Context) bool {
		// Skip logging for static files
		path := c.Request().URL.Path
		return strings.HasPrefix(path, "/media/") ||
			strings.HasSuffix(path, ".css") ||
			strings.HasSuffix(path, ".js") ||
			strings.HasSuffix(path, ".html")
	}))
	e.Use(custommiddleware.RecoverWithLogger())
	e.Use(middleware.CORS())

	// Rate limit middleware for auth endpoints
	authRateLimit := custommiddleware.RateLimitByIP(m.bredisClient, 10, time.Minute)

	// Public routes - define these BEFORE static files
	e.GET("/health", m.authHandler.HealthCheck)
	e.POST("/register", m.authHandler.Register, authRateLimit)
	e.POST("/login", m.authHandler.Login, authRateLimit)

	// Config endpoint for HTML to get API settings
	e.GET("/config.js", configHandler)

	// Serve uploaded files from tmp folder (use /media to avoid conflict with uploads.html)
	mediaPath := cmd.ResolvePath("tmp")
	e.Static("/media", mediaPath)

	// Protected routes (require authentication)
	protected := e.Group("/api")
	protected.Use(custommiddleware.JWTMiddleware(func(token string) (interface{}, error) {
		return m.jwtService.ValidateToken(token)
	}))
	{
		protected.POST("/revoke", m.authHandler.RevokeToken)
		protected.GET("/protected", m.authHandler.Protected)
		protected.POST("/upload", m.uploadHandler.Upload)
		protected.GET("/uploads", m.uploadHandler.GetUserUploads)
		protected.GET("/uploads/:id", m.uploadHandler.GetUploadByID)
	}

	// Serve static HTML files - LAST so it doesn't override API routes
	htmlPath := cmd.ResolvePath("html")
	e.Static("/", htmlPath)

	// Start server
	serverAddr := ":" + env.E.GetServerPort()
	logger.Infof("Server starting on %s...", serverAddr)
	logger.Info("Available endpoints:")
	logger.Info("  GET  /              - Web UI for testing")
	logger.Info("  POST /register      - Register a new user")
	logger.Info("  POST /login         - Login and get JWT token")
	logger.Info("  POST /api/revoke    - Revoke tokens (requires auth)")
	logger.Info("  GET  /api/protected - Protected endpoint (requires auth)")
	logger.Info("  POST /api/upload    - Upload image file (requires auth, max 8MB)")
	logger.Info("  GET  /api/uploads   - Get all uploads for user (requires auth)")
	logger.Info("  GET  /api/uploads/:id - Get specific upload (requires auth)")
	logger.Info("  GET  /health        - Health check")

	go func() {
		if err := e.Start(serverAddr); err != nil {
			logger.Errorf("Server stopped: %v", err)
		}
	}()
}

// configHandler returns JavaScript config for HTML pages
func configHandler(c echo.Context) error {
	// Get API base URL from env config
	apiBase := env.E.GetAPIBaseURL()

	// Return as JavaScript
	js := fmt.Sprintf(`// Auto-generated config from server
const CONFIG = {
    API_BASE: "%s",
    SERVER_NAME: "%s",
    ENVIRONMENT: "%s"
};
`, apiBase, env.E.ServerName, env.E.Environment)

	return c.Blob(http.StatusOK, "application/javascript", []byte(js))
}
