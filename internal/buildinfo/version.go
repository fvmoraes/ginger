package buildinfo

import (
	_ "embed"
	"regexp"
	"runtime/debug"
	"strings"
)

// FallbackVersion is used when no stable version metadata is available.
const FallbackVersion = "dev"

//go:embed version.txt
var embeddedVersion string

// releaseVersion can be injected via -ldflags for official release binaries.
var releaseVersion string

var pseudoVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+-(0\.)?\d{14}-[0-9a-f]{12}$`)

var readBuildInfo = debug.ReadBuildInfo

// Version returns the best stable Ginger version available for user-facing output
// and scaffolding. It prefers build-time release metadata, then Go build info,
// then the embedded repository release version.
func Version() string {
	buildVersion := ""
	info, ok := readBuildInfo()
	if ok && info != nil {
		buildVersion = info.Main.Version
	}

	return selectVersion(releaseVersion, buildVersion, embeddedVersion)
}

// ResolveVersion normalizes a raw module version string into a stable semantic version.
func ResolveVersion(raw string) string {
	v := strings.TrimPrefix(strings.TrimSpace(raw), "v")
	if v == "" || v == "(devel)" || strings.Contains(v, "+dirty") || pseudoVersionPattern.MatchString(v) {
		return FallbackVersion
	}
	return v
}

func selectVersion(candidates ...string) string {
	for _, candidate := range candidates {
		if version := ResolveVersion(candidate); version != FallbackVersion {
			return version
		}
	}

	return FallbackVersion
}
