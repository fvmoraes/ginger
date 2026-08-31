// Package integrations handles `ginger add <integration>` by generating
// platform integration files and updating go.mod dependencies.
package integrations

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/fvmoraes/ginger/internal/manifest"
	"github.com/fvmoraes/ginger/internal/plan"
	"github.com/fvmoraes/ginger/internal/project"
	"github.com/fvmoraes/ginger/internal/region"
	"gopkg.in/yaml.v3"
)

// hexSha256 returns the hex digest of b (GIN-002 provenance hash).
func hexSha256(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// hexSha256From is a readability alias for hexSha256.
func hexSha256From(b []byte) string { return hexSha256(b) }

// ErrIntegrationExists is returned when the target integration file already exists.
var ErrIntegrationExists = errors.New("integration already exists")

var execCommand = exec.Command

// IntegrationSpec holds the declarative metadata of a cataloged integration.
// GIN-004: single source of truth — the capability registry (constraints) is
// derived from this catalog, so an integration cannot bypass validations by
// being absent from it. Unknown integration names now fail closed.
type IntegrationSpec struct {
	Name string
	// Description for help/docs.
	Description string
	// Pkg is the `go get` target ("" = stdlib-only).
	Pkg string
	// File is the generated file path (relative to the project root).
	File string
	// MinGo is the minimum Go version ("" = none).
	MinGo string
	// ProjectTypes allowed; empty = allowed for any type.
	ProjectTypes []string
}

type integration struct {
	name         string
	description  string // GIN-004
	minGo        string // GIN-004
	projectTypes []string
	pkg          string // go get package
	file         string // output file path
	tmpl         string // file template
}

// Spec exposes the declarative metadata (GIN-004 catalog).
func (i integration) Spec() IntegrationSpec {
	return IntegrationSpec{
		Name:         i.name,
		Description:  i.description,
		Pkg:          i.pkg,
		File:         i.file,
		MinGo:        i.minGo,
		ProjectTypes: i.projectTypes,
	}
}

// Catalog returns the declarative specs of every registered integration.
func Catalog() []IntegrationSpec {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	specs := make([]IntegrationSpec, 0, len(names))
	for _, n := range names {
		specs = append(specs, registry[n].Spec())
	}
	return specs
}

// IsCataloged reports whether the integration exists in the catalog (GIN-004).
func IsCataloged(name string) bool {
	_, ok := registry[name]
	return ok
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

// UnmarshalYAML (GIN-002 extensão): aceita a forma abreviada do Compose
// (`build: .` — string) além da forma de mapeamento (`build: {context: .}`).
// Sem isso, um compose customizado com a forma abreviada falhava o
// `ginger add` inteiro com um erro de parse.
func (b *composeBuild) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var asString string
	if err := unmarshal(&asString); err == nil {
		b.Context = asString
		return nil
	}
	var asMap struct {
		Context    string `yaml:"context"`
		Dockerfile string `yaml:"dockerfile"`
	}
	if err := unmarshal(&asMap); err != nil {
		return fmt.Errorf("compose build: expected string or mapping: %w", err)
	}
	b.Context = asMap.Context
	b.Dockerfile = asMap.Dockerfile
	return nil
}

var registry = map[string]integration{
	// ── Databases ──────────────────────────────────────────────────────────
	"postgres": {
		name:        "postgres",
		description: "PostgreSQL database adapter",
		pkg:         "github.com/lib/pq",
		file:        "platform/database/postgres.go",
		tmpl:        postgresTmpl,
	},
	"mysql": {
		name:        "mysql",
		description: "MySQL database adapter",
		pkg:         "github.com/go-sql-driver/mysql",
		file:        "platform/database/mysql.go",
		tmpl:        mysqlTmpl,
	},
	"sqlite": {
		name:        "sqlite",
		description: "SQLite database adapter",
		pkg:         "github.com/mattn/go-sqlite3",
		file:        "platform/database/sqlite.go",
		tmpl:        sqliteTmpl,
	},
	"sqlserver": {
		name:        "sqlserver",
		description: "SQL Server database adapter",
		pkg:         "github.com/microsoft/go-mssqldb",
		file:        "platform/database/sqlserver.go",
		tmpl:        sqlserverTmpl,
	},
	"gorm": {
		name: "gorm", description: "GORM ORM integration", pkg: "gorm.io/gorm", file: "platform/database/gorm.go", tmpl: gormTmpl,
	},
	"sqlx": {
		name: "sqlx", description: "sqlx query builder integration", pkg: "github.com/jmoiron/sqlx github.com/lib/pq", file: "platform/database/sqlx.go", tmpl: sqlxTmpl,
	},
	"bun": {
		name:        "bun",
		description: "Bun ORM integration",
		pkg:         "github.com/uptrace/bun github.com/uptrace/bun/dialect/pgdialect github.com/uptrace/bun/driver/pgdriver",
		file:        "platform/database/bun.go", tmpl: bunTmpl,
	},
	// ── Cache ──────────────────────────────────────────────────────────────
	"redis": {
		name:        "redis",
		description: "Redis cache client",
		pkg:         "github.com/redis/go-redis/v9",
		file:        "platform/cache/redis.go",
		tmpl:        redisTmpl,
	},
	// ── NoSQL / Analytical ─────────────────────────────────────────────────
	"couchbase": {
		name:        "couchbase",
		description: "Couchbase NoSQL client",
		pkg:         "github.com/couchbase/gocb/v2",
		file:        "platform/nosql/couchbase.go",
		tmpl:        couchbaseTmpl,
	},
	"mongodb": {
		name:        "mongodb",
		description: "MongoDB client",
		pkg:         "go.mongodb.org/mongo-driver/v2/mongo",
		file:        "platform/nosql/mongo.go",
		tmpl:        mongoTmpl,
	},
	"swagger": {
		name:         "swagger",
		description:  "OpenAPI documentation generator",
		projectTypes: []string{"service"},
		pkg:          "",
		file:         "internal/api/swagger.go",
		tmpl:         swaggerTmpl,
	},
	"clickhouse": {
		name:        "clickhouse",
		description: "ClickHouse analytical store client",
		pkg:         "github.com/ClickHouse/clickhouse-go/v2",
		file:        "platform/database/clickhouse.go",
		tmpl:        clickhouseTmpl,
	},
	// ── Messaging ──────────────────────────────────────────────────────────
	"kafka": {
		name:        "kafka",
		description: "Kafka producer/consumer",
		pkg:         "github.com/segmentio/kafka-go",
		file:        "platform/messaging/kafka.go",
		tmpl:        kafkaTmpl,
	},
	"rabbitmq": {
		name:        "rabbitmq",
		description: "RabbitMQ producer/consumer",
		pkg:         "github.com/rabbitmq/amqp091-go",
		file:        "platform/messaging/rabbitmq.go",
		tmpl:        rabbitmqTmpl,
	},
	"nats": {
		name:        "nats",
		description: "NATS producer/consumer",
		pkg:         "github.com/nats-io/nats.go",
		file:        "platform/messaging/nats.go",
		tmpl:        natsTmpl,
	},
	"pubsub": {
		name:        "pubsub",
		description: "Google Cloud Pub/Sub producer/consumer",
		pkg:         "cloud.google.com/go/pubsub/v2",
		file:        "platform/messaging/pubsub.go",
		tmpl:        pubsubTmpl,
	},
	// ── UI / Real-time ─────────────────────────────────────────────────────────
	"sse": {
		name:        "sse",
		description: "Server-Sent Events handler",
		pkg:         "",
		file:        "internal/api/handlers/sse_handler.go",
		tmpl:        sseTmpl,
	},
	"websocket": {
		name:        "websocket",
		description: "WebSocket handler",
		pkg:         "",
		file:        "internal/api/handlers/ws_handler.go",
		tmpl:        wsTmpl,
	},
	// ── Observability ──────────────────────────────────────────────────────
	"otel": {
		name:         "otel",
		description:  "OpenTelemetry tracing",
		minGo:        "1.25",
		projectTypes: []string{"service", "worker"},
		pkg: "go.opentelemetry.io/otel@v1.43.0 " +
			"go.opentelemetry.io/otel/sdk@v1.43.0 " +
			"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@v1.43.0",
		file: "platform/telemetry/otel.go",
		tmpl: otelTmpl,
	},
	"prometheus": {
		name:         "prometheus",
		description:  "Prometheus metrics endpoint",
		projectTypes: []string{"service", "worker"},
		pkg:          "github.com/prometheus/client_golang/prometheus",
		file:         "platform/metrics/prometheus.go",
		tmpl:         prometheusTmpl,
	},
	// ── Protocols ──────────────────────────────────────────────────────────
	"grpc": {
		name:         "grpc",
		description:  "gRPC server integration",
		projectTypes: []string{"service"},
		pkg:          "google.golang.org/grpc",
		file:         "platform/grpc/server.go",
		tmpl:         grpcTmpl,
	},
	"mcp": {
		name:         "mcp",
		description:  "Model Context Protocol server",
		projectTypes: []string{"service", "cli"},
		pkg:          "",
		file:         "platform/mcp/server.go",
		tmpl:         mcpTmpl,
	},
}

// ApplyWithRollback (GIN-005) applies the plan, runs PostApply, and undoes
// the ENTIRE apply (creates + modifies + manifest) if the post-apply step
// fails — a failed `go get` must not leave half-applied integrations behind.
func ApplyWithRollback(p *plan.Plan, name, projectRoot string) error {
	snapshot, err := p.Snapshot()
	if err != nil {
		return fmt.Errorf("snapshot before apply: %w", err)
	}

	if err := p.Apply(); err != nil {
		return fmt.Errorf("apply: %w", err)
	}

	if !NeedsPostApply(name, p) {
		return nil
	}

	if err := PostApply(name, projectRoot); err != nil {
		fmt.Fprintf(os.Stderr, "post-apply failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "↩ rolling back this apply (files restored to pre-apply state)")
		if rbErr := snapshot.Restore(); rbErr != nil {
			return fmt.Errorf("post-apply failed (%v) AND rollback failed: %w", err, rbErr)
		}
		fmt.Fprintln(os.Stderr, "✓ rollback complete — project restored")
		return fmt.Errorf("post-apply: %w (apply rolled back)", err)
	}
	return nil
}

// Add generates the integration file and runs go get for the required package.
const prometheusComposeConfigTmpl = `global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "{{.Name}}"
    static_configs:
      - targets: ["{{.Name}}:8080"]
`

// planPrometheusConfig (GIN-006) adds the prometheus.yml creation to the
// plan when the integration is prometheus and the config does not exist yet.
func planPrometheusConfig(prj *project.Project, p *plan.Plan) error {
	configPath := filepath.Join(prj.Root, "devops", "docker", "prometheus.yml")
	if _, err := os.Stat(configPath); err == nil || !os.IsNotExist(err) {
		return nil
	}

	// App name: infer from the compose when it exists, fallback "app".
	appName := "app"
	composePath := filepath.Join(prj.Root, "devops", "docker", "docker-compose.yml")
	if data, err := os.ReadFile(composePath); err == nil {
		var compose composeFile
		if yaml.Unmarshal(data, &compose) == nil && compose.Services != nil {
			appName = detectComposeAppService(compose.Services)
		}
	}

	var rendered strings.Builder
	tmpl, err := template.New("prometheus-compose-config").Parse(prometheusComposeConfigTmpl)
	if err != nil {
		return fmt.Errorf("prometheus config template: %w", err)
	}
	if err := tmpl.Execute(&rendered, struct{ Name string }{Name: appName}); err != nil {
		return fmt.Errorf("prometheus config render: %w", err)
	}
	p.AddCreate(configPath, []byte(rendered.String()), false)
	return nil
}

func detectComposeAppService(services map[string]composeService) string {
	if _, ok := services["app"]; ok {
		return "app"
	}
	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) > 0 {
		return names[0]
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

// Plan builds a root-aware plan for adding an integration. Existing user files
// are modified only when owned by the manifest or through a managed region.
func Plan(name string, prj *project.Project, overwrite bool) (*plan.Plan, error) {
	intg, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf(
			"unknown integration: %s\n\navailable integrations:\n"+
				"  databases  : postgres, mysql, sqlite, sqlserver\n"+
				"  orm        : gorm, sqlx, bun\n"+
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

	p := plan.New(prj.Root)
	p.CreateMissingDirs = prj.YAML.Rules.CreateMissingDirs
	filePath, err := integrationPath(prj, name, intg.file)
	if err != nil {
		return nil, err
	}
	ownership, err := manifest.Load(prj.Root)
	if err != nil {
		return nil, err
	}

	// Render template
	var buf strings.Builder
	tmpl, err := template.New("").Parse(intg.tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	if err := tmpl.Execute(&buf, nil); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	content := []byte(buf.String())
	if name == "swagger" || name == "sse" || name == "websocket" {
		pkg := detectGoPackage(filepath.Dir(filePath), map[string]string{
			"swagger": "api", "sse": "handlers", "websocket": "handlers",
		}[name])
		content = []byte(strings.Replace(string(content), "package "+map[string]string{
			"swagger": "api", "sse": "handlers", "websocket": "handlers",
		}[name], "package "+pkg, 1))
	}
	relMain, _ := filepath.Rel(prj.Root, filePath)
	before := len(p.Changes)
	p.AddCreate(filePath, content, overwrite || ownership.ManagesFullFile(relMain))
	var ownedEntries []manifest.Entry
	if len(p.Changes) > before && isWritable(p.Changes[len(p.Changes)-1]) {
		ownedEntries = append(ownedEntries, manifest.Entry{Path: filepath.ToSlash(relMain), FullFile: true})
	}

	if intg.pkg != "" {
		p.AddWarning(fmt.Sprintf("apply will run 'go get %s' from the project root", intg.pkg))
	}

	// Swagger adds a configured OpenAPI document and proposes router wiring.
	if name == "swagger" {
		docsDir, resolveErr := prj.ResolvePath("docs")
		if resolveErr != nil {
			return nil, resolveErr
		}
		docPath := filepath.Join(docsDir, "openapi.json")
		relDoc, _ := filepath.Rel(prj.Root, docPath)
		docContent, routeCount, docErr := buildOpenAPI(prj)
		if docErr != nil {
			return nil, docErr
		}
		before = len(p.Changes)
		p.AddCreate(docPath, docContent, overwrite || ownership.ManagesFullFile(relDoc))
		if len(p.Changes) > before && isWritable(p.Changes[len(p.Changes)-1]) {
			ownedEntries = append(ownedEntries, manifest.Entry{Path: filepath.ToSlash(relDoc), FullFile: true})
		}
		if routeCount > 0 {
			p.AddWarning(fmt.Sprintf("detected %d route(s); response schemas remain TODO because they could not be inferred safely", routeCount))
		} else {
			p.AddWarning("no routes were inferred; add // ginger:route METHOD /path annotations")
		}
		routerEntry, routerErr := planSwaggerRouter(prj, p, ownership)
		if routerErr != nil {
			return nil, routerErr
		}
		if routerEntry != nil {
			ownedEntries = append(ownedEntries, *routerEntry)
		}
	}

	composeEntry, err := planComposePatch(prj, p, ownership, name)
	if err != nil {
		return nil, err
	}
	if composeEntry != nil {
		ownedEntries = append(ownedEntries, *composeEntry)
	}

	// prometheus.yml entra no plano (antes só o legado criava — gap da
	// migração plan-based). Idempotente: cria apenas se não existir.
	if name == "prometheus" {
		if err := planPrometheusConfig(prj, p); err != nil {
			return nil, err
		}
	}
	if len(ownedEntries) > 0 {
		if err := manifest.PlanUpdate(p, ownedEntries...); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// PostApply installs a declared Go dependency. File changes, including compose
// changes, are always handled by the plan itself.
func PostApply(name, projectRoot string) error {
	intg, ok := registry[name]
	if !ok {
		return fmt.Errorf("unknown integration: %s", name)
	}

	if intg.pkg != "" {
		fmt.Printf("  → go get %s\n", intg.pkg)
		args := append([]string{"get"}, strings.Fields(intg.pkg)...)
		cmd := execCommand("go", args...)
		cmd.Dir = projectRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go get %s: %w", intg.pkg, err)
		}
		fmt.Printf("  ✓ dependency added\n")
	}

	return nil
}

// NeedsPostApply reports whether the plan created or replaced the integration
// implementation and therefore needs dependency installation.
func NeedsPostApply(name string, p *plan.Plan) bool {
	intg, ok := registry[name]
	if !ok || intg.pkg == "" {
		return false
	}
	base := filepath.Base(intg.file)
	for _, change := range p.Changes {
		if isWritable(change) && filepath.Base(change.Path) == base {
			return true
		}
	}
	return false
}

func buildOpenAPI(prj *project.Project) ([]byte, int, error) {
	report, err := project.Inspect(prj)
	if err != nil {
		return nil, 0, err
	}
	paths := make(map[string]map[string]any)
	for _, route := range report.Routes {
		if route.Method == "ANY" {
			continue
		}
		method := strings.ToLower(route.Method)
		if paths[route.Path] == nil {
			paths[route.Path] = make(map[string]any)
		}
		paths[route.Path][method] = map[string]any{
			"summary":         fmt.Sprintf("TODO: document %s %s", route.Method, route.Path),
			"x-ginger-source": fmt.Sprintf("%s:%d", route.File, route.Line),
			"responses":       map[string]any{"200": map[string]any{"description": "TODO: describe response"}},
		}
	}
	document := map[string]any{
		"openapi": "3.0.3",
		"info":    map[string]any{"title": "Existing Go API", "version": "0.1.0"},
		"paths":   paths,
	}
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, 0, fmt.Errorf("marshal OpenAPI document: %w", err)
	}
	return append(data, '\n'), len(paths), nil
}

func integrationPath(prj *project.Project, name, registered string) (string, error) {
	switch name {
	case "swagger":
		dir, err := prj.ResolvePath("api")
		return filepath.Join(dir, "swagger.go"), err
	case "sse", "websocket":
		dir, err := prj.ResolvePath("handlers")
		return filepath.Join(dir, filepath.Base(registered)), err
	default:
		return prj.ResolveRelative(registered)
	}
}

func detectGoPackage(dir, fallback string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fallback
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, entry.Name()), nil, parser.PackageClauseOnly)
		if parseErr == nil && file.Name != nil {
			return file.Name.Name
		}
	}
	return fallback
}

func planSwaggerRouter(prj *project.Project, p *plan.Plan, ownership *manifest.Manifest) (*manifest.Entry, error) {
	apiDir, err := prj.ResolvePath("api")
	if err != nil {
		return nil, err
	}
	routerPath := filepath.Join(apiDir, "router.go")
	data, err := os.ReadFile(routerPath)
	if os.IsNotExist(err) {
		p.AddWarning("router.go was not found; register SwaggerUI and OpenAPISpec manually")
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read router for swagger: %w", err)
	}
	content := string(data)
	if strings.Contains(content, "registerSwaggerRoutes(r)") {
		p.AddWarning("swagger routes are already registered")
		return nil, nil
	}
	rel, _ := filepath.Rel(prj.Root, routerPath)
	gingerRouter := strings.Contains(content, "github.com/fvmoraes/ginger/pkg/router") || strings.Contains(content, "*router.Router")
	if current := region.FindRegion(content, "routes"); current != nil && gingerRouter && strings.Contains(current.Content, "r.") {
		replacement := strings.TrimRight(current.Content, "\n") + "\n\tregisterSwaggerRoutes(r)"
		updated, replaceErr := region.ReplaceRegion(content, "routes", replacement)
		if replaceErr != nil {
			return nil, replaceErr
		}
		p.AddModify(routerPath, []byte(updated), true)
		entry := manifest.Entry{Path: filepath.ToSlash(rel), Regions: []string{"routes"}}
		return &entry, nil
	}

	patchPath := filepath.Join(prj.Root, ".ginger", "patches", filepath.FromSlash(filepath.ToSlash(rel)+".patch"))
	patch := fmt.Sprintf("# Ginger suggestion for %s\n# No source file was modified.\n\nInside your router setup, register:\n\n    registerSwaggerRoutes(r)\n", filepath.ToSlash(rel))
	patchRel, _ := filepath.Rel(prj.Root, patchPath)
	p.AddCreate(patchPath, []byte(patch), ownership.ManagesFullFile(patchRel))
	p.AddWarning(fmt.Sprintf("router has no managed 'routes' region; review %s", filepath.ToSlash(patchRel)))
	entry := manifest.Entry{Path: filepath.ToSlash(patchRel), FullFile: true}
	return &entry, nil
}

func planComposePatch(prj *project.Project, p *plan.Plan, ownership *manifest.Manifest, name string) (*manifest.Entry, error) {
	composePath := filepath.Join(prj.Root, "devops", "docker", "docker-compose.yml")
	data, err := os.ReadFile(composePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read compose file: %w", err)
	}
	var compose composeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("parse compose file: %w", err)
	}
	if compose.Services == nil {
		compose.Services = make(map[string]composeService)
	}
	if compose.Volumes == nil {
		compose.Volumes = make(map[string]map[string]string)
	}
	appName := detectComposeAppService(compose.Services)
	app := compose.Services[appName]

	// GIN-026: snapshot published ports BEFORE the merge to detect collisions
	// between the integration's services and everything already declared.
	publishedBefore := map[string]string{} // hostPort → owning service
	for svcName, svc := range compose.Services {
		for _, mapping := range svc.Ports {
			publishedBefore[hostPort(mapping)] = svcName
		}
	}

	if !mergeIntegrationIntoCompose(&compose, appName, &app, name) {
		return nil, nil
	}
	compose.Services[appName] = app

	for _, conflict := range detectPortConflicts(publishedBefore, DockerServicesByIntegration[name]) {
		p.AddWarning(fmt.Sprintf("port conflict: %s is already published by service %q — adjust one of them before running docker compose", conflict.hostPort, conflict.owner))
	}

	out, err := yaml.Marshal(&compose)
	if err != nil {
		return nil, err
	}
	rel, _ := filepath.Rel(prj.Root, composePath)
	if ownership.ManagesFullFile(rel) {
		// GIN-002: merge condicional por hash de proveniência.
		// - intacto desde a geração (hash atual == hash gravado) → merge direto
		//   (fluxo de scaffold novo preservado);
		// - modificado pelo usuário OU hash ausente (manifest antigo) → patch
		//   revisável (data-safe: nunca destruir conteúdo customizado).
		currentHash := hexSha256(data)
		recorded := ownership.GeneratedHash(rel)
		switch {
		case recorded != "" && recorded == currentHash:
			p.AddModify(composePath, out, true)
			return &manifest.Entry{Path: filepath.ToSlash(rel), FullFile: true, GeneratedHash: hexSha256From(out)}, nil
		default:
			if recorded == "" {
				p.AddWarning("compose provenance hash is missing (older manifest); proposing a patch instead of rewriting")
			} else {
				p.AddWarning("compose file was customized after generation; proposing a patch instead of rewriting (GIN-002)")
			}
			patchPath := filepath.Join(prj.Root, ".ginger", "patches", filepath.FromSlash(filepath.ToSlash(rel)+".patch"))
			patchRel, _ := filepath.Rel(prj.Root, patchPath)
			p.AddCreate(patchPath, out, ownership.ManagesFullFile(patchRel))
			return &manifest.Entry{Path: filepath.ToSlash(patchRel), FullFile: true}, nil
		}
	}
	patchPath := filepath.Join(prj.Root, ".ginger", "patches", filepath.FromSlash(filepath.ToSlash(rel)+".patch"))
	patchRel, _ := filepath.Rel(prj.Root, patchPath)
	p.AddCreate(patchPath, out, ownership.ManagesFullFile(patchRel))
	p.AddWarning(fmt.Sprintf("compose file is not Ginger-owned; proposed content written to %s", filepath.ToSlash(patchRel)))
	return &manifest.Entry{Path: filepath.ToSlash(patchRel), FullFile: true}, nil
}

func isWritable(change plan.PlannedChange) bool {
	return change.Type == plan.ChangeCreate || change.Type == plan.ChangeModify
}
