package acme

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

const (
	nonceHeader = "Replay-Nonce"
	nonceSize   = 16 // 128 bits of entropy
)

// NonceMiddleware adds a nonce to all responses
func NonceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate new nonce
		nonce, err := generateNonce()
		if err != nil {
			http.Error(w, "Failed to generate nonce", http.StatusInternalServerError)
			return
		}

		// Add nonce to response header
		w.Header().Set(nonceHeader, nonce)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// generateNonce creates a new random nonce
func generateNonce() (string, error) {
	nonceBytes := make([]byte, nonceSize)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(nonceBytes), nil
}
