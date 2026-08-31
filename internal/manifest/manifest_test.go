package manifest

import (
	"os"

	"github.com/fvmoraes/ginger/internal/plan"
	"path/filepath"
	"testing"
)

func TestGeneratedHashRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ginger"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	manifestPath := filepath.Join(dir, ".ginger", "manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte("managed:\n  - path: devops/docker/docker-compose.yml\n    full_file: true\n    generated_hash: abc123\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	m, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !m.ManagesFullFile("devops/docker/docker-compose.yml") {
		t.Fatal("full file ownership lost")
	}
	if m.GeneratedHash("devops/docker/docker-compose.yml") != "abc123" {
		t.Fatalf("hash = %q", m.GeneratedHash("devops/docker/docker-compose.yml"))
	}
	if m.GeneratedHash("other") != "" {
		t.Fatal("unknown path must have empty hash")
	}
}

func TestOldManifestWithoutHashParses(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ginger"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Pre-GIN-002 manifest: no generated_hash field.
	if err := os.WriteFile(filepath.Join(dir, ".ginger", "manifest.yaml"),
		[]byte("managed:\n  - path: ginger.yaml\n    full_file: true\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	m, err := Load(dir)
	if err != nil {
		t.Fatalf("old manifest must parse: %v", err)
	}
	if m.GeneratedHash("ginger.yaml") != "" {
		t.Fatal("old manifest must report empty hash (data-safe fallback)")
	}
}

func TestAddRefreshesHash(t *testing.T) {
	m := &Manifest{}
	m.Add(Entry{Path: "a.yml", FullFile: true, GeneratedHash: "h1"})
	m.Add(Entry{Path: "a.yml", FullFile: true, GeneratedHash: "h2"})
	if m.GeneratedHash("a.yml") != "h2" {
		t.Fatalf("hash not refreshed: %q", m.GeneratedHash("a.yml"))
	}
	// Empty hash must not wipe a recorded one.
	m.Add(Entry{Path: "a.yml", FullFile: true})
	if m.GeneratedHash("a.yml") != "h2" {
		t.Fatalf("empty hash wiped existing: %q", m.GeneratedHash("a.yml"))
	}
}

func TestUnknownFieldsRejected(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ginger"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".ginger", "manifest.yaml"),
		[]byte("managed:\n  - path: x\n    typo_field: true\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("unknown manifest fields must be rejected")
	}
}

func TestManagesRegionAndAdd(t *testing.T) {
	m := &Manifest{}
	m.Add(Entry{Path: "internal/api/router.go", FullFile: true, Regions: []string{"routes"}})
	if !m.ManagesRegion("internal/api/router.go", "routes") {
		t.Fatal("region not managed")
	}
	if !m.ManagesFullFile("internal/api/router.go") {
		t.Fatal("full file not managed")
	}
	if m.ManagesRegion("internal/api/router.go", "other") {
		t.Fatal("unregistered region reported as managed")
	}
	if m.ManagesFullFile("other.go") {
		t.Fatal("unknown path reported as managed")
	}
	// Add merges regions without duplicating.
	m.Add(Entry{Path: "internal/api/router.go", Regions: []string{"routes", "imports"}})
	if len(m.Managed[0].Regions) != 2 {
		t.Fatalf("regions not merged: %v", m.Managed[0].Regions)
	}
}

func TestPlanUpdateAddsManifestPlan(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ginger"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	p := plan.New(dir)
	if err := PlanUpdate(p, Entry{Path: "new-file.go", FullFile: true, GeneratedHash: "h"}); err != nil {
		t.Fatalf("PlanUpdate: %v", err)
	}
	if !p.HasChanges() {
		t.Fatal("PlanUpdate must add a manifest change")
	}
	// Apply and verify the on-disk manifest.
	if err := p.Apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	m, err := Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if m.GeneratedHash("new-file.go") != "h" {
		t.Fatalf("hash not persisted: %q", m.GeneratedHash("new-file.go"))
	}
}
