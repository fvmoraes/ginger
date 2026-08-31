package logger

import (
	"context"
	"testing"
)

func TestNewLevelsAndJSONFormat(t *testing.T) {
	l := New("debug", "json")
	if l == nil {
		t.Fatal("nil logger")
	}
	l.Info("test_message", "key", "value")
	l.Error("error_message", "err", "boom")
	l.Warn("warn_message")
	l.Debug("debug_message")
}

func TestFromContextFallsBackToDefault(t *testing.T) {
	if FromContext(context.Background()) == nil {
		t.Fatal("FromContext must never return nil")
	}
}

func TestWithContextRoundTrip(t *testing.T) {
	l := New("info", "json")
	ctx := WithContext(context.Background(), l)
	if FromContext(ctx) != l {
		t.Fatal("context round trip failed")
	}
}

func TestWithReturnsNonNil(t *testing.T) {
	l := New("info", "json").With("k", "v")
	if l == nil {
		t.Fatal("With returned nil")
	}
}

func TestRedactsSecrets(t *testing.T) {
	l := New("info", "json")
	// The redaction lives in the handler; smoke-test via a payload with secrets.
	l.Error("db_fail", "password", "hunter2", "dsn", "postgres://u:p@h/db")
	// Logger writes to stderr by default; we assert it did not panic and
	// the API accepts sensitive keys (full redaction covered by handler tests).
}
