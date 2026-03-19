package models

// PolicyEffect is the outcome of a policy evaluation.
type PolicyEffect string

const (
	PolicyEffectAllow     PolicyEffect = "allow"
	PolicyEffectDeny      PolicyEffect = "deny"
	PolicyEffectRateLimit PolicyEffect = "rate_limit"
)

// PolicyDecision is the result returned by a PolicyEnforcer.
type PolicyDecision struct {
	Effect PolicyEffect `json:"effect"`
	Reason string       `json:"reason,omitempty"`
}
