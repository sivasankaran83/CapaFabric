package transport

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
	proxymanifest "github.com/sivasankaran83/CapaFabric/proxy/internal/manifest"
	"github.com/sivasankaran83/CapaFabric/shared/models"
	"github.com/sivasankaran83/CapaFabric/shared/resilience"
)

// HTTPTransportAdapter forwards capability invocations to a backing HTTP application.
// It builds the request via endpoint_mapper and extracts results via response_mapper.
// Each adapter may carry a shared circuit breaker (from Manager) to protect the endpoint.
type HTTPTransportAdapter struct {
	client     *http.Client
	appBaseURL string // e.g. "http://localhost:8081/api"
	breaker    *resilience.CircuitBreaker
	retry      resilience.RetryConfig
	logger     *slog.Logger
}

// newHTTP returns an HTTPTransportAdapter. breaker may be nil (disables circuit breaking).
func newHTTP(appBaseURL string, breaker *resilience.CircuitBreaker, logger *slog.Logger) *HTTPTransportAdapter {
	return &HTTPTransportAdapter{
		appBaseURL: appBaseURL,
		breaker:    breaker,
		retry:      resilience.DefaultRetryConfig(),
		logger:     logger.With("component", "http-transport"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Invoke forwards the invocation to the application and returns the result.
func (t *HTTPTransportAdapter) Invoke(
	ctx context.Context,
	cap models.CapabilityMetadata,
	ictx models.InvocationContext,
) (models.InvocationResult, error) {
	start := time.Now()

	req, err := proxymanifest.BuildRequest(t.appBaseURL, cap, ictx.Arguments)
	if err != nil {
		return models.InvocationResult{}, capaerrors.Wrap(capaerrors.ErrTransportError,
			fmt.Sprintf("building request for %s", cap.CapabilityID), err)
	}

	req = req.WithContext(ctx)

	// Propagate CapaFabric call-chain headers.
	if len(ictx.CallChain) > 0 {
		req.Header.Set("X-CapaFabric-Call-Chain", strings.Join(ictx.CallChain, ","))
		req.Header.Set("X-CapaFabric-Depth", fmt.Sprintf("%d", ictx.Depth))
		req.Header.Set("X-CapaFabric-Max-Depth", fmt.Sprintf("%d", ictx.MaxDepth))
	}
	if ictx.TraceID != "" {
		req.Header.Set("X-Trace-ID", ictx.TraceID)
	}

	var resp *http.Response
	doErr := t.doWithResilience(ctx, func(ctx context.Context) error {
		r := req.Clone(ctx)
		var e error
		resp, e = t.client.Do(r)
		return e
	})

	durationMS := time.Since(start).Milliseconds()

	if doErr != nil {
		return models.InvocationResult{
			RequestID:    ictx.RequestID,
			CapabilityID: cap.CapabilityID,
			Success:      false,
			DurationMS:   durationMS,
			Error: &models.InvokeError{
				Code:    capaerrors.ErrCodeTransportError,
				Message: doErr.Error(),
			},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return models.InvocationResult{
			RequestID:    ictx.RequestID,
			CapabilityID: cap.CapabilityID,
			Success:      false,
			DurationMS:   durationMS,
			Error: &models.InvokeError{
				Code:    capaerrors.ErrCodeInvocationFailed,
				Message: fmt.Sprintf("application returned HTTP %d", resp.StatusCode),
			},
		}, nil
	}

	result, err := proxymanifest.ExtractResult(resp, cap.Endpoint.Response.From)
	if err != nil {
		return models.InvocationResult{
			RequestID:    ictx.RequestID,
			CapabilityID: cap.CapabilityID,
			Success:      false,
			DurationMS:   durationMS,
			Error: &models.InvokeError{
				Code:    capaerrors.ErrCodeInvocationFailed,
				Message: "extracting result: " + err.Error(),
			},
		}, nil
	}

	return models.InvocationResult{
		RequestID:    ictx.RequestID,
		CapabilityID: cap.CapabilityID,
		Success:      true,
		Result:       result,
		DurationMS:   durationMS,
	}, nil
}

// doWithResilience runs fn wrapped in retry and (optionally) circuit breaker.
func (t *HTTPTransportAdapter) doWithResilience(ctx context.Context, fn func(context.Context) error) error {
	retryFn := func(ctx context.Context) error {
		return fn(ctx)
	}

	if t.breaker != nil {
		return resilience.Retry(ctx, t.retry, func(ctx context.Context) error {
			return t.breaker.Do(ctx, retryFn)
		})
	}

	return resilience.Retry(ctx, t.retry, retryFn)
}
