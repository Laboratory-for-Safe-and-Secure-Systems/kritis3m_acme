package types

import (
	"encoding/json"
)

// AccountStatus represents the status of an ACME account
type AccountStatus string

const (
	AccountStatusValid       AccountStatus = "valid"
	AccountStatusDeactivated AccountStatus = "deactivated"
	AccountStatusRevoked     AccountStatus = "revoked"
)

// Account represents an ACME account
type Account struct {
	ID                   string          `json:"id"`
	Key                  json.RawMessage `json:"key"`
	Contact              []string        `json:"contact,omitempty"`
	Status               AccountStatus   `json:"status"`
	TermsOfServiceAgreed bool            `json:"termsOfServiceAgreed"`
	CreatedAt            int64           `json:"createdAt"`
	InitialIP            string          `json:"initialIp"`
	OrdersURL            string          `json:"orders"`
}

// AccountRequest represents the JSON payload for a new-account request
type AccountRequest struct {
	Contact              []string `json:"contact"`
	TermsOfServiceAgreed bool     `json:"termsOfServiceAgreed"`
	OnlyReturnExisting   bool     `json:"onlyReturnExisting"`
}

// Error implements the error interface for Problem
func (p *Problem) Error() string {
	return p.Detail
}
