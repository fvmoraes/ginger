package cli

import (
	"fmt"
	"os"

	"github.com/ginger-framework/ginger/internal/integrations"
)

func runAdd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ginger add <postgres|redis|kafka|otel|prometheus>")
		os.Exit(1)
	}

	integration := args[0]
	if err := integrations.Add(integration); err != nil {
		fmt.Fprintf(os.Stderr, "add error: %v\n", err)
		os.Exit(1)
	}
}
