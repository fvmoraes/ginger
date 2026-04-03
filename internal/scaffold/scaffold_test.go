package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewProjectRejectsExistingDirectory(t *testing.T) {
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

	projectDir := filepath.Join(tmp, "demo")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = NewProject("demo", "api")
	if !errors.Is(err, ErrProjectExists) {
		t.Fatalf("expected ErrProjectExists, got %v", err)
	}
}

func TestNewProjectRejectsPathLikeName(t *testing.T) {
	err := NewProject(filepath.Join("tmp", "demo"), "api")
	if !errors.Is(err, ErrInvalidProjectName) {
		t.Fatalf("expected ErrInvalidProjectName, got %v", err)
	}
}
