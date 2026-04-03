package cli

import (
	"fmt"
	"os"

	"github.com/fvmoraes/ginger/internal/scaffold"
)

func runNew(args []string) {
	fs := mustFlag("new")
	isSvc := fs.Bool("service", false, "service project  → cmd/<name>")
	isWorker := fs.Bool("worker", false, "worker project   → cmd/<name>-worker")
	isCLI := fs.Bool("cli", false, "CLI project      → cmd/<name>")

	// Short aliases.
	isSvcShort := fs.Bool("s", false, "alias for --service")
	isWorkerShort := fs.Bool("w", false, "alias for --worker")
	isCLIShort := fs.Bool("c", false, "alias for --cli")

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
		fmt.Fprintln(os.Stderr, "usage: ginger new <name> [--service|--worker|--cli]")
		fmt.Fprintln(os.Stderr, "  (no flag)  generic   → cmd/<name>")
		fmt.Fprintln(os.Stderr, "  --service  service   → cmd/<name>")
		fmt.Fprintln(os.Stderr, "  --worker   worker    → cmd/<name>-worker")
		fmt.Fprintln(os.Stderr, "  --cli      cli       → cmd/<name>")
		os.Exit(1)
	}

	name := fs.Arg(0)

	projectType := "generic"
	switch {
	case *isSvc || *isSvcShort:
		projectType = "service"
	case *isWorker || *isWorkerShort:
		projectType = "worker"
	case *isCLI || *isCLIShort:
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
