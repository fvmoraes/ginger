package doctor

import "testing"

func TestEvaluateChecksReturnsFalseWhenAnyCheckFails(t *testing.T) {
	results, allOK := evaluateChecks([]check{
		{label: "ok", fn: func() checkResult { return checkResult{state: checkPass} }},
		{label: "skip", fn: func() checkResult { return checkResult{state: checkSkip, reason: "not installed"} }},
		{label: "fail", fn: func() checkResult { return checkResult{state: checkFail} }},
	})

	if allOK {
		t.Fatal("expected allOK to be false when a check fails")
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 evaluated checks, got %d", len(results))
	}
}

func TestEvaluateChecksReturnsTrueWhenNoCheckFails(t *testing.T) {
	_, allOK := evaluateChecks([]check{
		{label: "ok", fn: func() checkResult { return checkResult{state: checkPass} }},
		{label: "skip", fn: func() checkResult { return checkResult{state: checkSkip, reason: "not installed"} }},
	})

	if !allOK {
		t.Fatal("expected allOK to be true when no check fails")
	}
}

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
