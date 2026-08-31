package ws

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ── test helpers ─────────────────────────────────────────────────────────────

// startWSServer runs an HTTP server whose handler upgrades via HandleWithOptions.
func startWSServer(t *testing.T, opts Options, handler Handler) (url string, close func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWithOptions(w, r, opts, handler)
	}))
	return "ws" + strings.TrimPrefix(srv.URL, "http"), srv.Close
}

// dialWS performs the RFC 6455 handshake over raw TCP and returns the conn.
func dialWS(t *testing.T, wsURL, origin string) (net.Conn, *bufio.Reader) {
	t.Helper()
	tcpURL := strings.Replace(wsURL, "ws://", "", 1)
	conn, err := net.DialTimeout("tcp", tcpURL, 5*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	host := tcpURL
	req := "GET / HTTP/1.1\r\nHost: " + host + "\r\n" +
		"Upgrade: websocket\r\nConnection: Upgrade\r\n" +
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n" +
		"Sec-WebSocket-Version: 13\r\n"
	if origin != "" {
		req += "Origin: " + origin + "\r\n"
	}
	req += "\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		t.Fatalf("read handshake response: %v", err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101 upgrade, got %d", resp.StatusCode)
	}
	return conn, br
}

// clientSendFrame writes a masked client frame (RFC 6455 §5.3).
func clientSendFrame(t *testing.T, conn net.Conn, opcode byte, payload []byte) {
	t.Helper()
	var header []byte
	first := byte(0x80 | opcode)
	n := len(payload)
	switch {
	case n <= 125:
		header = []byte{first, 0x80 | byte(n)}
	case n <= 65535:
		header = []byte{first, 0x80 | 126, 0, 0}
		binary.BigEndian.PutUint16(header[2:], uint16(n))
	default:
		header = []byte{first, 0x80 | 127, 0, 0, 0, 0, 0, 0, 0, 0}
		binary.BigEndian.PutUint64(header[2:], uint64(n))
	}
	mask := []byte{0x11, 0x22, 0x33, 0x44}
	masked := make([]byte, n)
	for i := range payload {
		masked[i] = payload[i] ^ mask[i%4]
	}
	if _, err := conn.Write(append(append(header, mask...), masked...)); err != nil {
		t.Fatalf("send frame: %v", err)
	}
}

// clientReadFrame reads one server frame (unmasked).
func clientReadFrame(t *testing.T, br *bufio.Reader) (opcode byte, payload []byte, err error) {
	t.Helper()
	var hdr [2]byte
	if _, err := io.ReadFull(br, hdr[:]); err != nil {
		return 0, nil, err
	}
	opcode = hdr[0] & 0x0F
	n := int64(hdr[1] & 0x7F)
	switch hdr[1] & 0x7F {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(br, ext[:]); err != nil {
			return 0, nil, err
		}
		n = int64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(br, ext[:]); err != nil {
			return 0, nil, err
		}
		n = int64(binary.BigEndian.Uint64(ext[:]))
	}
	payload = make([]byte, n)
	if _, err := io.ReadFull(br, payload); err != nil {
		return 0, nil, err
	}
	return opcode, payload, nil
}

// ── tests ────────────────────────────────────────────────────────────────────

func TestEchoWorks(t *testing.T) {
	url, closeSrv := startWSServer(t, Options{}, func(conn *Conn) {
		for {
			var msg Message
			if err := conn.Recv(&msg); err != nil {
				return
			}
			_ = conn.Send(Message{Type: "echo", Data: msg.Data})
		}
	})
	defer closeSrv()

	conn, br := dialWS(t, url, "")
	clientSendFrame(t, conn, 0x1, []byte(`{"type":"chat","data":"hello"}`))
	opcode, payload, err := clientReadFrame(t, br)
	if err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if opcode != 0x1 {
		t.Fatalf("expected text frame, got opcode %#x", opcode)
	}
	var msg Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		t.Fatalf("decode echo: %v", err)
	}
	if msg.Type != "echo" || msg.Data != "hello" {
		t.Fatalf("unexpected echo: %+v", msg)
	}
}

func TestUnmaskedFrameRejected(t *testing.T) {
	url, closeSrv := startWSServer(t, Options{}, func(conn *Conn) { _ = conn.Recv(&Message{}) })
	defer closeSrv()

	conn, _ := dialWS(t, url, "")
	// Send an UNMASKED data frame — must be rejected (RFC 6455 §5.1).
	first := byte(0x80 | 0x1)
	if _, err := conn.Write([]byte{first, byte(len(`{}`)), '{', '}'}); err != nil {
		t.Fatalf("write: %v", err)
	}
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("expected clean EOF after unmasked frame, got %v", err)
	}
}

func TestOversizedFrameRejected(t *testing.T) {
	url, closeSrv := startWSServer(t, Options{MaxMessageSize: 64}, func(conn *Conn) { _ = conn.Recv(&Message{}) })
	defer closeSrv()

	conn, _ := dialWS(t, url, "")
	clientSendFrame(t, conn, 0x1, bytes.Repeat([]byte("a"), 65))
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("expected clean EOF after oversized frame, got %v", err)
	}
}

func TestHugeDeclaredLengthRejected(t *testing.T) {
	// The original GIN-003 attack: declare 2^40 payload, send nothing.
	url, closeSrv := startWSServer(t, Options{}, func(conn *Conn) { _ = conn.Recv(&Message{}) })
	defer closeSrv()

	conn, _ := dialWS(t, url, "")
	header := []byte{0x81, 0x80 | 127, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	header = append(header, 0x11, 0x22, 0x33, 0x44)
	if _, err := conn.Write(header); err != nil {
		t.Fatalf("write: %v", err)
	}
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("expected clean EOF for huge declared length, got %v", err)
	}
}

func TestOriginCheck(t *testing.T) {
	url, closeSrv := startWSServer(t, Options{}, func(conn *Conn) { _ = conn.Recv(&Message{}) })
	defer closeSrv()

	tcp := strings.Replace(url, "ws://", "", 1)
	dialRaw := func(origin string) (net.Conn, *http.Response) {
		t.Helper()
		conn, err := net.Dial("tcp", tcp)
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		t.Cleanup(func() { _ = conn.Close() })
		req := "GET / HTTP/1.1\r\nHost: " + tcp + "\r\n"
		if origin != "" {
			req += "Origin: " + origin + "\r\n"
		}
		req += "Upgrade: websocket\r\nConnection: Upgrade\r\n" +
			"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n"
		if _, err := conn.Write([]byte(req)); err != nil {
			t.Fatalf("write: %v", err)
		}
		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			t.Fatalf("read response: %v", err)
		}
		return conn, resp
	}

	// Mismatched origin → 403 before upgrade.
	_, resp := dialRaw("http://evil.example.com")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-origin, got %d", resp.StatusCode)
	}

	// Same origin → upgraded.
	_, resp2 := dialRaw("http://" + tcp)
	if resp2.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101 for same-origin, got %d", resp2.StatusCode)
	}

	// No Origin header (non-browser) → allowed by default.
	_, resp3 := dialRaw("")
	if resp3.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101 without Origin, got %d", resp3.StatusCode)
	}
}

func TestPingPong(t *testing.T) {
	url, closeSrv := startWSServer(t, Options{}, func(conn *Conn) {
		for {
			var msg Message
			if err := conn.Recv(&msg); err != nil {
				return
			}
		}
	})
	defer closeSrv()

	conn, br := dialWS(t, url, "")
	clientSendFrame(t, conn, 0x9, []byte("ping"))
	opcode, payload, err := clientReadFrame(t, br)
	if err != nil {
		t.Fatalf("read pong: %v", err)
	}
	if opcode != 0xA {
		t.Fatalf("expected pong opcode 0xA, got %#x", opcode)
	}
	if string(payload) != "ping" {
		t.Fatalf("pong payload: %q", payload)
	}
}

func TestReadTimeoutClosesIdleConn(t *testing.T) {
	url, closeSrv := startWSServer(t, Options{ReadTimeout: 100 * time.Millisecond}, func(conn *Conn) {
		_ = conn.Recv(&Message{})
	})
	defer closeSrv()

	conn, _ := dialWS(t, url, "")
	start := time.Now()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err := io.ReadAll(conn)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("expected clean EOF from read deadline, got %v", err)
	}
	if elapsed < 50*time.Millisecond {
		t.Fatalf("closed too early (%v) — not deadline-driven", elapsed)
	}
	if elapsed > 3*time.Second {
		t.Fatalf("closed too late (%v) — deadline not applied", elapsed)
	}
}

func TestLargeButWithinLimitEcho(t *testing.T) {
	url, closeSrv := startWSServer(t, Options{MaxMessageSize: 128 * 1024}, func(conn *Conn) {
		for {
			var msg Message
			if err := conn.Recv(&msg); err != nil {
				return
			}
			_ = conn.Send(Message{Type: "echo", Data: msg.Data})
		}
	})
	defer closeSrv()

	conn, br := dialWS(t, url, "")
	big := strings.Repeat("x", 100*1024)
	clientSendFrame(t, conn, 0x1, []byte(`{"type":"chat","data":"`+big+`"}`))
	_, payload, err := clientReadFrame(t, br)
	if err != nil {
		t.Fatalf("read large echo: %v", err)
	}
	if !strings.Contains(string(payload), big) {
		t.Fatal("large payload truncated")
	}
}

// ── unit: decodeFrame invariants ─────────────────────────────────────────────

func TestDecodeFrameReservedBitsRejected(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte{0xC1, 0x80, 0x00, 0x00, 0x00, 0x00}) // RSV1 set, masked, len 0
	rw := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(io.Discard))
	fr := newFrameReader(rw, &dummyConn{}, frameLimits{maxSize: 1024})
	if _, _, err := fr.decodeFrame(); err == nil || !strings.Contains(err.Error(), "reserved bits") {
		t.Fatalf("expected reserved bits error, got %v", err)
	}
}

func TestDecodeFrameControlTooLargeRejected(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte{0x89, 0x80 | 126, 0x00, 0x80}) // ping declared 128 > 125
	rw := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(io.Discard))
	fr := newFrameReader(rw, &dummyConn{}, frameLimits{maxSize: 1024})
	if _, _, err := fr.decodeFrame(); err == nil || !strings.Contains(err.Error(), "125") {
		t.Fatalf("expected control size error, got %v", err)
	}
}

func TestErrMessageTooLargeIsExported(t *testing.T) {
	if !errors.Is(ErrMessageTooLarge, ErrMessageTooLarge) {
		t.Fatal("unreachable")
	}
}
