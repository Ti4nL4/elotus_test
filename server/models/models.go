package models

import (
	"elotus_test/server/bsql"
	"elotus_test/server/cmd"
	"elotus_test/server/env"
	"elotus_test/server/logger"
	"elotus_test/server/models/auth"
	"elotus_test/server/models/upload"
	"elotus_test/server/models/user"
	"elotus_test/server/psql"
)

// Models holds all application components
type Models struct {
	db              *bsql.DB
	userStore       user.Repository
	uploadStore     upload.Repository
	revocationStore *auth.TokenRevocationStore
	jwtService      *auth.JWTService
	authHandler     *auth.Handler
	uploadHandler   *upload.Handler
}

// NewModels creates and initializes all application components
func NewModels(cmdMode bool) *Models {
	m := &Models{}

	// Database is required
	logger.Info("Connecting to PostgreSQL...")

	dbConfigPath := cmd.ResolvePath(env.E.DatabaseConfigFilePath)
	dbConfig, err := bsql.LoadDatabaseConfig(dbConfigPath)
	if err != nil {
		logger.Fatalf("Failed to load database config: %v", err)
	}

	logger.Infof("  Host: %s:%s", dbConfig.Host, dbConfig.Port)
	logger.Infof("  Database: %s", dbConfig.Database)
	logger.Infof("  User: %s", dbConfig.Username)

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
	logger.Info("Running database migrations...")
	migPath := cmd.ResolvePath("db/migrations")
	if err := psql.MigrateUp(m.db, migPath); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	// Use PostgreSQL repository
	m.userStore = user.NewPostgresRepository(m.db)
	logger.Info("Using PostgreSQL for user storage")

	// Initialize upload repository
	m.uploadStore = upload.NewPostgresRepository(m.db)
	logger.Info("Using PostgreSQL for file upload storage")

	// Initialize token revocation store (DB-based)
	m.revocationStore = auth.NewTokenRevocationStore(m.db)

	// Initialize JWT service
	jwtConfig := &auth.Config{
		SecretKey:     []byte(env.E.JWTSigningKey),
		TokenDuration: env.E.GetJWTDuration(),
	}
	m.jwtService = auth.NewJWTService(jwtConfig, m.revocationStore)

	// Initialize auth handler
	m.authHandler = auth.NewHandler(m.userStore, m.jwtService)

	// Initialize upload handler
	m.uploadHandler = upload.NewHandler(m.db, m.uploadStore)

	if !cmdMode {
		m.SetupRoutes()
	}

	return m
}

// RunCmd runs command mode
func (m *Models) RunCmd(c string) {
	switch c {
	default:
		logger.Warnf("Unknown command: %s", c)
	}
}
