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
	"postgres": {
		name: "postgres",
		pkg:  "github.com/lib/pq",
		file: "platform/database/postgres.go",
		tmpl: postgresTmpl,
	},
	"redis": {
		name: "redis",
		pkg:  "github.com/redis/go-redis/v9",
		file: "platform/cache/redis.go",
		tmpl: redisTmpl,
	},
	"kafka": {
		name: "kafka",
		pkg:  "github.com/segmentio/kafka-go",
		file: "platform/messaging/kafka.go",
		tmpl: kafkaTmpl,
	},
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
}

// Add generates the integration file and runs go get for the required package.
func Add(name string) error {
	intg, ok := registry[name]
	if !ok {
		return fmt.Errorf("unknown integration: %s\navailable: postgres, redis, kafka, otel, prometheus", name)
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
