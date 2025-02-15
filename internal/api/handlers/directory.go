package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/server"
)

func GetDirectory(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger(r.Context())

	log.Infow("Received directory request",
		"headers", r.Header,
		"method", r.Method,
		"url", r.URL.String(),
		"proto", r.Proto,
	)

	// Determine base URL (respect X-Forwarded-Proto header and ASL state for proxied requests)
	scheme := "http"
	if aslState := r.Context().Value(server.TLSStateKey); aslState != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := scheme + "://" + r.Host

	// Create directory response
	dir := types.Directory{
		NewNonce:   baseURL + "/new-nonce",
		NewAccount: baseURL + "/new-account",
		NewOrder:   baseURL + "/new-order",
		RevokeCert: baseURL + "/revoke-cert",
		KeyChange:  baseURL + "/key-change",
		Meta: &types.DirectoryMetadata{
			TermsOfService:          baseURL + "/terms",
			Website:                 "https://github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme",
			CAAIdentities:           []string{"kritis3m.example.com"},
			ExternalAccountRequired: false,
		},
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	// Write response
	if err := json.NewEncoder(w).Encode(dir); err != nil {
		log.Errorf("Failed to encode directory response: %v", err)
		writeError(w, newInternalServerError("Failed to encode response"))
		return
	}

	log.Debugw("Directory request handled successfully",
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent(),
	)
}

// Helper function to write error responses
func writeError(w http.ResponseWriter, problem *types.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	json.NewEncoder(w).Encode(problem)
}

// Helper function to create internal server error
func newInternalServerError(detail string) *types.Problem {
	return &types.Problem{
		Type:   "urn:ietf:params:acme:error:serverInternal",
		Detail: detail,
		Status: http.StatusInternalServerError,
	}
}

func NewNonce(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
