# ADR-008: Agents as Capabilities (Multi-Agent Pattern)

**Status**: Accepted
**Date**: 2026-03-19
**Decision Makers**: CapaFabric Core Team

## Context
Multi-agent orchestration traditionally requires a special agent-to-agent protocol
(A2A, AMCP). This adds complexity and a new protocol to maintain.

## Decision
An agent's goal endpoint is registered in the manifest as a capability — identical
to any deterministic capability. The Supervisor Agent discovers sub-agents via
the same `/discover` call used for deterministic capabilities. No special multi-agent
protocol, no agent-to-agent messaging framework.

Each sub-agent encapsulates its own Agent with its own `CognitiveProvider` and
internal capabilities. To the Supervisor, calling a sub-agent is indistinguishable
from calling a deterministic capability.

## Consequences
- Multi-agent orchestration = capability discovery + invocation (no new concepts)
- Sub-agents are independently deployable, scalable, and language-independent
- Supervisor prompt is goal-only — doesn't name sub-agents or prescribe order
- Adding a new agent = deploy + manifest. Supervisor discovers it automatically.
