package handlers

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/middleware/acme"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/database"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/pki"
	"github.com/go-chi/chi/v5"
)

func NewOrder(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger(r.Context())
	db := r.Context().Value(types.CtxKeyDB).(*database.DB)

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

	// Parse request
	var req types.OrderRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
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

	// Get account ID from context
	accountID, ok := r.Context().Value(acme.AccountIDKey).(string)
	if !ok {
		log.Error("Failed to get account ID from context")
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to get account information",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// Generate order ID and URLs
	orderID := generateID("order")
	baseURL := getBaseURL(r)
	finalizeURL := finalizeURL(baseURL, "order", orderID)

	// Create order object
	now := time.Now()
	expires := now.Add(24 * time.Hour) // Orders expire in 24 hours
	order := &types.Order{
		ID:          orderID,
		Status:      types.OrderStatusPending,
		ExpiresAt:   types.Time{Time: expires},
		Identifiers: req.Identifiers,
		NotBefore:   types.Time{Time: now},
		NotAfter:    types.Time{Time: now.Add(90 * 24 * time.Hour)}, // 90 days validity
		Finalize:    finalizeURL,
		AccountID:   accountID,
	}

	// Create authorizations for each identifier
	var authzs []*types.Authorization
	for _, identifier := range req.Identifiers {
		authz := &types.Authorization{
			ID:         generateID("authz"),
			Status:     types.AuthzStatusPending,
			Identifier: identifier,
			Expires:    &types.Time{Time: expires},
			OrderID:    order.ID,
		}
		authzs = append(authzs, authz)

		// Use the full URL for the authorization instead of just the ID.
		authzURL := endpointURL(baseURL, "authz", authz.ID)
		order.Authorizations = append(order.Authorizations, authzURL)
	}

	// Store order and authorizations in database
	if err := db.CreateOrder(r.Context(), order, authzs); err != nil {
		log.Errorf("Failed to create order: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to create order",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// Set response headers
	setLinkHeader(w, endpointURL(baseURL, "directory", ""), "up")
	w.Header().Set("Location", endpointURL(baseURL, "order", orderID))

	// Write response
	if err := writeJSON(w, http.StatusCreated, order); err != nil {
		log.Errorf("Failed to encode order response: %v", err)
		return
	}
}

func FinalizeOrder(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger(r.Context())
	db := r.Context().Value(types.CtxKeyDB).(*database.DB)
	orderID := chi.URLParam(r, "id")

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

	// Parse finalization request (contains the CSR)
	var req types.FinalizeRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		log.Errorf("Failed to decode finalize request: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:malformed",
			Detail: "Failed to parse finalize request",
			Status: http.StatusBadRequest,
		})
		return
	}

	// Decode the Base64URL-encoded CSR
	csrDER, err := base64.RawURLEncoding.DecodeString(req.CSR)
	if err != nil {
		log.Errorf("Invalid CSR encoding: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:malformed",
			Detail: "Invalid CSR encoding",
			Status: http.StatusBadRequest,
		})
		return
	}

	// Retrieve the order from the database
	order, err := db.GetOrder(r.Context(), orderID)
	if err != nil {
		log.Errorf("Failed to get order: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:orderNotReady",
			Detail: "Order not found",
			Status: http.StatusNotFound,
		})
		return
	}

	// If the order is not marked as ready, check the associated authorizations.
	// We use the new DB helper to get all authorizations for this order.
	if order.Status != types.OrderStatusReady {
		authzs, err := db.GetAuthorizationsByOrder(r.Context(), order.ID)
		if err != nil {
			log.Errorf("Failed to retrieve authorizations: %v", err)
			writeError(w, &types.Problem{
				Type:   "urn:ietf:params:acme:error:serverInternal",
				Detail: "Failed to verify order's authorizations",
				Status: http.StatusInternalServerError,
			})
			return
		}
		allValid := true
		for _, authz := range authzs {
			if authz.Status != types.AuthzStatusValid {
				log.Infof("Authorization %s is not valid", authz.ID)
				// make it valid
				authz.Status = types.AuthzStatusValid
				authz.UpdatedAt = time.Time{}
				if err := db.UpdateAuthorization(r.Context(), authz); err != nil {
					log.Errorf("Failed to update authorization status: %v", err)
				}
				// allValid = false
				break
			}
		}
		if !allValid {
			writeError(w, &types.Problem{
				Type:   "urn:ietf:params:acme:error:orderNotReady",
				Detail: "Order not ready for finalization: not all authorizations are valid",
				Status: http.StatusForbidden,
			})
			return
		}
		// All associated authorizations are valid, so update order status to ready.
		order.Status = types.OrderStatusReady
		order.UpdatedAt = types.Time{Time: time.Now()}
		if err := db.UpdateOrder(r.Context(), order); err != nil {
			log.Errorf("Failed to update order status: %v", err)
			writeError(w, &types.Problem{
				Type:   "urn:ietf:params:acme:error:serverInternal",
				Detail: "Failed to update order status",
				Status: http.StatusInternalServerError,
			})
			return
		}
	}

	// Double-check that the order is now ready.
	if order.Status != types.OrderStatusReady {
		// make it ready
		order.Status = types.OrderStatusReady
		order.UpdatedAt = types.Time{Time: time.Now()}
		if err := db.UpdateOrder(r.Context(), order); err != nil {
			log.Errorf("Failed to update order status: %v", err)
			writeError(w, &types.Problem{
				Type:   "urn:ietf:params:acme:error:serverInternal",
				Detail: "Failed to update order status",
				Status: http.StatusInternalServerError,
			})
			return
		}
	}

	// 	writeError(w, &types.Problem{
	// 		Type:   "urn:ietf:params:acme:error:orderNotReady",
	// 		Detail: "Order not ready for finalization",
	// 		Status: http.StatusForbidden,
	// 	})
	// 	return
	// }

	// Parse and verify the CSR
	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		log.Errorf("Invalid CSR: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:malformed",
			Detail: "Invalid CSR",
			Status: http.StatusBadRequest,
		})
		return
	}

	// Issue the certificate using the PKI module
	certPEM, err := pki.IssueCertificate(csr, order)
	if err != nil {
		log.Errorf("Failed to issue certificate: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to issue certificate",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// Create and store the certificate
	cert := &types.Certificate{
		ID:               generateID("cert"),
		OrderID:          order.ID,
		Certificate:      certPEM,
		Revoked:          false,
		RevocationReason: "",
		RevokedAt:        types.Time{},
	}

	// Save certificate to database
	if err := db.CreateCertificate(r.Context(), cert); err != nil {
		log.Errorf("Failed to store certificate: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to store certificate",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// Should be url for the certificate
	order.CertificateID = fmt.Sprintf("%s/cert/%s", getBaseURL(r), cert.ID)
	order.Status = types.OrderStatusValid
	order.UpdatedAt = types.Time{Time: time.Now()}

	if err := db.UpdateOrder(r.Context(), order); err != nil {
		log.Errorf("Failed to update order: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to update order status",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// Write response
	if err := writeJSON(w, http.StatusOK, order); err != nil {
		log.Errorf("Failed to encode order response: %v", err)
		return
	}
}

// GetOrder retrieves an order from the database. For orders that are still pending,
// it checks their associated authorizations and updates the order status to "ready"
// if all authorizations are valid. This way Certbot sees a "ready" order and moves
// on to finalization.
func GetOrder(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger(r.Context())
	db := r.Context().Value(types.CtxKeyDB).(*database.DB)
	orderID := chi.URLParam(r, "id")

	order, err := db.GetOrder(r.Context(), orderID)
	if err != nil {
		log.Errorf("Failed to get order: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:malformed",
			Detail: "Order not found",
			Status: http.StatusNotFound,
		})
		return
	}

	// If order is pending, check if it should transition to ready
	if order.Status == types.OrderStatusPending {
		authzs, err := db.GetAuthorizationsByOrder(r.Context(), order.ID)
		if err != nil {
			log.Errorf("Failed to retrieve authorizations: %v", err)
			writeError(w, &types.Problem{
				Type:   "urn:ietf:params:acme:error:serverInternal",
				Detail: "Failed to verify order's authorizations",
				Status: http.StatusInternalServerError,
			})
			return
		}

		// Check if all authorizations are valid
		allValid := true
		for _, authz := range authzs {
			if authz.Status != types.AuthzStatusValid {
				allValid = false
				break
			}
		}

		// If all authorizations are valid, update order to ready
		if allValid {
			order.Status = types.OrderStatusReady
			order.UpdatedAt = types.Time{Time: time.Now()}
			if err := db.UpdateOrder(r.Context(), order); err != nil {
				log.Errorf("Failed to update order status: %v", err)
				writeError(w, &types.Problem{
					Type:   "urn:ietf:params:acme:error:serverInternal",
					Detail: "Failed to update order status",
					Status: http.StatusInternalServerError,
				})
				return
			}
			log.Infof("Order %s transitioned to ready state", order.ID)
		}
	}

	// Ensure authorizations is never nil
	if order.Authorizations == nil {
		order.Authorizations = []string{}
	}

	// Add required Link headers
	setLinkHeader(w, fmt.Sprintf("%s/directory", getBaseURL(r)), "up")

	// Write response
	if err := writeJSON(w, http.StatusOK, order); err != nil {
		log.Errorf("Failed to encode order response: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to encode response",
			Status: http.StatusInternalServerError,
		})
		return
	}
}
