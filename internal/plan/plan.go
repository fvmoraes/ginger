// Package plan provides safe plan/apply semantics for file generation.
package plan

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ChangeType classifies what a planned change would do.
type ChangeType string

const (
	ChangeCreate ChangeType = "create"
	ChangeModify ChangeType = "modify"
	ChangeSkip   ChangeType = "skip"
	ChangeError  ChangeType = "error"
)

// PlannedChange describes a single file operation.
type PlannedChange struct {
	Type    ChangeType
	Path    string
	Reason  string
	Content []byte

	expectedExists bool
	expectedHash   [sha256.Size]byte
	expectedMode   os.FileMode
}

// Plan holds a collection of planned changes.
type Plan struct {
	ProjectRoot       string
	CreateMissingDirs bool
	Changes           []PlannedChange
	Warnings          []string
	Errors            []string
}

// New creates a plan rooted at projectRoot. Generated paths must remain below
// this directory, including after resolving existing symlink ancestors.
func New(projectRoot string) *Plan {
	return &Plan{ProjectRoot: projectRoot, CreateMissingDirs: true}
}

// HasErrors returns true if the plan contains blocking errors.
func (p *Plan) HasErrors() bool { return len(p.Errors) > 0 }

// HasChanges returns true if the plan has any create or modify changes.
func (p *Plan) HasChanges() bool {
	for _, c := range p.Changes {
		if c.Type == ChangeCreate || c.Type == ChangeModify {
			return true
		}
	}
	return false
}

// AddCreate plans creation of path. An existing file is never overwritten
// unless overwrite is explicitly true.
func (p *Plan) AddCreate(path string, content []byte, overwrite bool) {
	path, err := p.safePath(path)
	if err != nil {
		p.addPathError(path, err)
		return
	}
	info, err := os.Lstat(path)
	switch {
	case err == nil:
		if !info.Mode().IsRegular() {
			p.addPathError(path, errors.New("target exists and is not a regular file"))
			return
		}
		if !overwrite {
			p.Changes = append(p.Changes, PlannedChange{
				Type: ChangeSkip, Path: path,
				Reason: "file already exists; user content is preserved",
			})
			return
		}
		p.addExisting(ChangeModify, path, content, "explicit overwrite requested", info)
	case os.IsNotExist(err):
		if parentErr := p.validateParent(path); parentErr != nil {
			p.addPathError(path, parentErr)
			return
		}
		p.Changes = append(p.Changes, PlannedChange{
			Type: ChangeCreate, Path: path, Content: append([]byte(nil), content...),
		})
	default:
		p.addPathError(path, fmt.Errorf("inspect target: %w", err))
	}
}

// AddModify plans replacement of an existing file. Callers should pass
// overwrite only for an explicitly forced change or a safely managed file.
func (p *Plan) AddModify(path string, content []byte, overwrite bool) {
	path, err := p.safePath(path)
	if err != nil {
		p.addPathError(path, err)
		return
	}
	info, err := os.Lstat(path)
	switch {
	case os.IsNotExist(err):
		if parentErr := p.validateParent(path); parentErr != nil {
			p.addPathError(path, parentErr)
			return
		}
		p.Changes = append(p.Changes, PlannedChange{
			Type: ChangeCreate, Path: path, Content: append([]byte(nil), content...),
		})
	case err != nil:
		p.addPathError(path, fmt.Errorf("inspect target: %w", err))
	case !info.Mode().IsRegular():
		p.addPathError(path, errors.New("target exists and is not a regular file"))
	case !overwrite:
		p.Changes = append(p.Changes, PlannedChange{
			Type: ChangeSkip, Path: path,
			Reason: "file is not managed and overwrite is disabled",
		})
	default:
		p.addExisting(ChangeModify, path, content, "safe managed update", info)
	}
}

func (p *Plan) addExisting(changeType ChangeType, path string, content []byte, reason string, info os.FileInfo) {
	data, err := os.ReadFile(path)
	if err != nil {
		p.addPathError(path, fmt.Errorf("read current content: %w", err))
		return
	}
	if bytes.Equal(data, content) {
		p.Changes = append(p.Changes, PlannedChange{
			Type: ChangeSkip, Path: path, Reason: "already up to date",
		})
		return
	}
	p.Changes = append(p.Changes, PlannedChange{
		Type: changeType, Path: path, Reason: reason,
		Content: append([]byte(nil), content...), expectedExists: true,
		expectedHash: sha256.Sum256(data), expectedMode: info.Mode().Perm(),
	})
}

func (p *Plan) addPathError(path string, err error) {
	msg := fmt.Sprintf("unsafe change %q: %v", path, err)
	p.Errors = append(p.Errors, msg)
	p.Changes = append(p.Changes, PlannedChange{Type: ChangeError, Path: path, Reason: err.Error()})
}

func (p *Plan) validateParent(path string) error {
	if p.CreateMissingDirs {
		return nil
	}
	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("parent directory is missing and rules.create_missing_dirs is false: %w", err)
	}
	if !info.IsDir() {
		return errors.New("target parent is not a directory")
	}
	return nil
}

// AddWarning adds a non-blocking warning.
func (p *Plan) AddWarning(msg string) { p.Warnings = append(p.Warnings, msg) }

// AddError adds a blocking error.
func (p *Plan) AddError(msg string) { p.Errors = append(p.Errors, msg) }

// Render prints the plan in a human-readable format.
func (p *Plan) Render() {
	fmt.Println()
	fmt.Println("Plan:")
	fmt.Printf("  Project root: %s\n", p.ProjectRoot)
	if len(p.Changes) == 0 {
		fmt.Println("  (no changes)")
	}
	for _, c := range p.Changes {
		s := map[ChangeType]string{
			ChangeCreate: "+", ChangeModify: "~", ChangeSkip: "○", ChangeError: "✗",
		}[c.Type]
		fmt.Printf("  %s %-7s %s", s, c.Type, relPath(p.ProjectRoot, c.Path))
		if c.Reason != "" {
			fmt.Printf("  (%s)", c.Reason)
		}
		fmt.Println()
	}
	for _, w := range p.Warnings {
		fmt.Printf("  ⚠ %s\n", w)
	}
	for _, e := range p.Errors {
		fmt.Printf("  ✗ ERROR: %s\n", e)
	}
	fmt.Println()
}

// Apply preflights the complete plan and then writes each file. Creates use
// O_EXCL and modifications are written through a temporary file. A plan is
// rejected if a target changed after planning.
func (p *Plan) Apply() error {
	if p.HasErrors() {
		return errors.New("cannot apply a plan with errors")
	}
	for i := range p.Changes {
		if err := p.preflight(&p.Changes[i]); err != nil {
			return err
		}
	}
	for _, c := range p.Changes {
		switch c.Type {
		case ChangeCreate:
			if err := p.writeCreate(c); err != nil {
				return err
			}
			fmt.Printf("  ✓ create %s\n", relPath(p.ProjectRoot, c.Path))
		case ChangeModify:
			if err := p.writeModify(c); err != nil {
				return err
			}
			fmt.Printf("  ~ modify %s\n", relPath(p.ProjectRoot, c.Path))
		case ChangeSkip:
			fmt.Printf("  ○ skipped %s (%s)\n", relPath(p.ProjectRoot, c.Path), c.Reason)
		}
	}
	return nil
}

// Snapshot captures the on-disk state of every writable change so a failed
// apply+post-apply sequence can be undone (GIN-005 — N4: restores creates,
// modifies and the manifest, not just created files).
type Snapshot struct {
	root    string
	entries []snapshotEntry
}

type snapshotEntry struct {
	path    string
	existed bool
	content []byte
	mode    os.FileMode
}

// Snapshot reads the current state of the files this plan would change.
// Call BEFORE Apply().
func (p *Plan) Snapshot() (*Snapshot, error) {
	s := &Snapshot{root: p.ProjectRoot}
	for _, c := range p.Changes {
		if c.Type != ChangeCreate && c.Type != ChangeModify {
			continue
		}
		e := snapshotEntry{path: c.Path, mode: 0o644}
		if data, err := os.ReadFile(c.Path); err == nil {
			e.existed = true
			e.content = data
			if info, statErr := os.Stat(c.Path); statErr == nil {
				e.mode = info.Mode()
			}
		}
		s.entries = append(s.entries, e)
	}
	return s, nil
}

// Restore undoes the plan's writes: created files are removed and modified
// files are rewritten with their pre-apply content.
func (s *Snapshot) Restore() error {
	for _, e := range s.entries {
		if !e.existed {
			if err := os.Remove(e.path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("rollback: remove created %s: %w", e.path, err)
			}
			continue
		}
		if err := os.WriteFile(e.path, e.content, e.mode); err != nil {
			return fmt.Errorf("rollback: restore %s: %w", e.path, err)
		}
	}
	return nil
}

func (p *Plan) preflight(c *PlannedChange) error {
	if c.Type != ChangeCreate && c.Type != ChangeModify {
		return nil
	}
	path, err := p.safePath(c.Path)
	if err != nil {
		return fmt.Errorf("preflight %s: %w", c.Path, err)
	}
	c.Path = path
	parent := filepath.Dir(path)
	if _, err := os.Stat(parent); err != nil {
		if !os.IsNotExist(err) || !p.CreateMissingDirs {
			return fmt.Errorf("preflight %s: parent directory is unavailable: %w", path, err)
		}
	}
	info, statErr := os.Lstat(path)
	if c.Type == ChangeCreate {
		if statErr == nil {
			return fmt.Errorf("stale plan: %s was created after planning", path)
		}
		if !os.IsNotExist(statErr) {
			return fmt.Errorf("preflight %s: %w", path, statErr)
		}
		return nil
	}
	if statErr != nil {
		return fmt.Errorf("stale plan: %s is no longer available: %w", path, statErr)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("preflight %s: target is not a regular file", path)
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("preflight %s: %w", path, err)
	}
	if !c.expectedExists || sha256.Sum256(current) != c.expectedHash {
		return fmt.Errorf("stale plan: %s changed after planning", path)
	}
	return nil
}

func (p *Plan) writeCreate(c PlannedChange) error {
	if err := p.ensureParent(c.Path); err != nil {
		return err
	}
	f, err := os.OpenFile(c.Path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", c.Path, err)
	}
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = os.Remove(c.Path)
		}
	}()
	if _, err := f.Write(c.Content); err != nil {
		return fmt.Errorf("write %s: %w", c.Path, err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync %s: %w", c.Path, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close %s: %w", c.Path, err)
	}
	ok = true
	return nil
}

func (p *Plan) writeModify(c PlannedChange) error {
	current, err := os.ReadFile(c.Path)
	if err != nil || sha256.Sum256(current) != c.expectedHash {
		return fmt.Errorf("stale plan: %s changed before write", c.Path)
	}
	tmp, err := os.CreateTemp(filepath.Dir(c.Path), ".ginger-write-*")
	if err != nil {
		return fmt.Errorf("create temporary file for %s: %w", c.Path, err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	mode := c.expectedMode
	if mode == 0 {
		mode = 0o644
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := io.Copy(tmp, bytes.NewReader(c.Content)); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temporary file for %s: %w", c.Path, err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, c.Path); err != nil {
		return fmt.Errorf("replace %s: %w", c.Path, err)
	}
	return nil
}

func (p *Plan) ensureParent(path string) error {
	parent := filepath.Dir(path)
	if !p.CreateMissingDirs {
		return nil
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create parent for %s: %w", path, err)
	}
	_, err := p.safePath(path)
	return err
}

func (p *Plan) safePath(path string) (string, error) {
	root, err := filepath.Abs(p.ProjectRoot)
	if err != nil {
		return path, err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(root); resolveErr == nil {
		root = resolved
	}
	p.ProjectRoot = filepath.Clean(root)
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	path = filepath.Clean(path)
	if !within(root, path) || path == root {
		return path, errors.New("target is outside the project root")
	}

	ancestor := path
	for {
		if _, statErr := os.Lstat(ancestor); statErr == nil {
			resolved, resolveErr := filepath.EvalSymlinks(ancestor)
			if resolveErr != nil {
				return path, fmt.Errorf("resolve symlink ancestor: %w", resolveErr)
			}
			if !within(root, resolved) {
				return path, errors.New("target traverses a symlink outside the project root")
			}
			break
		} else if !os.IsNotExist(statErr) {
			return path, statErr
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return path, errors.New("could not find an existing project ancestor")
		}
		ancestor = parent
	}
	return path, nil
}

func within(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func relPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return rel
	}
	return path
}
