package models

import (
	"context"
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

	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

type Models struct {
	db           *bsql.DB
	bredisClient *bredis.Client
	echo         *echo.Echo

	userStore     user.Repository
	uploadStore   upload.Repository
	jwtService    *auth.JWTService
	authHandler   *auth.Handler
	uploadHandler *upload.Handler
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func NewModels(cmdMode bool) *Models {
	m := &Models{}

	logger.Info("════════════════════════════════════════════════════════")
	logger.Info("🚀 Starting Server Initialization...")
	logger.Info("════════════════════════════════════════════════════════")

	logger.Info("")
	logger.Info("🐘 Connecting to PostgreSQL...")

	dbConfigPath := cmd.ResolvePath(env.E.DatabaseConfigFilePath)
	dbConfig, err := bsql.LoadDatabaseConfig(dbConfigPath)
	if err != nil {
		logger.Fatalf("Failed to load database config: %v", err)
	}

	logger.Infof("   Host: %s:%s", dbConfig.Host, dbConfig.Port)
	logger.Infof("   Database: %s", dbConfig.Database)
	logger.Infof("   User: %s", dbConfig.Username)

	m.db = bsql.Open(
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Database,
		dbConfig.MaxIdleConnection,
		dbConfig.MaxOpenConnection,
	)
	logger.Info("✅ PostgreSQL connected!")

	logger.Info("")
	logger.Info("📦 Running database migrations...")
	migPath := cmd.ResolvePath("db/migrations")
	if err := psql.MigrateUp(m.db, migPath); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}
	logger.Info("✅ Migrations completed!")

	logger.Info("")
	m.bredisClient = m.initRedis()

	logger.Info("")
	logger.Info("📂 Initializing repositories...")
	m.userStore = user.NewPostgresRepository(m.db)
	m.uploadStore = upload.NewPostgresRepository(m.db)
	logger.Info("✅ Repositories initialized!")

	logger.Info("")
	logger.Info("🔐 Initializing JWT service...")
	revocationStore := auth.NewTokenRevocationStore(m.db, m.bredisClient)
	jwtConfig := &auth.Config{
		SecretKey:     []byte(env.E.JWTSigningKey),
		TokenDuration: env.E.GetJWTDuration(),
	}
	m.jwtService = auth.NewJWTService(jwtConfig, revocationStore)
	logger.Infof("   Token Duration: %v", env.E.GetJWTDuration())
	logger.Info("✅ JWT service initialized!")

	logger.Info("")
	logger.Info("🎯 Initializing handlers...")
	m.authHandler = auth.NewHandler(m.db, m.userStore, m.jwtService, m.bredisClient)
	m.uploadHandler = upload.NewHandler(m.db, m.uploadStore, m.bredisClient)
	logger.Info("✅ Handlers initialized!")

	logger.Info("")
	logger.Info("════════════════════════════════════════════════════════")
	logger.Info("✅ Server initialization completed!")
	logger.Info("════════════════════════════════════════════════════════")

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
		logger.Warnf("⚠️  Redis config not found at %s, Redis disabled", redisConfigPath)
		return nil
	}

	var cfg RedisConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		logger.Warnf("⚠️  Invalid redis config: %v", err)
		return nil
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "6379"
	}

	logger.Info("🔴 Connecting to Redis...")
	logger.Infof("   Host: %s:%s", cfg.Host, cfg.Port)
	logger.Infof("   DB: %d", cfg.DB)
	logger.Infof("   Key Prefix: %s", env.E.ServerName)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	client := bredis.New(addr, cfg.Password, cfg.DB, env.E.ServerName)
	if client == nil {
		logger.Warnf("❌ Failed to connect to Redis, Redis disabled")
		return nil
	}

	testKey := "startup_test"
	testValue := "connected"
	if err := client.Set(testKey, testValue, 5*60*1000000000); err == nil {
		var result string
		if client.Get(testKey, &result) == nil && result == testValue {
			logger.Info("✅ Redis connected and working!")
			logger.Info("   Features enabled:")
			logger.Info("   • Rate Limiting (IP & Login)")
			logger.Info("   • Token Revocation Cache")
			logger.Info("   • Uploads List Cache")
			client.Delete(testKey)
		}
	}

	return client
}

func (m *Models) RunCmd(c string) {
	switch c {
	default:
		logger.Warnf("Unknown command: %s", c)
	}
}

func (m *Models) Shutdown(ctx context.Context) error {
	logger.Info("Closing connections...")

	if m.echo != nil {
		if err := m.echo.Shutdown(ctx); err != nil {
			logger.Errorf("Error shutting down HTTP server: %v", err)
		}
		logger.Info("✅ HTTP server stopped")
	}

	if m.bredisClient != nil {
		if err := m.bredisClient.Close(); err != nil {
			logger.Errorf("Error closing Redis: %v", err)
		}
		logger.Info("✅ Redis connection closed")
	}

	if m.db != nil {
		if err := m.db.Close(); err != nil {
			logger.Errorf("Error closing database: %v", err)
		}
		logger.Info("✅ Database connection closed")
	}

	return nil
}
