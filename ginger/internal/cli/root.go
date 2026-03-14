// Package cli implements the Ginger CLI commands.
package cli

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.1.0"

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
		fmt.Println("ginger version " + version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`Ginger — Agilize e padronize projetos Go

Usage:
  ginger <command> [arguments]

Commands:
  new <name> [--type api|microservice|cli|worker]
                              Scaffold a new project
  run                         Run the app  (go run ./cmd/app)
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
                                observ.    : otel, prometheus
  doctor                      Diagnose project health
  version                     Print Ginger version
  help                        Show this help

Examples:
  ginger new my-api
  ginger new my-worker --type worker
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
