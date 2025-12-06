package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"elotus_test/server/bsql"
	"elotus_test/server/env"
	"elotus_test/server/models"
	"elotus_test/server/psql"
	"elotus_test/server/renv"
)

var cmd = flag.String("cmd", "", "Command mode")
var db = flag.String("db", "", "Database command: migrate, rollback, generate, status")
var migrationName = flag.String("name", "", "Migration name (for generate)")
var steps = flag.Int("steps", 1, "Number of migrations to rollback")

func main() {
	flag.Parse()
	log.Println("Starting elotus_test...")

	// Parse environment configuration
	var envConfig *env.ENV
	renv.ParseCmd(&envConfig)
	envConfig.SetDefaults()
	env.E = envConfig

	log.Printf("Environment: %s", env.E.Environment)
	log.Printf("Server Name: %s", env.E.ServerName)

	// Handle database commands
	if *db != "" {
		handleDBCommand(*db)
		return
	}

	// Handle other commands
	if *cmd != "" {
		instance := models.NewModels(true)
		instance.RunCmd(*cmd)
		return
	}

	// Start server
	models.NewModels(false)
	select {}
}

func handleDBCommand(command string) {
	// Resolve database config path
	dbConfigPath := resolvePath(env.E.DatabaseConfigFilePath)

	// Load database config
	dbConfig, err := bsql.LoadDatabaseConfig(dbConfigPath)
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	// Connect to database
	database := bsql.Open(
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Database,
		dbConfig.MaxIdleConnection,
		dbConfig.MaxOpenConnection,
	)
	defer database.Close()

	// Resolve migrations path
	migPath := resolvePath("db/migrations")

	switch command {
	case "migrate":
		log.Println("Running migrations...")
		if err := psql.MigrateUp(database, migPath); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations completed successfully")

	case "rollback":
		log.Printf("Rolling back %d migration(s)...\n", *steps)
		if err := psql.MigrateDown(database, migPath, *steps); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("Rollback completed successfully")

	case "generate":
		if *migrationName == "" {
			fmt.Println("Usage: server -db generate -name \"migration name\"")
			os.Exit(1)
		}
		if err := psql.GenerateMigration(migPath, *migrationName); err != nil {
			log.Fatalf("Failed to generate migration: %v", err)
		}

	case "status":
		if err := psql.MigrationStatus(database, migPath); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}

	default:
		fmt.Println("Unknown database command:", command)
		fmt.Println("Available commands: migrate, rollback, generate, status")
		os.Exit(1)
	}
}

// resolvePath tries to find the path in current dir or server/ dir
func resolvePath(path string) string {
	// Try current directory first
	if _, err := os.Stat(path); err == nil {
		return path
	}
	// Try server/ prefix
	serverPath := "server/" + path
	if _, err := os.Stat(serverPath); err == nil {
		return serverPath
	}
	// Return original path
	return path
}
