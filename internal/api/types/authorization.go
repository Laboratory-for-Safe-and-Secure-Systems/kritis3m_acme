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
	Status     AuthorizationStatus `json:"status"`
	Expires    *Time               `json:"expires,omitempty"`
	Identifier Identifier          `json:"identifier"`
	Challenges []Challenge         `json:"challenges"`
	Wildcard   *bool               `json:"wildcard,omitempty"`
}

type Challenge struct {
	Type      string          `json:"type"`
	URL       string          `json:"url"`
	Status    ChallengeStatus `json:"status"`
	Validated *Time           `json:"validated,omitempty"`
	Error     *Problem        `json:"error,omitempty"`
	Token     string          `json:"token"`
}

type ChallengeStatus string

const (
	ChallengeStatusPending    ChallengeStatus = "pending"
	ChallengeStatusProcessing ChallengeStatus = "processing"
	ChallengeStatusValid      ChallengeStatus = "valid"
	ChallengeStatusInvalid    ChallengeStatus = "invalid"
)

// Time is a wrapper around time.Time that serializes to RFC3339
type Time struct {
	time.Time
}
