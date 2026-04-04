package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

var gitOutput = func(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

type projectBuildMetadata struct {
	Version string
	Commit  string
	Date    string
}

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

	goArgs := append([]string{"run"}, cliBuildFlagArgs()...)
	goArgs = append(goArgs, cmdDir)
	goArgs = append(goArgs, args...)

	cmd := exec.Command("go", goArgs...)
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

	goArgs := append([]string{"build"}, cliBuildFlagArgs()...)
	goArgs = append(goArgs, "-o", output, cmdDir)

	cmd := exec.Command("go", goArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Built: %s\n", output)
}

func cliBuildFlagArgs() []string {
	modulePath, ok := detectCLIProjectModule()
	if !ok {
		return nil
	}

	meta := resolveProjectBuildMetadata()
	ldflags := strings.Join([]string{
		fmt.Sprintf("-X %s/internal/commands.version=%s", modulePath, meta.Version),
		fmt.Sprintf("-X %s/internal/commands.commit=%s", modulePath, meta.Commit),
		fmt.Sprintf("-X %s/internal/commands.date=%s", modulePath, meta.Date),
	}, " ")

	return []string{"-ldflags", ldflags}
}

func detectCLIProjectModule() (string, bool) {
	if _, err := os.Stat(filepath.Join("internal", "commands", "version.go")); err != nil {
		return "", false
	}

	modulePath, err := readModulePath("go.mod")
	if err != nil || modulePath == "" {
		return "", false
	}

	return modulePath, true
}

func resolveProjectBuildMetadata() projectBuildMetadata {
	meta := projectBuildMetadata{
		Version: "dev",
		Commit:  "local",
		Date:    time.Now().UTC().Format(time.RFC3339),
	}

	if tag, err := gitOutput("describe", "--tags", "--exact-match"); err == nil && tag != "" {
		meta.Version = strings.TrimPrefix(tag, "v")
	}
	if commit, err := gitOutput("rev-parse", "--short", "HEAD"); err == nil && commit != "" {
		meta.Commit = commit
	}

	return meta
}

func readModulePath(goModPath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("module path not found in %s", goModPath)
}
