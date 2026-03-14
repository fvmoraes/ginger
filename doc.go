// Package ginger provides a lightweight, opinionated Go framework for building web applications and APIs.
//
// Ginger accelerates and standardizes Go project development with a complete CLI tool,
// 13 production-ready packages, and 20+ integrations.
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
// # Core Packages
//
// Ginger provides the following core packages:
//
//   - app: Application bootstrap with graceful shutdown
//   - router: HTTP routing with method helpers and groups
//   - middleware: Built-in middlewares (Logger, CORS, Recover, RequestID)
//   - errors: Typed errors with HTTP status mapping
//   - response: JSON response envelopes
//   - config: YAML configuration with environment overrides
//   - logger: Structured logging with slog
//   - database: Database connection management
//   - health: Health check system
//   - telemetry: OpenTelemetry integration
//   - sse: Server-Sent Events for real-time streaming
//   - ws: WebSocket implementation (zero dependencies)
//   - testhelper: Testing utilities
//
// # Example Usage
//
// Basic HTTP server:
//
//	package main
//
//	import (
//	    "github.com/fvmoraes/ginger/pkg/app"
//	    "github.com/fvmoraes/ginger/pkg/router"
//	    "github.com/fvmoraes/ginger/pkg/response"
//	)
//
//	func main() {
//	    a := app.New()
//	    r := router.New()
//
//	    r.GET("/hello", func(w http.ResponseWriter, r *http.Request) {
//	        response.OK(w, map[string]string{"message": "Hello, World!"})
//	    })
//
//	    a.Run(":8080", r)
//	}
//
// # CLI Commands
//
// The ginger CLI provides the following commands:
//
//   - new: Scaffold a new project with complete structure
//   - generate: Generate code (handler, service, repository, CRUD)
//   - add: Add integrations (PostgreSQL, Redis, Kafka, etc.)
//   - run: Run development server with hot reload
//   - build: Build production binary
//   - doctor: Diagnose project health
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
//
// # Requirements
//
// Go 1.25+ is required (due to OpenTelemetry v1.42 dependency).
package ginger
