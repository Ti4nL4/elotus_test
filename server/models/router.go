package models

import (
	"log"

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

	// Public routes
	e.GET("/health", m.authHandler.HealthCheck)
	e.POST("/register", m.authHandler.Register)
	e.POST("/login", m.authHandler.Login)

	// Protected routes (require authentication)
	protected := e.Group("")
	protected.Use(auth.JWTMiddleware(m.jwtService))
	{
		protected.POST("/revoke", m.authHandler.RevokeToken)
		protected.GET("/protected", m.authHandler.Protected)
	}

	// Start server
	serverAddr := ":" + env.E.GetServerPort()
	log.Printf("Server starting on %s...", serverAddr)
	log.Println("Available endpoints:")
	log.Println("  POST /register   - Register a new user (username, password)")
	log.Println("  POST /login      - Login and get JWT token")
	log.Println("  POST /revoke     - Revoke tokens (requires auth)")
	log.Println("  GET  /protected  - Protected endpoint (requires auth)")
	log.Println("  GET  /health     - Health check")

	go func() {
		if err := e.Start(serverAddr); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()
}
