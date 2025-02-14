package types

import "time"

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusReady      OrderStatus = "ready"
	OrderStatusValid      OrderStatus = "valid"
	OrderStatusInvalid    OrderStatus = "invalid"
	OrderStatusProcessing OrderStatus = "processing"
)

type Order struct {
	ID             string       `json:"id"`
	Status         OrderStatus  `json:"status"`
	Expires        time.Time    `json:"expires"`
	Identifiers    []Identifier `json:"identifiers"`
	NotBefore      *time.Time   `json:"notBefore,omitempty"`
	NotAfter       *time.Time   `json:"notAfter,omitempty"`
	Error          *Problem     `json:"error,omitempty"`
	Authorizations []string     `json:"authorizations"`
	Finalize       string       `json:"finalize"`
	Certificate    string       `json:"certificate,omitempty"`
}

type Identifier struct {
	Type  string `json:"type"`  // "dns" or "ip"
	Value string `json:"value"` // domain name or IP address
}

type OrderRequest struct {
	Identifiers []Identifier `json:"identifiers"`
	NotBefore   *time.Time   `json:"notBefore,omitempty"`
	NotAfter    *time.Time   `json:"notAfter,omitempty"`
}

type FinalizeRequest struct {
	CSR string `json:"csr"` // Base64URL-encoded PKCS#10 CSR
}
