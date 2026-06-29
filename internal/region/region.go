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

// ReplaceRegion replaces the content of a managed region. If the region does not
// exist, an error is returned (use InsertRegion for new regions).
func ReplaceRegion(content, name, replacement string) (string, error) {
	r := FindRegion(content, name)
	if r == nil {
		return "", fmt.Errorf("region %q not found", name)
	}
	return content[:r.Start] + "\n" + replacement + "\n" + content[r.End:], nil
}

// InsertRegion inserts a new managed region into the content. Returns an error
// if the region already exists.
func InsertRegion(content, name, insertion string) (string, error) {
	if HasRegion(content, name) {
		return "", fmt.Errorf("region %q already exists — use ReplaceRegion", name)
	}
	marker := fmt.Sprintf("// ginger:begin %s\n%s\n// ginger:end %s\n", name, insertion, name)
	return content + "\n" + marker, nil
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
