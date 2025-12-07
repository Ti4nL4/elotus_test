package psql

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"elotus_test/server/bsql"
)

type Migration struct {
	Version   string
	Name      string
	UpSQL     string
	DownSQL   string
	AppliedAt *time.Time
}

func MigrateUp(db *bsql.DB, migrationsPath string) error {
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	migrations, err := loadMigrations(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	for _, m := range migrations {
		if _, ok := applied[m.Version]; ok {
			continue
		}

		fmt.Printf("Applying migration %s: %s\n", m.Version, m.Name)

		if _, err := db.Exec(m.UpSQL); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", m.Version, err)
		}

		if err := recordMigration(db, m.Version, m.Name); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", m.Version, err)
		}

		fmt.Printf("Migration %s applied successfully\n", m.Version)
	}

	return nil
}

func MigrateDown(db *bsql.DB, migrationsPath string, steps int) error {
	applied, err := getAppliedMigrationsOrdered(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		fmt.Println("No migrations to rollback")
		return nil
	}

	migrations, err := loadMigrations(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	migrationMap := make(map[string]*Migration)
	for _, m := range migrations {
		migrationMap[m.Version] = m
	}

	count := 0
	for i := len(applied) - 1; i >= 0 && count < steps; i-- {
		version := applied[i]
		m, ok := migrationMap[version]
		if !ok {
			fmt.Printf("Warning: Migration file for %s not found, skipping\n", version)
			continue
		}

		if m.DownSQL == "" {
			fmt.Printf("Warning: No down migration for %s, skipping\n", version)
			continue
		}

		fmt.Printf("Rolling back migration %s: %s\n", m.Version, m.Name)

		if _, err := db.Exec(m.DownSQL); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", m.Version, err)
		}

		if err := removeMigration(db, m.Version); err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", m.Version, err)
		}

		fmt.Printf("Migration %s rolled back successfully\n", m.Version)
		count++
	}

	return nil
}

func GenerateMigration(migrationsPath, name string) error {
	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	version := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s.sql", version, strings.ToLower(strings.ReplaceAll(name, " ", "_")))
	filePath := filepath.Join(migrationsPath, fileName)

	content := fmt.Sprintf(`-- Migration: %s
-- Created at: %s

-- +migrate Up


-- +migrate Down

`, name, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	fmt.Printf("Created migration: %s\n", filePath)
	return nil
}

func MigrationStatus(db *bsql.DB, migrationsPath string) error {
	if err := createMigrationsTable(db); err != nil {
		return err
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}

	migrations, err := loadMigrations(migrationsPath)
	if err != nil {
		return err
	}

	fmt.Println("Migration Status:")
	fmt.Println("=================")
	for _, m := range migrations {
		status := "Pending"
		if _, ok := applied[m.Version]; ok {
			status = "Applied"
		}
		fmt.Printf("[%s] %s - %s\n", status, m.Version, m.Name)
	}

	return nil
}

func createMigrationsTable(db *bsql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255),
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func getAppliedMigrations(db *bsql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, nil
}

func getAppliedMigrationsOrdered(db *bsql.DB) ([]string, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, nil
}

func recordMigration(db *bsql.DB, version, name string) error {
	_, err := db.Exec(
		"INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
		version, name,
	)
	return err
}

func removeMigration(db *bsql.DB, version string) error {
	_, err := db.Exec("DELETE FROM schema_migrations WHERE version = $1", version)
	return err
}

func loadMigrations(path string) ([]*Migration, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var migrations []*Migration
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		m, err := parseMigrationFile(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", file.Name(), err)
		}
		migrations = append(migrations, m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseMigrationFile(filePath string) (*Migration, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fileName := filepath.Base(filePath)
	parts := strings.SplitN(fileName, "_", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid migration file name: %s", fileName)
	}

	version := parts[0]
	name := strings.TrimSuffix(parts[1], ".sql")
	name = strings.ReplaceAll(name, "_", " ")

	contentStr := string(content)
	upSQL := ""
	downSQL := ""

	upIdx := strings.Index(contentStr, "-- +migrate Up")
	downIdx := strings.Index(contentStr, "-- +migrate Down")

	if upIdx != -1 {
		start := upIdx + len("-- +migrate Up")
		end := len(contentStr)
		if downIdx != -1 && downIdx > upIdx {
			end = downIdx
		}
		upSQL = strings.TrimSpace(contentStr[start:end])
	}

	if downIdx != -1 {
		start := downIdx + len("-- +migrate Down")
		downSQL = strings.TrimSpace(contentStr[start:])
	}

	return &Migration{
		Version: version,
		Name:    name,
		UpSQL:   upSQL,
		DownSQL: downSQL,
	}, nil
}
