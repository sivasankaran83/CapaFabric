package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	proxytransport "github.com/sivasankaran83/CapaFabric/proxy/internal/transport"
	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// CapabilityLookup is the minimal interface the invoke handler needs.
type CapabilityLookup interface {
	Capabilities() []models.CapabilityMetadata
}

// InvokeHandler handles POST /api/v1/invoke/{id}.
type InvokeHandler struct {
	source     CapabilityLookup
	transport  *proxytransport.Manager
	appBaseURL string
	logger     *slog.Logger
}

// NewInvokeHandler creates an InvokeHandler.
// appBaseURL is empty in agent mode; the proxy routes to cap.ProxyURL instead.
func NewInvokeHandler(source CapabilityLookup, appBaseURL string, logger *slog.Logger) *InvokeHandler {
	return &InvokeHandler{
		source:     source,
		transport:  proxytransport.NewManager(logger),
		appBaseURL: appBaseURL,
		logger:     logger.With("handler", "invoke"),
	}
}

func (h *InvokeHandler) Invoke(w http.ResponseWriter, r *http.Request) error {
	capabilityID := r.PathValue("id")
	if capabilityID == "" {
		return capaerrors.New(capaerrors.ErrValidation, "capability_id is required")
	}

	// ADR-010: loop-protection header enforcement.
	if err := checkCallChainHeaders(r); err != nil {
		return err
	}

	var req struct {
		Arguments map[string]any `json:"arguments"`
		CallerID  string         `json:"caller_id"`
		GoalID    string         `json:"goal_id"`
		TenantID  string         `json:"tenant_id"`
		CallChain []string       `json:"call_chain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return capaerrors.Wrap(capaerrors.ErrValidation, "invalid request body", err)
	}

	cap, ok := h.findCapability(capabilityID)
	if !ok {
		return capaerrors.NotFound("capability", capabilityID)
	}
	if cap.Status == models.CapabilityStatusUnhealthy {
		return capaerrors.New(capaerrors.ErrCapabilityUnhealthy,
			"capability is currently unhealthy")
	}

	// Agent mode: relay to the remote capability proxy.
	if h.appBaseURL == "" && cap.ProxyURL != "" {
		result, err := forwardToRemoteProxy(r.Context(), cap.ProxyURL, capabilityID,
			req.Arguments, req.CallerID, req.GoalID, req.TenantID, r.Header)
		if err != nil {
			return capaerrors.Wrap(capaerrors.ErrTransportError,
				fmt.Sprintf("remote proxy relay failed for %s", capabilityID), err)
		}
		return writeJSON(w, http.StatusOK, result)
	}

	// Capability mode: invoke via local HTTP transport with retry + circuit breaker.
	ictx := models.InvocationContext{
		RequestID: uuid.NewString(),
		GoalID:    req.GoalID,
		CallerID:  req.CallerID,
		TenantID:  req.TenantID,
		CallChain: req.CallChain,
		Arguments: req.Arguments,
		StartedAt: time.Now().UTC(),
		MaxDepth:  10,
	}

	adapter, err := h.transport.Adapter(cap, h.appBaseURL)
	if err != nil {
		return capaerrors.Wrap(capaerrors.ErrTransportError, "transport init failed", err)
	}

	result, err := adapter.Invoke(r.Context(), cap, ictx)
	if err != nil {
		return capaerrors.Wrap(capaerrors.ErrInvocationFailed,
			fmt.Sprintf("invocation failed for %s", capabilityID), err)
	}

	h.logger.Info("invocation completed",
		"capability_id", capabilityID,
		"success", result.Success,
		"duration_ms", result.DurationMS,
	)
	return writeJSON(w, http.StatusOK, result)
}

func (h *InvokeHandler) findCapability(id string) (models.CapabilityMetadata, bool) {
	for _, cap := range h.source.Capabilities() {
		if cap.CapabilityID == id {
			return cap, true
		}
	}
	return models.CapabilityMetadata{}, false
}

// checkCallChainHeaders enforces loop-protection (ADR-010).
func checkCallChainHeaders(r *http.Request) error {
	chain := r.Header.Get("X-CapaFabric-Call-Chain")
	if chain == "" {
		return nil
	}
	var depth, maxDepth int
	fmt.Sscanf(r.Header.Get("X-CapaFabric-Depth"), "%d", &depth)
	fmt.Sscanf(r.Header.Get("X-CapaFabric-Max-Depth"), "%d", &maxDepth)
	if maxDepth == 0 {
		maxDepth = 10
	}
	if depth >= maxDepth {
		ids := strings.Split(chain, ",")
		return capaerrors.WithDetail(capaerrors.ErrMaxDepthExceeded,
			"max agent depth exceeded",
			fmt.Sprintf("depth=%d max=%d chain=%v", depth, maxDepth, ids),
		)
	}
	return nil
}

// forwardToRemoteProxy relays an invoke call to a remote capability proxy.
func forwardToRemoteProxy(
	ctx context.Context,
	proxyBaseURL, capabilityID string,
	args map[string]any,
	callerID, goalID, tenantID string,
	incomingHeaders http.Header,
) (any, error) {
	url := proxyBaseURL + "/api/v1/invoke/" + capabilityID
	body, _ := json.Marshal(map[string]any{
		"arguments": args,
		"caller_id": callerID,
		"goal_id":   goalID,
		"tenant_id": tenantID,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url,
		strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	for _, h := range []string{
		"X-CapaFabric-Call-Chain",
		"X-CapaFabric-Depth",
		"X-CapaFabric-Max-Depth",
		"X-Trace-ID",
	} {
		if v := incomingHeaders.Get(h); v != "" {
			req.Header.Set(h, v)
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
