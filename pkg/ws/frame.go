package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

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
	frame := encodeFrame(p)
	_, err := fw.conn.Write(frame)
	return len(p), err
}

func encodeFrame(payload []byte) []byte {
	n := len(payload)
	var header []byte
	switch {
	case n <= 125:
		header = []byte{0x81, byte(n)}
	case n <= 65535:
		header = make([]byte, 4)
		header[0] = 0x81
		header[1] = 126
		binary.BigEndian.PutUint16(header[2:], uint16(n))
	default:
		header = make([]byte, 10)
		header[0] = 0x81
		header[1] = 127
		binary.BigEndian.PutUint64(header[2:], uint64(n))
	}
	return append(header, payload...)
}

// ── Frame reader ─────────────────────────────────────────────────────────────

type frameReader struct {
	buf  *bufio.ReadWriter
	conn net.Conn
}

func newFrameReader(buf *bufio.ReadWriter, conn net.Conn) *frameReader {
	return &frameReader{buf: buf, conn: conn}
}

// Read decodes the next WebSocket frame and returns its unmasked payload.
func (fr *frameReader) Read(p []byte) (int, error) {
	payload, err := decodeFrame(fr.buf)
	if err != nil {
		return 0, err
	}
	return copy(p, payload), nil
}

func decodeFrame(r *bufio.ReadWriter) ([]byte, error) {
	// Read first 2 bytes: FIN+opcode, MASK+length
	header := make([]byte, 2)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	masked := header[1]&0x80 != 0
	payloadLen := int(header[1] & 0x7F)

	switch payloadLen {
	case 126:
		ext := make([]byte, 2)
		if _, err := io.ReadFull(r, ext); err != nil {
			return nil, err
		}
		payloadLen = int(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		if _, err := io.ReadFull(r, ext); err != nil {
			return nil, err
		}
		payloadLen = int(binary.BigEndian.Uint64(ext))
	}

	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(r, maskKey[:]); err != nil {
			return nil, err
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}
	return payload, nil
}
