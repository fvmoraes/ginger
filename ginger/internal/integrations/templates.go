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
