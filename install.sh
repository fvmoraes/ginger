#!/bin/bash
set -e

# Ginger Framework Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/fvmoraes/ginger/main/install.sh | bash

VERSION="${GINGER_VERSION:-}"
INSTALL_DIR="${GINGER_INSTALL_DIR:-/usr/local/bin}"
REPO="fvmoraes/ginger"
TARGET_BIN=""

resolve_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | \
            sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | \
            sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1
    fi
}

path_contains() {
    case ":$PATH:" in
        *":$1:"*) return 0 ;;
        *) return 1 ;;
    esac
}

resolve_install_dir() {
    if [ -n "${GINGER_INSTALL_DIR:-}" ]; then
        echo "$GINGER_INSTALL_DIR"
        return
    fi

    existing_bin="$(command -v ginger 2>/dev/null || true)"
    if [ -n "$existing_bin" ]; then
        dirname "$existing_bin"
        return
    fi

    if command -v go >/dev/null 2>&1; then
        gopath_bin="$(go env GOPATH)/bin"
        if path_contains "$gopath_bin"; then
            echo "$gopath_bin"
            return
        fi
    fi

    echo "$INSTALL_DIR"
}

if [ -z "$VERSION" ]; then
    VERSION="$(resolve_latest_version)"
fi

if [ -z "$VERSION" ]; then
    echo "❌ Could not resolve the latest Ginger release. Set GINGER_VERSION manually and try again."
    exit 1
fi

INSTALL_DIR="$(resolve_install_dir)"
TARGET_BIN="${INSTALL_DIR}/ginger"

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

BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
DOWNLOAD_URL="${BASE_URL}/${BINARY}"

# Temp working dir — never overwrite a user file named ./ginger in the CWD.
WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT
cd "$WORK_DIR"

echo "📦 Downloading ${BINARY}..."
if command -v curl >/dev/null 2>&1; then
    curl --proto '=https' -fsSL "$DOWNLOAD_URL" -o ginger
elif command -v wget >/dev/null 2>&1; then
    wget -q "$DOWNLOAD_URL" -O ginger
else
    echo "❌ Neither curl nor wget found. Please install one of them."
    exit 1
fi

# Verify the published checksum (GIN-011). Missing checksums.txt → explicit
# warning and continue (forks/mirrors/old releases); mismatch → hard failure.
CHECKSUM_URL="${BASE_URL}/checksums.txt"
if command -v curl >/dev/null 2>&1; then
    curl --proto '=https' -fsSL "$CHECKSUM_URL" -o checksums.txt 2>/dev/null || \
    curl --proto '=https' -fsSL "$CHECKSUM_URL" -o checksums.txt
elif command -v wget >/dev/null 2>&1; then
    wget -q "$CHECKSUM_URL" -O checksums.txt || { echo "⚠ checksums.txt not found for ${VERSION} — skipping verification"; }
fi

if [ -f checksums.txt ]; then
    EXPECTED="$(grep " ${BINARY}\$" checksums.txt | awk '{print $1}')"
    if [ -z "$EXPECTED" ]; then
        echo "⚠ ${BINARY} not listed in checksums.txt — skipping verification"
    else
        if command -v sha256sum >/dev/null 2>&1; then
            ACTUAL="$(sha256sum ginger | awk '{print $1}')"
        elif command -v shasum >/dev/null 2>&1; then
            ACTUAL="$(shasum -a 256 ginger | awk '{print $1}')"
        else
            ACTUAL=""
            echo "⚠ no sha256sum/shasum available — skipping checksum verification"
        fi
        if [ -n "$ACTUAL" ]; then
            if [ "$ACTUAL" = "$EXPECTED" ]; then
                echo "🔒 Checksum verified"
            else
                echo "❌ Checksum mismatch for ${BINARY}!"
                echo "   expected: ${EXPECTED}"
                echo "   actual:   ${ACTUAL}"
                exit 1
            fi
        fi
    fi
fi

chmod +x ginger

echo "📂 Installing to ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
    mv ginger "$TARGET_BIN"
else
    echo "🔐 Requesting sudo permissions to install to ${INSTALL_DIR}..."
    sudo mv ginger "$TARGET_BIN"
fi

echo "✅ Ginger ${VERSION} installed successfully!"

if command -v go >/dev/null 2>&1; then
    GOPATH_BIN="$(go env GOPATH)/bin"
    SHELL_RC=""

    case "$(basename "${SHELL:-}")" in
        zsh)  SHELL_RC="$HOME/.zshrc" ;;
        bash)
            if [ -f "$HOME/.bash_profile" ]; then
                SHELL_RC="$HOME/.bash_profile"
            else
                SHELL_RC="$HOME/.bashrc"
            fi
            ;;
    esac

    if [ -n "$SHELL_RC" ] && ! grep -q "$GOPATH_BIN" "$SHELL_RC" 2>/dev/null; then
        printf '\n# Added by Ginger installer\nexport PATH="$PATH:%s"\n' "$GOPATH_BIN" >> "$SHELL_RC"
        echo "  Added $GOPATH_BIN to PATH in $SHELL_RC"
        echo "  Run: source $SHELL_RC"
    fi
fi

if [ -n "${TARGET_BIN}" ]; then
    echo "  Installed binary: ${TARGET_BIN}"
fi

echo ""
echo "Quick start:"
echo "   ginger new foobar --service  # Service   → cmd/foobar"
echo "   ginger new foobar --worker   # Worker    → cmd/foobar-worker"
echo "   ginger new foobar --cli      # CLI       → cmd/foobar"
echo "   ginger new foobar            # Generic   → cmd/foobar"
echo ""
echo "Documentation: https://github.com/fvmoraes/ginger#readme"
