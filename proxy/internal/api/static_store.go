package api

import "github.com/sivasankaran83/CapaFabric/shared/models"

// StaticCapabilityStore implements CapabilityStore over a fixed slice.
// Used in capability mode where the capabilities come from the manifest.
type StaticCapabilityStore struct {
	caps []models.CapabilityMetadata
}

// NewStaticStore wraps a fixed capability list as a CapabilityStore.
func NewStaticStore(caps []models.CapabilityMetadata) *StaticCapabilityStore {
	return &StaticCapabilityStore{caps: caps}
}

// Capabilities returns the static capability list.
func (s *StaticCapabilityStore) Capabilities() []models.CapabilityMetadata {
	return s.caps
}
