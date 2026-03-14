package cli

import (
	"fmt"
	"os"

	"github.com/ginger-framework/ginger/internal/generator"
)

func runGenerate(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: ginger generate <handler|service|repository|crud> <name>")
		os.Exit(1)
	}

	kind := args[0]
	name := args[1]

	var err error
	switch kind {
	case "handler", "h":
		err = generator.Handler(name)
	case "service", "s":
		err = generator.Service(name)
	case "repository", "repo", "r":
		err = generator.Repository(name)
	case "crud", "c":
		err = generator.CRUD(name)
	default:
		fmt.Fprintf(os.Stderr, "unknown generator: %s\n", kind)
		fmt.Fprintln(os.Stderr, "available: handler, service, repository, crud")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "generate error: %v\n", err)
		os.Exit(1)
	}
}
