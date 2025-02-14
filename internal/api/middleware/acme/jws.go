package acme

import (
	"net/http"
)

// JWSVerificationMiddleware verifies the JWS signature of ACME requests
func JWSVerificationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement JWS verification
		next.ServeHTTP(w, r)
	})
}
