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
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/middleware/acme"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/database"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
)

func New(ctx context.Context, db *database.DB) *chi.Mux {
	r := chi.NewRouter()

	// Add database to context middleware if provided
	if db != nil {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), types.CtxKeyDB, db)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
	}

	// Global middleware
	r.Use(withLogger(logger.GetLogger(ctx)))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "Content-Length"},
		MaxAge:         300, // Maximum value not ignored by any of major browsers
	}))

	// Public endpoints (no nonce or JWT verification required)
	r.Get("/health", handlers.HealthCheck)
	r.Get("/directory", handlers.GetDirectory)

	// ACME protocol endpoints
	r.Group(func(r chi.Router) {
		// Add nonce middleware to all ACME endpoints
		r.Use(acme.NonceMiddleware)

		// Nonce endpoints
		r.Get("/new-nonce", handlers.NewNonce)
		r.Head("/new-nonce", handlers.NewNonce) // Required by RFC 8555

		// Account management (with JWT verification)
		r.Group(func(r chi.Router) {
			r.Use(acme.JWSVerificationMiddleware)
			r.Post("/new-account", handlers.NewAccount)
			r.Post("/account/{id}", handlers.UpdateAccount)
			r.Post("/account/{id}/key-change", handlers.KeyChange)

			// Order management
			r.Post("/new-order", handlers.NewOrder)
			r.Get("/order/{id}", handlers.GetOrder)
			r.Post("/order/{id}", handlers.GetOrder)
			r.Post("/order/{id}/finalize", handlers.FinalizeOrder)

			// Authorization management
			r.Get("/authz/{id}", handlers.GetAuthorization)
			r.Post("/authz/{id}", handlers.GetAuthorization)
			// Challenge management
			r.Post("/challenge/{id}", handlers.ProcessChallenge)

			// Certificate management
			r.Post("/cert/{certID}", handlers.GetCertificate)
			r.Post("/revoke-cert", handlers.RevokeCertificate)
		})
	})

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
				scheme := "https"

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

			ctx := context.WithValue(r.Context(), types.CtxKeyLogger, logger)
			next.ServeHTTP(ww, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
