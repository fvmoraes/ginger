package project

import "testing"

// TestMergeGingerYAML (GIN-021): customizações do usuário vencem; ausentes vêm dos defaults.
func TestMergeGingerYAML(t *testing.T) {
	existing := &GingerYAML{
		Project:   ProjectConfig{Type: "service", Root: "custom-root"},
		Structure: StructureConfig{Handlers: "custom/handlers", Docs: "custom/docs"},
		Rules:     RulesConfig{Overwrite: true, RequirePlanBeforeApply: true},
	}
	detected := DefaultGingerYAML("library")
	got := MergeGingerYAML(existing, detected)

	if got.Project.Type != "service" {
		t.Errorf("Type: want custom service, got %q", got.Project.Type)
	}
	if got.Project.Root != "custom-root" {
		t.Errorf("Root: want custom-root, got %q", got.Project.Root)
	}
	if got.Structure.Handlers != "custom/handlers" {
		t.Errorf("Handlers: want custom, got %q", got.Structure.Handlers)
	}
	if got.Structure.Docs != "custom/docs" {
		t.Errorf("Docs: want custom, got %q", got.Structure.Docs)
	}
	if got.Structure.Cmd == "" || got.Structure.Models == "" {
		t.Errorf("missing fields should be filled from defaults; cmd=%q models=%q", got.Structure.Cmd, got.Structure.Models)
	}
	if !got.Rules.Overwrite || !got.Rules.RequirePlanBeforeApply {
		t.Errorf("rules should be preserved: %+v", got.Rules)
	}

	if nilMerged := MergeGingerYAML(nil, detected); nilMerged != detected {
		t.Error("nil existing should return detected")
	}
}
