# CapaFabric

The control plane for enterprise agentic AI capability integration.

Any capability. Any language. Any transport. Discovered and governed automatically.

## What is CapaFabric?

CapaFabric connects your existing enterprise applications to agentic AI systems without changing a single line of application code. Write a YAML manifest that describes your API endpoints, add a thin proxy sidecar, and your application becomes AI-discoverable, AI-governable, and AI-composable.

- **Capabilities** are what your systems can do.
- **Goals** are what the AI should achieve.
- The LLM discovers available capabilities and reasons about the invocation path.

No hardcoded workflows. No DAGs. No step definitions.

## Key Features

- **Zero code change** — integrate any HTTP service with a YAML manifest
- **Framework agnostic** — works with Semantic Kernel, LangGraph, CrewAI, or any agent framework
- **Language agnostic** — .NET, Go, Java, Python, TypeScript, or any language with an HTTP client
- **Transport agnostic** — HTTP, gRPC, Pub/Sub, Webhook, MCP
- **Enterprise governance** — authentication, policy enforcement, guardrails, OTEL tracing, audit log
- **Multi-agent orchestration** — agents are just capabilities to other agents (ADR-008)
- **Component swappable** — every component replaceable via language-neutral contracts

## Architecture

```
┌─────────────────────────────────────────────────┐
│                 Control Plane :8080              │
│  Registry · Policy · Health · Config Distribution│
└────────────────────┬────────────────────────────┘
                     │ register / heartbeat / discover
          ┌──────────┴───────────┐
          │                      │
  ┌───────▼────────┐    ┌────────▼───────┐
  │  Agent Proxy   │    │Capability Proxy│
  │   :3500        │    │   :3501        │
  │ (agent mode)   │    │(capability mode│
  └───────┬────────┘    └────────┬───────┘
          │                      │
  ┌───────▼────────┐    ┌────────▼───────┐
  │  AI Agent      │    │  Your App      │
  │ (any language) │    │ (zero changes) │
  └────────────────┘    └────────────────┘
```

### Components

| Component | Language | Description |
|-----------|----------|-------------|
| Control Plane | Go | Centralized registry, policy, health monitoring, config distribution |
| Thin Proxy | Go | Per-pod sidecar; capability or agent mode; local pipeline for guardrails and routing |
| Agent SDK | .NET / Go / TS / Python | Thin HTTP clients over the proxy API |
| Capability Manifest | YAML | Describes API endpoints, authentication, guardrails — the only required change |

## Quickstart

### Prerequisites

- Go 1.23+
- (Optional) .NET 8+ for the .NET SDK and examples
- (Optional) Docker for the full stack

### Build

```bash
git clone https://github.com/sivasankaran83/CapaFabric.git
cd CapaFabric
make build
```

Binaries are written to `bin/`:
- `bin/capafabric` — Control Plane
- `bin/cfproxy` — Thin Proxy

### Run locally (dev mode)

```bash
# 1. Start the Control Plane (in-memory, no external deps)
make run-cp

# 2. Start the proxy in agent mode (AI agent side)
make run-proxy-goal

# 3. Start the proxy in capability mode (application side)
make run-proxy-cap MANIFEST=examples/capability-dotnet-oracle/manifest.yaml
```

### Run with Docker Compose

```bash
make up      # start all services
make down    # stop all services
```

## Capability Manifest

The manifest is the only artifact you add to an existing application:

```yaml
app:
  name: oracle-service
  base_path: /api
  health_path: /health
  port: 8081

capabilities:
  - id: query-inventory
    name: Query Inventory
    description: Returns current stock levels for a product SKU
    tags: [inventory, read]
    endpoint:
      method: GET
      path: /inventory/{sku}
      parameters:
        - name: sku
          in: path
          required: true
    response:
      from: body
```

Run the capability proxy with this manifest and the capability is instantly discoverable by any AI agent connected to the control plane.

## Repository Layout

```
CapaFabric/
├── shared/                  # Shared Go module (models, interfaces, errors, resilience)
│   ├── models/              # Core domain types (CapabilityMetadata, InvocationContext, ...)
│   ├── interfaces/          # Cross-package interfaces (registry, discovery, transport, ...)
│   ├── errors/              # Typed ErrorCode + AppError + MapToHTTPStatus
│   └── resilience/          # CircuitBreaker + Retry (exponential backoff + jitter)
├── control-plane/           # Control Plane Go module
│   ├── cmd/capafabric/      # Entry point (main.go)
│   ├── internal/
│   │   ├── config/          # YAML config loading + validation
│   │   ├── registry/        # In-memory capability registry
│   │   ├── discovery/       # Capability discovery + filtering
│   │   └── api/             # HTTP router, middleware, handlers
│   └── configs/             # Dev and production config files
├── proxy/                   # Thin Proxy Go module
│   ├── cmd/cfproxy/         # Entry point (main.go)
│   ├── internal/
│   │   ├── config/          # YAML config + mode (agent | capability)
│   │   ├── manifest/        # Manifest loader, endpoint mapper, response mapper
│   │   ├── transport/       # HTTP transport adapter + per-capability circuit breakers
│   │   ├── registration/    # CP registration + periodic heartbeat
│   │   ├── cache/           # TTL config cache (serves stale on CP outage)
│   │   ├── fallback/        # Circuit breaker tuned for CP connectivity
│   │   └── api/             # HTTP router, middleware, handlers
│   └── configs/             # Agent and capability mode config files
├── sdk/                     # Language SDKs
│   └── dotnet/              # .NET SDK (CapaFabricAgent base class)
├── examples/                # End-to-end examples
│   ├── agent-dotnet/        # .NET AI agent using Semantic Kernel
│   ├── capability-dotnet-oracle/ # .NET app exposed as a capability
│   ├── capability-go-matching/   # Go app exposed as a capability
│   └── multi-agent-stallion/    # Multi-agent orchestration example
├── contracts/               # OpenAPI 3.1 + JSON Schema (language-neutral contracts)
├── docs/                    # Architecture and design documentation
│   ├── adr/                 # Architecture Decision Records (ADR-001 through ADR-010)
│   └── CAPAFABRIC_CODEBASE.md
└── config/                  # LiteLLM and shared infra configs
```

## API Reference

All routes are prefixed `/api/v1/`.

### Control Plane `:8080`

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/capabilities/register` | Register a capability |
| `DELETE` | `/api/v1/capabilities/{id}` | Unregister a capability |
| `GET` | `/api/v1/capabilities` | List all registered capabilities |
| `POST` | `/api/v1/discover` | Discover capabilities by goal/tags/tenant |
| `POST` | `/api/v1/invoke/{id}` | Invoke a capability |
| `POST` | `/api/v1/heartbeat` | Capability heartbeat |
| `GET` | `/api/v1/health` | Control plane health |
| `GET` | `/api/v1/health/capabilities` | Capability health summary |

### Proxy `:3500` (agent) / `:3501` (capability)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/discover` | Discover capabilities (from cache in agent mode) |
| `POST` | `/api/v1/invoke/{id}` | Invoke a capability |
| `GET` | `/api/v1/health` | Proxy health |
| `GET` | `/api/v1/health/app` | Backing application health (capability mode only) |

## Architecture Decisions

Key decisions are documented in [docs/adr/](docs/adr/):

| ADR | Title |
|-----|-------|
| [001](docs/adr/001-capability-goal-vocabulary.md) | Capability / Goal vocabulary |
| [002](docs/adr/002-control-plane-thin-proxy.md) | Control Plane + Thin Proxy split |
| [003](docs/adr/003-language-neutral-contracts.md) | Language-neutral contracts |
| [004](docs/adr/004-cognitive-provider.md) | Pluggable cognitive provider |
| [005](docs/adr/005-yaml-manifest-primary-contract.md) | YAML manifest as primary contract |
| [006](docs/adr/006-zero-hardcoded-workflow.md) | Zero hardcoded workflow |
| [007](docs/adr/007-governance-opt-in-yaml.md) | Governance opt-in via YAML |
| [008](docs/adr/008-agents-as-capabilities.md) | Agents are capabilities |
| [009](docs/adr/009-agent-base-class.md) | Agent base class pattern |
| [010](docs/adr/010-agent-loop-protection.md) | Agent loop protection |

## Development

```bash
make test    # run all tests (shared + control-plane + proxy)
make lint    # go vet all modules
make clean   # remove build artifacts
```

See [CLAUDE.md](CLAUDE.md) for Go conventions, error handling patterns, and notes for AI-assisted development.

## Documentation

| Document | Description |
|----------|-------------|
| [CAPAFABRIC_ALGORITHM.md](docs/CAPAFABRIC_ALGORITHM.md) | Core algorithm and architecture spec |
| [CAPAFABRIC_CODEBASE.md](docs/CAPAFABRIC_CODEBASE.md) | Go project structure and contracts |
| [CAPAFABRIC_PROJECT_SETUP.md](docs/CAPAFABRIC_PROJECT_SETUP.md) | Project setup and ADRs |
| [CAPAFABRIC_GO_BEST_PRACTICES.md](docs/CAPAFABRIC_GO_BEST_PRACTICES.md) | Go coding standards |
| [CAPAFABRIC_DOTNET_GUIDE.md](docs/CAPAFABRIC_DOTNET_GUIDE.md) | .NET agent implementation guide |
| [CAPAFABRIC_SINGLE_AGENT.md](docs/CAPAFABRIC_SINGLE_AGENT.md) | Single agent example walkthrough |
| [CAPAFABRIC_MULTI_AGENT.md](docs/CAPAFABRIC_MULTI_AGENT.md) | Multi-agent orchestration guide |

## License

MIT
