package region

import (
	"strings"
	"testing"
)

const testFile = `package router

import "net/http"

func SetupRoutes(r *Router) {
	// ginger:begin routes
	r.GET("/api/v1/users", listUsers)
	r.POST("/api/v1/users", createUser)
	// ginger:end routes

	// ginger:begin health
	r.GET("/health", healthCheck)
	// ginger:end health
}
`

func TestFindRegion(t *testing.T) {
	r := FindRegion(testFile, "routes")
	if r == nil {
		t.Fatal("expected to find region 'routes'")
		return
	}
	content := strings.TrimSpace(r.Content)
	if !strings.Contains(content, "r.GET") {
		t.Errorf("unexpected region content: %s", content)
	}
}

func TestFindRegion_NotFound(t *testing.T) {
	r := FindRegion(testFile, "nonexistent")
	if r != nil {
		t.Error("expected nil for nonexistent region")
	}
}

func TestHasRegion(t *testing.T) {
	if !HasRegion(testFile, "routes") {
		t.Error("expected HasRegion('routes') = true")
	}
	if HasRegion(testFile, "nonexistent") {
		t.Error("expected HasRegion('nonexistent') = false")
	}
}

func TestReplaceRegion(t *testing.T) {
	newRoutes := "r.GET(\"/api/v2/items\", listItems)"
	result, err := ReplaceRegion(testFile, "routes", newRoutes)
	if err != nil {
		t.Fatalf("ReplaceRegion: %v", err)
	}
	if !strings.Contains(result, newRoutes) {
		t.Errorf("expected new content in result:\n%s", result)
	}
	if strings.Contains(result, "r.GET(\"/api/v1/users\"") {
		t.Error("old content should be replaced")
	}
	if !strings.Contains(result, "ginger:begin health") {
		t.Error("health region should be preserved")
	}
}

func TestReplaceRegion_NotFound(t *testing.T) {
	_, err := ReplaceRegion(testFile, "nonexistent", "content")
	if err == nil {
		t.Error("expected error for nonexistent region")
	}
}

func TestInsertRegion(t *testing.T) {
	content := "package main\n"
	result, err := InsertRegion(content, "imports", `import "fmt"`)
	if err != nil {
		t.Fatalf("InsertRegion: %v", err)
	}
	if !strings.Contains(result, "ginger:begin imports") {
		t.Error("expected begin marker in result")
	}
	if !strings.Contains(result, `import "fmt"`) {
		t.Error("expected inserted content in result")
	}
}

func TestInsertRegion_AlreadyExists(t *testing.T) {
	_, err := InsertRegion(testFile, "routes", "new")
	if err == nil {
		t.Error("expected error when inserting existing region")
	}
}

func TestExtractRegions(t *testing.T) {
	regions := ExtractRegions(testFile)
	if len(regions) != 2 {
		t.Fatalf("expected 2 regions, got %d", len(regions))
	}
	names := []string{regions[0].Name, regions[1].Name}
	if names[0] != "routes" || names[1] != "health" {
		t.Errorf("unexpected region names: %v", names)
	}
}

func TestFindRegionRejectsMismatchedEndMarker(t *testing.T) {
	content := "// ginger:begin routes\nvalue\n// ginger:end other\n"
	if got := FindRegion(content, "routes"); got != nil {
		t.Fatalf("mismatched region should not be managed: %+v", got)
	}
}
