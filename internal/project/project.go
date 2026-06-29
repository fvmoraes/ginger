// Package project provides project-aware operations: root discovery, ginger.yaml
// loading, and automatic structure detection for existing Go projects.
package project

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrNoProject is returned when no project root can be found.
var ErrNoProject = errors.New("no project root found")

// GingerYAML represents the ginger.yaml contract file.
type GingerYAML struct {
	Project   ProjectConfig   `yaml:"project"`
	Structure StructureConfig `yaml:"structure"`
	Rules     RulesConfig     `yaml:"rules"`
}

// ProjectConfig holds project-level metadata.
type ProjectConfig struct {
	Type string `yaml:"type"` // service, worker, cli, generic, library
	Root string `yaml:"root"`
}

// StructureConfig maps logical names to directory paths.
type StructureConfig struct {
	Cmd          string `yaml:"cmd"`
	API          string `yaml:"api"`
	Handlers     string `yaml:"handlers"`
	Middlewares  string `yaml:"middlewares"`
	Models       string `yaml:"models"`
	Services     string `yaml:"services"`
	Repositories string `yaml:"repositories"`
	Ports        string `yaml:"ports"`
	Adapters     string `yaml:"adapters"`
	Config       string `yaml:"config"`
	Docs         string `yaml:"docs"`
	Tests        string `yaml:"tests"`
	Migrations   string `yaml:"migrations"`
}

// RulesConfig controls safe generation behavior.
type RulesConfig struct {
	Overwrite              bool `yaml:"overwrite"`
	CreateMissingDirs      bool `yaml:"create_missing_dirs"`
	RequirePlanBeforeApply bool `yaml:"require_plan_before_apply"`
}

// Project holds the resolved project context.
type Project struct {
	Root     string
	YAML     *GingerYAML
	Module   string
	IsGinger bool
}

// DefaultGingerYAML returns a default ginger.yaml for autodetected projects.
func DefaultGingerYAML(projectType string) *GingerYAML {
	return &GingerYAML{
		Project: ProjectConfig{
			Type: projectType,
			Root: ".",
		},
		Structure: StructureConfig{
			Cmd:          "cmd/app",
			API:          "internal/api",
			Handlers:     "internal/api/handlers",
			Middlewares:  "internal/api/middlewares",
			Models:       "internal/models",
			Services:     "internal/services",
			Repositories: "internal/repositories",
			Ports:        "internal/ports",
			Adapters:     "internal/adapters",
			Config:       "internal/config",
			Docs:         "docs",
			Tests:        "tests",
			Migrations:   "migrations",
		},
		Rules: RulesConfig{
			Overwrite:              false,
			CreateMissingDirs:      true,
			RequirePlanBeforeApply: true,
		},
	}
}

// FindRoot walks up from startDir looking for project markers in priority order:
// 1. ginger.yaml, 2. go.mod, 3. .git.
func FindRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("project root: %w", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(dir); resolveErr == nil {
		dir = resolved
	}

	for {
		if p, ok := hasMarker(dir, "ginger.yaml"); ok {
			return p, nil
		}
		if p, ok := hasMarker(dir, "go.mod"); ok {
			return p, nil
		}
		if p, ok := hasMarker(dir, ".git"); ok {
			return p, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNoProject
		}
		dir = parent
	}
}

// Load loads the project context from the given root directory.
// If ginger.yaml exists, it is loaded and validated.
// If it does not exist, the project structure is auto-detected and a default config is returned.
func Load(root string) (*Project, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("project root: %w", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(root); resolveErr == nil {
		root = resolved
	}
	p := &Project{Root: filepath.Clean(root)}

	gingerPath := filepath.Join(root, "ginger.yaml")
	data, err := os.ReadFile(gingerPath)
	if err == nil {
		detected := autoDetect(root, "")
		gy := DefaultGingerYAML(detected.Project.Type)
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(true)
		if err := decoder.Decode(gy); err != nil {
			return nil, fmt.Errorf("ginger.yaml: %w", err)
		}
		if err := Validate(root, gy); err != nil {
			return nil, fmt.Errorf("ginger.yaml: %w", err)
		}
		p.YAML = gy
		p.IsGinger = true
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read ginger.yaml: %w", err)
	} else {
		p.YAML = autoDetect(root, "")
	}

	base, err := p.BaseDir()
	if err != nil {
		return nil, err
	}
	if modPath, readErr := readModulePath(filepath.Join(base, "go.mod")); readErr == nil {
		p.Module = modPath
	}
	if !p.IsGinger {
		p.YAML = autoDetect(root, p.Module)
	}
	return p, nil
}

// Detect builds a project context from the existing filesystem without
// reading ginger.yaml. It is used by ginger init and never moves files.
func Detect(root string) (*Project, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(root); resolveErr == nil {
		root = resolved
	}
	p := &Project{Root: filepath.Clean(root)}
	if module, readErr := readModulePath(filepath.Join(p.Root, "go.mod")); readErr == nil {
		p.Module = module
	}
	p.YAML = autoDetect(p.Root, p.Module)
	return p, nil
}

// Validate verifies that the contract cannot address files outside the
// discovered project root and that declared values are supported.
func Validate(root string, gy *GingerYAML) error {
	if gy == nil {
		return errors.New("empty configuration")
	}
	allowedTypes := map[string]bool{
		"service": true, "worker": true, "cli": true, "generic": true,
		"library": true, "unknown": true,
	}
	if !allowedTypes[gy.Project.Type] {
		return fmt.Errorf("unsupported project.type %q", gy.Project.Type)
	}
	if gy.Project.Root == "" {
		gy.Project.Root = "."
	}
	paths := map[string]string{
		"project.root":           gy.Project.Root,
		"structure.cmd":          gy.Structure.Cmd,
		"structure.api":          gy.Structure.API,
		"structure.handlers":     gy.Structure.Handlers,
		"structure.middlewares":  gy.Structure.Middlewares,
		"structure.models":       gy.Structure.Models,
		"structure.services":     gy.Structure.Services,
		"structure.repositories": gy.Structure.Repositories,
		"structure.ports":        gy.Structure.Ports,
		"structure.adapters":     gy.Structure.Adapters,
		"structure.config":       gy.Structure.Config,
		"structure.docs":         gy.Structure.Docs,
		"structure.tests":        gy.Structure.Tests,
		"structure.migrations":   gy.Structure.Migrations,
	}
	for name, value := range paths {
		if value == "" {
			continue
		}
		if filepath.IsAbs(value) {
			return fmt.Errorf("%s must be relative to the project root", name)
		}
		clean := filepath.Clean(value)
		if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("%s escapes the project root: %q", name, value)
		}
	}
	base := filepath.Join(root, filepath.Clean(gy.Project.Root))
	if !pathWithin(root, base) {
		return fmt.Errorf("project.root escapes the project root: %q", gy.Project.Root)
	}
	return nil
}

// HasGingerYAML returns true if ginger.yaml exists at the project root.
func (p *Project) HasGingerYAML() bool {
	_, err := os.Stat(filepath.Join(p.Root, "ginger.yaml"))
	return err == nil
}

// BaseDir returns the directory used as the base for structure paths.
func (p *Project) BaseDir() (string, error) {
	if p.YAML == nil {
		return "", errors.New("project configuration is not loaded")
	}
	base := filepath.Join(p.Root, filepath.Clean(p.YAML.Project.Root))
	if !pathWithin(p.Root, base) {
		return "", fmt.Errorf("project.root escapes the project root: %q", p.YAML.Project.Root)
	}
	return base, nil
}

// ResolvePath resolves a logical structure key inside the project root.
func (p *Project) ResolvePath(key string) (string, error) {
	if p.YAML == nil {
		return "", errors.New("project configuration is not loaded")
	}
	s := p.YAML.Structure
	m := map[string]string{
		"cmd":          s.Cmd,
		"api":          s.API,
		"handlers":     s.Handlers,
		"middlewares":  s.Middlewares,
		"models":       s.Models,
		"services":     s.Services,
		"repositories": s.Repositories,
		"ports":        s.Ports,
		"adapters":     s.Adapters,
		"config":       s.Config,
		"docs":         s.Docs,
		"tests":        s.Tests,
		"migrations":   s.Migrations,
	}
	v, ok := m[key]
	if !ok {
		return "", fmt.Errorf("unknown structure key %q", key)
	}
	if v == "" {
		return "", fmt.Errorf("structure.%s is not configured", key)
	}
	base, err := p.BaseDir()
	if err != nil {
		return "", err
	}
	resolved := filepath.Join(base, filepath.Clean(v))
	if !pathWithin(p.Root, resolved) {
		return "", fmt.Errorf("structure.%s escapes the project root: %q", key, v)
	}
	return resolved, nil
}

// ResolveRelative resolves a project-relative path after validating containment.
func (p *Project) ResolveRelative(path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("path must be relative to the project root: %q", path)
	}
	base, err := p.BaseDir()
	if err != nil {
		return "", err
	}
	resolved := filepath.Join(base, filepath.Clean(path))
	if !pathWithin(p.Root, resolved) {
		return "", fmt.Errorf("path escapes the project root: %q", path)
	}
	return resolved, nil
}

// ShouldOverwrite returns true if rules allow overwriting files.
func (p *Project) ShouldOverwrite() bool {
	if p.YAML == nil {
		return false
	}
	return p.YAML.Rules.Overwrite
}

// ProjectType returns the detected project type.
func (p *Project) ProjectType() string {
	if p.YAML == nil {
		return "generic"
	}
	if p.YAML.Project.Type != "" {
		return p.YAML.Project.Type
	}
	return "generic"
}

// GoVersion returns the version declared by the project's go.mod.
func (p *Project) GoVersion() string {
	base, err := p.BaseDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(base, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "go" {
			return fields[1]
		}
	}
	return ""
}

// autoDetect attempts to determine project structure from directories and files.
func autoDetect(root, module string) *GingerYAML {
	apiDir := detectAPIDir(root)
	handlersDir := detectNamedDir(root, "handlers", "handler", "endpoints")
	candidates := []struct {
		projectType string
		markers     []string
	}{
		{"service", []string{handlersDir, apiDir, "internal/api/handlers", "internal/api"}},
		{"worker", []string{"internal/worker"}},
		{"cli", []string{"internal/commands"}},
	}

	projectType := "generic"
	for _, candidate := range candidates {
		for _, m := range candidate.markers {
			if m == "" {
				continue
			}
			if info, err := os.Stat(filepath.Join(root, m)); err == nil && info.IsDir() {
				projectType = candidate.projectType
				break
			}
		}
		if projectType != "generic" {
			break
		}
	}

	// If module path suggests a library (no cmd/ or main package), mark as library.
	if projectType == "generic" {
		if _, err := os.Stat(filepath.Join(root, "cmd")); os.IsNotExist(err) {
			if module != "" && !strings.HasSuffix(module, "/example") {
				projectType = "library"
			}
		}
	}

	gy := DefaultGingerYAML(projectType)

	// Detect actual cmd dir
	gy.Structure.Cmd = detectCmdDir(root)

	// Detect actual handler dirs
	if apiDir != "" {
		gy.Structure.API = apiDir
	}
	if handlersDir != "" {
		gy.Structure.Handlers = handlersDir
	} else if d := detectExistingDir(root, "internal/api/handlers"); d != "" {
		gy.Structure.Handlers = d
	}
	if d := detectExistingDir(root, "internal/api/middlewares"); d != "" {
		gy.Structure.Middlewares = d
	}
	if d := firstDetectedDir(root, "models", "model", "domain"); d != "" {
		gy.Structure.Models = d
	}
	if d := firstDetectedDir(root, "services", "service", "usecase", "usecases", "core"); d != "" {
		gy.Structure.Services = d
	}
	if d := firstDetectedDir(root, "repositories", "repository", "store", "storage"); d != "" {
		gy.Structure.Repositories = d
	}
	if d := detectExistingDir(root, "internal/ports"); d != "" {
		gy.Structure.Ports = d
	}
	if d := detectExistingDir(root, "internal/adapters"); d != "" {
		gy.Structure.Adapters = d
	}
	if d := detectExistingDir(root, "internal/config"); d != "" {
		gy.Structure.Config = d
	}
	if _, err := os.Stat(filepath.Join(root, "configs")); err == nil {
		gy.Structure.Config = "configs"
	}
	if d := detectExistingDir(root, "docs"); d != "" {
		gy.Structure.Docs = d
	}
	if d := detectExistingDir(root, "tests"); d != "" {
		gy.Structure.Tests = d
	}
	if d := detectExistingDir(root, "migrations"); d != "" {
		gy.Structure.Migrations = d
	}

	return gy
}

func firstDetectedDir(root string, names ...string) string {
	for _, name := range names {
		if d := detectExistingDir(root, filepath.Join("internal", name)); d != "" {
			return d
		}
	}
	return detectNamedDir(root, names...)
}

func detectNamedDir(root string, names ...string) string {
	wanted := make(map[string]bool, len(names))
	for _, name := range names {
		wanted[name] = true
	}
	var found []string
	internalRoot := filepath.Join(root, "internal")
	_ = filepath.WalkDir(internalRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !entry.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		if strings.Count(filepath.Clean(rel), string(filepath.Separator)) > 3 {
			return filepath.SkipDir
		}
		if wanted[strings.ToLower(entry.Name())] {
			found = append(found, filepath.ToSlash(rel))
			return filepath.SkipDir
		}
		return nil
	})
	sort.Strings(found)
	if len(found) > 0 {
		return found[0]
	}
	return ""
}

func detectAPIDir(root string) string {
	for _, candidate := range []string{"internal/api", "internal/httpapi", "internal/http", "api"} {
		if d := detectExistingDir(root, candidate); d != "" {
			return d
		}
	}
	return ""
}

func detectExistingDir(root, path string) string {
	if info, err := os.Stat(filepath.Join(root, path)); err == nil && info.IsDir() {
		return path
	}
	return ""
}

func detectCmdDir(root string) string {
	entries, err := os.ReadDir(filepath.Join(root, "cmd"))
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			return "cmd/" + e.Name()
		}
	}
	return ""
}

func hasMarker(dir, name string) (string, bool) {
	info, err := os.Stat(filepath.Join(dir, name))
	if err == nil && ((name == ".git" && (info.IsDir() || info.Mode().IsRegular())) || (name != ".git" && info.Mode().IsRegular())) {
		return dir, true
	}
	return "", false
}

func pathWithin(root, candidate string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(candidate))
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func readModulePath(goModPath string) (string, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module path not found in %s", goModPath)
}
