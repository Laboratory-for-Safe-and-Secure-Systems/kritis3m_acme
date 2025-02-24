package server

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/go-asl"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/go-asl/listener"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
)

type contextKey string

const TLSStateKey contextKey = "TLSState"

type Server struct {
	// Get Logger from context
	logger      *logger.Logger
	aslListener *listener.ASLListener

	*http.Server
}

func ASLServer(config *ASLServerConfig, handler http.Handler, ctx context.Context) (*Server, error) {
	// Create ASL Listener
	logger := logger.GetLogger(ctx)

	aslListener, err := NewASLListener(config)
	if err != nil {
		return nil, err
	}

	// Create HTTP Server
	srv := &http.Server{
		Addr:     config.Address,
		ErrorLog: log.New(logger, "", 0),
		Handler:  handler,
		// Add custom connection context to access ASL state
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			if aslConn, ok := c.(*listener.ASLConn); ok {
				return context.WithValue(ctx, TLSStateKey, aslConn.TLSState)
			}
			return ctx
		},
	}

	return &Server{
		logger:      logger,
		Server:      srv,
		aslListener: aslListener,
	}, nil
}

func (s *Server) ListenAndServeASL() error {
	return s.Server.Serve(s.aslListener)
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Shutdown the HTTP server
	err := s.Server.Shutdown(ctx)
	if err != nil {
		s.logger.Errorf("Error shutting down HTTP server: %v", err)
	}

	// Shutdown the ASL endpoint
	asl.ASLFreeEndpoint(s.aslListener.Endpoint)

	return nil
}
