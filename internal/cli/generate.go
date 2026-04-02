package cli

import (
	"fmt"
	"os"

	"github.com/fvmoraes/ginger/internal/generator"
)

func runGenerate(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ginger generate <handler|service|repository|crud|test|swagger> [name]")
		os.Exit(1)
	}

	kind := args[0]

	var err error
	switch kind {
	case "handler", "h":
		requireGenerateName(args, kind)
		name := args[1]
		err = generator.Handler(name)
	case "service", "s":
		requireGenerateName(args, kind)
		name := args[1]
		err = generator.Service(name)
	case "repository", "repo", "r":
		requireGenerateName(args, kind)
		name := args[1]
		err = generator.Repository(name)
	case "crud", "c":
		requireGenerateName(args, kind)
		name := args[1]
		err = generator.CRUD(name)
	case "test", "tests", "t":
		err = runGenerateTest(args[1:])
	case "swagger", "openapi":
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		err = generator.Swagger(name)
	default:
		fmt.Fprintf(os.Stderr, "unknown generator: %s\n", kind)
		fmt.Fprintln(os.Stderr, "available: handler, service, repository, crud, test, swagger")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "generate error: %v\n", err)
		os.Exit(1)
	}
}

func runGenerateTest(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ginger generate test <name|app> [handler|service|repository|unit|all]")
	}

	name := args[0]
	scope := "unit"
	if len(args) > 1 {
		scope = args[1]
	}

	if name == "app" {
		return generator.AppTest()
	}

	return generator.Tests(name, scope)
}

func requireGenerateName(args []string, kind string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: ginger generate %s <name>\n", kind)
		os.Exit(1)
	}
}
