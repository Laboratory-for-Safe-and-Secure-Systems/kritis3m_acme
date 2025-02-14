package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/go-chi/chi/v5"
)

func GetCertificate(w http.ResponseWriter, r *http.Request) {
	certID := chi.URLParam(r, "certID")
	log := logger.GetLogger(r.Context())

	// TODO:
	// 1. Retrieve certificate from storage
	// 2. Verify requestor has authorization to access certificate
	// 3. Return certificate in proper format (application/pem-certificate-chain)

	log.Infof("Get certificate request for ID: %s", certID)

	// For now, return not implemented
	writeError(w, &types.Problem{
		Type:   "urn:ietf:params:acme:error:serverInternal",
		Detail: "Certificate retrieval not implemented",
		Status: http.StatusNotImplemented,
	})
}

func RevokeCertificate(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger(r.Context())

	// Parse revocation request
	var req types.RevocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Errorf("Failed to decode revocation request: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:malformed",
			Detail: "Failed to parse revocation request",
			Status: http.StatusBadRequest,
		})
		return
	}

	// Validate certificate
	certDER, err := base64.RawURLEncoding.DecodeString(req.Certificate)
	if err != nil {
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:malformed",
			Detail: "Invalid certificate encoding",
			Status: http.StatusBadRequest,
		})
		return
	}

	// TODO:
	// 1. Parse certificate
	// 2. Verify authorization to revoke
	//    - Account must have authorized the original issuance, or
	//    - Request must be signed by the certificate key, or
	//    - Request must be signed by the issuing CA
	// 3. Add certificate to CRL/OCSP
	// 4. Update certificate status in database

	log.Infof("Revoke certificate request received, certificate size: %d bytes", len(certDER))

	// Return success response (empty 200 OK)
	w.WriteHeader(http.StatusOK)
}
