# Ginger Framework v1.2.3

**Agilize e padronize projetos Go** | **Accelerate and standardize Go projects**

---

## 🎯 Release

Ajustes no CLI e build dinâmico para cmd/<nome>

## 🚀 Installation

### Option 1: One-line install (recommended)
```bash
curl -sSL https://raw.githubusercontent.com/fvmoraes/ginger/main/install.sh | bash
```

### Option 2: Download binary
Download from the assets below, make executable, and move to your PATH.

### Option 3: Go install
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@v1.2.3
```

Or simply:
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@latest
```

### Option 4: Build from source
```bash
git clone https://github.com/fvmoraes/ginger
cd ginger
git checkout v1.2.3
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
01ae5f89bd099ace3cddeabfd06e9aba1c118ddbfd61091e576e34136430e904  ginger-darwin-amd64
492ad506ae34de8f5175459cf5dc133da170475eebab180cb008fde52a7f70d2  ginger-darwin-arm64
59b726f7ff91acde7611f9eae56e2a360dfa5f155efa402a1b2c60849e3986f7  ginger-linux-amd64
96c9f053cc19316d9d8d11f15293ac700ac86641bd4147675e10c6fb7054bbc0  ginger-linux-arm64
c3ab563c8b5197bd9a9722ee70a0de24f16a2537f398d4253c9589b4e26863e5  ginger-windows-amd64.exe
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

---

**Built with ❤️ and idiomatic Go**
