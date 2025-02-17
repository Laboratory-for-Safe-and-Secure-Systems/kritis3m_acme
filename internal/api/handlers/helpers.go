package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/server"
)

// getBaseURL determines the base URL from the request
func getBaseURL(r *http.Request) string {
	scheme := "http"
	if aslState := r.Context().Value(server.TLSStateKey); aslState != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// generateID creates a unique ID with a prefix
func generateID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// writeError writes a Problem Details error response
func writeError(w http.ResponseWriter, problem *types.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	json.NewEncoder(w).Encode(problem)
}

// writeJSON writes a JSON response with the given status code
func writeJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// setLinkHeader sets the Link header with the given relation
func setLinkHeader(w http.ResponseWriter, url string, rel string) {
	w.Header().Set("Link", fmt.Sprintf(`<%s>;rel="%s"`, url, rel))
}

// endpointURL returns the URL for the <Endpoint> endpoint
func endpointURL(baseURL string, endpoint string, id string) string {
	return fmt.Sprintf("%s/%s/%s", baseURL, endpoint, id)
}

// finalizeURL returns the URL for the finalize endpoint
func finalizeURL(baseURL string, endpoint string, orderID string) string {
	return endpointURL(baseURL, endpoint, orderID) + "/finalize"
}

// generateToken now creates a random 16-byte (32-hex-character) token.
func generateToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback token if random generation fails.
		return "token-123"
	}
	return hex.EncodeToString(b)
}
