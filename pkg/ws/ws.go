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
//
// Hardening (GIN-003): Handle applies safe defaults — 64 KiB payload limit,
// same-origin check, per-message read deadline and masked-frame enforcement.
// Use HandleWithOptions to override (e.g. allow a dev frontend on another port):
//
//	ws.HandleWithOptions(w, r, ws.Options{
//	    CheckOrigin: func(r *http.Request) bool { return true },
//	}, fn)
package ws

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"
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

// Options configures connection hardening. Zero values adopt the safe
// defaults documented on each field.
type Options struct {
	// MaxMessageSize caps the payload of an incoming frame in bytes.
	// 0 → 64 KiB (DefaultMaxMessageSize). Negative → unlimited (not recommended).
	MaxMessageSize int64
	// CheckOrigin validates the Origin header. nil → same-origin policy
	// (Origin host must match the request Host; requests without Origin —
	// non-browser clients — are allowed).
	CheckOrigin func(r *http.Request) bool
	// ReadTimeout is the per-message read deadline. 0 → 60s.
	// Negative → no deadline (not recommended; idle connections never expire).
	ReadTimeout time.Duration
}

// DefaultMaxMessageSize is the default per-frame payload cap (64 KiB).
const DefaultMaxMessageSize int64 = 64 << 10

// DefaultReadTimeout is the default per-message read deadline.
const DefaultReadTimeout = 60 * time.Second

// effective resolves the zero-value defaults.
func (o Options) effective() Options {
	if o.MaxMessageSize == 0 {
		o.MaxMessageSize = DefaultMaxMessageSize
	}
	if o.CheckOrigin == nil {
		o.CheckOrigin = sameOrigin
	}
	if o.ReadTimeout == 0 {
		o.ReadTimeout = DefaultReadTimeout
	}
	return o
}

// sameOrigin implements the default same-origin policy (GIN-003).
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Non-browser clients (CLI tools, tests) may omit Origin.
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return u.Host == r.Host
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

// Handle upgrades the HTTP connection to WebSocket and calls fn with the
// safe defaults (see Options).
func Handle(w http.ResponseWriter, r *http.Request, fn Handler) {
	HandleWithOptions(w, r, Options{}, fn)
}

// HandleWithOptions upgrades the HTTP connection to WebSocket with explicit
// hardening options and calls fn. The connection is closed when fn returns.
func HandleWithOptions(w http.ResponseWriter, r *http.Request, opts Options, fn Handler) {
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

	effective := opts.effective()
	if !effective.CheckOrigin(r) {
		http.Error(w, "ws: origin not allowed", http.StatusForbidden)
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

	limits := frameLimits{maxSize: effective.MaxMessageSize, readDeadline: effective.ReadTimeout}
	if effective.MaxMessageSize < 0 {
		limits.maxSize = 0 // unlimited
	}
	if effective.ReadTimeout < 0 {
		limits.readDeadline = 0 // no deadline
	}

	conn := &Conn{
		enc:    json.NewEncoder(newFrameWriter(netConn)),
		dec:    json.NewDecoder(newFrameReader(buf, netConn, limits)),
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
