// Package doctor analyzes a Ginger project and reports best-practice violations.
package doctor

import (
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
	fmt.Println("\n🩺 Ginger Doctor\n")

	checks := []check{
		{"valid project structure", checkStructure},
		{"go.mod present", checkGoMod},
		{"configs/app.yaml present", checkConfig},
		{"Dockerfile present", checkDockerfile},
		{"health check endpoint", checkHealthEndpoint},
		{"graceful shutdown configured", checkGracefulShutdown},
		{"tests present", checkTests},
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
	required := []string{
		"cmd",
		"internal",
		"configs",
	}
	for _, d := range required {
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
	_, err := os.Stat(filepath.Join("configs", "app.yaml"))
	return err == nil
}

func checkDockerfile() bool {
	_, err := os.Stat("Dockerfile")
	return err == nil
}

func checkHealthEndpoint() bool {
	// Check for /health route or health package import
	return grepInDir(".", "/health") ||
		grepInDir(".", `health.New`) ||
		grepInDir(".", `gingerapp.New`) // app.New registers /health automatically
}

func checkGracefulShutdown() bool {
	// Check for explicit shutdown or use of gingerapp.New (which includes graceful shutdown)
	return grepInDir(".", "Shutdown") ||
		grepInDir(".", "SIGTERM") ||
		grepInDir(".", "gingerapp.New") ||
		grepInDir(".", `app.New`)
}

func checkTests() bool {
	matches, _ := filepath.Glob("**/*_test.go")
	if len(matches) > 0 {
		return true
	}
	// walk manually since ** doesn't work in filepath.Glob on all platforms
	found := false
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && len(path) > 8 && path[len(path)-8:] == "_test.go" {
			found = true
		}
		return nil
	})
	return found
}

func checkLint() bool {
	_, err := exec.LookPath("golangci-lint")
	if err != nil {
		return false // not installed, skip
	}
	cmd := exec.Command("golangci-lint", "run", "--fast", "--timeout", "30s")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// grepInDir checks if a string appears in any .go file under dir.
func grepInDir(dir, pattern string) bool {
	found := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return nil
		}
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git") {
			return filepath.SkipDir
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			data, err := os.ReadFile(path)
			if err == nil && contains(data, pattern) {
				found = true
			}
		}
		return nil
	})
	return found
}

func contains(data []byte, s string) bool {
	return len(data) > 0 && indexBytes(data, []byte(s)) >= 0
}

func indexBytes(haystack, needle []byte) int {
	if len(needle) == 0 {
		return 0
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if string(haystack[i:i+len(needle)]) == string(needle) {
			return i
		}
	}
	return -1
}
