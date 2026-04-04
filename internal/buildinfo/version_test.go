package buildinfo

import "testing"

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "stable release", raw: "v1.3.1", want: "1.3.1"},
		{name: "stable release without v", raw: "1.3.1", want: "1.3.1"},
		{name: "stable release with whitespace", raw: "  v1.3.4\n", want: "1.3.4"},
		{name: "devel falls back", raw: "(devel)", want: FallbackVersion},
		{name: "empty falls back", raw: "", want: FallbackVersion},
		{name: "dirty release falls back", raw: "v1.3.1+dirty", want: FallbackVersion},
		{name: "pseudo version falls back", raw: "v0.0.0-20260403120000-abcdef123456", want: FallbackVersion},
		{name: "pseudo version with dash zero falls back", raw: "v1.3.2-0.20260403120000-abcdef123456", want: FallbackVersion},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveVersion(tc.raw); got != tc.want {
				t.Fatalf("ResolveVersion(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestSelectVersionPrefersFirstStableCandidate(t *testing.T) {
	got := selectVersion("v1.3.5", "v1.3.4", "1.3.3")
	if got != "1.3.5" {
		t.Fatalf("selectVersion() = %q, want %q", got, "1.3.5")
	}
}

func TestSelectVersionFallsBackToEmbeddedStableVersion(t *testing.T) {
	got := selectVersion("", "(devel)", "1.3.4\n")
	if got != "1.3.4" {
		t.Fatalf("selectVersion() = %q, want %q", got, "1.3.4")
	}
}

func TestSelectVersionFallsBackToDefaultWhenNoStableCandidateExists(t *testing.T) {
	got := selectVersion("", "(devel)", "v0.0.0-20260403120000-abcdef123456")
	if got != FallbackVersion {
		t.Fatalf("selectVersion() = %q, want %q", got, FallbackVersion)
	}
}
