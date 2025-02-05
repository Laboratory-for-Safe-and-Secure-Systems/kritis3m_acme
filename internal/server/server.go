package server

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/go-asl"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/go-asl/lib"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
)

const TLSStateKey = "TLSState"

type Server struct {
	// Get Logger from context
	logger *logger.Logger

	*http.Server
	aslListener *lib.ASLListener
}

func ASLServer(config *ASLServerConfig, handler http.Handler, ctx context.Context) (*Server, error) {
	// Create ASL Listener
	logger := logger.GetLogger(ctx)

	listener, err := NewASLListener(config)
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
			if aslConn, ok := c.(*lib.ASLConn); ok {
				return context.WithValue(ctx, TLSStateKey, aslConn.TLSState)
			}
			return ctx
		},
	}

	return &Server{
		logger:      logger,
		Server:      srv,
		aslListener: listener,
	}, nil
}

func (s *Server) ListenAndServeASL() error {
	return s.Server.Serve(s.aslListener)
}

func (s *Server) Shutdown(ctx context.Context) error {
	asl.ASLFreeEndpoint(s.aslListener.Endpoint)
	if err := s.Server.Shutdown(ctx); err != nil {
		s.logger.Errorf("Error shutting down server: %v", err)
		return err
	}

	if err := s.aslListener.Close(); err != nil {
		s.logger.Errorf("Error closing ASL listener: %v", err)
		return err
	}

	return nil
}
