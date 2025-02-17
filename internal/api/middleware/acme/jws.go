package acme

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/database"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/go-jose/go-jose/v3"
)

// JWSHeader represents the protected header of a JWS
type JWSHeader struct {
	Alg   string      `json:"alg"`           // The algorithm used for the signature
	Nonce string      `json:"nonce"`         // The nonce value
	URL   string      `json:"url"`           // The request URL
	Kid   string      `json:"kid,omitempty"` // Key identifier (for existing accounts)
	Jwk   interface{} `json:"jwk,omitempty"` // JWK (for new accounts)
}

// JWSRequest represents a JWS request structure
type JWSRequest struct {
	Protected string `json:"protected"` // Base64URL(UTF8(JWS Protected Header))
	Payload   string `json:"payload"`   // Base64URL(JWS Payload)
	Signature string `json:"signature"` // Base64URL(JWS Signature)
}

// contextKey type for request context keys
type contextKey string

const (
	// Context keys for storing verified data
	jwsPayloadKey     contextKey = "jws_payload"
	JwsProtectedKey   contextKey = "jws_protected"
	AccountIDKey      contextKey = "account_id"
	DecodedPayloadKey contextKey = "decoded_payload"

	// Error types
	errMalformed    = "urn:ietf:params:acme:error:malformed"
	errUnauthorized = "urn:ietf:params:acme:error:unauthorized"
	errBadNonce     = "urn:ietf:params:acme:error:badNonce"

	// Add constant at the top
	maxRequestSize = 1 << 20 // 1MB
)

// Add a simple in-memory nonce store (in production, use Redis or similar)
var (
	nonceStore = NewMemoryNonceStore()
)

// JWSVerificationMiddleware verifies the JWS signature of ACME requests
func JWSVerificationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger(r.Context())

		// Only verify POST requests
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Read and parse the JWS request
		body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestSize))
		if err != nil {
			log.Errorw("Failed to read request body",
				"error", err,
				"remote_addr", r.RemoteAddr,
				"content_length", r.ContentLength,
			)
			writeError(w, errMalformed, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Create a new reader with the body for subsequent handlers
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		var jws JWSRequest
		if err := json.Unmarshal(body, &jws); err != nil {
			log.Errorf("Failed to parse JWS: %v", err)
			writeError(w, errMalformed, "Invalid JWS format", http.StatusBadRequest)
			return
		}

		// Decode and verify the protected header
		protected, err := decodeJWSProtected(jws.Protected)
		if err != nil {
			log.Errorf("Failed to decode protected header: %v", err)
			writeError(w, errMalformed, "Invalid JWS protected header", http.StatusBadRequest)
			return
		}

		// Verify the nonce
		if err := verifyNonce(protected.Nonce); err != nil {
			log.Errorf("Nonce verification failed: %v", err)
			writeError(w, errBadNonce, "Invalid or missing nonce", http.StatusBadRequest)
			return
		}

		// Verify URL matches the request URL
		if err := verifyRequestURL(r, protected.URL); err != nil {
			log.Errorf("URL verification failed: %v", err)
			writeError(w, errMalformed, "JWS URL does not match request URL", http.StatusBadRequest)
			return
		}

		// Decode the payload and add it to the context
		payloadBytes, err := base64.RawURLEncoding.DecodeString(jws.Payload)
		if err != nil {
			log.Errorf("Invalid payload encoding: %v", err)
			writeError(w, errMalformed, "Invalid payload encoding", http.StatusBadRequest)
			return
		}

		// Verify signature based on key type (JWK or KID)
		if protected.Kid != "" {
			// Existing account - verify using stored public key
			if err := verifyExistingAccount(protected.Kid, jws, r); err != nil {
				log.Errorf("Account verification failed: %v", err)
				writeError(w, errUnauthorized, "Invalid account or signature", http.StatusUnauthorized)
				return
			}
		} else if protected.Jwk != nil {
			// New account - verify using provided JWK
			if err := verifyNewAccount(protected.Jwk, jws); err != nil {
				log.Errorf("JWK verification failed: %v", err)
				writeError(w, errMalformed, "Invalid JWK or signature", http.StatusBadRequest)
				return
			}
		} else {
			writeError(w, errMalformed, "Either 'kid' or 'jwk' must be present", http.StatusBadRequest)
			return
		}

		// Add both raw and decoded payload to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, jwsPayloadKey, jws.Payload)
		ctx = context.WithValue(ctx, DecodedPayloadKey, payloadBytes)
		ctx = context.WithValue(ctx, JwsProtectedKey, protected)

		// Call the next handler with our new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper functions to be implemented:

func writeError(w http.ResponseWriter, typ, detail string, status int) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"type":   typ,
		"detail": detail,
		"status": status,
	})
}

func decodeJWSProtected(protected string) (*JWSHeader, error) {
	// Decode base64url
	decoded, err := base64.RawURLEncoding.DecodeString(protected)
	if err != nil {
		return nil, fmt.Errorf("failed to decode protected header: %w", err)
	}

	// Parse JSON
	var header JWSHeader
	if err := json.Unmarshal(decoded, &header); err != nil {
		return nil, fmt.Errorf("failed to parse protected header: %w", err)
	}

	// Validate required fields
	if header.Alg == "" {
		return nil, fmt.Errorf("missing 'alg' in protected header")
	}
	if header.Nonce == "" {
		return nil, fmt.Errorf("missing 'nonce' in protected header")
	}
	if header.URL == "" {
		return nil, fmt.Errorf("missing 'url' in protected header")
	}
	if header.Kid == "" && header.Jwk == nil {
		return nil, fmt.Errorf("either 'kid' or 'jwk' must be present")
	}

	allowedAlgs := map[string]bool{
		"ES256": true, // Required by RFC 8555
		"ES384": true, // Recommended
		"ES512": true, // Recommended
		"RS256": true, // Recommended
	}
	if !allowedAlgs[header.Alg] {
		return nil, fmt.Errorf("algorithm %q not supported, must be one of: ES256, ES384, ES512", header.Alg)
	}

	return &header, nil
}

func verifyNonce(nonce string) error {
	return nonceStore.ValidateAndRemove(nonce)
}

func verifyRequestURL(r *http.Request, jwsURL string) error {
	// Parse the JWS URL
	parsedJWSURL, err := url.Parse(jwsURL)
	if err != nil {
		return fmt.Errorf("invalid JWS URL: %w", err)
	}

	// Compare only the path and host parts
	if !strings.EqualFold(parsedJWSURL.Path, r.URL.Path) {
		return fmt.Errorf("URL path mismatch: expected %s, got %s", r.URL.Path, parsedJWSURL.Path)
	}

	// Extract host without port
	jwsHost := parsedJWSURL.Hostname()
	reqHost := r.Host
	if host, _, err := net.SplitHostPort(r.Host); err == nil {
		reqHost = host
	}

	if !strings.EqualFold(jwsHost, reqHost) {
		return fmt.Errorf("host mismatch: expected %s, got %s", reqHost, jwsHost)
	}

	return nil
}

func verifyExistingAccount(kid string, jws JWSRequest, r *http.Request) error {
	log := logger.GetLogger(r.Context())
	log.Debugw("Verifying existing account", "kid", kid)

	// Parse the kid URL
	kidURL, err := url.Parse(kid)
	if err != nil {
		log.Errorw("Failed to parse kid URL", "kid", kid, "error", err)
		return fmt.Errorf("invalid kid URL: %w", err)
	}

	// Extract account ID from path
	parts := strings.Split(strings.Trim(kidURL.Path, "/"), "/")
	if len(parts) != 2 || parts[0] != "account" {
		log.Errorw("Invalid account URL format",
			"kid", kid,
			"parts", parts,
			"path", kidURL.Path)
		return fmt.Errorf("invalid account URL format: %s", kid)
	}
	accountID := parts[1]
	log.Debugw("Extracted account ID", "accountID", accountID)

	// Get the account from database using accountID
	db := r.Context().Value(types.CtxKeyDB).(*database.DB)
	account, err := db.GetAccount(r.Context(), accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Parse the stored public key
	var publicKey jose.JSONWebKey
	if err := json.Unmarshal(account.Key, &publicKey); err != nil {
		return fmt.Errorf("failed to parse stored public key: %w", err)
	}

	// Parse the JWS using proper JSON structure
	rawJWS := map[string]string{
		"protected": jws.Protected,
		"payload":   jws.Payload,
		"signature": jws.Signature,
	}
	rawJWSBytes, err := json.Marshal(rawJWS)
	if err != nil {
		return fmt.Errorf("failed to marshal JWS: %w", err)
	}

	sig, err := jose.ParseSigned(string(rawJWSBytes))
	if err != nil {
		return fmt.Errorf("failed to parse JWS: %w", err)
	}

	// Verify the signature using the public key
	_, err = sig.Verify(&publicKey)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// Add account ID to context for later use
	ctx := context.WithValue(r.Context(), AccountIDKey, account.ID)
	*r = *r.WithContext(ctx)

	return nil
}

func verifyNewAccount(jwk interface{}, jws JWSRequest) error {
	// Parse the JWS using proper JSON structure
	rawJWS := map[string]string{
		"protected": jws.Protected,
		"payload":   jws.Payload,
		"signature": jws.Signature,
	}
	rawJWSBytes, err := json.Marshal(rawJWS)
	if err != nil {
		return fmt.Errorf("failed to marshal JWS: %w", err)
	}
	sig, err := jose.ParseSigned(string(rawJWSBytes))
	if err != nil {
		return fmt.Errorf("failed to parse JWS: %w", err)
	}

	// Parse the JWK properly
	jwkBytes, err := json.Marshal(jwk)
	if err != nil {
		return fmt.Errorf("failed to marshal JWK: %w", err)
	}

	var key jose.JSONWebKey
	if err := json.Unmarshal(jwkBytes, &key); err != nil {
		return fmt.Errorf("invalid JWK format: %w", err)
	}

	// Verify the signature
	_, err = sig.Verify(&key)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}
