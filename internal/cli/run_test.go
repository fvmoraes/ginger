package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectCmdDirRejectsMultipleEntrypoints(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()

	for _, name := range []string{"service", "worker"} {
		dir := filepath.Join("cmd", name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll returned error: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644); err != nil {
			t.Fatalf("WriteFile returned error: %v", err)
		}
	}

	_, err = detectCmdDir()
	if err == nil || !strings.Contains(err.Error(), "multiple app entrypoints found") {
		t.Fatalf("expected multiple entrypoints error, got %v", err)
	}
}
