module github.com/fvmoraes/ginger/example

go 1.25.0

require github.com/fvmoraes/ginger v0.0.0

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	go.opentelemetry.io/otel v1.42.0 // indirect
	go.opentelemetry.io/otel/trace v1.42.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/fvmoraes/ginger => ../
