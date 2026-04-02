# Release Script Usage

The `scripts/release.sh` script automates the full release flow:

- Detects the latest semantic tag (`vX.Y.Z`)
- Calculates the next version (`patch` by default)
- Generates an English release message
- Updates `README.md` version badge and `CHANGELOG.md`
- Builds binaries and checksums
- Creates `RELEASE_NOTES.md` in English
- Commits, tags, pushes, and publishes the GitHub release

## Usage

```bash
./scripts/release.sh [--type patch|minor|major] [--message "Release message"]
```

## Examples

```bash
./scripts/release.sh
./scripts/release.sh --type minor
./scripts/release.sh --type patch --message "CLI improvements and reliability fixes"
```
