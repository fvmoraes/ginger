package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// detectCmdDir finds the first subdirectory of cmd/ that contains a main.go.
func detectCmdDir() (string, error) {
	entries, err := os.ReadDir("cmd")
	if err != nil {
		return "", fmt.Errorf("no cmd/ directory found — are you inside a Ginger project?")
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join("cmd", e.Name(), "main.go")); err == nil {
			return "./" + filepath.Join("cmd", e.Name()), nil
		}
	}
	return "", fmt.Errorf("no main.go found inside cmd/ — are you inside a Ginger project?")
}

func runRun(args []string) {
	cmdDir, err := detectCmdDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cmd := exec.Command("go", append([]string{"run", cmdDir}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
		os.Exit(1)
	}
}

func runBuild(args []string) {
	cmdDir, err := detectCmdDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Derive binary name from the cmd subdirectory name
	binName := filepath.Base(cmdDir)
	output := filepath.Join("./bin", binName)
	if len(args) > 0 {
		output = args[0]
	}

	cmd := exec.Command("go", "build", "-o", output, cmdDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Built: %s\n", output)
}
