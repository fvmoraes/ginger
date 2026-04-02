// Package cli implements the Ginger CLI commands.
package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

const fallbackVersion = "1.2.6"

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
	fmt.Print(`Ginger — Accelerate and standardize Go projects

Usage:
  ginger <command> [arguments]

Commands:
  new <name> [--api|--service|--worker|--cli]
                              Scaffold a new project
                                (no flag)  generic   → cmd/<name>
                                --api      api       → cmd/<name>-api
                                --service  service   → cmd/<name>-service
                                --worker   worker    → cmd/<name>-worker
                                --cli      cli       → cmd/<name>-cli
  run                         Run the app in dev mode
  build [output]              Build the binary
  generate handler  <name>    Generate an HTTP handler
  generate service  <name>    Generate a service
  generate repository <name>  Generate a repository
  generate crud <name>        Generate full CRUD (model+handler+service+repo+test)
  add <integration>           Add an integration:
                                databases  : postgres, mysql, sqlite, sqlserver
                                nosql      : couchbase, mongodb
                                analytical : clickhouse
                                cache      : redis
                                messaging  : kafka, rabbitmq, nats, pubsub
                                protocols  : grpc, mcp
                                realtime   : sse, websocket
                                observ.    : otel, prometheus
  doctor                      Diagnose project health
  version                     Print Ginger version
  help                        Show this help

Examples:
  ginger new foobar --api
  ginger new foobar --worker
  ginger generate crud user
  ginger add postgres
  ginger add grpc
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
