# CapaFabric — Architecture & Implementation Algorithm

> **CapaFabric** is the control plane for enterprise agentic AI capability integration.
> Framework-agnostic, language-agnostic, transport-agnostic. Any capability in any language
> over any transport, discovered and governed through a centralized control plane with
> thin data plane proxies. No dependency on any specific agent framework.
>
> *"Istio is the control plane for microservices. CapaFabric is the control plane for agentic AI."*

---

## Vocabulary

| Term | Definition |
|---|---|
| **Capability** | A discrete, typed unit of work the system can perform. Deterministic. Has input/output schema and a transport endpoint. Any language. Registered via YAML manifest. |
| **Goal** | An outcome the system should achieve. Non-deterministic. A Agent reasons about which capabilities to invoke. No hardcoded workflow. |
| **Agent** | The cognitive core that pursues a goal by discovering and invoking capabilities. Uses an LLM for reasoning. Any framework. |
| **AgentDecision** | Structured output a Agent produces each iteration: thought process, selected capability, confidence, completion status. |
| **CapabilityManifest** | YAML file that maps existing application endpoints to discoverable capabilities. The ONLY integration artifact. |
| **Control Plane** | Centralized Go binary: registry, policy, guardrails, health, config distribution, admin UI. |
| **Thin Proxy** | Per-pod Go binary: local pipeline enforcement (guardrails, tracing, routing). Caches config from CP. |

---

## Architectural Principle

```
CapaFabric Control Plane:   Centralized intelligence — registry, policy, health, config
Thin Proxy (per pod/node):  Local enforcement — guardrails, tracing, request routing
Agent (any framework):  Calls localhost proxy — doesn't know CapaFabric exists
Capability (any language):  Registers via YAML manifest — doesn't change application code
```

**Three-Axis Independence:**
```
Axis 1 — LANGUAGE:    Python, Go, .NET, Java, Rust, Node.js, COBOL-behind-REST
Axis 2 — TRANSPORT:   HTTP/REST, gRPC, Pub/Sub, Webhook, In-Process, MCP
Axis 3 — FRAMEWORK:   Semantic Kernel, LangGraph, CrewAI, AutoGen, custom
```

**Core rule**: Policy is configuration, not code. Every cross-cutting concern is off by
default, YAML-activated, abstract interface, injectable implementation. Zero code change
in the Agent or capability when governance layers change.

---

## Deployment Topology

```
┌──────────────────────────────────────────────────────────────────────┐
│                    CAPAFABRIC CONTROL PLANE                          │
│                    (Go binary, replicated, stateless)                │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────────────┐ │
│  │ Capability   │ │ Policy &     │ │ Config Distribution          │ │
│  │ Registry     │ │ Guardrail    │ │ (push to proxies)            │ │
│  │              │ │ Engine       │ │                              │ │
│  └──────────────┘ └──────────────┘ └──────────────────────────────┘ │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────────────┐ │
│  │ Health       │ │ Load         │ │ Admin UI (:8080)             │ │
│  │ Monitor      │ │ Balancer     │ │                              │ │
│  └──────────────┘ └──────────────┘ └──────────────────────────────┘ │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────────────┐ │
│  │ State Store  │ │ Audit        │ │ Context Management           │ │
│  │ Manager      │ │ Aggregator   │ │ Strategies                   │ │
│  └──────────────┘ └──────────────┘ └──────────────────────────────┘ │
└──────────┬──────────────────────────────────────┬────────────────────┘
           │                                       │
    ┌──────▼─────────────┐          ┌──────────────▼─────────────┐
    │  GOALAGENT POD      │          │  CAPABILITY POD             │
    │  Any Agent      │          │  .NET / Go / Java / etc.   │
    │  (.NET / Python)    │          │  (UNCHANGED)               │
    │    │                │          │    ▲                        │
    │    ▼                │          │    │                        │
    │  Thin Proxy         │◄────────►│  Thin Proxy                │
    │  localhost:3500     │          │  localhost:3500             │
    └─────────────────────┘          └────────────────────────────┘
```

---

## Project Structure

```
capafabric/
├── CLAUDE.md
├── control-plane/                     # Go — centralized control plane
│   ├── cmd/capafabric/main.go
│   ├── internal/
│   │   ├── registry/                 # Capability registry
│   │   ├── discovery/                # Discovery providers
│   │   ├── policy/                   # Auth + authorization
│   │   ├── guardrails/               # Guardrail engine
│   │   ├── context/                  # Context management
│   │   ├── health/                   # Health monitor
│   │   ├── loadbalancer/             # LB strategies
│   │   ├── state/                    # State store
│   │   ├── audit/                    # Audit aggregator
│   │   ├── config/                   # Config distributor
│   │   ├── manifest/                 # YAML manifest parser
│   │   ├── api/                      # REST API
│   │   └── ui/                       # Admin UI
│   └── configs/
│       ├── capafabric.yaml
│       └── capafabric.dev.yaml
├── proxy/                             # Go — thin data plane proxy
│   ├── cmd/cfproxy/main.go           # --mode=capability|agent
│   ├── internal/
│   │   ├── pipeline/                 # Request pipeline
│   │   ├── transport/                # HTTP, gRPC, pub/sub, webhook
│   │   ├── cache/                    # Config cache from CP
│   │   ├── otel/                     # OTEL emission
│   │   ├── manifest/                 # YAML manifest loader
│   │   └── api/                      # Localhost API
│   └── configs/
├── sdk/                               # Optional language SDKs
│   ├── dotnet/                        # NuGet: CapaFabric.Client
│   ├── python/                        # PyPI: capafabric
│   ├── go/
│   └── java/
├── proto/
│   ├── capability_manifest.schema.json
│   ├── capability_invoke.proto
│   └── capability_result.schema.json
├── config/
│   ├── litellm_config.yaml
│   └── litellm_config.dev.yaml
├── examples/
│   ├── agent-dotnet/
│   ├── capability-dotnet-oracle/
│   ├── capability-go-matching/
│   └── capability-java-email/
└── docker-compose.yaml
```

---

## Abstract Interfaces (9 ABCs)

### Value Objects

```
CapabilityMetadata      — capability_id, name, description, parameters_schema,
                          return_schema, source, transport, security, tags,
                          language, version, health_endpoint

TransportConfig         — kind, endpoint, http_*, grpc_*, pubsub_*, webhook_*, auth_*

SecurityConfig          — required_roles, required_scopes, allowed_callers,
                          denied_callers, classification, audit_level, max_calls_per_minute

InvocationContext       — request_id, capability_id, caller_id, goal_id,
                          tenant_id, iteration, parent_span_id, raw_token

AuthIdentity            — caller_id, authenticated, auth_method, roles, scopes, claims

PolicyDecision          — allowed, reason, requires_approval, rate_limited

AgentDecision            — thought_process, selected_capability, capability_input,
                          confidence_score, is_goal_complete, alternative_capabilities

AgentState               — goal_id, goal_description, tenant_id, iteration,
                          max_iterations, token_budget, tokens_used, status, history
```

### Interfaces

```
1. CapabilityRegistry    — stores/retrieves capability metadata + handlers
     namespace (property), register, unregister, get_all, get_by_id,
     search, get_handler, filter_by_source

2. DiscoveryProvider     — resolves which capabilities to expose for a goal
     discover(context, max_tools, sources?, tags?) → list[CapabilityMetadata]

3. ToolFormatAdapter     — converts metadata to LLM provider tool format
     to_tools(capabilities) → list[dict]
     parse_tool_call(raw_call) → (capability_id, arguments)

4. TransportAdapter      — invokes capability over wire protocol
     invoke(metadata, arguments) → Any
     health_check(metadata) → bool
     supports(transport_kind) → bool

5. AuthProvider          — authenticates caller identity
     authenticate(context) → AuthIdentity

6. PolicyEnforcer        — authorizes caller against SecurityConfig
     authorize(identity, metadata, context) → PolicyDecision

7. GuardrailProvider     — checks inbound prompts and outbound responses
     check_inbound(prompt, context) → GuardrailResult
     check_outbound(response, context) → GuardrailResult

8. ContextManager        — manages LLM context window budget
     prepare_context(messages, tools, goal_state, model_config) → ManagedContext

9. CapabilityInvoker    — invokes capability by ID
     invoke(capability_id, arguments) → Any
```

---

## Implementation Phases

### Phase 1: Core Abstractions + In-Memory
Go interfaces for all 9 ABCs. In-memory implementations.
Single Go binary, zero external dependencies.

### Phase 2: Control Plane REST API
```
POST   /api/v1/capabilities/register
DELETE /api/v1/capabilities/{id}
GET    /api/v1/capabilities
POST   /api/v1/discover
POST   /api/v1/invoke/{capability_id}
POST   /api/v1/goals
GET    /api/v1/goals/{goal_id}
GET/POST/DELETE /api/v1/state/{scope}/{key}
POST   /api/v1/heartbeat
GET    /ui/
```

### Phase 3: Thin Proxy
```
cfproxy --mode=capability|agent --config=proxy.yaml

Capability mode: load manifest → register with CP → probe app → heartbeat → serve :3500
Agent mode:  connect to CP → cache registry → serve :3500
  POST /discover → cached registry → LLM tool format
  POST /invoke/{id} → auth → policy → guardrails → transport → trace → audit
  POST /llm/chat → inbound guardrails → context mgmt → LiteLLM → outbound guardrails
```

### Phase 4: Transport Adapters
InProcess, HTTP, gRPC, PubSub, Webhook, MCP — each: invoke, health_check, supports

### Phase 5: Capability Manifest YAML
```yaml
apiVersion: capafabric/v1
kind: CapabilityManifest
metadata:
  agent_id: stallion-cash-matching
  language: dotnet
app:
  port: 8081
  protocol: http
  health_path: /health
  base_path: /api
capabilities:
  - capability_id: stallion.retrieve_invoice_details
    name: retrieve_invoice_details
    description: >
      Retrieves invoice details from Oracle Accounts Receivable
      including line items, customer info, and payment status.
    endpoint:
      method: GET
      path: /invoices/{invoice_id}
      arguments:
        invoice_id: { in: path }
      response:
        from: body
    security:
      required_roles: [finance_analyst]
      classification: confidential
```

### Phase 6: Authentication & Authorization
AuthProvider: TokenAuth (JWT/JWKS), MTLSAuth, APIKeyAuth, DIDAuth
PolicyEnforcer: RBAC, ABAC, OPA (Rego)
Config: auth.provider, policy.enforcer in CP YAML

### Phase 7: OpenTelemetry Observability
InvocationTracer: on_invocation_start/end, on_auth_decision, on_discovery
Provided: OTelTracer, StructuredLogTracer
AuditWriter: FileAudit, CosmosAudit, S3Audit
Standard metrics: invocations_total, duration_ms, errors_total, auth_denied_total

### Phase 8: Guardrails
GuardrailProvider: PromptInjection, PIIRedaction, TopicRestriction,
OutputSafety, SchemaCompliance, DataLeakage
All off by default. YAML-activated. action: block | warn | log_only
Pipeline: inbound guardrails → context mgmt → LLM → outbound guardrails

### Phase 9: Context Window Management
ContextManager: Passthrough, SlidingWindow, Summarizing, RAG, Adaptive
Lazy trigger: compress only when history exceeds 80% of window budget
Uses cheaper model for summarization via LiteLLM

### Phase 10: State Management
Scopes: invocation (ephemeral), goal (durable), session (durable), global (permanent)
API: GET/POST/DELETE /state/{scope}/{key}, lock/unlock for distributed coordination
Backends: inmemory, redis, postgres, cosmos

### Phase 11: Load Balancing
Strategies: round_robin, least_connections, affinity (goal_id), weighted, priority
Health-aware: unhealthy instances excluded automatically

### Phase 12: LiteLLM Integration
LiteLLM Proxy as infrastructure container. Proxy forwards /llm/chat to LiteLLM.
Model routing, failover, cost tracking, provider normalization.
Reasoner calls localhost:3500 → proxy → LiteLLM → any LLM provider

---

## Proxy Pipeline (Complete Request Flow)

```
Agent → localhost:3500/llm/chat
  → INBOUND GUARDRAILS (parallel: injection, PII, topic)
  → CONTEXT MANAGEMENT (compress if budget exceeded)
  → FORWARD TO LLM (via LiteLLM)
  → OUTBOUND GUARDRAILS (parallel: safety, schema, leakage)
  → RETURN TO GOALAGENT

Agent → localhost:3500/invoke/{capability_id}
  → AUTHENTICATE (AuthProvider)
  → AUTHORIZE (PolicyEnforcer) → denied + requires_approval → HITL queue
  → START OTEL SPAN
  → RESOLVE TARGET (CP load balancer, cached)
  → ROUTE VIA TRANSPORT (http | grpc | pubsub | webhook | mcp)
  → END OTEL SPAN + AUDIT
  → RETURN RESULT
```

---

## Control Plane Resilience

```
Proxy caches ALL config from CP with TTL:
  Registry: 30s, Guardrails: 60s, Auth keys: 300s, LB table: 10s

CP outage → proxy continues on cached config
  New registrations queue. Health reports queue. State ops → 503 with retry.
  Log: "Operating on cached config, CP unreachable"
  Auto-reconnect when CP recovers.

Agent survives proxy outage via circuit breaker:
  Primary: localhost:3500. Fallback: cached capability list + direct HTTP.
  Fallback mode: NO auth, NO guardrails, NO tracing (deliberate trade-off).
  Configurable per-capability: direct_fallback: true|false
```

---

## Enterprise Adoption Levels

```
L0 — Manifest only      Write manifest.yaml + add proxy       App changes: NONE
L1 — Manifest + mapping  Add argument/response transforms      App changes: NONE
L2 — Proxy plugin        Go plugin for legacy (no API)         App changes: NONE
L3 — SDK integration     Use language SDK for new services      New code only
```

---

## Development Levels (Zero to Full Stack)

```
Level 0 — Agent + Ollama only       (2 processes)
Level 1 — Add proxy + capability         (4 processes)
Level 2 — Add LiteLLM                    (5 processes)
Level 3 — Add control plane + admin UI   (6 processes)
Level 4 — Add OTEL + Jaeger             (8 processes)
Level 5 — Add Redis + full governance    (9 processes)
Level 6 — Docker Compose                 (everything)
```

---

## Extension Points

| What to Extend | Interface | Inject Via |
|---|---|---|
| Storage backend | `CapabilityRegistry` | CP config: registry.type |
| Discovery logic | `DiscoveryProvider` | CP config or plugin |
| LLM tool format | `ToolFormatAdapter` | Proxy config: provider |
| Wire protocol | `TransportAdapter` | Proxy plugin |
| Authentication | `AuthProvider` | CP config: auth.provider |
| Authorization | `PolicyEnforcer` | CP config: policy.enforcer |
| Guardrail algorithm | `GuardrailProvider` | CP config: guardrails.inbound[].provider |
| Context strategy | `ContextManager` | CP config: context_management.strategy |
| Cognitive engine | `CognitiveProvider` | SDK DI: per Agent, per invocation |
| Observability | `InvocationTracer` | CP config: observability.tracer |
| Audit storage | `AuditWriter` | CP config: audit.writer |
| State store | CP config | state.type |
| Load balancing | CP config | load_balancing.strategy |
| New capability | Manifest YAML | Write manifest + add proxy |

Every seam accepts the abstract interface. No concrete class is ever required.
Every cross-cutting concern is off by default, YAML-activated, independently evolvable.

**Component swappability**: Every component (CP, Proxy, SDKs, Admin UI) communicates
via language-neutral contracts (OpenAPI specs, gRPC protos, JSON Schemas) stored in
`contracts/`. Any component can be rewritten in any language at any time. Contract
tests in `contract-tests/` validate any implementation against the specs.

---

## Compatibility Matrix

| Agent Framework | Integration |
|---|---|
| .NET Semantic Kernel | Proxy on localhost:3500 |
| Python LangGraph | Proxy on localhost:3500 |
| Python CrewAI / AutoGen | Proxy on localhost:3500 |
| TypeScript / Node.js | Proxy on localhost:3500 |
| Custom (any language) | HTTP to localhost:3500 |

| Capability Language | Integration |
|---|---|
| .NET / Go / Java / TypeScript / Node / Python | manifest.yaml + proxy |
| Legacy (no API) | manifest.yaml + proxy plugin |

| SDK | Package |
|---|---|
| .NET 10+ | NuGet: CapaFabric.Client |
| TypeScript | npm: @capafabric/client |
| Python | PyPI: capafabric |
| Go | github.com/psiog/capafabric-go |
| Java 23+ | Maven: com.psiog.capafabric |

| LLM Provider | Integration |
|---|---|
| OpenAI / Azure / Anthropic / Gemini / Bedrock / Ollama / vLLM | Via CognitiveProvider → LiteLLM Proxy |

---

## Implementation Order

```
Week 1:  CP skeleton + in-memory registry + REST API + proxy skeleton
Week 2:  .NET Agent with Semantic Kernel + proxy integration
Week 3:  LiteLLM + manifest parser + endpoint mapping
Week 4:  Auth + policy + config push to proxies
Week 5:  Guardrails + OTEL + Jaeger
Week 6:  Context management + Redis state + audit
Week 7:  Admin UI + health monitoring + load balancing
Week 8:  Resilience (cache, circuit breaker, CP replication) + Docker Compose
```

---

## Dependencies

```
Control Plane (Go): Go 1.23+, single static binary
Thin Proxy (Go): Go 1.23+, ~10MB binary, ~64MB RAM
Admin UI (optional): React + Tailwind + shadcn/ui, embedded in CP via go:embed
.NET SDK: .NET 10+ (LTS), Microsoft.SemanticKernel (portability: net8.0 on need basis)
Python SDK: Python 3.10+, pydantic, httpx
Java SDK: Java 23+ / OpenJDK 23+ (portability: Java 17 on need basis)
Infrastructure: LiteLLM, Ollama (dev), Redis/Postgres/Cosmos (prod),
                OTEL Collector + Jaeger/Tempo (observability), OPA (optional)
```
