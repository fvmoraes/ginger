package cli

import (
	"fmt"
	"os"

	"github.com/ginger-framework/ginger/internal/scaffold"
)

func runNew(args []string) {
	fs := mustFlag("new")
	projectType := fs.String("type", "api", "project type: api | microservice | cli | worker")

	// Support both: ginger new myapp --type worker  AND  ginger new --type worker myapp
	// Reorder so flags come before positional args
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		if len(args[i]) > 0 && args[i][0] == '-' {
			flags = append(flags, args[i])
			// consume next arg if it's the flag value (not another flag)
			if i+1 < len(args) && (len(args[i+1]) == 0 || args[i+1][0] != '-') {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, args[i])
		}
	}
	fs.Parse(append(flags, positional...)) //nolint:errcheck

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: ginger new <project-name> [--type api|microservice|cli|worker]")
		os.Exit(1)
	}

	name := fs.Arg(0)
	if err := scaffold.NewProject(name, *projectType); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Project '%s' (%s) created successfully!\n\n", name, *projectType)
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  ginger run\n\n")
}
