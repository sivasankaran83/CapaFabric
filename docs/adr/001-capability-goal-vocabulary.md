# ADR-001: Capability + Goal Vocabulary

**Status**: Accepted
**Date**: 2026-03-19
**Decision Makers**: CapaFabric Core Team

## Context
The system needs consistent vocabulary that distinguishes between deterministic
units of work and LLM-driven goal pursuit. Terms like "skill," "tool," "function,"
and "agent" are overloaded across the industry (MCP skills, SK skills, LangChain tools).

## Decision
- **Capability**: A discrete, typed unit of work. Deterministic. Any language. Registered via manifest.
- **Goal**: An outcome the system should achieve. Non-deterministic. LLM-driven.
- **Agent**: The cognitive core that pursues a goal by discovering and invoking capabilities.
- **AgentDecision**: Structured output per reasoning iteration.
- **CognitiveProvider**: The polymorphic LLM abstraction (Deliberate method).

## Consequences
- Clear separation between "what the system can do" and "what it should achieve"
- No naming collisions with MCP, Semantic Kernel, or other frameworks
- Manifest vocabulary ("capabilities:") reads naturally
