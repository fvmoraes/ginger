// Package generator — worker handler generator for --worker projects.
package generator

import (
	"path/filepath"
)

const workerHandlerGenTmpl = generatedGoFileHeader + `package worker

import "context"

// {{.NameTitle}}Handler processes messages of type {{.NameTitle}}.
type {{.NameTitle}}Handler struct{}

// New{{.NameTitle}}Handler returns a new {{.NameTitle}}Handler.
func New{{.NameTitle}}Handler() *{{.NameTitle}}Handler { return &{{.NameTitle}}Handler{} }

// Handle processes a single message payload.
// Return a non-nil error to signal that the message should be retried.
func (h *{{.NameTitle}}Handler) Handle(ctx context.Context, msg []byte) error {
	// PT-BR: Implemente a lógica de processamento aqui.
	// EN: Implement processing logic here.
	_ = msg
	return nil
}
`

const workerHandlerTestGenTmpl = generatedGoFileHeader + `package worker

import (
	"context"
	"testing"
)

func Test{{.NameTitle}}Handler_Handle(t *testing.T) {
	h := New{{.NameTitle}}Handler()
	if err := h.Handle(context.Background(), []byte("test")); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}
}
`

// WorkerHandler generates internal/worker/<name>_handler.go and its test.
// Use this for --worker projects via: ginger generate handler <name>
func WorkerHandler(name string) error {
	data := newData(name)

	if err := generate(
		filepath.Join("internal", "worker", data.FileName+"_handler.go"),
		workerHandlerGenTmpl,
		data,
	); err != nil {
		return err
	}

	return generate(
		filepath.Join("internal", "worker", data.FileName+"_handler_test.go"),
		workerHandlerTestGenTmpl,
		data,
	)
}
