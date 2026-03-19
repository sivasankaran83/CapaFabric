package models

// GuardrailAction is the action taken when a guardrail triggers.
type GuardrailAction string

const (
	GuardrailActionAllow  GuardrailAction = "allow"
	GuardrailActionBlock  GuardrailAction = "block"
	GuardrailActionRedact GuardrailAction = "redact"
	GuardrailActionLog    GuardrailAction = "log"
)

// GuardrailType identifies which guardrail check was run.
type GuardrailType string

const (
	GuardrailTypePromptInjection GuardrailType = "prompt_injection"
	GuardrailTypePIIRedaction    GuardrailType = "pii_redaction"
	GuardrailTypeTopicRestrict   GuardrailType = "topic_restriction"
	GuardrailTypeOutputSafety    GuardrailType = "output_safety"
	GuardrailTypeSchemaCompliance GuardrailType = "schema_compliance"
	GuardrailTypeDataLeakage     GuardrailType = "data_leakage"
)

// GuardrailResult is the outcome of a single guardrail check.
type GuardrailResult struct {
	Type        GuardrailType   `json:"type"`
	Action      GuardrailAction `json:"action"`
	Triggered   bool            `json:"triggered"`
	Reason      string          `json:"reason,omitempty"`
	ModifiedContent string      `json:"modified_content,omitempty"` // populated when action == redact
}
