package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	_ = os.WriteFile(path, []byte("app:\n  name: demo\n  env: prod\nhttp:\n  port: 9000\n"), 0o644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.App.Name != "demo" || cfg.App.Env != "prod" || cfg.HTTP.Port != 9000 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadInvalidYAMLErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	_ = os.WriteFile(path, []byte("http: [unclosed\n"), 0o644)
	if _, err := Load(path); err == nil {
		t.Fatal("invalid yaml must error")
	}
}

// Runtime config is tolerant by design: missing file → defaults (documented).
// The strict contract (KnownFields) applies to ginger.yaml, not runtime config.
func TestLoadMissingFileFallsBackToDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil {
		t.Fatalf("missing file must fall back to defaults, got %v", err)
	}
	if cfg.HTTP.Port != 8080 {
		t.Fatalf("default port = %d, want 8080", cfg.HTTP.Port)
	}
}

func TestDefaults(t *testing.T) {
	cfg := defaults()
	if cfg.HTTP.ReadHeaderTimeout != 5 {
		t.Fatalf("ReadHeaderTimeout = %d, want 5 (GIN-023)", cfg.HTTP.ReadHeaderTimeout)
	}
	if cfg.HTTP.Host != "0.0.0.0" || cfg.App.Env != "development" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if !strings.Contains("", "") {
		t.Fatal("unreachable")
	}
}
