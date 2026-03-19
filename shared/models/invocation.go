package models

import "time"

// InvocationContext carries request-scoped metadata through the pipeline.
type InvocationContext struct {
	RequestID    string            `json:"request_id"`
	GoalID       string            `json:"goal_id,omitempty"`
	CallerID     string            `json:"caller_id,omitempty"`
	TenantID     string            `json:"tenant_id,omitempty"`
	CallChain    []string          `json:"call_chain,omitempty"`
	Depth        int               `json:"depth"`
	MaxDepth     int               `json:"max_depth"`
	Arguments    map[string]any    `json:"arguments,omitempty"`
	TraceID      string            `json:"trace_id,omitempty"`
	SpanID       string            `json:"span_id,omitempty"`
	StartedAt    time.Time         `json:"started_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// InvocationResult holds the outcome of a capability invocation.
type InvocationResult struct {
	RequestID    string        `json:"request_id"`
	CapabilityID string        `json:"capability_id"`
	Success      bool          `json:"success"`
	Result       any           `json:"result,omitempty"`
	Error        *InvokeError  `json:"error,omitempty"`
	DurationMS   int64         `json:"duration_ms"`
	TraceID      string        `json:"trace_id,omitempty"`
}

// InvokeError describes a failure during invocation.
type InvokeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}
