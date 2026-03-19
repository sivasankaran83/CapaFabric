package config

// ProxyMode controls which pipeline the proxy runs.
type ProxyMode string

const (
	ModeCapability ProxyMode = "capability" // sidecar next to an application
	ModeAgent      ProxyMode = "agent"      // sidecar next to an AI agent
)

// ProxyConfig is the top-level configuration for cfproxy.
type ProxyConfig struct {
	Mode         ProxyMode          `yaml:"mode"`
	Port         int                `yaml:"port"`
	ControlPlane ControlPlaneConfig `yaml:"control_plane"`
	App          AppConfig          `yaml:"app"`           // capability mode only
	LLM          LLMConfig          `yaml:"llm"`           // agent mode only
	Manifest     string             `yaml:"manifest"`      // path to manifest.yaml (capability mode)
	Observability ObservabilityConfig `yaml:"observability"`
}

// ControlPlaneConfig holds connection settings for the Control Plane.
type ControlPlaneConfig struct {
	URL            string `yaml:"url"`
	CacheTTLSeconds int   `yaml:"cache_ttl_seconds"`
}

// AppConfig describes the local application the proxy fronts (capability mode).
type AppConfig struct {
	Port                  int    `yaml:"port"`
	HealthCheckIntervalMS int    `yaml:"health_check_interval_ms"`
	HealthPath            string `yaml:"health_path"`
}

// LLMConfig holds settings for the upstream LLM gateway (agent mode).
type LLMConfig struct {
	Endpoint             string `yaml:"endpoint"`
	FallbackEndpoint     string `yaml:"fallback_endpoint"`
	HealthCheckIntervalMS int   `yaml:"health_check_interval_ms"`
}

// ObservabilityConfig holds tracing and metrics settings.
type ObservabilityConfig struct {
	Tracer       string `yaml:"tracer"`        // otel | log | none
	OTELEndpoint string `yaml:"otel_endpoint"`
	ServiceName  string `yaml:"service_name"`
}
