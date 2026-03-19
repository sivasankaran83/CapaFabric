package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// Load reads, substitutes env vars, and parses a YAML config file.
// Values from the file override defaults.
func Load(path string) (*Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	// Substitute ${ENV_VAR} patterns before YAML parsing.
	expanded := envVarPattern.ReplaceAllFunc(data, func(match []byte) []byte {
		name := envVarPattern.FindSubmatch(match)[1]
		val := os.Getenv(string(name))
		if val == "" {
			return match // leave unexpanded if not set
		}
		return []byte(val)
	})

	if err := yaml.Unmarshal(expanded, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}

	return &cfg, nil
}
