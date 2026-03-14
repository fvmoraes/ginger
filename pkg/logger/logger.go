// Package logger provides a structured logging abstraction built on log/slog.
// JSON format is the default (recommended for production); use "text" for local dev.
package logger

import (
	"context"
	"log/slog"
	"os"
)

// contextKey is an unexported type for context keys in this package.
// Using a named type prevents collisions with keys from other packages.
type contextKey struct{}

// nopLogger is a shared no-op fallback used by FromContext when no logger
// is stored in the context. Allocated once to avoid per-call heap pressure.
var nopLogger = New("error", "json")

// Logger wraps slog.Logger with context helpers.
type Logger struct {
	*slog.Logger
}

// New creates a structured logger.
//   - level:  "debug" | "info" | "warn" | "error" (default: info)
//   - format: "json" | "text" (default: json)
func New(level, format string) *Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}
	var h slog.Handler
	if format == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	return &Logger{slog.New(h)}
}

// WithContext stores l in ctx and returns the new context.
func WithContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext retrieves the logger stored by WithContext.
// Returns a shared no-op logger (level=error) when none is found,
// avoiding a heap allocation on the hot path.
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(contextKey{}).(*Logger); ok {
		return l
	}
	return nopLogger
}

// With returns a new Logger with additional structured key-value pairs.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}
