# Tool Consolidation Analysis

**Date:** 2026-03-10  
**Status:** Draft execution map for the current tool surface

## Summary

Current surface:
- `38` base tools
- optional `39th` tool on Apple Silicon: `apple_foundation_models`

Conclusion:
- We should keep the current capability set.
- We should not keep all of it as first-class top-level user-facing tools.
- Best path is consolidation by user intent, with compatibility aliases for at least one release.

## Keep As First-Class Tools

These are coherent user-facing entry points and should remain top-level:

| Tool | Reason |
|---|---|
| `task_workflow` | Core CRUD/lifecycle surface for tasks |
| `task_analysis` | Core backlog analysis/planning surface |
| `task_discovery` | Distinct discovery workflow from code/docs |
| `report` | Clear reporting/scorecard/plan surface |
| `health` | Clear health/status surface |
| `session` | Distinct session/handoff surface |
| `automation` | Scheduled/routine orchestration surface |
| `testing` | Clear test execution/coverage surface |
| `lint` | Clear lint/fix surface |
| `security` | Clear security scan/report surface |
| `git_tools` | Clear git/history/diff surface |
| `text_generate` | Unified text-generation entry point |
| `tool_catalog` | Discovery and migration helper for users |
| `workflow_mode` | Useful UX-level workflow control |
| `generate_config` | Clear standalone config-generation surface |
| `setup_hooks` | Clear standalone environment/bootstrap surface |
| `memory` | Primary memory interaction surface |
| `memory_maint` | Distinct maintenance/cleanup surface |
| `recommend` | Stable recommendation/advisor selector surface |

## Keep, But Treat As Specialist Tools

These are valid tools, but they should be documented as specialist/advanced tools rather than part of the primary surface:

| Tool | Reason |
|---|---|
| `ollama` | Backend-specific |
| `cursor_cloud_agent` | Platform/integration-specific |
| `fm_plan_and_execute` | Advanced orchestration helper |
| `prompt_tracking` | Useful, but narrow |
| `check_attribution` | Release/compliance specialist flow |
| `add_external_tool_hints` | Internal/tooling specialist flow |
| `analyze_alignment` | Valid, but likely secondary to task/report surfaces |

## Convert To Compatibility Aliases

These should continue to exist short term, but conceptually belong inside broader tools:

| Current Tool | Preferred Surface | Recommendation |
|---|---|---|
| `scan_dependency_security` | `security(action="scan")` | Keep as alias only |
| `context_budget` | `context(action="budget")` | Keep as alias only |
| `infer_task_progress` | `task_analysis` or `task_workflow` | Keep as alias now, fold later |
| `infer_session_mode` | `session` | Keep as alias now, fold later |
| `task_execute` | `task_workflow(action="run_with_ai")` or a future `task_workflow(action="execute")` | Keep as alias now, fold later |

## Strong Merge Candidates

These tools are the best refactor targets if the goal is to reduce top-level surface area without losing capability:

### 1. `scan_dependency_security` -> `security`
- Current state: already an alias in behavior.
- Action: keep name as compatibility alias, remove from primary docs/catalog, route users to `security`.

### 2. `context_budget` -> `context`
- Current state: semantically a sub-action, not a separate user intent.
- Action: document `context` as primary, keep `context_budget` only for compatibility and scripts.

### 3. `infer_session_mode` -> `session`
- Current state: session-adjacent implementation detail.
- Action: fold into `session(action="prime"|...)` or `session(sub_action="infer_mode")`.

### 4. `infer_task_progress` -> `task_analysis`
- Current state: task-backlog inference/analysis, not a separate product concept.
- Action: move toward `task_analysis(action="infer_progress")` or `task_workflow(action="sync_progress")`.

### 5. `task_execute` -> `task_workflow`
- Current state: task lifecycle execution helper, not really a separate domain.
- Action: merge into task workflow once the action contract is clear.

### 6. `fm_plan_and_execute` -> `text_generate`
- Current state: advanced planning/execution variant of generation/orchestration.
- Action: consider folding into `text_generate(provider="auto", task_type="plan_execute")` only if UX stays clear.

## Do Not Merge Yet

These may look mergeable on count alone, but should stay separate for now:

| Tool | Why not merge yet |
|---|---|
| `report` and `health` | Different user intents: reporting vs operational checks |
| `task_analysis` and `task_workflow` | Analysis and mutation should stay distinct |
| `automation` and `report` | Automation orchestrates many tools; report is a leaf surface |
| `memory` and `memory_maint` | User interaction vs maintenance/cleanup |
| `testing`, `lint`, `security` | Separate operational domains with different expectations |
| `read_resource` and `list_resources` | Protocol-facing primitives; should stay available even if downplayed |

## Proposed Primary Surface

If we optimize for user-facing simplicity, the recommended primary tool list is:

- `task_workflow`
- `task_analysis`
- `task_discovery`
- `report`
- `health`
- `session`
- `automation`
- `testing`
- `lint`
- `security`
- `git_tools`
- `memory`
- `memory_maint`
- `recommend`
- `text_generate`
- `workflow_mode`
- `tool_catalog`
- `generate_config`
- `setup_hooks`

Everything else should be documented as:
- compatibility alias
- backend-specific specialist tool
- protocol primitive

## Recommended Phases

### Phase 1: Documentation-Only Consolidation
- Mark primary tools in docs and tool catalog.
- Mark alias tools as compatibility surfaces.
- Mark backend-specific tools as advanced.

### Phase 2: Alias Formalization
- Ensure these are explicitly described as aliases:
  - `scan_dependency_security`
  - `context_budget`
- Add tool-catalog migration guidance for:
  - `infer_task_progress`
  - `infer_session_mode`
  - `task_execute`

### Phase 3: Action Merges
- Add equivalent actions to the destination tools.
- Keep old names as thin wrappers for one release.
- Remove old names only after usage/docs migration is complete.

## Recommended Concrete Map

| Tool | Disposition |
|---|---|
| `analyze_alignment` | Keep specialist |
| `generate_config` | Keep primary |
| `health` | Keep primary |
| `setup_hooks` | Keep primary |
| `check_attribution` | Keep specialist |
| `add_external_tool_hints` | Keep specialist |
| `memory` | Keep primary |
| `memory_maint` | Keep primary |
| `report` | Keep primary |
| `security` | Keep primary |
| `scan_dependency_security` | Alias to `security` |
| `task_analysis` | Keep primary |
| `task_discovery` | Keep primary |
| `task_workflow` | Keep primary |
| `infer_task_progress` | Merge target: `task_analysis` or `task_workflow`; keep alias now |
| `testing` | Keep primary |
| `automation` | Keep primary |
| `tool_catalog` | Keep primary |
| `workflow_mode` | Keep primary |
| `lint` | Keep primary |
| `estimation` | Keep specialist |
| `git_tools` | Keep primary |
| `session` | Keep primary |
| `infer_session_mode` | Merge target: `session`; keep alias now |
| `ollama` | Keep specialist |
| `context_budget` | Alias to `context` |
| `context` | Keep specialist |
| `text_generate` | Keep primary |
| `task_execute` | Merge target: `task_workflow`; keep alias now |
| `prompt_tracking` | Keep specialist |
| `recommend` | Keep primary |
| `research_aggregator` | Keep specialist; reassess later |
| `cursor_cloud_agent` | Keep specialist |
| `fm_plan_and_execute` | Keep specialist; possible later merge into `text_generate` |
| `read_resource` | Keep protocol primitive |
| `list_resources` | Keep protocol primitive |
| `apple_foundation_models` | Keep specialist, conditional registration |

## Recommendation

Recommended immediate action:
1. Do not remove tools yet.
2. Reclassify the surface in docs and `tool_catalog`.
3. Treat `scan_dependency_security` and `context_budget` as compatibility aliases immediately.
4. Plan real merges for `infer_task_progress`, `infer_session_mode`, and `task_execute`.

This gets you a smaller effective user-facing surface without taking on unnecessary compatibility risk.
