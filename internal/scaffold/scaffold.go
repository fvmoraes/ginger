// Package scaffold generates a new Ginger project from templates.
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

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

	cmdDir := CmdDir(name, projectType)
	data := projectData{Name: name, Module: name, Type: projectType, CmdDir: cmdDir}

	dirs := baseDirs(data)
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(name, d), 0755); err != nil {
			return fmt.Errorf("scaffold: mkdir %s: %w", d, err)
		}
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

func baseDirs(d projectData) []string {
	common := []string{
		"configs",
		"scripts",
		"tests",
		"docs",
	}
	switch d.Type {
	case "cli":
		return append(common, d.CmdDir, "internal/config", "pkg")
	case "worker":
		return append(common, d.CmdDir, "internal/worker", "internal/config", "platform", "pkg")
	case "generic":
		return append(common, d.CmdDir, "internal/config", "pkg")
	default: // api, service
		return append(common,
			d.CmdDir,
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

func baseFiles(d projectData) map[string]string {
	common := map[string]string{
		"go.mod":           goModTmpl,
		"configs/app.yaml": appYamlTmpl,
		".env.example":     envExampleTmpl,
		"Makefile":         makefileTmpl,
		".gitignore":       gitignoreTmpl,
		"README.md":        readmeTmpl,
	}

	mainPath := d.CmdDir + "/main.go"
	switch d.Type {
	case "cli":
		common[mainPath] = cliMainTmpl
	case "worker":
		common[mainPath] = workerMainTmpl
		common["internal/worker/worker.go"] = workerTmpl
		common["Dockerfile"] = dockerfileTmpl
	case "generic":
		common[mainPath] = cliMainTmpl
	default: // api, service
		common[mainPath] = mainTmpl
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
