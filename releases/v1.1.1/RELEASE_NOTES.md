# Ginger Framework v1.1.1

**Agilize e padronize projetos Go** | **Accelerate and standardize Go projects**

---

## 🎉 First Official Release

This is the first official release of the Ginger Framework - a complete CLI tool and package ecosystem for building production-ready Go applications with speed and consistency.

## ✨ Highlights

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
- **Protocols:** gRPC, MCP
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

## 🚀 Quick Start

### Installation

**Option 1: Download binary**
```bash
# macOS Apple Silicon
curl -L https://github.com/fvmoraes/ginger/releases/download/v1.1.1/ginger-darwin-arm64 -o ginger
chmod +x ginger
sudo mv ginger /usr/local/bin/

# Linux AMD64
curl -L https://github.com/fvmoraes/ginger/releases/download/v1.1.1/ginger-linux-amd64 -o ginger
chmod +x ginger
sudo mv ginger /usr/local/bin/
```

**Option 2: Build from source**
```bash
git clone https://github.com/fvmoraes/ginger
cd ginger
go build -o /usr/local/bin/ginger ./cmd/ginger
```

**Option 3: Go install**
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@v1.1.1
```

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

## 📋 Requirements

- **Go 1.25+** (required by OpenTelemetry v1.42)

Check your version:
```bash
go version
```

## 📦 Binary Downloads

| Platform | Architecture | Download |
|----------|-------------|----------|
| Linux | AMD64 | [ginger-linux-amd64](https://github.com/fvmoraes/ginger/releases/download/v1.1.1/ginger-linux-amd64) |
| Linux | ARM64 | [ginger-linux-arm64](https://github.com/fvmoraes/ginger/releases/download/v1.1.1/ginger-linux-arm64) |
| macOS | Intel | [ginger-darwin-amd64](https://github.com/fvmoraes/ginger/releases/download/v1.1.1/ginger-darwin-amd64) |
| macOS | Apple Silicon | [ginger-darwin-arm64](https://github.com/fvmoraes/ginger/releases/download/v1.1.1/ginger-darwin-arm64) |
| Windows | AMD64 | [ginger-windows-amd64.exe](https://github.com/fvmoraes/ginger/releases/download/v1.1.1/ginger-windows-amd64.exe) |

**Verify downloads:** [checksums.txt](https://github.com/fvmoraes/ginger/releases/download/v1.1.1/checksums.txt)

## 🔐 Checksums (SHA256)

```
fe093ccbe371781e3a712268f9fadf7f015c4ca1a5942d98a0adf89d356ea167  ginger-darwin-amd64
cea6270b47c164afb2c372e07775fc6c1b23c34c1976fa653153cd1a0e02dbaf  ginger-darwin-arm64
dd931f1ed12f018cff77c5463fa9affaf040b542b267f72a14d101835c39e953  ginger-linux-amd64
addc7c2a94fa04c6be9e89a5347d064941b1e2033f1f0a9f1ac327db3185803f  ginger-linux-arm64
aac4b652d0a9045a3ed69a8ab67950228f28b898203b210c36aeb2d6d6286eff  ginger-windows-amd64.exe
```

## 📖 Documentation

- [README](https://github.com/fvmoraes/ginger#readme)
- [Getting Started](https://github.com/fvmoraes/ginger/blob/main/docs/GETTING_STARTED.md)
- [Architecture](https://github.com/fvmoraes/ginger/blob/main/docs/ARCHITECTURE.md)
- [Package Reference](https://github.com/fvmoraes/ginger/blob/main/docs/PACKAGES.md)
- [Integrations](https://github.com/fvmoraes/ginger/blob/main/docs/INTEGRATIONS.md)
- [Testing Guide](https://github.com/fvmoraes/ginger/blob/main/docs/TESTING.md)
- [Deployment](https://github.com/fvmoraes/ginger/blob/main/docs/DEPLOYMENT.md)
- [Copy-Paste Examples](https://github.com/fvmoraes/ginger/blob/main/docs/COPY_PASTE.md)

## 🎯 Design Principles

- **Minimal dependencies** - only what's strictly necessary
- **Fast compilation** - no magic, no reflection-heavy DI
- **Idiomatic Go** - standard interfaces, standard patterns
- **Developer productivity** - less setup, more shipping
- **Clear structure** - every team member knows where things live

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📝 License

MIT License - see [LICENSE](https://github.com/fvmoraes/ginger/blob/main/LICENSE) file for details.

## 💬 Support

- **Issues:** [GitHub Issues](https://github.com/fvmoraes/ginger/issues)
- **Discussions:** [GitHub Discussions](https://github.com/fvmoraes/ginger/discussions)
- **Email:** fvmoraes@gmail.com

---

**Built with ❤️ and idiomatic Go**
