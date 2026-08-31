package cli

import (
	"fmt"
	"github.com/fvmoraes/ginger/internal/project"
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

// projectRoot resolves the project root from the current directory.
// GIN-009: run/build work from any subdirectory, not just the repo root.
// Falls back to the CWD when no project root can be detected (keeps the
// historical behavior outside Ginger projects).
func projectRoot() string {
	if root, err := resolveRoot(); err == nil && root != "" {
		return root
	}
	return "."
}

// detectCmdDir finds the only subdirectory of cmd/ that contains a main.go.
// Paths are relative to the project root (go run/build must run from there).
func detectCmdDir(root string) (string, error) {
	entries, err := os.ReadDir(filepath.Join(root, "cmd"))
	if err != nil {
		return "", fmt.Errorf("no cmd/ directory found — are you inside a Ginger project?")
	}

	var matches []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, "cmd", e.Name(), "main.go")); err == nil {
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
	root := projectRoot()
	cmdDir, err := detectCmdDir(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	goArgs := append([]string{"run"}, cliBuildFlagArgs(root)...)
	goArgs = append(goArgs, cmdDir)
	goArgs = append(goArgs, args...)

	cmd := exec.Command("go", goArgs...)
	cmd.Dir = root
	setupProcessGroup(cmd) // GIN-029: signal reaches go run AND the app
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
		_ = signalChild(cmd, sig)
		<-waitCh
		// GIN-029: Ctrl-C must not look like success for scripts — exit
		// 128+N for the signal we received (the child's own exit code is an
		// implementation detail of `go run`, which reports 1 even when its
		// grandchild died by signal).
		if s, ok := sig.(syscall.Signal); ok {
			os.Exit(128 + int(s))
		}
		os.Exit(1)
	}
}

func runBuild(args []string) {
	root := projectRoot()
	cmdDir, err := detectCmdDir(root)
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

	goArgs := append([]string{"build"}, cliBuildFlagArgs(root)...)
	goArgs = append(goArgs, "-o", output, cmdDir)

	cmd := exec.Command("go", goArgs...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Built: %s\n", output)
}

func cliBuildFlagArgs(root string) []string {
	modulePath, ok := detectCLIProjectModule(root)
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

func detectCLIProjectModule(root string) (string, bool) {
	if _, err := os.Stat(filepath.Join(root, "internal", "commands", "version.go")); err != nil {
		return "", false
	}

	modulePath, err := project.ReadModulePath(filepath.Join(root, "go.mod"))
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
