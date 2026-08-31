package doctor

import (
	"os"
	"path/filepath"

	"testing"

	"github.com/fvmoraes/ginger/internal/project"
	"github.com/fvmoraes/ginger/internal/scaffold"
)

// TestDiagnoseFreshServiceScaffold (GIN-008): doctor 0 falhas em scaffold novo.
func TestDiagnoseFreshServiceScaffold(t *testing.T) {
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

	d, err := Diagnose(prj)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}
	if !d.Healthy() {
		for _, c := range d.Checks {
			if c.Status == "fail" {
				t.Errorf("unexpected fail: %s — %s", c.Name, c.Detail)
			}
		}
	}
	// Catalogued capabilities must appear.
	found := false
	for _, name := range d.AvailableCapabilities {
		if name == "postgres" {
			found = true
		}
	}
	if !found {
		t.Error("postgres capability missing from available list")
	}
}

// TestDiagnosticHealthySemantics: warn não é fail (GIN-008).
func TestDiagnosticHealthySemantics(t *testing.T) {
	d := &Diagnostic{Root: "/x"}
	d.Checks = append(d.Checks, DiagnosticCheck{Name: "a", Status: "pass"})
	if !d.Healthy() {
		t.Fatal("pass-only diagnostic must be healthy")
	}
	d.Checks = append(d.Checks, DiagnosticCheck{Name: "tests", Status: "warn"})
	if !d.Healthy() {
		t.Fatal("warn must not make the diagnostic unhealthy (GIN-008)")
	}
	d.Checks = append(d.Checks, DiagnosticCheck{Name: "vet", Status: "fail"})
	if d.Healthy() {
		t.Fatal("fail must make the diagnostic unhealthy")
	}
}

// TestMissingDepsDetection: padrões de dependência ausente (GIN-008).
func TestMissingDepsDetection(t *testing.T) {
	cases := []struct {
		detail string
		want   bool
	}{
		{"cmd/x/main.go:1: missing go.sum entry for module providing package", true},
		{"no required module provides package github.com/x/y", true},
		{"cannot find package \"x\"", true},
		{"expected declaration, found badcode", false},
	}
	for _, c := range cases {
		if got := missingDeps(c.detail); got != c.want {
			t.Errorf("missingDeps(%q) = %v, want %v", c.detail, got, c.want)
		}
	}
}
