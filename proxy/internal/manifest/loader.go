package manifest

import (
	"fmt"
	"time"

	sharedmanifest "github.com/sivasankaran83/CapaFabric/shared/manifest"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// LoadCapabilities parses manifest.yaml and converts each capability entry into
// a CapabilityMetadata record ready for registration with the Control Plane.
// proxyURL is the base URL of this proxy instance (e.g. "http://localhost:3501").
func LoadCapabilities(manifestPath, proxyURL string) ([]models.CapabilityMetadata, error) {
	m, err := sharedmanifest.Parse(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}

	now := time.Now().UTC()
	caps := make([]models.CapabilityMetadata, 0, len(m.Capabilities))

	for _, mc := range m.Capabilities {
		caps = append(caps, models.CapabilityMetadata{
			CapabilityID:     mc.CapabilityID,
			Name:             mc.Name,
			Description:      mc.Description,
			Tags:             mc.Tags,
			Language:         m.Metadata.Language,
			AgentID:          m.Metadata.AgentID,
			Source:           models.CapabilitySourceManifest,
			Idempotent:       mc.Idempotent,
			SideEffects:      mc.SideEffects,
			RequiresApproval: mc.RequiresApproval,
			Endpoint:         mc.Endpoint,
			Security:         mc.Security,
			Transport:        mc.Transport,
			Status:           models.CapabilityStatusUnknown,
			RegisteredAt:     now,
			ProxyURL:         proxyURL,
			AppPort:          m.App.Port,
		})
	}

	return caps, nil
}

// ManifestApp returns app-level settings from the manifest (port, health path, base path).
func ManifestApp(manifestPath string) (*models.ManifestApp, error) {
	m, err := sharedmanifest.Parse(manifestPath)
	if err != nil {
		return nil, err
	}
	return &m.App, nil
}
