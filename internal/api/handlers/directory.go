package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
)

func GetDirectory(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger(r.Context())

	log.Infow("Received directory request",
		"headers", r.Header,
		"method", r.Method,
		"url", r.URL.String(),
		"proto", r.Proto,
	)

	baseURL := getBaseURL(r)

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
	setLinkHeader(w, fmt.Sprintf("%s/directory", baseURL), "up")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	// Write response
	if err := writeJSON(w, http.StatusOK, dir); err != nil {
		log.Errorf("Failed to encode directory response: %v", err)
		writeError(w, newInternalServerError("Failed to encode response"))
		return
	}

	log.Debugw("Directory request handled successfully",
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent(),
	)
}

func NewNonce(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
