package sse

import (
	"net/http"
	"net/http/httptest"
)

// flushRecorder implements http.Flusher on top of httptest.ResponseRecorder
// so New() accepts it in tests.
type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed int
}

func newFlushRecorder() *flushRecorder {
	return &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
}

func (f *flushRecorder) Flush() { f.flushed++ }

var (
	_ http.Flusher        = (*flushRecorder)(nil)
	_ http.ResponseWriter = (*flushRecorder)(nil)
)
