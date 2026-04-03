// Package ginger provides a lightweight, opinionated Go framework for building web applications and APIs.
//
// # Install
//
//	go install github.com/fvmoraes/ginger/cmd/ginger@latest
//
// # Create a new project
//
//	Both long and short flags work for project types.
//	As flags longas e curtas funcionam para os tipos de projeto.
//
//	ginger new foobar --api | -a       # API     → cmd/foobar-api
//	ginger new foobar --service | -s   # Service → cmd/foobar-service
//	ginger new foobar --worker | -w    # Worker  → cmd/foobar-worker
//	ginger new foobar --cli | -c       # CLI     → cmd/foobar-cli
//	ginger new foobar                  # Generic → cmd/foobar
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
//	import "github.com/fvmoraes/ginger/pkg/telemetry"
//
// # Documentation
//
// https://github.com/fvmoraes/ginger
package ginger
