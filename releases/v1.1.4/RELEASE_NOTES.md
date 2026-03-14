# Ginger Framework v1.1.4

**Agilize e padronize projetos Go** | **Accelerate and standardize Go projects**

---

## 📚 Documentation Release

This release adds comprehensive package-level documentation for pkg.go.dev, making it easier to discover and understand the framework's API.

## ✨ Added

- **Package Documentation:** Complete doc.go with framework overview, examples, and usage patterns
- **API Reference:** Detailed documentation for all 13 core packages visible on pkg.go.dev
- **Code Examples:** Inline examples showing common usage patterns

## 🎯 Recommended Release

This is now the **recommended version** to use. It includes all features from v1.1.3 plus improved documentation.

## 🚀 Features

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
- Complete API reference on pkg.go.dev
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
go install github.com/fvmoraes/ginger/cmd/ginger@v1.1.4
```

Or simply:
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@latest
```

### Option 4: Build from source
```bash
git clone https://github.com/fvmoraes/ginger
cd ginger
git checkout v1.1.4
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
e619e4c43eebdc8a06d4238f7a0713ee5d094a8d0ea5dde1ce40bee8cbd287a7  ginger-darwin-amd64
db26e22a96b8f0f0c5d897b2265cb06db9c848c5ed34a49099ae496e75522ff6  ginger-darwin-arm64
7e715691e2eae26f9df01cf780b77625a0fd66969ab47cc9be0250c9fff8b954  ginger-linux-amd64
377c61f39e2fade01cfbd314c7acc37b8f36b8d4f7793f4fce184aef86a025e8  ginger-linux-arm64
ca4c6a78f255c5d8996163ae3e27e4f70c23a70a3bdf90af076037af629a7da4  ginger-windows-amd64.exe
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
- [pkg.go.dev API Reference](https://pkg.go.dev/github.com/fvmoraes/ginger)

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

- **v1.1.4** (current) - Added package documentation for pkg.go.dev
- **v1.1.3** - Retracted v1.1.1, recommended version
- **v1.1.2** - Fixed module path issue
- **v1.1.1** - ⚠️ RETRACTED (incorrect module path, do not use)

---

**Built with ❤️ and idiomatic Go**
