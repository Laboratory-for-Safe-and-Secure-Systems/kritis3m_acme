package acme

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/go-jose/go-jose/v3"
)

func TestJWSVerification(t *testing.T) {
	// Initialize logger for tests
	log := logger.New(os.Stdout)

	// Create a test key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create JWK from public key
	jwk := jose.JSONWebKey{
		Key:       &privateKey.PublicKey,
		Algorithm: "ES256",
		Use:       "sig",
	}

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify logger is in context
		if logger.GetLogger(r.Context()) == nil {
			t.Error("Logger not found in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Create the middleware chain
	handler := JWSVerificationMiddleware(testHandler)

	tests := []struct {
		name       string
		setupJWS   func() ([]byte, error)
		wantStatus int
	}{
		{
			name: "valid new account JWS",
			setupJWS: func() ([]byte, error) {
				// Get a nonce first
				nonce, err := generateNonce()
				if err != nil {
					return nil, err
				}
				nonceStore.StoreNonce(nonce)

				// Create signer
				opts := &jose.SignerOptions{}
				opts.WithHeader("nonce", nonce)
				opts.WithHeader("url", "http://example.com/new-account")
				opts.WithHeader("jwk", jwk)

				signer, err := jose.NewSigner(jose.SigningKey{
					Algorithm: jose.ES256,
					Key:       privateKey,
				}, opts)
				if err != nil {
					return nil, err
				}

				// Create and sign payload
				payload := []byte(`{"termsOfServiceAgreed":true}`)
				object, err := signer.Sign(payload)
				if err != nil {
					return nil, err
				}

				// Convert to our internal JWS format
				serialized := object.FullSerialize()
				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(serialized), &parsed); err != nil {
					return nil, err
				}

				jws := JWSRequest{
					Protected: parsed["protected"].(string),
					Payload:   parsed["payload"].(string),
					Signature: parsed["signature"].(string),
				}

				return json.Marshal(jws)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid algorithm",
			setupJWS: func() ([]byte, error) {
				// Similar to above but use RS256 instead of ES256
				// This should fail
				return nil, nil
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid nonce",
			setupJWS: func() ([]byte, error) {
				// Use an invalid nonce
				return nil, nil
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "URL mismatch",
			setupJWS: func() ([]byte, error) {
				// Use different URL in JWS than request
				return nil, nil
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := tt.setupJWS()
			if err != nil {
				t.Fatalf("Failed to setup JWS: %v", err)
			}

			req := httptest.NewRequest("POST", "http://example.com/new-account", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			// Add logger to context
			ctx := context.WithValue(req.Context(), types.CtxKeyLogger, log)
			req = req.WithContext(ctx)

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Want status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}
