package interfaces

import (
	"context"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// AuditRecord is a single immutable audit event.
type AuditRecord struct {
	EventType    string                  `json:"event_type"`
	InvocationContext models.InvocationContext `json:"invocation_context"`
	Result       *models.InvocationResult `json:"result,omitempty"`
	Identity     *models.AuthIdentity     `json:"identity,omitempty"`
	Policy       *models.PolicyDecision   `json:"policy,omitempty"`
	Guardrails   []models.GuardrailResult `json:"guardrails,omitempty"`
}

// AuditWriter appends audit records to a durable store.
type AuditWriter interface {
	// Write appends a record. Implementations must be safe for concurrent use.
	Write(ctx context.Context, record AuditRecord) error
}
