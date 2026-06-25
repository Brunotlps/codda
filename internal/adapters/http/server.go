package http

import (
	"context"
	"errors"
	"net/http"
	"time"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

// Server wraps an http.Server configured with production-ready timeouts and
// graceful shutdown support.
type Server struct {
	httpServer *http.Server
}

// NewServer creates a Server that will listen on addr (e.g. ":8080" or
// "0.0.0.0:8080") and serve handler.
func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}
}

// Start begins serving requests, blocking until the server stops. It
// returns nil if the server was stopped via Shutdown, or the error that
// caused it to stop otherwise.
func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully stops the server: it stops accepting new connections
// and waits for in-flight requests to finish, until ctx is done.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
