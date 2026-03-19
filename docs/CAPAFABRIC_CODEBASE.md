# CapaFabric — Codebase Structure

> Complete Go project layout for the CapaFabric Control Plane and Thin Proxy.
> Both are single-binary Go projects following the Standard Go Project Layout.
> Drop as `CLAUDE.md` in the `capafabric/` monorepo root.

---

## Monorepo Structure

```
capafabric/
├── CLAUDE.md                                   # This file
├── go.work                                     # Go workspace (links CP + proxy)
├── go.work.sum
├── Makefile                                    # Build, test, lint, run targets
├── LICENSE
├── README.md
│
├── shared/                                     # Shared Go packages (used by both CP and proxy)
│   ├── go.mod                                  # module github.com/psiog/capafabric/shared
│   ├── models/
│   │   ├── capability.go                       # CapabilityMetadata, CapabilitySource enum
│   │   ├── transport.go                        # TransportConfig
│   │   ├── security.go                         # SecurityConfig
│   │   ├── invocation.go                       # InvocationContext, InvocationResult
│   │   ├── identity.go                         # AuthIdentity
│   │   ├── policy.go                           # PolicyDecision
│   │   ├── goal.go                             # AgentDecision, AgentState
│   │   ├── guardrail.go                        # GuardrailResult
│   │   └── manifest.go                         # CapabilityManifest (parsed from YAML)
│   ├── interfaces/
│   │   ├── registry.go                         # CapabilityRegistry interface
│   │   ├── discovery.go                        # DiscoveryProvider interface
│   │   ├── adapter.go                          # ToolFormatAdapter interface
│   │   ├── transport.go                        # TransportAdapter interface
│   │   ├── auth.go                             # AuthProvider interface
│   │   ├── policy.go                           # PolicyEnforcer interface
│   │   ├── guardrail.go                        # GuardrailProvider interface
│   │   ├── context.go                          # ContextManager interface
│   │   ├── invoker.go                         # CapabilityInvoker interface
│   │   ├── state.go                            # StateStore interface
│   │   ├── audit.go                            # AuditWriter interface
│   │   └── tracer.go                           # InvocationTracer interface
│   ├── manifest/
│   │   ├── parser.go                           # YAML manifest parser
│   │   ├── validator.go                        # JSON Schema validation
│   │   └── parser_test.go
│   ├── otel/
│   │   ├── spans.go                            # Shared OTEL span helpers
│   │   ├── metrics.go                          # Shared metric definitions
│   │   └── propagation.go                      # Trace context propagation
│   └── errors/
│       ├── errors.go                           # Structured error types
│       └── codes.go                            # Error code constants
│
├── control-plane/                              # The CapaFabric Control Plane
│   ├── go.mod                                  # module github.com/psiog/capafabric/control-plane
│   ├── go.sum
│   │
│   ├── cmd/
│   │   └── capafabric/
│   │       └── main.go                         # Entry point
│   │           # 1. Parse flags: --config, --port, --dev
│   │           # 2. Load config YAML
│   │           # 3. Initialize all components via DI
│   │           # 4. Start HTTP server + gRPC config stream
│   │           # 5. Start health monitor goroutine
│   │           # 6. Block until SIGINT/SIGTERM
│   │           # 7. Graceful shutdown
│   │
│   ├── internal/
│   │   │
│   │   ├── config/
│   │   │   ├── config.go                       # Config struct (maps to YAML)
│   │   │   ├── loader.go                       # YAML loader + env var substitution
│   │   │   ├── defaults.go                     # Sensible defaults for every field
│   │   │   ├── validator.go                    # Startup validation with actionable errors
│   │   │   └── loader_test.go
│   │   │
│   │   ├── registry/
│   │   │   ├── inmemory.go                     # InMemoryCapabilityRegistry
│   │   │   ├── inmemory_test.go
│   │   │   ├── redis.go                        # RedisCapabilityRegistry
│   │   │   ├── redis_test.go
│   │   │   ├── postgres.go                     # PostgresCapabilityRegistry
│   │   │   └── postgres_test.go
│   │   │   # Each implements shared/interfaces/CapabilityRegistry
│   │   │   # Factory: NewRegistry(config) → CapabilityRegistry
│   │   │
│   │   ├── discovery/
│   │   │   ├── full.go                         # FullDiscoveryProvider (return all, filter)
│   │   │   ├── semantic.go                     # SemanticDiscoveryProvider (keyword search)
│   │   │   ├── full_test.go
│   │   │   └── semantic_test.go
│   │   │   # Each implements shared/interfaces/DiscoveryProvider
│   │   │
│   │   ├── policy/
│   │   │   ├── auth/
│   │   │   │   ├── jwt.go                      # TokenAuthProvider (JWT + JWKS)
│   │   │   │   ├── jwt_test.go
│   │   │   │   ├── mtls.go                     # MTLSAuthProvider
│   │   │   │   ├── apikey.go                   # APIKeyAuthProvider
│   │   │   │   └── factory.go                  # NewAuthProvider(config) → AuthProvider
│   │   │   │   # Each implements shared/interfaces/AuthProvider
│   │   │   │
│   │   │   ├── enforcer/
│   │   │   │   ├── rbac.go                     # RBACPolicyEnforcer
│   │   │   │   ├── rbac_test.go
│   │   │   │   ├── opa.go                      # OPAPolicyEnforcer (HTTP to OPA server)
│   │   │   │   ├── opa_test.go
│   │   │   │   └── factory.go                  # NewPolicyEnforcer(config) → PolicyEnforcer
│   │   │   │   # Each implements shared/interfaces/PolicyEnforcer
│   │   │   │
│   │   │   └── ratelimit/
│   │   │       ├── sliding_window.go           # Per-caller sliding window rate limiter
│   │   │       └── sliding_window_test.go
│   │   │
│   │   ├── guardrails/
│   │   │   ├── engine.go                       # CompositeGuardrailProvider (chains multiple)
│   │   │   ├── engine_test.go
│   │   │   ├── prompt_injection.go             # Pattern-based prompt injection detection
│   │   │   ├── prompt_injection_test.go
│   │   │   ├── pii_redaction.go                # PII/PHI redaction (regex + Presidio integration)
│   │   │   ├── pii_redaction_test.go
│   │   │   ├── topic_restriction.go            # Allowed/blocked topic enforcement
│   │   │   ├── output_safety.go                # Toxicity/safety check (Azure Content Safety)
│   │   │   ├── schema_compliance.go            # Structured output validation
│   │   │   ├── data_leakage.go                 # Sensitive data in output prevention
│   │   │   └── factory.go                      # NewGuardrailEngine(config) → GuardrailProvider
│   │   │   # Each implements shared/interfaces/GuardrailProvider
│   │   │   # Engine runs inbound guardrails in parallel, outbound in parallel
│   │   │
│   │   ├── context/
│   │   │   ├── passthrough.go                  # No modification (default)
│   │   │   ├── sliding_window.go               # Keep recent N, summarize older
│   │   │   ├── sliding_window_test.go
│   │   │   ├── summarizing.go                  # Compress all history each iteration
│   │   │   ├── rag.go                          # Vector store retrieval
│   │   │   ├── adaptive.go                     # Auto-select based on model window + iteration
│   │   │   └── factory.go                      # NewContextManager(config) → ContextManager
│   │   │   # Each implements shared/interfaces/ContextManager
│   │   │
│   │   ├── health/
│   │   │   ├── monitor.go                      # Background goroutine: check all capabilities
│   │   │   ├── monitor_test.go
│   │   │   ├── heartbeat.go                    # Track heartbeat timestamps per agent_id
│   │   │   └── heartbeat_test.go
│   │   │   # Unhealthy capabilities marked unavailable (not unregistered)
│   │   │   # TTL expiry: no heartbeat within window → evict
│   │   │
│   │   ├── loadbalancer/
│   │   │   ├── round_robin.go
│   │   │   ├── round_robin_test.go
│   │   │   ├── least_connections.go
│   │   │   ├── affinity.go                     # Consistent hash on goal_id
│   │   │   ├── affinity_test.go
│   │   │   ├── weighted.go
│   │   │   ├── priority.go                     # HITL-escalated → dedicated instance
│   │   │   └── factory.go                      # NewLoadBalancer(config) → LoadBalancer
│   │   │
│   │   ├── state/
│   │   │   ├── inmemory.go                     # InMemoryStateStore
│   │   │   ├── inmemory_test.go
│   │   │   ├── redis.go                        # RedisStateStore
│   │   │   ├── redis_test.go
│   │   │   ├── lock.go                         # Distributed lock implementation
│   │   │   ├── lock_test.go
│   │   │   └── factory.go                      # NewStateStore(config) → StateStore
│   │   │   # Each implements shared/interfaces/StateStore
│   │   │
│   │   ├── audit/
│   │   │   ├── writer_file.go                  # Append-only JSONL file
│   │   │   ├── writer_cosmos.go                # Azure Cosmos DB
│   │   │   ├── writer_postgres.go              # PostgreSQL
│   │   │   ├── aggregator.go                   # Collects audit records from proxies
│   │   │   └── factory.go                      # NewAuditWriter(config) → AuditWriter
│   │   │   # Each implements shared/interfaces/AuditWriter
│   │   │
│   │   ├── configdist/
│   │   │   ├── distributor.go                  # Pushes config to proxies (gRPC stream or poll)
│   │   │   ├── distributor_test.go
│   │   │   └── snapshot.go                     # Config snapshot: registry + guardrails + policy
│   │   │   # Proxies connect and receive config updates in real-time
│   │   │   # On proxy reconnect: full snapshot sent, then incremental
│   │   │
│   │   ├── api/
│   │   │   ├── router.go                       # HTTP router setup (chi or stdlib mux)
│   │   │   ├── middleware.go                    # Logging, recovery, CORS, request ID
│   │   │   │
│   │   │   ├── handlers/
│   │   │   │   ├── capabilities.go             # POST /register, DELETE /{id}, GET /
│   │   │   │   ├── discover.go                 # POST /discover
│   │   │   │   ├── invoke.go                  # POST /invoke/{capability_id}
│   │   │   │   ├── goals.go                    # POST /goals, GET /goals/{id}
│   │   │   │   ├── state.go                    # GET/POST/DELETE /state/{scope}/{key}
│   │   │   │   ├── heartbeat.go                # POST /heartbeat
│   │   │   │   ├── health.go                   # GET /health, GET /health/capabilities
│   │   │   │   ├── admin.go                    # GET /admin/config, GET /admin/metrics
│   │   │   │   └── proxy_config.go             # GET /proxy/config (proxies pull config)
│   │   │   │
│   │   │   └── dto/
│   │   │       ├── requests.go                 # Request DTOs (RegisterRequest, DiscoverRequest, etc.)
│   │   │       └── responses.go                # Response DTOs (DiscoverResponse, InvokeResponse, etc.)
│   │   │
│   │   └── ui/
│   │       ├── embed.go                        # go:embed for static files
│   │       └── dist/                           # Built admin UI (React static files)
│   │           ├── index.html
│   │           ├── assets/
│   │           └── ...
│   │           # Admin UI is OPTIONAL. Embedded in the CP binary via go:embed.
│   │           # Accessible at GET /ui/ when admin_ui: true in config.
│   │           #
│   │           # PAGES:
│   │           #   Dashboard        — overview: capabilities, goals, health, metrics
│   │           #   Capabilities     — table: registered capabilities, health status,
│   │           #                      language, transport, last heartbeat
│   │           #   Goals            — active goals: status, iteration, duration, trace link
│   │           #   Guardrails       — activity: blocked, warned, passed counts
│   │           #   Auth & Policy    — decisions: allowed, denied, rate limited
│   │           #   Traces           — OTEL trace links → click opens Jaeger/Tempo
│   │           #   Config           — view/edit CP YAML config (live reload)
│   │           #
│   │           # MANIFEST EDITOR (drag-and-drop):
│   │           #   The visual manifest editor allows non-developers to:
│   │           #   - Drag existing API endpoints from a discovered OpenAPI spec
│   │           #     (or manually entered endpoint list) onto a canvas
│   │           #   - Drop them to create capability entries
│   │           #   - Fill in name, description, tags via guided form fields
│   │           #   - Map arguments (path/query/body/header) via dropdowns
│   │           #   - Map response fields via JSON path picker
│   │           #   - Set security config via checkboxes (roles, classification, audit)
│   │           #   - Preview the generated manifest.yaml in real-time
│   │           #   - Download the manifest.yaml or save directly to a Git repo
│   │           #   - Validate the manifest against the live application endpoint
│   │           #
│   │           #   This lowers the manifest authoring barrier from "YAML-literate
│   │           #   developer" to "anyone who knows their API endpoints."
│   │           #
│   │           # TECH STACK:
│   │           #   React + Tailwind + shadcn/ui
│   │           #   Built separately: cd ui && npm run build → copies to dist/
│   │           #   Embedded in Go binary via go:embed (zero runtime dependency)
│   │           #   API calls: /api/v1/* (same CP REST API)
│   │
│   ├── ui/                                     # Admin UI source (React + Tailwind)
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   ├── vite.config.ts
│   │   ├── tailwind.config.js
│   │   ├── src/
│   │   │   ├── App.tsx
│   │   │   ├── main.tsx
│   │   │   ├── pages/
│   │   │   │   ├── Dashboard.tsx              # Overview metrics + health
│   │   │   │   ├── Capabilities.tsx           # Registered capabilities table
│   │   │   │   ├── Goals.tsx                  # Active goals list
│   │   │   │   ├── Guardrails.tsx             # Guardrail activity
│   │   │   │   ├── Config.tsx                 # YAML config viewer/editor
│   │   │   │   └── ManifestEditor.tsx         # Drag-and-drop manifest builder
│   │   │   ├── components/
│   │   │   │   ├── ManifestCanvas.tsx         # Drop zone for endpoint → capability
│   │   │   │   ├── EndpointCard.tsx           # Draggable endpoint card
│   │   │   │   ├── CapabilityForm.tsx         # Guided form: name, description, tags
│   │   │   │   ├── ArgumentMapper.tsx         # in: path/query/body/header dropdowns
│   │   │   │   ├── ResponseMapper.tsx         # JSON path picker for response.from
│   │   │   │   ├── SecurityForm.tsx           # Roles, classification, audit checkboxes
│   │   │   │   ├── YamlPreview.tsx            # Live YAML preview panel
│   │   │   │   └── ManifestValidator.tsx      # Validate against live app endpoint
│   │   │   └── api/
│   │   │       └── client.ts                  # Fetch wrapper for /api/v1/*
│   │   └── public/
│   │   # Build: cd ui && npm install && npm run build
│   │   # Output: copies to internal/ui/dist/ → embedded via go:embed
│   │   # OPTIONAL: UI is not required for CP operation.
│   │   # CP works fully via REST API without the UI.
│   │
│   ├── configs/
│   │   ├── capafabric.yaml                     # Production config template
│   │   └── capafabric.dev.yaml                 # Dev config (in-memory everything)
│   │
│   └── Dockerfile                              # Full build WITH Admin UI
│       # FROM node:22-alpine AS ui-builder
│       # WORKDIR /ui
│       # COPY ui/package*.json .
│       # RUN npm ci
│       # COPY ui/ .
│       # RUN npm run build
│       #
│       # FROM golang:1.23-alpine AS builder
│       # COPY . .
│       # COPY --from=ui-builder /ui/dist ./internal/ui/dist
│       # RUN go build -o /capafabric ./cmd/capafabric
│       #
│       # FROM alpine:latest
│       # COPY --from=builder /capafabric /usr/local/bin/capafabric
│       # ENTRYPOINT ["capafabric"]
│   │
│   └── Dockerfile.slim                         # Slim build WITHOUT Admin UI
│       # FROM golang:1.23-alpine AS builder
│       # COPY . .
│       # RUN go build -tags no_ui -o /capafabric ./cmd/capafabric
│       #
│       # FROM alpine:latest
│       # COPY --from=builder /capafabric /usr/local/bin/capafabric
│       # ENTRYPOINT ["capafabric"]
│       #
│       # Build tags:
│       #   Default (no tag):  embeds Admin UI from internal/ui/dist/
│       #   -tags no_ui:       serves GET /ui/ → 200 "Admin UI not included.
│       #                      Build with: make build-with-ui"
│       #
│       #   control-plane/internal/ui/embed.go:
│       #     //go:build !no_ui
│       #     //go:embed dist/*
│       #     var uiFiles embed.FS
│       #
│       #   control-plane/internal/ui/embed_no_ui.go:
│       #     //go:build no_ui
│       #     // Returns placeholder response for /ui/ routes
│
├── proxy/                                      # The CapaFabric Thin Proxy
│   ├── go.mod                                  # module github.com/psiog/capafabric/proxy
│   ├── go.sum
│   │
│   ├── cmd/
│   │   └── cfproxy/
│   │       └── main.go                         # Entry point
│   │           # 1. Parse flags: --mode, --config, --manifest, --port
│   │           # 2. Load config YAML
│   │           # 3. Mode switch:
│   │           #    capability: load manifest → register → probe app → heartbeat → serve
│   │           #    agent:  connect to CP → cache config → serve
│   │           # 4. Start localhost HTTP server
│   │           # 5. Block until signal
│   │
│   ├── internal/
│   │   │
│   │   ├── config/
│   │   │   ├── config.go                       # ProxyConfig struct
│   │   │   ├── loader.go                       # YAML loader
│   │   │   └── defaults.go                     # Defaults per mode (capability vs agent)
│   │   │
│   │   ├── cache/
│   │   │   ├── config_cache.go                 # TTL-based config cache from CP
│   │   │   ├── config_cache_test.go
│   │   │   └── refresh.go                      # Background goroutine: poll CP or receive push
│   │   │   # On CP disconnect: serve from cache
│   │   │   # On CP reconnect: refresh immediately
│   │   │
│   │   ├── pipeline/
│   │   │   ├── pipeline.go                     # Composable request pipeline
│   │   │   ├── pipeline_test.go
│   │   │   └── stage.go                        # PipelineStage interface
│   │   │   # Pipeline for /llm/chat:
│   │   │   #   InboundGuardrails → ContextManagement → LLMForward → OutboundGuardrails
│   │   │   # Pipeline for /invoke:
│   │   │   #   Auth → Policy → OTelStart → Transport → OTelEnd → Audit
│   │   │   # Stages are composable. Disabled stages are no-ops (zero overhead).
│   │   │
│   │   ├── transport/
│   │   │   ├── http.go                         # HttpTransportAdapter
│   │   │   ├── http_test.go
│   │   │   ├── grpc.go                         # GrpcTransportAdapter
│   │   │   ├── pubsub.go                       # PubSubTransportAdapter
│   │   │   ├── webhook.go                      # WebhookTransportAdapter
│   │   │   ├── mcp.go                          # MCPTransportAdapter (JSON-RPC)
│   │   │   ├── inprocess.go                    # InProcessTransportAdapter (Python SDK only)
│   │   │   └── factory.go                      # NewTransportAdapter(kind) → TransportAdapter
│   │   │   # Each implements shared/interfaces/TransportAdapter
│   │   │
│   │   ├── manifest/
│   │   │   ├── loader.go                       # Load + parse manifest YAML
│   │   │   ├── loader_test.go
│   │   │   ├── endpoint_mapper.go              # Map capability args → HTTP request
│   │   │   ├── endpoint_mapper_test.go
│   │   │   └── response_mapper.go              # Map HTTP response → capability result
│   │   │   # Handles: path params, query params, body params, response.from
│   │   │
│   │   ├── registration/
│   │   │   ├── registrar.go                    # Register capabilities with CP on startup
│   │   │   ├── registrar_test.go
│   │   │   └── heartbeat.go                    # Periodic heartbeat to CP
│   │   │   # Probe app health before registering
│   │   │   # Retry with backoff if CP is unavailable
│   │   │   # Unregister on graceful shutdown
│   │   │
│   │   ├── guardrails/
│   │   │   ├── invoker.go                     # Invoke guardrail rules from cached config
│   │   │   └── invoker_test.go
│   │   │   # Rules come from CP config (not embedded)
│   │   │   # Prompt injection: pattern matching (local, no external call)
│   │   │   # PII redaction: regex-based (local) or Presidio call (configurable)
│   │   │   # Output safety: Azure Content Safety call (configurable)
│   │   │   # All run in parallel. Cached results for identical inputs.
│   │   │
│   │   ├── context/
│   │   │   ├── manager.go                      # Invoke context management from cached config
│   │   │   └── manager_test.go
│   │   │   # Strategy comes from CP config
│   │   │   # Summarization calls cheaper model via LiteLLM
│   │   │   # Lazy trigger: only when > 80% of window budget
│   │   │
│   │   ├── otel/
│   │   │   ├── tracer.go                       # OTEL span creation + attribute setting
│   │   │   ├── tracer_test.go
│   │   │   ├── metrics.go                      # Prometheus counters + histograms
│   │   │   └── propagation.go                  # W3C trace context propagation
│   │   │   # Span hierarchy:
│   │   │   #   [goal.invoke] → [capability.invoke X] → [transport.http]
│   │   │
│   │   ├── llm/
│   │   │   ├── forwarder.go                    # Forward /llm/chat to LiteLLM
│   │   │   ├── forwarder_test.go
│   │   │   └── health.go                       # LiteLLM health check + fallback endpoint
│   │   │   # Pipeline: inbound guardrails → context → forward → outbound guardrails
│   │   │
│   │   ├── fallback/
│   │   │   ├── circuit_breaker.go              # Circuit breaker for CP connectivity
│   │   │   └── circuit_breaker_test.go
│   │   │   # Trips after 3 failures. Resets on success.
│   │   │   # When tripped: serve from cache, queue registrations.
│   │   │
│   │   └── api/
│   │       ├── router.go                       # Localhost HTTP router
│   │       ├── middleware.go                    # Request ID, logging
│   │       │
│   │       └── handlers/
│   │           ├── discover.go                 # POST /discover (agent mode)
│   │           ├── invoke.go                  # POST /invoke/{capability_id} (both modes)
│   │           ├── llm.go                      # POST /llm/chat (agent mode)
│   │           ├── state.go                    # GET/POST/DELETE /state/{scope}/{key}
│   │           ├── health.go                   # GET /health, GET /health/app
│   │           └── metrics.go                  # GET /metrics (Prometheus)
│   │
│   ├── configs/
│   │   ├── proxy-agent.yaml                # Agent mode config
│   │   └── proxy-capability.yaml               # Capability mode config
│   │
│   └── Dockerfile
│       # Same multi-stage as CP. ~10MB final binary.
│
├── contracts/                                  # ══ LANGUAGE-NEUTRAL CONTRACTS ══
│   │                                           #
│   │                                           # These files ARE the architecture.
│   │                                           # Any component in any language is a valid
│   │                                           # drop-in replacement if it conforms to
│   │                                           # these contracts. The contracts are the
│   │                                           # INVARIANT. The implementations are VARIANTS.
│   │
│   ├── openapi/                                # REST API contracts (OpenAPI 3.1)
│   │   ├── control-plane-api.openapi.yaml      # Full CP REST API spec
│   │   │   # Every endpoint: /capabilities/register, /discover, /invoke,
│   │   │   # /goals, /state, /heartbeat, /health, /admin, /proxy/config
│   │   │   # Request/response schemas, error codes, auth headers
│   │   │   # ANY language implementing this spec IS a valid Control Plane
│   │   │
│   │   ├── proxy-api.openapi.yaml              # Full Proxy localhost API spec
│   │   │   # Every endpoint: /discover, /invoke, /llm/chat, /state, /health
│   │   │   # Request/response schemas, error codes
│   │   │   # ANY language implementing this spec IS a valid Proxy
│   │   │
│   │   └── admin-ui-api.openapi.yaml           # API surface the Admin UI depends on
│   │       # Subset of CP API + additional admin endpoints
│   │       # ANY frontend framework calling these endpoints IS a valid Admin UI
│   │
│   ├── grpc/                                   # gRPC contracts (Protocol Buffers)
│   │   ├── capability_invoke.proto             # Capability invocation service
│   │   │   # service CapabilityService {
│   │   │   #   rpc Invoke(CapabilityRequest) returns (CapabilityResponse);
│   │   │   #   rpc HealthCheck(HealthRequest) returns (HealthResponse);
│   │   │   # }
│   │   │
│   │   ├── proxy_config.proto                  # CP → Proxy config distribution
│   │   │   # service ConfigService {
│   │   │   #   rpc Subscribe(ConfigRequest) returns (stream ConfigSnapshot);
│   │   │   # }
│   │   │
│   │   ├── health_report.proto                 # Proxy → CP health reports
│   │   │   # service HealthReportService {
│   │   │   #   rpc Report(HealthReport) returns (HealthAck);
│   │   │   # }
│   │   │
│   │   └── cognitive.proto                     # CognitiveProvider gRPC contract
│   │       # service CognitiveService {
│   │       #   rpc Deliberate(CognitiveRequest) returns (CognitiveResult);
│   │       #   rpc DeliberateStream(CognitiveRequest) returns (stream CognitiveToken);
│   │       # }
│   │
│   ├── jsonschema/                             # JSON Schema contracts
│   │   ├── capability_manifest.schema.json     # Validates manifest YAML
│   │   │   # Used by: manifest parser (any language)
│   │   │   # Used by: CI pipeline to validate manifests
│   │   │   # Used by: Admin UI manifest editor validation
│   │   │
│   │   ├── capability_result.schema.json       # Invocation result envelope
│   │   │   # { request_id, capability_id, success, result, error, duration_ms }
│   │   │
│   │   ├── agent_state.schema.json             # AgentState structure
│   │   ├── agent_decision.schema.json          # AgentDecision structured output
│   │   ├── agent_context.schema.json           # AgentContext with call_chain + invocation_log
│   │   ├── invocation_context.schema.json      # InvocationContext
│   │   ├── auth_identity.schema.json           # AuthIdentity
│   │   ├── policy_decision.schema.json         # PolicyDecision
│   │   ├── guardrail_result.schema.json        # GuardrailResult
│   │   └── cognitive_request.schema.json       # CognitiveRequest / CognitiveResult
│   │
│   ├── headers/                                # Standard HTTP headers (proxy enforced)
│   │   └── call_chain.md                       # Call chain propagation contract:
│   │       # X-CapaFabric-Call-Chain: comma-separated agent IDs
│   │       #   e.g., "supervisor,ingestion,extraction"
│   │       # X-CapaFabric-Depth: current nesting depth (integer)
│   │       # X-CapaFabric-Max-Depth: maximum allowed (default: 10)
│   │       #
│   │       # Proxy behavior:
│   │       #   Target in chain → 409 Conflict (circular)
│   │       #   Depth >= max → 429 Too Deep
│   │       #   Otherwise → append current agent, forward
│   │       #
│   │       # These headers are also defined in:
│   │       #   contracts/openapi/proxy-api.openapi.yaml
│   │
│   └── README.md                               # Contract governance rules:
│       # 1. Contracts are versioned (apiVersion field)
│       # 2. Breaking changes require major version bump
│       # 3. All implementations must pass contract tests
│       # 4. Contracts are the source of truth — code is generated from them
│       # 5. SDK models can be auto-generated from JSON Schemas
│       # 6. gRPC stubs can be auto-generated from .proto files
│
├── contract-tests/                             # ══ CONTRACT VALIDATION TESTS ══
│   │                                           #
│   │                                           # Language-agnostic tests that validate
│   │                                           # ANY implementation against the contracts.
│   │                                           # Run these against Go v1, Rust v2, etc.
│   │
│   ├── cp-api/
│   │   ├── test_register_capability.py         # POST /capabilities/register
│   │   ├── test_discover.py                    # POST /discover
│   │   ├── test_invoke.py                     # POST /invoke/{id}
│   │   ├── test_goals.py                       # POST /goals, GET /goals/{id}
│   │   ├── test_state.py                       # GET/POST/DELETE /state
│   │   ├── test_health.py                      # GET /health
│   │   └── conftest.py                         # CP_BASE_URL from env (point at any impl)
│   │
│   ├── proxy-api/
│   │   ├── test_discover.py                    # POST /discover
│   │   ├── test_invoke.py                     # POST /invoke/{id}
│   │   ├── test_llm_chat.py                    # POST /llm/chat
│   │   ├── test_state.py                       # GET/POST/DELETE /state
│   │   ├── test_call_chain.py                  # Loop protection header enforcement:
│   │   │   # test_circular_chain_returns_409
│   │   │   # test_max_depth_returns_429
│   │   │   # test_chain_propagated_to_downstream
│   │   │   # test_missing_chain_header_starts_new_chain
│   │   └── conftest.py                         # PROXY_BASE_URL from env
│   │
│   ├── manifest/
│   │   ├── test_manifest_validation.py         # Validate sample manifests against schema
│   │   └── sample_manifests/                   # Valid + invalid manifest examples
│   │
│   ├── requirements.txt                        # pytest, httpx, jsonschema
│   └── README.md
│       # Usage:
│       #   CP_BASE_URL=http://localhost:8080 pytest contract-tests/cp-api/
│       #   PROXY_BASE_URL=http://localhost:3500 pytest contract-tests/proxy-api/
│       #
│       # These tests work against ANY implementation:
│       #   Go v1, Rust v2, Java v3 — as long as it passes, it's valid.
│
├── sdk/                                        # ══ LANGUAGE SDKs ══
│   │                                           #
│   │                                           # Thin HTTP client wrappers over the proxy API.
│   │                                           # Models can be AUTO-GENERATED from:
│   │                                           #   contracts/jsonschema/*.schema.json
│   │                                           #   contracts/openapi/proxy-api.openapi.yaml
│   │                                           #
│   │                                           # Tools: openapi-generator, nswag, oapi-codegen,
│   │                                           #        quicktype (JSON Schema → any language)
│   │
│   ├── dotnet/                                 # NuGet: CapaFabric.Client (.NET 10+)
│   │   ├── src/
│   │   │   └── CapaFabric.Client/
│   │   │       ├── CapaFabric.Client.csproj
│   │   │       ├── Agent.cs                     # Abstract Agent base class
│   │   │       │   # PursueGoalAsync: discover → guard → deliberate → invoke → evaluate
│   │   │       │   # Override points: OnNoCapabilitiesFound, OnLowConfidence,
│   │   │       │   #   GetCognitiveConfig, GetConfidenceThreshold
│   │   │       │   # Loop guards: circular chain, max depth, retry detection
│   │   │       ├── AgentRunner.cs               # Iteration loop wrapper (max iter, budget, checkpoint)
│   │   │       ├── AgentResult.cs               # Completed | Continue | EscalateToHuman | Retry | Failed
│   │   │       ├── AgentContext.cs              # CallChain, InvocationLog, MaxDepth, MaxRetries
│   │   │       ├── ICapabilityDiscoveryClient.cs
│   │   │       ├── ICapabilityInvocationClient.cs
│   │   │       ├── ICognitiveProvider.cs        # Deliberate() polymorphic interface
│   │   │       ├── IStateClient.cs
│   │   │       ├── Models/                      # Can be generated from JSON Schemas
│   │   │       │   ├── CapabilityMetadata.cs
│   │   │       │   ├── AgentDecision.cs
│   │   │       │   ├── AgentState.cs
│   │   │       │   ├── InvocationContext.cs
│   │   │       │   ├── CognitiveRequest.cs
│   │   │       │   └── CognitiveResult.cs
│   │   │       ├── Providers/                   # CognitiveProvider implementations
│   │   │       │   ├── LiteLLMCognitiveProvider.cs
│   │   │       │   ├── OllamaCognitiveProvider.cs
│   │   │       │   ├── SemanticKernelCognitiveProvider.cs
│   │   │       │   └── MockCognitiveProvider.cs
│   │   │       ├── ProxyDiscoveryClient.cs
│   │   │       ├── ProxyInvocationClient.cs
│   │   │       ├── ProxyStateClient.cs
│   │   │       └── InMemoryStateClient.cs
│   │   └── tests/
│   │       └── CapaFabric.Client.Tests/
│   │
│   ├── typescript/                              # npm: @capafabric/client (TypeScript/Node.js)
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   ├── src/
│   │   │   ├── index.ts                        # Public API exports
│   │   │   ├── agent.ts                        # Abstract Agent base class
│   │   │   │   # Same pattern as .NET: discover → guard → deliberate → invoke
│   │   │   │   # TypeScript: abstract class + method override
│   │   │   ├── agent-runner.ts                 # AgentRunner iteration loop
│   │   │   ├── agent-result.ts                 # AgentResult union type
│   │   │   ├── agent-context.ts                # AgentContext with callChain + invocationLog
│   │   │   ├── client.ts                       # CapaFabricClient (discover, invoke, state)
│   │   │   ├── cognitive.ts                    # ICognitiveProvider + Deliberate overloads
│   │   │   ├── models/                         # Can be generated from JSON Schemas
│   │   │   │   ├── capability.ts
│   │   │   │   ├── agent.ts                    # AgentDecision, AgentState
│   │   │   │   ├── invocation.ts
│   │   │   │   ├── cognitive.ts
│   │   │   │   └── security.ts
│   │   │   ├── providers/
│   │   │   │   ├── litellm.ts
│   │   │   │   ├── ollama.ts
│   │   │   │   └── openai.ts
│   │   │   └── utils/
│   │   │       ├── http.ts                     # Fetch wrapper with retry + circuit breaker
│   │   │       └── streaming.ts                # SSE/streaming response handler
│   │   ├── tests/
│   │   │   ├── agent.test.ts                   # Agent base class + loop guard tests
│   │   │   └── client.test.ts
│   │   └── README.md
│   │       # Usage:
│   │       #   import { CapaFabricClient } from '@capafabric/client'
│   │       #   const client = new CapaFabricClient('http://localhost:3500')
│   │       #   const capabilities = await client.discover({ goal: '...' })
│   │       #   const result = await client.invoke('stallion.fetch_invoice', { invoice_id: '...' })
│   │
│   ├── python/                                  # PyPI: capafabric
│   │   ├── pyproject.toml
│   │   └── capafabric/
│   │       ├── __init__.py
│   │       ├── agent.py                        # Abstract Agent base class (ABC)
│   │       ├── agent_runner.py                 # AgentRunner iteration loop
│   │       ├── agent_context.py                # AgentContext with call_chain + invocation_log
│   │       ├── client.py                       # ProxyClient (discover, invoke, state)
│   │       ├── cognitive.py                    # ICognitiveProvider + implementations
│   │       ├── models.py                       # Can be generated from JSON Schemas
│   │       └── decorator.py                    # @capability() decorator for Python capabilities
│   │
│   ├── go/                                      # Go module: github.com/psiog/capafabric-go
│   │   ├── go.mod
│   │   └── client/
│   │       ├── agent.go                        # Agent interface + BaseAgent struct embedding
│   │       ├── agent_runner.go                 # AgentRunner
│   │       ├── agent_context.go                # AgentContext with CallChain
│   │       ├── client.go                       # ProxyClient
│   │       ├── cognitive.go                    # CognitiveProvider interface
│   │       └── models.go                       # Can be generated from JSON Schemas
│   │
│   └── java/                                    # Maven: com.psiog.capafabric (Java 23+)
│       ├── pom.xml
│       └── src/main/java/com/psiog/capafabric/
│           ├── Agent.java                      # Abstract Agent base class
│           ├── AgentRunner.java                # AgentRunner
│           ├── AgentContext.java               # AgentContext with callChain
│           ├── CapaFabricClient.java
│           ├── CognitiveProvider.java          # ICognitiveProvider interface
│           └── models/                          # Can be generated from JSON Schemas
│
├── examples/
│   │
│   ├── agent-dotnet/                       # .NET Semantic Kernel Agent
│   │   ├── Agent.csproj
│   │   ├── Program.cs
│   │   ├── Agents/
│   │   │   ├── AgentRunner.cs
│   │   │   ├── CapabilityBridge.cs
│   │   │   └── HITLGateway.cs
│   │   └── appsettings.Development.json
│   │
│   ├── agent-python/                       # Python LangGraph Agent
│   │   ├── pyproject.toml
│   │   └── agent.py
│   │
│   ├── capability-dotnet-oracle/               # .NET Oracle connector + manifest
│   │   ├── OracleConnector.csproj
│   │   ├── Program.cs
│   │   └── manifest.yaml
│   │
│   ├── capability-go-matching/                 # Go matching engine + manifest
│   │   ├── go.mod
│   │   ├── main.go
│   │   └── manifest.yaml
│   │
│   ├── capability-java-email/                  # Java email ingestion + manifest
│   │   ├── pom.xml
│   │   ├── src/main/java/.../EmailService.java
│   │   └── manifest.yaml
│   │
│   └── multi-agent-stallion/                   # Full 4-stage pipeline
│       ├── supervisor/                          # Supervisor Agent
│       ├── ingestion/                           # Ingestion Agent + manifest
│       ├── extraction/                          # Extraction Agent + manifest
│       ├── matching/                            # Matching Agent (Go) + manifest
│       ├── control/                             # Control Agent + manifest
│       └── docker-compose.yaml
│
├── config/
│   ├── litellm_config.yaml                     # Production LiteLLM config
│   ├── litellm_config.dev.yaml                 # Dev LiteLLM config (Ollama)
│   └── otel-collector-config.yaml              # OTEL Collector pipeline
│
├── scripts/
│   ├── dev-setup.sh                            # Install Ollama, pull model, setup Redis
│   ├── validate-manifests.sh                   # Validate all manifest.yaml files
│   └── generate-proto.sh                       # Generate Go code from .proto files
│
├── docker-compose.yaml                         # Full stack (Level 6)
├── docker-compose.dev.yaml                     # Dev services only
│
└── .github/
    └── workflows/
        ├── ci.yaml                             # Build + test + lint
        ├── release.yaml                        # Build binaries + Docker images
        └── validate-manifests.yaml             # CI manifest validation
```

---

## Control Plane Config (capafabric.yaml)

```yaml
# control-plane/configs/capafabric.yaml
server:
  port: 8080
  admin_ui: true

registry:
  type: redis                        # inmemory | redis | postgres
  url: redis://redis:6379

state:
  type: redis
  url: redis://state:6379

auth:
  provider: jwt                      # jwt | mtls | apikey | none
  jwks_url: https://login.microsoftonline.com/.../discovery/v2.0/keys
  audience: api://capafabric

policy:
  enforcer: rbac                     # rbac | opa | abac | none

guardrails:
  enabled: true
  fail_mode: block
  inbound:
    - type: prompt_injection
      provider: local
      action: block
    - type: pii_redaction
      provider: presidio
      action: redact
      phi_mode: true
  outbound:
    - type: output_safety
      provider: azure_content_safety
      action: block
    - type: schema_compliance
      action: block

context_management:
  enabled: true
  strategy: adaptive

observability:
  tracer: otel
  otel_endpoint: http://otel-collector:4317
  service_name: capafabric-control-plane
  audit:
    enabled: true
    writer: cosmos

load_balancing:
  strategy: round_robin
  health_check_interval_ms: 10000

health:
  heartbeat_ttl_seconds: 60
  check_interval_ms: 30000
```

---

## Control Plane Dev Config (capafabric.dev.yaml)

```yaml
# control-plane/configs/capafabric.dev.yaml
# Minimal — everything in-memory, no external dependencies
server:
  port: 8080
  admin_ui: true

registry:
  type: inmemory

state:
  type: inmemory

auth:
  provider: none

policy:
  enforcer: none

guardrails:
  enabled: false

context_management:
  enabled: false

observability:
  tracer: log
  audit:
    enabled: false

load_balancing:
  strategy: round_robin
```

---

## Proxy Config — Agent Mode

```yaml
# proxy/configs/proxy-agent.yaml
mode: agent
port: 3500

control_plane:
  url: http://localhost:8080
  cache_ttl_seconds: 30

llm:
  endpoint: http://localhost:4000/v1          # LiteLLM
  fallback_endpoint: http://localhost:11434/v1 # Ollama direct
  health_check_interval_ms: 10000

observability:
  tracer: otel
  otel_endpoint: http://localhost:4317
  service_name: cfproxy-agent
```

---

## Proxy Config — Capability Mode

```yaml
# proxy/configs/proxy-capability.yaml
mode: capability
port: 3501

control_plane:
  url: http://localhost:8080

app:
  port: 8081                                   # The application's port
  health_check_interval_ms: 10000

manifest: ./manifest.yaml                      # Path to capability manifest

observability:
  tracer: otel
  otel_endpoint: http://localhost:4317
  service_name: cfproxy-oracle-connector
```

---

## Makefile

```makefile
.PHONY: build test lint run-cp run-proxy-goal run-proxy-cap dev validate

# Build
build:
	cd control-plane && go build -o ../bin/capafabric ./cmd/capafabric
	cd proxy && go build -o ../bin/cfproxy ./cmd/cfproxy

# Test
test:
	cd shared && go test ./...
	cd control-plane && go test ./...
	cd proxy && go test ./...

# Lint
lint:
	cd shared && golangci-lint run
	cd control-plane && golangci-lint run
	cd proxy && golangci-lint run

# Run — Dev mode
run-cp:
	cd control-plane && go run ./cmd/capafabric --config=configs/capafabric.dev.yaml

run-proxy-goal:
	cd proxy && go run ./cmd/cfproxy --mode=agent --config=configs/proxy-agent.yaml

run-proxy-cap:
	cd proxy && go run ./cmd/cfproxy --mode=capability --config=configs/proxy-capability.yaml \
		--manifest=$(MANIFEST)

# Full dev stack
dev:
	@echo "Starting Ollama..."
	@ollama serve &
	@echo "Starting Control Plane..."
	@make run-cp &
	@echo "Starting Agent Proxy..."
	@make run-proxy-goal &
	@echo "Dev stack ready. CP: :8080, Proxy: :3500"

# Validate all manifests
validate:
	@find examples -name "manifest.yaml" -exec \
		go run ./shared/manifest/cmd/validate.go {} \;

# Docker
docker-build:
	docker build -t capafabric/control-plane:latest ./control-plane
	docker build -t capafabric/proxy:latest ./proxy

# Full stack
up:
	docker compose up -d

down:
	docker compose down
```

---

## go.work (Workspace)

```
go 1.23

use (
	./shared
	./control-plane
	./proxy
)
```

---

## Implementation Order

```
Week 1:  shared/models + shared/interfaces + shared/manifest
         control-plane/cmd + control-plane/internal/config
         control-plane/internal/registry/inmemory
         control-plane/internal/api (register, discover, invoke, health)
         → CP starts, accepts manifest registration, returns capabilities

Week 2:  proxy/cmd + proxy/internal/config
         proxy/internal/manifest (loader + endpoint_mapper)
         proxy/internal/registration (registrar + heartbeat)
         proxy/internal/api (discover, invoke, health)
         proxy/internal/transport/http
         → Proxy reads manifest, registers with CP, routes /invoke to app

Week 3:  sdk/dotnet (CapaFabric.Client NuGet)
         examples/agent-dotnet (AgentRunner + CapabilityBridge)
         examples/capability-dotnet-oracle (+ manifest)
         → .NET Agent discovers and invokes capabilities via proxy

Week 4:  control-plane/internal/policy (auth + enforcer)
         control-plane/internal/configdist (push to proxies)
         proxy/internal/cache (config cache)
         → Auth + policy enforced, config pushed to proxies

Week 5:  control-plane/internal/guardrails (engine + providers)
         proxy/internal/guardrails (invoker from cached rules)
         proxy/internal/pipeline (composable pipeline)
         proxy/internal/llm (forwarder with guardrail pipeline)
         → Guardrails active on /llm/chat and /invoke

Week 6:  control-plane/internal/context (strategies)
         proxy/internal/context (manager from cached config)
         proxy/internal/otel (tracer + metrics)
         control-plane/internal/audit
         → Context management + OTEL + audit trail

Week 7:  control-plane/internal/state (redis)
         control-plane/internal/health (monitor + heartbeat)
         control-plane/internal/loadbalancer
         proxy/internal/fallback (circuit breaker)
         → State, health, LB, resilience

Week 8:  control-plane/internal/ui (admin dashboard)
         examples/multi-agent-stallion (full 4-stage pipeline)
         docker-compose.yaml
         CI pipeline (.github/workflows)
         → Production-ready
```

---

## Component Swappability Matrix

```
Every component boundary is a network protocol, not a language binding.
No component imports another component's code. They communicate via
HTTP, gRPC, and JSON. Any component can be rewritten in any language
at any time — as long as it passes the contract tests.

Component       │ v1 (ship fast)  │ v2 (optimize)      │ Contract
────────────────┼─────────────────┼────────────────────┼──────────────────────────────
Control Plane   │ Go              │ Rust / Zig / Go    │ control-plane-api.openapi.yaml
Proxy           │ Go              │ Rust / C++         │ proxy-api.openapi.yaml
.NET SDK        │ C# (.NET 10)    │ C# / F#           │ proxy-api.openapi.yaml
TypeScript SDK  │ TypeScript      │ TypeScript / Bun   │ proxy-api.openapi.yaml
Python SDK      │ Python          │ Python / Rust+PyO3 │ proxy-api.openapi.yaml
Go SDK          │ Go              │ Go / Rust          │ proxy-api.openapi.yaml
Java SDK        │ Java 23         │ Java / Kotlin      │ proxy-api.openapi.yaml
Admin UI        │ React           │ Svelte/Vue/HTMX    │ admin-ui-api.openapi.yaml
CognitiveProvider│ Per-SDK impl   │ gRPC service       │ cognitive.proto

Validation: run contract-tests/ against any implementation.
If tests pass → valid drop-in replacement. Zero changes to other components.
```

---

## Code Generation from Contracts

```
The contracts/ directory enables automatic code generation for any language.
This eliminates hand-written model drift across SDKs.

Source                                  │ Tool                  │ Output
────────────────────────────────────────┼───────────────────────┼──────────────────────
contracts/openapi/proxy-api.openapi.yaml│ openapi-generator     │ SDK client stubs
                                        │ nswag (for .NET)      │   (any language)
                                        │ oapi-codegen (for Go) │
                                        │ openapi-ts (for TS)   │
────────────────────────────────────────┼───────────────────────┼──────────────────────
contracts/jsonschema/*.schema.json      │ quicktype             │ Model classes
                                        │                       │   .NET, TS, Python,
                                        │                       │   Go, Java, Rust
────────────────────────────────────────┼───────────────────────┼──────────────────────
contracts/grpc/*.proto                  │ protoc                │ gRPC stubs + models
                                        │ + grpc-go             │   Go, .NET, Python,
                                        │ + grpc-dotnet         │   Java, TS, Rust
                                        │ + grpc-java           │
────────────────────────────────────────┼───────────────────────┼──────────────────────

Makefile targets:
  make generate-models     → quicktype JSON Schemas → all SDK model files
  make generate-grpc       → protoc .proto → all SDK gRPC stubs
  make generate-clients    → openapi-generator → SDK client stubs
  make generate-all        → all of the above

CI pipeline validates: generated code matches committed code (no drift)
```

---

## Contract Governance Rules

```
1. CONTRACTS FIRST — Write the OpenAPI spec / JSON Schema / .proto BEFORE
   implementing. The contract is the design. The code is the implementation.

2. VERSION EVERYTHING — Every contract has an apiVersion field.
   Breaking changes require a major version bump (v1 → v2).
   Non-breaking additions (new optional fields) are minor bumps.

3. BACKWARD COMPATIBILITY — A v2 Control Plane must accept v1 proxy
   connections. A v2 proxy must work with a v1 Control Plane.
   Use optional fields and graceful degradation.

4. CONTRACT TESTS ARE CI GATES — No PR merges unless contract tests pass
   for every implementation in the repo.

5. GENERATED CODE IS COMMITTED — Generated model files are committed to
   the repo (not .gitignored). CI validates they match the source schemas.
   This ensures SDK users don't need code generation tools installed.

6. CONTRACTS LIVE IN contracts/ — Never in a language-specific directory.
   The contracts/ directory is the single source of truth that all
   implementations derive from.
```
