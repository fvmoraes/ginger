#!/bin/bash

# Ginger Framework - Development Build Script
#
# This script builds binaries for DEVELOPMENT/TESTING without creating a release.
# Use this for:
# - Testing changes locally
# - Building for specific platforms
# - Quick iterations during development
#
# For official releases, use: ./scripts/release.sh
#
# Usage: ./scripts/build.sh [platform]
#
# Platforms:
#   all (default) - Build for all platforms
#   local         - Build for current platform only
#   linux         - Build for Linux (amd64 + arm64)
#   darwin        - Build for macOS (amd64 + arm64)
#   windows       - Build for Windows (amd64)
#
# Examples:
#   ./scripts/build.sh              # Build all platforms
#   ./scripts/build.sh local        # Build for current OS only
#   ./scripts/build.sh linux        # Build Linux binaries only

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

# Configuration
PLATFORM="${1:-all}"
OUTPUT_DIR="bin"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")

log_info "Building Ginger CLI (version: $VERSION)"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build function
build_binary() {
    local goos=$1
    local goarch=$2
    local output="$OUTPUT_DIR/ginger-$goos-$goarch"
    
    if [ "$goos" = "windows" ]; then
        output="$output.exe"
    fi
    
    log_info "Building $goos/$goarch..."
    GOOS=$goos GOARCH=$goarch go build -ldflags="-s -w -X main.version=$VERSION" -o "$output" ./cmd/ginger
    
    # Make executable (except Windows)
    if [ "$goos" != "windows" ]; then
        chmod +x "$output"
    fi
    
    # Show file size
    local size=$(du -h "$output" | cut -f1)
    log_success "Built $output ($size)"
}

# Determine what to build
case "$PLATFORM" in
    all)
        log_info "Building for all platforms..."
        echo ""
        build_binary "linux" "amd64"
        build_binary "linux" "arm64"
        build_binary "darwin" "amd64"
        build_binary "darwin" "arm64"
        build_binary "windows" "amd64"
        ;;
    
    local)
        log_info "Building for current platform..."
        echo ""
        GOOS=$(go env GOOS)
        GOARCH=$(go env GOARCH)
        build_binary "$GOOS" "$GOARCH"
        
        # Create convenient symlink
        if [ "$GOOS" = "windows" ]; then
            ln -sf "ginger-$GOOS-$GOARCH.exe" "$OUTPUT_DIR/ginger.exe"
        else
            ln -sf "ginger-$GOOS-$GOARCH" "$OUTPUT_DIR/ginger"
        fi
        log_success "Symlink created: $OUTPUT_DIR/ginger"
        ;;
    
    linux)
        log_info "Building for Linux..."
        echo ""
        build_binary "linux" "amd64"
        build_binary "linux" "arm64"
        ;;
    
    darwin|macos)
        log_info "Building for macOS..."
        echo ""
        build_binary "darwin" "amd64"
        build_binary "darwin" "arm64"
        ;;
    
    windows)
        log_info "Building for Windows..."
        echo ""
        build_binary "windows" "amd64"
        ;;
    
    *)
        log_error "Unknown platform: $PLATFORM"
        echo ""
        echo "Valid platforms: all, local, linux, darwin, windows"
        exit 1
        ;;
esac

echo ""
log_success "Build complete!"
echo ""
echo "Binaries location: $OUTPUT_DIR/"
ls -lh "$OUTPUT_DIR/"
echo ""

# Show usage hint
if [ "$PLATFORM" = "local" ]; then
    echo "To install locally:"
    echo "  sudo cp $OUTPUT_DIR/ginger /usr/local/bin/"
    echo ""
    echo "Or add to PATH:"
    echo "  export PATH=\"\$PATH:$(pwd)/$OUTPUT_DIR\""
    echo ""
fi

log_info "For official releases, use: ./scripts/release.sh"
