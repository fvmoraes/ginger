// Package integrations handles `ginger add <integration>` by generating
// platform integration files and updating go.mod dependencies.
package integrations

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// ErrIntegrationExists is returned when the target integration file already exists.
var ErrIntegrationExists = errors.New("integration already exists")

type integration struct {
	name string
	pkg  string // go get package
	file string // output file path
	tmpl string // file template
}

var registry = map[string]integration{
	// ── Databases ──────────────────────────────────────────────────────────
	"postgres": {
		name: "postgres",
		pkg:  "github.com/lib/pq",
		file: "platform/database/postgres.go",
		tmpl: postgresTmpl,
	},
	"mysql": {
		name: "mysql",
		pkg:  "github.com/go-sql-driver/mysql",
		file: "platform/database/mysql.go",
		tmpl: mysqlTmpl,
	},
	"sqlite": {
		name: "sqlite",
		pkg:  "github.com/mattn/go-sqlite3",
		file: "platform/database/sqlite.go",
		tmpl: sqliteTmpl,
	},
	"sqlserver": {
		name: "sqlserver",
		pkg:  "github.com/microsoft/go-mssqldb",
		file: "platform/database/sqlserver.go",
		tmpl: sqlserverTmpl,
	},
	// ── Cache ──────────────────────────────────────────────────────────────
	"redis": {
		name: "redis",
		pkg:  "github.com/redis/go-redis/v9",
		file: "platform/cache/redis.go",
		tmpl: redisTmpl,
	},
	// ── Messaging ──────────────────────────────────────────────────────────
	"kafka": {
		name: "kafka",
		pkg:  "github.com/segmentio/kafka-go",
		file: "platform/messaging/kafka.go",
		tmpl: kafkaTmpl,
	},
	"rabbitmq": {
		name: "rabbitmq",
		pkg:  "github.com/rabbitmq/amqp091-go",
		file: "platform/messaging/rabbitmq.go",
		tmpl: rabbitmqTmpl,
	},
	"nats": {
		name: "nats",
		pkg:  "github.com/nats-io/nats.go",
		file: "platform/messaging/nats.go",
		tmpl: natsTmpl,
	},
	"pubsub": {
		name: "pubsub",
		pkg:  "cloud.google.com/go/pubsub",
		file: "platform/messaging/pubsub.go",
		tmpl: pubsubTmpl,
	},
	// ── Observability ──────────────────────────────────────────────────────
	"otel": {
		name: "otel",
		pkg:  "go.opentelemetry.io/otel",
		file: "platform/telemetry/otel.go",
		tmpl: otelTmpl,
	},
	"prometheus": {
		name: "prometheus",
		pkg:  "github.com/prometheus/client_golang/prometheus",
		file: "platform/metrics/prometheus.go",
		tmpl: prometheusTmpl,
	},
	// ── Protocols ──────────────────────────────────────────────────────────
	"grpc": {
		name: "grpc",
		pkg:  "google.golang.org/grpc",
		file: "platform/grpc/server.go",
		tmpl: grpcTmpl,
	},
	"mcp": {
		name: "mcp",
		pkg:  "",
		file: "platform/mcp/server.go",
		tmpl: mcpTmpl,
	},
}

// Add generates the integration file and runs go get for the required package.
func Add(name string) error {
	intg, ok := registry[name]
	if !ok {
		return fmt.Errorf(
			"unknown integration: %s\n\navailable integrations:\n"+
				"  databases  : postgres, mysql, sqlite, sqlserver\n"+
				"  cache      : redis\n"+
				"  messaging  : kafka, rabbitmq, nats, pubsub\n"+
				"  protocols  : grpc, mcp\n"+
				"  observ.    : otel, prometheus",
			name,
		)
	}

	if err := os.MkdirAll(filepath.Dir(intg.file), 0755); err != nil {
		return fmt.Errorf("add: mkdir: %w", err)
	}

	if _, err := os.Stat(intg.file); err == nil {
		return fmt.Errorf("%w: %s", ErrIntegrationExists, intg.file)
	}

	f, err := os.Create(intg.file)
	if err != nil {
		return fmt.Errorf("add: create file: %w", err)
	}
	defer f.Close()

	tmpl, err := template.New("").Parse(intg.tmpl)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(f, nil); err != nil {
		return err
	}

	fmt.Printf("  ✓ created %s\n", intg.file)

	// MCP is stdlib-only — no external dependency needed.
	if intg.pkg == "" {
		fmt.Printf("\n✓ Integration '%s' added successfully!\n\n", name)
		return nil
	}

	// go get the dependency
	fmt.Printf("  → go get %s\n", intg.pkg)
	cmd := exec.Command("go", "get", intg.pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("  ■ go get failed — add manually: go get %s\n", intg.pkg)
	} else {
		fmt.Printf("  ✓ dependency added\n")
	}

	fmt.Printf("\n✓ Integration '%s' added successfully!\n\n", name)
	return nil
}
