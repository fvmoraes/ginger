// Package cli implements the Ginger CLI commands.
package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/fvmoraes/ginger/internal/buildinfo"
)

// Run is the CLI entrypoint. It dispatches to the appropriate subcommand.
func Run() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "new":
		runNew(args)
	case "run":
		runRun(args)
	case "build":
		runBuild(args)
	case "generate", "g":
		runGenerate(args)
	case "add":
		runAdd(args)
	case "init":
		runInit(args)
	case "inspect":
		runInspect(args)
	case "docs":
		runDocs(args)
	case "doctor":
		runDoctor(args)
	case "version", "--version", "-v":
		fmt.Println("ginger " + buildVersion())
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`Ginger
Safe project framework for Go.
It helps you create, organize, inspect and evolve Go projects without overwriting your work.

Usage:
  ginger <command> [arguments]

Project Commands:
  new <name> [--service|--worker|--cli]        Create a new project
    default   generic  -> cmd/<name>
    --service service  -> cmd/<name>
    --worker  worker   -> cmd/<name>-worker
    --cli     cli      -> cmd/<name>

  init [--force]                                Initialize ginger.yaml in an existing project
  inspect [--json]                              Analyze current project structure
  docs [--plan] [--force]                       Generate inspection-based documentation safely
  run [args...]                                 Run the detected app entrypoint
  build [output]                                Build the detected app entrypoint
  doctor                                        Diagnose project health

Generation Commands:
  generate crud <name> [--plan] [--force]       Generate a REST CRUD base (--service projects)
  generate command <name>                       Generate a Cobra subcommand (--cli projects)
  generate handler <name>                       Generate a worker handler (--worker projects)
  generate service <name>                       Generate a business service (--cli/--worker projects)
  generate test <name>                          Generate tests for a generated resource
  generate tests --scan [--plan]                Generate TODO tests for existing code safely
  generate smoke-test                           Generate app smoke test under tests/integration
  generate swagger [name]                       Generate docs/openapi.json

Integration Commands:
  add <integration> [--plan] [--force]          Add an integration to the current project
    --plan    show what would be done without applying
    --force   overwrite existing files

    infrastructure adapters -> platform/...
    ready-to-mount HTTP endpoints -> internal/api/handlers/...
    proposes a patch for unowned docker-compose files

    databases   : postgres, mysql, sqlite, sqlserver
    orm         : gorm, sqlx, bun
    nosql       : couchbase, mongodb
    analytical  : clickhouse
    cache       : redis
    messaging   : kafka, rabbitmq, nats, pubsub
    protocols   : grpc, mcp
    realtime    : sse, websocket
    observ.     : otel, prometheus
    docs        : swagger

Other Commands:
  version                                      Print ginger x.y.z
  help                                         Show this help

Aliases:
  generate = g
  version  = -v, --version
  help     = -h, --help

Examples:
  ginger new foobar --service
  ginger init
  ginger inspect
  ginger generate crud foobar
  ginger add swagger --plan
  ginger add postgres
  ginger doctor
  ginger run
`)
}

// mustFlag returns a FlagSet for a subcommand, exiting on parse error.
func mustFlag(name string) *flag.FlagSet {
	return flag.NewFlagSet(name, flag.ExitOnError)
}

func buildVersion() string {
	return buildinfo.Version()
}
