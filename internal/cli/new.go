package cli

import (
	"fmt"
	"os"

	"github.com/fvmoraes/ginger/internal/scaffold"
)

func runNew(args []string) {
	fs := mustFlag("new")
	isAPI := fs.Bool("a", false, "API project      → cmd/<name>-api")
	isSvc := fs.Bool("s", false, "service project  → cmd/<name>-service")
	isWorker := fs.Bool("w", false, "worker project   → cmd/<name>-worker")
	isCLI := fs.Bool("c", false, "CLI project      → cmd/<name>-cli")

	// Reorder so flags come before positional args
	var flags, positional []string
	for _, arg := range args {
		if len(arg) > 0 && arg[0] == '-' {
			flags = append(flags, arg)
		} else {
			positional = append(positional, arg)
		}
	}
	fs.Parse(append(flags, positional...)) //nolint:errcheck

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: ginger new <name> [-a|-s|-w|-c]")
		fmt.Fprintln(os.Stderr, "  (no flag)  generic   → cmd/<name>")
		fmt.Fprintln(os.Stderr, "  -a         api       → cmd/<name>-api")
		fmt.Fprintln(os.Stderr, "  -s         service   → cmd/<name>-service")
		fmt.Fprintln(os.Stderr, "  -w         worker    → cmd/<name>-worker")
		fmt.Fprintln(os.Stderr, "  -c         cli       → cmd/<name>-cli")
		os.Exit(1)
	}

	name := fs.Arg(0)

	projectType := "generic"
	switch {
	case *isAPI:
		projectType = "api"
	case *isSvc:
		projectType = "service"
	case *isWorker:
		projectType = "worker"
	case *isCLI:
		projectType = "cli"
	}

	if err := scaffold.NewProject(name, projectType); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cmdDir := scaffold.CmdDir(name, projectType)
	fmt.Printf("\n✓ Project '%s' created successfully!\n\n", name)
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  go run ./%s\n\n", cmdDir)
}
