package region

import (
	"strings"
	"testing"
)

// Matriz 6.7 (GIN-015) — validação explícita da malformação de marcadores.

func TestValidateRegionsNestedRejected(t *testing.T) {
	content := "// ginger:begin a\n// ginger:begin b\nx\n// ginger:end b\n// ginger:end a\n"
	err := ValidateRegions(content)
	if err == nil || !strings.Contains(err.Error(), "nested or interleaved") {
		t.Fatalf("expected nested error, got %v", err)
	}
	if _, err := InsertRegion(content, "new", "x"); err == nil {
		t.Fatal("InsertRegion must refuse nested markers")
	}
}

func TestValidateRegionsInterleavedRejected(t *testing.T) {
	content := "// ginger:begin a\n// ginger:begin b\n// ginger:end a\n// ginger:end b\n"
	err := ValidateRegions(content)
	if err == nil || !strings.Contains(err.Error(), "nested or interleaved") {
		t.Fatalf("expected interleaved error, got %v", err)
	}
}

func TestValidateRegionsUnclosedRejected(t *testing.T) {
	content := "// ginger:begin a\ncontent\n"
	err := ValidateRegions(content)
	if err == nil || !strings.Contains(err.Error(), "no end marker") {
		t.Fatalf("expected unclosed error, got %v", err)
	}
}

func TestValidateRegionsOrphanEndRejected(t *testing.T) {
	content := "// ginger:end a\n"
	err := ValidateRegions(content)
	if err == nil || !strings.Contains(err.Error(), "orphan end") {
		t.Fatalf("expected orphan end error, got %v", err)
	}
}

func TestValidateRegionsDuplicateRejected(t *testing.T) {
	content := "// ginger:begin a\nx\n// ginger:end a\n// ginger:begin a\ny\n// ginger:end a\n"
	err := ValidateRegions(content)
	if err == nil || !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestValidateRegionsMismatchedNamesRejected(t *testing.T) {
	content := "// ginger:begin a\n// ginger:end b\n"
	err := ValidateRegions(content)
	if err == nil || !strings.Contains(err.Error(), "mismatched") {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestValidateRegionsWellFormedOK(t *testing.T) {
	content := "// ginger:begin a\nx\n// ginger:end a\n\n// ginger:begin b\ny\n// ginger:end b\n"
	if err := ValidateRegions(content); err != nil {
		t.Fatalf("well-formed regions must pass: %v", err)
	}
}

func TestInsertRegionPreservesCRLF(t *testing.T) {
	content := "line1\r\nline2\r\n"
	got, err := InsertRegion(content, "routes", "new content")
	if err != nil {
		t.Fatalf("InsertRegion: %v", err)
	}
	if strings.Count(got, "\r\n") < 4 {
		t.Fatalf("expected CRLF preserved, got %q", got)
	}
	if strings.Contains(got, "\n\n") && !strings.Contains(got, "\r\n\r\n") {
		t.Fatalf("mixed EOLs introduced: %q", got)
	}
}

func TestReplaceRegionRejectsMalformed(t *testing.T) {
	content := "// ginger:begin routes\nold\n" // unclosed
	if _, err := ReplaceRegion(content, "routes", "new"); err == nil {
		t.Fatal("ReplaceRegion must refuse malformed markers")
	}
}

func TestInsertRegionEmptyContent(t *testing.T) {
	got, err := InsertRegion("", "routes", "x")
	if err != nil {
		t.Fatalf("InsertRegion empty: %v", err)
	}
	if !strings.Contains(got, "ginger:begin routes") {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestInsertRegionRoundTrip(t *testing.T) {
	content := "package main\n"
	got, err := InsertRegion(content, "routes", "generated")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	r := FindRegion(got, "routes")
	if r == nil || r.Content != "generated" {
		t.Fatalf("round trip failed: %+v", r)
	}
	replaced, err := ReplaceRegion(got, "routes", "updated")
	if err != nil {
		t.Fatalf("replace: %v", err)
	}
	if !strings.Contains(replaced, "updated") {
		t.Fatalf("replace lost content: %q", replaced)
	}
}

// FuzzValidateRegions: validação e operações nunca devem panicar (GIN-015).
func FuzzValidateRegions(f *testing.F) {
	seeds := []string{
		"// ginger:begin a\nx\n// ginger:end a\n",
		"// ginger:begin a\n// ginger:begin b\n// ginger:end b\n// ginger:end a\n",
		"// ginger:begin a\n",
		"# ginger:end a\n",
		"<!-- ginger:begin a -->\n<!-- ginger:end a -->\n",
		"// ginger:begin a\n// ginger:begin a\n",
		"no markers at all",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, content string) {
		_ = ValidateRegions(content)
		_ = FindRegion(content, "a")
		_ = ExtractRegions(content)
		if got, err := InsertRegion(content, "zzz-fuzz", "x"); err == nil && len(got) < len(content) {
			t.Fatalf("InsertRegion shrank content")
		}
	})
}
