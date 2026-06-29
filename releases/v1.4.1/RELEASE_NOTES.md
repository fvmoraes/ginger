# Ginger Framework v1.4.1

Documentation validation, consolidation, and cleanup release.

## Highlights
- Release date (UTC): 2026-06-29
- Base tag: v1.4.0

## What's New
- **Documentation consolidation**: Removed 4 redundant docs (COPY_PASTE, SUMMARY, SAFE_EVOLUTION, docs/CHANGELOG)
- **CLI Reference updated**: EN and PT now list all 11 commands (init, inspect, docs, tests --scan)
- **Integration tables fixed**: Added ORMs (gorm, sqlx, bun), corrected mongodb and pubsub package paths
- **Safe generation flags**: --plan/--force documented across all relevant sections
- **Project safety positioning**: README updated to reflect "safe project framework" identity
- **Docker/CI tags updated**: alpine 3.19→3.21, actions v3→v4, buildx v2→v3
- **Auth + File Upload patterns**: Added to Quick Reference
- **Portuguese fixes**: Removed obsolete "api" mode reference, added missing commands

## Internal
- plan/ directory committed with evolution roadmap and doc validation plans
- .gitignore updated to unblock plan/ from tracking

## Installation
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@v1.4.1
```

## Checksums
See `checksums.txt` in the release assets.
