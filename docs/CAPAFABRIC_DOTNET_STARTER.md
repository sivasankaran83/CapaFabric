# CapaFabric — .NET Agent Starter (DI-First, SOLID)

> Copy these files into examples/agent-dotnet/ to get started.
> Every dependency is injected. Every interface has one responsibility.
> Nothing is built until it's needed (YAGNI). Nothing is duplicated (DRY).
> Everything is as simple as possible (KISS).

---

## Design Principles Applied

```
SOLID:
  S — Single Responsibility: each interface does ONE thing
      ICognitiveProvider → thinks
      ICapabilityDiscoveryClient → discovers
      ICapabilityInvocationClient → invokes
      IStateClient → persists
      Agent → orchestrates (but delegates everything)

  O — Open/Closed: Agent base class is open for extension (override points)
      but closed for modification (PursueGoalAsync never changes)

  I — Interface Segregation: small, focused interfaces
      Not one god-interface, but 4 narrow ones

  D — Dependency Inversion: Agent depends on abstractions, not concrete classes
      Program.cs wires concrete implementations via DI

DRY:
  Agent base class implements the universal pattern ONCE
  Concrete agents only provide: AgentId, Persona, optional overrides

YAGNI:
  Level 0: no proxy client, no state client, no OTEL — just Agent + CognitiveProvider
  Add each layer only when you need it (Level 1, 2, 3...)

KISS:
  Creating an Agent = extend base, write Persona, done
  4 files to understand the entire system
```

---

## File: Core/ICognitiveProvider.cs

```csharp
namespace CapaFabric.Core;

/// <summary>
/// The cognitive engine. Thinks. Decides. Polymorphic per agent.
/// Single responsibility: LLM interaction only.
/// </summary>
public interface ICognitiveProvider
{
    /// Simple: prompt in, result out
    Task<CognitiveResult> DeliberateAsync(
        CognitiveRequest request,
        CancellationToken ct = default);

    /// With tool calling: provider loops internally, calls toolInvoker for each tool_call
    Task<CognitiveResult> DeliberateAsync(
        CognitiveRequest request,
        Func<ToolCall, CancellationToken, Task<ToolResult>> toolInvoker,
        CancellationToken ct = default);
}
```

## File: Core/ICapabilityDiscoveryClient.cs

```csharp
namespace CapaFabric.Core;

/// <summary>
/// Discovers capabilities. Single responsibility: find what's available.
/// </summary>
public interface ICapabilityDiscoveryClient
{
    Task<IReadOnlyList<CapabilityMetadata>> DiscoverAsync(
        string? goal = null,
        int maxTools = 20,
        IEnumerable<string>? tags = null,
        CancellationToken ct = default);
}
```

## File: Core/ICapabilityInvocationClient.cs

```csharp
namespace CapaFabric.Core;

/// <summary>
/// Invokes a capability. Single responsibility: call a capability by ID.
/// </summary>
public interface ICapabilityInvocationClient
{
    Task<JsonElement> InvokeAsync(
        string capabilityId,
        JsonElement arguments,
        InvocationContext context,
        CancellationToken ct = default);
}
```

## File: Core/IStateClient.cs

```csharp
namespace CapaFabric.Core;

/// <summary>
/// Persists state. Single responsibility: get/set/delete by scope+key.
/// </summary>
public interface IStateClient
{
    Task<T?> GetAsync<T>(string scope, string key, CancellationToken ct = default);
    Task SetAsync<T>(string scope, string key, T value, CancellationToken ct = default);
    Task DeleteAsync(string scope, string key, CancellationToken ct = default);
}
```

---

## File: Core/Models.cs

```csharp
using System.Text.Json;
using System.Text.Json.Serialization;

namespace CapaFabric.Core;

// ── Capability ──

public record CapabilityMetadata
{
    [JsonPropertyName("capability_id")] public required string CapabilityId { get; init; }
    [JsonPropertyName("name")] public required string Name { get; init; }
    [JsonPropertyName("description")] public required string Description { get; init; }
    [JsonPropertyName("parameters_schema")] public JsonElement ParametersSchema { get; init; }
    [JsonPropertyName("tags")] public List<string> Tags { get; init; } = [];
    [JsonPropertyName("is_idempotent")] public bool IsIdempotent { get; init; } = true;
    [JsonPropertyName("has_side_effects")] public bool HasSideEffects { get; init; }
    [JsonPropertyName("requires_approval")] public bool RequiresApproval { get; init; }
}

// ── Cognitive ──

public record CognitiveRequest
{
    public required string Persona { get; init; }
    public required string Goal { get; init; }
    public List<ChatMessage> Context { get; init; } = [];
    public List<ToolDefinition> Capabilities { get; init; } = [];
    public required CognitiveConfig Config { get; init; }
}

public record CognitiveResult
{
    public string? Content { get; init; }
    public AgentDecision? Decision { get; init; }
    public List<ToolCall> ToolCalls { get; init; } = [];
    public int TokensUsed { get; init; }
}

public record CognitiveConfig
{
    public required string Model { get; init; }
    public int MaxTokens { get; init; } = 4096;
}

public record ChatMessage(string Role, string Content);
public record ToolDefinition(string Name, string Description, JsonElement Parameters);
public record ToolCall(string Id, string CapabilityId, JsonElement Arguments);
public record ToolResult(string Id, JsonElement Result);

// ── Agent ──

public record AgentDecision
{
    [JsonPropertyName("thought_process")] public required string ThoughtProcess { get; init; }
    [JsonPropertyName("selected_capability")] public string? SelectedCapability { get; init; }
    [JsonPropertyName("confidence_score")] public required double ConfidenceScore { get; init; }
    [JsonPropertyName("is_goal_complete")] public bool IsGoalComplete { get; init; }
}

public record AgentState
{
    public string GoalId { get; init; } = Guid.NewGuid().ToString()[..12];
    public required string GoalDescription { get; init; }
    public int Iteration { get; init; }
    public int MaxIterations { get; init; } = 10;
    public int TokensUsed { get; init; }
    public int TokenBudget { get; init; } = 50000;
    public string Status { get; init; } = "pending";
    public AgentDecision? LastDecision { get; init; }
}

// ── Context ──

public record InvocationContext
{
    public string RequestId { get; init; } = Guid.NewGuid().ToString();
    public required string CapabilityId { get; init; }
    public required string CallerId { get; init; }
    public string? GoalId { get; init; }
    public List<string> CallChain { get; init; } = [];
    public int MaxDepth { get; init; } = 10;
}

public record InvocationRecord(string CapabilityId, string ArgsHash);

// ── Agent Context (carries call chain + history) ──

public record AgentContext
{
    public required string GoalId { get; init; }
    public List<ChatMessage> History { get; init; } = [];
    public List<string> CallChain { get; init; } = [];
    public List<InvocationRecord> InvocationLog { get; init; } = [];
    public int MaxDepth { get; init; } = 10;
    public int MaxConsecutiveRetries { get; init; } = 3;

    public AgentContext WithAgent(string agentId) => this with
    {
        CallChain = [.. CallChain, agentId]
    };
}

// ── Agent Result ──

public enum AgentResultType { Completed, Continue, EscalateToHuman, Retry, Failed }

public record AgentResult
{
    public required AgentResultType Type { get; init; }
    public AgentDecision? Decision { get; init; }
    public string? Reason { get; init; }

    public static AgentResult Completed(AgentDecision decision) => new()
        { Type = AgentResultType.Completed, Decision = decision };
    public static AgentResult ContinueIteration(AgentDecision? decision) => new()
        { Type = AgentResultType.Continue, Decision = decision };
    public static AgentResult EscalateToHuman(string reason) => new()
        { Type = AgentResultType.EscalateToHuman, Reason = reason };
    public static AgentResult RetryIteration(string reason) => new()
        { Type = AgentResultType.Retry, Reason = reason };
    public static AgentResult Failed(string reason) => new()
        { Type = AgentResultType.Failed, Reason = reason };
}
```

---

## File: Agents/Agent.cs (Abstract Base)

```csharp
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using CapaFabric.Core;

namespace CapaFabric.Agents;

/// <summary>
/// Abstract Agent base class. THE universal building block of CapaFabric.
///
/// Implements: discover → guard → deliberate → invoke → evaluate
/// Concrete agents provide: AgentId, Persona, optional overrides.
///
/// All dependencies injected. No concrete types referenced.
/// </summary>
public abstract class Agent
{
    private readonly ICognitiveProvider _cognitive;
    private readonly ICapabilityDiscoveryClient _discovery;
    private readonly ICapabilityInvocationClient _invoker;

    protected Agent(
        ICognitiveProvider cognitive,
        ICapabilityDiscoveryClient discovery,
        ICapabilityInvocationClient invoker)
    {
        _cognitive = cognitive;
        _discovery = discovery;
        _invoker = invoker;
    }

    // ── Identity (concrete agents provide these) ──

    public abstract string AgentId { get; }
    public abstract string Persona { get; }

    // ── The universal pattern ──

    public async Task<AgentResult> PursueGoalAsync(
        string goal, AgentContext context, CancellationToken ct = default)
    {
        // Guard 1: Circular chain
        if (context.CallChain.Contains(AgentId))
        {
            var chain = string.Join(" -> ", context.CallChain) + " -> " + AgentId;
            return AgentResult.Failed($"Circular agent chain detected: {chain}");
        }

        // Guard 2: Max depth
        if (context.CallChain.Count >= context.MaxDepth)
            return AgentResult.EscalateToHuman(
                $"Max agent depth ({context.MaxDepth}) exceeded.");

        var downstream = context.WithAgent(AgentId);

        // Discover
        var capabilities = await _discovery.DiscoverAsync(goal: goal, ct: ct);
        if (capabilities.Count == 0)
            return await OnNoCapabilitiesFound(goal, context, ct);

        // Deliberate (with tool invocation callback)
        var result = await _cognitive.DeliberateAsync(
            new CognitiveRequest
            {
                Persona = Persona,
                Goal = goal,
                Context = context.History,
                Capabilities = capabilities
                    .Select(c => new ToolDefinition(c.Name, c.Description, c.ParametersSchema))
                    .ToList(),
                Config = GetCognitiveConfig(),
            },
            toolInvoker: async (toolCall, ct2) =>
            {
                // Guard 3: Retry loop
                var argsHash = HashArguments(toolCall.Arguments);
                var lastN = context.InvocationLog.TakeLast(context.MaxConsecutiveRetries).ToList();
                if (lastN.Count >= context.MaxConsecutiveRetries &&
                    lastN.All(r => r.CapabilityId == toolCall.CapabilityId && r.ArgsHash == argsHash))
                {
                    var errorResponse = JsonSerializer.SerializeToElement(new
                    {
                        error = "retry_loop_detected",
                        message = $"'{toolCall.CapabilityId}' invoked {context.MaxConsecutiveRetries} times " +
                                  "with identical arguments. Try a different approach or escalate."
                    });
                    return new ToolResult(toolCall.Id, errorResponse);
                }

                // Invoke via the injected client (proxy or direct)
                var invokeResult = await _invoker.InvokeAsync(
                    toolCall.CapabilityId,
                    toolCall.Arguments,
                    new InvocationContext
                    {
                        CapabilityId = toolCall.CapabilityId,
                        CallerId = AgentId,
                        GoalId = context.GoalId,
                        CallChain = downstream.CallChain,
                        MaxDepth = downstream.MaxDepth,
                    },
                    ct2);

                context.InvocationLog.Add(new InvocationRecord(toolCall.CapabilityId, argsHash));
                return new ToolResult(toolCall.Id, invokeResult);
            },
            ct);

        // Evaluate
        if (result.Decision?.IsGoalComplete == true)
            return AgentResult.Completed(result.Decision);

        if (result.Decision is not null &&
            result.Decision.ConfidenceScore < GetConfidenceThreshold(result.Decision))
            return await OnLowConfidence(result.Decision, goal, context, ct);

        return AgentResult.ContinueIteration(result.Decision);
    }

    // ── Override points (Open/Closed principle) ──

    protected virtual Task<AgentResult> OnNoCapabilitiesFound(
        string goal, AgentContext context, CancellationToken ct)
        => Task.FromResult(AgentResult.EscalateToHuman(
            $"No capabilities found for goal: {goal}"));

    protected virtual Task<AgentResult> OnLowConfidence(
        AgentDecision decision, string goal, AgentContext context, CancellationToken ct)
        => Task.FromResult(AgentResult.EscalateToHuman(
            $"Confidence {decision.ConfidenceScore:F2} below threshold for '{decision.SelectedCapability}'"));

    protected virtual CognitiveConfig GetCognitiveConfig()
        => new() { Model = "reasoning-heavy" };

    protected virtual double GetConfidenceThreshold(AgentDecision decision)
        => 0.85;

    // ── Private helpers ──

    private static string HashArguments(JsonElement args)
        => Convert.ToHexString(SHA256.HashData(Encoding.UTF8.GetBytes(args.GetRawText())))[..16];
}
```

---

## File: Agents/AgentRunner.cs

```csharp
using CapaFabric.Core;
using Microsoft.Extensions.Logging;

namespace CapaFabric.Agents;

/// <summary>
/// Runs the Agent iteration loop. Single responsibility: iteration control.
/// Handles: max iterations, token budget, checkpointing, status transitions.
/// Does NOT contain reasoning logic — that's in Agent.PursueGoalAsync.
/// </summary>
public class AgentRunner
{
    private readonly IStateClient _state;
    private readonly ILogger<AgentRunner> _logger;

    public AgentRunner(IStateClient state, ILogger<AgentRunner> logger)
    {
        _state = state;
        _logger = logger;
    }

    public async Task<AgentState> RunAsync(
        Agent agent, string goal, CancellationToken ct = default)
    {
        var state = new AgentState { GoalDescription = goal, Status = "running" };
        var context = new AgentContext { GoalId = state.GoalId };

        _logger.LogInformation("[{AgentId}] Goal {GoalId}: {Goal}",
            agent.AgentId, state.GoalId, goal);

        while (state.Status == "running"
            && state.Iteration < state.MaxIterations
            && state.TokensUsed < state.TokenBudget)
        {
            var result = await agent.PursueGoalAsync(goal, context, ct);

            state = state with
            {
                Iteration = state.Iteration + 1,
                LastDecision = result.Decision,
                Status = result.Type switch
                {
                    AgentResultType.Completed => "completed",
                    AgentResultType.EscalateToHuman => "hitl_suspended",
                    AgentResultType.Failed => "failed",
                    _ => "running"
                }
            };

            if (result.Decision is not null)
                context.History.Add(new ChatMessage("assistant", result.Decision.ThoughtProcess));

            _logger.LogInformation("[{AgentId}] Iteration {Iter}: {Status}",
                agent.AgentId, state.Iteration, state.Status);

            // Checkpoint
            await _state.SetAsync("agent", state.GoalId, state, ct);
        }

        if (state.Iteration >= state.MaxIterations && state.Status == "running")
            state = state with { Status = "max_iterations" };
        if (state.TokensUsed >= state.TokenBudget && state.Status == "running")
            state = state with { Status = "budget_exhausted" };

        await _state.SetAsync("agent", state.GoalId, state, ct);
        _logger.LogInformation("[{AgentId}] Goal {GoalId}: final status = {Status}",
            agent.AgentId, state.GoalId, state.Status);

        return state;
    }
}
```

---

## File: Agents/CashMatchingAgent.cs (Concrete — 15 lines)

```csharp
using CapaFabric.Core;

namespace CapaFabric.Agents;

/// <summary>
/// Stallion Cash Matching Agent.
/// DRY: Only provides what's unique — AgentId, Persona, thresholds.
/// Everything else inherited from Agent base class.
/// </summary>
public class CashMatchingAgent : Agent
{
    public CashMatchingAgent(
        ICognitiveProvider cognitive,
        ICapabilityDiscoveryClient discovery,
        ICapabilityInvocationClient invoker)
        : base(cognitive, discovery, invoker) { }

    public override string AgentId => "stallion-cash-matching";

    public override string Persona => """
        You are a financial reconciliation agent for Stallion Insurance.

        Your objective is to achieve a 3-way match between an invoice,
        a corresponding bank cash receipt, and a purchase order.

        RULES:
        - Gather all necessary data before attempting to match.
        - Never post a financial transaction unless a match is confirmed.
        - If any variance exceeds tolerance, escalate to a human reviewer.
        - Generate an audit trail after every completed or escalated match.
        - Use the available capabilities to accomplish your objective.
          Do not assume which capabilities exist.

        Think step by step. Explain your reasoning at each step.
        """;

    protected override CognitiveConfig GetCognitiveConfig()
        => new() { Model = "anthropic/claude-sonnet-4" };

    protected override double GetConfidenceThreshold(AgentDecision decision)
        => 0.95;  // Strict for financial operations
}
```

---

## File: Infrastructure/OpenRouterCognitiveProvider.cs

```csharp
using System.Text.Json;
using CapaFabric.Core;

namespace CapaFabric.Infrastructure;

/// <summary>
/// CognitiveProvider that calls OpenRouter (or any OpenAI-compatible API).
/// Single responsibility: LLM HTTP interaction.
/// Injected via DI — Agent doesn't know this exists.
/// </summary>
public class OpenRouterCognitiveProvider : ICognitiveProvider
{
    private readonly HttpClient _http;

    public OpenRouterCognitiveProvider(HttpClient http)
    {
        _http = http;
    }

    public async Task<CognitiveResult> DeliberateAsync(
        CognitiveRequest request, CancellationToken ct = default)
    {
        var body = BuildRequestBody(request);
        var response = await _http.PostAsJsonAsync("chat/completions", body, ct);
        response.EnsureSuccessStatusCode();
        var json = await response.Content.ReadFromJsonAsync<JsonElement>(ct);

        return ParseResponse(json);
    }

    public async Task<CognitiveResult> DeliberateAsync(
        CognitiveRequest request,
        Func<ToolCall, CancellationToken, Task<ToolResult>> toolInvoker,
        CancellationToken ct = default)
    {
        var body = BuildRequestBody(request);

        // Tool calling loop: call LLM → if tool_call → invoke → feed back → repeat
        while (true)
        {
            var response = await _http.PostAsJsonAsync("chat/completions", body, ct);
            response.EnsureSuccessStatusCode();
            var json = await response.Content.ReadFromJsonAsync<JsonElement>(ct);

            var choice = json.GetProperty("choices")[0];
            var message = choice.GetProperty("message");
            var finishReason = choice.GetProperty("finish_reason").GetString();

            if (finishReason == "tool_calls" &&
                message.TryGetProperty("tool_calls", out var toolCalls))
            {
                // LLM wants to invoke capabilities
                var messages = body.GetProperty("messages").Deserialize<List<JsonElement>>()!;
                messages.Add(message);

                foreach (var tc in toolCalls.EnumerateArray())
                {
                    var toolCall = new ToolCall(
                        tc.GetProperty("id").GetString()!,
                        tc.GetProperty("function").GetProperty("name").GetString()!,
                        tc.GetProperty("function").GetProperty("arguments").Clone());

                    var result = await toolInvoker(toolCall, ct);

                    messages.Add(JsonSerializer.SerializeToElement(new
                    {
                        role = "tool",
                        tool_call_id = result.Id,
                        content = result.Result.GetRawText()
                    }));
                }

                // Rebuild body with updated messages for next LLM call
                body = RebuildWithMessages(body, messages);
                continue;
            }

            // No more tool calls — return final result
            return ParseResponse(json);
        }
    }

    private static JsonElement BuildRequestBody(CognitiveRequest request)
    {
        var messages = new List<object>
        {
            new { role = "system", content = request.Persona }
        };

        foreach (var msg in request.Context)
            messages.Add(new { role = msg.Role, content = msg.Content });

        messages.Add(new { role = "user", content = request.Goal });

        var body = new
        {
            model = request.Config.Model,
            messages,
            max_tokens = request.Config.MaxTokens,
            tools = request.Capabilities.Select(c => new
            {
                type = "function",
                function = new
                {
                    name = c.Name,
                    description = c.Description,
                    parameters = c.Parameters
                }
            }).ToArray()
        };

        return JsonSerializer.SerializeToElement(body);
    }

    private static JsonElement RebuildWithMessages(JsonElement body, List<JsonElement> messages)
    {
        var dict = body.Deserialize<Dictionary<string, JsonElement>>()!;
        dict["messages"] = JsonSerializer.SerializeToElement(messages);
        return JsonSerializer.SerializeToElement(dict);
    }

    private static CognitiveResult ParseResponse(JsonElement json)
    {
        var content = json.GetProperty("choices")[0]
            .GetProperty("message")
            .GetProperty("content")
            .GetString() ?? "";

        var tokensUsed = json.TryGetProperty("usage", out var usage)
            ? usage.GetProperty("total_tokens").GetInt32()
            : 0;

        AgentDecision? decision = null;
        try { decision = JsonSerializer.Deserialize<AgentDecision>(content); }
        catch { /* LLM returned free text, not structured — that's OK */ }

        return new CognitiveResult
        {
            Content = content,
            Decision = decision,
            TokensUsed = tokensUsed
        };
    }
}
```

---

## File: Infrastructure/InMemoryDiscoveryClient.cs (Level 0)

```csharp
using System.Text.Json;
using CapaFabric.Core;

namespace CapaFabric.Infrastructure;

/// <summary>
/// Level 0: hardcoded capabilities for development.
/// Replace with ProxyDiscoveryClient when proxy is available (Level 1+).
/// YAGNI: we don't build the proxy client until we need it.
/// </summary>
public class InMemoryDiscoveryClient : ICapabilityDiscoveryClient
{
    private readonly List<CapabilityMetadata> _capabilities;

    public InMemoryDiscoveryClient(IEnumerable<CapabilityMetadata> capabilities)
    {
        _capabilities = capabilities.ToList();
    }

    public Task<IReadOnlyList<CapabilityMetadata>> DiscoverAsync(
        string? goal = null, int maxTools = 20,
        IEnumerable<string>? tags = null, CancellationToken ct = default)
    {
        IEnumerable<CapabilityMetadata> result = _capabilities;

        if (tags is not null)
        {
            var tagSet = tags.ToHashSet();
            result = result.Where(c => c.Tags.Any(t => tagSet.Contains(t)));
        }

        return Task.FromResult<IReadOnlyList<CapabilityMetadata>>(
            result.Take(maxTools).ToList());
    }
}
```

---

## File: Infrastructure/HttpInvocationClient.cs (Level 0)

```csharp
using System.Text.Json;
using CapaFabric.Core;

namespace CapaFabric.Infrastructure;

/// <summary>
/// Level 0: Direct HTTP invocation (no proxy).
/// Replace with ProxyInvocationClient when proxy is available (Level 1+).
/// </summary>
public class HttpInvocationClient : ICapabilityInvocationClient
{
    private readonly HttpClient _http;

    public HttpInvocationClient(HttpClient http)
    {
        _http = http;
    }

    public async Task<JsonElement> InvokeAsync(
        string capabilityId, JsonElement arguments,
        InvocationContext context, CancellationToken ct = default)
    {
        // Level 0: direct call to capability service
        // In production, ProxyInvocationClient calls localhost:3500/invoke/{capabilityId}
        var response = await _http.PostAsJsonAsync(
            $"invoke/{capabilityId}",
            new { arguments, caller_id = context.CallerId, goal_id = context.GoalId },
            ct);

        response.EnsureSuccessStatusCode();
        return await response.Content.ReadFromJsonAsync<JsonElement>(ct);
    }
}
```

---

## File: Infrastructure/InMemoryStateClient.cs

```csharp
using System.Collections.Concurrent;
using System.Text.Json;
using CapaFabric.Core;

namespace CapaFabric.Infrastructure;

/// <summary>
/// In-process state. No external dependency. Use for Level 0-3.
/// Replace with ProxyStateClient when sidecar state is needed.
/// </summary>
public class InMemoryStateClient : IStateClient
{
    private readonly ConcurrentDictionary<string, string> _store = new();

    public Task<T?> GetAsync<T>(string scope, string key, CancellationToken ct = default)
    {
        var compositeKey = $"{scope}:{key}";
        return Task.FromResult(
            _store.TryGetValue(compositeKey, out var json)
                ? JsonSerializer.Deserialize<T>(json)
                : default);
    }

    public Task SetAsync<T>(string scope, string key, T value, CancellationToken ct = default)
    {
        _store[$"{scope}:{key}"] = JsonSerializer.Serialize(value);
        return Task.CompletedTask;
    }

    public Task DeleteAsync(string scope, string key, CancellationToken ct = default)
    {
        _store.TryRemove($"{scope}:{key}", out _);
        return Task.CompletedTask;
    }
}
```

---

## File: Program.cs (DI Composition Root)

```csharp
using CapaFabric.Agents;
using CapaFabric.Core;
using CapaFabric.Infrastructure;

var builder = WebApplication.CreateBuilder(args);
var config = builder.Configuration;

// ══ Register infrastructure (swap implementations here, nowhere else) ══

// CognitiveProvider: OpenRouter via OpenAI-compatible API
builder.Services.AddHttpClient<ICognitiveProvider, OpenRouterCognitiveProvider>(client =>
{
    client.BaseAddress = new Uri(config["LLM:Endpoint"]!);
    client.DefaultRequestHeaders.Add("Authorization", $"Bearer {config["LLM:ApiKey"]}");
});

// Discovery: in-memory for Level 0 (swap to ProxyDiscoveryClient for Level 1+)
builder.Services.AddSingleton<ICapabilityDiscoveryClient>(_ =>
    new InMemoryDiscoveryClient(SeedCapabilities()));

// Invocation: direct HTTP for Level 0 (swap to ProxyInvocationClient for Level 1+)
builder.Services.AddHttpClient<ICapabilityInvocationClient, HttpInvocationClient>(client =>
{
    client.BaseAddress = new Uri(config["Capabilities:BaseUrl"] ?? "http://localhost:8081/api/");
});

// State: in-memory for Level 0 (swap to ProxyStateClient for Level 4+)
builder.Services.AddSingleton<IStateClient, InMemoryStateClient>();

// ══ Register agents (each gets dependencies injected automatically) ══

builder.Services.AddScoped<CashMatchingAgent>();
builder.Services.AddScoped<AgentRunner>();

var app = builder.Build();

// ══ Endpoints ══

app.MapPost("/goals", async (GoalRequest req, CashMatchingAgent agent, AgentRunner runner, CancellationToken ct) =>
    Results.Ok(await runner.RunAsync(agent, req.Goal, ct)));

app.MapGet("/goals/{goalId}", async (string goalId, IStateClient state, CancellationToken ct) =>
{
    var goal = await state.GetAsync<AgentState>("agent", goalId, ct);
    return goal is not null ? Results.Ok(goal) : Results.NotFound();
});

app.MapGet("/health", () => Results.Ok(new { status = "healthy", agent = "stallion-cash-matching" }));

app.Run();

// ══ Request DTO ══

record GoalRequest(string Goal);

// ══ Seed capabilities for Level 0 (replace with discovery in Level 1+) ══

static List<CapabilityMetadata> SeedCapabilities() =>
[
    new()
    {
        CapabilityId = "stallion.retrieve_invoice_details",
        Name = "retrieve_invoice_details",
        Description = "Retrieves invoice details from Oracle AR including line items, customer info, and payment status.",
        Tags = ["oracle", "invoice", "read"],
    },
    new()
    {
        CapabilityId = "stallion.locate_matching_receipts",
        Name = "locate_matching_receipts",
        Description = "Searches bank lockbox cash receipts by customer ID and date range.",
        Tags = ["bank", "receipt", "read"],
    },
    new()
    {
        CapabilityId = "stallion.verify_three_way_match",
        Name = "verify_three_way_match",
        Description = "Performs deterministic 3-way match between invoice, receipt, and PO amounts.",
        Tags = ["matching", "reconciliation"],
    },
    new()
    {
        CapabilityId = "stallion.commit_financial_posting",
        Name = "commit_financial_posting",
        Description = "Posts matched receipt against invoice in Oracle AR. Irreversible financial transaction.",
        Tags = ["oracle", "posting", "write", "critical"],
        HasSideEffects = true,
        RequiresApproval = true,
    },
    new()
    {
        CapabilityId = "stallion.escalate_variance_to_reviewer",
        Name = "escalate_variance_to_reviewer",
        Description = "Sends variance notification to AR clerk for manual review.",
        Tags = ["notification", "alert"],
        HasSideEffects = true,
    },
    new()
    {
        CapabilityId = "stallion.produce_compliance_audit",
        Name = "produce_compliance_audit",
        Description = "Generates PDF audit document recording the complete matching journey.",
        Tags = ["audit", "compliance"],
        HasSideEffects = true,
    },
];
```

---

## File: appsettings.Development.json

```json
{
  "Logging": {
    "LogLevel": { "Default": "Information" }
  },
  "LLM": {
    "Endpoint": "https://openrouter.ai/api/v1/",
    "ApiKey": "your-openrouter-key-here"
  },
  "Capabilities": {
    "BaseUrl": "http://localhost:8081/api/"
  }
}
```

---

## Upgrade Path (YAGNI — add only when needed)

```
Level 0 (now):
  ICognitiveProvider     → OpenRouterCognitiveProvider
  ICapabilityDiscovery   → InMemoryDiscoveryClient (hardcoded list)
  ICapabilityInvocation  → HttpInvocationClient (direct HTTP)
  IStateClient           → InMemoryStateClient

Level 1 (when proxy is built):
  ICapabilityDiscovery   → ProxyDiscoveryClient    ← swap one line in Program.cs
  ICapabilityInvocation  → ProxyInvocationClient   ← swap one line in Program.cs

Level 2 (when LiteLLM is running):
  ICognitiveProvider     → LiteLLMCognitiveProvider ← swap one line in Program.cs

Level 4 (when Redis is running):
  IStateClient           → ProxyStateClient         ← swap one line in Program.cs

Each level: change ONE registration in Program.cs. Zero changes to Agent, AgentRunner,
or CashMatchingAgent. That's Dependency Inversion in practice.
```
