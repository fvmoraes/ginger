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

func TestAddUpdatesDockerComposeForMessagingIntegration(t *testing.T) {
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

	if err := os.MkdirAll(filepath.Join("devops", "docker"), 0755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	compose := `version: "3.9"
services:
  app:
    build:
      context: ../..
      dockerfile: devops/docker/Dockerfile
    environment:
      APP_ENV: development
`
	if err := os.WriteFile(filepath.Join("devops", "docker", "docker-compose.yml"), []byte(compose), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	originalExecCommand := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("true")
	}
	defer func() {
		execCommand = originalExecCommand
	}()

	if err := Add("rabbitmq"); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("devops", "docker", "docker-compose.yml"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	content := string(data)
	for _, want := range []string{"rabbitmq:", "rabbitmq:3-management-alpine", "RABBITMQ_URL", "depends_on"} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected compose to contain %q, got:\n%s", want, content)
		}
	}
}

func TestAddSkipsComposeUpdateWhenComposeFileDoesNotExist(t *testing.T) {
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

	originalExecCommand := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("true")
	}
	defer func() {
		execCommand = originalExecCommand
	}()

	if err := Add("postgres"); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join("platform", "database", "postgres.go")); err != nil {
		t.Fatalf("expected generated integration file to exist: %v", err)
	}
}

func TestAddMongoDBGeneratesValidTemplateOutput(t *testing.T) {
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

	originalExecCommand := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("true")
	}
	defer func() {
		execCommand = originalExecCommand
	}()

	if err := Add("mongodb"); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("platform", "nosql", "mongo.go"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(data), `bson.D{bson.E{Key: "ping", Value: 1}}`) {
		t.Fatalf("expected escaped bson command in generated file, got:\n%s", string(data))
	}
	if !strings.Contains(string(data), `go.mongodb.org/mongo-driver/v2/mongo`) {
		t.Fatalf("expected mongo template to use the v2 driver import path, got:\n%s", string(data))
	}
}

func TestAddSQLiteTemplateIncludesTimeImport(t *testing.T) {
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

	originalExecCommand := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("true")
	}
	defer func() {
		execCommand = originalExecCommand
	}()

	if err := Add("sqlite"); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("platform", "database", "sqlite.go"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(data), `"time"`) {
		t.Fatalf("expected sqlite template to import time, got:\n%s", string(data))
	}
}

func TestMessagingTemplatesUseTransportSpecificHelperNames(t *testing.T) {
	if strings.Contains(kafkaTmpl, "func Publish(") {
		t.Fatalf("kafka template should not declare a generic Publish function")
	}
	if !strings.Contains(kafkaTmpl, "func PublishKafka(") {
		t.Fatalf("expected kafka template to declare PublishKafka")
	}

	if strings.Contains(natsTmpl, "func Publish(") {
		t.Fatalf("nats template should not declare a generic Publish function")
	}
	if !strings.Contains(natsTmpl, "func PublishNATS(") {
		t.Fatalf("expected nats template to declare PublishNATS")
	}
	if !strings.Contains(natsTmpl, "func SubscribeNATS(") {
		t.Fatalf("expected nats template to declare SubscribeNATS")
	}
}

func TestRegistryUsesCurrentMongoAndPubSubModules(t *testing.T) {
	if got := registry["mongodb"].pkg; got != "go.mongodb.org/mongo-driver/v2/mongo" {
		t.Fatalf("expected mongodb integration to use v2 driver, got %q", got)
	}
	if got := registry["pubsub"].pkg; got != "cloud.google.com/go/pubsub/v2" {
		t.Fatalf("expected pubsub integration to use v2 module, got %q", got)
	}
}

func TestAddSwaggerRegistersRoutesInServiceRouter(t *testing.T) {
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

	if err := os.MkdirAll(filepath.Join("internal", "api"), 0755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	routerSource := `package api

import (
	"example/internal/api/middlewares"
	"github.com/fvmoraes/ginger/pkg/router"
)

func Register(r *router.Router) {
	v1 := r.Group("/api/v1", middlewares.RequestID)
	_ = v1
}
`
	if err := os.WriteFile(filepath.Join("internal", "api", "router.go"), []byte(routerSource), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if err := Add("swagger"); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("internal", "api", "router.go"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(data), "registerSwaggerRoutes(r)") {
		t.Fatalf("expected swagger integration to patch router registration, got:\n%s", string(data))
	}

	swaggerFile, err := os.ReadFile(filepath.Join("internal", "api", "swagger.go"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(swaggerFile), "package api") {
		t.Fatalf("expected swagger integration file to be generated in package api, got:\n%s", string(swaggerFile))
	}
}
