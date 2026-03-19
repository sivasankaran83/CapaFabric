# CapaFabric — Project Setup Guide

> Everything you need to set up CapaFabric on your laptop and start building.
> From zero to a running Agent with capabilities in under 30 minutes.

---

## Development Environment Requirements

### Minimum Hardware

```
Processor:  Intel i5 7th Gen or equivalent (4 cores recommended)
RAM:        16 GB DDR4
Storage:    250 GB (SSD preferred — Go and .NET builds benefit from fast I/O)
GPU:        Not required (Ollama runs on CPU for 3B models)
OS:         Windows 10, Windows 11, macOS, Linux
```

### RAM Budget by Development Level

```
With local Ollama (llama3.2 3B):

Level 0 — Agent + Ollama:                ~5 GB
  Ollama (llama3.2 3B):    ~2.5 GB
  .NET Agent:          ~0.3 GB
  VS Code:                 ~1.5 GB
  Windows OS:              ~2.0 GB
  Headroom:                ~9.7 GB    ✅ Comfortable

With OpenRouter (no local model — recommended for 16GB machines):

Level 0 — Agent + OpenRouter:            ~4 GB
  .NET Agent:          ~0.3 GB
  VS Code:                 ~1.5 GB
  Windows OS:              ~2.0 GB
  Headroom:                ~12 GB     ✅ Excellent (no Ollama RAM cost)

Level 1 — + Proxy + CP + Capability:         ~5 GB
  + Go Control Plane:      ~0.1 GB
  + Go Proxy × 2:          ~0.2 GB
  + .NET Capability:       ~0.3 GB
  Headroom:                ~11 GB     ✅ Excellent

Level 2 — + LiteLLM:                         ~6 GB
  + LiteLLM Proxy:         ~0.5 GB
  Headroom:                ~10 GB     ✅ Comfortable

Level 3 — + OTEL + Jaeger:                   ~7 GB
  + OTEL Collector:        ~0.3 GB
  + Jaeger:                ~0.5 GB
  Headroom:                ~9 GB      ✅ Comfortable

Level 4 — + State Store:                     ~7.5 GB
  + SQLite (embedded):     ~0 GB      ← no separate process
  OR
  + Redis:                 ~0.1 GB    ← separate process
  Headroom:                ~8.5 GB    ✅ Comfortable

Level 5 — Full governance (native):          ~8 GB
  All of the above
  Headroom:                ~8 GB      ✅ Fine

Level 6 — Docker Compose:                    ~12 GB
  Docker Desktop:          ~2.0 GB
  All containers:          ~4.0 GB
  Headroom:                ~4 GB      ⚠️ Tight but possible with OpenRouter

RECOMMENDATION for 16GB / i5-7th-Gen:
  Use OpenRouter instead of local Ollama — saves 2.5GB RAM
  Use SQLite instead of Redis for state — saves one process
  Daily development:  Level 0-4 natively (comfortable at ~7-8GB)
  Integration test:   Level 5 natively (fine at ~8GB)
  Full stack Docker:  Level 6 only if using OpenRouter (no Ollama container)
```

### State Store Options (Level 4+)

```
state.type: inmemory   — fastest, no persistence, zero overhead
                         Use for: Level 0-3 development

state.type: sqlite     — durable, no separate process, ~0 RAM overhead
                         Use for: Level 4-5 on memory-constrained machines
                         Single file: ./capafabric-state.db

state.type: redis      — durable, requires separate process (~100MB RAM)
                         Use for: production, multi-instance deployments

state.type: postgres   — durable, requires separate process (~200MB RAM)
                         Use for: production, when Postgres already in stack

state.type: cosmos     — Azure Cosmos DB (cloud)
                         Use for: Azure production deployments

All implement the same StateStore interface. Swap via YAML config.
No code changes. No Agent changes.
```

### Windows 10 Specific Notes

```
- Use Windows Terminal (not cmd.exe) for better experience
- Ollama runs natively on Windows (no WSL needed)
- Go and .NET run natively (no WSL needed)
- LiteLLM: install via pip in a Python venv to avoid conflicts
    python -m venv .venv
    .venv\Scripts\activate
    pip install litellm
- If using Docker: Docker Desktop requires WSL2 + Hyper-V (uses 2-4GB RAM)
  → Prefer running everything natively (no Docker) on 16GB machines
- Long path support: enable in Windows if you hit path length errors
    git config --global core.longpaths true
- Line endings: configure Git for Windows
    git config --global core.autocrlf true
```

### Model Selection for 16GB RAM

```
RECOMMENDED (stays under 3GB):
  ollama pull llama3.2         # 2.0 GB — good general reasoning
  ollama pull phi3:mini        # 1.7 GB — faster, slightly less capable

AVOID on 16GB (too large):
  ollama pull llama3.1:8b      # 4.7 GB — will starve other processes
  ollama pull codellama:13b    # 7.4 GB — won't leave room for anything else

TIP: Set Ollama to release memory when idle:
  set OLLAMA_KEEP_ALIVE=5m     # Windows: release model after 5 min idle
```

---

## Prerequisites

### Required

```bash
# .NET 10 SDK (LTS — target latest, port to lower versions on need basis)
# Download: https://dotnet.microsoft.com/download/dotnet/10.0
dotnet --version   # Should show 10.0.x or higher
# Portability: The codebase targets net10.0. If a client requires .NET 8 or .NET 6,
# update <TargetFramework> in .csproj files. Avoid .NET 10-only APIs (file-based apps,
# C# 14 extension members) unless guarded by #if NET10_0_OR_GREATER.

# Go 1.23+
# Download: https://go.dev/dl/
go version         # Should show go1.23.x or higher

# OpenJDK 23+ (target latest, port to lower versions on need basis)
# macOS:
brew install openjdk
# Windows:
# Download from https://adoptium.net/temurin/releases/
# Linux:
sudo apt install openjdk-23-jdk
java --version     # Should show 23.x or higher
# Portability: The codebase targets Java 23. If a client requires Java 17 or Java 11,
# update <maven.compiler.release> in pom.xml. Avoid Java 23-only APIs (unnamed patterns,
# string templates) unless guarded. Java 17 is the minimum viable port target.

# Maven (Java build tool)
# macOS:
brew install maven
# Windows/Linux:
# Download from https://maven.apache.org/download.cgi
mvn --version

# Ollama (local LLM)
# macOS:
brew install ollama
# Windows:
# Download installer from https://ollama.com/download
# Linux:
curl -fsSL https://ollama.com/install.sh | sh
```

### Optional (add as you progress through levels)

```bash
# LiteLLM (Level 2+)
pip install litellm

# Redis (Level 5+)
# macOS:
brew install redis
# Windows: Use Docker or Memurai (https://www.memurai.com/)
# Linux:
sudo apt install redis-server

# Jaeger (Level 4+)
# Download binary from: https://www.jaegertracing.io/download/
# Or run via Docker: docker run -d -p 16686:16686 jaegertracing/all-in-one

# Node.js 22+ (ONLY needed to BUILD the Admin UI)
# macOS:
brew install node
# Windows/Linux: https://nodejs.org/
# The Admin UI source is always in the repo (control-plane/ui/).
# Node.js is only needed when you want to build and package the UI.
# The CP binary can be built WITH or WITHOUT the UI (see Makefile targets).

# Docker (Level 6)
# Download: https://www.docker.com/products/docker-desktop/
```

### VS Code Extensions

```bash
# Install these extensions:
code --install-extension ms-dotnettools.csdevkit
code --install-extension ms-dotnettools.csharp
code --install-extension golang.go
code --install-extension redhat.vscode-yaml
code --install-extension ms-azuretools.vscode-docker
code --install-extension humao.rest-client
```

---

## Step 1: Create Repository

```bash
mkdir capafabric && cd capafabric
git init
```

---

## Step 2: Create Directory Structure

```bash
# Shared module
mkdir -p shared/{models,interfaces,manifest,otel,errors}

# Control Plane
mkdir -p control-plane/cmd/capafabric
mkdir -p control-plane/internal/{config,registry,discovery,policy/auth,policy/enforcer}
mkdir -p control-plane/internal/{guardrails,context,health,loadbalancer,state,audit}
mkdir -p control-plane/internal/{configdist,api/handlers,api/dto,ui/dist}
mkdir -p control-plane/ui/src/{pages,components,api}
mkdir -p control-plane/configs

# Proxy
mkdir -p proxy/cmd/cfproxy
mkdir -p proxy/internal/{config,cache,pipeline,transport,manifest,registration}
mkdir -p proxy/internal/{guardrails,context,otel,llm,fallback,api/handlers}
mkdir -p proxy/configs

# SDK
mkdir -p sdk/dotnet/src/CapaFabric.Client/{Models,Providers}
mkdir -p sdk/dotnet/tests/CapaFabric.Client.Tests
mkdir -p sdk/typescript/src/{models,providers,utils}
mkdir -p sdk/typescript/tests
mkdir -p sdk/python/capafabric
mkdir -p sdk/go/client
mkdir -p sdk/java/src/main/java/com/psiog/capafabric/models

# Proto / Contracts
mkdir -p contracts/{openapi,grpc,jsonschema}
mkdir -p contract-tests/{cp-api,proxy-api,manifest/sample_manifests}

# Config
mkdir -p config

# Examples
mkdir -p examples/agent-dotnet/{Core,Agents,Infrastructure,Configuration,Endpoints}
mkdir -p examples/capability-dotnet-oracle
mkdir -p examples/capability-go-matching
mkdir -p examples/multi-agent-stallion/{supervisor,ingestion,extraction,matching,control}

# Scripts
mkdir -p scripts

# CI
mkdir -p .github/workflows

# Docs + ADRs
mkdir -p docs/adr
```

---

## Step 2b: Create Initial ADRs

```bash
# ADR template
cat > docs/adr/000-template.md << 'ADRTPL'
# ADR-{number}: {title}

**Status**: Proposed | Accepted | Deprecated | Superseded by ADR-{n}
**Date**: {YYYY-MM-DD}
**Decision Makers**: {names}

## Context
What is the issue that we're seeing that is motivating this decision?

## Decision
What is the change that we're proposing and/or doing?

## Consequences
What becomes easier or more difficult to do because of this change?
ADRTPL

# ADR-001: Capability + Goal vocabulary
cat > docs/adr/001-capability-goal-vocabulary.md << 'EOF'
# ADR-001: Capability + Goal Vocabulary

**Status**: Accepted
**Date**: 2026-03-19

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
EOF

# ADR-002: Control Plane + Thin Proxy architecture
cat > docs/adr/002-control-plane-thin-proxy.md << 'EOF'
# ADR-002: Control Plane + Thin Proxy Architecture

**Status**: Accepted
**Date**: 2026-03-19

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
EOF

# ADR-003: Language-neutral contracts for component swappability
cat > docs/adr/003-language-neutral-contracts.md << 'EOF'
# ADR-003: Language-Neutral Contracts for Component Swappability

**Status**: Accepted
**Date**: 2026-03-19

## Context
Components (CP, Proxy, SDKs, Admin UI) must be rewritable in any language without
affecting other components. Initial implementation is Go (CP, Proxy) but future
versions may use Rust (proxy), or other languages.

## Decision
All component boundaries are defined by language-neutral contracts in contracts/:
- OpenAPI 3.1 specs for REST APIs (CP, Proxy, Admin UI)
- Protocol Buffer definitions for gRPC services
- JSON Schema for all value objects and manifests

Contract tests (contract-tests/) validate any implementation against these specs.
SDK models are auto-generated from JSON Schemas via quicktype.

## Consequences
- Any component can be rewritten without touching others
- Contract tests are CI gates — no merge without passing
- Generated code eliminates model drift across SDKs
- Added overhead: maintaining contracts as source of truth
EOF

# ADR-004: CognitiveProvider as polymorphic LLM abstraction
cat > docs/adr/004-cognitive-provider.md << 'EOF'
# ADR-004: CognitiveProvider as Polymorphic LLM Abstraction

**Status**: Accepted
**Date**: 2026-03-19

## Context
The Agent needs to call LLMs, but the LLM provider varies per agent (Supervisor
uses Claude, Ingestion uses Haiku, Matching uses local Ollama) and per invocation
within the same agent. Coupling to LiteLLM or any specific provider is unacceptable.

## Decision
Introduce ICognitiveProvider with polymorphic Deliberate() method:
- Deliberate(request) → simple request/response
- Deliberate(request, toolInvoker) → with tool calling (delegate callback pattern)
- DeliberateStream(request) → streaming tokens
- DeliberateStream(request, toolInvoker) → streaming + tool calling

CognitiveConfig is provider-specific (LiteLLMConfig, OllamaConfig, SemanticKernelConfig).
Injected per-agent via DI. Can vary per invocation.

## Consequences
- Agent is completely decoupled from any LLM provider
- Per-agent model selection enables cost optimization (60-70% cost reduction)
- toolInvoker delegate keeps CognitiveProvider decoupled from CapaFabric invocation
- 4 overloads cover every agentic interaction pattern
EOF

# ADR-005: YAML manifest as primary integration contract
cat > docs/adr/005-yaml-manifest-primary-contract.md << 'EOF'
# ADR-005: YAML Manifest as Primary Integration Contract

**Status**: Accepted
**Date**: 2026-03-19

## Context
Enterprise applications have existing APIs. Requiring them to adopt an SDK, install
a framework, or modify their code creates adoption friction. Auto-discovery from
OpenAPI specs is unreliable in production (specs often missing or outdated).

## Decision
The YAML CapabilityManifest is the primary and only required integration artifact.
It maps existing API endpoints to discoverable capabilities without any application
code changes. The manifest includes: endpoint mapping (path/query/body/header),
response mapping (JSON path), security config, and LLM-optimized descriptions.

Auto-discovery from OpenAPI is a future convenience layer on top, not a replacement.

## Consequences
- L0 adoption: write YAML + add proxy = AI-discoverable (zero code changes)
- Explicit over implicit: manifests are version-controlled, PR-reviewable
- Manifest drift risk: must validate manifests against live apps in CI
- Description quality matters: descriptions are the LLM's discovery interface
EOF

# ADR-006: Zero hardcoded workflow logic
cat > docs/adr/006-zero-hardcoded-workflow.md << 'EOF'
# ADR-006: Zero Hardcoded Workflow Logic

**Status**: Accepted
**Date**: 2026-03-19

## Context
Traditional orchestration systems (Temporal, Airflow, Step Functions) require
predefined workflow DAGs. When business processes change, developers must modify
the DAG definition, test, and deploy.

## Decision
CapaFabric has zero workflow definitions anywhere in the system. No DAGs, no BPMN,
no step definitions, no if/else routing rules. Invocation paths emerge from the
Agent's LLM reasoning over: the goal, the available capability descriptions,
and the current context. Different goals produce different paths through the same
capabilities.

Agent system prompts describe GOAL INTENT only — never capability names or
sequences. Constraint rules ("never post without confirmation") serve as guardrails
on reasoning, not as workflow steps.

## Consequences
- Adding a new capability = deploy service + manifest. Zero Agent changes.
- Adding a new process step = deploy capability with descriptive manifest. LLM discovers it.
- Quality depends on capability description quality (LLM must understand when to use it)
- Less predictable than hardcoded workflows (trade-off: flexibility vs determinism)
- HITL gating and confidence thresholds provide safety net
EOF

# ADR-007: Governance layers as opt-in YAML configuration
cat > docs/adr/007-governance-opt-in-yaml.md << 'EOF'
# ADR-007: Governance Layers as Opt-In YAML Configuration

**Status**: Accepted
**Date**: 2026-03-19

## Context
Enterprise deployments need auth, guardrails, tracing, audit, context management,
and load balancing. Development environments need none of these. Embedding governance
in code creates overhead everywhere.

## Decision
Every cross-cutting concern follows the same pattern:
- Off by default
- YAML-activated in CP config
- Abstract interface with injectable implementation
- Zero code change in Agent or capability when enabled/changed

Auth, policy, guardrails, tracing, audit, context management, load balancing —
all configured per-environment, per-tenant, per-capability via YAML.

## Consequences
- Dev environment: zero governance overhead
- Staging: guardrails in log_only mode for confidence building
- Production: full governance stack
- PHI deployment: add Presidio redaction + forensic audit
- Same binary, same code — different YAML
EOF

# ADR-008: Agents as capabilities (multi-agent pattern)
cat > docs/adr/008-agents-as-capabilities.md << 'EOF'
# ADR-008: Agents as Capabilities (Multi-Agent Pattern)

**Status**: Accepted
**Date**: 2026-03-19

## Context
Multi-agent orchestration traditionally requires a special agent-to-agent protocol
(A2A, AMCP). This adds complexity and a new protocol to maintain.

## Decision
An agent's goal endpoint is registered in the manifest as a capability — identical
to any deterministic capability. The Supervisor Agent discovers sub-agents via
the same /discover call used for deterministic capabilities. No special multi-agent
protocol, no agent-to-agent messaging framework.

Each sub-agent encapsulates its own Agent with its own CognitiveProvider and
internal capabilities. To the Supervisor, calling a sub-agent is indistinguishable
from calling a deterministic capability.

## Consequences
- Multi-agent orchestration = capability discovery + invocation (no new concepts)
- Sub-agents are independently deployable, scalable, and language-independent
- Supervisor prompt is goal-only — doesn't name sub-agents or prescribe order
- Adding a new agent = deploy + manifest. Supervisor discovers it automatically.
EOF

# ADR-009: Agent base class with polymorphic Deliberate
cat > docs/adr/009-agent-base-class.md << 'EOF'
# ADR-009: Agent Base Class with Polymorphic Deliberate

**Status**: Accepted
**Date**: 2026-03-19

## Context
Every Agent in the system follows the same pattern: discover capabilities,
deliberate via CognitiveProvider, invoke capabilities, evaluate confidence,
escalate to HITL if needed. Without a shared base class, each Agent reimplements
this logic, leading to inconsistency and bugs.

## Decision
Introduce an abstract Agent base class (in every SDK language) that implements
the universal pattern:
  1. Discover capabilities relevant to the goal
  2. If none found → OnNoCapabilitiesFound (default: HITL escalation)
  3. Deliberate via CognitiveProvider with tool invoker callback
  4. Evaluate result: complete, continue, low confidence, or unparsable
  5. Low confidence → OnLowConfidence (default: HITL escalation)

Concrete Agents extend the base and provide:
  - AgentId (identity)
  - Persona (system prompt — goal intent only)
  - Override points: OnNoCapabilitiesFound, OnLowConfidence, GetCognitiveConfig,
    GetConfidenceThreshold

An AgentRunner wraps the iteration loop (max iterations, token budget,
checkpointing) and calls Agent.PursueGoalAsync repeatedly until completion.

The base class is implemented in all SDK languages: C#, TypeScript, Python,
Go, Java, Rust — each using the language's native abstraction mechanism
(abstract class, ABC, trait, interface + struct embedding).

## Consequences
- Creating a new Agent = extend base, write Persona, optionally override thresholds
- Universal pattern enforced: all Agents discover → deliberate → invoke → evaluate
- Override points allow per-Agent customization without changing the base
- CognitiveProvider is per-Agent (Supervisor uses Sonnet, Ingestion uses Haiku)
- Polyglot: same pattern works identically in 6 languages
EOF

# ADR-010: Agent loop protection
cat > docs/adr/010-agent-loop-protection.md << 'EOF'
# ADR-010: Agent Loop Protection (Circular Chain, Max Depth, Retry Detection)

**Status**: Accepted
**Date**: 2026-03-19

## Context
In a multi-agent system where Agents discover and invoke Capabilities (which
may themselves be Agents), three types of infinite loops can occur:
  1. Agent calls itself (circular self-reference)
  2. Agent A → Agent B → Agent C → Agent A (circular chain)
  3. Agent retries the same failing capability with identical arguments

## Decision
Three guards implemented at TWO levels (belt and suspenders):

**SDK Level (Agent base class):**
  Guard 1 — Circular chain: AgentContext carries a CallChain (list of agent_ids).
    Before executing, Agent checks if its own ID is already in the chain.
    If yes → return AgentResult.Failed with chain trace.

  Guard 2 — Max depth: CallChain.Count checked against configurable MaxDepth
    (default: 10). If exceeded → escalate to HITL.

  Guard 3 — Retry loop: InvocationLog tracks (capability_id, args_hash) for
    last N calls. If same capability + same args repeated N times consecutively
    (default: 3) → inject structured error into tool result so LLM can
    self-correct. LLM receives: "You've tried this 3 times with identical
    input. Try a different approach or escalate."

**Infrastructure Level (Proxy):**
  Call chain propagated via HTTP headers:
    X-CapaFabric-Call-Chain: supervisor,ingestion,extraction
    X-CapaFabric-Depth: 3
    X-CapaFabric-Max-Depth: 10

  Proxy validates before routing:
    - Target agent already in chain → 409 Conflict
    - Depth >= max_depth → 429 Too Deep

  This catches cross-service loops that the SDK can't see (Agent A in .NET
  calls Agent B in Go which calls Agent A — the .NET process doesn't know
  about the Go call, but the proxy headers carry the full chain).

## Consequences
- Three loop types covered at two enforcement levels
- CallChain propagates across language boundaries via HTTP headers
- All guards produce structured errors the LLM can reason about
- All guards escalate to HITL by default (configurable per Agent)
- Guards are defined in contracts/openapi/proxy-api.openapi.yaml
- MaxDepth and MaxConsecutiveRetries configurable per Agent and per tenant
- ADR-009 Agent base class embeds all three guards in PursueGoalAsync
EOF
```

---

## Step 3: Initialize Go Modules

```bash
# Shared module
cd shared
cat > go.mod << 'EOF'
module github.com/psiog/capafabric/shared

go 1.23
EOF
cd ..

# Control Plane module
cd control-plane
cat > go.mod << 'EOF'
module github.com/psiog/capafabric/control-plane

go 1.23

require github.com/psiog/capafabric/shared v0.0.0
EOF
cd ..

# Proxy module
cd proxy
cat > go.mod << 'EOF'
module github.com/psiog/capafabric/proxy

go 1.23

require github.com/psiog/capafabric/shared v0.0.0
EOF
cd ..

# Go workspace
cat > go.work << 'EOF'
go 1.23

use (
	./shared
	./control-plane
	./proxy
)
EOF
```

---

## Step 4: Initialize .NET Projects

```bash
# Solution
dotnet new sln -n CapaFabric

# SDK client library
cd sdk/dotnet/src/CapaFabric.Client
dotnet new classlib -n CapaFabric.Client --framework net10.0
cd ../../../..
dotnet sln add sdk/dotnet/src/CapaFabric.Client/CapaFabric.Client.csproj

# SDK test project
cd sdk/dotnet/tests/CapaFabric.Client.Tests
dotnet new xunit -n CapaFabric.Client.Tests --framework net10.0
dotnet add reference ../../src/CapaFabric.Client/CapaFabric.Client.csproj
cd ../../../..
dotnet sln add sdk/dotnet/tests/CapaFabric.Client.Tests/CapaFabric.Client.Tests.csproj

# Example Agent
cd examples/agent-dotnet
dotnet new web -n Agent --framework net10.0
dotnet add package Microsoft.SemanticKernel --version "1.*"
dotnet add package Microsoft.SemanticKernel.Connectors.OpenAI --version "1.*"
dotnet add package Microsoft.Extensions.Http --version "8.*"
dotnet add package OpenTelemetry.Extensions.Hosting --version "1.*"
dotnet add package OpenTelemetry.Exporter.OpenTelemetryProtocol --version "1.*"
dotnet add reference ../../sdk/dotnet/src/CapaFabric.Client/CapaFabric.Client.csproj
cd ../..
dotnet sln add examples/agent-dotnet/Agent.csproj

# Example Oracle Connector capability
cd examples/capability-dotnet-oracle
dotnet new web -n OracleConnector --framework net10.0
cd ../..
dotnet sln add examples/capability-dotnet-oracle/OracleConnector.csproj
```

---

## Step 5: Pull Ollama Model (Optional — skip if using OpenRouter)

```bash
# Only needed if you want local LLM inference.
# Skip this step if using OpenRouter — saves 2.5GB RAM.
ollama pull llama3.2
# ~2GB download. Takes 2-5 minutes.
# For faster iteration: ollama pull phi3:mini (~1.7GB)
```

---

## Step 6: Create Config Files

### LiteLLM Dev Config

```bash
cat > config/litellm_config.dev.yaml << 'EOF'
# Option A: OpenRouter (recommended for 16GB machines — no local model needed)
model_list:
  - model_name: "reasoning-heavy"
    litellm_params:
      model: openrouter/anthropic/claude-sonnet-4
      api_key: os.environ/OPENROUTER_API_KEY

  - model_name: "fast-classification"
    litellm_params:
      model: openrouter/anthropic/claude-haiku-4
      api_key: os.environ/OPENROUTER_API_KEY

# Option B: Local Ollama (if you have RAM to spare)
#  - model_name: "reasoning-heavy"
#    litellm_params:
#      model: ollama/llama3.2
#      api_base: http://localhost:11434
#
#  - model_name: "fast-classification"
#    litellm_params:
#      model: ollama/llama3.2
#      api_base: http://localhost:11434

general_settings:
  master_key: sk-dev-key
EOF
```

### Control Plane Dev Config

```bash
cat > control-plane/configs/capafabric.dev.yaml << 'EOF'
server:
  port: 8080
  admin_ui: false

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
EOF
```

### Proxy Agent Mode Config

```bash
cat > proxy/configs/proxy-agent.yaml << 'EOF'
mode: agent
port: 3500

control_plane:
  url: http://localhost:8080
  cache_ttl_seconds: 30

llm:
  endpoint: http://localhost:4000/v1
  fallback_endpoint: http://localhost:11434/v1
  health_check_interval_ms: 10000

observability:
  tracer: log
  service_name: cfproxy-agent
EOF
```

### Proxy Capability Mode Config

```bash
cat > proxy/configs/proxy-capability.yaml << 'EOF'
mode: capability
port: 3501

control_plane:
  url: http://localhost:8080

app:
  port: 8081
  health_check_interval_ms: 10000

observability:
  tracer: log
  service_name: cfproxy-capability
EOF
```

### Agent Dev Settings

```bash
# Option A: OpenRouter (recommended for 16GB machines)
cat > examples/agent-dotnet/appsettings.Development.json << 'EOF'
{
  "Logging": {
    "LogLevel": {
      "Default": "Information"
    }
  },
  "LLM": {
    "ModelId": "anthropic/claude-sonnet-4",
    "Endpoint": "https://openrouter.ai/api/v1",
    "ApiKey": "${OPENROUTER_API_KEY}"
  },
  "Agent": {
    "UseProxy": false,
    "ProxyUrl": "http://localhost:3500",
    "DefaultConfidenceThreshold": 0.85,
    "MaxIterations": 10,
    "TokenBudget": 50000,
    "EnableOTel": false
  }
}
EOF

# Option B: Local Ollama
# Change LLM section to:
#   "ModelId": "llama3.2",
#   "Endpoint": "http://localhost:11434/v1",
#   "ApiKey": "ollama"
```

### Agent Production Settings

```bash
cat > examples/agent-dotnet/appsettings.Production.json << 'EOF'
{
  "LLM": {
    "ModelId": "reasoning-heavy",
    "Endpoint": "http://localhost:3500/llm/v1",
    "ApiKey": "${LITELLM_API_KEY}"
  },
  "Agent": {
    "UseProxy": true,
    "ProxyUrl": "http://localhost:3500",
    "ControlPlaneUrl": "http://capafabric-cp:8080",
    "DefaultConfidenceThreshold": 0.85,
    "MaxIterations": 10,
    "TokenBudget": 50000,
    "EnableOTel": true,
    "OTelEndpoint": "http://otel-collector:4317"
  }
}
EOF
```

---

## Step 7: Create Example Capability Manifest

```bash
cat > examples/capability-dotnet-oracle/manifest.yaml << 'EOF'
apiVersion: capafabric/v1
kind: CapabilityManifest

metadata:
  agent_id: stallion-cash-matching
  language: dotnet
  version: "1.0.0"

app:
  port: 8081
  protocol: http
  health_path: /health
  base_path: /api

capabilities:
  - capability_id: stallion.retrieve_invoice_details
    name: retrieve_invoice_details
    description: >
      Retrieves invoice details from Oracle Accounts Receivable including
      line items, customer info, amount, due date, and payment status.
    tags: [oracle, erp, invoice, read, finance]
    idempotent: true
    side_effects: false
    endpoint:
      method: GET
      path: /invoices/{invoice_id}
      arguments:
        invoice_id: { in: path }
      response:
        from: body
    security:
      required_roles: [finance_analyst, ar_clerk]
      classification: confidential

  - capability_id: stallion.locate_matching_receipts
    name: locate_matching_receipts
    description: >
      Searches bank lockbox cash receipts by customer ID and date range.
      Returns matching receipts with amounts and remittance info.
    tags: [bank, receipt, cash, read, finance]
    idempotent: true
    side_effects: false
    endpoint:
      method: POST
      path: /receipts/search
      arguments:
        customer_id: { in: body }
        date_range_start: { in: body }
        date_range_end: { in: body }
      response:
        from: body.receipts

  - capability_id: stallion.verify_three_way_match
    name: verify_three_way_match
    description: >
      Performs deterministic 3-way match between invoice amount, bank receipt
      amount, and purchase order amount. Returns match status and variance.
    tags: [matching, reconciliation, finance]
    idempotent: true
    side_effects: false
    endpoint:
      method: POST
      path: /matching/three-way
      arguments:
        invoice_amount: { in: body }
        receipt_amount: { in: body }
        po_amount: { in: body }
        tolerance: { in: body, default: 15.0 }
      response:
        from: body

  - capability_id: stallion.commit_financial_posting
    name: commit_financial_posting
    description: >
      Posts a matched receipt against an invoice in Oracle AR. Irreversible.
    tags: [oracle, erp, posting, write, finance, critical]
    idempotent: false
    side_effects: true
    requires_approval: true
    endpoint:
      method: POST
      path: /oracle/post-match
      arguments:
        invoice_id: { in: body }
        receipt_id: { in: body }
        match_status: { in: body }
      response:
        from: body
    security:
      required_roles: [ar_manager, finance_admin]
      classification: restricted
      audit_level: forensic
      max_calls_per_minute: 10

  - capability_id: stallion.escalate_variance_to_reviewer
    name: escalate_variance_to_reviewer
    description: >
      Sends a variance notification to the AR clerk for manual review.
    tags: [notification, alert, finance]
    idempotent: true
    side_effects: true
    endpoint:
      method: POST
      path: /notifications/variance-alert
      arguments:
        invoice_id: { in: body }
        customer_name: { in: body }
        variance_amount: { in: body }
        assigned_to: { in: body, default: "ar_clerk" }
      response:
        from: body

  - capability_id: stallion.produce_compliance_audit
    name: produce_compliance_audit
    description: >
      Generates a PDF audit document recording the complete matching journey.
    tags: [audit, pdf, compliance, finance]
    idempotent: true
    side_effects: true
    endpoint:
      method: POST
      path: /audit/generate-trail
      arguments:
        invoice_id: { in: body }
        decision_count: { in: body }
      response:
        from: body
EOF
```

---

## Step 8: Create Makefile

```bash
cat > Makefile << 'MAKEFILE'
.PHONY: build build-with-ui build-without-ui test lint run-cp run-proxy-goal run-proxy-cap dev clean setup

# ── Setup ──
setup:
	@echo "Installing Ollama model..."
	ollama pull llama3.2
	@echo "Restoring .NET packages..."
	dotnet restore
	@echo "Setup complete."

# ── Build ──

# Build WITHOUT Admin UI (no Node.js required)
build: build-without-ui

build-without-ui:
	cd control-plane && go build -tags no_ui -o ../bin/capafabric ./cmd/capafabric
	cd proxy && go build -o ../bin/cfproxy ./cmd/cfproxy
	dotnet build

# Build WITH Admin UI (requires Node.js 22+)
build-with-ui:
	cd control-plane/ui && npm install && npm run build
	cd control-plane && go build -o ../bin/capafabric ./cmd/capafabric
	cd proxy && go build -o ../bin/cfproxy ./cmd/cfproxy
	dotnet build

# ── Test ──
test:
	cd shared && go test ./...
	cd control-plane && go test ./...
	cd proxy && go test ./...
	dotnet test

# ── Lint ──
lint:
	cd shared && go vet ./...
	cd control-plane && go vet ./...
	cd proxy && go vet ./...

# ── Run Components ──
run-ollama:
	ollama serve

run-cp:
	cd control-plane && go run ./cmd/capafabric --config=configs/capafabric.dev.yaml

run-proxy-goal:
	cd proxy && go run ./cmd/cfproxy --mode=agent --config=configs/proxy-agent.yaml

run-proxy-cap:
	cd proxy && go run ./cmd/cfproxy --mode=capability --config=configs/proxy-capability.yaml \
		--manifest=$(MANIFEST)

run-litellm:
	litellm --config config/litellm_config.dev.yaml --port 4000

run-agent:
	cd examples/agent-dotnet && dotnet run

run-oracle-capability:
	cd examples/capability-dotnet-oracle && dotnet run --urls=http://localhost:8081

# ── Development Levels ──
level0:  ## Agent + Ollama (2 processes)
	@echo "Level 0: Start ollama in Terminal 1: make run-ollama"
	@echo "Level 0: Start Agent in Terminal 2: make run-agent"

level1:  ## + Proxy + Capability (4 processes)
	@echo "Level 1: Terminal 1: make run-ollama"
	@echo "Level 1: Terminal 2: make run-cp"
	@echo "Level 1: Terminal 3: make run-oracle-capability"
	@echo "Level 1: Terminal 4: make run-proxy-cap MANIFEST=examples/capability-dotnet-oracle/manifest.yaml"
	@echo "Level 1: Terminal 5: make run-proxy-goal"
	@echo "Level 1: Terminal 6: make run-agent (with UseProxy=true)"

level2:  ## + LiteLLM (5 processes)
	@echo "Level 2: Add to Level 1: make run-litellm"

# ── Validate ──
validate:
	@echo "Validating manifests..."
	@find examples -name "manifest.yaml" -exec echo "  Validating: {}" \;

# ── Clean ──
clean:
	rm -rf bin/
	dotnet clean

# ── Docker ──

# Docker WITHOUT Admin UI
docker-build:
	docker build -f control-plane/Dockerfile.slim -t capafabric/control-plane:latest-slim ./control-plane
	docker build -t capafabric/proxy:latest ./proxy

# Docker WITH Admin UI
docker-build-full:
	docker build -t capafabric/control-plane:latest ./control-plane
	docker build -t capafabric/proxy:latest ./proxy

up:
	docker compose up -d

down:
	docker compose down
MAKEFILE
```

---

## Step 9: Create VS Code Configuration

```bash
mkdir -p .vscode

cat > .vscode/launch.json << 'EOF'
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": ".NET Agent",
      "type": "coreclr",
      "request": "launch",
      "program": "${workspaceFolder}/examples/agent-dotnet/bin/Debug/net10.0/Agent.dll",
      "cwd": "${workspaceFolder}/examples/agent-dotnet",
      "env": {
        "ASPNETCORE_ENVIRONMENT": "Development",
        "ASPNETCORE_URLS": "http://localhost:5000"
      },
      "preLaunchTask": "build-agent"
    },
    {
      "name": "Oracle Capability",
      "type": "coreclr",
      "request": "launch",
      "program": "${workspaceFolder}/examples/capability-dotnet-oracle/bin/Debug/net10.0/OracleConnector.dll",
      "cwd": "${workspaceFolder}/examples/capability-dotnet-oracle",
      "env": {
        "ASPNETCORE_URLS": "http://localhost:8081"
      },
      "preLaunchTask": "build-oracle"
    },
    {
      "name": "Go Control Plane",
      "type": "go",
      "request": "launch",
      "program": "${workspaceFolder}/control-plane/cmd/capafabric",
      "args": ["--config=configs/capafabric.dev.yaml"],
      "cwd": "${workspaceFolder}/control-plane"
    },
    {
      "name": "Go Proxy (Agent)",
      "type": "go",
      "request": "launch",
      "program": "${workspaceFolder}/proxy/cmd/cfproxy",
      "args": ["--mode=agent", "--config=configs/proxy-agent.yaml"],
      "cwd": "${workspaceFolder}/proxy"
    },
    {
      "name": "Go Proxy (Capability)",
      "type": "go",
      "request": "launch",
      "program": "${workspaceFolder}/proxy/cmd/cfproxy",
      "args": [
        "--mode=capability",
        "--config=configs/proxy-capability.yaml",
        "--manifest=../../examples/capability-dotnet-oracle/manifest.yaml"
      ],
      "cwd": "${workspaceFolder}/proxy"
    }
  ],
  "compounds": [
    {
      "name": "Level 0: Agent Only",
      "configurations": [".NET Agent"]
    },
    {
      "name": "Level 1: Full Stack (no LiteLLM)",
      "configurations": [
        "Go Control Plane",
        "Oracle Capability",
        "Go Proxy (Capability)",
        "Go Proxy (Agent)",
        ".NET Agent"
      ]
    }
  ]
}
EOF

cat > .vscode/tasks.json << 'EOF'
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "build-agent",
      "type": "shell",
      "command": "dotnet build",
      "options": { "cwd": "${workspaceFolder}/examples/agent-dotnet" },
      "group": "build"
    },
    {
      "label": "build-oracle",
      "type": "shell",
      "command": "dotnet build",
      "options": { "cwd": "${workspaceFolder}/examples/capability-dotnet-oracle" },
      "group": "build"
    },
    {
      "label": "build-all",
      "type": "shell",
      "command": "make build",
      "group": { "kind": "build", "isDefault": true }
    },
    {
      "label": "test-all",
      "type": "shell",
      "command": "make test",
      "group": { "kind": "test", "isDefault": true }
    },
    {
      "label": "Start Ollama",
      "type": "shell",
      "command": "ollama serve",
      "isBackground": true,
      "problemMatcher": []
    },
    {
      "label": "Start LiteLLM",
      "type": "shell",
      "command": "litellm --config config/litellm_config.dev.yaml --port 4000",
      "isBackground": true,
      "problemMatcher": []
    }
  ]
}
EOF

cat > .vscode/extensions.json << 'EOF'
{
  "recommendations": [
    "ms-dotnettools.csdevkit",
    "ms-dotnettools.csharp",
    "golang.go",
    "redhat.vscode-yaml",
    "ms-azuretools.vscode-docker",
    "humao.rest-client"
  ]
}
EOF

cat > .vscode/settings.json << 'EOF'
{
  "go.toolsManagement.autoUpdate": true,
  "go.useLanguageServer": true,
  "dotnet-test-explorer.testProjectPath": "**/*.Tests.csproj",
  "yaml.schemas": {
    "proto/capability_manifest.schema.json": "**/manifest.yaml"
  }
}
EOF
```

---

## Step 10: Create .gitignore

```bash
cat > .gitignore << 'EOF'
# Go
bin/
*.exe

# .NET
**/bin/
**/obj/
*.user
*.suo

# IDE
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Environment
.env
*.env.local

# Build
dist/
node_modules/
EOF
```

---

## Step 11: Create REST Client Test File

```bash
cat > test.http << 'EOF'
### ===== Level 0: Direct Agent =====

### Submit a goal
POST http://localhost:5000/goals
Content-Type: application/json

{
  "goal": "Match invoice INV-4821 to cash receipt for customer CUST-4821"
}

### Check goal status
GET http://localhost:5000/goals/GOAL-abc123

### Health check
GET http://localhost:5000/health


### ===== Level 1+: Via Proxy =====

### Discover capabilities
POST http://localhost:3500/discover
Content-Type: application/json

{
  "context": {
    "goal": "Match invoice INV-4821 to cash receipt"
  },
  "provider": "openai",
  "max_tools": 20
}

### Invoke a capability
POST http://localhost:3500/invoke/stallion.retrieve_invoice_details
Content-Type: application/json

{
  "arguments": {
    "invoice_id": "INV-4821"
  },
  "caller_id": "stallion-agent",
  "goal_id": "GOAL-001",
  "tenant_id": "stallion"
}


### ===== Level 3+: Control Plane =====

### List registered capabilities
GET http://localhost:8080/api/v1/capabilities

### Health
GET http://localhost:8080/api/v1/health

### Capabilities health
GET http://localhost:8080/api/v1/health/capabilities
EOF
```

---

## Step 12: First Run (Level 0)

```bash
# Option A: With OpenRouter (recommended — no Ollama needed, saves 2.5GB RAM)

# Set your OpenRouter API key
set OPENROUTER_API_KEY=sk-or-v1-your-key-here

# Terminal 1: Start Agent (configure to use OpenRouter directly)
cd examples/agent-dotnet
dotnet run

# Option B: With local Ollama

# Terminal 1: Start Ollama
ollama serve

# Terminal 2: Start Agent
cd examples/agent-dotnet
dotnet run

# ── Test (same for both options) ──

# Terminal (new):
curl -X POST http://localhost:5000/goals ^
  -H "Content-Type: application/json" ^
  -d "{\"goal\": \"Match invoice INV-4821 to cash receipt for customer CUST-4821\"}"
```

If you see a response with a `goal_id` and `status`, CapaFabric Level 0 is working.

---

## Document Inventory

Place these in the repo root as `CLAUDE.md` (pick the one most relevant to what you're building):

| File | Use As | Purpose |
|---|---|---|
| `CAPAFABRIC_ALGORITHM.md` | `CLAUDE.md` in monorepo root | Full architecture spec (9 ABCs, 12 phases, CP + proxy) |
| `CAPAFABRIC_DOTNET_GUIDE.md` | `CLAUDE.md` in .NET Agent dir | .NET implementation with SK, 7-week plan |
| `CAPAFABRIC_SINGLE_AGENT.md` | Reference | Single agent with capabilities + manifest |
| `CAPAFABRIC_MULTI_AGENT.md` | Reference | 4-stage multi-agent pipeline |
| `CAPAFABRIC_CODEBASE.md` | `CLAUDE.md` in monorepo root | Go project structure for CP + proxy |
| `CAPAFABRIC_PROJECT_SETUP.md` | This file | Step-by-step setup guide |

---

## What to Build First

```
Week 1: Follow CAPAFABRIC_SINGLE_AGENT.md
  - Implement the Oracle Connector (examples/capability-dotnet-oracle/Program.cs)
  - Implement the Agent (examples/agent-dotnet/)
  - Run Level 0: Agent + Ollama
  - Verify: goal submitted → LLM reasons → capabilities invoked → result

Week 2: Follow CAPAFABRIC_CODEBASE.md Week 1-2
  - Implement shared/models and shared/interfaces
  - Implement control-plane with in-memory registry + REST API
  - Implement proxy with manifest loading + localhost API
  - Run Level 1: CP + Proxy + Capability + Agent

Week 3+: Follow CAPAFABRIC_ALGORITHM.md phases 6-12
  - Add each governance layer incrementally
```
