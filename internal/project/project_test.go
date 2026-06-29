package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRoot_GoMod(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	root, err := FindRoot(sub)
	if err != nil {
		t.Fatalf("FindRoot: %v", err)
	}
	if root != dir {
		t.Errorf("FindRoot = %s, want %s", root, dir)
	}
}

func TestFindRoot_GingerYAML(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "x", "y")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	writeProjectTestFile(t, filepath.Join(dir, "ginger.yaml"), "project:\n  type: service\n")
	writeProjectTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/test\n")

	// ginger.yaml should win over go.mod
	root, err := FindRoot(sub)
	if err != nil {
		t.Fatalf("FindRoot: %v", err)
	}
	if root != dir {
		t.Errorf("FindRoot = %s, want %s", root, dir)
	}
}

func TestFindRoot_NoProject(t *testing.T) {
	dir := t.TempDir()
	_, err := FindRoot(dir)
	if err == nil {
		t.Fatal("expected error for directory without project markers")
	}
}

func TestFindRoot_GitWorktreeFile(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "nested")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	writeProjectTestFile(t, filepath.Join(dir, ".git"), "gitdir: /tmp/example\n")
	root, err := FindRoot(sub)
	if err != nil {
		t.Fatalf("FindRoot: %v", err)
	}
	if root != dir {
		t.Fatalf("root = %s, want %s", root, dir)
	}
}

func TestLoad_WithoutGingerYAML(t *testing.T) {
	dir := t.TempDir()
	writeProjectTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/myapp\n")
	if err := os.MkdirAll(filepath.Join(dir, "cmd", "app"), 0755); err != nil {
		t.Fatal(err)
	}
	writeProjectTestFile(t, filepath.Join(dir, "cmd", "app", "main.go"), "package main\n")
	if err := os.MkdirAll(filepath.Join(dir, "internal", "api", "handlers"), 0755); err != nil {
		t.Fatal(err)
	}

	p, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.Module != "example.com/myapp" {
		t.Errorf("Module = %s, want example.com/myapp", p.Module)
	}
	if p.ProjectType() != "service" {
		t.Errorf("ProjectType = %s, want service", p.ProjectType())
	}
	if p.HasGingerYAML() {
		t.Error("HasGingerYAML should be false")
	}
}

func TestLoad_WithGingerYAML(t *testing.T) {
	dir := t.TempDir()
	writeProjectTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/test\n")
	writeProjectTestFile(t, filepath.Join(dir, "ginger.yaml"), `
project:
  type: worker
structure:
  cmd: cmd/worker
rules:
  overwrite: true
`)

	p, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.ProjectType() != "worker" {
		t.Errorf("ProjectType = %s, want worker", p.ProjectType())
	}
	if !p.ShouldOverwrite() {
		t.Error("ShouldOverwrite should be true")
	}
	if !p.HasGingerYAML() {
		t.Error("HasGingerYAML should be true")
	}
}

func TestResolvePath(t *testing.T) {
	dir := t.TempDir()
	writeProjectTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/test\n")
	p, _ := Load(dir)

	got, err := p.ResolvePath("handlers")
	if err != nil {
		t.Fatalf("ResolvePath: %v", err)
	}
	expected := filepath.Join(dir, "internal", "api", "handlers")
	if got != expected {
		t.Errorf("ResolvePath(handlers) = %s, want %s", got, expected)
	}
}

func TestLoadRejectsPathsOutsideProject(t *testing.T) {
	dir := t.TempDir()
	writeProjectTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/test\n")
	writeProjectTestFile(t, filepath.Join(dir, "ginger.yaml"), `project:
  type: service
  root: .
structure:
  handlers: ../../outside
`)

	if _, err := Load(dir); err == nil {
		t.Fatal("expected an escaping structure path to be rejected")
	}
}

func TestLoadRejectsUnknownContractFields(t *testing.T) {
	dir := t.TempDir()
	writeProjectTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/test\n")
	writeProjectTestFile(t, filepath.Join(dir, "ginger.yaml"), `project:
  type: service
rules:
  overrite: true
`)

	if _, err := Load(dir); err == nil {
		t.Fatal("expected unknown ginger.yaml fields to be rejected")
	}
}

func TestDetectCustomExistingServiceStructure(t *testing.T) {
	dir := t.TempDir()
	writeProjectTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/existing\n")
	for _, path := range []string{"cmd/server", "internal/httpapi/handlers", "internal/core", "internal/store"} {
		if err := os.MkdirAll(filepath.Join(dir, path), 0755); err != nil {
			t.Fatal(err)
		}
	}

	p, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if p.ProjectType() != "service" {
		t.Fatalf("ProjectType = %s, want service", p.ProjectType())
	}
	if p.YAML.Structure.API != "internal/httpapi" || p.YAML.Structure.Services != "internal/core" || p.YAML.Structure.Repositories != "internal/store" {
		t.Fatalf("unexpected detected structure: %+v", p.YAML.Structure)
	}
}

func writeProjectTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}
