package server

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/hardikphalet/go-redis/internal/store"
)

type Server struct {
	listener net.Listener
	store    store.Store
	port     string
	wg       sync.WaitGroup
	quit     chan struct{}
	mu       sync.Mutex
	stopped  bool
}

func New(address string) *Server {
	return &Server{
		port:    address,
		store:   store.NewMemoryStore(),
		quit:    make(chan struct{}),
		stopped: false,
	}
}

func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.port)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	log.Printf("Server listening on %s", s.port)

	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return nil
	}
	s.stopped = true
	close(s.quit)
	s.mu.Unlock()

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	s.wg.Wait()
	return nil
}

func (s *Server) acceptConnections() {
	defer s.wg.Done()

	for {
		select {
		case <-s.quit:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.quit:
					return
				default:
					log.Printf("Error accepting connection: %v", err)
					continue
				}
			}

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	remoteAddr := conn.RemoteAddr().String()
	log.Printf("New client connection from %s", remoteAddr)

	handler := NewHandler(conn, s.store)
	if err := handler.Handle(); err != nil {
		log.Printf("Error handling connection from %s: %v", remoteAddr, err)
	} else {
		log.Printf("Client %s disconnected", remoteAddr)
	}
}
