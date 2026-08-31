package integrations

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/fvmoraes/ginger/internal/project"
	"github.com/fvmoraes/ginger/internal/scaffold"
)

// TestApplyWithRollbackOnFailedGoGet (GIN-005 / N4): falha do `go get`
// desfaz TODO o apply — creates, modifies (compose) e o manifest.
func TestApplyWithRollbackOnFailedGoGet(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := scaffold.NewProject("demo", "service"); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	root := filepath.Join(dir, "demo")
	prj, err := project.Load(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	composeBefore, err := os.ReadFile(filepath.Join(root, "devops", "docker", "docker-compose.yml"))
	if err != nil {
		t.Fatalf("read compose: %v", err)
	}
	manifestBefore, err := os.ReadFile(filepath.Join(root, ".ginger", "manifest.yaml"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	// Force the post-apply `go get` to fail.
	origExec := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}
	t.Cleanup(func() { execCommand = origExec })

	p, err := Plan("redis", prj, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if err := ApplyWithRollback(p, "redis", root); err == nil {
		t.Fatal("expected ApplyWithRollback to fail on simulated go get error")
	}

	// Created file must be gone.
	if _, err := os.Stat(filepath.Join(root, "platform", "cache", "redis.go")); !os.IsNotExist(err) {
		t.Fatalf("created file survived rollback: %v", err)
	}
	// Modified compose restored byte a byte.
	composeAfter, _ := os.ReadFile(filepath.Join(root, "devops", "docker", "docker-compose.yml"))
	if string(composeAfter) != string(composeBefore) {
		t.Error("compose was not restored byte a byte")
	}
	// Manifest restored.
	manifestAfter, _ := os.ReadFile(filepath.Join(root, ".ginger", "manifest.yaml"))
	if string(manifestAfter) != string(manifestBefore) {
		t.Error("manifest was not restored")
	}
}

// TestApplyWithRollbackSucceedsNormally: caminho feliz sem rollback.
func TestApplyWithRollbackSucceedsNormally(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := scaffold.NewProject("demo", "service"); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	root := filepath.Join(dir, "demo")
	prj, err := project.Load(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// Stdlib-only integration: no post-apply, no rollback needed.
	p, err := Plan("mcp", prj, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if err := ApplyWithRollback(p, "mcp", root); err != nil {
		t.Fatalf("ApplyWithRollback: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "platform", "mcp", "server.go")); err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Logf("mcp file stat: %v", err)
	}
}
