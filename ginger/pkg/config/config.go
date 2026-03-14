// Package config provides configuration loading from env and yaml files.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	App      AppConfig      `yaml:"app"`
	HTTP     HTTPConfig     `yaml:"http"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Env     string `yaml:"env"`
	Version string `yaml:"version"`
}

type HTTPConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ShutdownTimeout int    `yaml:"shutdown_timeout"` // seconds
}

type DatabaseConfig struct {
	Driver  string `yaml:"driver"`
	DSN     string `yaml:"dsn"`
	MaxOpen int    `yaml:"max_open"`
	MaxIdle int    `yaml:"max_idle"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // json | text
}

// Load reads config from a yaml file, then overrides with env vars.
func Load(path string) (*Config, error) {
	cfg := defaults()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("config: read file: %w", err)
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

func applyEnv(cfg *Config) {
	if v := os.Getenv("APP_NAME"); v != "" {
		cfg.App.Name = v
	}
	if v := os.Getenv("APP_ENV"); v != "" {
		cfg.App.Env = v
	}
	if v := os.Getenv("APP_VERSION"); v != "" {
		cfg.App.Version = v
	}
	if v := os.Getenv("HTTP_HOST"); v != "" {
		cfg.HTTP.Host = v
	}
	if v := os.Getenv("HTTP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.HTTP.Port = p
		}
	}
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.Database.DSN = v
	}
	if v := os.Getenv("DATABASE_DRIVER"); v != "" {
		cfg.Database.Driver = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Log.Level = strings.ToLower(v)
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.Log.Format = strings.ToLower(v)
	}
}
