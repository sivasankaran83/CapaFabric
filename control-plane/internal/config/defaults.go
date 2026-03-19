package config

// Defaults returns a Config populated with sensible defaults.
func Defaults() Config {
	return Config{
		Server: ServerConfig{
			Port:    8080,
			AdminUI: false,
		},
		Registry: RegistryConfig{
			Type: "inmemory",
		},
		State: StateConfig{
			Type: "inmemory",
		},
		Auth: AuthConfig{
			Provider: "none",
		},
		Policy: PolicyConfig{
			Enforcer: "none",
		},
		Guardrails: GuardrailsConfig{
			Enabled:  false,
			FailMode: "block",
		},
		ContextMgmt: ContextMgmtConfig{
			Enabled:  false,
			Strategy: "passthrough",
		},
		Observability: ObservabilityConfig{
			Tracer:      "log",
			ServiceName: "capafabric-control-plane",
			Audit: AuditConfig{
				Enabled: false,
			},
		},
		LoadBalancing: LoadBalancingConfig{
			Strategy:              "round_robin",
			HealthCheckIntervalMS: 10000,
		},
		Health: HealthConfig{
			HeartbeatTTLSeconds: 60,
			CheckIntervalMS:     30000,
		},
	}
}
