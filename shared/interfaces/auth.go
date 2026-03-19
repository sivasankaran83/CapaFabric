package interfaces

import (
	"context"
	"net/http"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// AuthProvider authenticates an inbound HTTP request and returns an identity.
type AuthProvider interface {
	// Authenticate extracts and validates credentials from r.
	// Returns an anonymous identity (not an error) when auth is disabled.
	Authenticate(ctx context.Context, r *http.Request) (models.AuthIdentity, error)
}
