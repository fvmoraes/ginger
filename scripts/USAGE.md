# Scripts Usage

## `release.sh`

The `scripts/release.sh` script automates the full release flow:

- Detects the latest semantic tag (`vX.Y.Z`)
- Calculates the next version (`patch` by default)
- Generates an English release message
- Updates `README.md` version badge and `CHANGELOG.md`
- Builds binaries and checksums
- Creates `RELEASE_NOTES.md` in English
- Commits, tags, pushes, and publishes the GitHub release

### Usage

```bash
./scripts/release.sh [--type patch|minor|major] [--message "Release message"]
```

### Examples

```bash
./scripts/release.sh
./scripts/release.sh --type minor
./scripts/release.sh --type patch --message "CLI improvements and reliability fixes"
```

---

## `test-ginger-massive.sh`

The `scripts/test-ginger-massive.sh` script performs a full end-to-end validation of Ginger in an isolated workspace:

- Copies the local working tree into a fresh workspace when `REPO_URL` is a local path
- Clones Ginger into a fresh workspace when `REPO_URL` is a remote repository URL
- Installs the CLI into a private `bin/`
- Exports `PATH` during the run
- Validates `ginger version` and `ginger help`
- Exercises `ginger new` for `generic`, `service`, `worker`, and `cli`
- Exercises all supported generators
- Exercises all supported `ginger add` integrations
- Runs `go test`, `go build`, `ginger build`, `ginger doctor`, and runtime checks
- Streams colorful output in real time
- Prints a final report with `OK` / `FAIL`, duration, and log path per step
- Supports an optional `DEEP_MODE=1` for extra docker-compose smoke checks

### Defaults

- `REPO_URL`: current repository root
- `CHECKOUT_REF`: current branch
- `WORKSPACE_DIR`: `./my-local/workspace`
- `DEEP_MODE`: `0`

### Behavior notes

- Local `REPO_URL` values validate the current filesystem state, including uncommitted changes
- Remote `REPO_URL` values use `git clone --branch "$CHECKOUT_REF"`
- `DEEP_MODE=1` requires Docker Compose and a running Docker daemon for runtime smoke checks

### Usage

```bash
./scripts/test-ginger-massive.sh
```

### Useful overrides

```bash
REPO_URL=https://github.com/fvmoraes/ginger.git \
CHECKOUT_REF=main \
WORKSPACE_DIR=/tmp/ginger-massive \
./scripts/test-ginger-massive.sh
```

### Verbosity and color

```bash
VERBOSE=0 ./scripts/test-ginger-massive.sh
FORCE_COLOR=1 ./scripts/test-ginger-massive.sh
COLOR_ENABLED=0 ./scripts/test-ginger-massive.sh
DEEP_MODE=1 ./scripts/test-ginger-massive.sh
DEEP_MODE=1 DEEP_HEAVY=1 ./scripts/test-ginger-massive.sh
```
