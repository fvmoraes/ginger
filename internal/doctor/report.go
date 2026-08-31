package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fvmoraes/ginger/internal/capability"
	"github.com/fvmoraes/ginger/internal/project"
)

// DiagnosticCheck is one serializable doctor result.
type DiagnosticCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// CapabilityBlock explains why a capability cannot be used.
type CapabilityBlock struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// Diagnostic is the structured doctor output.
type Diagnostic struct {
	Root                  string            `json:"root"`
	Checks                []DiagnosticCheck `json:"checks"`
	AvailableCapabilities []string          `json:"available_capabilities"`
	BlockedCapabilities   []CapabilityBlock `json:"blocked_capabilities,omitempty"`
}

// Healthy reports whether all mandatory checks passed.
func (d *Diagnostic) Healthy() bool {
	for _, check := range d.Checks {
		if check.Status == "fail" {
			return false
		}
	}
	return true
}

// Diagnose validates a resolved project from its real root.
func Diagnose(prj *project.Project) (*Diagnostic, error) {
	report, err := project.Inspect(prj)
	if err != nil {
		return nil, err
	}
	d := &Diagnostic{Root: prj.Root}
	d.add("ginger.yaml valid", report.GingerYAML, "run 'ginger init' to create the project contract")
	_, goModErr := os.Stat(filepath.Join(prj.Root, "go.mod"))
	d.add("go.mod present", goModErr == nil, "go.mod is required for Go code generation")
	d.add("project type detected", report.Type != "unknown", "set project.type in ginger.yaml")
	// GIN-008: a fresh scaffold has no tests by design — warn, don't fail.
	d.addWarnable("tests present", report.TestFiles > 0, "run 'ginger generate tests --scan --plan'")

	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = prj.Root
	if output, vetErr := cmd.CombinedOutput(); vetErr != nil {
		detail := compact(string(output), vetErr.Error())
		// GIN-008: vet failures caused by missing dependencies (before
		// `go mod tidy`) are an incomplete setup, not a broken project —
		// warn + suggest tidy; real code errors keep failing hard.
		if missingDeps(detail) {
			detail = strings.TrimSpace(detail) + "\n→ run 'go mod tidy' and re-run doctor"
			d.Checks = append(d.Checks, DiagnosticCheck{Name: "go vet", Status: "warn", Detail: detail})
		} else {
			d.Checks = append(d.Checks, DiagnosticCheck{Name: "go vet", Status: "fail", Detail: detail})
		}
	} else {
		d.Checks = append(d.Checks, DiagnosticCheck{Name: "go vet", Status: "pass"})
	}

	registry := capability.DefaultRegistry()
	for _, c := range registry.Available(prj.ProjectType(), prj.GoVersion()) {
		d.AvailableCapabilities = append(d.AvailableCapabilities, c.Name)
	}
	for _, blocked := range registry.Blocked(prj.ProjectType(), prj.GoVersion()) {
		d.BlockedCapabilities = append(d.BlockedCapabilities, CapabilityBlock{Name: blocked.Name, Reason: blocked.Reason})
	}
	return d, nil
}

func (d *Diagnostic) add(name string, passed bool, detail string) {
	status := "pass"
	if !passed {
		status = "fail"
	}
	d.Checks = append(d.Checks, DiagnosticCheck{Name: name, Status: status, Detail: detail})
}

// addWarnable adds a check whose absence is a warning (not a failure):
// expected on freshly scaffolded projects (GIN-008).
func (d *Diagnostic) addWarnable(name string, passed bool, detail string) {
	status := "pass"
	if !passed {
		status = "warn"
	}
	d.Checks = append(d.Checks, DiagnosticCheck{Name: name, Status: status, Detail: detail})
}

func compact(output, fallback string) string {
	if output == "" {
		return fallback
	}
	if len(output) > 500 {
		return fmt.Sprintf("%s...", output[:500])
	}
	return output
}

// missingDeps detects go vet failures caused by unresolvable imports
// (incomplete `go mod tidy` setup) rather than real code errors.
func missingDeps(detail string) bool {
	patterns := []string{
		"no required module provides package",
		"cannot find package",
		"no Go files in",
		"missing go.sum entry",
		"missing module for import",
	}
	for _, p := range patterns {
		if strings.Contains(detail, p) {
			return true
		}
	}
	return false
}
