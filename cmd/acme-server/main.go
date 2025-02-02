package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/router"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize router
	r := router.New()

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting ACME server on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
