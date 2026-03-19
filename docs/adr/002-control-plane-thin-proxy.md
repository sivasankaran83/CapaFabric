# ADR-002: Control Plane + Thin Proxy Architecture

**Status**: Accepted
**Date**: 2026-03-19
**Decision Makers**: CapaFabric Core Team

## Context
Initial design used a single sidecar per pod. This created single-point-of-failure
risks, configuration complexity per pod, and no centralized visibility.

## Decision
Split into Control Plane (centralized, replicated, stateless) and Thin Proxy
(per-pod, lightweight, caches config from CP). Inspired by Istio (istiod + Envoy).

- CP owns: registry, policy engine, guardrail rules, health monitoring, load balancing, admin UI
- Proxy owns: local pipeline invocation (guardrails, tracing, routing), config cache

## Consequences
- Proxy survives CP outage via cached config (30-300s TTL)
- Configuration is centralized, not per-pod YAML
- Admin UI provides single-pane visibility
- Added complexity: two binaries, config distribution protocol
