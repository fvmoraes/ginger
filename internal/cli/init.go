package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fvmoraes/ginger/internal/manifest"
	"github.com/fvmoraes/ginger/internal/plan"
	"github.com/fvmoraes/ginger/internal/project"
	"gopkg.in/yaml.v3"
)

func runInit(args []string) {
	fs := mustFlag("init")
	force := fs.Bool("force", false, "overwrite existing ginger.yaml")
	fs.Parse(args) //nolint:errcheck

	root, err := project.FindRoot(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v; ginger init requires an existing Go project with go.mod\n", err)
		os.Exit(1)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		fmt.Fprintln(os.Stderr, "error: ginger init requires go.mod at the detected project root")
		os.Exit(1)
	}

	gingerPath := filepath.Join(root, "ginger.yaml")
	if _, err := os.Stat(gingerPath); err == nil && !*force {
		fmt.Printf("ginger.yaml already exists at %s; nothing changed.\n", gingerPath)
		return
	}

	p, err := project.Detect(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading project: %v\n", err)
		os.Exit(1)
	}

	if p.YAML == nil {
		fmt.Fprintf(os.Stderr, "error: could not detect project structure\n")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Ginger detected an existing Go project.")
	fmt.Println()
	fmt.Printf("Project:\n")
	fmt.Printf("  Module:  %s\n", p.Module)
	fmt.Printf("  Type:    %s\n", p.ProjectType())
	fmt.Printf("  Root:    %s\n", p.Root)

	if p.ProjectType() == "service" || p.ProjectType() == "worker" {
		fmt.Printf("  Handlers: %s\n", p.YAML.Structure.Handlers)
		fmt.Printf("  Models:  %s\n", p.YAML.Structure.Models)
		fmt.Printf("  Config:  %s\n", p.YAML.Structure.Config)
	}

	tests := "no"
	if inspection, inspectErr := project.Inspect(p); inspectErr == nil && inspection.TestFiles > 0 {
		tests = "partial"
	}
	fmt.Printf("  Tests:   %s\n", tests)

	fmt.Println()

	data, err := yaml.Marshal(p.YAML)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling ginger.yaml: %v\n", err)
		os.Exit(1)
	}

	changePlan := plan.New(root)
	changePlan.AddCreate(gingerPath, data, *force)
	if err := manifest.PlanUpdate(changePlan, manifest.Entry{Path: "ginger.yaml", FullFile: true}); err != nil {
		fmt.Fprintf(os.Stderr, "error planning manifest: %v\n", err)
		os.Exit(1)
	}
	changePlan.Render()
	if changePlan.HasErrors() {
		fmt.Fprintln(os.Stderr, "init blocked by plan errors")
		os.Exit(1)
	}
	if err := changePlan.Apply(); err != nil {
		fmt.Fprintf(os.Stderr, "error applying init plan: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Created:")
	fmt.Printf("  + ginger.yaml\n  + .ginger/manifest.yaml\n")
	fmt.Println()
	fmt.Println("Next:")
	fmt.Println("  ginger inspect")
	fmt.Println("  ginger add swagger --plan")
	fmt.Println("  ginger generate tests --plan")
	fmt.Println()
}
