package bsql

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type              string `yaml:"type"`
	Host              string `yaml:"host"`
	Port              string `yaml:"port"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	Database          string `yaml:"database"`
	MaxIdleConnection int    `yaml:"maxIdleConnection"`
	MaxOpenConnection int    `yaml:"maxOpenConnection"`
}

// LoadDatabaseConfig loads database configuration from yaml file
func LoadDatabaseConfig(path string) (*DatabaseConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read database config file: %w", err)
	}

	var config DatabaseConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse database config file: %w", err)
	}

	// Set defaults
	if config.Type == "" {
		config.Type = "postgres"
	}
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == "" {
		config.Port = "5432"
	}
	if config.MaxIdleConnection == 0 {
		config.MaxIdleConnection = 40
	}
	if config.MaxOpenConnection == 0 {
		config.MaxOpenConnection = 80
	}

	return &config, nil
}

// OpenFromConfig opens database from config file
func OpenFromConfig(configPath string) (*DB, error) {
	config, err := LoadDatabaseConfig(configPath)
	if err != nil {
		return nil, err
	}

	db := Open(
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.MaxIdleConnection,
		config.MaxOpenConnection,
	)

	return db, nil
}
