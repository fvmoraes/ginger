package cli

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
)

// detectCmdDir finds the only subdirectory of cmd/ that contains a main.go.
func detectCmdDir() (string, error) {
	entries, err := os.ReadDir("cmd")
	if err != nil {
		return "", fmt.Errorf("no cmd/ directory found — are you inside a Ginger project?")
	}

	var matches []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join("cmd", e.Name(), "main.go")); err == nil {
			matches = append(matches, "./"+filepath.Join("cmd", e.Name()))
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no main.go found inside cmd/ — are you inside a Ginger project?")
	case 1:
		return matches[0], nil
	default:
		sort.Strings(matches)
		return "", fmt.Errorf("multiple app entrypoints found: %v", matches)
	}
}

func runRun(args []string) {
	cmdDir, err := detectCmdDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	cmd := exec.Command("go", append([]string{"run", cmdDir}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
		os.Exit(1)
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	select {
	case err := <-waitCh:
		if err != nil {
			fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
			os.Exit(1)
		}
	case sig := <-sigCh:
		if cmd.Process != nil {
			_ = cmd.Process.Signal(sig)
		}
		<-waitCh
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
