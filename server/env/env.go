package env

import (
	"os"
	"time"
)

var E *ENV

type ENV struct {
	Environment            string `yaml:"environment"`
	DatabaseConfigFilePath string `yaml:"database_config_file_path"`
	RedisConfigFilePath    string `yaml:"redis_config_file_path"`

	ServerName string `yaml:"server_name"`

	Backend  *BackendHost  `yaml:"backend"`
	Frontend *FrontendHost `yaml:"frontend"`

	JWTSigningKey       string `yaml:"jwt_signing_key"`
	JWTTokenDuration    string `yaml:"jwt_token_duration"`
	TokenRevokeDuration string `yaml:"token_revoke_duration"`

	TimeZoneOffset int    `yaml:"time_zone_offset"`
	TimeZoneName   string `yaml:"time_zone_name"`

	Features *Features `yaml:"features"`
}

type BackendHost struct {
	HTTPHost   string `yaml:"host_http"`
	SocketHost string `yaml:"host_socket"`
	Port       string `yaml:"port"`
}

type FrontendHost struct {
	APIBaseURL string `yaml:"api_base_url"`
}

type Features struct {
	EnableRegistration bool `yaml:"enable_registration"`
	EnableTokenRevoke  bool `yaml:"enable_token_revoke"`
}

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

func (env *ENV) GetRevokeDuration() time.Duration {
	if env == nil || env.TokenRevokeDuration == "" {
		return 24 * time.Hour
	}
	duration, err := time.ParseDuration(env.TokenRevokeDuration)
	if err != nil {
		return 24 * time.Hour
	}
	return duration
}

func (env *ENV) GetServerPort() string {
	if env == nil || env.Backend == nil || env.Backend.Port == "" {
		return "8080"
	}
	return env.Backend.Port
}

func (env *ENV) GetAPIBaseURL() string {
	if env == nil || env.Frontend == nil || env.Frontend.APIBaseURL == "" {
		return "http://localhost:8080"
	}
	return env.Frontend.APIBaseURL
}

func (env *ENV) IsDevelopment() bool {
	return env != nil && env.Environment == "development"
}

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
	if env.Frontend == nil {
		env.Frontend = &FrontendHost{}
	}
	if env.Frontend.APIBaseURL == "" {
		env.Frontend.APIBaseURL = "http://localhost:" + env.Backend.Port
	}

	// JWT key: environment variable > config file (required, no default)
	if key := os.Getenv("JWT_SIGNING_KEY"); key != "" {
		env.JWTSigningKey = key
	}
	if env.JWTSigningKey == "" {
		panic("JWT_SIGNING_KEY is required. Set it via environment variable or config file.")
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
