package models

import (
	"log"
	"net/http"
	"os"

	"elotus_test/server/bsql"
	"elotus_test/server/env"
	"elotus_test/server/models/auth"
	"elotus_test/server/models/user"
	"elotus_test/server/psql"
)

// Models holds all application components
type Models struct {
	db              *bsql.DB
	userStore       user.Repository
	revocationStore *auth.TokenRevocationStore
	jwtService      *auth.JWTService
	authHandler     *auth.Handler
}

// M is the global models instance
var M *Models

// NewModels creates and initializes all application components
func NewModels(cmdMode bool) *Models {
	m := &Models{}

	// Initialize database if postgres is enabled
	if env.E.UsePostgres {
		log.Println("Connecting to PostgreSQL...")
		dbConfig, err := bsql.LoadDatabaseConfig(env.E.DatabaseConfigFilePath)
		if err != nil {
			log.Fatalf("Failed to load database config: %v", err)
		}

		log.Printf("  Host: %s:%s", dbConfig.Host, dbConfig.Port)
		log.Printf("  Database: %s", dbConfig.Database)
		log.Printf("  User: %s", dbConfig.Username)

		m.db = bsql.Open(
			dbConfig.Username,
			dbConfig.Password,
			dbConfig.Host,
			dbConfig.Port,
			dbConfig.Database,
			dbConfig.MaxIdleConnection,
			dbConfig.MaxOpenConnection,
		)

		// Run migrations using psql package
		log.Println("Running database migrations...")
		migPath := resolveMigrationsPath()
		if err := psql.MigrateUp(m.db, migPath); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		// Use PostgreSQL repository
		m.userStore = user.NewPostgresRepository(m.db)
		log.Println("Using PostgreSQL for user storage")
	} else {
		// Use in-memory storage
		m.userStore = user.NewMemoryRepository()
		log.Println("Using in-memory storage (data will be lost on restart)")
	}

	// Initialize token revocation store
	m.revocationStore = auth.NewTokenRevocationStore()

	// Initialize JWT service
	jwtConfig := &auth.Config{
		SecretKey:     []byte(env.E.JWTSigningKey),
		TokenDuration: env.E.GetJWTDuration(),
	}
	m.jwtService = auth.NewJWTService(jwtConfig, m.revocationStore)

	// Initialize auth handler
	m.authHandler = auth.NewHandler(m.userStore, m.jwtService)

	M = m

	if !cmdMode {
		m.startServer()
	}

	return m
}

// startServer starts the HTTP server
func (m *Models) startServer() {
	// Public routes
	http.HandleFunc("/health", m.authHandler.HealthCheck)
	http.HandleFunc("/register", m.authHandler.Register)
	http.HandleFunc("/login", m.authHandler.Login)

	// Protected routes (require authentication)
	http.HandleFunc("/revoke", auth.AuthMiddlewareFunc(m.jwtService, m.authHandler.RevokeToken))
	http.HandleFunc("/protected", auth.AuthMiddlewareFunc(m.jwtService, m.authHandler.Protected))

	serverAddr := ":" + env.E.GetServerPort()
	log.Printf("Server starting on %s...", serverAddr)
	log.Println("Available endpoints:")
	log.Println("  POST /register   - Register a new user (username, password)")
	log.Println("  POST /login      - Login and get JWT token")
	log.Println("  POST /revoke     - Revoke tokens (requires auth)")
	log.Println("  GET  /protected  - Protected endpoint (requires auth)")
	log.Println("  GET  /health     - Health check")

	go func() {
		if err := http.ListenAndServe(serverAddr, nil); err != nil {
			log.Fatal(err)
		}
	}()
}

// RunCmd runs command mode
func (m *Models) RunCmd(cmd string) {
	switch cmd {
	default:
		log.Printf("Unknown command: %s", cmd)
	}
}

// Close closes all resources
func (m *Models) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

// resolveMigrationsPath finds the migrations directory
func resolveMigrationsPath() string {
	paths := []string{
		"db/migrations",
		"server/db/migrations",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "db/migrations"
}
