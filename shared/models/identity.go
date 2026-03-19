package models

// AuthIdentity represents an authenticated caller.
type AuthIdentity struct {
	Subject  string            `json:"subject"`
	TenantID string            `json:"tenant_id,omitempty"`
	Roles    []string          `json:"roles,omitempty"`
	Claims   map[string]string `json:"claims,omitempty"`
	Provider string            `json:"provider"` // "jwt" | "mtls" | "apikey" | "none"
}
