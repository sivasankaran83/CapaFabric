# CapaFabric — Single Agent Example (Stallion Cash Matching)

> Complete .NET agent: Agent (Semantic Kernel) + Capabilities (ASP.NET Minimal API).
> The Agent discovers and invokes capabilities at runtime. No hardcoded workflow.
> System prompt describes goal intent only — never capability names or sequences.

---

## Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│  STALLION CASH MATCHING AGENT                                       │
│                                                                      │
│  ┌───────────────────────┐    ┌──────────────────────────────────┐ │
│  │  GOALAGENT             │    │  CAPABILITIES                    │ │
│  │  (Semantic Kernel)     │    │  (ASP.NET Minimal API)           │ │
│  │                        │    │                                  │ │
│  │  Goal: "Reconcile      │───►│  • retrieve_invoice_details      │ │
│  │   invoices of customer │    │  • locate_matching_receipts      │ │
│  │   ABC for Feb 2026"    │◄───│  • verify_three_way_match        │ │
│  │                        │    │  • commit_financial_posting       │ │
│  │  Confidence gating     │    │  • escalate_variance_to_reviewer │ │
│  │  HITL escalation       │    │  • produce_compliance_audit      │ │
│  └───────────────────────┘    └──────────────────────────────────┘ │
│           │                              │                          │
│    localhost:3500                  localhost:3500                    │
│    (Proxy: discover,              (Proxy: manifest                  │
│     invoke, llm/chat)             registration)                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## Capabilities Service: capabilities/oracle-connector/Program.cs

```csharp
// Standard ASP.NET Minimal API. ZERO CapaFabric dependency.
// The manifest.yaml maps these endpoints to discoverable capabilities.

var builder = WebApplication.CreateBuilder(args);
var app = builder.Build();

app.MapGet("/api/invoices/{invoiceId}", (string invoiceId) =>
    Results.Ok(new {
        invoice_id = invoiceId, amount = 15420.00m,
        customer_id = "CUST-4821", customer_name = "Meridian Logistics",
        due_date = "2026-03-15", status = "open",
        line_items = new[] {
            new { description = "Freight Services Q1", amount = 12000.00m },
            new { description = "Handling Charges", amount = 3420.00m },
        }
    }));

app.MapPost("/api/receipts/search", (ReceiptSearchRequest req) =>
    Results.Ok(new { receipts = new[] {
        new { receipt_id = "RCP-99201", amount = 15405.00m,
              customer_ref = req.CustomerId, bank_date = "2026-03-10" }
    }}));

app.MapPost("/api/matching/three-way", (ThreeWayMatchRequest req) =>
{
    var variance = Math.Abs(req.InvoiceAmount - req.ReceiptAmount);
    var poVariance = Math.Abs(req.InvoiceAmount - req.POAmount);
    if (variance <= req.Tolerance && poVariance <= req.Tolerance)
        return Results.Ok(new { status = "matched", variance, within_tolerance = true });
    return Results.Ok(new { status = "variance",
        detail = $"Variance ${variance:F2} exceeds tolerance ${req.Tolerance:F2}",
        requires_review = true });
});

app.MapPost("/api/oracle/post-match", (PostMatchRequest req) =>
    Results.Ok(new { posting_id = $"POST-{DateTime.UtcNow:yyyyMMdd}-001",
        invoice_id = req.InvoiceId, status = "posted", timestamp = DateTime.UtcNow }));

app.MapPost("/api/notifications/variance-alert", (VarianceAlertRequest req) =>
    Results.Ok(new { alert_id = $"ALERT-{req.InvoiceId}",
        sent_to = req.AssignedTo, channel = "email+slack", status = "sent" }));

app.MapPost("/api/audit/generate-trail", (AuditTrailRequest req) =>
    Results.Ok(new { audit_id = $"AUDIT-{req.InvoiceId}",
        pdf_path = $"/audits/{req.InvoiceId}_trail.pdf", pages = 3 }));

app.MapGet("/health", () => Results.Ok(new { status = "healthy" }));
app.Run();

record ReceiptSearchRequest(string CustomerId, string DateRangeStart, string DateRangeEnd);
record ThreeWayMatchRequest(decimal InvoiceAmount, decimal ReceiptAmount, decimal POAmount, decimal Tolerance = 15.0m);
record PostMatchRequest(string InvoiceId, string ReceiptId, string MatchStatus);
record VarianceAlertRequest(string InvoiceId, string CustomerName, decimal VarianceAmount, string AssignedTo = "ar_clerk");
record AuditTrailRequest(string InvoiceId, int DecisionCount);
```

---

## Capability Manifest: capabilities/oracle-connector/manifest.yaml

```yaml
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
      Returns all matching receipts with amounts and remittance info.
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
      Posts a matched receipt against an invoice in Oracle Accounts Receivable.
      This is an IRREVERSIBLE financial transaction. Only invoke after a
      confirmed match with high confidence.
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
      Sends a variance notification to the assigned AR clerk for manual
      review when a match detects a discrepancy exceeding tolerance.
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
      Generates a PDF audit document recording the complete matching
      journey including all decisions, invocations, and outcomes.
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
```

---

## Agent System Prompt

```csharp
// GOAL INTENT ONLY. No capability names. No sequences. No workflow.
private const string SystemPrompt = """
    You are a financial reconciliation agent for Stallion Insurance.

    Your objective is to achieve a 3-way match between an invoice,
    a corresponding bank cash receipt, and a purchase order.

    RULES:
    - Gather all necessary data before attempting to match.
    - Never post a financial transaction unless a match is confirmed.
    - If any variance exceeds tolerance, escalate to a human reviewer.
    - Generate an audit trail after every completed or escalated match.
    - Use the available capabilities to accomplish your objective.
      Do not assume which capabilities exist — discover them from what is provided.

    Think step by step. Explain your reasoning at each step.
    """;
```

---

## Agent: Agents/AgentRunner.cs

```csharp
public class AgentRunner
{
    private readonly Kernel _kernel;
    private readonly CapabilityBridge _bridge;
    private readonly HITLGateway _hitl;
    private readonly ICapabilityDiscoveryClient _discovery;
    private readonly IStateClient _state;
    private readonly ILogger<AgentRunner> _logger;

    public async Task<AgentState> PursueGoalAsync(string goalDescription, CancellationToken ct)
    {
        var goalId = $"GOAL-{Guid.NewGuid():N}"[..16];
        var context = new InvocationContext {
            RequestId = Guid.NewGuid().ToString(), CapabilityId = "",
            CallerId = "stallion-agent", GoalId = goalId, TenantId = "stallion" };

        // 1. DISCOVER CAPABILITIES
        var count = await _bridge.DiscoverAndRegisterAsync(_kernel, goalDescription, context, ct: ct);

        // 2. REASONING LOOP
        var chatHistory = new ChatHistory();
        chatHistory.AddSystemMessage(SystemPrompt);
        chatHistory.AddUserMessage($"Goal: {goalDescription}");
        var chatService = _kernel.GetRequiredService<IChatCompletionService>();
        var settings = new OpenAIPromptExecutionSettings {
            FunctionChoiceBehavior = FunctionChoiceBehavior.Auto(), MaxTokens = 4096 };

        var state = new AgentState { GoalId = goalId, GoalDescription = goalDescription, Status = "running" };

        while (state.Status == "running" && state.Iteration < state.MaxIterations
               && state.TokensUsed < state.TokenBudget)
        {
            // a. CALL LLM — SK auto-handles tool calling via CapabilityBridge
            var response = await chatService.GetChatMessageContentAsync(chatHistory, settings, _kernel, ct);
            chatHistory.Add(response);

            // b. UPDATE STATE
            state = state with { Iteration = state.Iteration + 1 };

            // c. CHECK COMPLETION
            var text = response.Content ?? "";
            AgentDecision? decision = null;
            try { decision = JsonSerializer.Deserialize<AgentDecision>(text); } catch { }
            if (decision?.IsGoalComplete == true) { state = state with { Status = "completed" }; break; }

            // d. HITL GATING
            if (decision is not null)
            {
                var capabilities = await _discovery.DiscoverAsync(ct: ct);
                var target = capabilities.FirstOrDefault(c => c.CapabilityId == decision.SelectedCapability);
                var hitl = _hitl.Evaluate(decision, target);
                if (hitl.RequiresHumanReview) { state = state with { Status = "hitl_suspended" }; break; }
            }

            // e. CHECKPOINT
            await _state.SetAsync("goal", goalId, state, ct);
        }

        await _state.SetAsync("goal", goalId, state, ct);
        return state;
    }
}
```

---

## Key Principles

```
1. Capabilities are PURE .NET — standard ASP.NET, zero CapaFabric dependency.
2. Manifest YAML is the ONLY integration artifact.
3. Agent uses Semantic Kernel — no custom LLM plumbing.
4. CapabilityBridge converts discovered capabilities to KernelFunctions.
5. System prompt describes GOAL INTENT — never capability names or sequences.
6. Proxy handles ALL cross-cutting: auth, guardrails, tracing, routing.
7. Agent code is IDENTICAL in Level 0 (no proxy) and Level 5 (full governance).
8. Adding a new capability = new endpoint + manifest entry. Zero Agent changes.
9. The LLM discovers capabilities from their descriptions and reasons about
   the invocation path. New capabilities are automatically usable.
```
