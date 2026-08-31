package cli

import "testing"

func TestParseAddArgsAcceptsFlagsAfterIntegration(t *testing.T) {
	name, planOnly, force, _, _, err := parseAddArgs([]string{"swagger", "--plan", "--force"})
	if err != nil {
		t.Fatalf("parseAddArgs: %v", err)
	}
	if name != "swagger" || !planOnly || !force {
		t.Fatalf("unexpected result: name=%q plan=%v force=%v", name, planOnly, force)
	}
}

func TestParseAddArgsJSONAndQuiet(t *testing.T) {
	name, planOnly, force, asJSON, quiet, err := parseAddArgs([]string{"redis", "--plan", "--json", "--quiet"})
	_ = force
	if err != nil {
		t.Fatalf("parseAddArgs: %v", err)
	}
	if name != "redis" || !planOnly || !asJSON || !quiet {
		t.Fatalf("unexpected: %q plan=%v json=%v quiet=%v", name, planOnly, asJSON, quiet)
	}
}

func TestParseAddArgsRejectsUnknownFlag(t *testing.T) {
	if _, _, _, _, _, err := parseAddArgs([]string{"swagger", "--plna"}); err == nil {
		t.Fatal("expected unknown flag error")
	}
}
