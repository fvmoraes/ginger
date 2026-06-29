package capability

import (
	"testing"

	"github.com/fvmoraes/ginger/internal/plan"
	"github.com/fvmoraes/ginger/internal/project"
)

func TestCapability_CheckGoVersion(t *testing.T) {
	tests := []struct {
		minGo     string
		currentGo string
		wantErr   bool
	}{
		{"", "1.22", false},
		{"1.25", "1.25", false},
		{"1.25", "1.26", false},
		{"1.25", "1.24", true},
		{"1.25", "1.22", true},
		{"go1.25", "go1.26", false},
	}

	for _, tt := range tests {
		c := &Capability{Name: "test", MinGo: tt.minGo}
		err := c.CheckGoVersion(tt.currentGo)
		if (err != nil) != tt.wantErr {
			t.Errorf("CheckGoVersion(min=%s, cur=%s) err=%v wantErr=%v",
				tt.minGo, tt.currentGo, err, tt.wantErr)
		}
	}
}

func TestCapability_SupportsProjectType(t *testing.T) {
	c := &Capability{ProjectTypes: []string{"service", "worker"}}

	if !c.SupportsProjectType("service") {
		t.Error("should support service")
	}
	if c.SupportsProjectType("cli") {
		t.Error("should not support cli")
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	r.Register(&Capability{
		Name:         "swagger",
		ProjectTypes: []string{"service"},
	})
	r.Register(&Capability{
		Name:  "otel",
		MinGo: "1.25",
	})

	c, ok := r.Get("swagger")
	if !ok {
		t.Fatal("expected swagger capability")
	}
	if !c.SupportsProjectType("service") {
		t.Error("swagger should support service")
	}

	available := r.Available("service", "1.24")
	if len(available) != 1 {
		t.Fatalf("expected 1 available capability for service/go1.24, got %d", len(available))
	}
	if available[0].Name != "swagger" {
		t.Errorf("expected swagger, got %s", available[0].Name)
	}

	blocked := r.Blocked("service", "1.24")
	if len(blocked) != 1 {
		t.Fatalf("expected 1 blocked capability, got %d", len(blocked))
	}
	if blocked[0].Name != "otel" {
		t.Errorf("expected otel blocked, got %s", blocked[0].Name)
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(&Capability{Name: "redis"})
	r.Register(&Capability{Name: "postgres"})

	names := r.List()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	if names[0] != "postgres" || names[1] != "redis" {
		t.Errorf("unexpected order: %v", names)
	}
}

// Ensure Context compiles
func TestContextType(t *testing.T) {
	ctx := Context{Force: true}
	_ = ctx
}

// Ensure Plan is referenced correctly
func TestPlanIntegration(t *testing.T) {
	p := &plan.Plan{ProjectRoot: "/tmp"}
	if p.ProjectRoot != "/tmp" {
		t.Error("plan field access")
	}
}

// Ensure Project is referenced correctly
func TestProjectIntegration(t *testing.T) {
	p := project.DefaultGingerYAML("service")
	if p.Project.Type != "service" {
		t.Error("default type mismatch")
	}
}
