package sse

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRejectsNonFlusher(t *testing.T) {
	// Plain wrapper — implements ResponseWriter only (no Flusher).
	w := plainResponseWriter{httptest.NewRecorder()}
	if _, err := New(w); !errors.Is(err, ErrStreamingUnsupported) {
		t.Fatalf("expected ErrStreamingUnsupported, got %v", err)
	}
}

type plainResponseWriter struct {
	http.ResponseWriter
}

func TestSendEscapesMultiLinePayload(t *testing.T) {
	w := newFlushRecorder()
	s, err := New(w)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := s.Send(Event{Data: "line1\nline2\r\nline3"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "data: line1\ndata: line2\ndata: line3\n\n") {
		t.Fatalf("multi-line payload not split into data: lines:\n%q", body)
	}
	if strings.Contains(body, "data: line1\nline2") {
		t.Fatal("protocol framing broken by payload newline")
	}
}

func TestSendStripsBreaksFromIDTypeAndComment(t *testing.T) {
	w := newFlushRecorder()
	s, _ := New(w)
	if err := s.Send(Event{ID: "abc\nevent: forged", Type: "chat\r\ndata: x", Data: "ok"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	s.SendComment("keep\nalive\r\ninjected")
	want := "id: abc event: forged\nevent: chat data: x\ndata: ok\n\n: keep alive injected\n\n"
	if got := w.Body.String(); got != want {
		t.Fatalf("output differs from sanitized expectation:\n got: %q\nwant: %q", got, want)
	}
}

func TestSendJSONEncodesStructs(t *testing.T) {
	w := newFlushRecorder()
	s, _ := New(w)
	if err := s.Send(Event{Data: map[string]any{"a": 1}}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !strings.Contains(w.Body.String(), `data: {"a":1}`) {
		t.Fatalf("json encoding missing: %q", w.Body.String())
	}
}
