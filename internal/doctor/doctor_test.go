package doctor

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestEvaluateChecksReturnsFalseWhenAnyCheckFails(t *testing.T) {
	results, sum, allOK := evaluateChecks([]check{
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
	if sum.passed != 1 || sum.skipped != 1 || sum.failed != 1 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}

func TestEvaluateChecksReturnsTrueWhenNoCheckFails(t *testing.T) {
	_, sum, allOK := evaluateChecks([]check{
		{label: "ok", fn: func() checkResult { return checkResult{state: checkPass} }},
		{label: "skip", fn: func() checkResult { return checkResult{state: checkSkip, reason: "not installed"} }},
	})

	if !allOK {
		t.Fatal("expected allOK to be true when no check fails")
	}
	if sum.passed != 1 || sum.skipped != 1 || sum.failed != 0 {
		t.Fatalf("unexpected summary: %+v", sum)
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

func TestRunReportsExecutedChecksWhenSomeAreSkipped(t *testing.T) {
	originalChecks := defaultChecks
	defaultChecks = func() []check {
		return []check{
			{label: "ok", fn: func() checkResult { return checkResult{state: checkPass} }},
			{label: "skip", fn: func() checkResult { return checkResult{state: checkSkip, reason: "not installed"} }},
		}
	}
	defer func() { defaultChecks = originalChecks }()

	output := captureStdout(t, func() {
		if !Run() {
			t.Fatal("expected Run to return true")
		}
	})

	if !strings.Contains(output, "All executed checks passed. Passed: 1, Skipped: 1.") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestRunReportsCountsWhenChecksFail(t *testing.T) {
	originalChecks := defaultChecks
	defaultChecks = func() []check {
		return []check{
			{label: "ok", fn: func() checkResult { return checkResult{state: checkPass} }},
			{label: "fail", fn: func() checkResult { return checkResult{state: checkFail} }},
		}
	}
	defer func() { defaultChecks = originalChecks }()

	output := captureStdout(t, func() {
		if Run() {
			t.Fatal("expected Run to return false")
		}
	})

	if !strings.Contains(output, "Some checks failed. Passed: 1, Failed: 1, Skipped: 0.") {
		t.Fatalf("unexpected output: %s", output)
	}
}

type notFoundError string

func (e notFoundError) Error() string { return string(e) }

func errNotFound(file string) error {
	return notFoundError(file + " not found")
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe returned error: %v", err)
	}
	os.Stdout = w

	outputCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outputCh <- buf.String()
	}()

	fn()

	_ = w.Close()
	os.Stdout = originalStdout
	output := <-outputCh
	_ = r.Close()
	return output
}
