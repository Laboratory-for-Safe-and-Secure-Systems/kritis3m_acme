package types

// DirectoryMetadata holds optional metadata for the directory object
type DirectoryMetadata struct {
	TermsOfService          string   `json:"termsOfService,omitempty"`
	Website                 string   `json:"website,omitempty"`
	CAAIdentities           []string `json:"caaIdentities,omitempty"`
	ExternalAccountRequired bool     `json:"externalAccountRequired,omitempty"`
}

// Directory represents the ACME directory object
type Directory struct {
	NewNonce   string             `json:"newNonce"`
	NewAccount string             `json:"newAccount"`
	NewOrder   string             `json:"newOrder"`
	RevokeCert string             `json:"revokeCert"`
	KeyChange  string             `json:"keyChange"`
	Meta       *DirectoryMetadata `json:"meta,omitempty"`
}

// Problem represents an ACME error response
type Problem struct {
	Type        string       `json:"type"`
	Detail      string       `json:"detail"`
	Status      int          `json:"status"`
	Instance    string       `json:"instance,omitempty"`
	Subproblems []Subproblem `json:"subproblems,omitempty"`
}

// Subproblem represents a subproblem in an ACME error response
type Subproblem struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
	Status int    `json:"status"`
}
