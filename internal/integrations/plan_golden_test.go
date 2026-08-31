package integrations

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fvmoraes/ginger/internal/project"
	"github.com/fvmoraes/ginger/internal/scaffold"
)

// Golden tests pinam o contrato do `ginger add --plan` (Fase 0, tarefa 8b —
// proteção anti-regressão). Diff intencional em um golden exige revisão e
// re-congelamento explícito (go test ./internal/integrations -update).
var updateGolden = flag.Bool("update", false, "rewrite golden files")

const customCompose = `# Custom compose with anchors, extension fields and networks (GIN-002 fixture)
# Nota: sem "build: ." — a forma abreviada causa falha dura de parse no
# planComposePatch (composeBuild é struct; see GIN-002 extensão no relatório).
x-common-env: &common-env
  APP_ENV: development

services:
  demo:
    image: demo:dev
    environment:
      <<: *common-env
    ports:
      - "8080:8080"

networks:
  custom-net:
    driver: bridge
`

func writeGolden(t *testing.T, path, got string) {
	t.Helper()
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run with -update to create): %v", err)
	}
	if got != string(want) {
		t.Fatalf("plan contract changed (intentional? re-run with -update and review the diff):\n--- got ---\n%s\n--- want ---\n%s", got, string(want))
	}
}

func setupServiceProject(t *testing.T) (root, origWd string) {
	t.Helper()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	if err := scaffold.NewProject("demo", "service"); err != nil {
		t.Fatalf("scaffold service project: %v", err)
	}
	return filepath.Join(dir, "demo"), origWd
}

func testPlanGolden(t *testing.T, integration string, customizeCompose bool, customGolden string) {
	t.Helper()
	root, origWd := setupServiceProject(t)

	composePath := filepath.Join(root, "devops", "docker", "docker-compose.yml")
	if customizeCompose {
		if err := os.WriteFile(composePath, []byte(customCompose), 0o644); err != nil {
			t.Fatalf("customize compose: %v", err)
		}
	}

	prj, err := project.Load(root)
	if err != nil {
		t.Fatalf("load project: %v", err)
	}

	p, err := Plan(integration, prj, false)
	if err != nil {
		t.Fatalf("Plan(%s): %v", integration, err)
	}

	var b strings.Builder
	for _, c := range p.Changes {
		rel, relErr := filepath.Rel(p.ProjectRoot, c.Path)
		if relErr != nil {
			t.Fatalf("rel path: %v", relErr)
		}
		b.WriteString(strings.TrimSpace(string(c.Type)) + " " + filepath.ToSlash(rel))
		if c.Reason != "" {
			b.WriteString(" (" + c.Reason + ")")
		}
		b.WriteString("\n")
	}
	for _, w := range p.Warnings {
		b.WriteString("warn: " + w + "\n")
	}
	for _, e := range p.Errors {
		b.WriteString("error: " + e + "\n")
	}

	// Golden path resolvido a partir do CWD original — o chdir do setup
	// aponta para o TempDir, e os goldens versionados vivem no repositório.
	writeGolden(t, filepath.Join(origWd, "testdata", "golden", customGolden+".golden"), b.String())
}

// Contrato atual: scaffold novo → compose gerenciado é modificado via plano.
func TestPlanGoldenRedisFreshService(t *testing.T) {
	testPlanGolden(t, "redis", false, "redis-fresh-service")
}

func TestPlanGoldenPostgresFreshService(t *testing.T) {
	testPlanGolden(t, "postgres", false, "postgres-fresh-service")
}

// GIN-002 CORRIGIDO (Fase 1): compose customizado → merge condicional por
// hash de proveniência → patch revisável em .ginger/patches/ e compose
// intocado. O golden anterior (ChangeModify lossy) foi substituído
// intencionalmente — ver evidence/regression-risk-validation.md.
func TestPlanGoldenRedisCustomizedComposePinsGIN002(t *testing.T) {
	testPlanGolden(t, "redis", true, "redis-customized-compose")
}
