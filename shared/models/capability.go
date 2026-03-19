package models

import "time"

// CapabilitySource indicates how a capability was registered.
type CapabilitySource string

const (
	CapabilitySourceManifest CapabilitySource = "manifest"
	CapabilitySourceDynamic  CapabilitySource = "dynamic"
)

// CapabilityStatus represents the health state of a capability.
type CapabilityStatus string

const (
	CapabilityStatusHealthy   CapabilityStatus = "healthy"
	CapabilityStatusUnhealthy CapabilityStatus = "unhealthy"
	CapabilityStatusUnknown   CapabilityStatus = "unknown"
)

// ArgumentLocation defines where an argument is passed in an HTTP request.
type ArgumentLocation string

const (
	ArgumentLocationPath   ArgumentLocation = "path"
	ArgumentLocationQuery  ArgumentLocation = "query"
	ArgumentLocationBody   ArgumentLocation = "body"
	ArgumentLocationHeader ArgumentLocation = "header"
)

// ArgumentSpec describes a single capability argument.
type ArgumentSpec struct {
	In      ArgumentLocation `json:"in" yaml:"in"`
	Default any              `json:"default,omitempty" yaml:"default,omitempty"`
}

// ResponseSpec describes how to extract the result from an HTTP response.
type ResponseSpec struct {
	From string `json:"from" yaml:"from"` // "body", "body.field", "header.X-Foo"
}

// EndpointSpec maps a capability to an HTTP endpoint on the application.
type EndpointSpec struct {
	Method    string                  `json:"method" yaml:"method"`
	Path      string                  `json:"path" yaml:"path"`
	Arguments map[string]ArgumentSpec `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	Response  ResponseSpec            `json:"response" yaml:"response"`
}

// CapabilityMetadata is the runtime record of a registered capability.
type CapabilityMetadata struct {
	CapabilityID     string            `json:"capability_id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Tags             []string          `json:"tags,omitempty"`
	Language         string            `json:"language,omitempty"`
	AgentID          string            `json:"agent_id"`
	TenantID         string            `json:"tenant_id,omitempty"`
	Source           CapabilitySource  `json:"source"`
	Idempotent       bool              `json:"idempotent"`
	SideEffects      bool              `json:"side_effects"`
	RequiresApproval bool              `json:"requires_approval"`
	Endpoint         EndpointSpec      `json:"endpoint"`
	Security         SecurityConfig    `json:"security,omitempty"`
	Transport        TransportConfig   `json:"transport,omitempty"`
	Status           CapabilityStatus  `json:"status"`
	RegisteredAt     time.Time         `json:"registered_at"`
	LastHeartbeat    *time.Time        `json:"last_heartbeat,omitempty"`
	ProxyURL         string            `json:"proxy_url,omitempty"` // URL of the proxy serving this capability
	AppPort          int               `json:"app_port,omitempty"`
}
