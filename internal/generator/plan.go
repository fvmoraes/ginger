package generator

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/fvmoraes/ginger/internal/manifest"
	"github.com/fvmoraes/ginger/internal/plan"
	"github.com/fvmoraes/ginger/internal/project"
)

type plannedTemplate struct {
	key      string
	file     string
	template string
}

// BuildPlan creates a real, root-aware generation plan. It is the path used by
// the CLI; legacy exported generator functions remain for API compatibility.
func BuildPlan(prj *project.Project, kind, name string, force bool) (*plan.Plan, error) {
	p := plan.New(prj.Root)
	p.CreateMissingDirs = prj.YAML.Rules.CreateMissingDirs
	data, err := projectData(prj, name)
	if err != nil {
		return nil, err
	}

	var templates []plannedTemplate
	switch kind {
	case "crud", "c":
		if err := requireGingerRouter(prj); err != nil {
			p.AddError(err.Error())
			return p, nil
		}
		templates = []plannedTemplate{
			{"models", data.FileName + ".go", modelTmpl},
			{"ports", data.FileName + "_repository.go", repositoryTmpl},
			{"adapters", data.FileName + "_memory_repository.go", adapterTmpl},
			{"services", data.FileName + "_service.go", serviceTmpl},
			{"handlers", data.FileName + "_handler.go", handlerTmpl},
			{"api", data.FileName + "_routes.go", apiRoutesTmpl},
			{"tests", filepath.Join("integration", data.FileName+"_test.go"), integrationTestTmpl},
		}
	case "test", "tests", "t":
		if name == "" {
			return nil, errorsUsage("generate tests requires a name or --scan")
		}
		if err := requireConfiguredResource(prj, data); err != nil {
			p.AddError(err.Error())
			return p, nil
		}
		templates = []plannedTemplate{
			{"handlers", data.FileName + "_handler_test.go", handlerTestTmpl},
			{"services", data.FileName + "_service_test.go", serviceTestTmpl},
			{"adapters", data.FileName + "_memory_repository_test.go", repositoryTestTmpl},
		}
	case "smoke-test", "smoke", "app-test":
		templates = []plannedTemplate{{"tests", filepath.Join("integration", "app_smoke_test.go"), appTestTmpl}}
	case "swagger", "openapi":
		templates = []plannedTemplate{{"docs", "openapi.json", openAPITmpl}}
	case "command":
		base, err := prj.ResolveRelative(filepath.Join("internal", "commands"))
		if err != nil {
			return nil, err
		}
		return planAbsoluteTemplates(prj, p, data, force, base, []plannedTemplate{
			{"", data.FileName + ".go", commandTmpl}, {"", data.FileName + "_test.go", commandTestTmpl},
		})
	case "handler":
		base, err := prj.ResolveRelative(filepath.Join("internal", "worker"))
		if err != nil {
			return nil, err
		}
		return planAbsoluteTemplates(prj, p, data, force, base, []plannedTemplate{
			{"", data.FileName + "_handler.go", workerHandlerGenTmpl},
			{"", data.FileName + "_handler_test.go", workerHandlerTestGenTmpl},
		})
	case "service":
		serviceDir, err := prj.ResolvePath("services")
		if err != nil {
			return nil, err
		}
		portsDir, err := prj.ResolvePath("ports")
		if err != nil {
			return nil, err
		}
		var serviceTemplate, serviceTestTemplate, portTemplate string
		switch prj.ProjectType() {
		case "cli":
			serviceTemplate, serviceTestTemplate, portTemplate = cliServiceTmpl, cliServiceTestTmpl, cliServicePortTmpl
		case "worker":
			serviceTemplate, serviceTestTemplate, portTemplate = workerServiceTmpl, workerServiceTestTmpl, workerServicePortTmpl
		default:
			return nil, fmt.Errorf("project service generation is not supported for %s projects", prj.ProjectType())
		}
		return planExplicitTemplates(prj, p, data, force, []absoluteTemplate{
			{filepath.Join(serviceDir, data.FileName+".go"), serviceTemplate},
			{filepath.Join(serviceDir, data.FileName+"_test.go"), serviceTestTemplate},
			{filepath.Join(portsDir, data.FileName+".go"), portTemplate},
		})
	default:
		return nil, fmt.Errorf("unknown generator: %s", kind)
	}

	return planTemplates(prj, p, data, force, templates)
}

type absoluteTemplate struct{ path, template string }

func planTemplates(prj *project.Project, p *plan.Plan, data genData, force bool, templates []plannedTemplate) (*plan.Plan, error) {
	var absolute []absoluteTemplate
	for _, item := range templates {
		dir, err := prj.ResolvePath(item.key)
		if err != nil {
			return nil, err
		}
		absolute = append(absolute, absoluteTemplate{filepath.Join(dir, item.file), item.template})
	}
	return planExplicitTemplates(prj, p, data, force, absolute)
}

func planAbsoluteTemplates(prj *project.Project, p *plan.Plan, data genData, force bool, base string, templates []plannedTemplate) (*plan.Plan, error) {
	var absolute []absoluteTemplate
	for _, item := range templates {
		absolute = append(absolute, absoluteTemplate{filepath.Join(base, item.file), item.template})
	}
	return planExplicitTemplates(prj, p, data, force, absolute)
}

func planExplicitTemplates(prj *project.Project, p *plan.Plan, data genData, force bool, templates []absoluteTemplate) (*plan.Plan, error) {
	ownership, err := manifest.Load(prj.Root)
	if err != nil {
		return nil, err
	}
	var managed []manifest.Entry
	for _, item := range templates {
		content, err := render(item.template, data)
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(prj.Root, item.path)
		if err != nil {
			return nil, err
		}
		allowOverwrite := force || prj.ShouldOverwrite() || ownership.ManagesFullFile(rel)
		before := len(p.Changes)
		p.AddCreate(item.path, content, allowOverwrite)
		if len(p.Changes) > before {
			change := p.Changes[len(p.Changes)-1]
			if change.Type == plan.ChangeCreate || change.Type == plan.ChangeModify {
				managed = append(managed, manifest.Entry{Path: filepath.ToSlash(rel), FullFile: true})
			}
		}
	}
	if len(managed) > 0 {
		if err := manifest.PlanUpdate(p, managed...); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func projectData(prj *project.Project, name string) (genData, error) {
	data := newData(name)
	if prj.Module == "" {
		data.Module = "yourmodule"
	} else {
		data.Module = prj.Module
	}
	types := []struct {
		key        string
		pkg        *string
		importPath *string
	}{
		{"api", &data.APIPackage, &data.APIImport},
		{"handlers", &data.HandlersPackage, &data.HandlersImport},
		{"models", &data.ModelsPackage, &data.ModelsImport},
		{"services", &data.ServicesPackage, &data.ServicesImport},
		{"ports", &data.PortsPackage, &data.PortsImport},
		{"adapters", &data.AdaptersPackage, &data.AdaptersImport},
		{"config", nil, &data.ConfigImport},
	}
	base, err := prj.BaseDir()
	if err != nil {
		return genData{}, err
	}
	for _, item := range types {
		dir, err := prj.ResolvePath(item.key)
		if err != nil {
			return genData{}, err
		}
		rel, err := filepath.Rel(base, dir)
		if err != nil {
			return genData{}, err
		}
		*item.importPath = strings.TrimSuffix(data.Module, "/") + "/" + filepath.ToSlash(rel)
		if item.pkg != nil {
			*item.pkg = detectPackage(dir)
		}
	}
	return data, nil
}

func detectPackage(dir string) string {
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}
			file, parseErr := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, entry.Name()), nil, parser.PackageClauseOnly)
			if parseErr == nil && file.Name != nil {
				return file.Name.Name
			}
		}
	}
	name := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return r
		}
		return '_'
	}, filepath.Base(dir))
	if name == "" || unicode.IsDigit([]rune(name)[0]) {
		return "generated"
	}
	return name
}

func render(source string, data genData) ([]byte, error) {
	var out strings.Builder
	tmpl, err := template.New("generated").Parse(source)
	if err != nil {
		return nil, fmt.Errorf("parse generator template: %w", err)
	}
	if err := tmpl.Execute(&out, data); err != nil {
		return nil, fmt.Errorf("render generator template: %w", err)
	}
	return []byte(out.String()), nil
}

func requireGingerRouter(prj *project.Project) error {
	api, err := prj.ResolvePath("api")
	if err != nil {
		return err
	}
	entries, _ := os.ReadDir(api)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(api, entry.Name()))
		if readErr == nil && strings.Contains(string(data), "generatedRouteRegistrars") {
			return nil
		}
	}
	return errorsUsage("CRUD generation requires a Ginger-compatible router; no files were changed")
}

func requireConfiguredResource(prj *project.Project, data genData) error {
	required := []struct{ key, file string }{
		{"handlers", data.FileName + "_handler.go"},
		{"services", data.FileName + "_service.go"},
		{"adapters", data.FileName + "_memory_repository.go"},
	}
	var missing []string
	for _, item := range required {
		dir, err := prj.ResolvePath(item.key)
		if err != nil {
			return err
		}
		path := filepath.Join(dir, item.file)
		if _, err := os.Stat(path); err != nil {
			rel, _ := filepath.Rel(prj.Root, path)
			missing = append(missing, rel)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("generate the resource first; missing: %s", strings.Join(missing, ", "))
	}
	return nil
}

// BuildScanTestsPlan creates compiling TODO tests beside existing Go files.
func BuildScanTestsPlan(prj *project.Project, force bool) (*plan.Plan, error) {
	p := plan.New(prj.Root)
	p.CreateMissingDirs = prj.YAML.Rules.CreateMissingDirs
	var sources []string
	for _, key := range []string{"handlers", "services", "repositories"} {
		dir, err := prj.ResolvePath(key)
		if err != nil {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
				sources = append(sources, filepath.Join(dir, entry.Name()))
			}
		}
	}
	sort.Strings(sources)
	var managed []manifest.Entry
	for _, source := range sources {
		target := strings.TrimSuffix(source, ".go") + "_test.go"
		pkg := detectPackage(filepath.Dir(source))
		name := title(splitNameTokens(strings.TrimSuffix(filepath.Base(source), ".go")))
		content := fmt.Sprintf("package %s\n\nimport \"testing\"\n\nfunc Test%s_GingerScaffold(t *testing.T) {\n\tt.Skip(\"TODO: add assertions for %s\")\n}\n", pkg, name, filepath.Base(source))
		before := len(p.Changes)
		p.AddCreate(target, []byte(content), force || prj.ShouldOverwrite())
		if len(p.Changes) > before && (p.Changes[len(p.Changes)-1].Type == plan.ChangeCreate || p.Changes[len(p.Changes)-1].Type == plan.ChangeModify) {
			rel, _ := filepath.Rel(prj.Root, target)
			managed = append(managed, manifest.Entry{Path: filepath.ToSlash(rel), FullFile: true})
		}
	}
	if len(sources) == 0 {
		p.AddWarning("no handlers, services, or repositories were found in configured paths")
	}
	if len(managed) > 0 {
		if err := manifest.PlanUpdate(p, managed...); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func errorsUsage(message string) error { return fmt.Errorf("%s", message) }
