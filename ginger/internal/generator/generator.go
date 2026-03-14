// Package generator produces boilerplate Go files for handlers, services, and repositories.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

type genData struct {
	Name       string // e.g. "user"
	NameTitle  string // e.g. "User"
	NamePlural string // e.g. "users"
}

func newData(name string) genData {
	return genData{
		Name:       strings.ToLower(name),
		NameTitle:  title(name),
		NamePlural: strings.ToLower(name) + "s",
	}
}

// Handler generates internal/api/handlers/<name>_handler.go
func Handler(name string) error {
	return generate(
		filepath.Join("internal", "api", "handlers", strings.ToLower(name)+"_handler.go"),
		handlerTmpl,
		newData(name),
	)
}

// Service generates internal/api/services/<name>_service.go
func Service(name string) error {
	return generate(
		filepath.Join("internal", "api", "services", strings.ToLower(name)+"_service.go"),
		serviceTmpl,
		newData(name),
	)
}

// Repository generates internal/api/repositories/<name>_repository.go
func Repository(name string) error {
	return generate(
		filepath.Join("internal", "api", "repositories", strings.ToLower(name)+"_repository.go"),
		repositoryTmpl,
		newData(name),
	)
}

func generate(path, tmplStr string, data genData) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists: %s", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("generator: create %s: %w", path, err)
	}
	defer f.Close()

	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(f, data); err != nil {
		return err
	}
	fmt.Printf("  ✓ created %s\n", path)
	return nil
}

func title(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
