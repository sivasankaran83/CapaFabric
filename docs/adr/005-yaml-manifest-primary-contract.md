# ADR-005: YAML Manifest as Primary Integration Contract

**Status**: Accepted
**Date**: 2026-03-19
**Decision Makers**: CapaFabric Core Team

## Context
Enterprise applications have existing APIs. Requiring them to adopt an SDK, install
a framework, or modify their code creates adoption friction. Auto-discovery from
OpenAPI specs is unreliable in production (specs often missing or outdated).

## Decision
The YAML `CapabilityManifest` is the primary and only required integration artifact.
It maps existing API endpoints to discoverable capabilities without any application
code changes. The manifest includes: endpoint mapping (path/query/body/header),
response mapping (JSON path), security config, and LLM-optimized descriptions.

Auto-discovery from OpenAPI is a future convenience layer on top, not a replacement.

## Consequences
- L0 adoption: write YAML + add proxy = AI-discoverable (zero code changes)
- Explicit over implicit: manifests are version-controlled, PR-reviewable
- Manifest drift risk: must validate manifests against live apps in CI
- Description quality matters: descriptions are the LLM's discovery interface
