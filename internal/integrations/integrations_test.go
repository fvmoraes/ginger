package integrations

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddRemovesCreatedFileWhenDependencyInstallFails(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()

	originalRegistry, ok := registry["testdep"]
	if ok {
		defer func() { registry["testdep"] = originalRegistry }()
	} else {
		defer delete(registry, "testdep")
	}

	registry["testdep"] = integration{
		name: "testdep",
		pkg:  "example.com/failing-dependency",
		file: filepath.Join("platform", "testdep", "client.go"),
		tmpl: "package testdep\n",
	}

	originalExecCommand := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() {
		execCommand = originalExecCommand
	}()

	err = Add("testdep")
	if err == nil {
		t.Fatalf("expected Add to fail when go get fails")
	}

	if _, statErr := os.Stat(filepath.Join("platform", "testdep", "client.go")); !os.IsNotExist(statErr) {
		t.Fatalf("expected generated file to be removed, stat err=%v", statErr)
	}
}

func TestRealtimeTemplatesUseHandlersPackage(t *testing.T) {
	for name, tmpl := range map[string]string{
		"sse":       sseTmpl,
		"websocket": wsTmpl,
	} {
		if !strings.Contains(tmpl, "package handlers") {
			t.Fatalf("%s template should declare package handlers", name)
		}
	}
}
