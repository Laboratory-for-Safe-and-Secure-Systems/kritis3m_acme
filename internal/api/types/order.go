package types

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
	AccountID      string       `json:"accountId" db:"account_id"`
	Status         OrderStatus  `json:"status"`
	ExpiresAt      Time         `json:"expires" db:"expires_at"`
	NotBefore      Time         `json:"notBefore,omitempty" db:"not_before"`
	NotAfter       Time         `json:"notAfter,omitempty" db:"not_after"`
	Identifiers    []Identifier `json:"identifiers"`
	Finalize       string       `json:"finalize"`
	Error          *Problem     `json:"error,omitempty"`
	CertificateID  string       `json:"certificate,omitempty" db:"certificate_id"`
	Authorizations []string     `json:"authorizations"`
	CreatedAt      Time         `json:"createdAt" db:"created_at"`
	UpdatedAt      Time         `json:"updatedAt" db:"updated_at"`
}

type Identifier struct {
	Type  string `json:"type"`  // "dns" or "ip"
	Value string `json:"value"` // domain name or IP address
}

type OrderRequest struct {
	Identifiers []Identifier `json:"identifiers"`
	NotBefore   Time         `json:"notBefore,omitempty"`
	NotAfter    Time         `json:"notAfter,omitempty"`
}

type FinalizeRequest struct {
	CSR string `json:"csr"` // Base64URL-encoded PKCS#10 CSR
}

type Certificate struct {
	ID               string `json:"id"`
	OrderID          string `json:"orderId"`
	Certificate      string `json:"certificate"` // PEM encoded certificate
	Revoked          bool   `json:"revoked"`
	RevocationReason string `json:"revocationReason,omitempty"`
	RevokedAt        Time   `json:"revokedAt,omitempty"`
	CreatedAt        Time   `json:"createdAt"`
}
