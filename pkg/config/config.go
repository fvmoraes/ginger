// Package config provides configuration loading from YAML files with
// environment variable overrides.
// YAML is read first; env vars take precedence, following the twelve-factor
// app methodology (https://12factor.net/config).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the full application configuration.
type Config struct {
	App      AppConfig      `yaml:"app"`
	HTTP     HTTPConfig     `yaml:"http"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
}

// AppConfig holds application-level metadata.
type AppConfig struct {
	Name    string `yaml:"name"`
	Env     string `yaml:"env"`
	Version string `yaml:"version"`
}

// HTTPConfig holds HTTP server settings.
type HTTPConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ShutdownTimeout int    `yaml:"shutdown_timeout"` // seconds
}

// DatabaseConfig holds SQL database connection settings.
type DatabaseConfig struct {
	Driver  string `yaml:"driver"`
	DSN     string `yaml:"dsn"`
	MaxOpen int    `yaml:"max_open"`
	MaxIdle int    `yaml:"max_idle"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // json | text
}

// Load reads config from path (YAML), then overrides with environment variables.
// If path is empty or the file does not exist, only defaults + env vars are used.
func Load(path string) (*Config, error) {
	cfg := defaults()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("config: read %s: %w", path, err)
		}
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("config: parse yaml: %w", err)
			}
		}
	}

	applyEnv(cfg)
	return cfg, nil
}

func defaults() *Config {
	return &Config{
		App:      AppConfig{Name: "ginger-app", Env: "development", Version: "0.0.1"},
		HTTP:     HTTPConfig{Host: "0.0.0.0", Port: 8080, ShutdownTimeout: 30},
		Log:      LogConfig{Level: "info", Format: "json"},
		Database: DatabaseConfig{MaxOpen: 25, MaxIdle: 5},
	}
}

// envString overrides dst with the env var value when non-empty.
func envString(dst *string, key string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

// envStringLower overrides dst with the lowercased env var value when non-empty.
func envStringLower(dst *string, key string) {
	if v := os.Getenv(key); v != "" {
		*dst = strings.ToLower(v)
	}
}

// envInt overrides dst with the parsed env var value when non-empty and valid.
func envInt(dst *int, key string) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*dst = n
		}
	}
}

// applyEnv overrides cfg fields from environment variables.
// Uses a table-driven approach to keep the mapping explicit and DRY.
func applyEnv(cfg *Config) {
	envString(&cfg.App.Name, "APP_NAME")
	envString(&cfg.App.Env, "APP_ENV")
	envString(&cfg.App.Version, "APP_VERSION")
	envString(&cfg.HTTP.Host, "HTTP_HOST")
	envInt(&cfg.HTTP.Port, "HTTP_PORT")
	envString(&cfg.Database.Driver, "DATABASE_DRIVER")
	envString(&cfg.Database.DSN, "DATABASE_DSN")
	envStringLower(&cfg.Log.Level, "LOG_LEVEL")
	envStringLower(&cfg.Log.Format, "LOG_FORMAT")
}
