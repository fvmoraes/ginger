package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanAddCreate_NewFile(t *testing.T) {
	dir := t.TempDir()
	p := &Plan{ProjectRoot: dir}
	target := filepath.Join(dir, "new.txt")

	p.AddCreate(target, []byte("hello"), false)

	if len(p.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(p.Changes))
	}
	if p.Changes[0].Type != ChangeCreate {
		t.Errorf("expected create, got %s", p.Changes[0].Type)
	}
}

func TestPlanAddCreate_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	p := &Plan{ProjectRoot: dir}
	p.AddCreate(target, []byte("new"), false)

	if p.Changes[0].Type != ChangeSkip {
		t.Errorf("expected skip for existing file without force, got %s", p.Changes[0].Type)
	}
}

func TestPlanAddCreate_ForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	p := &Plan{ProjectRoot: dir}
	p.AddCreate(target, []byte("new"), true)

	if p.Changes[0].Type != ChangeModify {
		t.Errorf("expected modify with force, got %s", p.Changes[0].Type)
	}
}

func TestPlanApply(t *testing.T) {
	dir := t.TempDir()
	p := &Plan{ProjectRoot: dir}
	target := filepath.Join(dir, "created.txt")

	p.AddCreate(target, []byte("content"), false)
	if err := p.Apply(); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("content = %q, want 'content'", string(data))
	}
}

func TestPlanApply_SkipDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(target, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	p := &Plan{ProjectRoot: dir}
	p.AddCreate(target, []byte("should-not-write"), false)
	if err := p.Apply(); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(target)
	if string(data) != "original" {
		t.Errorf("content = %q, want 'original' (skip must not overwrite)", string(data))
	}
}

func TestPlanRender(t *testing.T) {
	p := &Plan{ProjectRoot: "/tmp/test"}
	p.AddCreate("/tmp/test/a.txt", nil, false)
	p.AddWarning("test warning")

	var sb strings.Builder
	// Capture rendering by calling methods directly
	// (we test that it doesn't panic, not output format)
	p.Render()
	if len(p.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(p.Warnings))
	}
	_ = sb
}

func TestPlanRejectsPathOutsideRoot(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	p := New(dir)
	p.AddCreate(outside, []byte("unsafe"), false)
	if !p.HasErrors() {
		t.Fatal("expected outside path to create a plan error")
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatalf("outside path should not be written, stat err=%v", err)
	}
}

func TestPlanRejectsSymlinkEscape(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(dir, "linked")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	p := New(dir)
	p.AddCreate(filepath.Join(link, "escaped.txt"), []byte("unsafe"), false)
	if !p.HasErrors() {
		t.Fatal("expected symlink escape to create a plan error")
	}
}

func TestPlanRejectsStaleCreate(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "new.txt")
	p := New(dir)
	p.AddCreate(target, []byte("planned"), false)
	if err := os.WriteFile(target, []byte("user"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := p.Apply(); err == nil || !strings.Contains(err.Error(), "stale plan") {
		t.Fatalf("expected stale plan error, got %v", err)
	}
	data, _ := os.ReadFile(target)
	if string(data) != "user" {
		t.Fatalf("stale apply overwrote user content: %q", data)
	}
}

func TestPlanRejectsStaleModify(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "existing.txt")
	if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	p := New(dir)
	p.AddModify(target, []byte("planned"), true)
	if err := os.WriteFile(target, []byte("user edit"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := p.Apply(); err == nil || !strings.Contains(err.Error(), "stale plan") {
		t.Fatalf("expected stale plan error, got %v", err)
	}
}

func TestPlanRespectsCreateMissingDirs(t *testing.T) {
	dir := t.TempDir()
	p := New(dir)
	p.CreateMissingDirs = false
	p.AddCreate(filepath.Join(dir, "missing", "file.go"), []byte("package missing\n"), false)
	if !p.HasErrors() {
		t.Fatal("expected missing parent directory to block the plan")
	}
}

func TestPlanSkipsIdenticalManagedContent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "managed.txt")
	if err := os.WriteFile(target, []byte("same"), 0644); err != nil {
		t.Fatal(err)
	}
	p := New(dir)
	p.AddModify(target, []byte("same"), true)
	if len(p.Changes) != 1 || p.Changes[0].Type != ChangeSkip {
		t.Fatalf("identical content should be skipped: %+v", p.Changes)
	}
}
