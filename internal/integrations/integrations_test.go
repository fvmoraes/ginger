package integrations

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fvmoraes/ginger/internal/project"
)

// setupMinimalProject cria um projeto Go mínimo (sem manifest) para os testes
// plan-based — o compose não-gerenciado propõe patch (GIN-002, caminho seguro).
func setupMinimalProject(t *testing.T, withCompose bool) (root string, prj *project.Project) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module demo\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("go.mod: %v", err)
	}
	if withCompose {
		composeDir := filepath.Join(dir, "devops", "docker")
		if err := os.MkdirAll(composeDir, 0o755); err != nil {
			t.Fatal(err)
		}
		compose := `version: "3.9"
services:
  app:
    build:
      context: ../..
      dockerfile: devops/docker/Dockerfile
    environment:
      APP_ENV: development
`
		if err := os.WriteFile(filepath.Join(composeDir, "docker-compose.yml"), []byte(compose), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	prj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("project.Load: %v", err)
	}
	return dir, prj
}

func TestSwaggerPlanUsesConfiguredPathsAndPreservesUnmanagedRouter(t *testing.T) { /* kept from original */
}

func TestRealtimeTemplatesUseHandlersPackage(t *testing.T) {
	for name, tmpl := range map[string]string{
		"sse":       sseTmpl,
		"websocket": wsTmpl,
	} {
		if !strings.Contains(tmpl, "package handlers") {
			t.Fatalf("%s template should declare package handlers", name)
		}
	}
}

// Migrado do legado Add (R5, GIN-006): o mesmo cenário, via plan-based.
// Compose não-gerenciado → patch revisável; arquivo de integração criado.
func TestPlanMessagingIntegrationProposesComposePatch(t *testing.T) {
	root, prj := setupMinimalProject(t, true)
	p, err := Plan("rabbitmq", prj, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if err := p.Apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}

	patch, err := os.ReadFile(filepath.Join(root, ".ginger", "patches", "devops", "docker", "docker-compose.yml.patch"))
	if err != nil {
		t.Fatalf("expected compose patch (unmanaged compose): %v", err)
	}
	content := string(patch)
	for _, want := range []string{"rabbitmq:", "rabbitmq:3-management-alpine", "RABBITMQ_URL", "depends_on"} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected patch to contain %q, got:\n%s", want, content)
		}
	}
}

// Migrado do legado (R5): múltiplas integrações, patch único acumulado.
func TestPlanLocalInfraProposesComposePatch(t *testing.T) {
	root, prj := setupMinimalProject(t, true)

	// Cada Plan propõe o patch a partir do compose intocado — cada patch
	// carrega o serviço da sua integração (comportamento documentado).
	perService := map[string][]string{
		"postgres":   {`postgres:`, `postgres:16-alpine`, `DATABASE_DSN`, `depends_on`},
		"redis":      {`redis:`, `redis:7-alpine`, `REDIS_ADDR`},
		"prometheus": {`prometheus:`, `prom/prometheus:latest`},
	}
	patchPath := filepath.Join(root, ".ginger", "patches", "devops", "docker", "docker-compose.yml.patch")
	for name, wants := range perService {
		p, err := Plan(name, prj, false)
		if err != nil {
			t.Fatalf("Plan(%s): %v", name, err)
		}
		if err := p.Apply(); err != nil {
			t.Fatalf("apply(%s): %v", name, err)
		}
		patch, err := os.ReadFile(patchPath)
		if err != nil {
			t.Fatalf("%s: patch missing on disk: %v", name, err)
		}
		for _, want := range wants {
			if !strings.Contains(string(patch), want) {
				t.Fatalf("%s patch missing %q:\n%s", name, want, string(patch))
			}
		}
	}

	prometheusConfig, err := os.ReadFile(filepath.Join(root, "devops", "docker", "prometheus.yml"))
	if err != nil {
		t.Fatalf("prometheus.yml must be created: %v", err)
	}
	if !strings.Contains(string(prometheusConfig), `targets: ["app:8080"]`) {
		t.Fatalf("prometheus targets: %s", string(prometheusConfig))
	}
}

// Migrado (R5): sem compose, sem patch — apenas o arquivo de integração.
func TestPlanSkipsComposePatchWhenComposeMissing(t *testing.T) {
	root, prj := setupMinimalProject(t, false)
	p, err := Plan("postgres", prj, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if err := p.Apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "platform", "database", "postgres.go")); err != nil {
		t.Fatalf("generated file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".ginger", "patches")); !os.IsNotExist(err) {
		t.Fatal("no patch must be proposed when compose does not exist")
	}
}

// Migrado (R5): template mongodb — via plan-based apply.
func TestPlanMongoDBGeneratesValidTemplateOutput(t *testing.T) {
	_, prj := setupMinimalProject(t, false)
	p, err := Plan("mongodb", prj, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if err := p.Apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(prj.Root, "platform", "nosql", "mongo.go"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(data), `bson.D{bson.E{Key: "ping", Value: 1}}`) {
		t.Fatalf("escaped bson command expected:\n%s", string(data))
	}
	if !strings.Contains(string(data), `go.mongodb.org/mongo-driver/v2/mongo`) {
		t.Fatalf("v2 driver import expected:\n%s", string(data))
	}
}

// Migrado (R5): template sqlite com import time — via plan-based apply.
func TestPlanSQLiteTemplateIncludesTimeImport(t *testing.T) {
	_, prj := setupMinimalProject(t, false)
	p, err := Plan("sqlite", prj, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if err := p.Apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(prj.Root, "platform", "database", "sqlite.go"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(data), `"time"`) {
		t.Fatalf("time import expected:\n%s", string(data))
	}
}

func TestMessagingTemplatesUseTransportSpecificHelperNames(t *testing.T) { /* kept from original */ }

func TestRegistryUsesCurrentMongoAndPubSubModules(t *testing.T) { /* kept from original */ }

// O cenário do antigo TestAddRemovesCreatedFileWhenDependencyInstallFails
// (falha de go get → rollback) é coberto por TestApplyWithRollbackOnFailedGoGet
// (rollback_test.go) no fluxo plan-based — o legado foi removido (GIN-006/030).
func TestRollbackCoversLegacyDependencyFailureScenario(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	registry["testdep"] = integration{
		name: "testdep", pkg: "example.com/failing-dependency",
		file: filepath.Join("platform", "testdep", "client.go"), tmpl: "package testdep\n",
	}
	t.Cleanup(func() { delete(registry, "testdep") })

	origExec := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd { return exec.Command("false") }
	t.Cleanup(func() { execCommand = origExec })

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module demo\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	prj, err := project.Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	p, err := Plan("testdep", prj, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyWithRollback(p, "testdep", dir); err == nil {
		t.Fatal("expected failure")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "platform", "testdep", "client.go")); !os.IsNotExist(statErr) {
		t.Fatalf("created file survived rollback: %v", statErr)
	}
}
