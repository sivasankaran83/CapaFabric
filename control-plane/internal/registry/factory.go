package registry

import (
	"fmt"

	"github.com/sivasankaran83/CapaFabric/control-plane/internal/config"
	"github.com/sivasankaran83/CapaFabric/shared/interfaces"
)

// NewRegistry creates a CapabilityRegistry based on the configuration.
func NewRegistry(cfg config.RegistryConfig) (interfaces.CapabilityRegistry, error) {
	switch cfg.Type {
	case "inmemory", "":
		return NewInMemory(), nil
	default:
		return nil, fmt.Errorf("unsupported registry type %q (supported: inmemory)", cfg.Type)
	}
}
