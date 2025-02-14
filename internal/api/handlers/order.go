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

func NewOrder(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger(r.Context())

	// Parse request
	var req types.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Errorf("Failed to decode order request: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:malformed",
			Detail: "Failed to parse order request",
			Status: http.StatusBadRequest,
		})
		return
	}

	// Validate identifiers
	if len(req.Identifiers) == 0 {
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:malformed",
			Detail: "At least one identifier required",
			Status: http.StatusBadRequest,
		})
		return
	}

	// TODO: Generate proper order ID
	orderID := "order-123"
	baseURL := "http://" + r.Host // Should match your directory URL construction

	// Create new order
	order := &types.Order{
		ID:          orderID,
		Status:      types.OrderStatusPending,
		Expires:     time.Now().Add(24 * time.Hour), // 24 hour expiry
		Identifiers: req.Identifiers,
		NotBefore:   req.NotBefore,
		NotAfter:    req.NotAfter,
		// TODO: Generate proper authorization URLs
		Authorizations: []string{
			path.Join(baseURL, "authz", "auth-1"),
			path.Join(baseURL, "authz", "auth-2"),
		},
		Finalize: path.Join(baseURL, "order", orderID, "finalize"),
	}

	// TODO: Store order in database

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", path.Join(baseURL, "order", orderID))
	w.WriteHeader(http.StatusCreated)

	// Write response
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Errorf("Failed to encode order response: %v", err)
		writeError(w, newInternalServerError("Failed to encode response"))
		return
	}
}

func GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	log := logger.GetLogger(r.Context())

	// TODO: Retrieve order from database
	log.Infof("Get order request for ID: %s", orderID)
	w.WriteHeader(http.StatusNotImplemented)
}

func FinalizeOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	log := logger.GetLogger(r.Context())

	// Parse request
	var req types.FinalizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Errorf("Failed to decode finalize request: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:malformed",
			Detail: "Failed to parse finalize request",
			Status: http.StatusBadRequest,
		})
		return
	}

	// TODO:
	// 1. Verify order exists and is in valid state
	// 2. Verify CSR matches authorized identifiers
	// 3. Issue certificate
	// 4. Update order status

	log.Infof("Finalize order request for ID: %s", orderID)
	w.WriteHeader(http.StatusNotImplemented)
}
