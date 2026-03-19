# ADR-009: Agent Base Class with Polymorphic Deliberate

**Status**: Accepted
**Date**: 2026-03-19
**Decision Makers**: CapaFabric Core Team

## Context
Every Agent in the system follows the same pattern: discover capabilities,
deliberate via CognitiveProvider, invoke capabilities, evaluate confidence,
escalate to HITL if needed. Without a shared base class, each Agent reimplements
this logic, leading to inconsistency and bugs.

## Decision
Introduce an abstract Agent base class (in every SDK language) that implements
the universal pattern:
1. Discover capabilities relevant to the goal
2. If none found → `OnNoCapabilitiesFound` (default: HITL escalation)
3. Deliberate via `CognitiveProvider` with tool invoker callback
4. Evaluate result: complete, continue, low confidence, or unparsable
5. Low confidence → `OnLowConfidence` (default: HITL escalation)

Concrete Agents extend the base and provide:
- `AgentId` (identity)
- `Persona` (system prompt — goal intent only)
- Override points: `OnNoCapabilitiesFound`, `OnLowConfidence`, `GetCognitiveConfig`,
  `GetConfidenceThreshold`

An `AgentRunner` wraps the iteration loop (max iterations, token budget,
checkpointing) and calls `Agent.PursueGoalAsync` repeatedly until completion.

The base class is implemented in all SDK languages: C#, TypeScript, Python,
Go, Java, Rust — each using the language's native abstraction mechanism
(abstract class, ABC, trait, interface + struct embedding).

## Consequences
- Creating a new Agent = extend base, write Persona, optionally override thresholds
- Universal pattern enforced: all Agents discover → deliberate → invoke → evaluate
- Override points allow per-Agent customization without changing the base
- `CognitiveProvider` is per-Agent (Supervisor uses Sonnet, Ingestion uses Haiku)
- Polyglot: same pattern works identically in 6 languages
