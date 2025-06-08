package server

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/hardikphalet/go-redis/internal/store"
)

type mockConn struct {
	reader *bytes.Buffer
	writer *bytes.Buffer
}

func newMockConn(input string) mockConn {
	return mockConn{
		reader: bytes.NewBuffer([]byte(input)),
		writer: &bytes.Buffer{},
	}
}

func (m mockConn) Read(b []byte) (n int, err error)   { return m.reader.Read(b) }
func (m mockConn) Write(b []byte) (n int, err error)  { return m.writer.Write(b) }
func (m mockConn) Close() error                       { return nil }
func (m mockConn) LocalAddr() net.Addr                { return nil }
func (m mockConn) RemoteAddr() net.Addr               { return nil }
func (m mockConn) SetDeadline(t time.Time) error      { return nil }
func (m mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ping command",
			input:    "*1\r\n$4\r\nPING\r\n",
			expected: "+PONG\r\n",
		},
		{
			name:     "echo command",
			input:    "*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n",
			expected: "$5\r\nhello\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := newMockConn(tt.input)

			store := store.NewMemoryStore()
			handler := NewHandler(conn, store)

			errCh := make(chan error)
			go func() {
				errCh <- handler.Handle()
			}()

			time.Sleep(100 * time.Millisecond)

			if got := conn.writer.String(); got != tt.expected {
				t.Errorf("Handler.Handle() output = %q, want %q", got, tt.expected)
			}
		})
	}
}
