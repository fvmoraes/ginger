// Package region provides helpers for managed code regions.
// Ginger can safely modify content inside // ginger:begin / // ginger:end blocks
// without overwriting user code outside those blocks.
package region

import (
	"fmt"
	"regexp"
	"strings"
)

var regionPattern = regexp.MustCompile(`(?m)^[\t ]*(?://|#|<!--)\s*ginger:(begin|end)\s+(\S+)\s*(?:-->)?\s*$`)

// Region represents a managed region in a file.
type Region struct {
	Name    string
	Content string
	Start   int
	End     int
}

// marker is a parsed region marker with its position.
type marker struct {
	kind     string // "begin" | "end"
	name     string
	line     int // 1-based, for error messages
	startIdx int // byte offset of the line start
	endIdx   int // byte offset of the line end (exclusive)
}

// parseMarkers extracts and positions every region marker in the content.
func parseMarkers(content string) []marker {
	var out []marker
	for _, loc := range regionPattern.FindAllStringSubmatchIndex(content, -1) {
		kind := content[loc[2]:loc[3]]
		name := content[loc[4]:loc[5]]
		lineEnd := strings.Index(content[loc[0]:], "\n")
		lineNo := 1 + strings.Count(content[:loc[0]], "\n")
		end := len(content)
		if lineEnd >= 0 {
			end = loc[0] + lineEnd
		}
		out = append(out, marker{kind: kind, name: name, line: lineNo, startIdx: loc[0], endIdx: end})
	}
	return out
}

// ValidateRegions (GIN-015) verifies that every marker pairs correctly.
// It reports: unclosed begins, orphan ends, duplicated regions, nested and
// interleaved regions — all of which previously produced silent no-ops.
func ValidateRegions(content string) error {
	ms := parseMarkers(content)
	seenOpen := map[string]marker{} // region name → its begin marker
	closed := map[string]int{}      // region name → times closed
	lastBegin := marker{}
	haveLastBegin := false

	for _, m := range ms {
		switch m.kind {
		case "begin":
			if haveLastBegin {
				// A begin while another region is open: nested or interleaved.
				return fmt.Errorf("region %q (line %d): nested or interleaved region %q (line %d) is not supported — close %q first",
					lastBegin.name, lastBegin.line, m.name, m.line, lastBegin.name)
			}
			if prev, dup := seenOpen[m.name]; dup {
				return fmt.Errorf("region %q duplicated (first at line %d, duplicate at line %d) — update it with ReplaceRegion instead",
					m.name, prev.line, m.line)
			}
			seenOpen[m.name] = m
			lastBegin = m
			haveLastBegin = true
		case "end":
			if !haveLastBegin {
				return fmt.Errorf("orphan end marker for region %q (line %d) — no begin is open", m.name, m.line)
			}
			if m.name != lastBegin.name {
				return fmt.Errorf("region %q opened at line %d is closed by mismatched end %q (line %d)",
					lastBegin.name, lastBegin.line, m.name, m.line)
			}
			closed[m.name]++
			haveLastBegin = false
		}
	}

	if haveLastBegin {
		return fmt.Errorf("region %q (line %d) has no end marker", lastBegin.name, lastBegin.line)
	}
	return nil
}

// FindRegion finds a managed region by name in the given content.
// Returns nil if the region is not found.
func FindRegion(content, name string) *Region {
	matches := regionPattern.FindAllStringSubmatchIndex(content, -1)
	for i := 0; i < len(matches)-1; i++ {
		m := matches[i]
		nextM := matches[i+1]
		kind := content[m[2]:m[3]]
		regionName := content[m[4]:m[5]]
		nextKind := content[nextM[2]:nextM[3]]
		nextName := content[nextM[4]:nextM[5]]
		if kind == "begin" && regionName == name && nextKind == "end" && nextName == name {
			startLineEnd := strings.Index(content[m[1]:], "\n")
			if startLineEnd == -1 {
				startLineEnd = 0
			}
			start := m[1] + startLineEnd + 1
			end := nextM[0]
			if end < start {
				end = start
			}
			return &Region{
				Name:    name,
				Content: strings.Trim(content[start:end], "\r\n"),
				Start:   start,
				End:     end,
			}
		}
	}
	return nil
}

// HasRegion reports whether the content contains a managed region with the given name.
func HasRegion(content, name string) bool {
	return FindRegion(content, name) != nil
}

// ReplaceRegion replaces the content of a managed region. If the region does
// not exist, or the file's markers are malformed, an error is returned
// (GIN-015 — malformed markers used to fail silently).
func ReplaceRegion(content, name, replacement string) (string, error) {
	if err := ValidateRegions(content); err != nil {
		return "", err
	}
	r := FindRegion(content, name)
	if r == nil {
		return "", fmt.Errorf("region %q not found", name)
	}
	return content[:r.Start] + "\n" + replacement + "\n" + content[r.End:], nil
}

// InsertRegion inserts a new managed region into the content. Returns an error
// if the region already exists or the existing markers are malformed (GIN-015).
// Line endings match the file's dominant style (CRLF preserved — GIN-015).
func InsertRegion(content, name, insertion string) (string, error) {
	if err := ValidateRegions(content); err != nil {
		return "", err
	}
	if HasRegion(content, name) {
		return "", fmt.Errorf("region %q already exists — use ReplaceRegion", name)
	}
	eol := "\n"
	if strings.Count(content, "\r\n") > strings.Count(content, "\n")/2 {
		eol = "\r\n"
	}
	marker := fmt.Sprintf("// ginger:begin %s%s%s%s// ginger:end %s%s", name, eol, insertion, eol, name, eol)
	if content == "" {
		return marker, nil
	}
	return strings.TrimRight(content, eol) + eol + eol + marker, nil
}

// ExtractRegions returns all managed regions in the content.
func ExtractRegions(content string) []Region {
	var regions []Region
	matches := regionPattern.FindAllStringSubmatchIndex(content, -1)
	for i := 0; i < len(matches)-1; i++ {
		m := matches[i]
		nextM := matches[i+1]
		kind := content[m[2]:m[3]]
		regionName := content[m[4]:m[5]]
		nextKind := content[nextM[2]:nextM[3]]
		nextName := content[nextM[4]:nextM[5]]
		if kind == "begin" && nextKind == "end" && nextName == regionName {
			startLineEnd := strings.Index(content[m[1]:], "\n")
			if startLineEnd == -1 {
				startLineEnd = 0
			}
			start := m[1] + startLineEnd + 1
			end := nextM[0]
			if end < start {
				end = start
			}
			regions = append(regions, Region{
				Name:    regionName,
				Content: strings.Trim(content[start:end], "\r\n"),
				Start:   start,
				End:     end,
			})
		}
	}
	return regions
}
