package handlers

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/database"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/go-chi/chi/v5"
)

// GetAuthorization retrieves an authorization from the database if available.
// Otherwise it falls back to a mock authorization.
func GetAuthorization(w http.ResponseWriter, r *http.Request) {
	authzID := chi.URLParam(r, "id")
	log := logger.GetLogger(r.Context())
	baseURL := getBaseURL(r)

	// Try to retrieve the authorization from the database.
	db, ok := r.Context().Value(types.CtxKeyDB).(*database.DB)
	var authz *types.Authorization
	var err error
	if ok && db != nil {
		authz, err = db.GetAuthorization(r.Context(), authzID)
		if err != nil {
			log.Errorf("Failed to get authorization from database: %v", err)
			writeError(w, &types.Problem{
				Type:   "urn:ietf:params:acme:error:authorizationNotFound",
				Detail: fmt.Sprintf("Authorization %s not found", authzID),
				Status: http.StatusNotFound,
			})
			return
		}
		// Update challenge URLs based on the current host.
		for i := range authz.Challenges {
			if authz.Challenges[i].URL == "" {
				authz.Challenges[i].URL = endpointURL(baseURL, "challenge", authz.Challenges[i].Token)
			}
		}
	} else {
		// Fallback to a mock authorization.
		baseURL := "https://" + r.Host
		authz = &types.Authorization{
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
					Token:  generateToken(),
				},
				{
					Type:   "tls-alpn-01",
					URL:    path.Join(baseURL, "challenge", "chall-tls-alpn-01"),
					Status: types.ChallengeStatusPending,
					Token:  generateToken(),
				},
			},
			Wildcard: false,
		}
	}

	// add rel="up" to the header
	setLinkHeader(w, endpointURL(baseURL, "directory", ""), "up")
	if err := writeJSON(w, http.StatusOK, authz); err != nil {
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

// ProcessChallenge verifies that the challenge exists and is pending,
// then simulates validation by updating its status to valid.
func ProcessChallenge(w http.ResponseWriter, r *http.Request) {
	challengeID := chi.URLParam(r, "id")
	log := logger.GetLogger(r.Context())
	baseURL := getBaseURL(r)

	// Retrieve the challenge from the database.
	db, ok := r.Context().Value(types.CtxKeyDB).(*database.DB)
	if !ok || db == nil {
		log.Error("Database not available in context")
		writeError(w, newInternalServerError("Database not available"))
		return
	}

	challenge, err := db.GetChallenge(r.Context(), challengeID)
	if err != nil {
		log.Errorf("Challenge not found: %v", err)
		writeError(w, newNotFoundError(fmt.Sprintf("Challenge %s not found", challengeID), "challengeNotFound"))
		return
	}

	// Ensure the challenge is in a pending state.
	if challenge.Status != types.ChallengeStatusPending {
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:challengeInvalid",
			Detail: "Challenge is not in a pending state",
			Status: http.StatusBadRequest,
		})
		return
	}

	// Simulate challenge validation.
	newStatus := types.ChallengeStatusValid
	if err := db.UpdateChallengeStatus(r.Context(), challengeID, string(newStatus)); err != nil {
		log.Errorf("Failed to update challenge status: %v", err)
		writeError(w, newInternalServerError("Failed to update challenge status"))
		return
	}

	// Update the local challenge status.
	challenge.Status = newStatus

	// Set a challenge URL using the current request host.
	challenge.URL = endpointURL(baseURL, "challenge", challengeID)
	log.Infof("Challenge URL: %s", challenge.URL)

	// Update authorization status if all challenges are valid
	log.Infof("Challenge AuthorizationID: %s", challenge.AuthorizationID)
	authz, err := db.GetAuthorization(r.Context(), challenge.AuthorizationID)
	if err != nil {
		log.Errorf("Failed to get authorization: %v", err)
		writeError(w, newInternalServerError("Failed to update authorization status"))
		return
	}

	// Check if all challenges for this authorization are valid
	challenges, err := db.GetChallengesByAuthorization(r.Context(), authz.ID)
	if err != nil {
		log.Errorf("Failed to get challenges: %v", err)
		writeError(w, newInternalServerError("Failed to check challenge status"))
		return
	}

	allValid := true
	for _, c := range challenges {
		if c.Status != types.ChallengeStatusValid {
			allValid = false
			break
		}
	}

	if allValid {
		authz.Status = types.AuthzStatusValid
		if err := db.UpdateAuthorizationStatus(r.Context(), authz.ID, string(types.AuthzStatusValid)); err != nil {
			log.Errorf("Failed to update authorization status: %v", err)
			writeError(w, newInternalServerError("Failed to update authorization status"))
			return
		}
	}

	// Add required Link header pointing to the authorization
	setLinkHeader(w, endpointURL(baseURL, "authorization", challenge.AuthorizationID), "up")

	if err := writeJSON(w, http.StatusOK, challenge); err != nil {
		log.Errorf("Failed to encode challenge response: %v", err)
		writeError(w, newInternalServerError("Failed to encode response"))
		return
	}
}
