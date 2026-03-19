package models

// CapabilityManifest is the parsed representation of a manifest.yaml file.
type CapabilityManifest struct {
	APIVersion string             `json:"apiVersion" yaml:"apiVersion"`
	Kind       string             `json:"kind" yaml:"kind"`
	Metadata   ManifestMetadata   `json:"metadata" yaml:"metadata"`
	App        ManifestApp        `json:"app" yaml:"app"`
	Capabilities []ManifestCapability `json:"capabilities" yaml:"capabilities"`
}

// ManifestMetadata holds identifying information from the manifest header.
type ManifestMetadata struct {
	AgentID  string `json:"agent_id" yaml:"agent_id"`
	Language string `json:"language" yaml:"language"`
	Version  string `json:"version" yaml:"version"`
}

// ManifestApp describes the application the proxy will forward requests to.
type ManifestApp struct {
	Port        int    `json:"port" yaml:"port"`
	Protocol    string `json:"protocol" yaml:"protocol"`
	HealthPath  string `json:"health_path" yaml:"health_path"`
	BasePath    string `json:"base_path" yaml:"base_path"`
}

// ManifestCapability is a single capability entry inside a manifest.
type ManifestCapability struct {
	CapabilityID     string         `json:"capability_id" yaml:"capability_id"`
	Name             string         `json:"name" yaml:"name"`
	Description      string         `json:"description" yaml:"description"`
	Tags             []string       `json:"tags,omitempty" yaml:"tags,omitempty"`
	Idempotent       bool           `json:"idempotent" yaml:"idempotent"`
	SideEffects      bool           `json:"side_effects" yaml:"side_effects"`
	RequiresApproval bool           `json:"requires_approval,omitempty" yaml:"requires_approval,omitempty"`
	Endpoint         EndpointSpec   `json:"endpoint" yaml:"endpoint"`
	Security         SecurityConfig `json:"security,omitempty" yaml:"security,omitempty"`
	Transport        TransportConfig `json:"transport,omitempty" yaml:"transport,omitempty"`
}
