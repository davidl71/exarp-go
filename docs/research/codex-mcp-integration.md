# Codex MCP Integration Research

**Date:** 2026-03-24  
**Status:** Research / Backlog

---

## Overview

`exarp-go` already works well as an MCP server for Codex, but most of the current integration is still generic:

- Codex can call tools
- Codex can read resources
- Codex can follow skill documents
- `session action=prime` provides a useful bootstrap payload

That is enough for basic task execution, but it does not yet give Codex a strongly opinionated execution-cockpit surface. The next improvement is not deeper access to Codex internals. The next improvement is better MCP output design:

- more role-aware
- more task-scoped
- more execution-oriented
- more explicit about safe next actions

This document focuses on what `exarp-go` can realistically do for Codex through MCP.

---

## Current Capability

Today `exarp-go` can already guide Codex through four channels.

### 1. Tools

Primary tools:

- `task_workflow`
- `session`
- `report`
- `health`
- `task_analysis`
- `task_execute`

These are the strongest integration point because they let Codex mutate state, claim tasks, record execution runs, and inspect project status with structured results.

### 2. Resources

Primary resources:

- `prime://context`
- `stdio://tasks/{task_id}`
- `stdio://active-work`
- `stdio://tools`
- `stdio://agent/skills`
- `stdio://agent/skills/{name}`
- `stdio://agent/briefing`
- `stdio://agent/task/{task_id}/execution-pack`
- `stdio://agent/alerts`

This is the best path for low-token context loading. It is especially relevant now that `session prime` returns `lazy_context` pointers for task-specific skill and task resources.

### 3. Skills and Rules

Repo-local skills shape Codex behavior without requiring tool calls:

- `use-exarp-tools`
- `task-workflow`
- `session-handoff`
- `thinking-workflow`
- `database-maintenance`

This layer is effective, but only when the guidance stays current with the actual MCP surface.

### 4. Structured Session Bootstrap

`session action=prime` already provides:

- `suggested_next`
- `suggested_next_action`
- `active_claims`
- `active_runs`
- `lazy_context`
- continuity ledger injection
- status context and handoff alerting

This is the closest thing to a Codex-specific “control plane” today.

---

## Hard Limits

There are important limits to what MCP can do for Codex.

`exarp-go` cannot:

- inspect or control Codex hidden reasoning
- force Codex to delegate work internally
- become a fully bidirectional runtime for streaming intermediate reasoning
- guarantee that Codex will follow one suggested workflow over another

So the correct design target is not “talk to Codex internals.” The target is:

- make the best next action obvious
- make the safe action sequence explicit
- make the task context cheap to load
- make execution evidence easy to append

---

## Gaps

### 1. No Codex-specific execution packet

Codex often needs the same bundle of context when starting work:

- task body
- dependencies
- recommended tools
- skill URIs
- recent execution evidence
- active conflicts
- related files

Today this is spread across multiple calls (`task show`, `session prime`, `read_resource`, `report execution_briefing`).

### 2. No explicit next-step contract

`recommended_tools` exists, but it is still just a flat list. Codex would benefit from an ordered workflow such as:

1. `task_workflow claim`
2. `task_workflow start_run`
3. `report execution_briefing`
4. `task_workflow add_progress`
5. `task_workflow verify`
6. `task_workflow end_run`

That is much closer to how a coding agent actually works.

### 3. No dedicated Codex briefing surface

`session prime` is general-purpose. `execution_briefing` is execution-oriented. Neither is explicitly shaped as “what Codex should do next in this repo, in this session, with this task.”

### 4. Weak event-style awareness

Codex can poll resources, but there is no compact event/alert surface for:

- stale locks
- approval-needed items
- blocked active runs
- review-ready tasks
- changed recommended next task

### 5. Multi-agent guidance is still descriptive, not operational

The project now has:

- `agent_role`
- role-aware session inference
- execution runs
- active claims
- orchestration lanes in execution briefing

But the MCP surface still does not expose a concrete “delegate this next” or “spawn this role next” packet tuned for Codex.

---

## Recommended Additions

### A. `stdio://agent/briefing`

Purpose:

- one compact resource for Codex startup and task switching

Suggested contents:

- current status context
- suggested next task
- active claims / runs
- delegation suggestions
- continuity ledger excerpt
- top recommended tools/resources

Why it helps:

- reduces “which call should I make first?” overhead
- gives Codex one obvious bootstrap resource
- `stdio://codex/briefing` can remain as a compatibility alias

### B. `stdio://agent/task/{task_id}/execution-pack`

Purpose:

- one task-scoped packet for starting work

Suggested contents:

- task metadata
- dependencies
- recommended tools
- ordered workflow steps
- skill URIs
- related active run / latest progress / latest verification
- active conflicts or blockers

Why it helps:

- avoids repetitive multi-call context assembly
- matches how Codex actually starts a task
- `stdio://codex/task/{task_id}/execution-pack` can remain as a compatibility alias

### C. Ordered execution contracts

Add metadata such as:

- `recommended_workflow`
- `safe_next_actions`
- `required_preconditions`
- `completion_evidence_required`

This should extend `recommended_tools`, not replace it.

### D. Compact execution alerts resource

Example:

- `stdio://agent/alerts`
- optional alias: `stdio://codex/alerts`

Suggested contents:

- stale locks
- long-running runs
- review-needed tasks
- approval queue count
- tasks claimed by another host

Why it helps:

- gives Codex an interrupt-like signal without requiring many queries

### E. Stronger multi-agent delegation summaries

Current orchestration data should be extended with:

- which lane is idle
- which lane has queued work
- which specific tasks are best delegated next
- recommended role for the next child task or subagent

This is still MCP-friendly and does not depend on hidden model APIs.

---

## Recommended Sequencing

### 1. Codex briefing resource

Highest leverage, lowest coordination cost.

### 2. Task execution pack resource

Most practical improvement for day-to-day coding loops.

### 3. Ordered workflow metadata

Improves correctness and reduces tool-call drift.

### 4. Alert/event resource

Useful once the main execution packets exist.

### 5. Deeper multi-agent delegation packet

Best added after the execution-cockpit surfaces are stable.

---

## Suggested Task Set

1. Add `stdio://agent/briefing` resource for Codex-oriented startup context
2. Add `stdio://agent/task/{task_id}/execution-pack` resource
3. Extend task/session/report payloads with ordered execution workflow metadata
4. Add compact `stdio://agent/alerts` resource for stale/blocked/review-needed execution state

These are intentionally MCP-surface tasks, not generic “AI integration” tasks.

---

## Bottom Line

`exarp-go` can already help Codex in advanced ways, but the right strategy is not deeper access to Codex internals.

The right strategy is to make `exarp-go` a better execution-control plane for Codex by exposing:

- better startup briefing
- better task execution packets
- better ordered action guidance
- better compact alerting

That is realistic, MCP-native, and directly actionable in this codebase.
