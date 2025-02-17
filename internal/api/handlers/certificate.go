package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/go-chi/chi/v5"
)

func GetCertificate(w http.ResponseWriter, r *http.Request) {

	certID := chi.URLParam(r, "certID")
	log := logger.GetLogger(r.Context())

	// TODO:
	// 1. Retrieve certificate from storage if needed.
	// 2. Verify that the requestor is authorized to access the certificate.

	log.Infof("Get certificate request for ID: %s", certID)

	// --- Load CA Certificate Chain ---
	caChainPath := "/home/ayham/repo/kritis3m_workspace/certificates/pki/intermediate/ca_chain.pem"
	caChainBytes, err := os.ReadFile(caChainPath)
	if err != nil {
		log.Errorf("Failed to read CA chain file: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to read CA certificate chain",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// Assume the first certificate in the chain is the CA certificate.
	var caCert *x509.Certificate
	var block *pem.Block
	rest := caChainBytes
	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			caCert, err = x509.ParseCertificate(block.Bytes)
			if err != nil {
				log.Errorf("Failed to parse CA certificate: %v", err)
				writeError(w, &types.Problem{
					Type:   "urn:ietf:params:acme:error:serverInternal",
					Detail: "Failed to parse CA certificate",
					Status: http.StatusInternalServerError,
				})
				return
			}
			break
		}
	}
	if caCert == nil {
		log.Errorf("No CA certificate found in chain")
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "No CA certificate found in chain",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// --- Load CA Private Key ---
	caKeyPath := "/home/ayham/repo/kritis3m_workspace/certificates/pki/intermediate/privateKey.pem"
	caKeyBytes, err := os.ReadFile(caKeyPath)
	if err != nil {
		log.Errorf("Failed to read CA key file: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to read CA private key",
			Status: http.StatusInternalServerError,
		})
		return
	}

	block, _ = pem.Decode(caKeyBytes)
	if block == nil {
		log.Errorf("Failed to decode CA private key PEM")
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to decode CA private key",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// Try parsing as PKCS1; if that fails, try PKCS8.
	var caPrivateKey interface{}
	caPrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		caPrivateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			log.Errorf("Failed to parse CA private key: %v", err)
			writeError(w, &types.Problem{
				Type:   "urn:ietf:params:acme:error:serverInternal",
				Detail: "Failed to parse CA private key",
				Status: http.StatusInternalServerError,
			})
			return
		}
	}

	// --- Create a Certificate Template for "10.10.10.10" ---
	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1), // In production, use a proper random serial number.
		Subject: pkix.Name{
			CommonName: "10.10.10.10",
			// You can add Organization, Country, etc. if needed.
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for one year.
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// --- Generate a Key Pair for the New Certificate ---
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Errorf("Failed to generate key pair: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to generate key pair",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// --- Create the Certificate Signed by the CA ---
	certDER, err := x509.CreateCertificate(rand.Reader, certTemplate, caCert, &privKey.PublicKey, caPrivateKey)
	if err != nil {
		log.Errorf("Failed to create certificate: %v", err)
		writeError(w, &types.Problem{
			Type:   "urn:ietf:params:acme:error:serverInternal",
			Detail: "Failed to create certificate",
			Status: http.StatusInternalServerError,
		})
		return
	}

	// --- Encode the Certificate to PEM Format ---
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	// Optionally, append the CA chain to form a complete chain.
	fullChain := append(certPEM, caChainBytes...)

	// --- Return the PEM Certificate Chain in the Response ---
	w.Header().Set("Content-Type", "application/pem-certificate-chain")
	w.WriteHeader(http.StatusOK)
	w.Write(fullChain)
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
