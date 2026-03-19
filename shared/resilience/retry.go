package resilience

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
)

// RetryConfig controls retry behaviour.
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	// Retryable returns true if the error is transient and should be retried.
	// Defaults to DefaultRetryable when nil.
	Retryable func(error) bool
}

// DefaultRetryConfig returns a sensible retry config.
// Business errors (not-found, auth, validation) are not retried.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Retryable:    DefaultRetryable,
	}
}

// DefaultRetryable returns false for non-transient domain errors.
func DefaultRetryable(err error) bool {
	var ae *capaerrors.AppError
	if errors.As(err, &ae) {
		switch ae.Code {
		case capaerrors.ErrNotFound,
			capaerrors.ErrCapabilityNotFound,
			capaerrors.ErrValidation,
			capaerrors.ErrManifestInvalid,
			capaerrors.ErrAuthentication,
			capaerrors.ErrAuthorization,
			capaerrors.ErrCircularChain,
			capaerrors.ErrMaxDepthExceeded,
			capaerrors.ErrGuardrailBlocked:
			return false // non-transient — do not retry
		}
	}
	return true // transient — retry
}

// Retry executes fn up to cfg.MaxAttempts times with exponential backoff + jitter.
// Stops early if ctx is cancelled or fn returns a non-retryable error.
func Retry(ctx context.Context, cfg RetryConfig, fn func(context.Context) error) error {
	retryable := cfg.Retryable
	if retryable == nil {
		retryable = DefaultRetryable
	}

	delay := cfg.InitialDelay
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}
		if !retryable(lastErr) {
			return lastErr
		}
		if attempt == cfg.MaxAttempts {
			break
		}

		// Jitter: ±25% of delay.
		jitter := time.Duration(rand.Int63n(int64(delay / 4)))
		wait := delay + jitter

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}

		delay = time.Duration(float64(delay) * cfg.Multiplier)
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}

	return fmt.Errorf("max retries (%d) exhausted: %w", cfg.MaxAttempts, lastErr)
}
