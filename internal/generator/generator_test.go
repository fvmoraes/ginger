package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCRUDDoesNotGenerateTests(t *testing.T) {
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

	goMod := "module example.com/test\n\ngo 1.25\n"
	if err := os.WriteFile("go.mod", []byte(goMod), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if err := CRUD("user"); err != nil {
		t.Fatalf("CRUD returned error: %v", err)
	}

	expectedFiles := []string{
		filepath.Join("internal", "models", "user.go"),
		filepath.Join("internal", "api", "handlers", "user_handler.go"),
		filepath.Join("internal", "api", "services", "user_service.go"),
		filepath.Join("internal", "api", "repositories", "user_repository.go"),
	}
	for _, path := range expectedFiles {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected file %s to exist: %v", path, err)
		}
	}

	unexpectedFiles := []string{
		filepath.Join("internal", "api", "handlers", "user_handler_test.go"),
		filepath.Join("internal", "api", "services", "user_service_test.go"),
		filepath.Join("internal", "api", "repositories", "user_repository_test.go"),
	}
	for _, path := range unexpectedFiles {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected test file %s to be absent, got err=%v", path, err)
		}
	}
}
