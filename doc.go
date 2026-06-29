// Package ginger is a safe project framework for Go.
//
// Ginger helps you create, organize, inspect, and evolve Go projects without
// overwriting your work. It handles scaffolding, generation, and evolution so
// you can focus on business logic instead of repetitive setup.
//
// Ginger is not just another web framework. It is a project framework +
// safe generator + structure toolkit.
//
// # Install
//
//	go install github.com/fvmoraes/ginger/cmd/ginger@latest
//
// # Create a new project
//
//	ginger new foobar --service
//	ginger new foobar --worker
//	ginger new foobar --cli
//	ginger new foobar                # generic
//
// # Initialize an existing project
//
//	ginger init                      # detects structure, creates ginger.yaml
//	ginger inspect                   # analyze project structure
//
// # Safe generation (plan before apply)
//
//	ginger add swagger --plan        # see what would be created
//	ginger add postgres --plan       # plan without applying
//	ginger add redis --force         # overwrite existing files
//
//	ginger generate crud user --plan
//	ginger generate tests --scan --plan
//
// # Core Packages
//
// Import any package directly:
//
//	import "github.com/fvmoraes/ginger/pkg/app"
//	import "github.com/fvmoraes/ginger/pkg/router"
//	import "github.com/fvmoraes/ginger/pkg/middleware"
//	import "github.com/fvmoraes/ginger/pkg/response"
//	import "github.com/fvmoraes/ginger/pkg/logger"
//	import "github.com/fvmoraes/ginger/pkg/config"
//
// Optional OpenTelemetry support is a separate module (Go 1.25+):
//
//	go get github.com/fvmoraes/ginger/pkg/telemetry
//
// # Positioning
//
// Gin/Echo/Fiber are HTTP frameworks.
// Cobra is a CLI framework.
// Ginger is a project framework — it works with your existing code.
//
// # Documentation
//
// https://github.com/fvmoraes/ginger
package ginger
