// Package doctor analyses a Ginger project and reports best-practice violations.
package doctor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type check struct {
	label string
	fn    func() bool
}

// Run executes all checks and prints a diagnostic report.
func Run() {
	fmt.Print("\n🩺 Ginger Doctor\n\n")

	checks := []check{
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

	allOK := true
	for _, c := range checks {
		if c.fn() {
			fmt.Printf("  ✓ %s\n", c.label)
		} else {
			fmt.Printf("  ■ %s\n", c.label)
			allOK = false
		}
	}

	fmt.Println()
	if allOK {
		fmt.Println("✓ All checks passed. Your project looks healthy!")
	} else {
		fmt.Println("■ Some checks failed. Review the items above.")
	}
	fmt.Println()
}

func checkStructure() bool {
	for _, d := range []string{"cmd"} {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func checkGoMod() bool {
	_, err := os.Stat("go.mod")
	return err == nil
}

func checkConfig() bool {
	if !needsConfig() {
		return true
	}
	_, err := os.Stat(filepath.Join("configs", "app.yaml"))
	return err == nil
}

func checkDockerfile() bool {
	if !needsDockerfile() {
		return true
	}
	_, err := os.Stat(filepath.Join("devops", "docker", "Dockerfile"))
	return err == nil
}

func checkHealthEndpoint() bool {
	if !isHTTPProject() {
		return true
	}
	return grepInDir(".", "/health") ||
		grepInDir(".", "health.New") ||
		grepInDir(".", "gingerapp.New")
}

func checkGracefulShutdown() bool {
	if !isHTTPProject() {
		return true
	}
	return grepInDir(".", "Shutdown") ||
		grepInDir(".", "SIGTERM") ||
		grepInDir(".", "gingerapp.New") ||
		grepInDir(".", "app.New")
}

func checkTests() bool {
	return hasFileWithSuffix(".", "_test.go")
}

// checkGoVet runs go vet ./... and reports whether it passes.
func checkGoVet() bool {
	cmd := exec.Command("go", "vet", "./...")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func checkLint() bool {
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return false
	}
	cmd := exec.Command("golangci-lint", "run", "--fast", "--timeout", "30s")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
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
