package models

// AuditLevel controls the granularity of audit trail for a capability invocation.
type AuditLevel string

const (
	AuditLevelNone     AuditLevel = "none"
	AuditLevelBasic    AuditLevel = "basic"
	AuditLevelForensic AuditLevel = "forensic"
)

// ClassificationLevel indicates the data sensitivity level of a capability.
type ClassificationLevel string

const (
	ClassificationPublic       ClassificationLevel = "public"
	ClassificationInternal     ClassificationLevel = "internal"
	ClassificationConfidential ClassificationLevel = "confidential"
	ClassificationRestricted   ClassificationLevel = "restricted"
)

// SecurityConfig declares access control requirements for a capability.
type SecurityConfig struct {
	RequiredRoles       []string            `json:"required_roles,omitempty" yaml:"required_roles,omitempty"`
	Classification      ClassificationLevel `json:"classification,omitempty" yaml:"classification,omitempty"`
	AuditLevel          AuditLevel          `json:"audit_level,omitempty" yaml:"audit_level,omitempty"`
	MaxCallsPerMinute   int                 `json:"max_calls_per_minute,omitempty" yaml:"max_calls_per_minute,omitempty"`
}
