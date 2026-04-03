// Package scaffold generates a new Ginger project from templates.
package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ErrProjectExists is returned when the target project directory already exists.
var ErrProjectExists = errors.New("project directory already exists")

// ErrInvalidProjectName is returned when the requested project name is not a simple slug.
var ErrInvalidProjectName = errors.New("invalid project name")

type projectData struct {
	Name   string
	Module string
	Type   string
	CmdDir string
}

// CmdDir returns the cmd subdirectory for a given project name and type.
//
//	generic  → cmd/<name>
//	api      → cmd/<name>-api
//	service  → cmd/<name>-service
//	worker   → cmd/<name>-worker
//	cli      → cmd/<name>-cli
func CmdDir(name, projectType string) string {
	switch projectType {
	case "api", "service", "worker", "cli":
		return "cmd/" + name + "-" + projectType
	default: // generic
		return "cmd/" + name
	}
}

// NewProject scaffolds a complete project at ./<name> for the given type.
// Supported types: generic (default), api, service, cli, worker.
func NewProject(name, projectType string) error {
	switch projectType {
	case "generic", "api", "service", "cli", "worker":
	default:
		return fmt.Errorf("unknown project type %q — use: generic, api, service, cli, worker", projectType)
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
	data := projectData{Name: name, Module: name, Type: projectType, CmdDir: cmdDir}

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
		common[mainPath] = cliMainTmpl
	case "worker":
		common["configs/app.yaml"] = appYamlTmpl
		common[".env.example"] = envExampleTmpl
		common[mainPath] = workerMainTmpl
		common["internal/worker/worker.go"] = workerTmpl
		common["devops/docker/Dockerfile"] = dockerfileTmpl
		common["devops/pipelines/ci.yaml"] = pipelineTmpl
	case "generic":
		common[mainPath] = cliMainTmpl
	default: // api, service
		common["configs/app.yaml"] = appYamlTmpl
		common[".env.example"] = envExampleTmpl
		common[mainPath] = mainTmpl
		common["internal/config/config.go"] = internalConfigTmpl
		common["internal/api/handlers/health.go"] = healthHandlerTmpl
		common["devops/docker/Dockerfile"] = dockerfileTmpl
		common["devops/docker/docker-compose.yml"] = dockerComposeTmpl
		common["devops/docker/prometheus.yml"] = prometheusConfigTmpl
		common["devops/kubernetes/deployment.yaml"] = k8sDeploymentTmpl
		common["devops/helm/Chart.yaml"] = helmChartTmpl
		common["devops/helm/values.yaml"] = helmValuesTmpl
		common["devops/helm/templates/deployment.yaml"] = helmDeploymentTmpl
		common["devops/pipelines/ci.yaml"] = pipelineTmpl
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
