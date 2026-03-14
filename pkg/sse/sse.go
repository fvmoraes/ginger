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
// Returns ErrStreamingUnsupported if w does not implement http.Flusher.
func New(w http.ResponseWriter) (*Stream, error) {
	f, ok := w.(http.Flusher)
	if !ok {
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

// Send writes a single SSE event to the client and flushes immediately.
func (s *Stream) Send(e Event) error {
	if e.Retry > 0 {
		fmt.Fprintf(s.w, "retry: %d\n", e.Retry)
	}
	if e.ID != "" {
		fmt.Fprintf(s.w, "id: %s\n", e.ID)
	}
	eventType := e.Type
	if eventType == "" {
		eventType = "message"
	}
	fmt.Fprintf(s.w, "event: %s\n", eventType)

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

	fmt.Fprintf(s.w, "data: %s\n\n", payload)
	s.flusher.Flush()
	return nil
}

// SendComment writes an SSE comment line (ignored by clients, useful as keepalive).
func (s *Stream) SendComment(comment string) {
	fmt.Fprintf(s.w, ": %s\n\n", comment)
	s.flusher.Flush()
}
