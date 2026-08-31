// Package capability defines the modular capability system.
// Each capability represents a feature (swagger, redis, otel, etc.)
// that declares its requirements, supported project types, and generation logic.
package capability

import (
	"fmt"
	"github.com/fvmoraes/ginger/internal/integrations"
	"sort"
	"strconv"
	"strings"

	"github.com/fvmoraes/ginger/internal/plan"
	"github.com/fvmoraes/ginger/internal/project"
)

// Capability represents a modular feature that Ginger can add to a project.
type Capability struct {
	Name         string
	Description  string
	MinGo        string // minimum Go version required (empty = no constraint)
	ProjectTypes []string
	Plan         func(ctx Context) (*plan.Plan, error)
}

// Context holds the data available to a capability's Plan function.
type Context struct {
	Project *project.Project
	Force   bool
}

// CheckGoVersion returns an error if the current Go version is below the
// capability's minimum requirement.
func (c *Capability) CheckGoVersion(currentGo string) error {
	if c.MinGo == "" {
		return nil
	}
	minParts := strings.Split(strings.TrimPrefix(c.MinGo, "go"), ".")
	curParts := strings.Split(strings.TrimPrefix(currentGo, "go"), ".")

	if len(minParts) < 2 || len(curParts) < 2 {
		return fmt.Errorf("cannot compare Go versions %q and %q", currentGo, c.MinGo)
	}

	minMajor, err := strconv.Atoi(minParts[0])
	if err != nil {
		return fmt.Errorf("invalid minimum Go version %q", c.MinGo)
	}
	minMinor, err := strconv.Atoi(minParts[1])
	if err != nil {
		return fmt.Errorf("invalid minimum Go version %q", c.MinGo)
	}
	curMajor, err := strconv.Atoi(curParts[0])
	if err != nil {
		return fmt.Errorf("invalid project Go version %q", currentGo)
	}
	curMinor, err := strconv.Atoi(curParts[1])
	if err != nil {
		return fmt.Errorf("invalid project Go version %q", currentGo)
	}

	if curMajor < minMajor || (curMajor == minMajor && curMinor < minMinor) {
		return fmt.Errorf("%s requires Go >= %s, current Go %s", c.Name, c.MinGo, currentGo)
	}
	return nil
}

// DefaultRegistry returns the capabilities shipped with the CLI. Planning is
// delegated to the integration and generator packages to avoid package cycles.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	// GIN-004: derive capabilities from the integration catalog (single
	// source of truth) so constraints cannot be bypassed by absence here.
	for _, spec := range integrations.Catalog() {
		c := &Capability{
			Name:        spec.Name,
			Description: spec.Description,
			MinGo:       spec.MinGo,
		}
		if len(spec.ProjectTypes) > 0 {
			c.ProjectTypes = spec.ProjectTypes
		}
		r.Register(c)
	}
	// Non-integration capabilities (generators/features, not `add` targets).
	for _, c := range []*Capability{
		{Name: "tests", Description: "Tests for existing code"},
		{Name: "docker", Description: "Container development files"},
	} {
		r.Register(c)
	}
	return r
}

// SupportsProjectType returns true if the capability supports the given project type.
func (c *Capability) SupportsProjectType(pt string) bool {
	if len(c.ProjectTypes) == 0 {
		return true
	}
	for _, t := range c.ProjectTypes {
		if t == pt {
			return true
		}
	}
	return false
}

// Registry holds all registered capabilities.
type Registry struct {
	capabilities map[string]*Capability
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		capabilities: make(map[string]*Capability),
	}
}

// Register adds a capability to the registry.
func (r *Registry) Register(c *Capability) {
	r.capabilities[c.Name] = c
}

// Get returns a capability by name.
func (r *Registry) Get(name string) (*Capability, bool) {
	c, ok := r.capabilities[name]
	return c, ok
}

// List returns all registered capability names, sorted.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.capabilities))
	for name := range r.capabilities {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Available returns capabilities that are available for the given project type and Go version.
func (r *Registry) Available(projectType, goVersion string) []*Capability {
	var available []*Capability
	for _, c := range r.capabilities {
		if c.SupportsProjectType(projectType) && c.CheckGoVersion(goVersion) == nil {
			available = append(available, c)
		}
	}
	sort.Slice(available, func(i, j int) bool {
		return available[i].Name < available[j].Name
	})
	return available
}

// Blocked returns capabilities that are blocked for the given project type or Go version.
func (r *Registry) Blocked(projectType, goVersion string) []struct {
	Name   string
	Reason string
} {
	var blocked []struct {
		Name   string
		Reason string
	}
	for _, c := range r.capabilities {
		if !c.SupportsProjectType(projectType) {
			blocked = append(blocked, struct {
				Name   string
				Reason string
			}{c.Name, fmt.Sprintf("requires project type %v", c.ProjectTypes)})
			continue
		}
		if err := c.CheckGoVersion(goVersion); err != nil {
			blocked = append(blocked, struct {
				Name   string
				Reason string
			}{c.Name, err.Error()})
		}
	}
	sort.Slice(blocked, func(i, j int) bool {
		return blocked[i].Name < blocked[j].Name
	})
	return blocked
}
