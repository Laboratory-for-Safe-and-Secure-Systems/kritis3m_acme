package handlers

import (
	"encoding/json"
	"net/http"
	"path"
	"time"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/go-chi/chi/v5"
)

func GetAuthorization(w http.ResponseWriter, r *http.Request) {
	authzID := chi.URLParam(r, "id")
	log := logger.GetLogger(r.Context())

	// TODO: Retrieve authorization from database
	// For now, return a mock authorization
	baseURL := "http://" + r.Host

	// Create mock authorization
	authz := &types.Authorization{
		Status:  types.AuthzStatusPending,
		Expires: &types.Time{Time: time.Now().Add(24 * time.Hour)},
		Identifier: types.Identifier{
			Type:  "dns",
			Value: "example.com",
		},
		Challenges: []types.Challenge{
			{
				Type:   "http-01",
				URL:    path.Join(baseURL, "challenge", "chall-http-01"),
				Status: types.ChallengeStatusPending,
				Token:  generateToken(), // TODO: Implement proper token generation
			},
			{
				Type:   "tls-alpn-01",
				URL:    path.Join(baseURL, "challenge", "chall-tls-alpn-01"),
				Status: types.ChallengeStatusPending,
				Token:  generateToken(),
			},
		},
		Wildcard: nil, // Not a wildcard authorization
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write response
	if err := json.NewEncoder(w).Encode(authz); err != nil {
		log.Errorf("Failed to encode authorization response: %v", err)
		writeError(w, newInternalServerError("Failed to encode response"))
		return
	}

	log.Infow("Authorization request handled",
		"id", authzID,
		"status", authz.Status,
		"identifier", authz.Identifier,
	)
}

// Helper function to generate a random token
func generateToken() string {
	// TODO: Implement proper token generation
	return "token-123"
}

func ProcessChallenge(w http.ResponseWriter, r *http.Request) {
	challengeID := chi.URLParam(r, "id")
	log := logger.GetLogger(r.Context())

	// TODO:
	// 1. Verify challenge exists
	// 2. Verify challenge is in valid state
	// 3. Start challenge validation
	// 4. Update challenge status

	log.Infof("Process challenge request for ID: %s", challengeID)
	w.WriteHeader(http.StatusNotImplemented)
}
