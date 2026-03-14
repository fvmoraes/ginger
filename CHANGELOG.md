# Changelog

All notable changes to the Ginger Framework will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.3] - 2026-03-14

### Changed
- Added retract directive for v1.1.1 (incorrect module path)
- v1.1.1 is now marked as invalid and should not be used

## [1.1.2] - 2026-03-14

### Fixed
- Correct module path in go.mod for pkg.go.dev indexing
- Go proxy cache issue resolved

### Added
- pkg.go.dev badge in README

## [1.1.1] - 2026-03-14

### Added

#### CLI Tool
- `ginger new` - Scaffold new projects with complete structure
- `ginger generate` - Code generators (handler, service, repository, CRUD)
- `ginger add` - Integration system with 20+ integrations
- `ginger run` - Development server
- `ginger build` - Production builds
- `ginger doctor` - Project health diagnostics

#### Core Packages
- `pkg/app` - Application bootstrap with graceful shutdown
- `pkg/router` - HTTP routing with method helpers and groups
- `pkg/middleware` - Built-in middlewares (Logger, CORS, Recover, RequestID)
- `pkg/errors` - Typed errors with HTTP status mapping
- `pkg/config` - YAML configuration with environment overrides
- `pkg/logger` - Structured logging with slog
- `pkg/database` - Database connection management
- `pkg/health` - Health check system
- `pkg/telemetry` - OpenTelemetry integration
- `pkg/testhelper` - Testing utilities

#### UI Facilitators
- `pkg/response` - JSON response envelopes (OK, Created, Paginated, NoContent)
- `pkg/sse` - Server-Sent Events for real-time streaming
- `pkg/ws` - WebSocket implementation (zero dependencies, RFC 6455)

#### Integrations
- **Databases:** PostgreSQL, MySQL, SQLite, SQL Server, ClickHouse
- **NoSQL:** MongoDB, Couchbase
- **Cache:** Redis
- **Messaging:** Kafka, RabbitMQ, NATS, Google Pub/Sub
- **Protocols:** gRPC, MCP (Model Context Protocol)
- **Real-time:** SSE, WebSocket
- **Observability:** OpenTelemetry, Prometheus

#### Documentation
- Comprehensive bilingual documentation (English/Portuguese)
- Getting Started guide (5-minute tutorial)
- Copy-Paste ready code examples
- Architecture deep dive
- Complete package reference
- Integrations guide
- Testing guide
- Deployment guide (Docker, Kubernetes, Helm)
- Quick reference
- Quality checklist

#### Infrastructure
- Docker multi-stage builds
- Docker Compose templates
- Kubernetes manifests
- Helm chart templates
- Example application with full CRUD

### Technical Details
- Go 1.25+ required (OpenTelemetry v1.42 dependency)
- Zero external dependencies for core functionality
- Stdlib-first approach
- Fast compilation
- Idiomatic Go patterns

---

## Release Notes

### Installation

```bash
# Via git clone
git clone https://github.com/fvmoraes/ginger
cd ginger
go build -o /usr/local/bin/ginger ./cmd/ginger

# Via go install
go install github.com/fvmoraes/ginger/cmd/ginger@v1.1.1
```

### Quick Start

```bash
# Create new project
ginger new my-api
cd my-api
go mod tidy

# Run development server
ginger run
```

### Binary Downloads

Download pre-compiled binaries from the [releases page](https://github.com/fvmoraes/ginger/releases/tag/v1.1.1):

- **Linux AMD64:** `ginger-linux-amd64`
- **Linux ARM64:** `ginger-linux-arm64`
- **macOS Intel:** `ginger-darwin-amd64`
- **macOS Apple Silicon:** `ginger-darwin-arm64`
- **Windows AMD64:** `ginger-windows-amd64.exe`

Verify downloads with `checksums.txt`.

---

[1.1.1]: https://github.com/fvmoraes/ginger/releases/tag/v1.1.1
