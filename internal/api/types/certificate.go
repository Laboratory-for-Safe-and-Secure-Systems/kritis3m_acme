package types

// CertificateRequest represents a certificate signing request
type CertificateRequest struct {
	CSR string `json:"csr"` // Base64URL-encoded PKCS#10 CSR
}

// RevocationReason represents the reason for certificate revocation
type RevocationReason int

const (
	RevocationReasonUnspecified          RevocationReason = 0
	RevocationReasonKeyCompromise        RevocationReason = 1
	RevocationReasonCACompromise         RevocationReason = 2
	RevocationReasonAffiliationChanged   RevocationReason = 3
	RevocationReasonSuperseded           RevocationReason = 4
	RevocationReasonCessationOfOperation RevocationReason = 5
)

// RevocationRequest represents a request to revoke a certificate
type RevocationRequest struct {
	Certificate string           `json:"certificate"`   // Base64URL-encoded DER certificate
	Reason      RevocationReason `json:"reason"`        // Reason code for revocation
	JWK         interface{}      `json:"jwk,omitempty"` // JWK for external account auth
}
