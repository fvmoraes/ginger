# Ginger Framework v1.1.3

**Agilize e padronize projetos Go** | **Accelerate and standardize Go projects**

---

## 🎯 Recommended Release

This is the **recommended version** to use. It includes a retraction of v1.1.1 which had an incorrect module path.

## 🔄 Changed

- **Retracted v1.1.1:** Added `retract` directive in `go.mod` to mark v1.1.1 as invalid
- **Version Management:** `go get @latest` now correctly skips v1.1.1

## ⚠️ Important Note

**Do not use v1.1.1** - it has an incorrect module path and will not work properly. Use v1.1.2 or v1.1.3 instead.

## ✨ Features

### 🛠️ Complete CLI Tool
- Scaffold new projects with `ginger new`
- Generate code with `ginger generate` (handlers, services, repositories, full CRUD)
- Add integrations with `ginger add` (20+ integrations available)
- Run and build with `ginger run` and `ginger build`
- Diagnose projects with `ginger doctor`

### 📦 13 Core Packages
Production-ready packages for common needs:
- Application bootstrap with graceful shutdown
- HTTP routing with method helpers
- Built-in middlewares (Logger, CORS, Recover, RequestID)
- Typed errors with HTTP status mapping
- JSON response envelopes
- Server-Sent Events (SSE)
- WebSocket (zero dependencies)
- Configuration management
- Structured logging
- Database connections
- Health checks
- OpenTelemetry integration
- Testing utilities

### 🔌 20+ Integrations
One command to add:
- **Databases:** PostgreSQL, MySQL, SQLite, SQL Server, ClickHouse
- **NoSQL:** MongoDB, Couchbase
- **Cache:** Redis
- **Messaging:** Kafka, RabbitMQ, NATS, Google Pub/Sub
- **Protocols:** gRPC, MCP (Model Context Protocol)
- **Real-time:** SSE, WebSocket
- **Observability:** OpenTelemetry, Prometheus

### 📚 Comprehensive Documentation
- Bilingual (English/Portuguese)
- 5-minute getting started guide
- Copy-paste ready examples
- Architecture deep dive
- Complete API reference
- Testing strategies
- Deployment guides (Docker, K8s, Helm)

## 🚀 Installation

### Option 1: One-line install (recommended)
```bash
curl -sSL https://raw.githubusercontent.com/fvmoraes/ginger/main/install.sh | bash
```

### Option 2: Download binary
Download from the assets below, make executable, and move to your PATH.

### Option 3: Go install
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@v1.1.3
```

Or simply:
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@latest
```

### Option 4: Build from source
```bash
git clone https://github.com/fvmoraes/ginger
cd ginger
git checkout v1.1.3
go build -o /usr/local/bin/ginger ./cmd/ginger
```

## 📦 Binary Downloads

| Platform | Architecture | Download |
|----------|-------------|----------|
| Linux | AMD64 | ginger-linux-amd64 |
| Linux | ARM64 | ginger-linux-arm64 |
| macOS | Intel | ginger-darwin-amd64 |
| macOS | Apple Silicon | ginger-darwin-arm64 |
| Windows | AMD64 | ginger-windows-amd64.exe |

**Verify downloads:** checksums.txt

## 🔐 Checksums (SHA256)

```
addc493749c0abe4ffbb84432f9c2db9371b501490f4a77768ee43c79558bba1  ginger-darwin-amd64
abdc0b50f99b65916e9f39e8d002221f2e2ffcaf2d20dd8555c04932d8023fcc  ginger-darwin-arm64
e9494692e633795f1489da1933cabcdc3ba3dbdf69810ebfc62444aa07beb0ca  ginger-linux-amd64
a96a6615e6be7bdebc9b49764c3d54a5f9265a13c12f56ec163c6a5e07a2ed70  ginger-linux-arm64
254d8d185937ab112e00bd64cef92478846b97f47d6c41f694663219eebeb8aa  ginger-windows-amd64.exe
```

## 📋 Requirements

- **Go 1.25+** (required by OpenTelemetry v1.42)

Check your version:
```bash
go version
```

## 🚀 Quick Start

### Create Your First Project

```bash
# Create project
ginger new my-api
cd my-api
go mod tidy

# Run development server
ginger run
```

Your API is now running at `http://localhost:8080`

### Generate a Complete CRUD

```bash
ginger generate crud product
```

This creates:
- Model with validation
- Handler with all HTTP methods
- Service with business logic
- Repository with data access
- Tests for all layers

### Add Integrations

```bash
ginger add postgres    # PostgreSQL
ginger add redis       # Redis cache
ginger add kafka       # Kafka messaging
ginger add grpc        # gRPC protocol
```

## 📖 Documentation

- [README](https://github.com/fvmoraes/ginger#readme)
- [Getting Started (5 min)](https://github.com/fvmoraes/ginger/blob/main/docs/GETTING_STARTED.md)
- [Copy-Paste Examples](https://github.com/fvmoraes/ginger/blob/main/docs/COPY_PASTE.md)
- [Architecture](https://github.com/fvmoraes/ginger/blob/main/docs/ARCHITECTURE.md)
- [Package Reference](https://github.com/fvmoraes/ginger/blob/main/docs/PACKAGES.md)
- [Integrations](https://github.com/fvmoraes/ginger/blob/main/docs/INTEGRATIONS.md)
- [Testing](https://github.com/fvmoraes/ginger/blob/main/docs/TESTING.md)
- [Deployment](https://github.com/fvmoraes/ginger/blob/main/docs/DEPLOYMENT.md)
- [pkg.go.dev](https://pkg.go.dev/github.com/fvmoraes/ginger)

## 🎯 Design Principles

- **Minimal dependencies** — only what is strictly necessary
- **Fast compilation** — no magic, no reflection-heavy DI
- **Idiomatic Go** — standard interfaces, standard patterns
- **Simple CLI** — scaffold, generate, run, build
- **Clear project structure** — every team member knows where things live
- **Developer productivity** — less setup, more shipping

## 💬 Support

- **Issues:** [GitHub Issues](https://github.com/fvmoraes/ginger/issues)
- **Discussions:** [GitHub Discussions](https://github.com/fvmoraes/ginger/discussions)
- **Email:** fvmoraes@gmail.com

---

## 📝 Version History

- **v1.1.3** (current) - Retracted v1.1.1, recommended version
- **v1.1.2** - Fixed module path issue
- **v1.1.1** - ⚠️ RETRACTED (incorrect module path, do not use)

---

**Built with ❤️ and idiomatic Go**
