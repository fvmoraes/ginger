package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewProjectRejectsExistingDirectory(t *testing.T) {
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

	projectDir := filepath.Join(tmp, "demo")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = NewProject("demo", "service")
	if !errors.Is(err, ErrProjectExists) {
		t.Fatalf("expected ErrProjectExists, got %v", err)
	}
}

func TestNewProjectRejectsPathLikeName(t *testing.T) {
	err := NewProject(filepath.Join("tmp", "demo"), "service")
	if !errors.Is(err, ErrInvalidProjectName) {
		t.Fatalf("expected ErrInvalidProjectName, got %v", err)
	}
}

func TestResolveGoVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "version above minimum",
			input: "go version go1.26.0 darwin/amd64",
			want:  "1.26",
		},
		{
			name:  "version equal to minimum",
			input: "go version go1.25.0 linux/amd64",
			want:  "1.25",
		},
		{
			name:  "version below minimum",
			input: "go version go1.21.5 linux/amd64",
			want:  "1.25",
		},
		{
			name:  "detection failure empty string",
			input: "",
			want:  "1.25",
		},
		{
			name:  "detection failure garbage output",
			input: "not a go version string",
			want:  "1.25",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveGoVersion(tc.input)
			if got != tc.want {
				t.Errorf("resolveGoVersion(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestDetectGoVersionFallsBackOnCommandError(t *testing.T) {
	original := goVersionOutput
	goVersionOutput = func() ([]byte, error) {
		return nil, errors.New("boom")
	}
	defer func() { goVersionOutput = original }()

	if got := detectGoVersion(); got != minGoVersion {
		t.Fatalf("detectGoVersion() = %q, want %q", got, minGoVersion)
	}
}

func TestNewProjectWorkerIncludesConfigLoader(t *testing.T) {
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

	if err := NewProject("demo", "worker"); err != nil {
		t.Fatalf("NewProject returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("demo", "internal", "config", "config.go"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(data), `gingercfg.Load("configs/app.yaml")`) {
		t.Fatalf("expected worker scaffold config loader, got %s", string(data))
	}
}

func TestNewProjectServicePinsStableGingerVersionInGoMod(t *testing.T) {
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

	originalVersion := gingerVersion
	gingerVersion = func() string { return "9.9.9" }
	defer func() { gingerVersion = originalVersion }()

	if err := NewProject("demo", "service"); err != nil {
		t.Fatalf("NewProject returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("demo", "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if !strings.Contains(string(data), "require github.com/fvmoraes/ginger v9.9.9") {
		t.Fatalf("expected pinned Ginger version in go.mod, got %s", string(data))
	}
}

func TestNewProjectServiceRouterIncludesPingAndGeneratedRegistrars(t *testing.T) {
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

	if err := NewProject("demo", "service"); err != nil {
		t.Fatalf("NewProject returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("demo", "internal", "api", "router.go"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	content := string(data)
	for _, want := range []string{
		`generatedRouteRegistrars`,
		`registerGeneratedRoutes(v1)`,
		`v1.GET("/ping"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected service router to contain %q, got %s", want, content)
		}
	}
}

func TestNewProjectServiceComposeStartsMinimal(t *testing.T) {
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

	if err := NewProject("demo", "service"); err != nil {
		t.Fatalf("NewProject returned error: %v", err)
	}

	composeData, err := os.ReadFile(filepath.Join("demo", "devops", "docker", "docker-compose.yml"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	compose := string(composeData)
	for _, want := range []string{
		`services:`,
		`demo:`,
		`APP_ENV: development`,
		`HTTP_PORT: 8080`,
	} {
		if !strings.Contains(compose, want) {
			t.Fatalf("expected minimal service compose to contain %q, got %s", want, compose)
		}
	}

	for _, unwanted := range []string{
		`postgres:`,
		`redis:`,
		`prometheus:`,
		`grafana:`,
		`DATABASE_DSN`,
		`depends_on`,
	} {
		if strings.Contains(compose, unwanted) {
			t.Fatalf("expected minimal service compose to omit %q, got %s", unwanted, compose)
		}
	}

	if _, err := os.Stat(filepath.Join("demo", "devops", "docker", "prometheus.yml")); !os.IsNotExist(err) {
		t.Fatalf("expected prometheus config to be absent by default, stat err=%v", err)
	}

	envData, err := os.ReadFile(filepath.Join("demo", ".env.example"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	envExample := string(envData)
	for _, unwanted := range []string{
		`DATABASE_DRIVER`,
		`DATABASE_DSN`,
	} {
		if strings.Contains(envExample, unwanted) {
			t.Fatalf("expected minimal env example to omit %q, got %s", unwanted, envExample)
		}
	}
}

func TestNewProjectWorkerUsesStructuredLifecycle(t *testing.T) {
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

	if err := NewProject("demo", "worker"); err != nil {
		t.Fatalf("NewProject returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("demo", "cmd", "demo-worker", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	content := string(data)
	for _, want := range []string{
		`logger.New(cfg.Log.Level, cfg.Log.Format)`,
		`health.New()`,
		`worker_health_started`,
		`shutdown_signal_received`,
		`server.Shutdown(shutdownCtx)`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected worker main to contain %q, got %s", want, content)
		}
	}
}
