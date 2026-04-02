// Package logger provides a structured logging abstraction built on log/slog.
// Ginger always emits pretty-printed JSON logs to keep ingestion predictable
// across environments and observability backends.
package logger

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/trace"
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
//   - format is kept for backward compatibility and ignored; Ginger logs are
//     always emitted as pretty-printed JSON.
func New(level, _ string) *Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	h := newPrettyJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return &Logger{slog.New(h)}
}

// WithContext stores l in ctx and returns the new context.
func WithContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext retrieves the logger stored by WithContext.
// Returns a shared no-op logger (level=error) when none is found.
// When an OTel span is present, trace identifiers are attached automatically.
func FromContext(ctx context.Context) *Logger {
	base := nopLogger
	if l, ok := ctx.Value(contextKey{}).(*Logger); ok {
		base = l
	}

	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return base
	}

	return base.With(
		"trace_id", spanCtx.TraceID().String(),
		"span_id", spanCtx.SpanID().String(),
	)
}

// With returns a new Logger with additional structured key-value pairs.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}

type prettyJSONHandler struct {
	out    io.Writer
	opts   *slog.HandlerOptions
	attrs  []slog.Attr
	groups []string
	mu     *sync.Mutex
}

func newPrettyJSONHandler(out io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &prettyJSONHandler{
		out:  out,
		opts: opts,
		mu:   &sync.Mutex{},
	}
}

func (h *prettyJSONHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts != nil && h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *prettyJSONHandler) Handle(ctx context.Context, record slog.Record) error {
	payload := map[string]any{
		"time":  record.Time.UTC().Format("2006-01-02T15:04:05.000000000Z07:00"),
		"level": strings.ToLower(record.Level.String()),
		"msg":   record.Message,
	}

	for _, attr := range h.attrs {
		h.applyAttr(payload, attr)
	}

	record.Attrs(func(attr slog.Attr) bool {
		h.applyAttr(payload, attr)
		return true
	})

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		payload["trace_id"] = spanCtx.TraceID().String()
		payload["span_id"] = spanCtx.SpanID().String()
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err = h.out.Write(append(data, '\n'))
	return err
}

func (h *prettyJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &clone
}

func (h *prettyJSONHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	clone := *h
	clone.groups = append(append([]string{}, h.groups...), name)
	return &clone
}

func (h *prettyJSONHandler) applyAttr(dst map[string]any, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if h.opts != nil && h.opts.ReplaceAttr != nil {
		attr = h.opts.ReplaceAttr(h.groups, attr)
	}
	if attr.Equal(slog.Attr{}) || attr.Key == "" {
		return
	}

	value := attrValue(attr.Value)
	if isSensitiveKey(attr.Key) {
		value = redactValue(value)
	}

	path := append(append([]string{}, h.groups...), attr.Key)
	setNested(dst, path, value)
}

func attrValue(v slog.Value) any {
	switch v.Kind() {
	case slog.KindBool:
		return v.Bool()
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindString:
		return v.String()
	case slog.KindTime:
		return v.Time().UTC().Format("2006-01-02T15:04:05.000000000Z07:00")
	case slog.KindUint64:
		return v.Uint64()
	case slog.KindGroup:
		group := make(map[string]any)
		for _, attr := range v.Group() {
			attr.Value = attr.Value.Resolve()
			if attr.Key == "" {
				continue
			}
			value := attrValue(attr.Value)
			if isSensitiveKey(attr.Key) {
				value = redactValue(value)
			}
			group[attr.Key] = value
		}
		return group
	case slog.KindAny:
		return v.Any()
	default:
		return v.String()
	}
}

func setNested(dst map[string]any, path []string, value any) {
	if len(path) == 1 {
		dst[path[0]] = value
		return
	}

	head := path[0]
	next, ok := dst[head].(map[string]any)
	if !ok {
		next = make(map[string]any)
		dst[head] = next
	}
	setNested(next, path[1:], value)
}

func isSensitiveKey(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	switch k {
	case "authorization", "cookie", "set-cookie", "password", "passwd", "secret", "token", "access_token", "refresh_token", "api_key", "apikey", "dsn":
		return true
	default:
		return strings.Contains(k, "password") ||
			strings.Contains(k, "secret") ||
			strings.Contains(k, "token") ||
			strings.Contains(k, "authorization") ||
			strings.Contains(k, "cookie")
	}
}

func redactValue(v any) any {
	switch value := v.(type) {
	case string:
		if value == "" {
			return ""
		}
		return "[REDACTED]"
	default:
		return "[REDACTED]"
	}
}
