package models

import (
	"fmt"
	"log"
	"net/http"

	"elotus_test/server/cmd"
	"elotus_test/server/env"
	"elotus_test/server/models/auth"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupRoutes configures and starts the HTTP server
func (m *Models) SetupRoutes() {
	e := echo.New()
	e.HideBanner = true

	// Global middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Public routes - define these BEFORE static files
	e.GET("/health", m.authHandler.HealthCheck)
	e.POST("/register", m.authHandler.Register)
	e.POST("/login", m.authHandler.Login)
	e.GET("/upload-form", m.uploadHandler.UploadForm) // Simple HTML form for testing

	// Config endpoint for HTML to get API settings
	e.GET("/config.js", configHandler)

	// Serve uploaded files from tmp folder (use /media to avoid conflict with uploads.html)
	mediaPath := cmd.ResolvePath("tmp")
	e.Static("/media", mediaPath)

	// Protected routes (require authentication)
	protected := e.Group("/api")
	protected.Use(auth.JWTMiddleware(m.jwtService))
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
	log.Printf("Server starting on %s...", serverAddr)
	log.Println("Available endpoints:")
	log.Println("  GET  /              - Web UI for testing")
	log.Println("  POST /register      - Register a new user")
	log.Println("  POST /login         - Login and get JWT token")
	log.Println("  POST /api/revoke    - Revoke tokens (requires auth)")
	log.Println("  GET  /api/protected - Protected endpoint (requires auth)")
	log.Println("  POST /api/upload    - Upload image file (requires auth, max 8MB)")
	log.Println("  GET  /api/uploads   - Get all uploads for user (requires auth)")
	log.Println("  GET  /api/uploads/:id - Get specific upload (requires auth)")
	log.Println("  GET  /health        - Health check")

	go func() {
		if err := e.Start(serverAddr); err != nil {
			log.Printf("Server stopped: %v", err)
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
