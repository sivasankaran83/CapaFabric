# CapaFabric — .NET Agent Implementation Guide

> Implementation specification for building a .NET Semantic Kernel Agent
> integrated with the CapaFabric control plane and thin proxy.
> Drop as `CLAUDE.md` in your .NET Agent repository.

---

## Quick Start (Level 0 — Two Processes)

```bash
# Terminal 1: Ollama
ollama serve && ollama pull llama3.2

# Terminal 2: .NET Agent
cd src/Agent && dotnet run
```

---

## Repository Structure

```
capafabric-agent-dotnet/
├── CLAUDE.md
├── src/
│   ├── Agent/
│   │   ├── Agent.csproj
│   │   ├── Program.cs
│   │   ├── Core/
│   │   │   ├── ICapabilityDiscoveryClient.cs
│   │   │   ├── ICapabilityInvocationClient.cs
│   │   │   ├── IStateClient.cs
│   │   │   ├── CapabilityMetadata.cs
│   │   │   ├── AgentDecision.cs
│   │   │   ├── AgentState.cs
│   │   │   └── InvocationContext.cs
│   │   ├── Agents/
│   │   │   ├── AgentRunner.cs            # Main reasoning loop
│   │   │   ├── CapabilityBridge.cs       # Bridges discovered capabilities → SK KernelFunctions
│   │   │   └── HITLGateway.cs
│   │   ├── Infrastructure/
│   │   │   ├── ProxyDiscoveryClient.cs   # HTTP → localhost:3500/discover
│   │   │   ├── ProxyInvocationClient.cs   # HTTP → localhost:3500/invoke
│   │   │   ├── ProxyStateClient.cs       # HTTP → localhost:3500/state
│   │   │   └── InMemoryStateClient.cs    # Level 0: no proxy
│   │   ├── Configuration/
│   │   │   ├── AgentOptions.cs
│   │   │   └── LLMOptions.cs
│   │   └── Endpoints/
│   │       ├── GoalEndpoints.cs
│   │       └── HealthEndpoints.cs
│   └── Agent.Tests/
│       └── ...
├── capabilities/                          # Example capabilities
│   ├── oracle-connector/
│   │   ├── OracleConnector.csproj
│   │   ├── Program.cs
│   │   └── manifest.yaml
│   └── matching-engine/
│       ├── go.mod
│       ├── main.go
│       └── manifest.yaml
└── docker-compose.yaml
```

---

## Phase 1: Domain Model

### NuGet Packages

```xml
<!-- Target .NET 10 (LTS). Portability: change to net8.0 on need basis. -->
<!-- <TargetFramework>net10.0</TargetFramework> -->
<PackageReference Include="Microsoft.SemanticKernel" Version="1.*" />
<PackageReference Include="Microsoft.SemanticKernel.Connectors.OpenAI" Version="1.*" />
<PackageReference Include="Microsoft.Extensions.Http" Version="10.*" />
<PackageReference Include="OpenTelemetry.Extensions.Hosting" Version="1.*" />
<PackageReference Include="OpenTelemetry.Exporter.OpenTelemetryProtocol" Version="1.*" />
```

### Interfaces

```csharp
// Core/ICapabilityDiscoveryClient.cs
public interface ICapabilityDiscoveryClient
{
    Task<IReadOnlyList<CapabilityMetadata>> DiscoverAsync(
        string? goal = null, int maxTools = 20,
        IEnumerable<string>? tags = null, CancellationToken ct = default);
}

// Core/ICapabilityInvocationClient.cs
public interface ICapabilityInvocationClient
{
    Task<JsonElement> InvokeAsync(
        string capabilityId, JsonElement arguments,
        InvocationContext context, CancellationToken ct = default);
}

// Core/IStateClient.cs
public interface IStateClient
{
    Task<T?> GetAsync<T>(string scope, string key, CancellationToken ct = default);
    Task SetAsync<T>(string scope, string key, T value, CancellationToken ct = default);
    Task DeleteAsync(string scope, string key, CancellationToken ct = default);
}
```

### AgentDecision (Structured Output)

```csharp
public record AgentDecision
{
    [JsonPropertyName("thought_process")]
    public required string ThoughtProcess { get; init; }

    [JsonPropertyName("selected_capability")]
    public string? SelectedCapability { get; init; }

    [JsonPropertyName("capability_input")]
    public JsonElement? CapabilityInput { get; init; }

    [JsonPropertyName("confidence_score")]
    public required double ConfidenceScore { get; init; }

    [JsonPropertyName("is_goal_complete")]
    public bool IsGoalComplete { get; init; } = false;
}
```

---

## Phase 2: Agent

### CapabilityBridge.cs

```
Bridges CapaFabric-discovered capabilities into Semantic Kernel.

Algorithm:
1. Call ICapabilityDiscoveryClient.DiscoverAsync(goal)
2. For each CapabilityMetadata:
   a. Create KernelFunction with name, description, parameters from metadata
   b. Function body calls ICapabilityInvocationClient.InvokeAsync(capability_id, args)
3. Register all as a "CapaFabric" plugin on the Kernel

This is the .NET equivalent of CapaFabric's runtime capability discovery.
SK function calling ≡ discover() + invoke()
```

### AgentRunner.cs

```
Main reasoning loop. Algorithm:

1. Initialize AgentState → checkpoint via IStateClient
2. Discover capabilities → register as KernelFunctions via CapabilityBridge
3. Loop while !complete && iteration < max && tokens < budget:
   a. Build context (system prompt + goal + history summary)
   b. Call LLM via SK (FunctionChoiceBehavior.Auto)
   c. SK handles tool calling → CapabilityBridge → proxy → capability
   d. HITL gateway: evaluate confidence
   e. Monotonic progress check (3 no-change → escalate)
   f. Checkpoint AgentState

Agent does NOT know about: guardrails, context compression, auth, transport, tracing
Agent ONLY knows about: goal logic, HITL thresholds, iteration/budget limits
```

### HITLGateway.cs

```
Evaluates confidence against per-capability thresholds.
Tag-based: critical→0.95, finance→0.95, write→0.90, read→0.80
Side effects get stricter threshold. requires_approval always escalates.
```

---

## Phase 3: Infrastructure

### ProxyDiscoveryClient.cs
POST localhost:3500/discover → List<CapabilityMetadata>

### ProxyInvocationClient.cs
POST localhost:3500/invoke/{capabilityId} → result
Errors: 401→Auth, 403→Authz, 404→NotFound, 429→RateLimit

### ProxyStateClient.cs
GET/POST/DELETE localhost:3500/state/{scope}/{key}

### InMemoryStateClient.cs (Level 0)
ConcurrentDictionary — no proxy needed

---

## Phase 4: Program.cs

```csharp
var builder = WebApplication.CreateBuilder(args);
var config = builder.Configuration;
var useProxy = config.GetValue<bool>("Agent:UseProxy");

// Semantic Kernel
var kernelBuilder = Kernel.CreateBuilder();
kernelBuilder.AddOpenAIChatCompletion(
    modelId: config["LLM:ModelId"]!,        // "llama3.2" or "reasoning-heavy"
    endpoint: new Uri(config["LLM:Endpoint"]!), // Ollama or proxy→LiteLLM
    apiKey: config["LLM:ApiKey"]!
);
builder.Services.AddSingleton(kernelBuilder.Build());

// Infrastructure: swap via config
if (useProxy)
{
    var url = config["Agent:ProxyUrl"]!;
    builder.Services.AddHttpClient<ICapabilityDiscoveryClient, ProxyDiscoveryClient>(c =>
        c.BaseAddress = new Uri(url));
    builder.Services.AddHttpClient<ICapabilityInvocationClient, ProxyInvocationClient>(c =>
        c.BaseAddress = new Uri(url));
    builder.Services.AddHttpClient<IStateClient, ProxyStateClient>(c =>
        c.BaseAddress = new Uri(url));
}
else
{
    builder.Services.AddSingleton<ICapabilityDiscoveryClient, InProcessDiscoveryClient>();
    builder.Services.AddSingleton<ICapabilityInvocationClient, InProcessInvocationClient>();
    builder.Services.AddSingleton<IStateClient, InMemoryStateClient>();
}

builder.Services.AddSingleton(new HITLGateway(
    config.GetValue("Agent:DefaultConfidenceThreshold", 0.85)));
builder.Services.AddScoped<CapabilityBridge>();
builder.Services.AddScoped<AgentRunner>();

var app = builder.Build();

app.MapPost("/goals", async (GoalRequest req, AgentRunner agent, CancellationToken ct) =>
    Results.Ok(await agent.PursueGoalAsync(req.Goal, ct)));

app.MapGet("/goals/{goalId}", async (string goalId, IStateClient state, CancellationToken ct) =>
{
    var goal = await state.GetAsync<AgentState>("goal", goalId, ct);
    return goal is not null ? Results.Ok(goal) : Results.NotFound();
});

app.MapGet("/health", () => Results.Ok(new { status = "healthy" }));
app.Run();

record GoalRequest(string Goal);
```

---

## Configuration

```json
// appsettings.Development.json
{
  "LLM": { "ModelId": "llama3.2", "Endpoint": "http://localhost:11434/v1", "ApiKey": "ollama" },
  "Agent": { "UseProxy": false, "ProxyUrl": "http://localhost:3500",
                  "DefaultConfidenceThreshold": 0.85, "MaxIterations": 10, "EnableOTel": false }
}

// appsettings.Production.json
{
  "LLM": { "ModelId": "reasoning-heavy", "Endpoint": "http://localhost:3500/llm/v1", "ApiKey": "${LITELLM_API_KEY}" },
  "Agent": { "UseProxy": true, "ProxyUrl": "http://localhost:3500",
                  "DefaultConfidenceThreshold": 0.85, "EnableOTel": true,
                  "OTelEndpoint": "http://otel-collector:4317" }
}
```

---

## VS Code Configuration

```json
// .vscode/launch.json
{
  "configurations": [
    { "name": ".NET Agent", "type": "coreclr", "request": "launch",
      "program": "${workspaceFolder}/src/Agent/bin/Debug/net10.0/Agent.dll",
      "env": { "ASPNETCORE_ENVIRONMENT": "Development", "ASPNETCORE_URLS": "http://localhost:5000" } },
    { "name": "Oracle Capability", "type": "coreclr", "request": "launch",
      "program": "${workspaceFolder}/capabilities/oracle-connector/bin/Debug/net10.0/OracleConnector.dll",
      "env": { "ASPNETCORE_URLS": "http://localhost:8081" } }
  ],
  "compounds": [
    { "name": "Level 0: Agent Only", "configurations": [".NET Agent"] },
    { "name": "Level 1: Agent + Capability", "configurations": [".NET Agent", "Oracle Capability"] }
  ]
}
```

---

## Development Levels

```
Level 0:  Agent + Ollama                   (2 processes)
Level 1:  + proxy + capability                 (4 processes)
Level 2:  + LiteLLM                            (5 processes)
Level 3:  + control plane                      (6 processes)
Level 4:  + OTEL + Jaeger                      (8 processes)
Level 5:  + Redis + governance                 (9 processes)
Level 6:  docker compose up                    (everything)
```

---

## Implementation Order

```
Week 1:  Domain model + Program.cs + in-memory clients → Level 0
Week 2:  AgentRunner + HITL + CapabilityBridge → reasoning loop working
Week 3:  Proxy clients + Oracle capability + manifest → Level 1
Week 4:  LiteLLM + control plane → Level 2-3
Week 5:  OTEL + guardrails → Level 4
Week 6:  Redis + auth + audit → Level 5
Week 7:  Docker Compose + CI → Level 6
```
