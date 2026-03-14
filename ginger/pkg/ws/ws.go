// Package ws provides a lightweight WebSocket helper built on top of
// golang.org/x/net/websocket (stdlib-compatible, no CGO).
//
// For bidirectional real-time communication between server and frontend clients.
// For one-way server→client streams, use pkg/sse instead.
//
// Usage:
//
//	func chatHandler(w http.ResponseWriter, r *http.Request) {
//	    ws.Handle(w, r, func(conn *ws.Conn) {
//	        for {
//	            var msg ws.Message
//	            if err := conn.Recv(&msg); err != nil {
//	                return
//	            }
//	            conn.Send(ws.Message{Type: "echo", Data: msg.Data})
//	        }
//	    })
//	}
package ws

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
)

// ErrUpgradeFailed is returned when the HTTP→WebSocket upgrade fails.
var ErrUpgradeFailed = errors.New("ws: upgrade failed")

// Message is the standard JSON envelope for WebSocket messages.
//
//	{ "type": "chat", "data": { ... } }
type Message struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

// Conn wraps the underlying connection with typed send/receive helpers.
type Conn struct {
	mu     sync.Mutex
	enc    *json.Encoder
	dec    *json.Decoder
	closer interface{ Close() error }
}

// Send encodes v as JSON and writes it to the connection.
// Safe for concurrent use.
func (c *Conn) Send(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.enc.Encode(v)
}

// Recv decodes the next JSON message from the connection into v.
func (c *Conn) Recv(v any) error {
	return c.dec.Decode(v)
}

// Close closes the underlying connection.
func (c *Conn) Close() error {
	return c.closer.Close()
}

// Handler is a function that handles a WebSocket connection.
type Handler func(conn *Conn)

// Handle upgrades the HTTP connection to WebSocket and calls fn.
// It uses the standard HTTP hijack mechanism via golang.org/x/net/websocket
// semantics but implemented over net/http for zero extra dependencies.
//
// The connection is closed automatically when fn returns.
func Handle(w http.ResponseWriter, r *http.Request, fn Handler) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, ErrUpgradeFailed.Error(), http.StatusInternalServerError)
		return
	}

	// Perform the WebSocket handshake manually (RFC 6455).
	if !isWebSocketUpgrade(r) {
		http.Error(w, "ws: not a websocket upgrade request", http.StatusBadRequest)
		return
	}

	key := r.Header.Get("Sec-Websocket-Key")
	if key == "" {
		http.Error(w, "ws: missing Sec-WebSocket-Key", http.StatusBadRequest)
		return
	}

	accept := computeAccept(key)

	w.Header().Set("Upgrade", "websocket")
	w.Header().Set("Connection", "Upgrade")
	w.Header().Set("Sec-WebSocket-Accept", accept)
	w.WriteHeader(http.StatusSwitchingProtocols)

	netConn, buf, err := hj.Hijack()
	if err != nil {
		return
	}

	conn := &Conn{
		enc:    json.NewEncoder(newFrameWriter(netConn)),
		dec:    json.NewDecoder(newFrameReader(buf, netConn)),
		closer: netConn,
	}
	defer conn.Close()
	fn(conn)
}

// isWebSocketUpgrade checks the required upgrade headers.
func isWebSocketUpgrade(r *http.Request) bool {
	return r.Method == http.MethodGet &&
		r.Header.Get("Upgrade") == "websocket" &&
		r.Header.Get("Connection") == "Upgrade"
}
