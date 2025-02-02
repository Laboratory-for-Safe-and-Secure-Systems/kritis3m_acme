package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/handlers"
)

func New() *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
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
