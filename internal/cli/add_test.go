package cli

import "testing"

func TestParseAddArgsAcceptsFlagsAfterIntegration(t *testing.T) {
	name, planOnly, force, err := parseAddArgs([]string{"swagger", "--plan", "--force"})
	if err != nil {
		t.Fatalf("parseAddArgs: %v", err)
	}
	if name != "swagger" || !planOnly || !force {
		t.Fatalf("unexpected result: name=%q plan=%v force=%v", name, planOnly, force)
	}
}

func TestParseAddArgsRejectsUnknownFlag(t *testing.T) {
	if _, _, _, err := parseAddArgs([]string{"swagger", "--plna"}); err == nil {
		t.Fatal("expected unknown flag error")
	}
}
