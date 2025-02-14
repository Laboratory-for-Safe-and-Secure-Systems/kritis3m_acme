package types

// AccountStatus represents the status of an ACME account
type AccountStatus string

const (
	AccountStatusValid       AccountStatus = "valid"
	AccountStatusDeactivated AccountStatus = "deactivated"
	AccountStatusRevoked     AccountStatus = "revoked"
)

// Account represents an ACME account
type Account struct {
	ID                   string        `json:"id"`
	Key                  interface{}   `json:"key"`     // JWK
	Contact              []string      `json:"contact"` // Email addresses
	Status               AccountStatus `json:"status"`
	TermsOfServiceAgreed bool          `json:"termsOfServiceAgreed"`
	Orders               string        `json:"orders"` // URL of orders list
	CreatedAt            int64         `json:"createdAt"`
	InitialIP            string        `json:"initialIp"`
}

// AccountRequest represents the JSON payload for a new-account request
type AccountRequest struct {
	Contact              []string `json:"contact"`
	TermsOfServiceAgreed bool     `json:"termsOfServiceAgreed"`
	OnlyReturnExisting   bool     `json:"onlyReturnExisting"`
}
