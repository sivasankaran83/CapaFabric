# CapaFabric

The control plane for enterprise agentic AI capability integration.

Any capability. Any language. Any transport. Discovered and governed automatically.

## What is CapaFabric?

CapaFabric connects your existing enterprise applications to agentic AI systems
without changing a single line of application code. Write a YAML manifest that
describes your API endpoints, add a thin proxy, and your application becomes
AI-discoverable, AI-governable, and AI-composable.

Capabilities are what your systems can do.
Goals are what the AI should achieve.
The LLM discovers available capabilities and reasons about the invocation path.
No hardcoded workflows. No DAGs. No step definitions.

## Key Features

- Zero code change integration via YAML manifest
- Framework agnostic: Semantic Kernel, LangGraph, CrewAI, or any Agent
- Language agnostic: .NET, Go, Java, Python, TypeScript, or any language
- Transport agnostic: HTTP, gRPC, Pub/Sub, Webhook, MCP
- Enterprise governance: auth, policy, guardrails, OTEL tracing, audit
- Multi-agent orchestration: agents are just capabilities to other agents
- Component swappable: every component replaceable via language-neutral contracts

## Architecture

- Control Plane (Go): centralized registry, policy, health, config distribution
- Thin Proxy (Go): per-pod pipeline for guardrails, tracing, routing
- Agent (any): LLM-powered reasoning, calls localhost:3500
- Capability (any): existing app + YAML manifest, zero code changes

## Documentation

- CAPAFABRIC_ALGORITHM.md - Architecture spec
- CAPAFABRIC_CODEBASE.md - Go project structure and contracts
- CAPAFABRIC_DOTNET_GUIDE.md - .NET Agent implementation
- CAPAFABRIC_SINGLE_AGENT.md - Single agent example
- CAPAFABRIC_MULTI_AGENT.md - Multi-agent orchestration
- CAPAFABRIC_PROJECT_SETUP.md - Project setup and ADRs

## License

MIT
