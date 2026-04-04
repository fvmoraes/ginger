// Package scaffold generates a new Ginger project from templates.
package scaffold

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/fvmoraes/ginger/internal/buildinfo"
)

// ErrProjectExists is returned when the target project directory already exists.
var ErrProjectExists = errors.New("project directory already exists")

// ErrInvalidProjectName is returned when the requested project name is not a simple slug.
var ErrInvalidProjectName = errors.New("invalid project name")

type projectData struct {
	Name          string
	Module        string
	Type          string
	CmdDir        string
	GoVersion     string
	GingerVersion string
	UsesGinger    bool
}

var goVersionOutput = func() ([]byte, error) {
	return exec.Command("go", "version").Output()
}

var gingerVersion = buildinfo.Version

// CmdDir returns the cmd subdirectory for a given project name and type.
//
//	generic  → cmd/<name>
//	service  → cmd/<name>
//	worker   → cmd/<name>-worker
//	cli      → cmd/<name>
func CmdDir(name, projectType string) string {
	switch projectType {
	case "worker":
		return "cmd/" + name + "-worker"
	default: // generic, service, cli
		return "cmd/" + name
	}
}

// NewProject scaffolds a complete project at ./<name> for the given type.
// Supported types: generic (default), service, cli, worker.
func NewProject(name, projectType string) error {
	switch projectType {
	case "generic", "service", "cli", "worker":
	default:
		return fmt.Errorf("unknown project type %q — use: generic, service, cli, worker", projectType)
	}

	if err := validateProjectName(name); err != nil {
		return err
	}

	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("%w: %s", ErrProjectExists, name)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("scaffold: stat %s: %w", name, err)
	}

	cmdDir := CmdDir(name, projectType)
	usesGinger := projectType == "service" || projectType == "worker"
	data := projectData{
		Name:          name,
		Module:        name,
		Type:          projectType,
		CmdDir:        cmdDir,
		GoVersion:     detectGoVersion(),
		GingerVersion: gingerVersion(),
		UsesGinger:    usesGinger,
	}

	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("scaffold: mkdir %s: %w", name, err)
	}

	files := baseFiles(data)
	for path, tmplStr := range files {
		if err := writeTemplate(filepath.Join(name, path), tmplStr, data); err != nil {
			return err
		}
	}

	fmt.Printf("  created %s/ (%s → %s)\n", name, projectType, cmdDir)
	return nil
}

const minGoVersion = "1.25"

// detectGoVersion returns the local Go version if >= minGoVersion, otherwise minGoVersion.
func detectGoVersion() string {
	out, err := goVersionOutput()
	if err != nil {
		return minGoVersion
	}
	return resolveGoVersion(strings.TrimSpace(string(out)))
}

// resolveGoVersion parses "go version go1.X.Y ..." and returns the major.minor,
// clamped to minGoVersion from below.
func resolveGoVersion(goVersionOutput string) string {
	// "go version go1.25.0 darwin/amd64" → "1.25"
	parts := strings.Fields(goVersionOutput)
	for _, p := range parts {
		ver := strings.TrimPrefix(p, "go")
		if ver == p {
			continue
		}
		segs := strings.SplitN(ver, ".", 3)
		if len(segs) < 2 {
			continue
		}
		major, err1 := strconv.Atoi(segs[0])
		minor, err2 := strconv.Atoi(segs[1])
		if err1 != nil || err2 != nil {
			continue
		}
		minSegs := strings.SplitN(minGoVersion, ".", 2)
		minMajor, _ := strconv.Atoi(minSegs[0])
		minMinor, _ := strconv.Atoi(minSegs[1])
		if major < minMajor || (major == minMajor && minor < minMinor) {
			return minGoVersion
		}
		return fmt.Sprintf("%d.%d", major, minor)
	}
	return minGoVersion
}

func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidProjectName)
	}
	if filepath.Base(name) != name {
		return fmt.Errorf("%w: use a simple directory name, not a path: %s", ErrInvalidProjectName, name)
	}
	if strings.ContainsAny(name, `\/`) {
		return fmt.Errorf("%w: path separators are not allowed: %s", ErrInvalidProjectName, name)
	}
	if strings.Contains(name, " ") {
		return fmt.Errorf("%w: spaces are not allowed: %s", ErrInvalidProjectName, name)
	}
	return nil
}

func baseFiles(d projectData) map[string]string {
	common := map[string]string{
		"go.mod":     goModTmpl,
		"Makefile":   makefileTmpl,
		".gitignore": gitignoreTmpl,
		"README.md":  readmeTmpl,
	}

	mainPath := d.CmdDir + "/main.go"
	switch d.Type {
	case "cli":
		common[mainPath] = cliAppMainTmpl
		common["internal/commands/root.go"] = cliRootCommandTmpl
		common["internal/commands/version.go"] = cliVersionCommandTmpl
		common["internal/ports/ports.go"] = cliPortsTmpl
		common["internal/adapters/filesystem.go"] = cliFilesystemAdapterTmpl
		common["internal/config/config.go"] = cliConfigTmpl
		common["pkg/output/formatter.go"] = cliOutputFormatterTmpl
		common[".goreleaser.yaml"] = goreleaserTmpl
		common[".editorconfig"] = editorconfigTmpl
	case "worker":
		common["configs/app.yaml"] = appYamlTmpl
		common[".env.example"] = envExampleTmpl
		common[mainPath] = workerMainTmpl
		common["internal/config/config.go"] = internalConfigTmpl
		common["internal/worker/worker.go"] = workerTmpl
		common["internal/worker/handler.go"] = workerHandlerTmpl
		common["internal/ports/ports.go"] = workerPortsTmpl
		common["internal/adapters/memory_consumer.go"] = workerMemoryConsumerTmpl
		common["internal/services/processor.go"] = workerProcessorTmpl
		common["tests/integration/worker_test.go"] = workerIntegrationTestTmpl
		common["devops/docker/Dockerfile"] = dockerfileTmpl
		common["devops/docker/docker-compose.yml"] = workerDockerComposeTmpl
		common["devops/kubernetes/deployment.yaml"] = k8sDeploymentTmpl
		common["devops/helm/Chart.yaml"] = helmChartTmpl
		common["devops/helm/values.yaml"] = helmValuesTmpl
		common["devops/helm/templates/deployment.yaml"] = helmDeploymentTmpl
		common["devops/pipelines/ci.yaml"] = pipelineTmpl
		common[".editorconfig"] = editorconfigTmpl
	case "generic":
		common[mainPath] = cliMainTmpl
		common["internal/.gitkeep"] = ""
		common[".editorconfig"] = editorconfigTmpl
	default: // service
		common["configs/app.yaml"] = appYamlTmpl
		common[".env.example"] = envExampleTmpl
		common[mainPath] = mainTmpl
		common["internal/config/config.go"] = internalConfigTmpl
		common["internal/api/handlers/health.go"] = healthHandlerTmpl
		common["internal/api/router.go"] = serviceRouterTmpl
		common["internal/api/middlewares/request_id.go"] = serviceRequestIDMiddlewareTmpl
		common["internal/ports/ports.go"] = servicePortsTmpl
		common["internal/adapters/memory_store.go"] = serviceMemoryStoreTmpl
		common["internal/models/.gitkeep"] = ""
		common["migrations/.gitkeep"] = ""
		common["tests/integration/health_test.go"] = serviceHealthTestTmpl
		common["devops/docker/Dockerfile"] = dockerfileTmpl
		common["devops/docker/docker-compose.yml"] = dockerComposeTmpl
		common["devops/kubernetes/deployment.yaml"] = k8sDeploymentTmpl
		common["devops/helm/Chart.yaml"] = helmChartTmpl
		common["devops/helm/values.yaml"] = helmValuesTmpl
		common["devops/helm/templates/deployment.yaml"] = helmDeploymentTmpl
		common["devops/pipelines/ci.yaml"] = pipelineTmpl
		common[".editorconfig"] = editorconfigTmpl
	}
	return common
}

func writeTemplate(path, tmplStr string, data any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("scaffold: create %s: %w", path, err)
	}
	defer f.Close()

	tmpl, err := template.New("").Funcs(template.FuncMap{
		// titleCase capitalizes the first letter of a string.
		// Replaces the deprecated strings.Title.
		"title": titleCase,
	}).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("scaffold: parse template: %w", err)
	}
	return tmpl.Execute(f, data)
}

// titleCase returns s with the first Unicode letter uppercased.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 32
	}
	return string(r)
}
