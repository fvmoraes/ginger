package cli

import (
	"fmt"
	"os"

	"github.com/fvmoraes/ginger/internal/integrations"
)

const addUsage = `usage: ginger add <integration>

Storage convention:
  platform/...              external infrastructure adapters
  internal/api/handlers/... ready-to-mount HTTP endpoints

databases  : postgres, mysql, sqlite, sqlserver
nosql      : couchbase, mongodb
analytical : clickhouse
cache      : redis
messaging  : kafka, rabbitmq, nats, pubsub
protocols  : grpc, mcp
realtime   : sse, websocket
observ.    : otel, prometheus
docs       : swagger`

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
