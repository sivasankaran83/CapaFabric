# ADR-006: Zero Hardcoded Workflow Logic

**Status**: Accepted
**Date**: 2026-03-19
**Decision Makers**: CapaFabric Core Team

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
