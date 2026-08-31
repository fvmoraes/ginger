package cli

import (
	"fmt"
	"os"

	"github.com/fvmoraes/ginger/internal/generator"
	"github.com/fvmoraes/ginger/internal/project"
)

func runGenerate(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, generateUsage())
		os.Exit(1)
	}

	kind := args[0]
	rest := args[1:]

	// Parse --plan and --force from rest args
	planOnly := false
	force := false
	scan := false
	filtered := make([]string, 0, len(rest))
	for _, arg := range rest {
		switch arg {
		case "--plan":
			planOnly = true
		case "--force":
			force = true
		case "--scan":
			scan = true
		default:
			filtered = append(filtered, arg)
		}
	}

	// Find project root
	root, err := resolveRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'ginger init' to initialize a project or create one with 'ginger new'.")
		os.Exit(1)
	}

	prj, err := project.Load(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	projectType := prj.ProjectType()
	if err = validateGeneratorForProjectType(projectType, kind); err != nil {
		fmt.Fprintf(os.Stderr, "generate error: %v\n", err)
		os.Exit(1)
	}

	name := ""
	if len(filtered) > 0 {
		name = filtered[0]
	}
	// GIN-020: extra positional arguments were silently ignored before.
	if len(filtered) > 1 {
		fmt.Fprintf(os.Stderr, "error: unexpected extra argument %q — generate accepts <kind> [name]\n", filtered[1])
		os.Exit(2)
	}
	requiresName := kind == "crud" || kind == "c" || kind == "command" || kind == "handler" || kind == "service"
	if requiresName && name == "" {
		fmt.Fprintf(os.Stderr, "usage: ginger generate %s <name> [--plan] [--force]\n", kind)
		os.Exit(1)
	}

	var generationPlan interface {
		Render()
		HasErrors() bool
		HasChanges() bool
		Apply() error
	}
	if (kind == "test" || kind == "tests" || kind == "t") && scan {
		generationPlan, err = generator.BuildScanTestsPlan(prj, force)
	} else {
		generationPlan, err = generator.BuildPlan(prj, kind, name, force)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate error: %v\n", err)
		os.Exit(1)
	}
	generationPlan.Render()
	if planOnly {
		return
	}
	if generationPlan.HasErrors() {
		fmt.Fprintln(os.Stderr, "generate blocked: resolve plan errors before applying")
		os.Exit(1)
	}
	if !generationPlan.HasChanges() {
		fmt.Println("Nothing to do; existing files were preserved.")
		return
	}
	if err := generationPlan.Apply(); err != nil {
		fmt.Fprintf(os.Stderr, "apply error: %v\n", err)
		os.Exit(1)
	}
}

func generateUsage() string {
	return `usage: ginger generate <subcommand> [name]

subcommands:
  crud        <name>       generate model/handler/service/ports/adapter (--service)
  command     <name>       generate a Cobra subcommand (--cli)
  handler     <name>       generate a worker message handler (--worker)
  service     <name>       generate a business service (--cli/--worker)
  test        <name>       generate tests for a Ginger resource
  tests       --scan       generate compiling TODO tests for existing code
  smoke-test               generate an app-level smoke test
  swagger     [name]       generate OpenAPI spec

flags:
  --plan                   print the complete plan without writing
  --force                  explicitly replace existing targets
  --scan                   scan existing handlers/services/repositories`
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
