# ADR-004: CognitiveProvider as Polymorphic LLM Abstraction

**Status**: Accepted
**Date**: 2026-03-19
**Decision Makers**: CapaFabric Core Team

## Context
The Agent needs to call LLMs, but the LLM provider varies per agent (Supervisor
uses Claude, Ingestion uses Haiku, Matching uses local Ollama) and per invocation
within the same agent. Coupling to LiteLLM or any specific provider is unacceptable.

## Decision
Introduce `ICognitiveProvider` with polymorphic `Deliberate()` method:
- `Deliberate(request)` → simple request/response
- `Deliberate(request, toolInvoker)` → with tool calling (delegate callback pattern)
- `DeliberateStream(request)` → streaming tokens
- `DeliberateStream(request, toolInvoker)` → streaming + tool calling

`CognitiveConfig` is provider-specific (`LiteLLMConfig`, `OllamaConfig`, `SemanticKernelConfig`).
Injected per-agent via DI. Can vary per invocation.

## Consequences
- Agent is completely decoupled from any LLM provider
- Per-agent model selection enables cost optimization (60-70% cost reduction)
- `toolInvoker` delegate keeps `CognitiveProvider` decoupled from CapaFabric invocation
- 4 overloads cover every agentic interaction pattern
