package acme

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
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

		// Store the nonce
		if err := nonceStore.StoreNonce(nonce); err != nil {
			http.Error(w, "Failed to store nonce", http.StatusInternalServerError)
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

// Add after the imports
type NonceStore interface {
	StoreNonce(nonce string) error
	ValidateAndRemove(nonce string) error
	CleanExpired() error
}

type MemoryNonceStore struct {
	nonces map[string]*nonceEntry
	mu     sync.RWMutex
}

type nonceEntry struct {
	used      bool
	createdAt time.Time
}

func NewMemoryNonceStore() *MemoryNonceStore {
	store := &MemoryNonceStore{
		nonces: make(map[string]*nonceEntry),
	}
	// Start cleanup goroutine
	go store.periodicCleanup()
	return store
}

func (s *MemoryNonceStore) StoreNonce(nonce string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a default logger if none exists
	log := logger.New(os.Stdout)
	log.Debugw("Storing nonce", "nonce", nonce)

	s.nonces[nonce] = &nonceEntry{
		used:      false,
		createdAt: time.Now(),
	}
	return nil
}

func (s *MemoryNonceStore) ValidateAndRemove(nonce string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log := logger.New(os.Stdout)
	log.Debugw("Validating nonce", "nonce", nonce)

	entry, exists := s.nonces[nonce]
	if !exists {
		log.Errorw("Nonce not found", "nonce", nonce)
		return fmt.Errorf("invalid nonce")
	}
	if entry.used {
		log.Errorw("Nonce already used", "nonce", nonce)
		return fmt.Errorf("nonce already used")
	}

	// Check if nonce has expired (15 minutes max age)
	if time.Since(entry.createdAt) > 15*time.Minute {
		delete(s.nonces, nonce)
		return fmt.Errorf("nonce has expired")
	}

	// Mark as used and schedule for deletion
	entry.used = true
	return nil
}

func (s *MemoryNonceStore) CleanExpired() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for nonce, entry := range s.nonces {
		// Remove if used or older than 15 minutes
		if entry.used || now.Sub(entry.createdAt) > 15*time.Minute {
			delete(s.nonces, nonce)
		}
	}
	return nil
}

func (s *MemoryNonceStore) periodicCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		s.CleanExpired()
	}
}
