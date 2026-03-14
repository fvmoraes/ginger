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
}

// NewProject scaffolds a complete project at ./<name>.
func NewProject(name string) error {
	module := name // user can edit go.mod after
	data := projectData{Name: name, Module: module}

	dirs := []string{
		"cmd/app",
		"internal/api/handlers",
		"internal/api/services",
		"internal/api/repositories",
		"internal/api/middlewares",
		"internal/models",
		"internal/config",
		"pkg",
		"platform",
		"configs",
		"scripts",
		"tests",
		"docs",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(name, d), 0755); err != nil {
			return fmt.Errorf("scaffold: mkdir %s: %w", d, err)
		}
	}

	files := map[string]string{
		"go.mod":                          goModTmpl,
		"cmd/app/main.go":                 mainTmpl,
		"internal/config/config.go":       internalConfigTmpl,
		"internal/api/handlers/health.go": healthHandlerTmpl,
		"configs/app.yaml":                appYamlTmpl,
		".env.example":                    envExampleTmpl,
		"Dockerfile":                      dockerfileTmpl,
		"Makefile":                        makefileTmpl,
		".gitignore":                      gitignoreTmpl,
		"README.md":                       readmeTmpl,
	}

	for path, tmplStr := range files {
		if err := writeTemplate(filepath.Join(name, path), tmplStr, data); err != nil {
			return err
		}
	}

	fmt.Printf("  created %s/\n", name)
	return nil
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
