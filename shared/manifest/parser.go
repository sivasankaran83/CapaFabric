package manifest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// Parse reads and parses a manifest YAML file from the given path.
func Parse(path string) (*models.CapabilityManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, capaerrors.Wrap(capaerrors.ErrManifestParseFailed, "reading manifest file", err)
	}
	return ParseBytes(data)
}

// ParseBytes parses manifest YAML from a byte slice.
func ParseBytes(data []byte) (*models.CapabilityManifest, error) {
	var m models.CapabilityManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, capaerrors.Wrap(capaerrors.ErrManifestParseFailed, "parsing manifest YAML", err)
	}
	if err := Validate(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate checks a parsed manifest for required fields and consistency.
func Validate(m *models.CapabilityManifest) error {
	if m.APIVersion == "" {
		return capaerrors.New(capaerrors.ErrManifestInvalid, "apiVersion is required")
	}
	if m.Kind != "CapabilityManifest" {
		return capaerrors.New(capaerrors.ErrManifestInvalid,
			fmt.Sprintf("kind must be CapabilityManifest, got: %s", m.Kind))
	}
	if m.Metadata.AgentID == "" {
		return capaerrors.New(capaerrors.ErrManifestInvalid, "metadata.agent_id is required")
	}
	if m.App.Port == 0 {
		return capaerrors.New(capaerrors.ErrManifestInvalid, "app.port is required")
	}
	if len(m.Capabilities) == 0 {
		return capaerrors.New(capaerrors.ErrManifestInvalid, "at least one capability is required")
	}
	for i, cap := range m.Capabilities {
		if cap.CapabilityID == "" {
			return capaerrors.New(capaerrors.ErrManifestInvalid,
				fmt.Sprintf("capabilities[%d].capability_id is required", i))
		}
		if cap.Description == "" {
			return capaerrors.New(capaerrors.ErrManifestInvalid,
				fmt.Sprintf("capabilities[%d].description is required (LLM depends on it)", i))
		}
		if cap.Endpoint.Method == "" || cap.Endpoint.Path == "" {
			return capaerrors.New(capaerrors.ErrManifestInvalid,
				fmt.Sprintf("capabilities[%d].endpoint.method and path are required", i))
		}
	}
	return nil
}
