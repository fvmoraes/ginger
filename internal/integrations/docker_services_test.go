package integrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fvmoraes/ginger/internal/project"
	"github.com/fvmoraes/ginger/internal/scaffold"
)

// TestDockerServicesTableDrift (Fase 4): a tabela central deve refletir o
// que o switch de merge realmente produz (imagem e nome de serviço).
func TestDockerServicesTableDrift(t *testing.T) {
	cases := map[string]map[string]string{ // integration → service → image
		"postgres":   {"postgres": "postgres:16-alpine"},
		"mysql":      {"mysql": "mysql:8"},
		"redis":      {"redis": "redis:7-alpine"},
		"rabbitmq":   {"rabbitmq": "rabbitmq:3-management-alpine"},
		"kafka":      {"kafka": "bitnami/kafka:3.7"},
		"nats":       {"nats": "nats:2-alpine"},
		"mongodb":    {"mongodb": "mongo:7"},
		"clickhouse": {"clickhouse": "clickhouse/clickhouse-server:24.3"},
		"couchbase":  {"couchbase": "couchbase:community-7.6.2"},
		"prometheus": {"prometheus": "prom/prometheus:latest"},
		"otel":       {"otel-collector": "otel/opentelemetry-collector:0.102.1"},
	}
	for integration, want := range cases {
		services, ok := DockerServicesByIntegration[integration]
		if !ok {
			t.Errorf("integration %q missing from central table", integration)
			continue
		}
		for svcName, image := range want {
			found := false
			for _, s := range services {
				if s.Service == svcName {
					found = true
					if s.Image != image {
						t.Errorf("%s/%s image = %q, want %q", integration, svcName, s.Image, image)
					}
				}
			}
			if !found {
				t.Errorf("%s: service %q missing from central table", integration, svcName)
			}
		}
	}
}

// TestPortConflictWarning (GIN-026): compose com "db" publicando 5432 +
// `add postgres` → warning no plano (não bloqueia).
func TestPortConflictWarning(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module demo\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	composeDir := filepath.Join(dir, "devops", "docker")
	_ = os.MkdirAll(composeDir, 0o755)
	compose := `version: "3.9"
services:
  db:
    image: postgres:15
    ports:
      - "5432:5432"
`
	_ = os.WriteFile(filepath.Join(composeDir, "docker-compose.yml"), []byte(compose), 0o644)

	prj, err := project.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Plan("postgres", prj, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	found := false
	for _, w := range p.Warnings {
		if strings.Contains(w, "port conflict") && strings.Contains(w, "5432") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected port conflict warning, got warnings: %v", p.Warnings)
	}
}

// TestNoPortConflictWarningWhenFree: sem colisão, sem warning.
func TestNoPortConflictWarningWhenFree(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module demo\n\ngo 1.22\n"), 0o644)
	composeDir := filepath.Join(dir, "devops", "docker")
	_ = os.MkdirAll(composeDir, 0o755)
	_ = os.WriteFile(filepath.Join(composeDir, "docker-compose.yml"),
		[]byte("version: \"3.9\"\nservices:\n  app:\n    image: demo\n"), 0o644)

	prj, err := project.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Plan("postgres", prj, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range p.Warnings {
		if strings.Contains(w, "port conflict") {
			t.Fatalf("unexpected conflict warning: %v", p.Warnings)
		}
	}
}

// Scaffold → redis → add postgres: sem falso positivo (redis não publica 5432).
func TestScaffoldAddPostgresNoFalseConflict(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := scaffold.NewProject("demo", "service"); err != nil {
		t.Fatal(err)
	}
	prj, err := project.Load(filepath.Join(dir, "demo"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := Plan("postgres", prj, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range p.Warnings {
		if strings.Contains(w, "port conflict") {
			t.Fatalf("false positive conflict: %v", p.Warnings)
		}
	}
}
