package handlers

import (
	"encoding/json"
	"net/http"
)

type directoryResponse struct {
	NewNonce   string `json:"newNonce"`
	NewAccount string `json:"newAccount"`
	NewOrder   string `json:"newOrder"`
	RevokeCert string `json:"revokeCert"`
	KeyChange  string `json:"keyChange"`
}

func GetDirectory(w http.ResponseWriter, r *http.Request) {
	baseURL := "http://" + r.Host

	dir := directoryResponse{
		NewNonce:   baseURL + "/new-nonce",
		NewAccount: baseURL + "/new-account",
		NewOrder:   baseURL + "/new-order",
		RevokeCert: baseURL + "/revoke-cert",
		KeyChange:  baseURL + "/key-change",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dir)
}

func NewNonce(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

func NewAccount(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement account creation
	w.WriteHeader(http.StatusNotImplemented)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
