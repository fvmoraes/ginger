#!/bin/bash
set -e

# Ginger Framework Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/fvmoraes/ginger/main/install.sh | bash

VERSION="${GINGER_VERSION:-v1.1.1}"
INSTALL_DIR="${GINGER_INSTALL_DIR:-/usr/local/bin}"

echo "🌶️  Installing Ginger Framework ${VERSION}..."

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux*)
        OS="linux"
        ;;
    Darwin*)
        OS="darwin"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        OS="windows"
        ;;
    *)
        echo "❌ Unsupported operating system: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

BINARY="ginger-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY="${BINARY}.exe"
fi

DOWNLOAD_URL="https://github.com/fvmoraes/ginger/releases/download/${VERSION}/${BINARY}"

echo "📦 Downloading ${BINARY}..."
if command -v curl >/dev/null 2>&1; then
    curl -sSL "$DOWNLOAD_URL" -o ginger
elif command -v wget >/dev/null 2>&1; then
    wget -q "$DOWNLOAD_URL" -O ginger
else
    echo "❌ Neither curl nor wget found. Please install one of them."
    exit 1
fi

chmod +x ginger

echo "📂 Installing to ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
    mv ginger "$INSTALL_DIR/ginger"
else
    echo "🔐 Requesting sudo permissions to install to ${INSTALL_DIR}..."
    sudo mv ginger "$INSTALL_DIR/ginger"
fi

echo "✅ Ginger ${VERSION} installed successfully!"
echo ""
echo "🚀 Quick start:"
echo "   ginger new my-api"
echo "   cd my-api"
echo "   ginger run"
echo ""
echo "📚 Documentation: https://github.com/fvmoraes/ginger#readme"
