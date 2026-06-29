// Package manifest records which files and regions Ginger may manage.
package manifest

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fvmoraes/ginger/internal/plan"
	"gopkg.in/yaml.v3"
)

// Entry describes ownership of a file or named regions in that file.
type Entry struct {
	Path     string   `yaml:"path"`
	FullFile bool     `yaml:"full_file,omitempty"`
	Regions  []string `yaml:"regions,omitempty"`
}

// Manifest is stored at .ginger/manifest.yaml.
type Manifest struct {
	Managed []Entry `yaml:"managed"`
}

// Load reads a manifest. A missing manifest is returned as an empty value.
func Load(root string) (*Manifest, error) {
	path := filepath.Join(root, ".ginger", "manifest.yaml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Manifest{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &m, nil
}

// ManagesFullFile reports whether Ginger owns the whole relative path.
func (m *Manifest) ManagesFullFile(path string) bool {
	path = filepath.ToSlash(filepath.Clean(path))
	for _, entry := range m.Managed {
		if filepath.ToSlash(filepath.Clean(entry.Path)) == path && entry.FullFile {
			return true
		}
	}
	return false
}

// ManagesRegion reports whether Ginger owns a named region in relative path.
func (m *Manifest) ManagesRegion(path, region string) bool {
	path = filepath.ToSlash(filepath.Clean(path))
	for _, entry := range m.Managed {
		if filepath.ToSlash(filepath.Clean(entry.Path)) != path {
			continue
		}
		for _, current := range entry.Regions {
			if current == region {
				return true
			}
		}
	}
	return false
}

// Add merges ownership without discarding existing entries.
func (m *Manifest) Add(entry Entry) {
	entry.Path = filepath.ToSlash(filepath.Clean(entry.Path))
	for i := range m.Managed {
		if filepath.ToSlash(filepath.Clean(m.Managed[i].Path)) != entry.Path {
			continue
		}
		m.Managed[i].FullFile = m.Managed[i].FullFile || entry.FullFile
		for _, region := range entry.Regions {
			if !contains(m.Managed[i].Regions, region) {
				m.Managed[i].Regions = append(m.Managed[i].Regions, region)
			}
		}
		sort.Strings(m.Managed[i].Regions)
		return
	}
	m.Managed = append(m.Managed, entry)
	sort.Slice(m.Managed, func(i, j int) bool { return m.Managed[i].Path < m.Managed[j].Path })
}

// PlanUpdate adds a safe manifest update to p.
func PlanUpdate(p *plan.Plan, entries ...Entry) error {
	m, err := Load(p.ProjectRoot)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		m.Add(entry)
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	path := filepath.Join(p.ProjectRoot, ".ginger", "manifest.yaml")
	if _, err := os.Stat(path); err == nil {
		p.AddModify(path, data, true)
	} else if os.IsNotExist(err) {
		p.AddCreate(path, data, false)
	} else {
		return fmt.Errorf("inspect manifest: %w", err)
	}
	return nil
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
