package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fvmoraes/ginger/internal/doctor"
	"github.com/fvmoraes/ginger/internal/project"
)

func runDoctor(args []string) {
	fs := mustFlag("doctor")
	jsonOut := fs.Bool("json", false, "output as JSON")
	fix := fs.Bool("fix", false, "apply safe automatic fixes")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "doctor flags: %v\n", err)
		os.Exit(2)
	}
	root, err := resolveRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	prj, err := project.Load(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if *fix {
		fmt.Fprintln(os.Stderr, "doctor --fix has no implicit fixes yet; use the suggested plan commands")
	}
	diagnostic, err := doctor.Diagnose(prj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "doctor error: %v\n", err)
		os.Exit(1)
	}
	if *jsonOut {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(diagnostic); err != nil {
			fmt.Fprintf(os.Stderr, "doctor JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		printDiagnostic(diagnostic)
	}
	if !diagnostic.Healthy() {
		os.Exit(1)
	}
}

func printDiagnostic(d *doctor.Diagnostic) {
	fmt.Printf("\nGinger doctor\n  Root: %s\n\n", d.Root)
	for _, check := range d.Checks {
		symbol := "✓"
		if check.Status == "fail" {
			symbol = "✗"
		}
		fmt.Printf("  %s %s", symbol, check.Name)
		if check.Status == "fail" && check.Detail != "" {
			fmt.Printf(" — %s", check.Detail)
		}
		fmt.Println()
	}
	fmt.Println("\nAvailable capabilities:")
	for _, name := range d.AvailableCapabilities {
		fmt.Printf("  ✓ %s\n", name)
	}
	if len(d.BlockedCapabilities) > 0 {
		fmt.Println("\nBlocked capabilities:")
		for _, blocked := range d.BlockedCapabilities {
			fmt.Printf("  ✗ %s — %s\n", blocked.Name, blocked.Reason)
		}
	}
	fmt.Println()
}
