package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/fvmoraes/ginger/internal/capability"
	"github.com/fvmoraes/ginger/internal/integrations"
	"github.com/fvmoraes/ginger/internal/project"
)

const addUsage = `usage: ginger add <integration> [--plan] [--force]

Storage convention:
  platform/...              external infrastructure adapters
  internal/api/handlers/... ready-to-mount HTTP endpoints

If devops/docker/docker-compose.yml exists, Ginger also updates it with the
local infrastructure needed by the added integration when applicable.

database   : postgres, mysql, sqlite, sqlserver
orm        : gorm, sqlx, bun
nosql      : couchbase, mongodb
analytical : clickhouse
cache      : redis
messaging  : kafka, rabbitmq, nats, pubsub
protocols  : grpc, mcp
realtime   : sse, websocket
observ.    : otel, prometheus
docs       : swagger

Flags:
  --plan    Show what would be done without applying
  --force   Overwrite existing files`

func runAdd(args []string) {
	integrationName, planOnly, force, err := parseAddArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "add arguments: %v\n\n%s\n", err, addUsage)
		os.Exit(2)
	}
	if integrationName == "" {
		fmt.Fprintln(os.Stderr, addUsage)
		os.Exit(1)
	}

	// Find project root and load project context
	root, err := project.FindRoot(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'ginger init' to initialize a project or create one with 'ginger new'.")
		os.Exit(1)
	}

	prj, err := project.Load(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading project: %v\n", err)
		os.Exit(1)
	}

	// Check capability-level constraints (e.g., Go version for otel)
	if capErr := checkCapabilityConstraints(prj, integrationName); capErr != nil {
		fmt.Fprintf(os.Stderr, "add blocked: %v\n", capErr)
		os.Exit(1)
	}

	// Generate the integration plan
	p, err := integrations.Plan(integrationName, prj, prj.ShouldOverwrite() || force)
	if err != nil {
		fmt.Fprintf(os.Stderr, "add error: %v\n", err)
		os.Exit(1)
	}

	// Render plan
	p.Render()

	if planOnly {
		return
	}

	if p.HasErrors() {
		fmt.Fprintln(os.Stderr, "Cannot apply: plan has errors. Fix them or use --force.")
		os.Exit(1)
	}

	if !p.HasChanges() {
		fmt.Println("Nothing to do; existing files were preserved.")
		return
	}

	// Apply + post-apply with rollback (GIN-005): a failed `go get` undoes
	// the whole apply instead of leaving a half-installed integration.
	if err := integrations.ApplyWithRollback(p, integrationName, root); err != nil {
		fmt.Fprintf(os.Stderr, "add error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Integration '%s' added successfully!\n\n", integrationName)
}

func parseAddArgs(args []string) (name string, planOnly, force bool, err error) {
	for _, arg := range args {
		switch arg {
		case "--plan":
			planOnly = true
		case "--force":
			force = true
		default:
			if strings.HasPrefix(arg, "-") {
				return "", false, false, fmt.Errorf("unknown flag %s", arg)
			}
			if name != "" {
				return "", false, false, fmt.Errorf("expected one integration, got %q and %q", name, arg)
			}
			name = arg
		}
	}
	return name, planOnly, force, nil
}

func checkCapabilityConstraints(prj *project.Project, integrationName string) error {
	// GIN-004: fail-closed — the capability registry is derived from the
	// integration catalog, so an unknown name here means a programming error
	// or a tampered invocation path; refusing is safer than silently allowing.
	capabilityDef, ok := capability.DefaultRegistry().Get(integrationName)
	if !ok {
		return fmt.Errorf("integration %q is not cataloged — refusing to bypass constraints", integrationName)
	}
	if !capabilityDef.SupportsProjectType(prj.ProjectType()) {
		return fmt.Errorf("%s supports project types %v, current type is %s", integrationName, capabilityDef.ProjectTypes, prj.ProjectType())
	}
	if capabilityDef.MinGo != "" {
		return capabilityDef.CheckGoVersion(prj.GoVersion())
	}
	return nil
}
