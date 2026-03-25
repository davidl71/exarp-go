# Task Lanes and File Ownership Plan

**Date:** 2026-03-25  
**Status:** Planning  
**Purpose:** Make exarp-go parallel execution plans actionable by adding source-file ownership, collision detection, and lane-aware task planning.

---

## Problem

Current exarp-go planning is strong on task dependencies and waves, but weak on source ownership.

Today the system can answer:
- which tasks depend on which other tasks
- which tasks can run in the same wave
- which In Progress tasks overlap by dependency or declared files

It cannot reliably answer:
- which source files a task is likely to change
- which tasks will collide on the same high-churn files
- how to split work into stable ownership lanes before spawning subagents
- which files should be treated as merge hotspots and kept under one owner

This gap showed up clearly during recent multi-pane TUI and health work in Aether:
- the dependency graph was mostly flat
- many tasks were parallelizable in theory
- but several tasks converged on the same hotspot files such as `app.rs`, `input.rs`, and `ui/mod.rs`
- useful parallelization required manual lane planning by source-file ownership, not just dependency analysis

The result is that wave-based planning overstates safe parallelism.

---

## Goals

1. Add explicit source ownership to tasks when known.
2. Infer likely file ownership when tasks do not declare it.
3. Detect file-collision risk before parallel execution.
4. Emit lane-aware execution plans, not only dependency waves.
5. Preserve current Todo2 task model and graph analysis as the foundation.

Non-goals:
- perfect static prediction of every edited file
- replacing dependency analysis
- full semantic code ownership inference in the first cut

---

## Proposed model

### 1. Task ownership fields

Extend task metadata with optional ownership hints.

Suggested fields:
- `owned_files`: exact file paths this task is expected to modify
- `owned_globs`: glob patterns for broader ownership
- `forbidden_files`: files the task should avoid touching
- `ownership_confidence`: `explicit`, `inferred`, or `unknown`
- `lane`: optional stable lane label such as `tui-shell`, `backend-health`, `docs`, `source-architecture`

These should live in task metadata so they remain backward-compatible.

Example:

```json
{
  "lane": "tui-shell",
  "owned_files": [
    "agents/backend/services/tui_service/src/input.rs",
    "agents/backend/services/tui_service/src/app.rs",
    "agents/backend/services/tui_service/src/ui/mod.rs"
  ],
  "forbidden_files": [
    "agents/backend/services/tui_service/src/ui/settings.rs"
  ],
  "ownership_confidence": "explicit"
}
```

### 2. Merge hotspot registry

Add a lightweight project-level concept of hotspot files.

Suggested sources:
- repeated overlap among active tasks
- repeated edits across recent commits
- manual configuration in project config or plan files

Example hotspot output:
- `agents/backend/services/tui_service/src/app.rs`
- `agents/backend/services/tui_service/src/input.rs`
- `agents/backend/services/tui_service/src/ui/mod.rs`

Planner behavior:
- avoid placing two hotspot-sharing tasks in the same parallel batch
- surface a stronger warning when tasks touch hotspot files

### 3. Ownership inference

When `owned_files` are missing, infer likely write scope from:
- explicit file lists in task descriptions
- paths mentioned in comments or research notes
- recent git history for similar task titles or tags
- `rg` matches on symbols or module names from the task title
- tag-to-path heuristics

Inference should be best-effort and low-risk.

Suggested confidence levels:
- `explicit`: user or tool declared exact ownership
- `high`: task content strongly names files or modules
- `medium`: tag/path heuristics plus code matches
- `low`: weak text similarity only
- `unknown`: no usable signal

---

## Planning outputs

### 1. Collision-aware execution plan

Extend `task_analysis action=execution_plan` and related reports with:
- `likely_files`
- `hotspot_files`
- `collision_risk`
- `lane`
- `parallel_safe_with`
- `parallel_conflicts_with`

Example shape:

```json
{
  "task_id": "T-123",
  "lane": "tui-shell",
  "likely_files": [
    "services/tui_service/src/input.rs",
    "services/tui_service/src/app.rs"
  ],
  "hotspot_files": [
    "services/tui_service/src/app.rs"
  ],
  "collision_risk": "high",
  "parallel_conflicts_with": ["T-124", "T-130"],
  "parallel_safe_with": ["T-140", "T-141"]
}
```

### 2. Lane summary

Execution plans should include lane groupings alongside waves.

Example:
- `lane=tui-shell`: owns shell routing, focus, key dispatch
- `lane=tui-pane-ops`: owns Alerts/Logs/Settings local renderers
- `lane=backend-health`: owns health aggregation and transport metrics
- `lane=docs`: owns documentation-only tasks

This gives a coordinator enough structure to assign multiple workers without manual repo archaeology.

### 3. File-lease aware run claims

Optional next step: allow runs to claim files or globs in addition to task IDs.

Example:
- task run claims `agents/backend/services/tui_service/src/input.rs`
- a second run touching the same file warns or is blocked unless forced

This would extend current task locking into practical merge protection.

---

## Tooling changes

### `task_workflow`

Add support for:
- storing/updating `owned_files`, `owned_globs`, `forbidden_files`, `lane`
- showing ownership in `show` and `list`
- optionally adding ownership when creating tasks

Suggested actions:
- reuse `create` and `update`
- add `sub_action=ownership` only if needed later

### `task_analysis`

Add or extend:
- `action=execution_plan` to emit collision-aware ownership data
- `action=conflicts` to report file collisions for all candidate tasks, not only In Progress tasks
- `action=lanes` to suggest ownership lanes across a backlog slice
- `action=infer_ownership` to populate metadata heuristically

### `report`

Enhance plan outputs with:
- lane sections
- hotspot warnings
- explicit notes on files that should remain single-owner

---

## Suggested implementation phases

### Phase 1: Metadata and reporting

Low-risk, high-value.

1. Add optional metadata fields for ownership and lane.
2. Surface them in `task_workflow list/show`.
3. Extend conflict detection to consider exact-file ownership metadata.
4. Add lane and ownership sections to execution plan output.

This phase does not require inference.

### Phase 2: Ownership inference

1. Add heuristics from task text and tags.
2. Emit `ownership_confidence`.
3. Allow dry-run ownership suggestions.
4. Keep explicit ownership higher priority than inferred ownership.

### Phase 3: File-lease and hotspot policy

1. Track hotspot files.
2. Add warnings when starting runs on hotspot overlaps.
3. Optionally add file leases for active runs.

---

## Low-hanging fruit

These items would have helped immediately in the Aether work:

1. Add `owned_files` and `lane` metadata to tasks.
2. Extend `execution_plan` to show file collisions between candidate parallel tasks.
3. Add a hotspot report for files repeatedly touched by active tasks.
4. Teach session prime to warn when the suggested next tasks collide on the same files.

These are much cheaper than full ownership inference and would already improve parallel execution quality substantially.

---

## Recommended lane taxonomy

A small stable vocabulary is enough for most repos.

Suggested built-in lane labels:
- `shell`
- `ui-pane`
- `backend-health`
- `backend-runtime`
- `source-architecture`
- `docs`
- `config`
- `testing`
- `cleanup`

Projects can override or extend these with repo-specific tags.

---

## Boundary and ownership principles

The planner should encourage these rules:
- one owner for shell/composition files
- local pane or module files may be parallelized
- transport/model/aggregation layers should remain separate lanes where architecture already separates them
- docs-only tasks should be isolated from runtime code lanes when possible
- dependency waves are necessary but not sufficient for safe parallel execution

In practice, the most useful distinction is often:
- dependency-safe
- merge-safe

exarp-go already models the first well. This plan focuses on the second.

---

## Recommendation

Implement Phase 1 first.

If exarp-go gains only two new capabilities soon, they should be:
1. explicit per-task ownership metadata (`owned_files`, `owned_globs`, `lane`)
2. collision-aware `execution_plan` output

That would make the generated plan usable for real multi-agent coordination without requiring a human to manually map tasks to source files every time.
