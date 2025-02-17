package types

import "time"

type AuthorizationStatus string

const (
	AuthzStatusPending     AuthorizationStatus = "pending"
	AuthzStatusValid       AuthorizationStatus = "valid"
	AuthzStatusInvalid     AuthorizationStatus = "invalid"
	AuthzStatusDeactivated AuthorizationStatus = "deactivated"
	AuthzStatusExpired     AuthorizationStatus = "expired"
	AuthzStatusRevoked     AuthorizationStatus = "revoked"
)

type Authorization struct {
	ID         string              `json:"id"`
	Status     AuthorizationStatus `json:"status"`
	Identifier Identifier          `json:"identifier"`
	Expires    *Time               `json:"expires"`
	Challenges []Challenge         `json:"challenges"`
	OrderID    string              `json:"orderId"`
	Wildcard   bool                `json:"wildcard"`
	CreatedAt  time.Time           `json:"createdAt"`
	UpdatedAt  time.Time           `json:"updatedAt"`
}

type Challenge struct {
	ID              string          `json:"id"`
	AuthorizationID string          `json:"-"` // Foreign key to Authorization
	Type            string          `json:"type"`
	URL             string          `json:"url"`
	Status          ChallengeStatus `json:"status"`
	Token           string          `json:"token"`
	Validated       *Time           `json:"validated,omitempty"`
}

type ChallengeStatus string

const (
	ChallengeStatusPending    ChallengeStatus = "pending"
	ChallengeStatusProcessing ChallengeStatus = "processing"
	ChallengeStatusValid      ChallengeStatus = "valid"
	ChallengeStatusInvalid    ChallengeStatus = "invalid"
)
