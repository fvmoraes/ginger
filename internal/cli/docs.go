package cli

import (
	"fmt"
	"os"

	"github.com/fvmoraes/ginger/internal/docsgen"
	"github.com/fvmoraes/ginger/internal/project"
)

func runDocs(args []string) {
	fs := mustFlag("docs")
	planOnly := fs.Bool("plan", false, "show plan without applying")
	force := fs.Bool("force", false, "replace existing managed documentation")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "docs flags: %v\n", err)
		os.Exit(2)
	}
	root, err := project.FindRoot(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "docs error: %v\n", err)
		os.Exit(1)
	}
	prj, err := project.Load(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "docs error: %v\n", err)
		os.Exit(1)
	}
	p, err := docsgen.BuildPlan(prj, *force)
	if err != nil {
		fmt.Fprintf(os.Stderr, "docs error: %v\n", err)
		os.Exit(1)
	}
	p.Render()
	if *planOnly {
		return
	}
	if p.HasErrors() {
		fmt.Fprintln(os.Stderr, "docs blocked by plan errors")
		os.Exit(1)
	}
	if !p.HasChanges() {
		fmt.Println("Nothing to do; existing documentation was preserved.")
		return
	}
	if err := p.Apply(); err != nil {
		fmt.Fprintf(os.Stderr, "docs apply: %v\n", err)
		os.Exit(1)
	}
}
