package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fvmoraes/ginger/internal/project"
)

func runInspect(args []string) {
	fs := mustFlag("inspect")
	jsonOut := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "inspect flags: %v\n", err)
		os.Exit(2)
	}
	root, err := project.FindRoot(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	p, err := project.Load(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	report, err := project.Inspect(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "inspect error: %v\n", err)
		os.Exit(1)
	}
	if *jsonOut {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "inspect JSON: %v\n", err)
			os.Exit(1)
		}
		return
	}
	printInspection(report)
}

func printInspection(report *project.Inspection) {
	fmt.Printf("\nGinger project: %s\n\n", report.Module)
	fmt.Printf("  Root:        %s\n", report.Root)
	fmt.Printf("  Type:        %s\n", report.Type)
	fmt.Printf("  Go:          %s\n", report.GoVersion)
	fmt.Printf("  Ginger YAML: %s\n", boolToStr(report.GingerYAML))

	fmt.Println("\n  Structure:")
	keys := make([]string, 0, len(report.Structure))
	for key := range report.Structure {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		status := report.Structure[key]
		marker := "✗"
		if status.Exists {
			marker = "✓"
		}
		fmt.Printf("    %-15s %s %s\n", key+":", marker, status.Path)
	}

	fmt.Printf("\n  Routes (%d):\n", len(report.Routes))
	if len(report.Routes) == 0 {
		fmt.Println("    none detected; add // ginger:route METHOD /path annotations for explicit discovery")
	}
	for _, route := range report.Routes {
		fmt.Printf("    %-7s %-24s %s:%d\n", route.Method, route.Path, route.File, route.Line)
	}
	fmt.Printf("\n  Features:   %s\n", valueOrNone(report.Features))
	fmt.Printf("  Datastores: %s\n", valueOrNone(report.Datastores))
	fmt.Printf("  Test files: %d\n", report.TestFiles)
	for _, warning := range report.Warnings {
		fmt.Printf("  ⚠ %s\n", warning)
	}
	fmt.Println()
}

func boolToStr(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func valueOrNone(values []string) string {
	if len(values) == 0 {
		return "none detected"
	}
	return strings.Join(values, ", ")
}
