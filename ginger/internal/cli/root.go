// Package cli implements the Ginger CLI commands.
package cli

import (
	"flag"
	"fmt"
	"os"
)

// Run is the CLI entrypoint.
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
	case "version":
		fmt.Println("ginger version 0.1.0")
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`Ginger - Opinionated Go Framework

Usage:
  ginger <command> [arguments]

Commands:
  new <project-name>          Scaffold a new Ginger project
  run                         Run the application (go run ./cmd/app)
  build                       Build the application binary
  generate handler <name>     Generate a new HTTP handler
  generate service <name>     Generate a new service
  generate repository <name>  Generate a new repository
  version                     Print the Ginger version
  help                        Show this help message

Examples:
  ginger new my-api
  ginger generate handler user
  ginger generate service order
  ginger run
`)
}

// mustFlag returns a FlagSet for a subcommand.
func mustFlag(name string) *flag.FlagSet {
	return flag.NewFlagSet(name, flag.ExitOnError)
}
