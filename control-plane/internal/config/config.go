package config

// Config is the top-level control plane configuration structure.
type Config struct {
	Server          ServerConfig          `yaml:"server"`
	Registry        RegistryConfig        `yaml:"registry"`
	State           StateConfig           `yaml:"state"`
	Auth            AuthConfig            `yaml:"auth"`
	Policy          PolicyConfig          `yaml:"policy"`
	Guardrails      GuardrailsConfig      `yaml:"guardrails"`
	ContextMgmt     ContextMgmtConfig     `yaml:"context_management"`
	Observability   ObservabilityConfig   `yaml:"observability"`
	LoadBalancing   LoadBalancingConfig   `yaml:"load_balancing"`
	Health          HealthConfig          `yaml:"health"`
}

type ServerConfig struct {
	Port    int  `yaml:"port"`
	AdminUI bool `yaml:"admin_ui"`
}

type RegistryConfig struct {
	Type string `yaml:"type"` // inmemory | redis | postgres
	URL  string `yaml:"url,omitempty"`
}

type StateConfig struct {
	Type string `yaml:"type"` // inmemory | sqlite | redis | postgres | cosmos
	URL  string `yaml:"url,omitempty"`
	Path string `yaml:"path,omitempty"` // for sqlite
}

type AuthConfig struct {
	Provider string `yaml:"provider"` // jwt | mtls | apikey | none
	JWKSURL  string `yaml:"jwks_url,omitempty"`
	Audience string `yaml:"audience,omitempty"`
}

type PolicyConfig struct {
	Enforcer string `yaml:"enforcer"` // rbac | opa | none
	OPAURL   string `yaml:"opa_url,omitempty"`
}

type GuardrailsConfig struct {
	Enabled  bool                 `yaml:"enabled"`
	FailMode string               `yaml:"fail_mode"` // block | log
	Inbound  []GuardrailRuleConfig `yaml:"inbound,omitempty"`
	Outbound []GuardrailRuleConfig `yaml:"outbound,omitempty"`
}

type GuardrailRuleConfig struct {
	Type     string `yaml:"type"`
	Provider string `yaml:"provider,omitempty"`
	Action   string `yaml:"action"` // block | redact | log
	PHIMode  bool   `yaml:"phi_mode,omitempty"`
}

type ContextMgmtConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Strategy string `yaml:"strategy"` // passthrough | sliding_window | summarizing | rag | adaptive
}

type ObservabilityConfig struct {
	Tracer      string      `yaml:"tracer"` // otel | log | none
	OTELEndpoint string     `yaml:"otel_endpoint,omitempty"`
	ServiceName string      `yaml:"service_name,omitempty"`
	Audit       AuditConfig `yaml:"audit"`
}

type AuditConfig struct {
	Enabled bool   `yaml:"enabled"`
	Writer  string `yaml:"writer,omitempty"` // file | cosmos | postgres
	DSN     string `yaml:"dsn,omitempty"`
	Path    string `yaml:"path,omitempty"`
}

type LoadBalancingConfig struct {
	Strategy              string `yaml:"strategy"` // round_robin | least_connections | affinity | weighted | priority
	HealthCheckIntervalMS int    `yaml:"health_check_interval_ms"`
}

type HealthConfig struct {
	HeartbeatTTLSeconds int `yaml:"heartbeat_ttl_seconds"`
	CheckIntervalMS     int `yaml:"check_interval_ms"`
}
