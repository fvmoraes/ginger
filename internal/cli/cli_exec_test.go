package cli

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Fase 4 (GIN-020/GIN-029): golden tests da CLI via binário compilado —
// exit codes reais e saídas congeláveis. Execute com -update para
// re-congelar os goldens (diff exige revisão).

var updateGolden = flag.Bool("update-golden", false, "rewrite CLI golden files")
var gingerBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "ginger-cli-*")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	gingerBin = filepath.Join(dir, "ginger")
	build := exec.Command("go", "build", "-o", gingerBin, "github.com/fvmoraes/ginger/cmd/ginger")
	build.Dir = repoDir()
	if out, err := build.CombinedOutput(); err != nil {
		panic("build ginger: " + string(out))
	}
	os.Exit(m.Run())
}

func repoDir() string {
	wd, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			panic("repo root not found")
		}
		wd = parent
	}
}

// goldenCLI runs the binary and compares (or freezes) stdout+stderr+exit code.
func goldenCLI(t *testing.T, name string, dir string, stdin *string, timeout time.Duration, args ...string) {
	t.Helper()
	cmd := exec.Command(gingerBin, args...)
	cmd.Dir = dir
	if stdin != nil {
		cmd.Stdin = strings.NewReader(*stdin)
	}
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	exitCode := 0
	if ee, ok := err.(*exec.ExitError); ok {
		exitCode = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run %v: %v", args, err)
	}
	// Normaliza o root absoluto (TempDir muda a cada run — N3).
	normalized := strings.ReplaceAll(out.String(), dir, "<ROOT>")
	got := "exit: " + itoa(exitCode) + "\n" + normalized

	goldenPath := filepath.Join("testdata", "cli_golden", name+".golden")
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden missing (run with -update-golden): %v", err)
	}
	if got != string(want) {
		t.Fatalf("CLI output changed (intentional? re-run with -update-golden and review):\n--- got ---\n%s\n--- want ---\n%s", got, string(want))
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}

func TestGoldenHelp(t *testing.T) {
	goldenCLI(t, "help", t.TempDir(), nil, 10*time.Second, "help")
}

func TestGoldenVersion(t *testing.T) {
	goldenCLI(t, "version", t.TempDir(), nil, 10*time.Second, "version")
}

func TestGoldenNewExtraArgRejected(t *testing.T) {
	goldenCLI(t, "new-extra-arg", t.TempDir(), nil, 10*time.Second, "new", "foo", "bar")
}

func TestGoldenUnknownCommand(t *testing.T) {
	goldenCLI(t, "unknown-command", t.TempDir(), nil, 10*time.Second, "definitely-not-a-command")
}

// GIN-029: SIGINT durante `ginger run` propaga 128+N.
func TestRunPropagatesSIGINTExitCode(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "cmd", "app")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module sleepy\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	main := "package main\n\nimport (\n\t\"fmt\"\n\t\"time\"\n)\n\nfunc main() {\n\tfmt.Println(\"ready\")\n\tfor {\n\t\ttime.Sleep(time.Second)\n\t}\n}\n"
	if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(main), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(gingerBin, "run")
	cmd.Dir = dir
	// File-based stdout: polling reads are race-free (strings.Builder via
	// exec copy goroutine would race with the polling read).
	outFile, err := os.CreateTemp("", "ginger-run-out-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(outFile.Name()) }()
	cmd.Stdout = outFile
	cmd.Stderr = outFile
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	readOut := func() string {
		data, _ := os.ReadFile(outFile.Name())
		return string(data)
	}

	// Wait until the app printed "ready" (app compiled and running).
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) && !strings.Contains(readOut(), "ready") {
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(readOut(), "ready") {
		_ = cmd.Process.Kill()
		t.Fatalf("app never became ready: %s", readOut())
	}

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("signal: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("ginger run did not exit after SIGINT")
	}

	exitCode := cmd.ProcessState.ExitCode()
	// 130 = 128+SIGINT (GIN-029): the child (go run) is interrupted, ginger
	// must propagate instead of exiting 0.
	if exitCode == 0 {
		t.Fatalf("giner exited 0 after SIGINT (GIN-029 regression)")
	}
	if exitCode < 128 {
		t.Fatalf("expected 128+N propagation, got %d", exitCode)
	}
}

// GIN-020: add --plan --json (CI machine-readable contract).
func TestGoldenAddPlanJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module demo\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	composeDir := filepath.Join(dir, "devops", "docker")
	if err := os.MkdirAll(composeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	compose := "version: \"3.9\"\nservices:\n  app:\n    image: demo\n"
	if err := os.WriteFile(filepath.Join(composeDir, "docker-compose.yml"), []byte(compose), 0o644); err != nil {
		t.Fatal(err)
	}
	goldenCLI(t, "add-plan-json", dir, nil, 30*time.Second, "add", "redis", "--plan", "--json")
}
