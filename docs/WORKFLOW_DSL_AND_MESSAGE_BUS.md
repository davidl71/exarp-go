# Workflow DSL and message bus: Go libraries and future alignment

**Purpose:** Record decisions for (1) using Go libraries for workflow/pipeline execution and (2) aligning with the planned Redis + Asynq message bus. See also [.cursor/plans/redis-asynq-and-tui.plan.md](../.cursor/plans/redis-asynq-and-tui.plan.md) and [ALTERNATIVES_ITERATIVE_CI_SCRIPTS.md](ALTERNATIVES_ITERATIVE_CI_SCRIPTS.md).

---

## 1. Prefer Go libraries

- **Workflow/pipeline execution:** Implement in Go. No dependency on external workflow CLIs (e.g. Dagu as a separate binary). Use a small in-process runner that loads declarative YAML/JSON (e.g. `.exarp/workflows/*.yaml`) and executes steps by calling existing tool handlers (same pattern as `runDailyTask` in `automation_native.go`: dispatch by tool name to native handlers).
- **Message bus / queue:** Use the **already-planned** Go library: **[Asynq](https://github.com/hibiken/asynq)** (Redis-backed task queue, MIT). Plan: [.cursor/plans/redis-asynq-and-tui.plan.md](../.cursor/plans/redis-asynq-and-tui.plan.md). Todo2 tasks exist for producer, worker, config, TUI, optional dispatcher.
- **Not reinventing the wheel:** Asynq provides retries, scheduling, optional periodic enqueue, and observability. For “queue in SQLite” without Redis, the doc already references **neoq-sqlite** and **laqueue** as Go libraries; use one of those if we want a DB-backed queue later.

---

## 2. Future message bus (Redis + Asynq) — summary

From the Redis+Asynq plan and ALTERNATIVES doc:

| Component | Role |
|-----------|------|
| **Redis** | Broker: holds job payloads. |
| **Asynq** | Go library: producer enqueues jobs; workers pull and run handlers. Retries, priorities, optional periodic tasks. |
| **Job payloads** | `task_id` (run `task_execute`) or `{"trigger":"automation","action":"nightly"}` (run automation). |
| **Workers** | Claim in Todo2 → run exarp-go (task_execute or automation) → release. Optional: `exarp-go worker` or separate binary. |
| **Invariant** | Todo2 remains source of truth for tasks and locks; queue is transport only. |

No separate “message bus” product: the **message bus is Redis + Asynq** (Go library).

---

## 3. How workflow DSL fits the message bus

- **Project workflows** (e.g. `.exarp/workflows/daily.yaml`): Declarative list of steps (tool + action + params), optional parallel groups. Run by a **Go in-process runner** that resolves steps and calls the same native handlers used today in `runDailyTask` / `runParallelTasks`. No subprocess per step; no Dagu CLI.
- **Single-host:** Automation tool (or new `workflow run` command) loads project YAML and runs the runner in-process. Behavior matches current hardcoded daily/nightly but driven by project git.
- **With Redis+Asynq (future):**
  - **Option A — One job per workflow:** Producer enqueues one job: `{"trigger":"automation","workflow":"daily"}` (or `action":"daily"`). Worker runs the same in-process runner (loads `.exarp/workflows/daily.yaml` or built-in daily), so one worker executes the whole pipeline. Simple; good when the workflow is small or must run sequentially.
  - **Option B — One job per step (optional):** For maximum parallelism across workers, producer could enqueue one job per step (e.g. `{"trigger":"workflow_step","workflow":"daily","step_id":"analyze_alignment"}`). Worker runs that step only. Requires workflow runner to support “run step N” and possibly depend on prior steps (wave ordering or DAG). Defer until we need multi-worker parallelism within one workflow.
- **Compatibility:** The same workflow YAML and the same Go runner can be used (1) in-process from the automation tool and (2) inside an Asynq worker when the job payload says “run workflow daily.” No duplicate engine.

---

## 4. Recommended libraries (no reinvention)

| Need | Library | Notes |
|------|---------|--------|
| **Message bus / task queue** | **Asynq** (hibiken/asynq) | Planned; Redis-backed; producer + worker in Go. |
| **Workflow/pipeline runner** | **In-repo Go** | Small interpreter: load YAML → steps → dispatch to existing tool handlers (extend `runDailyTask`-style switch or a tool registry). No Dagu; no new process. |
| **Optional: queue without Redis** | **neoq-sqlite** or **laqueue** | SQLite-backed job queue (see ALTERNATIVES doc). Use if we want “queue in DB” and no Redis. |
| **Cron/scheduling (in-process)** | **robfig/cron** | If we add “run automation daily at 2am” inside the binary; already referenced in ALTERNATIVES. |

We are **not** introducing a generic workflow engine dependency (e.g. Dagu as library): the pipeline is a list of “call tool X with params,” which we already do in `automation_native.go`. The only new piece is loading that list from project YAML and iterating (sequential + parallel groups).

---

## 5. Implementation order (suggested)

1. **Workflow format and runner (Go):** Define a minimal YAML schema (e.g. `steps: [{ tool, action, params?, parallel?: false }]` or `parallel: [ step1, step2 ]; then: [ step3 ]`). Implement runner in `internal/tools/` or `internal/workflow/` that dispatches to existing handlers. Use it to drive built-in “daily” from a default YAML (or keep current hardcoded daily and add “if project has `.exarp/workflows/daily.yaml`, use that”).
2. **Project workflows in git:** Document that `.exarp/workflows/*.yaml` are project-specific; automation (or `exarp-go workflow run daily`) runs them from project root. No message bus required for this.
3. **Redis+Asynq (existing plan):** Implement M1–M5 from [redis-asynq-and-tui.plan.md](../.cursor/plans/redis-asynq-and-tui.plan.md). Worker handler for “automation” jobs runs the same workflow runner (in-process) for the requested workflow name.
4. **Optional:** Later, support “enqueue workflow” from automation or TUI (producer enqueues one job per workflow; worker runs workflow runner). No change to workflow YAML or runner.

---

## 6. Protobuf, hashing, and compression

Align workflow and message-bus work with existing protobuf usage and the context-reduction strategy ([CONTEXT_REDUCTION_OPTIONS.md](CONTEXT_REDUCTION_OPTIONS.md), [CONTEXT_REDUCTION_FOLLOWUP_TASKS.md](CONTEXT_REDUCTION_FOLLOWUP_TASKS.md)).

### 6.1 Protobuf

**Current use in exarp-go:** Tool requests/responses (e.g. `TaskWorkflowRequest`, `AutomationRequest`, `AutomationResponse`, `ReportRequest`), config (`.exarp/config.pb`), and report data (`ProjectOverviewData`) use protobuf for typed, compact serialization. Handlers parse proto first and fall back to JSON.

**For workflow and message bus:**

| Use case | Recommendation |
|----------|----------------|
| **Asynq job payloads** | Option A: Keep payloads as small JSON (`task_id`, `{"trigger":"automation","action":"daily"}`) for simplicity and debugging. Option B: Add a small proto (e.g. `WorkflowJobPayload` with `trigger`, `workflow_name`, `task_id`, `project_root`) and serialize to bytes for Asynq; worker deserializes proto. Proto gives typed fields, smaller size, and schema evolution. |
| **Workflow definition storage** | Primary: YAML in `.exarp/workflows/*.yaml` (human-editable, in git). Optional: compiled/canonical form as proto for cache or wire (e.g. `WorkflowDef` with repeated `Step`) so runner and producer share one schema; generate from YAML on load. |
| **Workflow run result** | Already covered by `AutomationResponse` (action + `result_json`). For structured result without one big JSON string, add `WorkflowRunResult` proto with `steps_run`, `succeeded`, `failed`, `summary` and use it when MCP/CLI expects proto. |

**Conclusion:** Use proto for job payloads if we want smaller, versioned messages in Redis; keep YAML as source of truth for workflow definitions. Extend `proto/tools.proto` (e.g. `WorkflowJobPayload`, `WorkflowRunResult`) when implementing the Asynq producer/worker.

### 6.2 Hashing

**Current use / plan:** `content_hash` (future) for task list and report to support client cache and skip re-fetch ([exarp-mcp-output.mdc](../.cursor/rules/exarp-mcp-output.mdc)). Project already has `github.com/cespare/xxhash/v2` (indirect).

**For workflow and message bus:**

| Use case | Recommendation |
|----------|----------------|
| **Workflow definition identity** | Hash the canonical form of the workflow (e.g. YAML bytes or serialized steps). Use as cache key (e.g. "skip re-run if workflow hash unchanged") or version id in job payload. Prefer xxhash for speed. |
| **Job idempotency / dedup** | Optionally include a hash of (workflow_name + params + project_root) in the job so Asynq or the worker can deduplicate (e.g. "already ran this workflow with these inputs"). |
| **Step output (future)** | If steps expose output for conditional logic, hash step output for "run step B only if step A output changed" (advanced; defer). |

**Conclusion:** Add workflow-definition hash when implementing the runner (for cache/skip or display). Optionally add payload hash for idempotent enqueue. Reuse xxhash; keep hashes short (e.g. first 8–16 chars of hex) for readability.

### 6.3 Compression

**Current use:** `compact=true` for JSON (no indentation) in session prime, task_workflow list, report ([response_compact.go](../internal/tools/response_compact.go)). Planned: gzip+base64 for large JSON payloads, optional DB compression for long text ([CONTEXT_REDUCTION_OPTIONS.md](CONTEXT_REDUCTION_OPTIONS.md) §§2–3, [CONTEXT_REDUCTION_FOLLOWUP_TASKS.md](CONTEXT_REDUCTION_FOLLOWUP_TASKS.md) items 2, 4).

**For workflow and message bus:**

| Use case | Recommendation |
|----------|----------------|
| **Workflow run result (MCP/CLI)** | Use `compact=true` when returning JSON (same as report/automation). For very large run results (many steps, big summaries), support `compress=true` (gzip+base64 envelope) as in context-reduction plan. |
| **Asynq job payloads** | Keep payloads small (task_id or short trigger JSON). If a job ever carries large data (e.g. full workflow YAML or big params), consider: (1) store in DB or file and put only an id/path in the job, or (2) gzip+base64 the payload before enqueue; worker decompresses first. Prefer (1) to avoid bloating Redis. |
| **Stored workflow definitions** | YAML on disk needs no compression. If we cache parsed workflow in DB or memory, optional gzip for large definitions (align with DB compression for `long_description` in followup tasks). |

**Conclusion:** Wire `compact` (and when implemented, `compress`) for workflow run results like other tools. Keep queue payloads small; avoid compressing job bodies unless payload size becomes a problem.

---

## 7. References

- [.cursor/plans/redis-asynq-and-tui.plan.md](../.cursor/plans/redis-asynq-and-tui.plan.md) — Redis + Asynq plan, M1–M5, Todo2 task names.
- [docs/ALTERNATIVES_ITERATIVE_CI_SCRIPTS.md](ALTERNATIVES_ITERATIVE_CI_SCRIPTS.md) — Message queue integration, Asynq/Machinery/gocraft/work, SQLite schema options, locking.
- [.cursor/rules/agent-locking.mdc](../.cursor/rules/agent-locking.mdc) — ClaimTaskForAgent, ReleaseTask, worker behavior.
- [internal/tools/automation_native.go](../internal/tools/automation_native.go) — `runDailyTask`, `runParallelTasks`, current step dispatch.
- [docs/CONTEXT_REDUCTION_OPTIONS.md](CONTEXT_REDUCTION_OPTIONS.md) — Compact JSON, content hash, gzip+base64, DB compression.
- [docs/CONTEXT_REDUCTION_FOLLOWUP_TASKS.md](CONTEXT_REDUCTION_FOLLOWUP_TASKS.md) — Follow-up tasks for hashing, compression, compact wiring.

---

## 8. Follow-up tasks (Todo2)

Create these with `exarp-go task create` (or MCP `task_workflow` action=create) when Todo2/DB is ready. Existing Redis+Asynq plan tasks (producer, worker, config, TUI) remain in [.cursor/plans/redis-asynq-and-tui.plan.md](../.cursor/plans/redis-asynq-and-tui.plan.md).

| # | Task name | Priority | Tags |
|---|-----------|----------|------|
| 1 | Workflow DSL: YAML schema and in-process Go runner | medium | workflow, dsl, feature |
| 2 | Project workflows in git: .exarp/workflows/*.yaml | medium | workflow, docs, feature |
| 3 | Workflow: protobuf and hashing (job payloads, content_hash) | low | workflow, protobuf, performance |
| 4 | Workflow run result: compact and compress wiring | low | workflow, context, performance |

**CLI commands (run from project root after `make build` or with `go run ./cmd/server`):**

```bash
exarp-go task create "Workflow DSL: YAML schema and in-process Go runner" \
  --description "Define minimal YAML schema (steps: tool, action, params; parallel groups). Implement runner that loads YAML and dispatches to existing tool handlers (runDailyTask pattern). See docs/WORKFLOW_DSL_AND_MESSAGE_BUS.md §5." \
  --priority medium --tags "workflow,dsl,feature"

exarp-go task create "Project workflows in git: .exarp/workflows/*.yaml" \
  --description "If .exarp/workflows/daily.yaml exists, automation uses it; else built-in daily. Document in README/docs. See docs/WORKFLOW_DSL_AND_MESSAGE_BUS.md §3, §5." \
  --priority medium --tags "workflow,docs,feature"

exarp-go task create "Workflow: protobuf and hashing (job payloads, content_hash)" \
  --description "Optional: WorkflowJobPayload/WorkflowRunResult in proto/tools.proto; workflow-definition hash (xxhash) for cache/dedup. See doc §6.1, §6.2." \
  --priority low --tags "workflow,protobuf,performance"

exarp-go task create "Workflow run result: compact and compress wiring" \
  --description "Wire compact=true (and compress=true when implemented) for workflow/automation run results. See doc §6.3, CONTEXT_REDUCTION_OPTIONS.md." \
  --priority low --tags "workflow,context,performance"
```
