// Package cli implements the Ginger CLI commands.
package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

const fallbackVersion = "1.3.0"

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
Minimal Go scaffolding that starts small and grows on demand.

Usage:
  ginger <command> [arguments]

Project Commands:
  new <name> [--service|--worker|--cli]        Create a new project
    default   generic  -> cmd/<name>
    --service service  -> cmd/<name>
    --worker  worker   -> cmd/<name>-worker
    --cli     cli      -> cmd/<name>

  run [args...]                                Run the detected app entrypoint
  build [output]                               Build the detected app entrypoint
  doctor                                       Diagnose project health

Generation Commands:
  generate crud <name>                         Generate a REST CRUD base (--service projects)
  generate command <name>                      Generate a Cobra subcommand (--cli projects)
  generate handler <name>                      Generate a worker handler (--worker projects)
  generate service <name>                      Generate a business service (--cli/--worker projects)
  generate test <name>                         Generate tests for a generated resource
  generate smoke-test                          Generate app smoke test under tests/integration
  generate swagger [name]                      Generate docs/openapi.json
                                               no name = starter spec
                                               name    = CRUD example for that resource

Integration Commands:
  add <integration>                            Add an integration file to the current project
    infrastructure adapters -> platform/...
    ready-to-mount HTTP endpoints -> internal/api/handlers/...
    updates devops/docker/docker-compose.yml when local infra is available

    databases   : postgres, mysql, sqlite, sqlserver
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
  ginger new foobar --worker
  ginger new foobar --cli
  ginger generate crud foobar
  ginger generate command deploy
  ginger generate handler order
  ginger generate service deployer
  ginger generate test foobar
  ginger generate smoke-test
  ginger generate swagger
  ginger generate swagger foobar
  ginger add postgres
  ginger add swagger
  ginger doctor
  ginger run
`)
}

// mustFlag returns a FlagSet for a subcommand, exiting on parse error.
func mustFlag(name string) *flag.FlagSet {
	return flag.NewFlagSet(name, flag.ExitOnError)
}

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fallbackVersion
	}

	mainVersion := strings.TrimPrefix(info.Main.Version, "v")
	if mainVersion != "" && mainVersion != "(devel)" && !isPseudoVersion(mainVersion) {
		return mainVersion
	}

	if mainVersion != "" && mainVersion != "(devel)" && isPseudoVersion(mainVersion) {
		return fallbackVersion
	}

	return fallbackVersion
}

func isPseudoVersion(v string) bool {
	return strings.Contains(v, "-0.") || strings.Contains(v, "+dirty")
}
