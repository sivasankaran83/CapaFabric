package config

import "fmt"

// Validate checks the loaded config for required values and invalid combinations.
func Validate(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be 1-65535, got %d", cfg.Server.Port)
	}

	validRegistryTypes := map[string]bool{"inmemory": true, "redis": true, "postgres": true}
	if !validRegistryTypes[cfg.Registry.Type] {
		return fmt.Errorf("registry.type must be one of [inmemory, redis, postgres], got %q", cfg.Registry.Type)
	}
	if (cfg.Registry.Type == "redis" || cfg.Registry.Type == "postgres") && cfg.Registry.URL == "" {
		return fmt.Errorf("registry.url is required for type %q", cfg.Registry.Type)
	}

	validAuthProviders := map[string]bool{"jwt": true, "mtls": true, "apikey": true, "none": true}
	if !validAuthProviders[cfg.Auth.Provider] {
		return fmt.Errorf("auth.provider must be one of [jwt, mtls, apikey, none], got %q", cfg.Auth.Provider)
	}
	if cfg.Auth.Provider == "jwt" && cfg.Auth.JWKSURL == "" {
		return fmt.Errorf("auth.jwks_url is required for jwt provider")
	}

	validEnforcers := map[string]bool{"rbac": true, "opa": true, "none": true}
	if !validEnforcers[cfg.Policy.Enforcer] {
		return fmt.Errorf("policy.enforcer must be one of [rbac, opa, none], got %q", cfg.Policy.Enforcer)
	}

	return nil
}
