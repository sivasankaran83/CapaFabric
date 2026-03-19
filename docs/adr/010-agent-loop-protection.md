# ADR-010: Agent Loop Protection (Circular Chain, Max Depth, Retry Detection)

**Status**: Accepted
**Date**: 2026-03-19
**Decision Makers**: CapaFabric Core Team

## Context
In a multi-agent system where Agents discover and invoke Capabilities (which
may themselves be Agents), three types of infinite loops can occur:
1. Agent calls itself (circular self-reference)
2. Agent A → Agent B → Agent C → Agent A (circular chain)
3. Agent retries the same failing capability with identical arguments

## Decision
Three guards implemented at TWO levels (belt and suspenders):

**SDK Level (Agent base class):**

Guard 1 — Circular chain: `AgentContext` carries a `CallChain` (list of `agent_id`s).
  Before executing, Agent checks if its own ID is already in the chain.
  If yes → return `AgentResult.Failed` with chain trace.

Guard 2 — Max depth: `CallChain.Count` checked against configurable `MaxDepth`
  (default: 10). If exceeded → escalate to HITL.

Guard 3 — Retry loop: `InvocationLog` tracks `(capability_id, args_hash)` for
  last N calls. If same capability + same args repeated N times consecutively
  (default: 3) → inject structured error into tool result so LLM can
  self-correct. LLM receives: "You've tried this 3 times with identical
  input. Try a different approach or escalate."

**Infrastructure Level (Proxy):**

Call chain propagated via HTTP headers:
```
X-CapaFabric-Call-Chain: supervisor,ingestion,extraction
X-CapaFabric-Depth: 3
X-CapaFabric-Max-Depth: 10
```

Proxy validates before routing:
- Target agent already in chain → `409 Conflict`
- Depth >= `max_depth` → `429 Too Deep`

This catches cross-service loops that the SDK can't see (Agent A in .NET
calls Agent B in Go which calls Agent A — the .NET process doesn't know
about the Go call, but the proxy headers carry the full chain).

## Consequences
- Three loop types covered at two enforcement levels
- `CallChain` propagates across language boundaries via HTTP headers
- All guards produce structured errors the LLM can reason about
- All guards escalate to HITL by default (configurable per Agent)
- Guards are defined in `contracts/openapi/proxy-api.openapi.yaml`
- `MaxDepth` and `MaxConsecutiveRetries` configurable per Agent and per tenant
- ADR-009 Agent base class embeds all three guards in `PursueGoalAsync`
