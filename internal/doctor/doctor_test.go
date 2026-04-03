package doctor

import "testing"

func TestCheckLintSkipsWhenGolangCILintIsMissing(t *testing.T) {
	originalLookPath := lookPath
	lookPath = func(file string) (string, error) {
		return "", errNotFound(file)
	}
	defer func() {
		lookPath = originalLookPath
	}()

	result := checkLint()
	if result.state != checkSkip {
		t.Fatalf("expected checkSkip, got %v", result.state)
	}
}

type notFoundError string

func (e notFoundError) Error() string { return string(e) }

func errNotFound(file string) error {
	return notFoundError(file + " not found")
}
