package cli

import (
	"fmt"
	"os"

	"github.com/fvmoraes/ginger/internal/generator"
)

// detectProjectType inspects the current working directory to determine the
// project type scaffolded by ginger.
//
//   - internal/commands/ present → cli
//   - internal/worker/   present → worker
//   - internal/api/      present → service
//   - otherwise          → generic
func detectProjectType() string {
	if dirExists("internal/commands") {
		return "cli"
	}
	if dirExists("internal/worker") {
		return "worker"
	}
	if dirExists("internal/api") {
		return "service"
	}
	return "generic"
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func runGenerate(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, generateUsage())
		os.Exit(1)
	}

	kind := args[0]
	projectType := detectProjectType()

	var err error
	switch kind {
	case "crud", "c":
		requireGenerateName(args, kind)
		name := args[1]
		if err = validateGeneratorForProjectType(projectType, kind); err == nil {
			err = generator.CRUD(name)
		}
	case "test", "tests", "t":
		err = runGenerateTest(args[1:])
	case "smoke-test", "smoke", "app-test":
		err = runGenerateSmokeTest(args[1:])
	case "swagger", "openapi":
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		err = generator.Swagger(name)
	case "command":
		requireGenerateName(args, kind)
		name := args[1]
		if err = validateGeneratorForProjectType(projectType, kind); err == nil {
			err = generator.Command(name)
		}
	case "handler":
		requireGenerateName(args, kind)
		name := args[1]
		if err = validateGeneratorForProjectType(projectType, kind); err == nil {
			err = generator.WorkerHandler(name)
		}
	case "service":
		requireGenerateName(args, kind)
		name := args[1]
		if err = validateGeneratorForProjectType(projectType, kind); err == nil {
			err = generator.ProjectService(name, projectType)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown generator: %s\n", kind)
		fmt.Fprintln(os.Stderr, generateUsage())
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "generate error: %v\n", err)
		os.Exit(1)
	}
}

func generateUsage() string {
	return `usage: ginger generate <subcommand> [name]

subcommands:
  crud        <name>   generate model/handler/service/ports/adapter (--service projects)
  command     <name>   generate a Cobra subcommand            (--cli projects)
  handler     <name>   generate a worker message handler      (--worker projects)
  service     <name>   generate a business service            (--cli/--worker projects)
  test        <name>   generate unit tests for a resource
  smoke-test           generate an app-level smoke test
  swagger     [name]   generate OpenAPI spec`
}

func runGenerateTest(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ginger generate test <name|app>")
	}

	if len(args) > 1 {
		return fmt.Errorf("test scopes are no longer supported; use 'ginger generate test <name>' or 'ginger generate test app'")
	}

	name := args[0]
	if name == "app" {
		return generator.AppTest()
	}

	return generator.Tests(name)
}

func runGenerateSmokeTest(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: ginger generate smoke-test")
	}

	return generator.AppTest()
}

func requireGenerateName(args []string, kind string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: ginger generate %s <name>\n", kind)
		os.Exit(1)
	}
}

func validateGeneratorForProjectType(projectType, kind string) error {
	switch kind {
	case "crud", "c":
		if projectType != "service" {
			return fmt.Errorf("generate %s is only available in --service projects", kind)
		}
	case "command":
		if projectType != "cli" {
			return fmt.Errorf("generate %s is only available in --cli projects", kind)
		}
	case "handler":
		if projectType != "worker" {
			return fmt.Errorf("generate %s is only available in --worker projects", kind)
		}
	case "service":
		if projectType != "cli" && projectType != "worker" {
			return fmt.Errorf("generate %s is only available in --cli and --worker projects", kind)
		}
	}
	return nil
}
