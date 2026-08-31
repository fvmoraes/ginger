package project

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// PathStatus describes one configured structural location.
type PathStatus struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	Files  int    `json:"files,omitempty"`
}

// Route describes an HTTP route discovered in the project.
// Source/Confidence (GIN-014): annotation (authoritative) > ast (composed
// prefixes, real call expressions) > regex fallback.
type Route struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Source     string `json:"source,omitempty"`
	Confidence string `json:"confidence,omitempty"`
}

// Inspection is the stable, serializable project analysis model.
type Inspection struct {
	Root       string                `json:"root"`
	Module     string                `json:"module,omitempty"`
	Type       string                `json:"type"`
	GoVersion  string                `json:"go_version,omitempty"`
	GingerYAML bool                  `json:"ginger_yaml"`
	Structure  map[string]PathStatus `json:"structure"`
	Routes     []Route               `json:"routes,omitempty"`
	Features   []string              `json:"features,omitempty"`
	Datastores []string              `json:"datastores,omitempty"`
	TestFiles  int                   `json:"test_files"`
	Warnings   []string              `json:"warnings,omitempty"`
}

var (
	annotationRoute = regexp.MustCompile(`^\s*//\s*ginger:route\s+([A-Z]+)\s+(\S+)`)
	methodRoute     = regexp.MustCompile(`\.(GET|POST|PUT|PATCH|DELETE|OPTIONS|HEAD)\(\s*"([^"]+)"`)
	handleRoute     = regexp.MustCompile(`HandleFunc\(\s*"(?:(GET|POST|PUT|PATCH|DELETE|OPTIONS|HEAD)\s+)?([^"]+)"`)
)

// Inspect analyzes files without executing project code.
func Inspect(p *Project) (*Inspection, error) {
	report := &Inspection{
		Root: p.Root, Module: p.Module, Type: p.ProjectType(), GoVersion: p.GoVersion(),
		GingerYAML: p.HasGingerYAML(), Structure: make(map[string]PathStatus),
	}
	keys := []string{"cmd", "api", "handlers", "middlewares", "models", "services", "repositories", "ports", "adapters", "config", "docs", "tests", "migrations"}
	for _, key := range keys {
		path, err := p.ResolvePath(key)
		if err != nil {
			continue
		}
		rel, _ := filepath.Rel(p.Root, path)
		status := PathStatus{Path: filepath.ToSlash(rel)}
		if info, statErr := os.Stat(path); statErr == nil {
			status.Exists = true
			if info.IsDir() {
				entries, _ := os.ReadDir(path)
				status.Files = len(entries)
			} else {
				status.Files = 1
			}
		}
		report.Structure[key] = status
	}

	features := map[string]bool{}
	datastores := map[string]bool{}
	err := filepath.WalkDir(p.Root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".ginger", "vendor", "node_modules":
				if path != p.Root {
					return filepath.SkipDir
				}
			}
			return nil
		}
		rel, _ := filepath.Rel(p.Root, path)
		rel = filepath.ToSlash(rel)
		lower := strings.ToLower(rel)
		switch {
		case strings.HasSuffix(lower, "_test.go"):
			report.TestFiles++
		case lower == "dockerfile" || strings.HasSuffix(lower, "/dockerfile"):
			features["docker"] = true
		case strings.Contains(lower, "docker-compose") || strings.Contains(lower, "compose.yaml"):
			features["docker-compose"] = true
		case strings.Contains(lower, "kubernetes") || strings.Contains(lower, "/k8s/"):
			features["kubernetes"] = true
		case strings.Contains(lower, "chart.yaml") || strings.Contains(lower, "/helm/"):
			features["helm"] = true
		case strings.Contains(lower, "openapi") || strings.Contains(lower, "swagger"):
			features["swagger/openapi"] = true
		}
		if !strings.HasSuffix(lower, ".go") && filepath.Base(path) != "go.mod" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil || len(data) > 2<<20 {
			return nil
		}
		text := string(data)
		for needle, name := range map[string]string{
			"go.opentelemetry.io/otel": "opentelemetry",
			"github.com/redis/":        "redis", "github.com/lib/pq": "postgres",
			"pgx": "postgres", "go-sql-driver/mysql": "mysql", "go-sqlite3": "sqlite",
			"mongo-driver": "mongodb",
		} {
			if strings.Contains(text, needle) {
				if name == "opentelemetry" {
					features[name] = true
				} else {
					datastores[name] = true
				}
			}
		}
		if strings.HasSuffix(lower, ".go") {
			generatedSwagger := strings.HasSuffix(lower, "/swagger.go") && strings.Contains(text, "Arquivo gerado pelo Ginger")
			if !generatedSwagger {
				// GIN-014: AST first (composed prefixes, annotations with
				// high confidence); regex fallback fills gaps at low confidence.
				astFound := astRoutes(text, rel)
				regexFound := scanRoutes(text, rel)
				report.Routes = append(report.Routes, mergeRouteResults(astFound, regexFound)...)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("inspect project: %w", err)
	}
	for name := range features {
		report.Features = append(report.Features, name)
	}
	for name := range datastores {
		report.Datastores = append(report.Datastores, name)
	}
	sort.Strings(report.Features)
	sort.Strings(report.Datastores)
	sort.Slice(report.Routes, func(i, j int) bool {
		if report.Routes[i].Path == report.Routes[j].Path {
			return report.Routes[i].Method < report.Routes[j].Method
		}
		return report.Routes[i].Path < report.Routes[j].Path
	})
	if !report.GingerYAML {
		report.Warnings = append(report.Warnings, "ginger.yaml is missing; paths were auto-detected")
	}
	return report, nil
}

func scanRoutes(content, file string) []Route {
	var routes []Route
	seen := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		matches := [][]string{}
		if match := annotationRoute.FindStringSubmatch(text); match != nil {
			matches = append(matches, match)
		}
		if match := methodRoute.FindStringSubmatch(text); match != nil {
			matches = append(matches, match)
		}
		if match := handleRoute.FindStringSubmatch(text); match != nil {
			method := match[1]
			if method == "" {
				method = "ANY"
			}
			matches = append(matches, []string{"", method, match[2]})
		}
		for _, match := range matches {
			key := match[1] + " " + match[2] + " " + file
			if seen[key] {
				continue
			}
			seen[key] = true
			routes = append(routes, Route{Method: match[1], Path: match[2], File: file, Line: line})
		}
	}
	return routes
}
