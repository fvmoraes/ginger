#!/usr/bin/env bash

set -euo pipefail

TYPE="patch"
CUSTOM_MESSAGE=""

usage() {
  cat <<USAGE
Usage: ./scripts/release.sh [--type patch|minor|major] [--message "English release message"]

Examples:
  ./scripts/release.sh
  ./scripts/release.sh --type minor
  ./scripts/release.sh --type patch --message "CLI improvements and reliability fixes"
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --type|-t)
      TYPE="${2:-}"
      shift 2
      ;;
    --message|-m)
      CUSTOM_MESSAGE="${2:-}"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1"
      usage
      exit 1
      ;;
  esac
done

if [[ ! "$TYPE" =~ ^(major|minor|patch)$ ]]; then
  echo "Invalid type: $TYPE"
  echo "Valid values: patch, minor, major"
  exit 1
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "This script must run inside a git repository."
  exit 1
fi

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Working tree is not clean. Commit or stash your changes before releasing."
  git status --short
  exit 1
fi

for cmd in go gh shasum; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Missing required command: $cmd"
    exit 1
  fi
done

LATEST_TAG="$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 1)"
if [[ -z "$LATEST_TAG" ]]; then
  LATEST_TAG="v0.0.0"
fi

VERSION_RAW="${LATEST_TAG#v}"
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION_RAW"

case "$TYPE" in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
esac

NEW_VERSION="$MAJOR.$MINOR.$PATCH"
NEW_TAG="v$NEW_VERSION"
RELEASE_DIR="releases/$NEW_TAG"
DATE_UTC="$(date -u +%Y-%m-%d)"

if git rev-parse "$NEW_TAG" >/dev/null 2>&1; then
  echo "Tag $NEW_TAG already exists."
  exit 1
fi

if [[ "$LATEST_TAG" == "v0.0.0" ]]; then
  COMMITS="$(git log --pretty=format:'%s')"
else
  COMMITS="$(git log --pretty=format:'%s' "$LATEST_TAG"..HEAD)"
fi

if [[ -z "$COMMITS" ]]; then
  echo "No new commits to release since $LATEST_TAG"
  exit 1
fi

if [[ -n "$CUSTOM_MESSAGE" ]]; then
  RELEASE_MESSAGE="$CUSTOM_MESSAGE"
else
  HAS_FEAT=0
  HAS_FIX=0
  if echo "$COMMITS" | rg -q '^feat(\(.+\))?:\s+'; then
    HAS_FEAT=1
  fi
  if echo "$COMMITS" | rg -q '^fix(\(.+\))?:\s+'; then
    HAS_FIX=1
  fi

  if [[ $HAS_FEAT -eq 1 && $HAS_FIX -eq 1 ]]; then
    RELEASE_MESSAGE="Feature improvements and bug fixes"
  elif [[ $HAS_FEAT -eq 1 ]]; then
    RELEASE_MESSAGE="New features and developer experience improvements"
  elif [[ $HAS_FIX -eq 1 ]]; then
    RELEASE_MESSAGE="Bug fixes and reliability improvements"
  else
    RELEASE_MESSAGE="Maintenance and internal improvements"
  fi
fi

echo "Preparing release $NEW_TAG"
echo "Latest tag: $LATEST_TAG"
echo "Release type: $TYPE"
echo "Release message: $RELEASE_MESSAGE"

echo "Updating README version badge..."
sed -E -i.bak "s/version-[0-9]+\.[0-9]+\.[0-9]+-blue/version-$NEW_VERSION-blue/g" README.md
rm -f README.md.bak

echo "Updating CHANGELOG.md..."
CHANGELOG_SECTION_TITLE="Changed"
case "$TYPE" in
  major) CHANGELOG_SECTION_TITLE="Changed" ;;
  minor) CHANGELOG_SECTION_TITLE="Added" ;;
  patch) CHANGELOG_SECTION_TITLE="Fixed" ;;
esac

{
  echo "## [$NEW_VERSION] - $DATE_UTC"
  echo
  echo "### $CHANGELOG_SECTION_TITLE"
  echo "- $RELEASE_MESSAGE"
  echo
  echo "### Commit Summary"
  echo "$COMMITS" | sed 's/^/- /'
  echo
} > /tmp/ginger_changelog_entry.txt

if rg -q '^## \[' CHANGELOG.md; then
  awk '
    BEGIN { inserted=0 }
    {
      if (!inserted && $0 ~ /^## \[/) {
        while ((getline line < "/tmp/ginger_changelog_entry.txt") > 0) print line
        inserted=1
      }
      print
    }
  ' CHANGELOG.md > CHANGELOG.md.tmp
  mv CHANGELOG.md.tmp CHANGELOG.md
else
  cat /tmp/ginger_changelog_entry.txt >> CHANGELOG.md
fi

if ! rg -q "\[$NEW_VERSION\]:" CHANGELOG.md; then
  echo "[$NEW_VERSION]: https://github.com/fvmoraes/ginger/releases/tag/$NEW_TAG" >> CHANGELOG.md
fi

rm -f /tmp/ginger_changelog_entry.txt

echo "Building binaries..."
mkdir -p "$RELEASE_DIR"

PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

for platform in "${PLATFORMS[@]}"; do
  IFS='/' read -r GOOS GOARCH <<< "$platform"
  OUTPUT="$RELEASE_DIR/ginger-$GOOS-$GOARCH"
  if [[ "$GOOS" == "windows" ]]; then
    OUTPUT+=".exe"
  fi

  GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="-s -w" -o "$OUTPUT" ./cmd/ginger
done

echo "Generating checksums..."
(
  cd "$RELEASE_DIR"
  shasum -a 256 ginger-* > checksums.txt
)

echo "Generating RELEASE_NOTES.md in English..."
{
  echo "# Ginger Framework $NEW_TAG"
  echo
  echo "$RELEASE_MESSAGE"
  echo
  echo "## Highlights"
  echo "- Release date (UTC): $DATE_UTC"
  echo "- Base tag: $LATEST_TAG"
  echo "- Total commits in this release: $(echo "$COMMITS" | wc -l | tr -d ' ')"
  echo
  echo "## Commit Summary"
  echo "$COMMITS" | sed 's/^/- /'
  echo
  echo "## Installation"
  echo '```bash'
  echo "go install github.com/fvmoraes/ginger/cmd/ginger@$NEW_TAG"
  echo '```'
  echo
  echo "## Checksums"
  echo 'See `checksums.txt` in the release assets.'
} > "$RELEASE_DIR/RELEASE_NOTES.md"

echo "Committing release files..."
git add README.md CHANGELOG.md "$RELEASE_DIR"
git commit -m "release: $NEW_TAG - $RELEASE_MESSAGE"

echo "Creating tag $NEW_TAG..."
git tag -a "$NEW_TAG" -m "$NEW_TAG - $RELEASE_MESSAGE"

echo "Pushing branch and tag..."
git push origin main
git push origin "$NEW_TAG"

echo "Publishing GitHub release with gh..."
gh release create "$NEW_TAG" "$RELEASE_DIR"/ginger-* "$RELEASE_DIR/checksums.txt" \
  --title "Ginger Framework $NEW_TAG" \
  --notes-file "$RELEASE_DIR/RELEASE_NOTES.md" \
  --latest

echo
echo "Release completed successfully: $NEW_TAG"
echo "Release URL: https://github.com/fvmoraes/ginger/releases/tag/$NEW_TAG"
