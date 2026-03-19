package interfaces

import (
	"context"
	"time"
)

// StateStore is a key-value store with optional TTL, scoped by (scope, key).
type StateStore interface {
	// Get retrieves a value. Returns (nil, nil) when not found.
	Get(ctx context.Context, scope, key string) ([]byte, error)

	// Set stores a value with optional TTL (zero = no expiry).
	Set(ctx context.Context, scope, key string, value []byte, ttl time.Duration) error

	// Delete removes a key.
	Delete(ctx context.Context, scope, key string) error

	// List returns all keys within a scope.
	List(ctx context.Context, scope string) ([]string, error)
}
