package integrations

const postgresTmpl = `// Package database provides a PostgreSQL connection helper.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Config holds PostgreSQL connection settings.
type Config struct {
	DSN     string
	MaxOpen int
	MaxIdle int
}

// Connect opens and validates a PostgreSQL connection.
func Connect(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return db, nil
}
`

const redisTmpl = `// Package cache provides a Redis client helper.
package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis connection settings.
type Config struct {
	Addr     string
	Password string
	DB       int
}

// Connect creates and validates a Redis client.
func Connect(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis: ping: %w", err)
	}
	return client, nil
}

// Checker implements health.Checker for Redis.
type Checker struct{ client *redis.Client }

func NewChecker(c *redis.Client) *Checker { return &Checker{client: c} }
func (c *Checker) Name() string           { return "redis" }
func (c *Checker) Check(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
`

const kafkaTmpl = `// Package messaging provides a Kafka producer/consumer helper.
package messaging

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

// ProducerConfig holds Kafka producer settings.
type ProducerConfig struct {
	Brokers []string
	Topic   string
}

// NewWriter creates a Kafka writer (producer).
func NewWriter(cfg ProducerConfig) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
}

// ConsumerConfig holds Kafka consumer settings.
type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// NewReader creates a Kafka reader (consumer).
func NewReader(cfg ConsumerConfig) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
}

// Publish sends a message to Kafka.
func Publish(ctx context.Context, w *kafka.Writer, key, value []byte) error {
	return w.WriteMessages(ctx, kafka.Message{Key: key, Value: value})
}
`

const otelTmpl = `// Package telemetry provides OpenTelemetry setup with OTLP exporter.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Setup initializes the global OTel tracer provider with OTLP HTTP exporter.
// Set OTEL_EXPORTER_OTLP_ENDPOINT env var to your collector endpoint.
func Setup(ctx context.Context, serviceName, version string) (func(context.Context) error, error) {
	exp, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("otel: exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel: resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

// Tracer returns a named tracer.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
`

const prometheusTmpl = `// Package metrics provides Prometheus metrics helpers.
package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// DefaultMetrics holds standard HTTP metrics.
type DefaultMetrics struct {
	RequestDuration *prometheus.HistogramVec
	RequestTotal    *prometheus.CounterVec
	ErrorTotal      *prometheus.CounterVec
}

// NewDefaultMetrics registers and returns standard HTTP metrics.
func NewDefaultMetrics(namespace string) *DefaultMetrics {
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	total := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	errors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_errors_total",
		Help:      "Total number of HTTP errors.",
	}, []string{"method", "path"})

	prometheus.MustRegister(duration, total, errors)

	return &DefaultMetrics{
		RequestDuration: duration,
		RequestTotal:    total,
		ErrorTotal:      errors,
	}
}

// Handler returns the Prometheus metrics HTTP handler (mount at /metrics).
func Handler() http.Handler {
	return promhttp.Handler()
}

// Middleware records request duration and count for each route.
func Middleware(m *DefaultMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			dur := time.Since(start).Seconds()
			status := http.StatusText(rw.status)
			m.RequestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(dur)
			m.RequestTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
`

// ─── Messaging ───────────────────────────────────────────────────────────────

const rabbitmqTmpl = `// Package messaging provides a RabbitMQ connection and channel helper.
package messaging

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ wraps an AMQP connection and channel.
type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

// ConnectRabbitMQ dials RabbitMQ and opens a channel.
// dsn example: amqp://guest:guest@localhost:5672/
func ConnectRabbitMQ(dsn string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq: channel: %w", err)
	}
	return &RabbitMQ{conn: conn, Channel: ch}, nil
}

// Publish sends a message to the given exchange/routing key.
func (r *RabbitMQ) Publish(ctx context.Context, exchange, key string, body []byte) error {
	return r.Channel.PublishWithContext(ctx, exchange, key, false, false,
		amqp.Publishing{ContentType: "application/json", Body: body},
	)
}

// Close releases the channel and connection.
func (r *RabbitMQ) Close() error {
	if err := r.Channel.Close(); err != nil {
		return err
	}
	return r.conn.Close()
}

// Checker implements health.Checker for RabbitMQ.
type RabbitMQChecker struct{ r *RabbitMQ }

func NewRabbitMQChecker(r *RabbitMQ) *RabbitMQChecker { return &RabbitMQChecker{r: r} }
func (c *RabbitMQChecker) Name() string                { return "rabbitmq" }
func (c *RabbitMQChecker) Check(_ context.Context) error {
	if c.r.conn.IsClosed() {
		return fmt.Errorf("rabbitmq: connection closed")
	}
	return nil
}
`

const natsTmpl = `// Package messaging provides a NATS connection helper.
package messaging

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

// ConnectNATS connects to a NATS server.
// url example: nats://localhost:4222
func ConnectNATS(url string) (*nats.Conn, error) {
	nc, err := nats.Connect(url, nats.MaxReconnects(5))
	if err != nil {
		return nil, fmt.Errorf("nats: connect: %w", err)
	}
	return nc, nil
}

// Publish sends a message to a NATS subject.
func Publish(nc *nats.Conn, subject string, data []byte) error {
	return nc.Publish(subject, data)
}

// Subscribe registers a handler for a NATS subject.
func Subscribe(nc *nats.Conn, subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return nc.Subscribe(subject, handler)
}

// NATSChecker implements health.Checker for NATS.
type NATSChecker struct{ nc *nats.Conn }

func NewNATSChecker(nc *nats.Conn) *NATSChecker { return &NATSChecker{nc: nc} }
func (c *NATSChecker) Name() string              { return "nats" }
func (c *NATSChecker) Check(_ context.Context) error {
	if !c.nc.IsConnected() {
		return fmt.Errorf("nats: not connected")
	}
	return nil
}
`

const pubsubTmpl = `// Package messaging provides a Google Cloud Pub/Sub helper.
package messaging

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

// PubSubClient wraps a Google Cloud Pub/Sub client.
type PubSubClient struct {
	client *pubsub.Client
}

// ConnectPubSub creates a Pub/Sub client for the given GCP project.
// Requires GOOGLE_APPLICATION_CREDENTIALS or Workload Identity.
func ConnectPubSub(ctx context.Context, projectID string) (*PubSubClient, error) {
	c, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub: client: %w", err)
	}
	return &PubSubClient{client: c}, nil
}

// Publish sends a message to the given topic.
func (p *PubSubClient) Publish(ctx context.Context, topicID string, data []byte) (string, error) {
	t := p.client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{Data: data})
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("pubsub: publish: %w", err)
	}
	return id, nil
}

// Subscribe pulls messages from a subscription.
func (p *PubSubClient) Subscribe(ctx context.Context, subID string, fn func(ctx context.Context, msg *pubsub.Message)) error {
	sub := p.client.Subscription(subID)
	return sub.Receive(ctx, fn)
}

// Close releases the Pub/Sub client.
func (p *PubSubClient) Close() error { return p.client.Close() }
`

// ─── Relational databases ─────────────────────────────────────────────────────

const mysqlTmpl = `// Package database provides a MySQL connection helper.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLConfig holds MySQL connection settings.
type MySQLConfig struct {
	DSN     string // user:pass@tcp(host:3306)/dbname?parseTime=true
	MaxOpen int
	MaxIdle int
}

// ConnectMySQL opens and validates a MySQL connection.
func ConnectMySQL(cfg MySQLConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("mysql: open: %w", err)
	}
	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("mysql: ping: %w", err)
	}
	return db, nil
}
`

const sqliteTmpl = `// Package database provides a SQLite connection helper.
package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// ConnectSQLite opens a SQLite database file.
// path example: "./data.db" or ":memory:"
func ConnectSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("sqlite: ping: %w", err)
	}
	return db, nil
}
`

const sqlserverTmpl = `// Package database provides a SQL Server connection helper.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

// SQLServerConfig holds SQL Server connection settings.
type SQLServerConfig struct {
	DSN     string // sqlserver://user:pass@host:1433?database=dbname
	MaxOpen int
	MaxIdle int
}

// ConnectSQLServer opens and validates a SQL Server connection.
func ConnectSQLServer(cfg SQLServerConfig) (*sql.DB, error) {
	db, err := sql.Open("sqlserver", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("sqlserver: open: %w", err)
	}
	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("sqlserver: ping: %w", err)
	}
	return db, nil
}
`

// ─── gRPC ─────────────────────────────────────────────────────────────────────

const grpcTmpl = `// Package grpc provides a gRPC server and client helper for Ginger projects.
package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// ServerConfig holds gRPC server settings.
type ServerConfig struct {
	Addr string // e.g. ":50051"
}

// Server wraps a *grpc.Server with lifecycle helpers.
type Server struct {
	srv  *grpc.Server
	addr string
}

// NewServer creates a gRPC server with health check and reflection enabled.
// Pass additional grpc.ServerOption values to customise (e.g. TLS, interceptors).
func NewServer(cfg ServerConfig, opts ...grpc.ServerOption) *Server {
	srv := grpc.NewServer(opts...)

	// Standard health check service (used by k8s probes, grpc-health-probe, etc.)
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())

	// Reflection lets tools like grpcurl discover services without .proto files.
	reflection.Register(srv)

	return &Server{srv: srv, addr: cfg.Addr}
}

// Register adds a gRPC service to the server.
// Call this before Serve.
func (s *Server) Register(fn func(*grpc.Server)) {
	fn(s.srv)
}

// Serve starts listening and blocks until the server stops.
func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("grpc: listen %s: %w", s.addr, err)
	}
	return s.srv.Serve(lis)
}

// Stop performs a graceful shutdown.
func (s *Server) Stop() { s.srv.GracefulStop() }

// NewClient dials a gRPC server and returns the connection.
// Uses insecure credentials by default — swap for TLS in production.
func NewClient(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	defaults := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.DialContext(ctx, target, append(defaults, opts...)...)
	if err != nil {
		return nil, fmt.Errorf("grpc: dial %s: %w", target, err)
	}
	return conn, nil
}
`

// ─── MCP (Model Context Protocol) ────────────────────────────────────────────

const mcpTmpl = `// Package mcp provides a minimal Model Context Protocol (MCP) server for
// exposing tools and resources to LLM clients (e.g. Claude, Cursor, Kiro).
//
// The MCP spec: https://modelcontextprotocol.io/specification
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Tool represents an MCP tool that can be called by an LLM client.
type Tool struct {
	Name        string          // unique tool identifier
	Description string          // shown to the LLM
	InputSchema json.RawMessage // JSON Schema for the input object
	Handler     ToolHandler     // called when the tool is invoked
}

// ToolHandler is the function signature for an MCP tool implementation.
type ToolHandler func(ctx context.Context, input json.RawMessage) (any, error)

// Server is a minimal MCP server that exposes tools over HTTP (SSE + JSON-RPC).
type Server struct {
	tools map[string]Tool
}

// NewServer creates an MCP server.
func NewServer() *Server {
	return &Server{tools: make(map[string]Tool)}
}

// Register adds a tool to the server.
func (s *Server) Register(t Tool) {
	s.tools[t.Name] = t
}

// Handler returns an http.Handler that implements the MCP HTTP transport.
// Mount it at a path, e.g.: mux.Handle("/mcp", mcpServer.Handler())
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/mcp/tools/list":
			s.handleList(w, r)
		case "/mcp/tools/call":
			s.handleCall(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// toolInfo is the wire format for a tool in the list response.
type toolInfo struct {
	Name        string          ` + "`json:\"name\"`" + `
	Description string          ` + "`json:\"description\"`" + `
	InputSchema json.RawMessage ` + "`json:\"inputSchema\"`" + `
}

func (s *Server) handleList(w http.ResponseWriter, _ *http.Request) {
	list := make([]toolInfo, 0, len(s.tools))
	for _, t := range s.tools {
		list = append(list, toolInfo{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"tools": list})
}

type callRequest struct {
	Name  string          ` + "`json:\"name\"`" + `
	Input json.RawMessage ` + "`json:\"input\"`" + `
}

func (s *Server) handleCall(w http.ResponseWriter, r *http.Request) {
	var req callRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResponse(fmt.Sprintf("invalid request: %v", err)))
		return
	}

	t, ok := s.tools[req.Name]
	if !ok {
		writeJSON(w, http.StatusNotFound, errResponse(fmt.Sprintf("tool not found: %s", req.Name)))
		return
	}

	result, err := t.Handler(r.Context(), req.Input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResponse("internal server error"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"result": result})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func errResponse(msg string) map[string]string {
	return map[string]string{"error": msg}
}
`

// ─── UI / Real-time ──────────────────────────────────────────────────────────

const sseTmpl = `// Package realtime provides a Server-Sent Events (SSE) handler example.
// For the full SSE helper, see github.com/fvmoraes/ginger/pkg/sse.
package realtime

import (
	"net/http"
	"time"

	"github.com/fvmoraes/ginger/pkg/sse"
)

// LiveFeedHandler streams real-time events to the client.
// Mount at: GET /api/v1/events
func LiveFeedHandler(w http.ResponseWriter, r *http.Request) {
	stream, err := sse.New(w)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Send an initial connected event.
	_ = stream.Send(sse.Event{Type: "connected", Data: map[string]string{"status": "ok"}})

	for {
		select {
		case <-r.Context().Done():
			return
		case t := <-ticker.C:
			_ = stream.Send(sse.Event{
				Type: "tick",
				Data: map[string]string{"time": t.Format(time.RFC3339)},
			})
		}
	}
}
`

const wsTmpl = `// Package realtime provides a WebSocket handler example.
// For the full WebSocket helper, see github.com/fvmoraes/ginger/pkg/ws.
package realtime

import (
	"net/http"

	"github.com/fvmoraes/ginger/pkg/ws"
)

// EchoHandler upgrades the connection to WebSocket and echoes every message.
// Mount at: GET /api/v1/ws
func EchoHandler(w http.ResponseWriter, r *http.Request) {
	ws.Handle(w, r, func(conn *ws.Conn) {
		for {
			var msg ws.Message
			if err := conn.Recv(&msg); err != nil {
				return // client disconnected
			}
			// Echo the message back with type "echo".
			if err := conn.Send(ws.Message{Type: "echo", Data: msg.Data}); err != nil {
				return
			}
		}
	})
}
`

// ─── ORM / Query builders ─────────────────────────────────────────────────────

const gormTmpl = `// Package database provides a GORM setup helper.
// GORM supports PostgreSQL, MySQL, SQLite and SQL Server via driver packages.
//
// Usage:
//
//	db, err := database.OpenGORM(database.GORMConfig{
//	    Dialector: postgres.Open(dsn),
//	    LogLevel:  logger.Info,
//	})
package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GORMConfig holds GORM connection settings.
type GORMConfig struct {
	// Dialector is the database-specific driver, e.g.:
	//   postgres.Open(dsn)  — gorm.io/driver/postgres
	//   mysql.Open(dsn)     — gorm.io/driver/mysql
	//   sqlite.Open(file)   — gorm.io/driver/sqlite
	//   sqlserver.Open(dsn) — gorm.io/driver/sqlserver
	Dialector gorm.Dialector
	LogLevel  logger.LogLevel // logger.Silent | Info | Warn | Error
	MaxOpen   int
	MaxIdle   int
}

// OpenGORM opens a GORM database connection and configures the connection pool.
func OpenGORM(cfg GORMConfig) (*gorm.DB, error) {
	lvl := cfg.LogLevel
	if lvl == 0 {
		lvl = logger.Warn
	}

	db, err := gorm.Open(cfg.Dialector, &gorm.Config{
		Logger: logger.Default.LogMode(lvl),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm: open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gorm: get sql.DB: %w", err)
	}

	if cfg.MaxOpen > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	}
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}
`

const sqlxTmpl = `// Package database provides a sqlx setup helper.
// sqlx extends database/sql with named queries, struct scanning and more.
//
// Usage:
//
//	db, err := database.OpenSQLx(database.SQLxConfig{
//	    Driver: "postgres",
//	    DSN:    "postgres://user:pass@localhost/db?sslmode=disable",
//	})
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"           // postgres  — swap as needed
	// _ "github.com/go-sql-driver/mysql"   // mysql
	// _ "github.com/mattn/go-sqlite3"      // sqlite
	// _ "github.com/microsoft/go-mssqldb"  // sqlserver
)

// SQLxConfig holds sqlx connection settings.
type SQLxConfig struct {
	Driver  string // "postgres" | "mysql" | "sqlite3" | "sqlserver"
	DSN     string
	MaxOpen int
	MaxIdle int
}

// OpenSQLx opens a sqlx database connection and validates it with a ping.
func OpenSQLx(cfg SQLxConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("sqlx: connect: %w", err)
	}

	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}
	db.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}

// SQLxChecker implements health.Checker for a *sqlx.DB.
type SQLxChecker struct{ db *sqlx.DB }

func NewSQLxChecker(db *sqlx.DB) *SQLxChecker { return &SQLxChecker{db: db} }
func (c *SQLxChecker) Name() string            { return "database" }
func (c *SQLxChecker) Check(ctx context.Context) error {
	return c.db.PingContext(ctx)
}
`

const bunTmpl = `// Package database provides a Bun ORM setup helper.
// Bun supports PostgreSQL, MySQL, SQLite and SQL Server.
//
// Usage:
//
//	db, err := database.OpenBun(database.BunConfig{
//	    Driver: "postgres",
//	    DSN:    "postgres://user:pass@localhost/db?sslmode=disable",
//	})
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	// Swap dialect/driver for other databases:
	// "github.com/uptrace/bun/dialect/mysqldialect"
	// "github.com/uptrace/bun/driver/sqliteshim"
	// "github.com/uptrace/bun/dialect/sqlitedialect"
)

// BunConfig holds Bun connection settings.
type BunConfig struct {
	DSN     string
	MaxOpen int
	MaxIdle int
}

// OpenBun opens a Bun database connection backed by PostgreSQL.
// Swap pgdriver/pgdialect for your target database.
func OpenBun(cfg BunConfig) (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.DSN)))

	if cfg.MaxOpen > 0 {
		sqldb.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		sqldb.SetMaxIdleConns(cfg.MaxIdle)
	}
	sqldb.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqldb.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("bun: ping: %w", err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())
	return db, nil
}

// BunChecker implements health.Checker for a *bun.DB.
type BunChecker struct{ db *bun.DB }

func NewBunChecker(db *bun.DB) *BunChecker { return &BunChecker{db: db} }
func (c *BunChecker) Name() string         { return "database" }
func (c *BunChecker) Check(ctx context.Context) error {
	return c.db.PingContext(ctx)
}
`

// ─── NoSQL / Analytical ───────────────────────────────────────────────────────

const couchbaseTmpl = `// Package nosql provides a Couchbase connection helper.
// Uses the official Couchbase Go SDK v2.
//
// Usage:
//
//	cluster, bucket, err := nosql.ConnectCouchbase(nosql.CouchbaseConfig{
//	    ConnectionString: "couchbase://localhost",
//	    Username:         "Administrator",
//	    Password:         "password",
//	    BucketName:       "my-bucket",
//	})
package nosql

import (
	"context"
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"
)

// CouchbaseConfig holds Couchbase connection settings.
type CouchbaseConfig struct {
	ConnectionString string // e.g. "couchbase://localhost"
	Username         string
	Password         string
	BucketName       string
}

// ConnectCouchbase connects to a Couchbase cluster and opens a bucket.
func ConnectCouchbase(cfg CouchbaseConfig) (*gocb.Cluster, *gocb.Bucket, error) {
	cluster, err := gocb.Connect(cfg.ConnectionString, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("couchbase: connect: %w", err)
	}

	bucket := cluster.Bucket(cfg.BucketName)
	if err := bucket.WaitUntilReady(10*time.Second, nil); err != nil {
		return nil, nil, fmt.Errorf("couchbase: bucket not ready: %w", err)
	}

	return cluster, bucket, nil
}

// CouchbaseChecker implements health.Checker for a Couchbase cluster.
type CouchbaseChecker struct{ cluster *gocb.Cluster }

func NewCouchbaseChecker(c *gocb.Cluster) *CouchbaseChecker { return &CouchbaseChecker{cluster: c} }
func (c *CouchbaseChecker) Name() string                     { return "couchbase" }
func (c *CouchbaseChecker) Check(_ context.Context) error {
	_, err := c.cluster.Ping(nil)
	if err != nil {
		return fmt.Errorf("couchbase: ping: %w", err)
	}
	return nil
}
`

const clickhouseTmpl = `// Package database provides a ClickHouse connection helper.
// Uses the official ClickHouse Go driver v2 (database/sql interface).
//
// Usage:
//
//	db, err := database.ConnectClickHouse(database.ClickHouseConfig{
//	    Addr:     []string{"localhost:9000"},
//	    Database: "default",
//	    Username: "default",
//	    Password: "",
//	})
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// ClickHouseConfig holds ClickHouse connection settings.
type ClickHouseConfig struct {
	Addr     []string // e.g. []string{"localhost:9000"}
	Database string
	Username string
	Password string
	MaxOpen  int
	MaxIdle  int
}

// ConnectClickHouse opens and validates a ClickHouse connection via database/sql.
func ConnectClickHouse(cfg ClickHouseConfig) (*sql.DB, error) {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: cfg.Addr,
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout:  5 * time.Second,
		MaxOpenConns: cfg.MaxOpen,
		MaxIdleConns: cfg.MaxIdle,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("clickhouse: ping: %w", err)
	}
	return conn, nil
}

// ClickHouseChecker implements health.Checker for ClickHouse.
type ClickHouseChecker struct{ db *sql.DB }

func NewClickHouseChecker(db *sql.DB) *ClickHouseChecker { return &ClickHouseChecker{db: db} }
func (c *ClickHouseChecker) Name() string                { return "clickhouse" }
func (c *ClickHouseChecker) Check(ctx context.Context) error {
	return c.db.PingContext(ctx)
}
`

const mongoTmpl = `// Package nosql provides a MongoDB connection helper.
// Uses the official MongoDB Go driver.
//
// Usage:
//
//	client, db, err := nosql.ConnectMongo(nosql.MongoConfig{
//	    URI:      "mongodb://localhost:27017",
//	    Database: "mydb",
//	})
package nosql

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
)

// MongoConfig holds MongoDB connection settings.
type MongoConfig struct {
	URI      string // e.g. "mongodb://localhost:27017"
	Database string
}

// ConnectMongo connects to MongoDB and returns the client and the target database.
func ConnectMongo(cfg MongoConfig) (*mongo.Client, *mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, nil, fmt.Errorf("mongo: connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, fmt.Errorf("mongo: ping: %w", err)
	}

	return client, client.Database(cfg.Database), nil
}

// MongoChecker implements health.Checker for MongoDB.
type MongoChecker struct{ client *mongo.Client }

func NewMongoChecker(c *mongo.Client) *MongoChecker { return &MongoChecker{client: c} }
func (c *MongoChecker) Name() string                { return "mongodb" }
func (c *MongoChecker) Check(ctx context.Context) error {
	return c.client.Database("admin").RunCommand(ctx, bson.D{{ Key: "ping", Value: 1 }}).Err()
}
`

const swaggerTmpl = `// Package handlers provides Swagger/OpenAPI endpoints.
package handlers

import (
	"os"
	"net/http"
)

// RegisterSwagger mounts a simple Swagger UI and an OpenAPI spec endpoint.
// Mount examples:
//   r.GET("/swagger", handlers.SwaggerUI)
//   r.GET("/swagger/openapi.json", handlers.OpenAPISpec)
func RegisterSwagger(r interface {
	GET(pattern string, h http.HandlerFunc)
}) {
	r.GET("/swagger", SwaggerUI)
	r.GET("/swagger/openapi.json", OpenAPISpec)
}

// SwaggerUI serves a lightweight Swagger UI page using the public CDN assets.
func SwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerHTML))
}

// OpenAPISpec serves a starter OpenAPI 3.0 document that teams can customize.
func OpenAPISpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	spec, err := os.ReadFile("docs/openapi.json")
	if err != nil {
		spec = []byte(openAPISpec)
	}
	_, _ = w.Write(spec)
}

const swaggerHTML = ` + "`" + `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/swagger/openapi.json',
      dom_id: '#swagger-ui'
    });
  </script>
</body>
</html>` + "`" + `

const openAPISpec = ` + "`" + `{
  "openapi": "3.0.3",
  "info": {
    "title": "Ginger API",
    "version": "1.0.0",
    "description": "Starter OpenAPI document generated by ginger add swagger"
  },
  "paths": {
    "/health": {
      "get": {
        "summary": "Health check",
        "responses": {
          "200": {
            "description": "Healthy"
          }
        }
      }
    }
  }
}` + "`" + `
`
