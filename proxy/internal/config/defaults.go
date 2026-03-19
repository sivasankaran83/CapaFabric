package config

// Defaults returns a ProxyConfig pre-populated with sensible defaults.
func Defaults() ProxyConfig {
	return ProxyConfig{
		Mode: ModeAgent,
		Port: 3500,
		ControlPlane: ControlPlaneConfig{
			URL:             "http://localhost:8080",
			CacheTTLSeconds: 30,
		},
		App: AppConfig{
			Port:                  8081,
			HealthCheckIntervalMS: 10000,
			HealthPath:            "/health",
		},
		LLM: LLMConfig{
			Endpoint:              "http://localhost:4000/v1",
			FallbackEndpoint:      "http://localhost:11434/v1",
			HealthCheckIntervalMS: 10000,
		},
		Observability: ObservabilityConfig{
			Tracer:      "log",
			ServiceName: "cfproxy",
		},
	}
}
