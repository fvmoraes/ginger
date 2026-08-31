package capability

import (
	"testing"

	"github.com/fvmoraes/ginger/internal/integrations"
)

// TestCatalogDrift (GIN-004): o capability registry deve cobrir EXATAMENTE o
// catálogo de integrações (fonte única de verdade) mais as capabilities de
// features que não são `add` targets.
func TestCatalogDrift(t *testing.T) {
	catalog := integrations.Catalog()
	if len(catalog) != 22 {
		t.Fatalf("expected 22 cataloged integrations, got %d", len(catalog))
	}

	registry := DefaultRegistry()
	for _, spec := range catalog {
		c, ok := registry.Get(spec.Name)
		if !ok {
			t.Errorf("capability %q missing — catalog and registry drifted", spec.Name)
			continue
		}
		if c.MinGo != spec.MinGo {
			t.Errorf("capability %q MinGo = %q, catalog says %q", spec.Name, c.MinGo, spec.MinGo)
		}
		if len(c.ProjectTypes) != len(spec.ProjectTypes) {
			t.Errorf("capability %q ProjectTypes = %v, catalog says %v", spec.Name, c.ProjectTypes, spec.ProjectTypes)
		}
	}

	// Feature capabilities que não são integrações.
	for _, name := range []string{"tests", "docker"} {
		if _, ok := registry.Get(name); !ok {
			t.Errorf("feature capability %q missing", name)
		}
	}
}

// TestCatalogSpecsValid (GIN-004): toda entry do catálogo tem os campos
// mínimos preenchidos.
func TestCatalogSpecsValid(t *testing.T) {
	for _, spec := range integrations.Catalog() {
		if spec.Name == "" || spec.File == "" || spec.Description == "" {
			t.Errorf("integration %q has empty required fields: %+v", spec.Name, spec)
		}
	}
}

// TestFailClosedMatrix (GIN-004): matriz integração×tipo×Go derivada do
// catálogo — restrições atuais preservadas (comportamento compatível).
func TestFailClosedMatrix(t *testing.T) {
	registry := DefaultRegistry()

	// otel: service/worker + minGo 1.25.
	otel, _ := registry.Get("otel")
	if otel.SupportsProjectType("generic") || otel.SupportsProjectType("cli") {
		t.Error("otel must not support generic/cli")
	}
	if !otel.SupportsProjectType("service") || !otel.SupportsProjectType("worker") {
		t.Error("otel must support service/worker")
	}
	if err := otel.CheckGoVersion("1.22"); err == nil {
		t.Error("otel must block go < 1.25")
	}
	if err := otel.CheckGoVersion("1.25"); err != nil {
		t.Errorf("otel must allow go 1.25: %v", err)
	}

	// swagger/grpc: apenas service.
	for _, name := range []string{"swagger", "grpc"} {
		c, _ := registry.Get(name)
		if c.SupportsProjectType("cli") || c.SupportsProjectType("worker") || c.SupportsProjectType("generic") {
			t.Errorf("%s must only support service", name)
		}
	}

	// mcp: service + cli.
	mcp, _ := registry.Get("mcp")
	if !mcp.SupportsProjectType("service") || !mcp.SupportsProjectType("cli") {
		t.Error("mcp must support service/cli")
	}
	if mcp.SupportsProjectType("worker") || mcp.SupportsProjectType("generic") {
		t.Error("mcp must not support worker/generic")
	}

	// Integrações sem restrição: permitidas em qualquer tipo.
	for _, name := range []string{"postgres", "mysql", "sqlite", "redis", "kafka", "mongodb", "sse", "websocket"} {
		c, ok := registry.Get(name)
		if !ok {
			t.Fatalf("integration %q missing from registry", name)
		}
		for _, pt := range []string{"generic", "service", "worker", "cli"} {
			if !c.SupportsProjectType(pt) {
				t.Errorf("%s must support %q (no restriction)", name, pt)
			}
		}
	}
}

// TestUnknownIntegrationFailsClosed (GIN-004): nomes fora do catálogo são
// recusados pelo checkCapabilityConstraints.
func TestUnknownIntegrationFailsClosed(t *testing.T) {
	if integrations.IsCataloged("totally-unknown-integration") {
		t.Fatal("unknown integration must not be cataloged")
	}
}
