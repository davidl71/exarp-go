# Task Tool Enrichment Design

**Created:** 2026-02-15  
**Status:** Draft  
**Summary:** Enrich Todo2 tasks with recommended tools (e.g. Tractatus Thinking, Context7) so the AI and users see per-task tool suggestions.

---

## User Guide: recommended_tools and Enrichment

Tasks can have a `recommended_tools` metadata key: an array of MCP tool identifiers (e.g. `["tractatus_thinking", "context7"]`). When present:

- **Session prime** — Each item in `suggested_next` includes `recommended_tools` so the AI knows which tools to use for that task.
- **Task show** — The task detail payload includes `recommended_tools`.
- **Tag-based enrichment** — Tags like `#research`, `#docs` can map to tools (e.g. `#research` → `tractatus_thinking`); an enrichment tool or `task_analysis` action applies these rules to backlog tasks.

**Canonical tool identifiers:** `tractatus_thinking`, `context7`, `task_workflow`, `task_analysis`.

**Manual assignment:** Use `task_workflow` update/create with `recommended_tools` in params, or (when implemented) `exarp-go task update T-123 --recommended-tools tractatus_thinking,context7`.

See `.cursor/rules/mcp-configuration.mdc` for Tractatus and Context7 setup.

---

## Overview

Today, exarp-go provides **global** tool hints at session prime (e.g. “Consider tractatus_thinking for logical decomposition”). It does **not** attach recommended tools to individual tasks. This design adds **task-level tool enrichment**: store and surface “recommended tools” (e.g. `tractatus_thinking`, `context7`) so that when a task is suggested or shown, the AI sees which tools are relevant for that task.

---

## Goals

1. **Per-task tool hints** – Each task can have zero or more recommended tools (MCP tool names or identifiers).
2. **Session prime** – When `suggested_next` includes a task, include its `recommended_tools` (or `tool_hints`) so the AI can use the right tool for that task.
3. **Task show / list** – Expose recommended tools in task details and, optionally, in list output.
4. **Enrichment mechanisms** – Support tag-based rules and, optionally, semantic (LLM) or manual assignment.

---

## Current State

| Mechanism | Scope | Tractatus |
|-----------|--------|-----------|
| Session prime `hints` | Global (all sessions) | Yes, in `getHintsForMode()` |
| `add_external_tool_hints` | Docs/markdown files | No; enriches files, not tasks |
| Task metadata | Flexible JSON blob | No `recommended_tools` today |

Task metadata already supports arbitrary keys (e.g. `preferred_backend`, `planning_doc`). We add a conventional key and consumption points.

---

## Design

### 1. Task metadata key

- **Key:** `recommended_tools`
- **Type:** `[]string` (array of tool identifiers)
- **Examples:** `["tractatus_thinking"]`, `["context7", "tractatus_thinking"]`
- **Storage:** Existing task metadata (DB and JSON); no schema migration required. Ensure `SanitizeMetadataForWrite` accepts `[]interface{}` for this key (slice of strings is already JSON-serializable).

Optional: add `tool_hints` as `[]string` for human-readable one-liners (e.g. “Use tractatus operation=start for decomposition”). For MVP, `recommended_tools` is sufficient; clients can map tool ID → hint text.

### 2. Canonical tool identifiers

Use stable names that match MCP tool names or a small well-known set:

| Identifier | MCP / usage |
|------------|-------------|
| `tractatus_thinking` | Tractatus Thinking MCP – logical decomposition |
| `context7` | Context7 MCP – library/docs lookup |
| `task_workflow` | exarp-go task_workflow |
| `task_analysis` | exarp-go task_analysis |
| (others) | Extend as needed |

### 3. Enrichment: how tasks get `recommended_tools`

**Option A – Tag-based rules (MVP)**  
- Config or code: map tags → tools (e.g. `#research` → `["tractatus_thinking"]`, `#docs` → `["context7"]`).  
- New tool or task_analysis action: `enrich_tool_hints` (or `discover_tool_hints`).  
- For each Todo/In Progress task, if tags match a rule, set or append `recommended_tools` in metadata (merge with existing).  
- Idempotent; can run on sync or on demand.

**Option B – Semantic (LLM)**  
- Use task title + description to suggest tools (e.g. via existing text_generate or recommend flow).  
- Write suggested list into metadata (with optional human review).  
- Defer to later phase.

**Option C – Manual / CLI**  
- `task_workflow` update: accept `recommended_tools` in params and write to task metadata.  
- CLI: `exarp-go task update T-123 --recommended-tools tractatus_thinking,context7`.  
- Useful for one-off and overrides.

MVP: implement **Option A** (tag-based) and **Option C** (manual/CLI update). Option B can follow.

### 4. Consumption

**Session prime**  
- When building `suggested_next`, load full task (or at least metadata) for each suggested task.  
- For each item, add `recommended_tools: task.Metadata["recommended_tools"]` (or `tool_hints` derived from it).  
- Prime response already includes task `id`, `content`, `priority`, `level`; add `recommended_tools` so the AI sees e.g. “For T-xxx use: tractatus_thinking”.

**Task show**  
- Include `recommended_tools` (and optional `tool_hints`) in the task detail payload.

**Task list**  
- Optional: include `recommended_tools` in list entries when a flag or param requests “include tool hints”. Default: list stays minimal; show/details carry full metadata.

### 5. Tag → tool mapping (MVP)

Suggested defaults (config or code):

| Tag | Recommended tools |
|-----|-------------------|
| `#research`, `#design`, `#planning` | `tractatus_thinking` |
| `#docs` | `context7` |
| `#task_workflow`, `#task_management` | `task_workflow`, `task_analysis` |

Config file (e.g. in centralized config or `.cursor/task_tool_rules.yaml`) can override or extend.

### 6. Files / components

| Component | Change |
|-----------|--------|
| `internal/tools/session.go` | In `getSuggestedNextTasksFromTasks` (or caller), attach `recommended_tools` from task metadata to each suggested task in the response. |
| `internal/tools/task_workflow_common.go` | Support `recommended_tools` in update/create params; write into `task.Metadata["recommended_tools"]`. |
| `internal/cli/task.go` | Optional: `--recommended-tools` flag for `task update` / `task create`. |
| New or existing tool | `enrich_tool_hints` (or task_analysis action): tag-based rule engine, read tasks, merge `recommended_tools` into metadata, save. |
| Config | Optional: `task_tool_rules` (tag → tools) in centralized config or separate file. |
| Docs | Update MCP/config docs to describe `recommended_tools` and optional `tool_hints`. |

### 7. Backward compatibility

- Tasks without `recommended_tools` behave as today (no hints).  
- Existing metadata keys unchanged.  
- Session prime and task show remain backward compatible; new field is additive.

---

## Tasks (implementation order)

1. **Design doc** – This document (done).
2. **Metadata + task_workflow** – Add support for `recommended_tools` in task metadata; task_workflow update (and create) accepts and persists it.
3. **Session prime** – Include `recommended_tools` in each `suggested_next` item when present.
4. **Task show** – Include `recommended_tools` in task detail response.
5. **Tag-based enrichment** – Implement `enrich_tool_hints` (or task_analysis action) with tag → tool rules; run on backlog tasks.
6. **CLI** – Optional `--recommended-tools` for `task update` / `task create`.
7. **Config** – Optional config for tag → tool mapping.
8. **Docs** – Document `recommended_tools` and enrichment in user-facing docs.

---

## References

- Session prime hints: `internal/tools/session.go` (`getHintsForMode`, suggested_next).
- Task metadata: `internal/database/tasks.go` (SerializeTaskMetadata, DeserializeTaskMetadata, SanitizeMetadataForWrite).
- MCP config (Tractatus, Context7): `.cursor/rules/mcp-configuration.mdc`, `.cursor/skills/tractatus-decompose/SKILL.md`.
