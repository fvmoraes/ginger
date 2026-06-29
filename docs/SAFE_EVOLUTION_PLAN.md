# Ginger safe evolution plan

The product direction is now:

> Ginger is a safe project framework for Go. It understands a codebase,
> organizes its structure, and evolves existing projects without overwriting
> user work.

This supersedes roadmap priorities centered on adding more HTTP framework
packages. Those features can still exist, but only as modular capabilities
after the project-safety foundation is mature.

## Non-negotiable invariants

1. Every project-aware command resolves the real root before reading or writing.
2. Every configured path is relative, validated, and contained by that root.
3. A plan is complete and has no side effects; apply revalidates the plan.
4. Existing unowned files are preserved. Managed regions or reviewable patches
   are used instead of arbitrary source rewrites.
5. `ginger.yaml` is the contract for structure and generation rules.
6. Optional capabilities do not raise the core module's Go requirement.
7. Generated projects and existing-project fixtures must build and test.

## Delivery phases

### Foundation — implemented

- root discovery from nested directories;
- strict `ginger.yaml` loading, defaults, auto-detection, and path containment;
- safe plan/apply with stale-plan and symlink checks;
- ownership manifest and managed regions;
- root-aware generators, add-ons, docs, inspect JSON, and doctor JSON;
- existing-project test scanning with no-overwrite behavior;
- OpenAPI route discovery with patch fallback;
- optional OpenTelemetry submodule and per-capability Go checks.

### Hardening — next

- transactional multi-file rollback and durable backups;
- real unified diffs for patches and an explicit `ginger apply-patch` workflow;
- richer router adapters (stdlib, Chi, Gin, Echo, Fiber) without source guessing;
- Go AST/type-aware handler and response model inference;
- subprocess changes (`go get`, formatters) represented in plans and lockfiles;
- CLI end-to-end tests for every documented flag order and nested-directory use.

### Extensibility

- capability metadata loaded from a single registry;
- capability-specific validation and deterministic dependency resolution;
- versioned generator contracts and user/org templates;
- documentation and test generation driven by the shared inspection model.

### Product capabilities

Authentication, rate limiting, migrations, caching, metrics, eventing, and
other framework features remain candidates. They must be opt-in, independently
versioned where appropriate, and unable to weaken the safety invariants above.

## Release gates

- `go test ./...`, `go vet ./...`, and `go build ./...` pass for the core;
- nested modules and `examples/existing-api` pass their own tests;
- plan mode produces no filesystem changes;
- applying from a nested directory writes only below the detected root;
- existing source/test hashes remain unchanged unless ownership or force is
  explicit;
- a capability blocked by Go version does not block the Ginger core.
