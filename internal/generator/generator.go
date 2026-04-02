// Package generator produces boilerplate Go files for handlers, services, repositories, models and tests.
package generator

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

// ErrFileExists is returned when a generate target already exists on disk.
var ErrFileExists = errors.New("file already exists")

type genData struct {
	Name       string // e.g. "user"
	NameTitle  string // e.g. "User"
	NamePlural string // e.g. "users"
	Module     string // e.g. "github.com/acme/foo"
}

func newData(name string) genData {
	return genData{
		Name:       strings.ToLower(name),
		NameTitle:  title(name),
		NamePlural: strings.ToLower(name) + "s",
		Module:     modulePath(),
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

// Model generates internal/models/<name>.go
func Model(name string) error {
	return generate(
		filepath.Join("internal", "models", strings.ToLower(name)+".go"),
		modelTmpl,
		newData(name),
	)
}

// HandlerTest generates a handler test file.
func HandlerTest(name string) error {
	return generate(
		filepath.Join("internal", "api", "handlers", strings.ToLower(name)+"_handler_test.go"),
		handlerTestTmpl,
		newData(name),
	)
}

// ServiceTest generates a service test file.
func ServiceTest(name string) error {
	return generate(
		filepath.Join("internal", "api", "services", strings.ToLower(name)+"_service_test.go"),
		serviceTestTmpl,
		newData(name),
	)
}

// RepositoryTest generates a repository test file.
func RepositoryTest(name string) error {
	return generate(
		filepath.Join("internal", "api", "repositories", strings.ToLower(name)+"_repository_test.go"),
		repositoryTestTmpl,
		newData(name),
	)
}

// AppTest generates a basic application smoke test under tests/integration.
func AppTest() error {
	return generate(
		filepath.Join("tests", "integration", "app_smoke_test.go"),
		appTestTmpl,
		genData{Module: modulePath()},
	)
}

// Swagger generates docs/openapi.json with a starter OpenAPI document.
// If name is provided, it generates CRUD examples for that resource.
func Swagger(name string) error {
	data := genData{}
	if name != "" {
		data = newData(name)
	}

	return generate(
		filepath.Join("docs", "openapi.json"),
		openAPITmpl,
		data,
	)
}

// Tests generates a test suite for a given resource and scope.
// Supported scopes: handler, service, repository, unit, all.
func Tests(name, scope string) error {
	switch scope {
	case "", "unit":
		if err := requireGeneratedResource(name, "handler", "service", "repository"); err != nil {
			return err
		}
		return generateMany(name, []struct {
			label string
			fn    func(string) error
		}{
			{"handler test", HandlerTest},
			{"service test", ServiceTest},
			{"repository test", RepositoryTest},
		})
	case "handler":
		if err := requireGeneratedResource(name, "handler"); err != nil {
			return err
		}
		return HandlerTest(name)
	case "service":
		if err := requireGeneratedResource(name, "service"); err != nil {
			return err
		}
		return ServiceTest(name)
	case "repository", "repo":
		if err := requireGeneratedResource(name, "repository"); err != nil {
			return err
		}
		return RepositoryTest(name)
	case "all":
		if err := requireGeneratedResource(name, "handler", "service", "repository"); err != nil {
			return err
		}
		if err := generateMany(name, []struct {
			label string
			fn    func(string) error
		}{
			{"handler test", HandlerTest},
			{"service test", ServiceTest},
			{"repository test", RepositoryTest},
		}); err != nil {
			return err
		}
		return AppTest()
	default:
		return fmt.Errorf("unknown test scope: %s", scope)
	}
}

// CRUD generates model + handler + service + repository + tests for a given name.
func CRUD(name string) error {
	fmt.Printf("\n  Generating CRUD for '%s'...\n\n", name)
	steps := []struct {
		label string
		fn    func(string) error
	}{
		{"model", Model},
		{"repository", Repository},
		{"service", Service},
		{"handler", Handler},
		{"handler test", HandlerTest},
		{"service test", ServiceTest},
		{"repository test", RepositoryTest},
	}
	if err := generateMany(name, steps); err != nil {
		return err
	}
	fmt.Printf("\n  ✓ CRUD for '%s' generated. Wire it up in your router!\n", name)
	return nil
}

func generateMany(name string, steps []struct {
	label string
	fn    func(string) error
}) error {
	for _, s := range steps {
		if err := s.fn(name); err != nil {
			return fmt.Errorf("%s: %w", s.label, err)
		}
	}
	return nil
}

func generate(path, tmplStr string, data genData) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %s", ErrFileExists, path)
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

func modulePath() string {
	f, err := os.Open("go.mod")
	if err != nil {
		return "yourmodule"
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}

	return "yourmodule"
}

func requireGeneratedResource(name string, kinds ...string) error {
	var missing []string

	for _, kind := range kinds {
		var path string
		switch kind {
		case "handler":
			path = filepath.Join("internal", "api", "handlers", strings.ToLower(name)+"_handler.go")
		case "service":
			path = filepath.Join("internal", "api", "services", strings.ToLower(name)+"_service.go")
		case "repository":
			path = filepath.Join("internal", "api", "repositories", strings.ToLower(name)+"_repository.go")
		default:
			continue
		}

		if _, err := os.Stat(path); err != nil {
			missing = append(missing, path)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("generate the resource first; missing: %s", strings.Join(missing, ", "))
}
