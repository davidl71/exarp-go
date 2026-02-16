# Parallel execution framework

**Purpose:** Define how exarp-go runs work in parallel (in-process tool batches and wave-based plan execution).

## Components

### 1. Wave-based execution (plan)

- **Source:** `BacklogExecutionOrder()` in `internal/tools/graph_helpers.go` computes dependency levels and waves (same wave = no ordering = can run in parallel). When config `thresholds.max_tasks_per_wave` is set (e.g. in `.exarp/config.yaml` or protobuf), `LimitWavesByMaxTasks()` splits any wave with more than that many tasks into consecutive waves so each wave has at most that many tasks.
- **Consumers:** `report` tool with `action=plan` (writes `.cursor/plans/exarp-go.plan.md` with `waves` in frontmatter) and `action=parallel_execution_plan` (writes `.cursor/plans/parallel-execution-subagents.plan.md` with one section per wave and task IDs); TUI waves view and run-wave.
- **Usage:** Run waves sequentially (Wave 0, then Wave 1, …). Within a wave, run one subagent (e.g. wave-task-runner) per task so they run in parallel. To cap tasks per wave, set `thresholds.max_tasks_per_wave` (e.g. `10`). See [.cursor/plans/parallel-execution-subagents.plan.md](../.cursor/plans/parallel-execution-subagents.plan.md).

### 2. In-process parallel tool runs (automation)

- **Source:** `runParallelTasks()` in `internal/tools/automation_native.go`. Runs a slice of `parallelTask` (tool name + params) with a max concurrency limit (`max_parallel_tasks`).
- **Consumer:** `automation` tool `action=daily` (and related) runs batches of tools in parallel (e.g. analyze_alignment, task_analysis, health).
- **Usage:** Pass `max_parallel_tasks` in automation params to cap concurrency.

### 3. Task grouping for parallelization (analysis)

- **Source:** `task_analysis` with `action=parallelization` or `action=dependencies_summary`; uses `findParallelizableTasks()` and dependency levels from `GetTaskLevels()` / `BacklogExecutionOrder()`.
- **Output:** Parallel groups (by dependency level) and execution order. Use to decide which Todo2 tasks can be run in parallel.

## Shared state management

State shared across agents and processes:

| State        | Location / API                    | Purpose                                      |
|-------------|------------------------------------|----------------------------------------------|
| **Todo2**   | `.todo2/todo2.db` (SQLite), JSON fallback | Single source of truth for tasks, status, deps |
| **Handoffs**| `.todo2/handoffs.json`            | Context from previous session (summary, next_steps) |
| **Task locks** | DB `task_agent_locks`           | Claim/release via `database.ClaimTaskForAgent`, `ReleaseTask`, `RenewLease`; prevents duplicate assignment |

**Usage:** Agents read/write Todo2 via `task_workflow` or database package; session prime and handoff tools read handoffs; claim before moving a task to In Progress and release when done. See [.cursor/rules/agent-locking.mdc](../.cursor/rules/agent-locking.mdc).

## Summary

| Layer            | Mechanism                    | Use case                          |
|------------------|-----------------------------|-----------------------------------|
| Plan / subagents | Waves + one agent per task  | Cursor subagents per wave         |
| Automation       | runParallelTasks + max N    | In-process parallel MCP tool runs |
| Analysis         | task_analysis parallelization | Inspect groups and order        |
| Shared state     | Todo2 + handoffs + task locks | Coordination across agents      |

## References

- [MULTI_AGENT_PLAN.md](MULTI_AGENT_PLAN.md) — Task dependency analysis and Phase 3 tasks
- [.cursor/plans/parallel-execution-subagents.plan.md](../.cursor/plans/parallel-execution-subagents.plan.md) — Wave prompts and task IDs
