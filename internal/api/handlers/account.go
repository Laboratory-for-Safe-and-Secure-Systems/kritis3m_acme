package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/go-chi/chi/v5"
)

func NewAccount(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger(r.Context())

	// Parse the request body
	var req types.AccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	// TODO: Extract JWK from JWS header for account key
	// TODO: Check if account already exists with this key
	// TODO: Store account in database

	// Create new account
	account := &types.Account{
		ID:                   "account-id", // TODO: Generate proper ID
		Status:               types.AccountStatusValid,
		Contact:              req.Contact,
		TermsOfServiceAgreed: true,
		CreatedAt:            time.Now().Unix(),
		InitialIP:            r.RemoteAddr,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/account/"+account.ID)
	w.WriteHeader(http.StatusCreated)

	// Write response
	if err := json.NewEncoder(w).Encode(account); err != nil {
		log.Errorf("Failed to encode account response: %v", err)
		writeError(w, newInternalServerError("Failed to encode response"))
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
