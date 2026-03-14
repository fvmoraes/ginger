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
		writeJSON(w, http.StatusInternalServerError, errResponse(err.Error()))
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
