package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// ErrMessageTooLarge is returned when an incoming frame exceeds the
// configured maximum payload size (GIN-003).
var ErrMessageTooLarge = fmt.Errorf("ws: message too large")

// computeAccept returns the Sec-WebSocket-Accept value for a given key.
func computeAccept(key string) string {
	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// ── Frame writer ─────────────────────────────────────────────────────────────

type frameWriter struct {
	conn net.Conn
}

func newFrameWriter(conn net.Conn) *frameWriter {
	return &frameWriter{conn: conn}
}

// Write wraps p in a single unmasked text frame (server→client, RFC 6455 §5).
func (fw *frameWriter) Write(p []byte) (int, error) {
	frame := encodeTextFrame(p)
	_, err := fw.conn.Write(frame)
	return len(p), err
}

func encodeTextFrame(payload []byte) []byte {
	return encodeFrame(0x1, payload)
}

// encodeFrame builds an unmasked server frame with the given opcode.
func encodeFrame(opcode byte, payload []byte) []byte {
	n := len(payload)
	first := byte(0x80 | opcode) // FIN + opcode
	var header []byte
	switch {
	case n <= 125:
		header = []byte{first, byte(n)}
	case n <= 65535:
		header = make([]byte, 4)
		header[0] = first
		header[1] = 126
		binary.BigEndian.PutUint16(header[2:], uint16(n))
	default:
		header = make([]byte, 10)
		header[0] = first
		header[1] = 127
		binary.BigEndian.PutUint64(header[2:], uint64(n))
	}
	return append(header, payload...)
}

// ── Frame reader ─────────────────────────────────────────────────────────────

type frameLimits struct {
	maxSize      int64         // maximum payload size (0 = no limit)
	readDeadline time.Duration // per-message deadline (0 = none)
}

type frameReader struct {
	buf     *bufio.ReadWriter
	conn    net.Conn
	limits  frameLimits
	pending []byte // remainder of the last decoded frame (io.Reader contract)
}

func newFrameReader(buf *bufio.ReadWriter, conn net.Conn, limits frameLimits) *frameReader {
	return &frameReader{buf: buf, conn: conn, limits: limits}
}

// Read decodes the next data frame and returns its unmasked payload.
// Frames larger than p are buffered and served across subsequent Read calls
// (json.Decoder reads in small chunks — dropping the remainder truncated
// messages larger than its internal buffer).
// Control frames are handled inline: ping → pong, close → EOF (GIN-003).
func (fr *frameReader) Read(p []byte) (int, error) {
	for len(fr.pending) == 0 {
		payload, opcode, err := fr.decodeFrame()
		if err != nil {
			return 0, err
		}
		switch opcode {
		case 0x9: // ping → pong, keep reading
			if _, err := fr.conn.Write(encodeFrame(0xA, payload)); err != nil {
				return 0, err
			}
			continue
		case 0x8: // close
			_, _ = fr.conn.Write(encodeFrame(0x8, payload))
			return 0, io.EOF
		default: // text (0x1), binary (0x2), continuation (0x0)
			fr.pending = payload
		}
	}
	n := copy(p, fr.pending)
	fr.pending = fr.pending[n:]
	return n, nil
}

// decodeFrame reads one frame with RFC 6455 safety checks:
// masked client frames only, no reserved bits, control frame size cap and a
// hard limit on the allocated payload (prevents OOM via fake 64-bit lengths).
func (fr *frameReader) decodeFrame() ([]byte, byte, error) {
	if fr.limits.readDeadline > 0 {
		_ = fr.conn.SetReadDeadline(time.Now().Add(fr.limits.readDeadline))
	}

	// Read first 2 bytes: FIN+opcode, MASK+length
	header := make([]byte, 2)
	if _, err := io.ReadFull(fr.buf, header); err != nil {
		return nil, 0, err
	}

	opcode := header[0] & 0x0F
	rsv := header[0] & 0x70
	masked := header[1]&0x80 != 0
	payloadLen := int64(header[1] & 0x7F)

	if rsv != 0 {
		return nil, 0, fmt.Errorf("ws: reserved bits set (no extension negotiated)")
	}
	if !masked {
		return nil, 0, fmt.Errorf("ws: client frames must be masked (RFC 6455 §5.1)")
	}
	if isControl(opcode) && payloadLen > 125 {
		return nil, 0, fmt.Errorf("ws: control frame payload exceeds 125 bytes")
	}
	if fr.limits.maxSize > 0 && payloadLen > fr.limits.maxSize {
		return nil, 0, ErrMessageTooLarge
	}

	switch payloadLen {
	case 126:
		ext := make([]byte, 2)
		if _, err := io.ReadFull(fr.buf, ext); err != nil {
			return nil, 0, err
		}
		payloadLen = int64(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		if _, err := io.ReadFull(fr.buf, ext); err != nil {
			return nil, 0, err
		}
		u := binary.BigEndian.Uint64(ext)
		if u > uint64(fr.limits.maxSize) && fr.limits.maxSize > 0 {
			return nil, 0, ErrMessageTooLarge
		}
		if u > uint64(maxInt) {
			return nil, 0, ErrMessageTooLarge
		}
		payloadLen = int64(u)
	}

	if fr.limits.maxSize > 0 && payloadLen > fr.limits.maxSize {
		return nil, 0, ErrMessageTooLarge
	}

	var maskKey [4]byte
	if _, err := io.ReadFull(fr.buf, maskKey[:]); err != nil {
		return nil, 0, err
	}

	// Safe allocation: payloadLen was validated against maxSize above.
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(fr.buf, payload); err != nil {
		return nil, 0, err
	}

	for i := range payload {
		payload[i] ^= maskKey[i%4]
	}
	return payload, opcode, nil
}

func isControl(opcode byte) bool {
	return opcode&0x8 != 0
}

// maxInt guards 64-bit length conversions on 32-bit platforms.
const maxInt = int64(^uint(0) >> 1)
