package server

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "valid address",
			addr:    "localhost:6379",
			wantErr: false,
		},
		{
			name:    "empty address",
			addr:    "",
			wantErr: false,
		},
		{
			name:    "port only",
			addr:    ":6379",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.addr)
			if s == nil {
				t.Error("New() returned nil server")
			}
			if s.port != tt.addr {
				t.Errorf("New() port = %v, want %v", s.port, tt.addr)
			}
			if s.store == nil {
				t.Error("New() store is nil")
			}
			if s.quit == nil {
				t.Error("New() quit channel is nil")
			}
		})
	}
}

func TestServer_StartAndStop(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "valid port",
			addr:    "localhost:6380",
			wantErr: false,
		},
		{
			name:    "already in use port",
			addr:    "localhost:6381",
			wantErr: true,
		},
		{
			name:    "invalid port",
			addr:    "localhost:999999",
			wantErr: true,
		},
	}

	// Create a server that will block the port for the "already in use" test
	blockingServer := New("localhost:6381")
	if err := blockingServer.Start(); err != nil {
		t.Fatalf("Failed to start blocking server: %v", err)
	}
	defer blockingServer.Stop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.addr)

			// Start server
			err := s.Start()
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.Start() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Try to connect to the server
				conn, err := net.Dial("tcp", tt.addr)
				if err != nil {
					t.Fatalf("Failed to connect to server: %v", err)
				}
				conn.Close()

				// Stop server
				if err := s.Stop(); err != nil {
					t.Fatalf("Server.Stop() error = %v", err)
				}

				// Verify server is stopped by trying to connect
				time.Sleep(100 * time.Millisecond)
				if _, err := net.Dial("tcp", tt.addr); err == nil {
					t.Error("Server still accepting connections after Stop()")
				}
			}
		})
	}
}

func TestServer_ConcurrentConnections(t *testing.T) {
	s := New("localhost:6382")
	if err := s.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer s.Stop()

	// Number of concurrent connections to test
	numConnections := 10
	var wg sync.WaitGroup
	wg.Add(numConnections)

	// Create multiple concurrent connections
	for i := 0; i < numConnections; i++ {
		go func() {
			defer wg.Done()

			conn, err := net.Dial("tcp", "localhost:6382")
			if err != nil {
				t.Errorf("Failed to connect: %v", err)
				return
			}
			defer conn.Close()

			// Keep connection open briefly
			time.Sleep(100 * time.Millisecond)
		}()
	}

	// Wait for all connections to complete
	wg.Wait()
}

func TestServer_GracefulShutdown(t *testing.T) {
	s := New("localhost:6383")
	if err := s.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Create a long-running connection
	conn, err := net.Dial("tcp", "localhost:6383")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Start graceful shutdown in a goroutine
	shutdownComplete := make(chan struct{})
	go func() {
		if err := s.Stop(); err != nil {
			t.Errorf("Server.Stop() error = %v", err)
		}
		close(shutdownComplete)
	}()

	// Keep connection open briefly
	time.Sleep(100 * time.Millisecond)

	// Close the connection
	conn.Close()

	// Wait for shutdown to complete with timeout
	select {
	case <-shutdownComplete:
		// Shutdown completed successfully
	case <-time.After(2 * time.Second):
		t.Error("Server shutdown did not complete in time")
	}

	// Verify server is stopped
	if _, err := net.Dial("tcp", "localhost:6383"); err == nil {
		t.Error("Server still accepting connections after Stop()")
	}
}

func TestServer_StopWithoutStart(t *testing.T) {
	s := New("localhost:6384")
	if err := s.Stop(); err != nil {
		t.Errorf("Server.Stop() error = %v, want nil", err)
	}
}

func TestServer_MultipleStops(t *testing.T) {
	s := New("localhost:6385")
	if err := s.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// First stop should succeed
	if err := s.Stop(); err != nil {
		t.Errorf("First Server.Stop() error = %v", err)
	}

	// Second stop should also succeed (idempotent)
	if err := s.Stop(); err != nil {
		t.Errorf("Second Server.Stop() error = %v", err)
	}
}

func TestServer_StartAfterStop(t *testing.T) {
	s := New("localhost:6386")

	// First start
	if err := s.Start(); err != nil {
		t.Fatalf("First Server.Start() error = %v", err)
	}

	// Stop
	if err := s.Stop(); err != nil {
		t.Fatalf("Server.Stop() error = %v", err)
	}

	// Second start should succeed
	if err := s.Start(); err != nil {
		t.Errorf("Second Server.Start() error = %v", err)
	}

	// Clean up
	s.Stop()
}
