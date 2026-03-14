# Ginger Framework v1.1.2

**Agilize e padronize projetos Go** | **Accelerate and standardize Go projects**

---

## 🔧 Bug Fix Release

This release fixes the module path issue that prevented v1.1.1 from being properly indexed on pkg.go.dev.

## 🐛 Fixed

- **Module Path:** Corrected `go.mod` module path for proper Go proxy indexing
- **pkg.go.dev:** Module now properly indexed and discoverable

## ✨ Features (Same as v1.1.1)

### 🛠️ Complete CLI Tool
- Scaffold new projects with `ginger new`
- Generate code with `ginger generate` (handlers, services, repositories, full CRUD)
- Add integrations with `ginger add` (20+ integrations available)
- Run and build with `ginger run` and `ginger build`
- Diagnose projects with `ginger doctor`

### 📦 13 Core Packages
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
- **Databases:** PostgreSQL, MySQL, SQLite, SQL Server, ClickHouse
- **NoSQL:** MongoDB, Couchbase
- **Cache:** Redis
- **Messaging:** Kafka, RabbitMQ, NATS, Google Pub/Sub
- **Protocols:** gRPC, MCP
- **Real-time:** SSE, WebSocket
- **Observability:** OpenTelemetry, Prometheus

## 🚀 Installation

### Option 1: One-line install
```bash
curl -sSL https://raw.githubusercontent.com/fvmoraes/ginger/main/install.sh | bash
```

### Option 2: Download binary
Download from the assets below, make executable, and move to your PATH.

### Option 3: Go install
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@v1.1.2
```

### Option 4: Build from source
```bash
git clone https://github.com/fvmoraes/ginger
cd ginger
git checkout v1.1.2
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
f4f7bfbdc12a841a815e9ef3ce1dff9c2750de7179b62f510c68dc7911dfffbf  ginger-darwin-amd64
b7f6499244adda23dd73817255f90300d5d02be983d45ec86ed46b7e63b055a6  ginger-darwin-arm64
cfe9f14474fc363b7a35eb0f135b948e081989672ab2a173d38d64a1b9e725b8  ginger-linux-amd64
9b945a3efafcc65da6be2523ebf82dad8cdeb2db08a52d21b16d4eaec7e4c3d8  ginger-linux-arm64
541dfaa624d32e73cfb8872118831bc5131d8391d22f48267da135eb23644c59  ginger-windows-amd64.exe
```

## 📋 Requirements

- **Go 1.25+** (required by OpenTelemetry v1.42)

## 🚀 Quick Start

```bash
# Create new project
ginger new my-api
cd my-api
go mod tidy

# Run development server
ginger run
```

Your API is now running at `http://localhost:8080`

## 📖 Documentation

- [README](https://github.com/fvmoraes/ginger#readme)
- [Getting Started](https://github.com/fvmoraes/ginger/blob/main/docs/GETTING_STARTED.md)
- [Architecture](https://github.com/fvmoraes/ginger/blob/main/docs/ARCHITECTURE.md)
- [Package Reference](https://github.com/fvmoraes/ginger/blob/main/docs/PACKAGES.md)
- [pkg.go.dev](https://pkg.go.dev/github.com/fvmoraes/ginger)

## 💬 Support

- **Issues:** [GitHub Issues](https://github.com/fvmoraes/ginger/issues)
- **Discussions:** [GitHub Discussions](https://github.com/fvmoraes/ginger/discussions)
- **Email:** fvmoraes@gmail.com

---

**Built with ❤️ and idiomatic Go**
