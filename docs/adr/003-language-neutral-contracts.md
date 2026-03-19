# ADR-003: Language-Neutral Contracts for Component Swappability

**Status**: Accepted
**Date**: 2026-03-19
**Decision Makers**: CapaFabric Core Team

## Context
Components (CP, Proxy, SDKs, Admin UI) must be rewritable in any language without
affecting other components. Initial implementation is Go (CP, Proxy) but future
versions may use Rust (proxy), or other languages.

## Decision
All component boundaries are defined by language-neutral contracts in `contracts/`:
- OpenAPI 3.1 specs for REST APIs (CP, Proxy, Admin UI)
- Protocol Buffer definitions for gRPC services
- JSON Schema for all value objects and manifests

Contract tests (`contract-tests/`) validate any implementation against these specs.
SDK models are auto-generated from JSON Schemas via quicktype.

## Consequences
- Any component can be rewritten without touching others
- Contract tests are CI gates — no merge without passing
- Generated code eliminates model drift across SDKs
- Added overhead: maintaining contracts as source of truth
