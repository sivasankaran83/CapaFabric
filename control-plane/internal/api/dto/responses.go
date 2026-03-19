package dto

import "github.com/sivasankaran83/CapaFabric/shared/models"

// RegisterResponse is returned by POST /api/v1/capabilities/register.
type RegisterResponse struct {
	CapabilityID string `json:"capability_id"`
	Status       string `json:"status"`
}

// DiscoverResponse is returned by POST /api/v1/discover.
type DiscoverResponse struct {
	Capabilities []models.CapabilityMetadata `json:"capabilities"`
	Count        int                         `json:"count"`
}

// InvokeResponse is returned by POST /api/v1/invoke/{capability_id}.
type InvokeResponse struct {
	models.InvocationResult
}

// CapabilityListResponse is returned by GET /api/v1/capabilities.
type CapabilityListResponse struct {
	Capabilities []models.CapabilityMetadata `json:"capabilities"`
	Count        int                         `json:"count"`
}

// HealthResponse is returned by GET /api/v1/health.
type HealthResponse struct {
	Status  string            `json:"status"`
	Version string            `json:"version,omitempty"`
	Checks  map[string]string `json:"checks,omitempty"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}
