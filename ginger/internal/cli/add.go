package cli

import (
	"fmt"
	"os"

	"github.com/ginger-framework/ginger/internal/integrations"
)

const addUsage = `usage: ginger add <integration>

databases  : postgres, mysql, sqlite, sqlserver
cache      : redis
messaging  : kafka, rabbitmq, nats, pubsub
protocols  : grpc, mcp
observ.    : otel, prometheus`

func runAdd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, addUsage)
		os.Exit(1)
	}

	if err := integrations.Add(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "add error: %v\n", err)
		os.Exit(1)
	}
}
