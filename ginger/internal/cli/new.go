package cli

import (
	"fmt"
	"os"

	"github.com/ginger-framework/ginger/internal/scaffold"
)

func runNew(args []string) {
	fs := mustFlag("new")
	fs.Parse(args) //nolint:errcheck

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: ginger new <project-name>")
		os.Exit(1)
	}

	name := fs.Arg(0)
	if err := scaffold.NewProject(name); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Project '%s' created successfully!\n\n", name)
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  ginger run\n\n")
}
