// Package ginger provides a lightweight, opinionated Go framework for building web applications and APIs.
//
// Ginger accelerates and standardizes Go project development with a complete CLI tool,
// 13 production-ready packages, and 20+ integrations. It is not a replacement for the
// standard library — it is a thin layer on top of it that enforces conventions,
// eliminates boilerplate, and ships with a CLI to scaffold new projects and generate code.
//
// # Quick Start
//
// Install the CLI tool:
//
//	go install github.com/fvmoraes/ginger/cmd/ginger@latest
//
// Create a new project:
//
//	ginger new my-api
//	cd my-api
//	go mod tidy
//	ginger run
//
// Your API is now running at http://localhost:8080
//
// # Core Packages
//
// Ginger provides 13 production-ready packages:
//
//   - app: Application bootstrap with graceful shutdown
//   - router: HTTP routing with method helpers and groups
//   - middleware: Built-in middlewares (Logger, CORS, Recover, RequestID)
//   - errors: Typed errors with HTTP status mapping
//   - response: JSON response envelopes (OK, Created, Paginated, NoContent)
//   - config: YAML configuration with environment overrides
//   - logger: Structured logging with slog
//   - database: Database connection management
//   - health: Health check system
//   - telemetry: OpenTelemetry integration
//   - sse: Server-Sent Events for real-time streaming
//   - ws: WebSocket implementation (zero dependencies, RFC 6455)
//   - testhelper: Testing utilities
//
// # Example Usage
//
// Basic HTTP server with routing and middleware:
//
//	package main
//
//	import (
//	    "net/http"
//	    "github.com/fvmoraes/ginger/pkg/app"
//	    "github.com/fvmoraes/ginger/pkg/router"
//	    "github.com/fvmoraes/ginger/pkg/response"
//	    "github.com/fvmoraes/ginger/pkg/middleware"
//	    "github.com/fvmoraes/ginger/pkg/config"
//	)
//
//	func main() {
//	    cfg, _ := config.Load("configs/app.yaml")
//	    a := app.New(cfg)
//
//	    // Add middlewares
//	    a.Router.Use(middleware.Logger(a.Logger))
//	    a.Router.Use(middleware.Recover(a.Logger))
//	    a.Router.Use(middleware.RequestID())
//	    a.Router.Use(middleware.CORS())
//
//	    // Define routes
//	    a.Router.GET("/hello", func(w http.ResponseWriter, r *http.Request) {
//	        response.OK(w, map[string]string{"message": "Hello, World!"})
//	    })
//
//	    // Run with graceful shutdown
//	    a.Run()
//	}
//
// # CLI Commands
//
// The ginger CLI provides the following commands:
//
//   - new: Scaffold a new project with complete structure
//   - generate: Generate code (handler, service, repository, CRUD)
//   - add: Add integrations (PostgreSQL, Redis, Kafka, gRPC, etc.)
//   - run: Run development server
//   - build: Build production binary
//   - doctor: Diagnose project health
//
// Generate a complete CRUD:
//
//	ginger generate crud product
//
// This creates model, handler, service, repository, and tests.
//
// Add integrations:
//
//	ginger add postgres    # PostgreSQL
//	ginger add redis       # Redis cache
//	ginger add kafka       # Kafka messaging
//	ginger add grpc        # gRPC protocol
//
// # Design Principles
//
//   - Minimal dependencies - only what is strictly necessary
//   - Fast compilation - no magic, no reflection-heavy DI
//   - Idiomatic Go - standard interfaces, standard patterns
//   - Simple CLI - scaffold, generate, run, build
//   - Clear project structure - every team member knows where things live
//   - Developer productivity - less setup, more shipping
//
// # Documentation
//
// For complete documentation, visit:
//   - GitHub: https://github.com/fvmoraes/ginger
//   - Getting Started: https://github.com/fvmoraes/ginger/blob/main/docs/GETTING_STARTED.md
//   - Architecture: https://github.com/fvmoraes/ginger/blob/main/docs/ARCHITECTURE.md
//   - Package Reference: https://github.com/fvmoraes/ginger/blob/main/docs/PACKAGES.md
//   - Integrations: https://github.com/fvmoraes/ginger/blob/main/docs/INTEGRATIONS.md
//
// # Requirements
//
// Go 1.25+ is required (due to OpenTelemetry v1.42 dependency).
package ginger
