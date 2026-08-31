package ws

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

// dummyConn satisfies net.Conn without real I/O — records pong writes.
type dummyConn struct {
	written bytes.Buffer
}

func (d *dummyConn) Read(b []byte) (int, error)         { return 0, io.EOF }
func (d *dummyConn) Write(b []byte) (int, error)        { return d.written.Write(b) }
func (d *dummyConn) Close() error                       { return nil }
func (d *dummyConn) LocalAddr() net.Addr                { return nil }
func (d *dummyConn) RemoteAddr() net.Addr               { return nil }
func (d *dummyConn) SetDeadline(t time.Time) error      { return nil }
func (d *dummyConn) SetReadDeadline(t time.Time) error  { return nil }
func (d *dummyConn) SetWriteDeadline(t time.Time) error { return nil }

// FuzzDecodeFrame (GIN-003): o decoder nunca deve panicar nem alocar memória
// proporcional a um length declarado arbitrário, independentemente da entrada.
func FuzzDecodeFrame(f *testing.F) {
	seeds := [][]byte{
		{0x81, 0x80, 0x00, 0x00, 0x00, 0x00},                          // masked empty text
		{0x81, 0x85, 0x11, 0x22, 0x33, 0x44, 'h', 'e', 'l', 'l', 'o'}, // masked "hello"
		{0x01, 0x80, 0, 0, 0, 0},                                      // unmasked (rejeitado)
		{0x89, 0x80, 0, 0, 0, 0},                                      // ping
		{0x88, 0x80, 0, 0, 0, 0},                                      // close
		{0x91, 0x80, 0, 0, 0, 0},                                      // reserved bits
		{0x81, 0xFF, 0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},        // 64-bit length enorme
		{0x81}, // truncado
	}
	for _, s := range seeds {
		f.Add(s)
	}

	limits := frameLimits{maxSize: 64 << 10}

	f.Fuzz(func(t *testing.T, data []byte) {
		conn := &dummyConn{}
		rw := bufio.NewReadWriter(bufio.NewReader(bytes.NewReader(data)), bufio.NewWriter(io.Discard))
		fr := newFrameReader(rw, conn, limits)
		// Ler até 4 frames; cada um termina em erro controlado ou payload ≤ limit.
		for i := 0; i < 4; i++ {
			payload, opcode, err := fr.decodeFrame()
			if err != nil {
				return // erro controlado — fim
			}
			switch opcode {
			case 0x9:
				continue
			case 0x8:
				return
			default:
				if int64(len(payload)) > limits.maxSize {
					t.Fatalf("payload %d exceeds limit %d", len(payload), limits.maxSize)
				}
			}
		}
	})
}
