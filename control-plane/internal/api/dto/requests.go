package dto

import "github.com/sivasankaran83/CapaFabric/shared/models"

// RegisterRequest is the body for POST /api/v1/capabilities/register.
type RegisterRequest struct {
	CapabilityID     string                  `json:"capability_id"`
	Name             string                  `json:"name"`
	Description      string                  `json:"description"`
	Tags             []string                `json:"tags,omitempty"`
	Language         string                  `json:"language,omitempty"`
	AgentID          string                  `json:"agent_id"`
	TenantID         string                  `json:"tenant_id,omitempty"`
	Idempotent       bool                    `json:"idempotent"`
	SideEffects      bool                    `json:"side_effects"`
	RequiresApproval bool                    `json:"requires_approval,omitempty"`
	Endpoint         models.EndpointSpec     `json:"endpoint"`
	Security         models.SecurityConfig   `json:"security,omitempty"`
	Transport        models.TransportConfig  `json:"transport,omitempty"`
	ProxyURL         string                  `json:"proxy_url,omitempty"`
	AppPort          int                     `json:"app_port,omitempty"`
}

// DiscoverRequest is the body for POST /api/v1/discover.
type DiscoverRequest struct {
	Goal     string   `json:"goal,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	TenantID string   `json:"tenant_id,omitempty"`
	Provider string   `json:"provider,omitempty"`
	MaxTools int      `json:"max_tools,omitempty"`
}

// InvokeRequest is the body for POST /api/v1/invoke/{capability_id}.
type InvokeRequest struct {
	Arguments map[string]any `json:"arguments,omitempty"`
	CallerID  string         `json:"caller_id,omitempty"`
	GoalID    string         `json:"goal_id,omitempty"`
	TenantID  string         `json:"tenant_id,omitempty"`
	CallChain []string       `json:"call_chain,omitempty"`
}

// HeartbeatRequest is the body for POST /api/v1/heartbeat.
type HeartbeatRequest struct {
	AgentID      string   `json:"agent_id"`
	CapabilityIDs []string `json:"capability_ids"`
}
