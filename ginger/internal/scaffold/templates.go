package scaffold

const goModTmpl = `module {{.Module}}

go 1.24
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

const dockerfileTmpl = `FROM golang:1.24-alpine AS builder
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

const makefileTmpl = `BIN=bin/app

.PHONY: run build test lint tidy

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
`

const gitignoreTmpl = `bin/
*.env
.env
*.log
`

const readmeTmpl = `# {{.Name}}

A Ginger-powered Go application.

## Getting started

` + "```" + `bash
go mod tidy
make run
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
tests/            # Integration tests
` + "```" + `
`
