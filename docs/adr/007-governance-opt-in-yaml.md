# ADR-007: Governance Layers as Opt-In YAML Configuration

**Status**: Accepted
**Date**: 2026-03-19
**Decision Makers**: CapaFabric Core Team

## Context
Enterprise deployments need auth, guardrails, tracing, audit, context management,
and load balancing. Development environments need none of these. Embedding governance
in code creates overhead everywhere.

## Decision
Every cross-cutting concern follows the same pattern:
- Off by default
- YAML-activated in CP config
- Abstract interface with injectable implementation
- Zero code change in Agent or capability when enabled/changed

Auth, policy, guardrails, tracing, audit, context management, load balancing —
all configured per-environment, per-tenant, per-capability via YAML.

## Consequences
- Dev environment: zero governance overhead
- Staging: guardrails in `log_only` mode for confidence building
- Production: full governance stack
- PHI deployment: add Presidio redaction + forensic audit
- Same binary, same code — different YAML
