package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/middleware/acme"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/database"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/go-chi/chi/v5"
)

// generateAccountID creates a unique account ID
func generateAccountID() string {
	// For now, use timestamp + random suffix
	return fmt.Sprintf("acct_%d", time.Now().UnixNano())
}

func NewAccount(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger(r.Context())
	db := r.Context().Value(types.CtxKeyDB).(*database.DB)

	// Get the protected header from context
	protected, ok := r.Context().Value(acme.JwsProtectedKey).(*acme.JWSHeader)
	if !ok {
		log.Error("Failed to get protected header from context")
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to get protected header",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// Get the decoded payload from context
	payloadBytes, ok := r.Context().Value(acme.DecodedPayloadKey).([]byte)
	if !ok {
		log.Error("Failed to get decoded payload from context")
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to get decoded payload",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// Parse the request body
	var req types.AccountRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		log.Errorf("Failed to decode account request: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:malformed",
			Detail: "Failed to parse account request",
			Status: http.StatusBadRequest,
		})
		return
	}

	// Check if Terms of Service were agreed to
	if !req.TermsOfServiceAgreed {
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:userActionRequired",
			Detail: "Must agree to terms of service",
			Status: http.StatusBadRequest,
		})
		return
	}

	// Generate a unique account ID
	accountID := generateAccountID()

	// Store the JWK as JSON
	jwkJSON, err := json.Marshal(protected.Jwk)
	if err != nil {
		log.Errorf("Failed to marshal JWK: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to process account key",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// Create new account
	account := &types.Account{
		ID:                   accountID,
		Key:                  jwkJSON,
		Status:               types.AccountStatusValid,
		Contact:              req.Contact,
		TermsOfServiceAgreed: true,
		CreatedAt:            time.Now().Unix(),
		InitialIP:            r.RemoteAddr,
		OrdersURL:            fmt.Sprintf("https://%s/orders/%s", r.Host, accountID),
	}

	// Store account in database
	if err := db.CreateAccount(r.Context(), account); err != nil {
		log.Errorf("Failed to create account: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to create account",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("https://%s/account/%s", r.Host, account.ID))
	w.WriteHeader(http.StatusCreated)

	// Write response
	if err := json.NewEncoder(w).Encode(account); err != nil {
		log.Errorf("Failed to encode account response: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to encode response",
			Status: http.StatusInternalServerError,
		})
		return
	}
}

func UpdateAccount(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "id")
	log := logger.GetLogger(r.Context())

	// TODO: Implement account update
	// 1. Verify account exists
	// 2. Verify JWS is signed by account key
	// 3. Update account details

	log.Infof("Update account request for ID: %s", accountID)
	w.WriteHeader(http.StatusNotImplemented)
}

func KeyChange(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "id")
	log := logger.GetLogger(r.Context())

	// TODO: Implement key change
	// 1. Verify account exists
	// 2. Verify old key signed outer JWS
	// 3. Verify new key signed inner JWS
	// 4. Update account key

	log.Infof("Key change request for account ID: %s", accountID)
	w.WriteHeader(http.StatusNotImplemented)
}
