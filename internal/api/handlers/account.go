package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/middleware/acme"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/database"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/go-chi/chi/v5"
)

func NewAccount(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger(r.Context())
	db := r.Context().Value(types.CtxKeyDB).(*database.DB)
	baseURL := getBaseURL(r)

	// Get the protected header from context
	protected, ok := r.Context().Value(acme.JwsProtectedKey).(*acme.JWSHeader)
	if !ok {
		log.Error("Failed to get protected header from context")
		writeError(w, newInternalServerError("Failed to get protected header"))
		return
	}

	// Get the decoded payload from context
	payloadBytes, ok := r.Context().Value(acme.DecodedPayloadKey).([]byte)
	if !ok {
		log.Error("Failed to get decoded payload from context")
		writeError(w, newInternalServerError("Failed to get decoded payload"))
		return
	}

	// Parse the request body
	var req types.AccountRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		log.Errorf("Failed to decode account request: %v", err)
		writeError(w, newMalformedError("Failed to parse account request"))
		return
	}

	// Check if Terms of Service were agreed to
	if !req.TermsOfServiceAgreed {
		writeError(w, newBadRequestError("Must agree to terms of service"))
		return
	}

	// Generate a unique account ID
	accountID := generateID("acct")

	// Store the JWK as JSON
	jwkJSON, err := json.Marshal(protected.Jwk)
	if err != nil {
		log.Errorf("Failed to marshal JWK: %v", err)
		writeError(w, newInternalServerError("Failed to process account key"))
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
		OrdersURL:            endpointURL(baseURL, "orders", accountID),
	}

	// Store account in database
	if err := db.CreateAccount(r.Context(), account); err != nil {
		log.Errorf("Failed to create account: %v", err)
		writeError(w, newInternalServerError("Failed to create account"))
		return
	}

	// Set response headers with correct account URL format
	accountURL := endpointURL(baseURL, "account", account.ID)
	w.Header().Set("Location", accountURL)

	// Also update the OrdersURL in the account
	account.OrdersURL = endpointURL(baseURL, "orders", account.ID)

	// Write response
	if err := writeJSON(w, http.StatusCreated, account); err != nil {
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
