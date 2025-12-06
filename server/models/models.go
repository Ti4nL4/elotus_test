package models

import (
	"fmt"
	"os"

	"elotus_test/server/bredis"
	"elotus_test/server/bsql"
	"elotus_test/server/cmd"
	"elotus_test/server/env"
	"elotus_test/server/logger"
	"elotus_test/server/models/auth"
	"elotus_test/server/models/upload"
	"elotus_test/server/models/user"
	"elotus_test/server/psql"

	"gopkg.in/yaml.v3"
)

// Models holds all application components
type Models struct {
	db           *bsql.DB
	bredisClient *bredis.Client

	userStore     user.Repository
	uploadStore   upload.Repository
	jwtService    *auth.JWTService
	authHandler   *auth.Handler
	uploadHandler *upload.Handler
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
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

	// Connect to Redis (optional)
	m.bredisClient = m.initRedis()

	// Initialize repositories
	m.userStore = user.NewPostgresRepository(m.db)
	m.uploadStore = upload.NewPostgresRepository(m.db)
	logger.Info("Using PostgreSQL for storage")

	// Initialize JWT service with revocation store
	revocationStore := auth.NewTokenRevocationStore(m.db, m.bredisClient)
	jwtConfig := &auth.Config{
		SecretKey:     []byte(env.E.JWTSigningKey),
		TokenDuration: env.E.GetJWTDuration(),
	}
	m.jwtService = auth.NewJWTService(jwtConfig, revocationStore)

	// Initialize handlers
	m.authHandler = auth.NewHandler(m.userStore, m.jwtService, m.bredisClient)
	m.uploadHandler = upload.NewHandler(m.db, m.uploadStore, m.bredisClient)

	if !cmdMode {
		m.SetupRoutes()
	}

	return m
}

func (m *Models) initRedis() *bredis.Client {
	redisConfigPath := cmd.ResolvePath(env.E.RedisConfigFilePath)
	if redisConfigPath == "" {
		redisConfigPath = cmd.ResolvePath("db/redis.yaml")
	}

	data, err := os.ReadFile(redisConfigPath)
	if err != nil {
		logger.Warnf("Redis config not found, disabled")
		return nil
	}

	var cfg RedisConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		logger.Warnf("Invalid redis config: %v", err)
		return nil
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "6379"
	}

	logger.Info("Connecting to Redis...")
	logger.Infof("  Host: %s:%s", cfg.Host, cfg.Port)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	client := bredis.New(addr, cfg.Password, cfg.DB, env.E.ServerName)
	if client == nil {
		logger.Warnf("Failed to connect to Redis, disabled")
		return nil
	}

	logger.Info("Redis connected")
	return client
}

// RunCmd runs command mode
func (m *Models) RunCmd(c string) {
	switch c {
	default:
		logger.Warnf("Unknown command: %s", c)
	}
}
