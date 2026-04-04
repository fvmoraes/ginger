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

func TestReadModulePath(t *testing.T) {
	tmp := t.TempDir()
	goModPath := filepath.Join(tmp, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module example.com/demo\n\ngo 1.25\n"), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	got, err := readModulePath(goModPath)
	if err != nil {
		t.Fatalf("readModulePath returned error: %v", err)
	}
	if got != "example.com/demo" {
		t.Fatalf("readModulePath() = %q, want %q", got, "example.com/demo")
	}
}

func TestCLIBuildFlagArgsUsesProjectMetadata(t *testing.T) {
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

	if err := os.MkdirAll(filepath.Join("internal", "commands"), 0755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile("go.mod", []byte("module example.com/demo\n\ngo 1.25\n"), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join("internal", "commands", "version.go"), []byte("package commands\n"), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	originalGitOutput := gitOutput
	gitOutput = func(args ...string) (string, error) {
		switch strings.Join(args, " ") {
		case "describe --tags --exact-match":
			return "v1.2.3", nil
		case "rev-parse --short HEAD":
			return "abc1234", nil
		default:
			t.Fatalf("unexpected git args: %v", args)
			return "", nil
		}
	}
	defer func() { gitOutput = originalGitOutput }()

	got := cliBuildFlagArgs()
	if len(got) != 2 || got[0] != "-ldflags" {
		t.Fatalf("cliBuildFlagArgs() = %v, want -ldflags pair", got)
	}

	for _, want := range []string{
		"example.com/demo/internal/commands.version=1.2.3",
		"example.com/demo/internal/commands.commit=abc1234",
		"example.com/demo/internal/commands.date=",
	} {
		if !strings.Contains(got[1], want) {
			t.Fatalf("expected ldflags to contain %q, got %q", want, got[1])
		}
	}
}
