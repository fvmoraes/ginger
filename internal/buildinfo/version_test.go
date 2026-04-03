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
