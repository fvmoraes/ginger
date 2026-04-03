// Package generator produces boilerplate Go files for handlers, services, ports, adapters, models and tests.
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
	"unicode/utf8"
)

// ErrFileExists is returned when a generate target already exists on disk.
var ErrFileExists = errors.New("file already exists")

type genData struct {
	FileName   string
	Name       string
	Slug       string
	NameTitle  string
	NamePlural string
	Module     string
}

func newData(name string) genData {
	tokens := splitNameTokens(name)
	identifier := strings.Join(tokens, "_")
	slug := strings.Join(tokens, "-")
	if identifier == "" {
		identifier = "resource"
	}
	if slug == "" {
		slug = "resource"
	}

	return genData{
		FileName:   identifier,
		Name:       identifier,
		Slug:       slug,
		NameTitle:  title(tokens),
		NamePlural: slug + "s",
		Module:     modulePath(),
	}
}

// Handler generates internal/api/handlers/<name>_handler.go.
func Handler(name string) error {
	data := newData(name)
	return generate(
		filepath.Join("internal", "api", "handlers", data.FileName+"_handler.go"),
		handlerTmpl,
		data,
	)
}

// Service generates internal/services/<name>_service.go.
func Service(name string) error {
	data := newData(name)
	return generate(
		filepath.Join("internal", "services", data.FileName+"_service.go"),
		serviceTmpl,
		data,
	)
}

// Repository generates internal/ports/<name>_repository.go.
func Repository(name string) error {
	data := newData(name)
	return generate(
		filepath.Join("internal", "ports", data.FileName+"_repository.go"),
		repositoryTmpl,
		data,
	)
}

// Adapter generates internal/adapters/<name>_memory_repository.go.
func Adapter(name string) error {
	data := newData(name)
	return generate(
		filepath.Join("internal", "adapters", data.FileName+"_memory_repository.go"),
		adapterTmpl,
		data,
	)
}

// Model generates internal/models/<name>.go.
func Model(name string) error {
	data := newData(name)
	return generate(
		filepath.Join("internal", "models", data.FileName+".go"),
		modelTmpl,
		data,
	)
}

// HandlerTest generates a handler test file.
func HandlerTest(name string) error {
	data := newData(name)
	return generate(
		filepath.Join("internal", "api", "handlers", data.FileName+"_handler_test.go"),
		handlerTestTmpl,
		data,
	)
}

// ServiceTest generates a service test file.
func ServiceTest(name string) error {
	data := newData(name)
	return generate(
		filepath.Join("internal", "services", data.FileName+"_service_test.go"),
		serviceTestTmpl,
		data,
	)
}

// RepositoryTest generates an adapter test file.
func RepositoryTest(name string) error {
	data := newData(name)
	return generate(
		filepath.Join("internal", "adapters", data.FileName+"_memory_repository_test.go"),
		repositoryTestTmpl,
		data,
	)
}

// IntegrationTest generates tests/integration/<name>_test.go.
func IntegrationTest(name string) error {
	data := newData(name)
	return generate(
		filepath.Join("tests", "integration", data.FileName+"_test.go"),
		integrationTestTmpl,
		data,
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

// Tests generates handler, service, and adapter tests for a given resource.
func Tests(name string) error {
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
}

// CRUD generates model + handler + service + port + adapter + integration test.
func CRUD(name string) error {
	fmt.Printf("\n  Generating CRUD for '%s'...\n\n", name)
	steps := []struct {
		label string
		fn    func(string) error
	}{
		{"model", Model},
		{"repository port", Repository},
		{"memory adapter", Adapter},
		{"service", Service},
		{"handler", Handler},
		{"integration test", IntegrationTest},
	}
	if err := generateMany(name, steps); err != nil {
		return err
	}
	fmt.Printf("\n  ✓ CRUD for '%s' generated.\n", name)
	return nil
}

// ProjectService generates a service scaffold for CLI or worker projects.
func ProjectService(name, projectType string) error {
	data := newData(name)

	switch projectType {
	case "cli":
		if err := generate(filepath.Join("internal", "services", data.FileName+".go"), cliServiceTmpl, data); err != nil {
			return err
		}
		if err := generate(filepath.Join("internal", "services", data.FileName+"_test.go"), cliServiceTestTmpl, data); err != nil {
			return err
		}
		return generate(filepath.Join("internal", "ports", data.FileName+".go"), cliServicePortTmpl, data)
	case "worker":
		if err := generate(filepath.Join("internal", "services", data.FileName+".go"), workerServiceTmpl, data); err != nil {
			return err
		}
		if err := generate(filepath.Join("internal", "services", data.FileName+"_test.go"), workerServiceTestTmpl, data); err != nil {
			return err
		}
		return generate(filepath.Join("internal", "ports", data.FileName+".go"), workerServicePortTmpl, data)
	default:
		return fmt.Errorf("project service generation is not supported for %s projects", projectType)
	}
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
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

func title(tokens []string) string {
	var b strings.Builder
	for _, token := range tokens {
		if token == "" {
			continue
		}
		r, size := utf8.DecodeRuneInString(token)
		if r == utf8.RuneError {
			continue
		}
		b.WriteRune(unicode.ToUpper(r))
		b.WriteString(token[size:])
	}
	if b.Len() == 0 {
		return "Resource"
	}
	return b.String()
}

func splitNameTokens(s string) []string {
	var tokens []string
	var current []rune

	flush := func() {
		if len(current) == 0 {
			return
		}
		tokens = append(tokens, strings.ToLower(string(current)))
		current = nil
	}

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current = append(current, r)
			continue
		}
		flush()
	}
	flush()

	if len(tokens) == 0 {
		return []string{"resource"}
	}
	return tokens
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
			path = filepath.Join("internal", "api", "handlers", newData(name).FileName+"_handler.go")
		case "service":
			path = filepath.Join("internal", "services", newData(name).FileName+"_service.go")
		case "repository":
			path = filepath.Join("internal", "adapters", newData(name).FileName+"_memory_repository.go")
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
