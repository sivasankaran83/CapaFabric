# CapaFabric — Multi-Agent Orchestration Example

> Stallion Insurance: 4-stage invoice cash matching pipeline.
> Each stage is an independent agent with its own Agent + capabilities.
> A Supervisor Agent discovers and orchestrates them via CapaFabric.
> No special multi-agent framework. Agents are just capabilities to other agents.

---

## The Core Insight

```
In CapaFabric, there is NO distinction between a "capability" and an "agent."

An agent is a service that:
  1. Has a Agent (LLM-powered reasoning)
  2. Has capabilities (deterministic functions)
  3. Exposes a goal endpoint via manifest

To another agent, that goal endpoint IS a capability.
The Supervisor doesn't know whether "ingest_incoming_correspondence"
is a simple HTTP endpoint or a full Agent with its own LLM calls.

Multi-agent orchestration = capability discovery + invocation.
No special protocol. No agent-to-agent messaging. Just manifests and proxies.
```

---

## System Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                        SUPERVISOR GOALAGENT                               │
│                        (.NET Semantic Kernel)                             │
│                                                                           │
│  Goal: "Process all pending invoices for today"                          │
│                                                                           │
│  Discovers 4 sub-agents as capabilities via CapaFabric:                  │
│    • stallion.ingest.process_correspondence                              │
│    • stallion.extract.extract_structured_data                            │
│    • stallion.match.perform_invoice_reconciliation                       │
│    • stallion.control.adjudicate_and_finalize                            │
│                                                                           │
│  Reasons about WHICH to call and WHEN — no hardcoded sequence.           │
└───────────┬──────────────┬──────────────┬──────────────┬─────────────────┘
            │              │              │              │
            ▼              ▼              ▼              ▼
┌───────────────┐ ┌────────────────┐ ┌──────────────┐ ┌──────────────────┐
│ INGESTION     │ │ EXTRACTION     │ │ MATCHING     │ │ CONTROL          │
│ AGENT         │ │ AGENT          │ │ AGENT        │ │ AGENT            │
│ (.NET)        │ │ (.NET)         │ │ (Go)         │ │ (.NET)           │
│               │ │                │ │              │ │                  │
│ Agent:    │ │ Agent:     │ │ Agent:   │ │ Agent:       │
│  Reasons      │ │  Reasons about │ │  Reasons     │ │  Evaluates       │
│  about email  │ │  attachment    │ │  about       │ │  match quality,  │
│  relevance    │ │  format +      │ │  matching    │ │  decides post    │
│               │ │  extraction    │ │  strategy    │ │  vs escalate     │
│ Capabilities: │ │ Capabilities:  │ │ Capabilities:│ │ Capabilities:    │
│  fetch_emails │ │  extract_pdf   │ │  verify_     │ │  commit_posting  │
│  classify_msg │ │  extract_csv   │ │   three_way  │ │  escalate_       │
│  filter_spam  │ │  standardize   │ │  resolve_    │ │   variance       │
│               │ │                │ │   approx     │ │  produce_audit   │
└───────────────┘ └────────────────┘ └──────────────┘ └──────────────────┘
     :8081             :8082              :8083             :8084
```

---

## How Agents Expose Themselves as Capabilities

Each agent exposes a **goal endpoint** — a single entry point that accepts a goal
and returns a result. Internally, it may use its own Agent with multiple LLM
calls. Externally, it's just another capability.

```
┌──────────────────────────────────────────────────────────────────┐
│  INGESTION AGENT (internal view)                                  │
│                                                                    │
│  POST /api/goals/ingest ← This is what the manifest exposes      │
│       │                                                            │
│       ▼                                                            │
│  IngestionAgent.PursueGoalAsync(goal)                        │
│       ├──► LLM: "Which emails should I process?"                  │
│       ├──► Capability: fetch_emails(date_range)                   │
│       ├──► LLM: "Are any of these irrelevant?"                   │
│       ├──► Capability: classify_message(email_id)                 │
│       └──► Return: { processed_count: 12, email_ids: [...] }     │
│                                                                    │
│  To the Supervisor, this is ONE capability call:                  │
│    invoke("stallion.ingest.process_correspondence", {...})       │
│    → { processed_count: 12, email_ids: [...] }                   │
└──────────────────────────────────────────────────────────────────┘
```

---

## Agent Manifests

### Ingestion Agent

```yaml
apiVersion: capafabric/v1
kind: CapabilityManifest
metadata:
  agent_id: stallion-ingestion
  language: dotnet
app:
  port: 8081
  protocol: http
  health_path: /health
  base_path: /api
capabilities:
  # Goal endpoint — what the Supervisor calls
  - capability_id: stallion.ingest.process_correspondence
    name: process_correspondence
    description: >
      Processes incoming emails for invoice-related content. Uses AI
      reasoning to filter by velocity, content relevance, and date range.
      Returns processed email IDs with extracted invoice references.
      Use when new correspondence needs ingestion into the pipeline.
    tags: [email, ingestion, stage, agent]
    side_effects: true
    endpoint:
      method: POST
      path: /goals/ingest
      arguments:
        date_range_start: { in: body }
        date_range_end: { in: body }
        max_emails: { in: body, default: 100 }
      response:
        from: body
    security:
      required_roles: [pipeline_operator]
      classification: confidential

  # Internal capabilities (used by this agent's own Agent)
  - capability_id: stallion.ingest.fetch_emails
    name: fetch_emails
    description: >
      Fetches emails from the configured mailbox within a date range.
    tags: [email, read, internal]
    idempotent: true
    endpoint:
      method: POST
      path: /capabilities/fetch-emails
      arguments:
        date_range_start: { in: body }
        date_range_end: { in: body }
      response:
        from: body.emails

  - capability_id: stallion.ingest.determine_message_relevance
    name: determine_message_relevance
    description: >
      Classifies an email as invoice-related, payment-related, or irrelevant.
    tags: [email, classification, internal]
    idempotent: true
    endpoint:
      method: POST
      path: /capabilities/classify
      arguments:
        email_id: { in: body }
        subject: { in: body }
        body_preview: { in: body }
      response:
        from: body
```

### Extraction Agent

```yaml
apiVersion: capafabric/v1
kind: CapabilityManifest
metadata:
  agent_id: stallion-extraction
  language: dotnet
app:
  port: 8082
  protocol: http
  health_path: /health
  base_path: /api
capabilities:
  - capability_id: stallion.extract.extract_structured_data
    name: extract_structured_data
    description: >
      Extracts structured invoice data from email bodies and attachments.
      Handles PDF, CSV, Excel, and image formats using AI reasoning.
      Returns normalized invoice records. Use after correspondence
      ingestion to transform raw content into structured data.
    tags: [extraction, content, stage, agent]
    side_effects: true
    endpoint:
      method: POST
      path: /goals/extract
      arguments:
        email_ids: { in: body }
      response:
        from: body
    security:
      required_roles: [pipeline_operator, finance_analyst]

  - capability_id: stallion.extract.extract_from_pdf
    name: extract_from_pdf
    description: Extracts text and tables from a PDF attachment.
    tags: [pdf, parsing, internal]
    idempotent: true
    endpoint:
      method: POST
      path: /capabilities/parse-pdf
      arguments:
        attachment_id: { in: body }
      response:
        from: body

  - capability_id: stallion.extract.standardize_to_canonical_format
    name: standardize_to_canonical_format
    description: Normalizes extracted data into the standard invoice record format.
    tags: [normalization, internal]
    idempotent: true
    endpoint:
      method: POST
      path: /capabilities/normalize
      arguments:
        raw_records: { in: body }
      response:
        from: body
```

### Matching Agent (Go)

```yaml
apiVersion: capafabric/v1
kind: CapabilityManifest
metadata:
  agent_id: stallion-matching
  language: go
app:
  port: 8083
  protocol: http
  health_path: /health
  base_path: /api
capabilities:
  - capability_id: stallion.match.perform_invoice_reconciliation
    name: perform_invoice_reconciliation
    description: >
      Performs intelligent invoice-to-receipt reconciliation. Uses AI
      reasoning to select the best matching strategy based on data
      quality. Returns match results with confidence scores. Use after
      content extraction to match invoices against receipts.
    tags: [matching, reconciliation, stage, agent]
    side_effects: true
    endpoint:
      method: POST
      path: /goals/reconcile
      arguments:
        invoice_records: { in: body }
      response:
        from: body

  - capability_id: stallion.match.verify_three_way_match
    name: verify_three_way_match
    description: Deterministic 3-way match between invoice, receipt, and PO.
    tags: [matching, deterministic, internal]
    idempotent: true
    endpoint:
      method: POST
      path: /capabilities/three-way-match
      arguments:
        invoice_amount: { in: body }
        receipt_amount: { in: body }
        po_amount: { in: body }
        tolerance: { in: body, default: 15.0 }
      response:
        from: body

  - capability_id: stallion.match.resolve_approximate_match
    name: resolve_approximate_match
    description: Fuzzy matching for invoices where exact amounts don't align.
    tags: [matching, fuzzy, internal]
    idempotent: true
    endpoint:
      method: POST
      path: /capabilities/fuzzy-match
      arguments:
        invoice: { in: body }
        candidates: { in: body }
      response:
        from: body
```

### Control Agent (HITL + Posting)

```yaml
apiVersion: capafabric/v1
kind: CapabilityManifest
metadata:
  agent_id: stallion-control
  language: dotnet
app:
  port: 8084
  protocol: http
  health_path: /health
  base_path: /api
capabilities:
  - capability_id: stallion.control.adjudicate_and_finalize
    name: adjudicate_and_finalize
    description: >
      Reviews match results and decides the appropriate action: post
      to Oracle (high confidence), send variance alert (mismatch), or
      escalate to human reviewer (ambiguous). Generates audit trail for
      every decision. This is the final stage — use after matching.
    tags: [control, posting, hitl, stage, agent]
    requires_approval: true
    side_effects: true
    endpoint:
      method: POST
      path: /goals/review
      arguments:
        match_results: { in: body }
      response:
        from: body
    security:
      required_roles: [ar_manager, finance_admin]
      classification: restricted
      audit_level: forensic

  - capability_id: stallion.control.commit_financial_posting
    name: commit_financial_posting
    description: Posts matched receipt against invoice in Oracle AR. Irreversible.
    tags: [oracle, posting, critical, internal]
    requires_approval: true
    side_effects: true
    endpoint:
      method: POST
      path: /capabilities/post-to-oracle
      arguments:
        invoice_id: { in: body }
        receipt_id: { in: body }
      response:
        from: body
    security:
      required_roles: [ar_manager]
      classification: restricted
      audit_level: forensic

  - capability_id: stallion.control.escalate_variance_to_reviewer
    name: escalate_variance_to_reviewer
    description: Notifies AR clerk of a variance requiring manual review.
    tags: [notification, alert, internal]
    side_effects: true
    endpoint:
      method: POST
      path: /capabilities/variance-alert
      arguments:
        invoice_id: { in: body }
        variance_amount: { in: body }
      response:
        from: body

  - capability_id: stallion.control.produce_compliance_audit
    name: produce_compliance_audit
    description: Generates PDF audit document of the complete matching journey.
    tags: [audit, compliance, internal]
    side_effects: true
    endpoint:
      method: POST
      path: /capabilities/audit-trail
      arguments:
        invoice_id: { in: body }
        journey: { in: body }
      response:
        from: body
```

---

## Context Awareness Capability

```yaml
# Registered by a shared data service — prevents redundant stage invocation
- capability_id: stallion.context.assess_reconciliation_readiness
  name: assess_reconciliation_readiness
  description: >
    Evaluates the current processing status for a customer's invoices.
    Returns which invoices exist, their current stage (ingested, extracted,
    matched, posted, variance), and whether matching data is available.
    Use FIRST when a goal references a specific customer or invoice to
    understand what work has already been done and avoid redundant processing.
  tags: [context, status, read, pre-check]
  idempotent: true
  side_effects: false
  endpoint:
    method: POST
    path: /capabilities/customer-status
    arguments:
      customer_id: { in: body }
      period: { in: body }
    response:
      from: body
```

---

## Supervisor Agent

### System Prompt (Goal Intent Only)

```csharp
private const string SystemPrompt = """
    You are the pipeline supervisor for Stallion Insurance's
    invoice cash matching operation.

    Your responsibility is to orchestrate the end-to-end process
    of matching incoming invoices to bank cash receipts and ensuring
    all matches are posted, all variances are escalated, and all
    actions are audited.

    The pipeline involves:
    - Ingesting correspondence that may contain invoice-related content
    - Extracting structured data from various formats
    - Matching invoices against bank receipts and purchase orders
    - Reviewing matches and taking the appropriate action

    RULES:
    - If the goal references a specific customer or invoice, check
      what data already exists before deciding which stages to run.
    - Process stages in logical dependency order — you cannot match
      before data is extracted, and you cannot extract before content
      is ingested.
    - If any stage fails, assess the failure before deciding whether
      to retry, skip, or escalate.
    - Use the available capabilities to accomplish your objective.
      Do not assume which capabilities exist — discover them from
      what is provided.

    Think step by step. After each stage, evaluate the results
    before deciding what to do next.
    """;
```

---

## Runtime Flow: "Process all pending invoices for today"

```
1. Supervisor discovers 5 capabilities: assess_reconciliation_readiness,
   process_correspondence, extract_structured_data,
   perform_invoice_reconciliation, adjudicate_and_finalize

2. LLM: "This is a broad goal, no specific customer. Start with ingestion."
   → invoke: process_correspondence({ date_range: "2026-03-10" })
   → Ingestion Agent internally: fetch_emails → classify → returns 12 emails

3. LLM: "12 emails processed. Now extract structured data."
   → invoke: extract_structured_data({ email_ids: [...] })
   → Extraction Agent internally: parse_pdf × 5, parse_csv × 3, normalize

4. LLM: "Got 12 invoice records. Now reconcile."
   → invoke: perform_invoice_reconciliation({ invoice_records: [...] })
   → Matching Agent (Go) internally: three_way_match × 10, fuzzy_match × 2

5. LLM: "10 matched, 2 variance. Send to control."
   → invoke: adjudicate_and_finalize({ match_results: [...] })
   → Control Agent internally: post × 10 (HITL), alert × 2, audit trail

6. Supervisor: "PIPELINE COMPLETE"
```

## Runtime Flow: "Reconcile invoices of customer ABC for Feb 2026"

```
1. Supervisor discovers same 5 capabilities.

2. LLM: "Specific customer referenced. Check what already exists."
   → invoke: assess_reconciliation_readiness({ customer_id: "ABC", period: "2026-02" })
   → Result: INV-4821 extracted, INV-4822 extracted, INV-4830 ingested only

3. LLM: "Two are ready, one needs extraction. Skip ingestion entirely."
   → invoke: extract_structured_data({ email_ids: ["E030"] })  ← only the one

4. LLM: "All three ready. Reconcile."
   → invoke: perform_invoice_reconciliation({ invoice_records: [4821, 4822, 4830] })

5. LLM: "Results ready. Finalize."
   → invoke: adjudicate_and_finalize({ match_results: [...] })

6. Supervisor: "GOAL COMPLETE — skipped ingestion, extracted 1, reconciled 3"
```

**Same Supervisor. Same prompt. Same capabilities. Different goal → different path.**

---

## Key Multi-Agent Principles

```
1. AGENTS ARE CAPABILITIES — An agent's goal endpoint is just another manifest
   entry. The Supervisor discovers it via CapaFabric like any other capability.

2. NESTED REASONING — Each agent has its OWN Agent with its OWN LLM calls.
   The Supervisor delegates a goal and receives a result. How the agent achieves
   the goal is encapsulated.

3. SUPERVISOR PROMPT IS GOAL-ONLY — No agent names, no sequence, no hardcoding.
   Discovers available agents from descriptions. Reasons about dependency order.

4. CONTEXT AWARENESS — The assess_reconciliation_readiness capability lets the
   Supervisor check existing state before deciding which stages to invoke.
   Prevents redundant processing. The LLM reasons about this naturally.

5. MIXED LANGUAGES — Matching Agent is Go. Rest are .NET. Supervisor doesn't care.

6. INDEPENDENT SCALING — Matching (CPU-heavy) scales to 5 replicas.
   Ingestion (I/O-heavy) scales to 2. CP load balances across replicas.

7. INDEPENDENT DEPLOYMENT — Update Extraction to support new format?
   Deploy only that agent. Nothing else changes.

8. OBSERVABLE END-TO-END — OTEL trace: Supervisor → Agent → Agent's internal
   capabilities. One Jaeger trace captures the entire pipeline.

9. HITL AT ANY LEVEL — Control Agent's commit_financial_posting has
   requires_approval=true. HITL happens inside Control's Agent.
   Supervisor simply waits for the result.

10. ADDING A NEW AGENT — Deploy service + manifest. Supervisor's next
    /discover sees it. If its description matches a need, the LLM uses it.
    Zero Supervisor changes. Zero prompt changes.

11. NO HARDCODED WORKFLOW — The invocation path emerges from the LLM's
    reasoning over the goal + context + capability descriptions. Different
    goals produce different paths through the same agents.
```
