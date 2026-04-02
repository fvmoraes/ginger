#!/bin/bash

# Ginger Framework - Release Automation Script
# 
# This script creates OFFICIAL RELEASES with:
# - Version bump in all files
# - CHANGELOG update
# - Binaries for all platforms
# - Git tag and push
# - Release notes
#
# Usage: ./scripts/release.sh <version> <type> [message]
# 
# Types:
#   major - Breaking changes (1.0.0 -> 2.0.0)
#   minor - New features (1.1.0 -> 1.2.0)
#   patch - Bug fixes (1.1.1 -> 1.1.2)
#
# Examples:
#   ./scripts/release.sh 1.2.0 minor "Add WebSocket support"
#   ./scripts/release.sh 1.1.5 patch "Fix CORS middleware"
#   ./scripts/release.sh 2.0.0 major "Complete rewrite"
#
# For development builds without release, use: ./scripts/build.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check arguments
if [ -z "$1" ] || [ -z "$2" ]; then
    log_error "Version and type are required"
    echo ""
    echo "Usage: ./scripts/release.sh <version> <type> [message]"
    echo ""
    echo "Types:"
    echo "  major - Breaking changes (1.0.0 -> 2.0.0)"
    echo "  minor - New features (1.1.0 -> 1.2.0)"
    echo "  patch - Bug fixes (1.1.1 -> 1.1.2)"
    echo ""
    echo "Examples:"
    echo "  ./scripts/release.sh 1.2.0 minor \"Add WebSocket support\""
    echo "  ./scripts/release.sh 1.1.5 patch \"Fix CORS middleware\""
    echo ""
    exit 1
fi

VERSION="$1"
TYPE="$2"
MESSAGE="${3:-Release v$VERSION}"
TAG="v$VERSION"
RELEASE_DIR="releases/$TAG"

# Validate type
if [[ ! "$TYPE" =~ ^(major|minor|patch)$ ]]; then
    log_error "Invalid type: $TYPE"
    echo "Valid types: major, minor, patch"
    exit 1
fi

log_info "Starting OFFICIAL RELEASE for version $VERSION ($TYPE)"
echo ""
log_warning "This will:"
echo "  • Update version in README.md and CHANGELOG.md"
echo "  • Build binaries for 5 platforms"
echo "  • Create git tag $TAG"
echo "  • Push to GitHub"
echo ""
read -p "Continue? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_error "Release cancelled"
    exit 1
fi
echo ""

# Step 1: Check if working directory is clean
log_info "Step 1/10: Checking working directory..."
if [ -n "$(git status --porcelain)" ]; then
    log_warning "Working directory has uncommitted changes"
    git status --short
    echo ""
    read -p "Do you want to continue? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_error "Release cancelled"
        exit 1
    fi
fi
log_success "Working directory checked"
echo ""

# Step 2: Update version in files
log_info "Step 2/10: Updating version in files..."

# Update README.md badge
if [ -f "README.md" ]; then
    sed -i.bak "s/version-[0-9]\+\.[0-9]\+\.[0-9]\+-blue/version-$VERSION-blue/g" README.md
    rm -f README.md.bak
    log_success "Updated README.md"
fi

# Update CHANGELOG.md
if [ -f "CHANGELOG.md" ]; then
    DATE=$(date +%Y-%m-%d)
    
    # Determine section based on type
    case "$TYPE" in
        major)
            SECTION="### Changed\n- $MESSAGE\n\n### ⚠️ Breaking Changes\n- See migration guide"
            ;;
        minor)
            SECTION="### Added\n- $MESSAGE"
            ;;
        patch)
            SECTION="### Fixed\n- $MESSAGE"
            ;;
    esac
    
    # Add new version entry at the top (after the header)
    awk -v version="$VERSION" -v date="$DATE" -v section="$SECTION" '
        /^## \[/ && !done {
            print ""
            print "## [" version "] - " date
            print ""
            print section
            print ""
            done=1
        }
        {print}
    ' CHANGELOG.md > CHANGELOG.md.tmp
    mv CHANGELOG.md.tmp CHANGELOG.md
    
    # Add badge link at the bottom if not exists
    if ! grep -q "\[$VERSION\]:" CHANGELOG.md; then
        echo "" >> CHANGELOG.md
        echo "[$VERSION]: https://github.com/fvmoraes/ginger/releases/tag/$TAG" >> CHANGELOG.md
    fi
    log_success "Updated CHANGELOG.md"
fi

log_success "Version updated in files"
echo ""

# Step 3: Create release directory
log_info "Step 3/10: Creating release directory..."
mkdir -p "$RELEASE_DIR"
log_success "Created $RELEASE_DIR"
echo ""

# Step 4: Build binaries
log_info "Step 4/10: Building binaries..."

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a parts <<< "$platform"
    GOOS="${parts[0]}"
    GOARCH="${parts[1]}"
    
    OUTPUT="$RELEASE_DIR/ginger-$GOOS-$GOARCH"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="$OUTPUT.exe"
    fi
    
    log_info "Building $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$OUTPUT" .
    log_success "Built $OUTPUT"
done

log_success "All binaries built"
echo ""

# Step 5: Generate checksums
log_info "Step 5/10: Generating checksums..."
cd "$RELEASE_DIR"
shasum -a 256 ginger-* | sed 's|releases/[^/]*/||g' > checksums.txt
cd - > /dev/null
log_success "Checksums generated"
echo ""

# Step 6: Create release notes
log_info "Step 6/10: Creating release notes..."

cat > "$RELEASE_DIR/RELEASE_NOTES.md" << EOF
# Ginger Framework $TAG

**Agilize e padronize projetos Go** | **Accelerate and standardize Go projects**

---

## 🎯 Release

$MESSAGE

## 🚀 Installation

### Option 1: One-line install (recommended)
\`\`\`bash
curl -sSL https://raw.githubusercontent.com/fvmoraes/ginger/main/install.sh | bash
\`\`\`

### Option 2: Download binary
Download from the assets below, make executable, and move to your PATH.

### Option 3: Go install
\`\`\`bash
go install github.com/fvmoraes/ginger@$TAG
\`\`\`

Or simply:
\`\`\`bash
go install github.com/fvmoraes/ginger@latest
\`\`\`

### Option 4: Build from source
\`\`\`bash
git clone https://github.com/fvmoraes/ginger
cd ginger
git checkout $TAG
go build -o /usr/local/bin/ginger .
\`\`\`

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

\`\`\`
$(cat "$RELEASE_DIR/checksums.txt")
\`\`\`

## 📋 Requirements

- **Go 1.25+** (required by OpenTelemetry v1.42)

## 🚀 Quick Start

\`\`\`bash
# Create project
ginger new my-api
cd my-api
go mod tidy

# Run development server
ginger run
\`\`\`

Your API is now running at \`http://localhost:8080\`

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
EOF

log_success "Release notes created"
echo ""

# Step 7: Commit changes
log_info "Step 7/10: Committing changes..."
git add README.md CHANGELOG.md "$RELEASE_DIR/"
git commit -m "release: $TAG - $MESSAGE"
log_success "Changes committed"
echo ""

# Step 8: Create and push tag
log_info "Step 8/10: Creating git tag..."
git tag -a "$TAG" -m "$TAG - $MESSAGE"
log_success "Tag $TAG created"
echo ""

# Step 9: Push to GitHub
log_info "Step 9/10: Pushing to GitHub..."
git push origin main
git push origin "$TAG"
log_success "Pushed to GitHub"
echo ""

# Step 10: Summary
log_info "Step 10/10: Release summary"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log_success "Release $TAG created successfully!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📦 Binaries location: $RELEASE_DIR/"
echo "📝 Release notes: $RELEASE_DIR/RELEASE_NOTES.md"
echo "🔐 Checksums: $RELEASE_DIR/checksums.txt"
echo ""
echo "Next steps:"
echo "1. Go to: https://github.com/fvmoraes/ginger/releases/new?tag=$TAG"
echo "2. Copy content from: $RELEASE_DIR/RELEASE_NOTES.md"
echo "3. Upload binaries from: $RELEASE_DIR/"
echo "4. Mark as latest release"
echo "5. Publish release"
echo ""
echo "Or use GitHub CLI:"
echo "  gh release create $TAG $RELEASE_DIR/ginger-* $RELEASE_DIR/checksums.txt \\"
echo "    --title \"Ginger Framework $TAG\" \\"
echo "    --notes-file $RELEASE_DIR/RELEASE_NOTES.md \\"
echo "    --latest"
echo ""
log_success "Done! 🎉"
