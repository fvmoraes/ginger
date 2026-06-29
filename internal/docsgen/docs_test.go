package docsgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fvmoraes/ginger/internal/plan"
	"github.com/fvmoraes/ginger/internal/project"
)

func TestBuildPlanUsesConfiguredDocsAndPreservesCustomFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/docs\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ginger.yaml"), []byte("project:\n  type: service\nstructure:\n  docs: documentation\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "documentation"), 0755); err != nil {
		t.Fatal(err)
	}
	custom := filepath.Join(dir, "documentation", "routes.md")
	if err := os.WriteFile(custom, []byte("# My custom routes\n"), 0644); err != nil {
		t.Fatal(err)
	}
	prj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	p, err := BuildPlan(prj, false)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	foundSkip := false
	for _, change := range p.Changes {
		if change.Path == custom && change.Type == plan.ChangeSkip {
			foundSkip = true
		}
	}
	if !foundSkip {
		t.Fatalf("custom routes.md was not preserved: %+v", p.Changes)
	}
}
