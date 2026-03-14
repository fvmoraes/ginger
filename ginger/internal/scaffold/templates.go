package scaffold

const goModTmpl = `module {{.Module}}

go 1.25
`

const mainTmpl = `package main

import (
	"log"

	"{{.Module}}/internal/config"
	"{{.Module}}/internal/api/handlers"
	gingerapp "github.com/ginger-framework/ginger/pkg/app"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	app := gingerapp.New(cfg)

	// Register routes
	handlers.Register(app.Router)

	if err := app.Run(); err != nil {
		log.Fatalf("app: %v", err)
	}
}
`

const cliMainTmpl = `package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: {{.Name}} <command>")
		os.Exit(1)
	}
	fmt.Printf("{{.Name}} CLI — command: %s\n", os.Args[1])
}
`

const workerMainTmpl = `package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"{{.Module}}/internal/worker"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	w := worker.New()
	log.Println("worker starting...")
	if err := w.Run(ctx); err != nil {
		log.Fatalf("worker: %v", err)
	}
	log.Println("worker stopped")
	os.Exit(0)
}
`

const workerTmpl = `package worker

import "context"

// Worker is the background job processor.
type Worker struct{}

func New() *Worker { return &Worker{} }

// Run starts the worker loop and blocks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		// TODO: add job processing logic here
		}
	}
}
`

const internalConfigTmpl = `package config

import (
	gingercfg "github.com/ginger-framework/ginger/pkg/config"
)

// Load reads the application configuration.
func Load() (*gingercfg.Config, error) {
	return gingercfg.Load("configs/app.yaml")
}
`

const healthHandlerTmpl = `package handlers

import (
	"net/http"

	"github.com/ginger-framework/ginger/pkg/router"
)

// Register mounts all application routes.
func Register(r *router.Router) {
	r.GET("/", func(w http.ResponseWriter, req *http.Request) {
		router.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}
`

const appYamlTmpl = `app:
  name: {{.Name}}
  env: development
  version: 0.1.0

http:
  host: 0.0.0.0
  port: 8080
  shutdown_timeout: 30

log:
  level: info
  format: json
`

const envExampleTmpl = `APP_NAME={{.Name}}
APP_ENV=development
HTTP_PORT=8080
LOG_LEVEL=info
LOG_FORMAT=json
DATABASE_DRIVER=postgres
DATABASE_DSN=postgres://user:pass@localhost:5432/{{.Name}}?sslmode=disable
`

const dockerfileTmpl = `FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/app ./cmd/app

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/bin/app .
COPY --from=builder /app/configs ./configs
EXPOSE 8080
ENTRYPOINT ["./app"]
`

const dockerComposeTmpl = `version: "3.9"

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      APP_ENV: development
      HTTP_PORT: 8080
      DATABASE_DSN: postgres://user:pass@postgres:5432/{{.Name}}?sslmode=disable
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: {{.Name}}
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./scripts/prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    depends_on:
      - prometheus

volumes:
  pgdata:
`

const k8sDeploymentTmpl = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
  labels:
    app: {{.Name}}
spec:
  replicas: 2
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
    spec:
      containers:
        - name: {{.Name}}
          image: {{.Name}}:latest
          ports:
            - containerPort: 8080
          envFrom:
            - configMapRef:
                name: {{.Name}}-config
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 15
            periodSeconds: 20
          resources:
            requests:
              cpu: "100m"
              memory: "64Mi"
            limits:
              cpu: "500m"
              memory: "256Mi"
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
spec:
  selector:
    app: {{.Name}}
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
`

const helmChartTmpl = `apiVersion: v2
name: {{.Name}}
description: Helm chart for {{.Name}}
type: application
version: 0.1.0
appVersion: "0.1.0"
`

const helmValuesTmpl = `replicaCount: 2

image:
  repository: {{.Name}}
  tag: latest
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

resources:
  requests:
    cpu: 100m
    memory: 64Mi
  limits:
    cpu: 500m
    memory: 256Mi

env:
  APP_ENV: production
  HTTP_PORT: "8080"
`

const helmDeploymentTmpl = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ "{{" }} .Release.Name {{ "}}" }}
spec:
  replicas: {{ "{{" }} .Values.replicaCount {{ "}}" }}
  selector:
    matchLabels:
      app: {{ "{{" }} .Release.Name {{ "}}" }}
  template:
    metadata:
      labels:
        app: {{ "{{" }} .Release.Name {{ "}}" }}
    spec:
      containers:
        - name: {{ "{{" }} .Release.Name {{ "}}" }}
          image: "{{ "{{" }} .Values.image.repository {{ "}}" }}:{{ "{{" }} .Values.image.tag {{ "}}" }}"
          ports:
            - containerPort: {{ "{{" }} .Values.service.targetPort {{ "}}" }}
          readinessProbe:
            httpGet:
              path: /health
              port: {{ "{{" }} .Values.service.targetPort {{ "}}" }}
          resources:
            {{- "{{" }} toYaml .Values.resources | nindent 12 {{ "}}" }}
`

const makefileTmpl = `BIN=bin/app

.PHONY: run build test lint tidy docker up down

run:
	go run ./cmd/app

build:
	go build -o $(BIN) ./cmd/app

test:
	go test ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

docker:
	docker build -t {{.Name}}:latest .

up:
	docker compose up -d

down:
	docker compose down
`

const gitignoreTmpl = `bin/
*.env
.env
*.log
.DS_Store
.vscode/
.idea/
`

const readmeTmpl = `# {{.Name}}

A Ginger-powered Go application (type: {{.Type}}).

## Getting started

` + "```" + `bash
go mod tidy
make run
` + "```" + `

## Docker

` + "```" + `bash
make up    # starts app + postgres + redis + prometheus + grafana
make down  # stops everything
` + "```" + `

## Project structure

` + "```" + `
cmd/app/          # Application entrypoint
internal/
  api/
    handlers/     # HTTP handlers
    services/     # Business logic
    repositories/ # Data access
    middlewares/  # App-specific middlewares
  models/         # Domain models
  config/         # Config loader
configs/          # YAML config files
platform/         # External integrations
kubernetes/       # K8s manifests
helm/             # Helm chart
tests/            # Integration tests
` + "```" + `
`
