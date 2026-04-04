// Package integrations handles `ginger add <integration>` by generating
// platform integration files and updating go.mod dependencies.
package integrations

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// ErrIntegrationExists is returned when the target integration file already exists.
var ErrIntegrationExists = errors.New("integration already exists")

var execCommand = exec.Command

type integration struct {
	name         string
	pkg          string // go get package
	file         string // output file path
	tmpl         string // file template
	postGenerate func() error
}

type composeFile struct {
	Version  string                       `yaml:"version,omitempty"`
	Services map[string]composeService    `yaml:"services,omitempty"`
	Volumes  map[string]map[string]string `yaml:"volumes,omitempty"`
}

type composeService struct {
	Image       string            `yaml:"image,omitempty"`
	Build       *composeBuild     `yaml:"build,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Command     []string          `yaml:"command,omitempty"`
	Restart     string            `yaml:"restart,omitempty"`
}

type composeBuild struct {
	Context    string `yaml:"context,omitempty"`
	Dockerfile string `yaml:"dockerfile,omitempty"`
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
	// ── NoSQL / Analytical ─────────────────────────────────────────────────
	"couchbase": {
		name: "couchbase",
		pkg:  "github.com/couchbase/gocb/v2",
		file: "platform/nosql/couchbase.go",
		tmpl: couchbaseTmpl,
	},
	"mongodb": {
		name: "mongodb",
		pkg:  "go.mongodb.org/mongo-driver/v2/mongo",
		file: "platform/nosql/mongo.go",
		tmpl: mongoTmpl,
	},
	"swagger": {
		name:         "swagger",
		pkg:          "",
		file:         "internal/api/swagger.go",
		tmpl:         swaggerTmpl,
		postGenerate: enableSwaggerRoutes,
	},
	"clickhouse": {
		name: "clickhouse",
		pkg:  "github.com/ClickHouse/clickhouse-go/v2",
		file: "platform/database/clickhouse.go",
		tmpl: clickhouseTmpl,
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
		pkg:  "cloud.google.com/go/pubsub/v2",
		file: "platform/messaging/pubsub.go",
		tmpl: pubsubTmpl,
	},
	// ── UI / Real-time ─────────────────────────────────────────────────────────
	"sse": {
		name: "sse",
		pkg:  "",
		file: "internal/api/handlers/sse_handler.go",
		tmpl: sseTmpl,
	},
	"websocket": {
		name: "websocket",
		pkg:  "",
		file: "internal/api/handlers/ws_handler.go",
		tmpl: wsTmpl,
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
				"  nosql      : couchbase, mongodb\n"+
				"  analytical : clickhouse\n"+
				"  cache      : redis\n"+
				"  messaging  : kafka, rabbitmq, nats, pubsub\n"+
				"  protocols  : grpc, mcp\n"+
				"  realtime   : sse, websocket\n"+
				"  observ.    : otel, prometheus\n"+
				"  docs       : swagger",
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

	if intg.postGenerate != nil {
		if err := intg.postGenerate(); err != nil {
			_ = os.Remove(intg.file)
			return err
		}
	}

	fmt.Printf("  ✓ created %s\n", intg.file)

	// MCP is stdlib-only — no external dependency needed.
	if intg.pkg == "" {
		if err := updateDockerCompose(name); err != nil {
			return err
		}
		fmt.Printf("\n✓ Integration '%s' added successfully!\n\n", name)
		return nil
	}

	// go get the dependency
	fmt.Printf("  → go get %s\n", intg.pkg)
	cmd := execCommand("go", "get", intg.pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(intg.file)
		return fmt.Errorf("add dependency %s: %w", intg.pkg, err)
	}
	fmt.Printf("  ✓ dependency added\n")

	if err := updateDockerCompose(name); err != nil {
		return err
	}

	fmt.Printf("\n✓ Integration '%s' added successfully!\n\n", name)
	return nil
}

func updateDockerCompose(integrationName string) error {
	composePath := filepath.Join("devops", "docker", "docker-compose.yml")
	if _, err := os.Stat(composePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("add: stat compose file: %w", err)
	}

	data, err := os.ReadFile(composePath)
	if err != nil {
		return fmt.Errorf("add: read compose file: %w", err)
	}

	var compose composeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return fmt.Errorf("add: parse compose file: %w", err)
	}

	if compose.Services == nil {
		compose.Services = make(map[string]composeService)
	}
	if compose.Volumes == nil {
		compose.Volumes = make(map[string]map[string]string)
	}

	appName := detectComposeAppService(compose.Services)
	app := compose.Services[appName]

	changed := mergeIntegrationIntoCompose(&compose, appName, &app, integrationName)
	if !changed {
		return nil
	}

	compose.Services[appName] = app

	out, err := yaml.Marshal(&compose)
	if err != nil {
		return fmt.Errorf("add: marshal compose file: %w", err)
	}

	if err := os.WriteFile(composePath, out, 0644); err != nil {
		return fmt.Errorf("add: write compose file: %w", err)
	}

	fmt.Printf("  ✓ updated %s\n", composePath)
	return nil
}

func detectComposeAppService(services map[string]composeService) string {
	wd, err := os.Getwd()
	if err == nil {
		projectName := filepath.Base(wd)
		if _, ok := services[projectName]; ok {
			return projectName
		}
	}

	for name := range services {
		return name
	}

	return "app"
}

func mergeIntegrationIntoCompose(compose *composeFile, appName string, app *composeService, integrationName string) bool {
	changed := false
	projectName := appName

	ensureAppEnv := func(key, value string) {
		if app.Environment == nil {
			app.Environment = make(map[string]string)
		}
		if _, ok := app.Environment[key]; !ok {
			app.Environment[key] = value
			changed = true
		}
	}

	addDependency := func(name string) {
		if !contains(app.DependsOn, name) {
			app.DependsOn = append(app.DependsOn, name)
			changed = true
		}
	}

	addService := func(name string, svc composeService) {
		if _, ok := compose.Services[name]; !ok {
			compose.Services[name] = svc
			changed = true
		}
	}

	addVolume := func(name string) {
		if _, ok := compose.Volumes[name]; !ok {
			compose.Volumes[name] = map[string]string{}
			changed = true
		}
	}

	switch integrationName {
	case "postgres":
		addDependency("postgres")
		ensureAppEnv("DATABASE_DSN", fmt.Sprintf("postgres://user:pass@postgres:5432/%s?sslmode=disable", projectName))
		addService("postgres", composeService{
			Image: "postgres:16-alpine",
			Environment: map[string]string{
				"POSTGRES_USER":     "user",
				"POSTGRES_PASSWORD": "pass",
				"POSTGRES_DB":       projectName,
			},
			Ports:   []string{"5432:5432"},
			Volumes: []string{"pgdata:/var/lib/postgresql/data"},
		})
		addVolume("pgdata")
	case "redis":
		addDependency("redis")
		ensureAppEnv("REDIS_ADDR", "redis:6379")
		addService("redis", composeService{
			Image: "redis:7-alpine",
			Ports: []string{"6379:6379"},
		})
	case "rabbitmq":
		addDependency("rabbitmq")
		ensureAppEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
		addService("rabbitmq", composeService{
			Image: "rabbitmq:3-management-alpine",
			Ports: []string{"5672:5672", "15672:15672"},
		})
	case "kafka":
		addDependency("kafka")
		ensureAppEnv("KAFKA_BROKERS", "kafka:9092")
		addService("kafka", composeService{
			Image: "bitnami/kafka:3.7",
			Ports: []string{"9092:9092"},
			Environment: map[string]string{
				"KAFKA_CFG_NODE_ID":                        "1",
				"KAFKA_CFG_PROCESS_ROLES":                  "broker,controller",
				"KAFKA_CFG_CONTROLLER_LISTENER_NAMES":      "CONTROLLER",
				"KAFKA_CFG_LISTENERS":                      "PLAINTEXT://:9092,CONTROLLER://:9093",
				"KAFKA_CFG_ADVERTISED_LISTENERS":           "PLAINTEXT://kafka:9092",
				"KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP": "PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT",
				"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS":       "1@kafka:9093",
				"ALLOW_PLAINTEXT_LISTENER":                 "yes",
			},
		})
	case "nats":
		addDependency("nats")
		ensureAppEnv("NATS_URL", "nats://nats:4222")
		addService("nats", composeService{
			Image:   "nats:2-alpine",
			Ports:   []string{"4222:4222", "8222:8222"},
			Command: []string{"-js"},
		})
	case "mongodb":
		addDependency("mongodb")
		ensureAppEnv("MONGODB_URI", fmt.Sprintf("mongodb://mongodb:27017/%s", projectName))
		addService("mongodb", composeService{
			Image: "mongo:7",
			Ports: []string{"27017:27017"},
		})
	case "mysql":
		addDependency("mysql")
		ensureAppEnv("DATABASE_DSN", fmt.Sprintf("root:root@tcp(mysql:3306)/%s?parseTime=true", projectName))
		addService("mysql", composeService{
			Image: "mysql:8",
			Environment: map[string]string{
				"MYSQL_ROOT_PASSWORD": "root",
				"MYSQL_DATABASE":      projectName,
			},
			Ports: []string{"3306:3306"},
		})
	case "clickhouse":
		addDependency("clickhouse")
		addService("clickhouse", composeService{
			Image: "clickhouse/clickhouse-server:24.3",
			Ports: []string{"8123:8123", "9000:9000"},
		})
	case "couchbase":
		addDependency("couchbase")
		addService("couchbase", composeService{
			Image: "couchbase:community-7.6.2",
			Ports: []string{"8091:8091", "11210:11210"},
		})
	case "prometheus":
		addService("prometheus", composeService{
			Image:   "prom/prometheus:latest",
			Ports:   []string{"9090:9090"},
			Volumes: []string{"./prometheus.yml:/etc/prometheus/prometheus.yml"},
		})
	case "otel":
		addService("otel-collector", composeService{
			Image: "otel/opentelemetry-collector:0.102.1",
			Ports: []string{"4317:4317", "4318:4318"},
		})
		ensureAppEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318")
		addDependency("otel-collector")
	}

	return changed
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func enableSwaggerRoutes() error {
	routerPath := filepath.Join("internal", "api", "router.go")
	data, err := os.ReadFile(routerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("add swagger: internal/api/router.go not found; swagger integration requires a service project")
		}
		return fmt.Errorf("add swagger: read router: %w", err)
	}

	content := string(data)
	if strings.Contains(content, "registerSwaggerRoutes(r)") {
		return nil
	}

	const marker = "\tv1 := r.Group(\"/api/v1\", middlewares.RequestID)\n"
	if !strings.Contains(content, marker) {
		return fmt.Errorf("add swagger: could not locate API group registration in %s", routerPath)
	}

	updated := strings.Replace(content, marker, "\tregisterSwaggerRoutes(r)\n"+marker, 1)
	if err := os.WriteFile(routerPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("add swagger: write router: %w", err)
	}

	return nil
}
