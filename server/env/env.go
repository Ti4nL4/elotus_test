package env

import (
	"time"
)

// E is the global environment config
var E *ENV

// ENV holds the environment configuration
type ENV struct {
	Environment            string `yaml:"environment"`
	DatabaseConfigFilePath string `yaml:"database_config_file_path"`

	ServerName string `yaml:"server_name"`

	Backend *BackendHost `yaml:"backend"`

	JWTSigningKey     string `yaml:"jwt_signing_key"`
	JWTTokenDuration  string `yaml:"jwt_token_duration"`
	JWTSigningKeyUser string `yaml:"jwt_signing_key_user"`

	TimeZoneOffset int    `yaml:"time_zone_offset"`
	TimeZoneName   string `yaml:"time_zone_name"`

	Features *Features `yaml:"features"`
}

// BackendHost holds backend server configuration
type BackendHost struct {
	HTTPHost   string `yaml:"host_http"`
	SocketHost string `yaml:"host_socket"`
	Port       string `yaml:"port"`
}

// Features holds feature flags
type Features struct {
	EnableRegistration bool `yaml:"enable_registration"`
	EnableTokenRevoke  bool `yaml:"enable_token_revoke"`
}

// GetJWTDuration returns JWT token duration as time.Duration
func (env *ENV) GetJWTDuration() time.Duration {
	if env == nil || env.JWTTokenDuration == "" {
		return 24 * time.Hour
	}
	duration, err := time.ParseDuration(env.JWTTokenDuration)
	if err != nil {
		return 24 * time.Hour
	}
	return duration
}

// GetServerPort returns the server port
func (env *ENV) GetServerPort() string {
	if env == nil || env.Backend == nil || env.Backend.Port == "" {
		return "8080"
	}
	return env.Backend.Port
}

// IsDevelopment returns true if environment is development
func (env *ENV) IsDevelopment() bool {
	return env != nil && env.Environment == "development"
}

// IsProduction returns true if environment is production
func (env *ENV) IsProduction() bool {
	return env != nil && env.Environment == "production"
}

// SetDefaults sets default values for ENV
func (env *ENV) SetDefaults() {
	if env.Environment == "" {
		env.Environment = "development"
	}
	if env.ServerName == "" {
		env.ServerName = "elotus-auth"
	}
	if env.Backend == nil {
		env.Backend = &BackendHost{}
	}
	if env.Backend.Port == "" {
		env.Backend.Port = "8080"
	}
	if env.Backend.HTTPHost == "" {
		env.Backend.HTTPHost = "localhost"
	}
	if env.JWTSigningKey == "" {
		env.JWTSigningKey = "your-256-bit-secret-key-here-change-in-production"
	}
	if env.JWTTokenDuration == "" {
		env.JWTTokenDuration = "24h"
	}
	if env.TimeZoneName == "" {
		env.TimeZoneName = "Asia/Ho_Chi_Minh"
	}
	if env.TimeZoneOffset == 0 {
		env.TimeZoneOffset = 7
	}
	if env.Features == nil {
		env.Features = &Features{
			EnableRegistration: true,
			EnableTokenRevoke:  true,
		}
	}
}
