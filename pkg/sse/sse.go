// Package sse provides a Server-Sent Events (SSE) writer for streaming
// real-time updates to frontend clients over a plain HTTP connection.
//
// SSE is ideal for one-way server→client streams (live feeds, notifications,
// progress updates). For bidirectional communication use pkg/ws instead.
//
// Usage:
//
//	func streamHandler(w http.ResponseWriter, r *http.Request) {
//	    stream, err := sse.New(w)
//	    if err != nil {
//	        http.Error(w, err.Error(), http.StatusInternalServerError)
//	        return
//	    }
//	    for {
//	        select {
//	        case <-r.Context().Done():
//	            return
//	        case event := <-eventCh:
//	            stream.Send(sse.Event{Data: event})
//	        }
//	    }
//	}
package sse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrStreamingUnsupported is returned when the ResponseWriter does not
// implement http.Flusher, which is required for SSE.
var ErrStreamingUnsupported = errors.New("sse: streaming unsupported by this ResponseWriter")

// Event represents a single SSE message.
type Event struct {
	// ID is the optional event ID (allows clients to resume after reconnect).
	ID string
	// Type is the optional event type (default: "message").
	Type string
	// Data is the event payload. Structs are JSON-encoded automatically.
	Data any
	// Retry instructs the client to wait N milliseconds before reconnecting.
	Retry int
}

// Stream wraps an http.ResponseWriter for SSE output.
type Stream struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// New prepares w for SSE and returns a Stream.
// Flushers are resolved through Unwrap chains (GIN-013) so SSE works behind
// wrapping middlewares. Returns ErrStreamingUnsupported if none is reachable.
func New(w http.ResponseWriter) (*Stream, error) {
	f := flusherOf(w)
	if f == nil {
		return nil, ErrStreamingUnsupported
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no") // disable nginx buffering
	w.WriteHeader(http.StatusOK)
	f.Flush()
	return &Stream{w: w, flusher: f}, nil
}

// flusherOf finds the http.Flusher, unwrapping ResponseWriter wrappers.
func flusherOf(w http.ResponseWriter) http.Flusher {
	if f, ok := w.(http.Flusher); ok {
		return f
	}
	if u, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
		return flusherOf(u.Unwrap())
	}
	return nil
}

// Send writes a single SSE event to the client and flushes immediately.
//
// Injection-safe (GIN-012): multi-line payloads are split into multiple
// `data:` lines (per the SSE spec) and CR/LF are stripped from ID, Type and
// comments so user-controlled content can never forge protocol fields.
func (s *Stream) Send(e Event) error {
	if e.Retry > 0 {
		if _, err := fmt.Fprintf(s.w, "retry: %d\n", e.Retry); err != nil {
			return err
		}
	}
	if e.ID != "" {
		if _, err := fmt.Fprintf(s.w, "id: %s\n", stripBreaks(e.ID)); err != nil {
			return err
		}
	}
	eventType := e.Type
	if eventType == "" {
		eventType = "message"
	}
	if _, err := fmt.Fprintf(s.w, "event: %s\n", stripBreaks(eventType)); err != nil {
		return err
	}

	// Encode Data as JSON if it's not already a string.
	var payload string
	switch v := e.Data.(type) {
	case string:
		payload = v
	case []byte:
		payload = string(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("sse: marshal data: %w", err)
		}
		payload = string(b)
	}

	for _, line := range payloadLines(payload) {
		if _, err := fmt.Fprintf(s.w, "data: %s\n", line); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprint(s.w, "\n"); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// SendComment writes an SSE comment line (ignored by clients, useful as keepalive).
func (s *Stream) SendComment(comment string) {
	_, _ = fmt.Fprintf(s.w, ": %s\n\n", stripBreaks(comment))
	s.flusher.Flush()
}

// stripBreaks removes CR/LF sequences that would terminate an SSE field early.
func stripBreaks(s string) string {
	r := strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ")
	return r.Replace(s)
}

// payloadLines splits a payload into SSE data lines, normalizing CRLF.
func payloadLines(payload string) []string {
	normalized := strings.ReplaceAll(payload, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return strings.Split(normalized, "\n")
}
