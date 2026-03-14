// Package logger provides a structured logging abstraction over slog.
package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey struct{}

// Logger wraps slog.Logger.
type Logger struct {
	*slog.Logger
}

// New creates a structured logger.
// level: "debug" | "info" | "warn" | "error" (defaults to info).
// format: "json" | "text" (defaults to json, recommended for production).
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
	var handler slog.Handler
	if format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return &Logger{slog.New(handler)}
}

// WithContext stores the logger in a context.
func WithContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext retrieves the logger from context, falling back to a default.
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(contextKey{}).(*Logger); ok {
		return l
	}
	return New("info", "json")
}

// With returns a new Logger with additional key-value pairs.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}
