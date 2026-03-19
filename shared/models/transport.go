package models

// TransportKind identifies the transport protocol used to invoke a capability.
type TransportKind string

const (
	TransportHTTP      TransportKind = "http"
	TransportGRPC      TransportKind = "grpc"
	TransportPubSub    TransportKind = "pubsub"
	TransportWebhook   TransportKind = "webhook"
	TransportMCP       TransportKind = "mcp"
	TransportInProcess TransportKind = "inprocess"
)

// TransportConfig holds transport-level settings for a capability.
type TransportConfig struct {
	Kind           TransportKind `json:"kind,omitempty" yaml:"kind,omitempty"`
	TimeoutSeconds int           `json:"timeout_seconds,omitempty" yaml:"timeout_seconds,omitempty"`
	RetryMax       int           `json:"retry_max,omitempty" yaml:"retry_max,omitempty"`
	TLSEnabled     bool          `json:"tls_enabled,omitempty" yaml:"tls_enabled,omitempty"`
}
