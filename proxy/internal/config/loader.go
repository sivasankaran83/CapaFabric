package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// Load reads, env-substitutes, and parses a proxy YAML config file.
// Returned config has all defaults applied before the file overrides.
func Load(path string) (*ProxyConfig, error) {
	cfg := Defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading proxy config %q: %w", path, err)
	}

	expanded := envVarPattern.ReplaceAllFunc(data, func(match []byte) []byte {
		name := envVarPattern.FindSubmatch(match)[1]
		if val := os.Getenv(string(name)); val != "" {
			return []byte(val)
		}
		return match
	})

	if err := yaml.Unmarshal(expanded, &cfg); err != nil {
		return nil, fmt.Errorf("parsing proxy config YAML: %w", err)
	}

	return &cfg, nil
}

// Validate returns an error if cfg contains invalid or missing values.
func Validate(cfg *ProxyConfig) error {
	if cfg.Mode != ModeCapability && cfg.Mode != ModeAgent {
		return fmt.Errorf("mode must be %q or %q, got %q", ModeCapability, ModeAgent, cfg.Mode)
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("port must be 1-65535, got %d", cfg.Port)
	}
	if cfg.ControlPlane.URL == "" {
		return fmt.Errorf("control_plane.url is required")
	}
	if cfg.Mode == ModeCapability && cfg.Manifest == "" {
		return fmt.Errorf("manifest path is required in capability mode (use --manifest flag)")
	}
	return nil
}
