// Package doctor analyses a Ginger project and reports best-practice violations.
package doctor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type checkState int

const (
	checkPass checkState = iota
	checkFail
	checkSkip
)

type checkResult struct {
	state  checkState
	reason string
}

type check struct {
	label string
	fn    func() checkResult
}

type evaluatedCheck struct {
	label  string
	result checkResult
}

var (
	lookPath    = exec.LookPath
	execCommand = exec.Command
)

func defaultChecks() []check {
	return []check{
		{"valid project structure", checkStructure},
		{"go.mod present", checkGoMod},
		{"configs/app.yaml present when applicable", checkConfig},
		{"DevOps Dockerfile present when applicable", checkDockerfile},
		{"health check endpoint when applicable", checkHealthEndpoint},
		{"graceful shutdown configured when applicable", checkGracefulShutdown},
		{"tests present", checkTests},
		{"go vet passes", checkGoVet},
		{"lint (golangci-lint)", checkLint},
	}
}

func evaluateChecks(checks []check) ([]evaluatedCheck, bool) {
	results := make([]evaluatedCheck, 0, len(checks))
	allOK := true

	for _, c := range checks {
		result := c.fn()
		results = append(results, evaluatedCheck{label: c.label, result: result})
		if result.state == checkFail {
			allOK = false
		}
	}

	return results, allOK
}

func printResults(results []evaluatedCheck) {
	for _, r := range results {
		switch r.result.state {
		case checkPass:
			fmt.Printf("  ✓ %s\n", r.label)
		case checkSkip:
			if r.result.reason != "" {
				fmt.Printf("  - %s (%s)\n", r.label, r.result.reason)
			} else {
				fmt.Printf("  - %s\n", r.label)
			}
		default:
			fmt.Printf("  ■ %s\n", r.label)
		}
	}
}

// Run executes all checks, prints a diagnostic report, and returns true when all mandatory checks pass.
func Run() bool {
	fmt.Print("\n🩺 Ginger Doctor\n\n")

	results, allOK := evaluateChecks(defaultChecks())
	printResults(results)

	fmt.Println()
	if allOK {
		fmt.Println("✓ All checks passed. Your project looks healthy!")
	} else {
		fmt.Println("■ Some checks failed. Review the items above.")
	}
	fmt.Println()
	return allOK
}

func checkStructure() checkResult {
	for _, d := range []string{"cmd"} {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			return checkResult{state: checkFail}
		}
	}
	return checkResult{state: checkPass}
}

func checkGoMod() checkResult {
	_, err := os.Stat("go.mod")
	if err == nil {
		return checkResult{state: checkPass}
	}
	return checkResult{state: checkFail}
}

func checkConfig() checkResult {
	if !needsConfig() {
		return checkResult{state: checkPass}
	}
	_, err := os.Stat(filepath.Join("configs", "app.yaml"))
	if err == nil {
		return checkResult{state: checkPass}
	}
	return checkResult{state: checkFail}
}

func checkDockerfile() checkResult {
	if !needsDockerfile() {
		return checkResult{state: checkPass}
	}
	_, err := os.Stat(filepath.Join("devops", "docker", "Dockerfile"))
	if err == nil {
		return checkResult{state: checkPass}
	}
	return checkResult{state: checkFail}
}

func checkHealthEndpoint() checkResult {
	if !isHTTPProject() {
		return checkResult{state: checkPass}
	}
	if grepInDir(".", "/health") ||
		grepInDir(".", "health.New") ||
		grepInDir(".", "gingerapp.New") {
		return checkResult{state: checkPass}
	}
	return checkResult{state: checkFail}
}

func checkGracefulShutdown() checkResult {
	if !isHTTPProject() {
		return checkResult{state: checkPass}
	}
	if grepInDir(".", "Shutdown") ||
		grepInDir(".", "SIGTERM") ||
		grepInDir(".", "gingerapp.New") ||
		grepInDir(".", "app.New") {
		return checkResult{state: checkPass}
	}
	return checkResult{state: checkFail}
}

func checkTests() checkResult {
	if hasFileWithSuffix(".", "_test.go") {
		return checkResult{state: checkPass}
	}
	return checkResult{state: checkFail}
}

// checkGoVet runs go vet ./... and reports whether it passes.
func checkGoVet() checkResult {
	cmd := execCommand("go", "vet", "./...")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if cmd.Run() == nil {
		return checkResult{state: checkPass}
	}
	return checkResult{state: checkFail}
}

func checkLint() checkResult {
	if _, err := lookPath("golangci-lint"); err != nil {
		return checkResult{state: checkSkip, reason: "golangci-lint not installed"}
	}
	cmd := execCommand("golangci-lint", "run", "--fast", "--timeout", "30s")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if cmd.Run() == nil {
		return checkResult{state: checkPass}
	}
	return checkResult{state: checkFail}
}

// grepInDir reports whether pattern appears in any .go file under dir.
func grepInDir(dir, pattern string) bool {
	needle := []byte(pattern)
	found := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error { //nolint:errcheck
		if err != nil || found {
			return nil
		}
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git") {
			return filepath.SkipDir
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			data, err := os.ReadFile(path)
			if err == nil && bytes.Contains(data, needle) {
				found = true
			}
		}
		return nil
	})
	return found
}

// hasFileWithSuffix checks if any file with the given suffix exists under dir.
func hasFileWithSuffix(dir, suffix string) bool {
	found := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error { //nolint:errcheck
		if err != nil || found {
			return nil
		}
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git") {
			return filepath.SkipDir
		}
		if !info.IsDir() && len(path) >= len(suffix) && path[len(path)-len(suffix):] == suffix {
			found = true
		}
		return nil
	})
	return found
}

func needsConfig() bool {
	if _, err := os.Stat(filepath.Join("configs", "app.yaml")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join("internal", "config", "config.go")); err == nil {
		return true
	}
	return false
}

func needsDockerfile() bool {
	if _, err := os.Stat(filepath.Join("devops", "docker", "Dockerfile")); err == nil {
		return true
	}
	if _, err := os.Stat("devops"); err == nil {
		return true
	}
	return false
}

func isHTTPProject() bool {
	if _, err := os.Stat(filepath.Join("internal", "api", "handlers")); err == nil {
		return true
	}
	return grepInDir(".", "gingerapp.New")
}
