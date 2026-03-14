// Package scaffold generates a new Ginger project from templates.
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type projectData struct {
	Name   string
	Module string
	Type   string
}

// NewProject scaffolds a complete project at ./<name> for the given type.
// Supported types: api (default), microservice, cli, worker.
func NewProject(name, projectType string) error {
	switch projectType {
	case "api", "microservice", "cli", "worker":
	default:
		return fmt.Errorf("unknown project type: %s (api|microservice|cli|worker)", projectType)
	}

	data := projectData{Name: name, Module: name, Type: projectType}

	dirs := baseDirs(projectType)
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(name, d), 0755); err != nil {
			return fmt.Errorf("scaffold: mkdir %s: %w", d, err)
		}
	}

	files := baseFiles(projectType)
	for path, tmplStr := range files {
		if err := writeTemplate(filepath.Join(name, path), tmplStr, data); err != nil {
			return err
		}
	}

	fmt.Printf("  created %s/ (%s)\n", name, projectType)
	return nil
}

func baseDirs(t string) []string {
	common := []string{
		"configs",
		"scripts",
		"tests",
		"docs",
	}
	switch t {
	case "cli":
		return append(common, "cmd/root", "internal/config", "pkg")
	case "worker":
		return append(common, "cmd/worker", "internal/worker", "internal/config", "platform", "pkg")
	default: // api, microservice
		return append(common,
			"cmd/app",
			"internal/api/handlers",
			"internal/api/services",
			"internal/api/repositories",
			"internal/api/middlewares",
			"internal/models",
			"internal/config",
			"pkg",
			"platform",
		)
	}
}

func baseFiles(t string) map[string]string {
	common := map[string]string{
		"go.mod":           goModTmpl,
		"configs/app.yaml": appYamlTmpl,
		".env.example":     envExampleTmpl,
		"Makefile":         makefileTmpl,
		".gitignore":       gitignoreTmpl,
		"README.md":        readmeTmpl,
	}

	switch t {
	case "cli":
		common["cmd/root/main.go"] = cliMainTmpl
	case "worker":
		common["cmd/worker/main.go"] = workerMainTmpl
		common["internal/worker/worker.go"] = workerTmpl
		common["Dockerfile"] = dockerfileTmpl
	default: // api, microservice
		common["cmd/app/main.go"] = mainTmpl
		common["internal/config/config.go"] = internalConfigTmpl
		common["internal/api/handlers/health.go"] = healthHandlerTmpl
		common["Dockerfile"] = dockerfileTmpl
		common["docker-compose.yml"] = dockerComposeTmpl
		common["kubernetes/deployment.yaml"] = k8sDeploymentTmpl
		common["helm/Chart.yaml"] = helmChartTmpl
		common["helm/values.yaml"] = helmValuesTmpl
		common["helm/templates/deployment.yaml"] = helmDeploymentTmpl
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
		"title": strings.Title, //nolint:staticcheck
	}).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("scaffold: parse template: %w", err)
	}
	return tmpl.Execute(f, data)
}
