# Ginger Framework v1.2.2

**Agilize e padronize projetos Go** | **Accelerate and standardize Go projects**

---

## 🎯 Release

Restore idiomatic cmd/ginger layout

## 🚀 Installation

### Option 1: One-line install (recommended)
```bash
curl -sSL https://raw.githubusercontent.com/fvmoraes/ginger/main/install.sh | bash
```

### Option 2: Download binary
Download from the assets below, make executable, and move to your PATH.

### Option 3: Go install
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@v1.2.2
```

Or simply:
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@latest
```

### Option 4: Build from source
```bash
git clone https://github.com/fvmoraes/ginger
cd ginger
git checkout v1.2.2
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
02132b06412914ec1d303058001287c94c761f931bc42a13097ac9140ed482b7  ginger-darwin-amd64
9bdcb610d91c1ec9385d489cc85b341d0c78b858c9504dbeb153735160ddb655  ginger-darwin-arm64
2adfdbd6931074d1c03bb6b1b90eed6cc6f27a13133923e3341beba8590ca96f  ginger-linux-amd64
a8fac1d7d1df26cc37eaf1daa36b9043295dcd71eb9d6ded7185c399eac7a365  ginger-linux-arm64
e56a4eaa2ae9dd98f0e33c8eccaa0ec710d2aa0d544c3b643d9486d4a23d94e1  ginger-windows-amd64.exe
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
