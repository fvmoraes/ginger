# Ginger Framework v1.4.0

Project awareness, safe generation with --plan/--force, managed regions, ginger init/inspect commands, and .ginger/ manifest support.

## Highlights
- Release date (UTC): 2026-06-29
- Base tag: v1.3.6
- Total commits in this release: 1

## What's New
- **ginger init**: Initialize ginger.yaml in existing Go projects with auto-detected structure
- **ginger inspect**: Analyze project structure, type, and detected features
- **--plan / --force flags**: Preview changes before applying; overwrite only with explicit --force
- **Managed regions**: Safe code injection via // ginger:begin / // ginger:end blocks
- **.ginger/ manifest**: Track owned files in .ginger/manifest.yaml
- **Root detection**: Commands work from any subdirectory (walks up to find project root)
- **No-overwrite by default**: Generation never clobbers user files without --force

## Internal packages added
- internal/project — root discovery (ginger.yaml > go.mod > .git)
- internal/plan — safe plan/apply system
- internal/region — managed code region parser
- internal/capability — modular capability registry

## Commit Summary
- feat: add project management and region handling features

## Installation
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@v1.4.0
```

## Checksums
See `checksums.txt` in the release assets.
