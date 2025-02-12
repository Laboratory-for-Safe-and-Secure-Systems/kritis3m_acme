package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/handlers"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
)

// ctxKeyLogger is the context key for the logger.
const (
	ctxKeyLogger = iota
)

func New(ctx context.Context) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(withLogger(logger.GetLogger(ctx)))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	// ACME Directory endpoints
	r.Get("/directory", handlers.GetDirectory)

	// Nonce endpoints
	r.Get("/new-nonce", handlers.NewNonce)

	// Account management
	r.Post("/new-account", handlers.NewAccount)

	// Health check
	r.Get("/health", handlers.HealthCheck)

	return r
}

// withLogger is middleware that logs each HTTP request.
func withLogger(logger *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			// Capture our own copy of the logger so change in this closure
			// won't affect the object passed-in.

			logger := logger

			if reqID := middleware.GetReqID(r.Context()); reqID != "" {
				logger = logger.With("HTTP Request ID", reqID)
			}

			// Defer a function to log and entry once the main handler
			// has returned.

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			defer func() {
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}

				logger.Infow("HTTP request",
					"Method", r.Method,
					"URI", fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI),
					"Protocol", r.Proto,
					"Remote Address", r.RemoteAddr,
					"Status", ww.Status(),
					"Bytes Written", ww.BytesWritten(),
					"Time Taken", time.Since(t1),
				)
			}()

			ctx := context.WithValue(r.Context(), ctxKeyLogger, logger)
			next.ServeHTTP(ww, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
