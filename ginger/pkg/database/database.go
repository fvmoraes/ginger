// Package database provides helpers for connecting to SQL databases.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	// Import your driver in main.go, e.g.:
	//   _ "github.com/lib/pq"           // postgres
	//   _ "github.com/mattn/go-sqlite3" // sqlite
)

// ErrNoDriver is returned when Connect is called without a driver name.
var ErrNoDriver = errors.New("database: driver is required")

// Config holds database connection settings.
type Config struct {
	Driver  string
	DSN     string
	MaxOpen int
	MaxIdle int
}

// Connect opens and validates a database connection.
// Returns ErrNoDriver if cfg.Driver is empty.
func Connect(cfg Config) (*sql.DB, error) {
	if cfg.Driver == "" {
		return nil, ErrNoDriver
	}
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database: ping: %w", err)
	}
	return db, nil
}

// Checker implements health.Checker for a *sql.DB.
type Checker struct {
	db *sql.DB
}

func NewChecker(db *sql.DB) *Checker { return &Checker{db: db} }
func (c *Checker) Name() string      { return "database" }
func (c *Checker) Check(ctx context.Context) error {
	return c.db.PingContext(ctx)
}
