# Ginger Framework v1.2.1

**Agilize e padronize projetos Go** | **Accelerate and standardize Go projects**

---

## 🎯 Release

Move main to root — go install github.com/fvmoraes/ginger@latest

## 🚀 Installation

### Option 1: One-line install (recommended)
```bash
curl -sSL https://raw.githubusercontent.com/fvmoraes/ginger/main/install.sh | bash
```

### Option 2: Download binary
Download from the assets below, make executable, and move to your PATH.

### Option 3: Go install
```bash
go install github.com/fvmoraes/ginger@v1.2.1
```

Or simply:
```bash
go install github.com/fvmoraes/ginger@latest
```

### Option 4: Build from source
```bash
git clone https://github.com/fvmoraes/ginger
cd ginger
git checkout v1.2.1
go build -o /usr/local/bin/ginger .
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
76d78afff76093217550f1a1dbc7b6a90ab2c568f6f9eaed6cb821520a5dbf6b  ginger-darwin-amd64
15f1178587318550e34b3dc9eb5f32585234556bfff0363de0c8c5092e9c43d5  ginger-darwin-arm64
9c7d741d9d0cf186bac0386a030a58862b6d16c3c6677f370224f8623be58d7b  ginger-linux-amd64
1383eb151eb997b904d82e72cc30bc22d64ad3c39b0de1701d6498af6ffeba3d  ginger-linux-arm64
b59bf0c92eb0a5cc5380207000274f3053b393f4241c47e50445c4a859107fea  ginger-windows-amd64.exe
```

## 📋 Requirements

- **Go 1.25+** (required by OpenTelemetry v1.42)

## 🚀 Quick Start

```bash
# Create project
ginger new my-api
cd my-api
go mod tidy

# Run development server
ginger run
```

Your API is now running at `http://localhost:8080`

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

## 💬 Support

- **Issues:** [GitHub Issues](https://github.com/fvmoraes/ginger/issues)
- **Discussions:** [GitHub Discussions](https://github.com/fvmoraes/ginger/discussions)
- **Email:** fvmoraes@gmail.com

---

**Built with ❤️ and idiomatic Go**
