package models

import (
	"log"

	"elotus_test/server/bsql"
	"elotus_test/server/cmd"
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

// NewModels creates and initializes all application components
func NewModels(cmdMode bool) *Models {
	m := &Models{}

	// Database is required
	log.Println("Connecting to PostgreSQL...")

	dbConfigPath := cmd.ResolvePath(env.E.DatabaseConfigFilePath)
	dbConfig, err := bsql.LoadDatabaseConfig(dbConfigPath)
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

	// Run migrations
	log.Println("Running database migrations...")
	migPath := cmd.ResolvePath("db/migrations")
	if err := psql.MigrateUp(m.db, migPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Use PostgreSQL repository
	m.userStore = user.NewPostgresRepository(m.db)
	log.Println("Using PostgreSQL for user storage")

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

	if !cmdMode {
		m.SetupRoutes()
	}

	return m
}

// RunCmd runs command mode
func (m *Models) RunCmd(c string) {
	switch c {
	default:
		log.Printf("Unknown command: %s", c)
	}
}
