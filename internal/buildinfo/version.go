package buildinfo

import (
	"regexp"
	"runtime/debug"
	"strings"
)

// FallbackVersion is used when the binary does not carry a stable semantic version.
const FallbackVersion = "1.3.3"

var pseudoVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+-(0\.)?\d{14}-[0-9a-f]{12}$`)

// Version returns the best stable Ginger version available for user-facing output
// and scaffolding. It strips a leading "v" and falls back when the build carries
// a pseudo-version or a dirty/devel marker.
func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return FallbackVersion
	}

	return ResolveVersion(info.Main.Version)
}

// ResolveVersion normalizes a raw module version string into a stable semantic version.
func ResolveVersion(raw string) string {
	v := strings.TrimPrefix(raw, "v")
	if v == "" || v == "(devel)" || strings.Contains(v, "+dirty") || pseudoVersionPattern.MatchString(v) {
		return FallbackVersion
	}
	return v
}
