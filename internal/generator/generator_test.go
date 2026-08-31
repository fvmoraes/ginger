package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fvmoraes/ginger/internal/plan"
	"github.com/fvmoraes/ginger/internal/project"
)

// loadTestProject cria um projeto mínimo em cwd (para plan-based BuildPlan).
// Para o gerador crud, cria também o router Ginger com a região routes
// (requireGingerRouter exige generatedRouteRegistrars).
func loadTestProject(t *testing.T) (*project.Project, error) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(wd, "go.mod"), []byte("module example.com/test\n\ngo 1.25\n"), 0o644); err != nil {
		return nil, err
	}

	apiDir := filepath.Join(wd, "internal", "api")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		return nil, err
	}
	routerSource := `package api

import "github.com/fvmoraes/ginger/pkg/router"

var generatedRouteRegistrars []func(*router.Router)

func Register(r *router.Router) {
	// ginger:begin routes
	// ginger:end routes
}
`
	if err := os.WriteFile(filepath.Join(apiDir, "router.go"), []byte(routerSource), 0o644); err != nil {
		return nil, err
	}

	return project.Load(wd)
}

func TestNewDataNormalizesKebabCase(t *testing.T) {
	data := newData("order-processor")

	if data.FileName != "order_processor" {
		t.Fatalf("expected FileName order_processor, got %q", data.FileName)
	}
	if data.Name != "order_processor" {
		t.Fatalf("expected Name order_processor, got %q", data.Name)
	}
	if data.Slug != "order-processor" {
		t.Fatalf("expected Slug order-processor, got %q", data.Slug)
	}
	if data.NameTitle != "OrderProcessor" {
		t.Fatalf("expected NameTitle OrderProcessor, got %q", data.NameTitle)
	}
	if data.NamePlural != "order-processors" {
		t.Fatalf("expected NamePlural order-processors, got %q", data.NamePlural)
	}
}

func TestBuildPlanRespectsConfiguredStructureWithoutWriting(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "go.mod"), "module example.com/custom\n\ngo 1.25\n")
	mustWriteFile(t, filepath.Join(dir, "ginger.yaml"), `project:
  type: service
  root: .
structure:
  api: internal/transport/http
  handlers: internal/transport/http/endpoints
  models: internal/domain
  services: internal/usecases
  ports: internal/contracts
  adapters: internal/storage
  config: internal/settings
  tests: test
`)
	mustMkdirAll(t, filepath.Join(dir, "internal", "transport", "http"))
	mustWriteFile(t, filepath.Join(dir, "internal", "transport", "http", "router.go"), "package httpapi\n\nvar generatedRouteRegistrars []func()\n")

	prj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	p, err := BuildPlan(prj, "crud", "user", false)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if p.HasErrors() {
		t.Fatalf("unexpected plan errors: %v", p.Errors)
	}
	want := filepath.Join(dir, "internal", "transport", "http", "endpoints", "user_handler.go")
	found := false
	for _, change := range p.Changes {
		if change.Path == want && change.Type == plan.ChangeCreate {
			found = strings.Contains(string(change.Content), `models "example.com/custom/internal/domain"`)
		}
	}
	if !found {
		t.Fatalf("configured handler path/import not found in plan: %+v", p.Changes)
	}
	if _, err := os.Stat(want); !os.IsNotExist(err) {
		t.Fatalf("planning wrote a target file, stat err=%v", err)
	}
}

func TestBuildScanTestsPlanPreservesExistingTests(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "go.mod"), "module example.com/existing\n\ngo 1.22\n")
	for _, path := range []string{"internal/httpapi/handlers", "internal/core", "internal/store"} {
		mustMkdirAll(t, filepath.Join(dir, path))
	}
	mustWriteFile(t, filepath.Join(dir, "internal/httpapi/handlers/user.go"), "package handlers\n")
	mustWriteFile(t, filepath.Join(dir, "internal/httpapi/handlers/user_test.go"), "package handlers\n")
	mustWriteFile(t, filepath.Join(dir, "internal/core/user_service.go"), "package core\n")
	mustWriteFile(t, filepath.Join(dir, "internal/store/user_repository.go"), "package store\n")
	prj, err := project.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	p, err := BuildScanTestsPlan(prj, false)
	if err != nil {
		t.Fatalf("BuildScanTestsPlan: %v", err)
	}
	var skipped, creates int
	for _, change := range p.Changes {
		switch change.Type {
		case plan.ChangeSkip:
			skipped++
		case plan.ChangeCreate:
			if strings.HasSuffix(change.Path, "_test.go") {
				creates++
			}
		}
	}
	if skipped != 1 || creates != 2 {
		t.Fatalf("expected one preserved test and two new tests, got skipped=%d creates=%d changes=%+v", skipped, creates, p.Changes)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
}

func TestCRUDDoesNotGenerateTests(t *testing.T) {
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

	goMod := "module example.com/test\n\ngo 1.25\n"
	if err := os.WriteFile("go.mod", []byte(goMod), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	prj, err := loadTestProject(t)
	if err != nil {
		t.Fatal(err)
	}
	p, err := BuildPlan(prj, "crud", "user", false)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if err := p.Apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}

	expectedFiles := []string{
		filepath.Join("internal", "models", "user.go"),
		filepath.Join("internal", "api", "user_routes.go"),
		filepath.Join("internal", "api", "handlers", "user_handler.go"),
		filepath.Join("internal", "services", "user_service.go"),
		filepath.Join("internal", "ports", "user_repository.go"),
		filepath.Join("internal", "adapters", "user_memory_repository.go"),
		filepath.Join("tests", "integration", "user_test.go"),
	}
	for _, path := range expectedFiles {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected file %s to exist: %v", path, err)
		}
	}

	adapterSource, err := os.ReadFile(filepath.Join("internal", "adapters", "user_memory_repository.go"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(adapterSource), "apperrors.NotFound") {
		t.Fatalf("expected generated adapter to map missing resources to apperrors.NotFound")
	}

	unexpectedFiles := []string{
		filepath.Join("internal", "api", "handlers", "user_handler_test.go"),
		filepath.Join("internal", "services", "user_service_test.go"),
		filepath.Join("internal", "adapters", "user_memory_repository_test.go"),
	}
	for _, path := range unexpectedFiles {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected test file %s to be absent, got err=%v", path, err)
		}
	}
}

func TestTestsGeneratesFullResourceSuite(t *testing.T) {
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

	goMod := "module example.com/test\n\ngo 1.25\n"
	if err := os.WriteFile("go.mod", []byte(goMod), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	prj, err := loadTestProject(t)
	if err != nil {
		t.Fatal(err)
	}
	p, err := BuildPlan(prj, "crud", "user", false)
	if err != nil {
		t.Fatalf("BuildPlan crud: %v", err)
	}
	if err := p.Apply(); err != nil {
		t.Fatalf("apply crud: %v", err)
	}
	tp, err := BuildPlan(prj, "tests", "user", false)
	if err != nil {
		t.Fatalf("BuildPlan tests: %v", err)
	}
	if err := tp.Apply(); err != nil {
		t.Fatalf("apply tests: %v", err)
	}

	expectedFiles := []string{
		filepath.Join("internal", "api", "handlers", "user_handler_test.go"),
		filepath.Join("internal", "services", "user_service_test.go"),
		filepath.Join("internal", "adapters", "user_memory_repository_test.go"),
	}
	for _, path := range expectedFiles {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected file %s to exist: %v", path, err)
		}
	}
}

func TestCommandGeneratorCreatesFailingStub(t *testing.T) {
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

	wd2, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd2) }()
	prj, err := loadTestProject(t)
	if err != nil {
		t.Fatal(err)
	}
	p, err := BuildPlan(prj, "command", "sync", false)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if err := p.Apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("internal", "commands", "sync.go"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, `return fmt.Errorf("sync: not yet implemented")`) {
		t.Fatalf("expected generated command to return a non-zero error, got %s", content)
	}
}
