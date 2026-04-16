package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Server is the pigo HTTP server.
type Server struct {
	http *http.Server
	addr string
}

// New creates a server that listens on the given host:port.
func New(host string, port int) *Server {
	addr := fmt.Sprintf("%s:%d", host, port)

	mux := http.NewServeMux()
	mux.HandleFunc("/command", handleCommand)
	mux.HandleFunc("/health", handleHealth)

	return &Server{
		http: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		addr: addr,
	}
}

// Start begins listening. It blocks until the server is shut down.
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.addr, err)
	}
	fmt.Printf("pigo server listening on %s\n", s.addr)
	return s.http.Serve(listener)
}

// Shutdown gracefully stops the server with a timeout.
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.http.Shutdown(ctx)
}

// Addr returns the configured listen address.
func (s *Server) Addr() string {
	return s.addr
}
