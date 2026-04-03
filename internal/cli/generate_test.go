package cli

import (
	"strings"
	"testing"
)

func TestRunGenerateTestRejectsLegacyScopes(t *testing.T) {
	err := runGenerateTest([]string{"user", "all"})
	if err == nil {
		t.Fatal("expected legacy scope error, got nil")
	}

	if !strings.Contains(err.Error(), "test scopes are no longer supported") {
		t.Fatalf("expected legacy scope guidance, got %v", err)
	}
}

func TestRunGenerateTestRequiresNameOrApp(t *testing.T) {
	err := runGenerateTest(nil)
	if err == nil {
		t.Fatal("expected usage error, got nil")
	}

	if !strings.Contains(err.Error(), "usage: ginger generate test <name|app>") {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestRunGenerateSmokeTestRejectsArguments(t *testing.T) {
	err := runGenerateSmokeTest([]string{"app"})
	if err == nil {
		t.Fatal("expected usage error, got nil")
	}

	if !strings.Contains(err.Error(), "usage: ginger generate smoke-test") {
		t.Fatalf("expected smoke-test usage error, got %v", err)
	}
}

func TestValidateGeneratorForProjectType(t *testing.T) {
	tests := []struct {
		name        string
		projectType string
		kind        string
		wantErr     string
	}{
		{name: "crud allowed for service", projectType: "service", kind: "crud"},
		{name: "command rejected outside cli", projectType: "service", kind: "command", wantErr: "--cli"},
		{name: "handler rejected outside worker", projectType: "cli", kind: "handler", wantErr: "--worker"},
		{name: "service allowed for cli", projectType: "cli", kind: "service"},
		{name: "service rejected for generic", projectType: "generic", kind: "service", wantErr: "--cli and --worker"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateGeneratorForProjectType(tc.projectType, tc.kind)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
			}
		})
	}
}
