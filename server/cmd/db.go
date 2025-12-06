package cmd

import (
	"fmt"
	"log"
	"os"

	"elotus_test/server/bsql"
	"elotus_test/server/env"
	"elotus_test/server/psql"
)

// HandleDB handles database-related commands
func HandleDB(command string, migrationName string, steps int) {
	// Resolve database config path
	dbConfigPath := ResolvePath(env.E.DatabaseConfigFilePath)

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
	migPath := ResolvePath("db/migrations")

	switch command {
	case "migrate":
		log.Println("Running migrations...")
		if err := psql.MigrateUp(database, migPath); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations completed successfully")

	case "rollback":
		log.Printf("Rolling back %d migration(s)...\n", steps)
		if err := psql.MigrateDown(database, migPath, steps); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("Rollback completed successfully")

	case "generate":
		if migrationName == "" {
			fmt.Println("Usage: server -db generate -name \"migration name\"")
			os.Exit(1)
		}
		if err := psql.GenerateMigration(migPath, migrationName); err != nil {
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

// ResolvePath tries to find the path in current dir or server/ dir
func ResolvePath(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}
	serverPath := "server/" + path
	if _, err := os.Stat(serverPath); err == nil {
		return serverPath
	}
	return path
}
