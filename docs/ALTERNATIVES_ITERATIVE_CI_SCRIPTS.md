# Alternatives to Iterative Scripts / CI (Jenkins, Shell, CMake)

**Purpose:** Options for systems that **continuously run**, **take tasks from waves or assignments**, and **execute** them — beyond Jenkins, shell scripts, and CMake.  
**Last updated:** 2026-02-17

---

## What you’re replacing or complementing

- **Iterative scripts** — cron, shell loops, Make targets that run on a schedule or trigger.
- **CI** — Jenkins, GitHub Actions, etc., that run on push/schedule.
- **Task source** — In exarp-go: **waves** (dependency-ordered groups from `BacklogExecutionOrder`), **assignments** (Todo2 status + agent locking), and **Todo2 backlog** as the single source of truth.

**exarp-go already provides:**

- **Waves** — `report action=plan` / `parallel_execution_plan` writes `.cursor/plans/...plan.md` with wave→task IDs; TUI “waves” view; within a wave, tasks can run in parallel (e.g. one agent per task).
- **Execution** — `task_execute` runs one Todo2 task (load → prompt → model → apply changes); `automation` runs daily/nightly/sprint/discover (fixed tool sequences).
- **State** — Todo2 (SQLite), task locks (`ClaimTaskForAgent`), handoffs.

So any “runner” only needs to: (1) run on a schedule or continuously, (2) get task IDs from waves/backlog/assignment, (3) call `exarp-go -tool task_execute -args '{"task_id":"T-xxx"}'` or trigger `automation` / Make targets.

**Why language and license matter:** exarp-go is **Go** and **open source**. A runner in **Go** (or **language-agnostic** YAML/shell) avoids adding another runtime (e.g. Python) just to orchestrate. **Open source** (or clear source-available terms) avoids vendor lock-in and lets you self-host or audit the stack.

---

## Do these tools obsolete exarp-go?

**No.** They **complement** exarp-go; they don’t replace its core functions. The division of responsibility is:

| exarp-go function | Could an external tool replace it? | Why not / nuance |
|-------------------|------------------------------------|-------------------|
| **Todo2 backlog** (tasks, deps, status, tags, comments) | ❌ No | External tools don’t define your *project* backlog or task model; they run *units of work* you give them. Todo2 is the source of truth; orchestrators consume task IDs from exarp-go or from a plan derived from it. |
| **Waves / execution order** (`BacklogExecutionOrder`, plan, parallel-execution plan) | ❌ No | Wave computation is domain logic (dependency DAG over Todo2). Orchestrators don’t know your tasks or deps; they run steps you define. exarp-go (or a script calling it) produces the ordered list; the external tool runs that list. |
| **task_execute** (load task → prompt → LLM → parse → apply changes → comment) | ❌ No | This is the *semantics* of “execute one task” for your stack (Todo2 + model-assisted workflow). External tools only *invoke* it (e.g. one job or activity = one `task_execute` call). They don’t implement it. |
| **task locking** (`ClaimTaskForAgent`, release, renew) | ❌ No | Prevents double assignment across agents/hosts. Orchestrators assign “run T-123” to one worker; they don’t manage Todo2 state or exarp-go’s lock table. Locking stays in exarp-go (or your shared Todo2 DB). |
| **Report / scorecard / health / task_analysis / session / memory / etc.** | ❌ No | These are exarp-go MCP tools and CLI tools. Orchestrators may *call* them (e.g. a step that runs `report overview`); they don’t replace their logic. |
| **automation** (daily / nightly / sprint / discover) | ⚠️ Only the *trigger* | *Content* (which tools, in which order, with which params) is fixed in exarp-go. What an external tool can replace is **who triggers** that run: e.g. cron today runs “exarp-go automation nightly”; Temporal or a scheduled pipeline could run the same command on a schedule. So: “scheduler of automation” can move outside exarp-go; “what automation does” does not. |
| **In-process parallelism** (`runParallelTasks` inside automation) | ⚠️ Optional superset | exarp-go runs several tools in parallel inside one process. An external orchestrator could run many `task_execute` (or other tool) calls *across* hosts instead. That’s a *stronger* form of distribution, not a replacement: you still need exarp-go to define and execute each unit of work. |

**Summary:** External tools handle **when** to run, **where** (which host), **how many in parallel**, retries, and observability. exarp-go remains the place that defines **what** the backlog is (Todo2), **what** the execution order is (waves), **what** “execute one task” means (`task_execute`), and **what** the automation steps are (automation tool content). Nothing in the comparison table obsoletes Todo2, waves, task_execute, locking, or the MCP tool set; at most they take over the role of cron or a single-host script as the *trigger* for exarp-go.

---

## Comparison table (all options)

| System | **Language** | **Open source** | Run model | Task source | Execution | **Distributed / multi-host** | Complexity | exarp-go fit |
|--------|---------------|------------------|-----------|-------------|-----------|-----------------------------|------------|--------------|
| **Buildkite** | Go (agent) | ✅ Agent MIT; service proprietary (SaaS) | Agents **pull** jobs from queue | Pipeline gets work (exarp-go plan/API), fan-out jobs | One job per task; script runs `task_execute` | ✅ **Yes** — many agents on many hosts, each pulls next job; queues by cluster/tags | Medium | Strong: one job per task_id, agents anywhere |
| **Harness Open Source** | Go / YAML | ✅ Apache 2.0 (open source CI offering) | Pipelines on event/schedule | Schedule → pipeline runs exarp-go to get work | Steps run exarp-go | ⚠️ **Limited** — pipeline runs on one runner; scale by more pipelines, not by distributing one run | Low–Medium | Good for scheduled automation |
| **GitLab CI** | Ruby (backend), YAML pipelines | ✅ GitLab CE: MIT | Runners pull jobs | Schedule/API → pipeline | Jobs run exarp-go | ✅ **Yes** — register many runners (any host); jobs distributed across runners | Medium | Good: distributed runners, one job per task |
| **Tekton** | Go | ✅ Apache 2.0 | Pipelines as Pods; Triggers/Cron | Cron/event → PipelineRun; step gets task list | One step per task | ✅ **Yes** — K8s cluster = many nodes; Pods scheduled across nodes | Medium–High | Good on K8s: fan-out steps across cluster |
| **Argo Workflows** | Go | ✅ Apache 2.0 (CNCF) | CronWorkflow / Workflow DAG | Cron/event → Workflow; first step gets waves | One step per task (container) | ✅ **Yes** — Workflow steps run as Pods across K8s nodes; parallelism = multiple Pods | Medium–High | Strong on K8s: DAG + multi-node |
| **Temporal** | Go (server); SDKs: Go, Java, Python, etc. | ✅ MIT | Workers **pull** from Task Queues | Workflow gets waves; schedules activities | Activity = `task_execute` | ✅ **Yes** — workers on any number of hosts; same task queue, scale by adding workers | Medium–High | Strong: durable, retries, multi-host workers; Go SDK fits exarp-go |
| **Prefect** | Python | ✅ Apache 2.0 | Workers pull flow runs | Schedule/API → flow gets work | Flow run calls exarp-go | ✅ **Yes** — multiple workers on multiple hosts; work pool / queue | Medium | Good if Python OK (adds Python runtime) |
| **Kestra** | Java (server), YAML workflows | ✅ Apache 2.0 | Triggers → workflow runs on server | Schedule/webhook; task gets task IDs | Tasks run script/container | ⚠️ **Depends** — workers can be separate (Kestra supports remote workers); single-server default | Low–Medium | Good for YAML + schedule; JVM on server |
| **Airflow** | Python | ✅ Apache 2.0 | Scheduler + workers | DAG trigger; task “get_waves” | Task = operator → exarp-go | ✅ **Yes** — CeleryExecutor/KubernetesExecutor: workers on many hosts or Pods | High | Good: DAG + distributed executors; Python runtime |
| **Dagster** | Python | ✅ Apache 2.0 | Scheduler + run launcher + workers | Schedule/sensor → job | Op runs exarp-go | ✅ **Yes** — run launcher can be K8s/ECS; multiple run workers | Medium–High | Good: ops as units, multi-host via launcher; Python |
| **Redis + Celery/Bull/Sidekiq** | C (Redis); workers: Python/Node/Ruby | ✅ Redis BSD; Celery Apache 2.0; Bull MIT; Sidekiq LGPL | Workers **pull** from queue | Producer pushes task_ids (from exarp-go) | Worker runs `task_execute` | ✅ **Yes** — classic: many worker processes on many hosts, one shared queue | Low–Medium | Strong: add hosts = add workers; pick worker lang to match stack |
| **K8s Jobs + CronJob** | N/A (K8s = Go); your script any | ✅ Kubernetes Apache 2.0 | CronJob creates Jobs | Dispatcher gets waves, creates Job per task | Job Pod runs `task_execute` | ✅ **Yes** — Jobs scheduled across cluster nodes; scale with cluster size | Medium | Good on K8s: one Job per task, any node; no extra lang |
| **Earthly** | Go | ✅ MPL 2.0 | Targets in containers | Generated from plan | One target per task | ⚠️ **Limited** — can run remote builders; not designed as multi-host task queue | Medium | OK: containerized, single-runner focus; Go-native |
| **Task / Just** | Go (Task), Rust (Just) | ✅ Task: MIT; Just: MIT | One command (e.g. run-wave) | Script reads plan or exarp-go | Script loops `task_execute` | ❌ **No** — single process, one host | Low | Minimal: one host, one command; no extra runtime |
| **cron / systemd / launchd** | C (OS) | ✅ Part of OS (POSIX/systemd/BSD) | Schedule one command | None (or wrapper script) | One process | ❌ **No** — one host per cron; replicate cron on each host for “distribution” | Low | Minimal: automation on schedule; no extra stack |
| **Shell script** (cron or loop) | Shell | ✅ N/A (POSIX shell) | Cron runs script, or script loops | Script calls exarp-go (plan/waves) or fixed list | Script runs `exarp-go -tool task_execute` / `automation` | ❌ **No** — single host; multi-host = run same script on each + exarp-go locking | Low | Same as cron: minimal; you implement waves loop and error handling |

### Quick filter: language and open source

- **Go or language-agnostic (no extra runtime):** Buildkite (agent), Harness Open Source, Tekton, Argo Workflows, Temporal (server + Go SDK), K8s Jobs, Earthly, Task, cron, **shell script**. *Fits exarp-go’s Go stack without adding Python/JVM; shell is the minimal option.*
- **Python-centric:** Prefect, Airflow, Dagster, Celery. *Adds Python; good if you already run Python tooling.*
- **Fully open source (permissive):** All table entries marked ✅ are OSS or source-available; Buildkite’s *agent* is MIT (hosted service is proprietary). Harness Open Source and GitLab CE are OSS; Temporal, Argo, Tekton, Kestra, Prefect, Airflow, Dagster, Redis/Celery/Bull/Sidekiq, Earthly, Task, Just are open source with the licenses noted.

---

## Shell script vs alternatives

A **shell script** (plus cron or a loop) can do the same *triggering* as many of the tools above: run `exarp-go -tool automation -args '{"action":"nightly"}'` or loop over task IDs and call `task_execute`. Comparison:

| Aspect | Shell script (+ cron / loop) | Alternatives (CI, workflow engines, Go libs) |
|--------|------------------------------|---------------------------------------------|
| **Setup** | Minimal: no extra services, no new runtime. | Varies: cron-level (robfig/cron in-process) to full platform (Temporal, Buildkite). |
| **Distribution** | Single host unless you run the same script on many hosts and rely on exarp-go locking. No built-in load balancing. | Multi-host by design: agents, workers, or Pods across machines; one queue or scheduler. |
| **Retries / observability** | Manual: log output, retry in script or cron. | Built-in in most: retry policies, dashboards, logs, metrics (e.g. Asynq, Temporal, Buildkite). |
| **Waves / ordering** | You implement: script calls exarp-go to get plan/waves, then loops over task IDs (or wave-by-wave). | Some give you DAG/ordering (Argo, Temporal); others still need your script to produce the list, then they run it. |
| **Failure handling** | One failing `task_execute` can stop the loop unless you add error handling. | Queues and workflows typically retry or isolate failures so other tasks still run. |
| **When to use shell** | One host, simple schedule (e.g. nightly automation), low ops. Good default before adding infra. | Multiple hosts, retries/visibility, or “run many task_executes in parallel” across machines. |

**Summary:** Shell + cron is enough for “run exarp-go automation on a schedule” or a single-host “get waves → loop task_execute.” Use an alternative when you need distribution, retries, or observability without hand-rolling them.

---

## Go frameworks and libraries (same space)

These are **Go-native** options for scheduling or task queues; they fit exarp-go’s language and can run in the same process or as separate workers that invoke exarp-go.

| Project | Purpose | License | exarp-go usage |
|---------|---------|---------|----------------|
| **[Asynq](https://github.com/hibiken/asynq)** | Redis-backed task queue: scheduling, retries, priorities, periodic tasks, web UI, metrics. | MIT | Producer enqueues `task_id` (from exarp-go plan); worker handler runs `exec.Command("exarp-go", "-tool", "task_execute", "-args", ...)`. Multi-host: run workers on many machines. **Planned:** [.cursor/plans/redis-asynq-and-tui.plan.md](../.cursor/plans/redis-asynq-and-tui.plan.md) — producer, worker, Todo2 locking, TUI updates. |
| **[Machinery](https://github.com/RichardKnop/machinery)** | Celery-like distributed task queue; Redis/RabbitMQ/SQS/etc.; chains, groups, chords. | MIT | Same pattern: enqueue “run task T-xxx”; worker runs exarp-go. Heavier; use when you need multiple brokers or complex workflows. |
| **[gocraft/work](https://github.com/gocraft/work)** | Lightweight Redis job queue (Sidekiq-like); cron-like scheduling, retries, unique jobs. | MIT | Lighter than Asynq; enqueue job with `task_id`, worker runs exarp-go. Good for simple, high-throughput “run many task_executes.” |
| **[robfig/cron](https://github.com/robfig/cron)** | In-process cron scheduler (no queue). | MIT | `c.AddFunc("0 2 * * *", func() { ... run exarp-go automation nightly ... })`. No distribution; replaces shell cron with Go, same host. |
| **Temporal Go SDK** | Durable workflows and activities (client to Temporal server). | MIT | Activity = call exarp-go `task_execute`; workflow iterates waves and starts activities. Use when you want Temporal’s guarantees and multi-host workers. |

**Comparison to shell:** Asynq, Machinery, and gocraft/work add a **queue** and **workers** (and usually Redis), so you get retries, distribution, and visibility without writing that in shell. robfig/cron is “cron in Go”—no queue, single process, similar to a shell cron entry.

**Comparison to external CI/orchestrators:** Go libraries run inside your Go process (or as a separate Go binary); no separate platform (Buildkite, Argo, etc.). Use them when you want to keep everything in Go and optionally embed the runner in the same binary as exarp-go (e.g. a “wave runner” mode that uses Asynq to distribute `task_execute`).

---

## Message queue integration

**Yes — a message queue can help** when you want decoupling, buffering, scaling, or async triggers without tying exarp-go to a specific orchestrator. Many of the options above already use one under the hood (Redis for Asynq/Celery/Bull, RabbitMQ for Machinery/Celery, etc.). You can also integrate a queue directly.

### When a queue helps

| Benefit | Without queue | With queue |
|---------|----------------|------------|
| **Decoupling** | Trigger (cron, CI, webhook) must run or invoke exarp-go directly; tight coupling. | Producer (cron, CI, webhook) pushes a message (e.g. `{"task_id":"T-123"}` or `{"action":"automation","kind":"nightly"}`); workers pull and run exarp-go. Trigger and workers can be on different hosts and deployed independently. |
| **Buffering** | If no worker is ready, work is lost or the trigger blocks/fails. | Messages sit in the queue until a worker consumes them; no work lost if workers are temporarily down. |
| **Scaling** | Single process or manual multi-host (e.g. duplicate cron + locking). | Add more worker processes/hosts that pull from the same queue; natural load distribution. |
| **Async** | “Run backlog” blocks until done or times out. | Trigger enqueues N messages (one per task_id) and returns; workers process in the background. |
| **Ordering (optional)** | Script enforces wave order by looping. | Use separate queues per wave (e.g. `wave-0`, `wave-1`) or a single queue with priority; workers drain in order. |

### What the queue carries

The queue holds **work items**, not task content:

- **Per-task:** `{"task_id":"T-xxx"}` (or just `"T-xxx"`). Worker runs `exarp-go -tool task_execute -args '{"task_id":"T-xxx"}'`. Task body, deps, and status stay in Todo2.
- **Automation:** `{"trigger":"automation","action":"nightly"}`. Worker runs `exarp-go -tool automation -args '{"action":"nightly"}'`.

Wave ordering is either: (1) producer pushes to `wave-0` then `wave-1` and workers respect that, or (2) producer pushes all task_ids and workers use exarp-go locking so only one runs each task (ordering then comes from when producer enqueues, or from a separate “wave” field and worker logic).

### Raw queue vs task-queue library

| Approach | Pros | Cons |
|----------|------|------|
| **Raw queue** (Redis Streams, RabbitMQ, Kafka): your code publishes/consumes messages. | Minimal dependency; you control format and semantics. | You implement retries, dead-letter, scheduling, and visibility yourself. |
| **Task-queue library** (Asynq, Machinery, gocraft/work, Celery): library uses Redis/RabbitMQ and adds jobs, retries, scheduling, UI. | Retries, backoff, periodic tasks, and (often) a dashboard out of the box. | Extra dependency and abstraction. |

**Recommendation:** For exarp-go, a **task-queue library** (e.g. Asynq with Redis) usually gives the best tradeoff: queue semantics plus retries and observability without building them from scratch. Use a **raw queue** if you already have a broker and want the smallest integration (e.g. one producer and one consumer that only forward to exarp-go).

### What the queue does not replace

Same as the rest of the doc: the queue does **not** replace Todo2 (backlog and state), wave computation (exarp-go or your script), or `task_execute` logic. It only **transports** “what to run” (task_id or automation trigger). exarp-go (or a shared Todo2 store) remains the source of truth for task content and locking.

---

## SQLite schema extensions that can help

Todo2 already uses SQLite (`.todo2/todo2.db`) with `tasks`, `task_tags`, `task_dependencies`, locking columns (`assignee`, `lock_until`), and migrations. Extending the **schema** (new columns or tables) can give you queue-like behavior or scheduling **inside the same DB**, so you don’t need a separate Redis or message broker for basic cases.

### 1. Schema-only extensions (no loadable .so)

These are new migrations (columns or tables) on the existing Todo2 DB. No SQLite loadable extension required.

| Extension | Purpose | How it helps |
|-----------|---------|--------------|
| **`execution_level` / `wave` on `tasks`** | Store the computed dependency level or wave (0, 1, 2, …) when you run `BacklogExecutionOrder` or plan. | Workers can `SELECT id FROM tasks WHERE status IN ('Todo','In Progress') AND execution_level = ? ORDER BY id` to get “next wave” without recomputing. Optional: backfill from exarp-go when plan is generated. |
| **`scheduled_run_at` (INTEGER)** | Optional “run after” timestamp (Unix) on `tasks`. | “Scheduled” tasks only appear for workers when `strftime('%s','now') >= scheduled_run_at`. Lets you push work into the DB and have it become visible later. |
| **`execution_queue` table** | A separate table: `(id, task_id, status, enqueued_at, run_after, worker_id, started_at, retries)`. Producer inserts rows; workers `SELECT ... WHERE status = 'pending' AND run_after <= ? ... LIMIT 1` and `UPDATE ... SET status = 'running', worker_id = ?`. | Turns SQLite into the queue: one row per “run task T-xxx”. Same pattern as [SQLite background job systems](https://jasongorman.uk/writing/sqlite-background-job-system/). Use transactions and `version`/optimistic lock to avoid double-claim. |
| **Indexes** | e.g. `(status, execution_level)`, `(status, run_after)` or queue `(status, run_after)`. | Fast “next work” queries for workers. |

**Tradeoff:** SQLite has limited concurrent writers (one write at a time per DB). For many workers on many hosts, a single shared SQLite file can become a bottleneck; use WAL and short transactions. For one or a few workers, or for “queue in DB” on a single host, schema-only queue is simple and avoids Redis.

### 2. Loadable SQLite extensions (optional)

| Extension | Purpose | Relevance to exarp-go |
|-----------|---------|------------------------|
| **JSON1** | Built-in in modern SQLite; `json_extract(metadata,'$.key')`. | Already available; useful for querying/filtering on `tasks.metadata` (e.g. by tag or custom fields) without app code. |
| **crsqlite** (vlcn) | Multi-master replication: merge changes across copies of the DB. | Helps if each host has its **own** copy of Todo2 and you sync later (e.g. offline agents). Not needed if all workers share one DB (NFS/PVC) or one API. |
| **No generic “queue” extension** | Queue semantics are usually implemented as tables (job table + status), not a loadable extension. | Use schema (execution_queue or task columns) plus your code or a Go library that implements the job table pattern. |

### 3. Go libraries that add a job table to SQLite

These use SQLite as the queue backend by creating their own **job table** (and optionally cron/scheduling). You can point them at the **same** `.todo2/todo2.db` (adding a `jobs` or `queue` table) or a separate DB.

| Library | License | How it fits |
|---------|---------|-------------|
| **[neoq-sqlite](https://github.com/pranavmodx/neoq-sqlite)** | MIT | Queue-agnostic Go job lib with SQLite backend; retries, cron, future scheduling. Worker handler runs `exec` to exarp-go `task_execute`; producer enqueues `task_id`. Can use same DB as Todo2 (adds its tables). |
| **[laqueue](https://github.com/nicotsx/laqueue)** | — | Lightweight Go SQLite queue; delayed execution, retries. Same idea: add queue table, worker runs exarp-go. |

**Benefit:** You get retries, scheduling, and “queue in DB” without designing the table and claim logic yourself; the library owns the schema it needs. exarp-go stays the executor; the library only holds “what to run” (e.g. task_id payload).

### Summary

- **Schema-only:** Add `execution_level`/`wave` and/or `scheduled_run_at` to `tasks`, or an `execution_queue` table, so workers can poll the **same** Todo2 DB for “next work” without Redis. Best for single-host or low-write multi-host.
- **Loadable:** JSON1 is already there; crsqlite only if you need multi-master sync of Todo2. No queue-in-a-box extension.
- **Go + SQLite queue lib:** Use neoq-sqlite or laqueue to add a job table (same or separate DB); workers run exarp-go. Easiest way to get “queue in SQLite” with retries and scheduling.

See [SQLITE_SCHEMA_DESIGN.md](SQLITE_SCHEMA_DESIGN.md) and `migrations/` for the current Todo2 schema; any new columns or tables should go in a new migration.

---

## Locking

Locking ensures a task is only executed by one agent at a time and that “who is doing what” is visible everywhere (TUI, task_workflow, report). exarp-go already provides **task-level locking in the Todo2 DB**; the question is how it fits with queues, schema extensions, and distributed runners.

### What exarp-go already has

| Mechanism | Where | Purpose |
|-----------|--------|---------|
| **`tasks.assignee`** | Todo2 DB (migration 002) | Which agent/host has claimed the task. |
| **`tasks.lock_until`** | Todo2 DB | Lease expiry (Unix timestamp); allows stale lock cleanup. |
| **`ClaimTaskForAgent(ctx, taskID, agentID, leaseDuration)`** | `internal/database` | Atomically claim a task if unclaimed or expired; sets status to In Progress, assignee, lock_until. Returns error if already locked by another agent. |
| **`ReleaseTask(ctx, taskID, agentID)`** | `internal/database` | Release the lock (clear assignee/lock_until); only the owning agent can release. |
| **`RenewLease(ctx, taskID, agentID, duration)`** | `internal/database` | Extend lock_until for long-running work. |
| **`GetLockStatus(ctx, taskID)`** | `internal/database` | Check if task is locked, by whom, and whether the lock is expired or stale. |
| **`DetectStaleLocks` / `CleanupExpiredLocksWithReport`** | `internal/database` | Find and optionally clear expired locks (e.g. after agent crash). |

So the **single source of truth for “who runs this task”** is the Todo2 DB. Any runner (shell, CI job, queue worker) that calls `task_execute` should **claim first** (via task_workflow or database API) so the UI and other tools stay consistent. See [.cursor/rules/agent-locking.mdc](../.cursor/rules/agent-locking.mdc) and [AGENT_LOCKING_STRATEGY.md](AGENT_LOCKING_STRATEGY.md).

### Two layers of “locking”

| Layer | Who provides it | What it guarantees |
|-------|------------------|--------------------|
| **Queue / orchestrator** | Redis, Temporal, Buildkite, K8s Job, etc. | “This unit of work (job/message/activity) is assigned to this worker.” At most one consumer per job. |
| **Todo2 (exarp-go)** | `tasks.assignee` + `lock_until` + Claim/Release | “This task is In Progress and claimed by this agent.” Prevents two runners from both running `task_execute` for the same task_id and keeps status/lock visible in exarp-go tools. |

You can use **both**: the queue assigns “run T-123” to one worker; that worker still **claims T-123 in Todo2** before calling `task_execute`, so the TUI and task_workflow show the task as In Progress and locked. If you skip the Todo2 claim and only rely on the queue, the queue prevents double execution but Todo2 status may not reflect “In Progress” until task_execute updates it (and you lose assignee/lock_until visibility and stale-lock cleanup).

### When to use Todo2 locking

| Scenario | Use Todo2 claim? | Reason |
|----------|------------------|--------|
| **Queue guarantees one job per task_id** (e.g. one Redis job per T-xxx) | **Recommended** | Claim in Todo2 before `task_execute` so status and assignee are correct everywhere; release (or let lease expire) when done. |
| **Workers poll Todo2 directly** (e.g. “give me a task”) | **Required** | No queue to prevent double assignment; claim is the only way to reserve the task. Use `ClaimTaskForAgent` then `task_execute` then `ReleaseTask`. |
| **SQLite `execution_queue` table** | **Two options** | (1) Atomic claim on the queue row (e.g. `UPDATE execution_queue SET status='running', worker_id=? WHERE id=? AND status='pending'`) so only one worker gets the row; then optionally claim in Todo2 as well for UI consistency. (2) Or use only Todo2: workers `SELECT` from queue without updating, then `ClaimTaskForAgent(task_id)`; only the one who wins the claim runs `task_execute`. |
| **Single host, single process** (cron + script) | Optional | No concurrent workers; locking still helps if you later add another host or process. |

### SQLite and concurrent claim

SQLite does not support `SELECT ... FOR UPDATE`. To avoid two workers claiming the same task or the same queue row:

- **Todo2 claim:** `ClaimTaskForAgent` uses a transaction: read task, check assignee/lock_until, then update in the same transaction. Only one writer wins; the other gets “already locked” or a version conflict.
- **execution_queue table:** Use a single transaction: `UPDATE execution_queue SET status='running', worker_id=?, started_at=? WHERE status='pending' AND (run_after IS NULL OR run_after <= ?) ORDER BY enqueued_at LIMIT 1 RETURNING id, task_id` (or equivalent: SELECT then UPDATE with a unique constraint / version so only one commit succeeds). Keep transactions short to reduce contention.

### Stale locks and cleanup

If a worker crashes after claiming but before releasing, the task stays “In Progress” with an assignee and lock_until. exarp-go’s **stale lock detection** (`DetectStaleLocks`) and **cleanup** (`CleanupExpiredLocksWithReport`) use `lock_until` to find and clear expired locks so the task can be claimed again. Run cleanup periodically (e.g. in automation nightly or a cron job). See [TASKS_AS_CI_AND_AUTOMATION.md](TASKS_AS_CI_AND_AUTOMATION.md) for nightly automation that reports stale lock counts.

### Summary

- **Locking lives in Todo2** (assignee, lock_until, Claim/Release/Renew); use it whenever multiple runners can touch the same task so only one runs it and tools show correct state.
- **Queue/orchestrator** can additionally ensure one job per task_id; still **claim in Todo2** before `task_execute` for consistency.
- **Schema extensions** (e.g. execution_queue): add atomic claim semantics on the queue row and/or rely on Todo2 claim after dequeue; use short transactions to avoid SQLite writer contention.
- **Stale locks:** use `lock_until` and periodic cleanup so crashed workers don’t leave tasks permanently locked.

---

## Using Context7 for deeper comparison

**Context7** (MCP for library/docs lookup) can help when you need API details or design patterns for these options:

- **Go task queues:** Use `resolve-library-id` with a query like “asynq” or “go task queue,” then `query-docs` to get API for enqueueing, workers, periodic tasks, and retries. Same for “machinery” or “gocraft work.”
- **Go cron:** `resolve-library-id` for “robfig cron” (or “go cron”), then `query-docs` for scheduling API and timezone support.
- **Workflow engines:** For Temporal, Prefect, or Kestra, resolve the library and query for “worker,” “activity,” “schedule,” or “trigger” to compare how they run and observe tasks.

That gives you a precise, doc-backed comparison (e.g. Asynq vs Machinery vs shell) without relying only on high-level summaries. See [RESEARCH_HELPERS_REFERENCE.md](RESEARCH_HELPERS_REFERENCE.md) and the Context7 MCP docs for `resolve-library-id` and `query-docs` usage.

---

## Distributed execution among multiple hosts

### What “distributed” means here

- **Multiple hosts** run work at the same time (not just one machine running steps sequentially).
- **Task/wave source** is shared (Todo2, plan file, or exarp-go API) so all hosts see the same backlog.
- **No double execution** of the same task: use exarp-go’s **task locking** (`ClaimTaskForAgent`) or the runner’s own assignment (e.g. “one job per task_id” so each job is unique).

### How each category supports multi-host

| Category | Multi-host mechanism | Notes |
|----------|----------------------|--------|
| **CI with pull (Buildkite, GitLab CI)** | **Agent/runner fleet** — install an agent or runner on each host; they pull jobs from a central queue. Jobs are assigned to one runner; no shared memory. | Buildkite: tag agents (e.g. `queue=exarp-go`); GitLab: register runners per host; both scale by adding machines. |
| **K8s-native (Argo, Tekton, K8s Jobs)** | **Cluster** — Pods run on any node; scheduler distributes them. One Pod per task (or per wave) = natural distribution across nodes. | Shared storage (e.g. Todo2 DB or plan file) via PVC or shared volume; or exarp-go/API on a service all Pods call. |
| **Workflow engines (Temporal, Prefect, Airflow, Dagster)** | **Worker pool** — run worker processes on many hosts; they poll the same task queue or accept work from the same server. | Temporal/Prefect: add workers on new hosts, same queue. Airflow: CeleryExecutor or KubernetesExecutor. Dagster: run launcher + workers. |
| **Job queues (Redis + Celery/Bull/Sidekiq)** | **Shared queue** — producers push job payloads to Redis (or DB); workers on any host pull. Adding hosts = adding worker processes. | Classic distributed pattern; ensure one queue (or one queue per wave) and workers only pull when ready. |
| **Task runners (Task, Just), cron** | **Not distributed** — single process. To “distribute” you’d run the same cron on multiple hosts and rely on exarp-go **locking** so only one host claims a given task. | Possible but ad hoc: each host runs “get next task, claim, execute”; locking prevents duplicates. Prefer a proper queue or CI if you need many hosts. |

### Design choices for distributed exarp-go execution

1. **Shared Todo2 and locks**  
   All hosts must see the same task state and use the same lock store. Options:
   - **Shared filesystem** for `.todo2/` (NFS, PVC) so every runner sees the same SQLite DB and lock table.
   - **Central Todo2 API** (if you add one): runners call the API to list/claim/update tasks; DB lives on one server.

2. **Who assigns work**  
   - **Central dispatcher:** One process (cron, pipeline, or workflow) gets waves from exarp-go, then creates one “unit of work” per task (job, queue message, or workflow activity). Many workers only execute; they don’t decide “what’s next.”  
   - **Pull by workers:** Workers call exarp-go or Todo2 to “give me a task I can run” (e.g. list Todo, claim one, run `task_execute`). exarp-go’s `ClaimTaskForAgent` is built for this.

3. **Waves and ordering**  
   If you need wave order (wave 0 before wave 1):
   - **Queue per wave:** Producer enqueues to `wave-0` then `wave-1`; workers drain `wave-0` before starting `wave-1`, or use two queues and two worker pools.
   - **DAG / workflow:** Argo, Temporal, Airflow express “wave 1 after wave 0” as dependencies; steps/Pods/activities run in order.

4. **Summary by platform**  
   - **Best “many hosts, minimal ops”:** Redis + Celery (or Bull) + producer that pushes task_ids from exarp-go; workers on N hosts.  
   - **Best “we’re on Kubernetes”:** Argo Workflows or K8s Jobs: one step/Job per task, shared volume or API for Todo2.  
   - **Best “we want durable workflows and retries”:** Temporal: workers on N hosts, one workflow that iterates waves and starts activities.  
   - **Best “we already have CI agents”:** Buildkite or GitLab CI: many agents/runners, pipeline that fans out one job per task_id.

---

## 1. CI / build systems with a “pull” model

These **pull jobs from a queue** instead of only reacting to git events. Good when you want “agents” that continuously take the next piece of work.

| System | How it runs | How it gets “tasks” | How it executes | Fit with exarp-go |
|--------|-------------|----------------------|-----------------|-------------------|
| **Buildkite** | Self-hosted **agents** poll for jobs; you control how many agents and when they run. | Jobs come from pipelines (triggered by push, schedule, or API). You can trigger a pipeline that “pulls” work (e.g. call exarp-go to get next wave/task, then run a job per task). | Each job runs a script (e.g. `./bin/exarp-go -tool task_execute -args '{"task_id":"T-xxx"}'`). | Use pipeline steps to: (1) sync Todo2, (2) get execution order (`report action=plan` or custom script), (3) fan-out one job per task ID (or per wave). |
| **Harness Open Source** | Pipelines run on events or **schedules**. No built-in “pull queue” of your tasks; you add a scheduled pipeline. | Scheduled pipeline runs exarp-go to get work (e.g. `task_workflow` list + filter, or read plan waves). | Pipeline steps run `exarp-go -tool task_execute` or `automation`. | Same as [HARNESS_OPEN_SOURCE_EXARP_GO.md](HARNESS_OPEN_SOURCE_EXARP_GO.md): one pipeline = one “run” of automation or wave; repeat via schedule. |
| **GitLab CI** | Runners pull jobs; can be self-hosted and always-on. | Jobs from pipeline YAML (triggered by push, schedule, or API). | Script in `.gitlab-ci.yml` invokes exarp-go. | Scheduled pipeline: “nightly” runs `automation action=nightly`; “wave runner” pipeline gets task IDs (from API/artifact) and runs `task_execute` per ID. |
| **Tekton (Kubernetes)** | Pipelines run as Pods; **Triggers** or CronJobs can start runs. | Event (webhook, cron) → PipelineRun; pipeline can call a “get work” step that returns task IDs, then fan-out. | One step or Task per `task_execute` call. | CronJob + Pipeline: (1) step gets waves/task list (exarp-go or DB), (2) parallel steps each run `task_execute` for one task_id. |
| **Argo Workflows** | **CronWorkflows** run on a schedule; **Workflow** can have DAG steps. | Cron or event submits Workflow; first step can query Todo2/plan to get task IDs; then DAG with one step per task or per wave. | Each step runs a container that calls exarp-go (CLI or HTTP if you add one). | Good fit: CronWorkflow “every 6h” → Workflow that (1) gets backlog/waves (script or exarp-go), (2) runs a DAG of `task_execute` steps; use `parallelism` and dependencies to respect waves. |

**Summary:** Use **Buildkite** if you want true “agent pulls next job”; use **Argo Workflows** or **Tekton** if you’re on Kubernetes and want scheduled + DAG; use **Harness** or **GitLab CI** for schedule-based runs that call exarp-go.

---

## 2. Workflow / orchestration engines (task queues, durable execution)

These are built for **workflows** and **task queues**: workers pull tasks, run them, and report back. They’re more general than CI and can represent “wave 0 → wave 1 → …” and retries.

| System | How it runs | How it gets “tasks” | How it executes | Fit with exarp-go |
|--------|-------------|----------------------|-----------------|-------------------|
| **Temporal** | **Workers** poll **Task Queues**; long-running, durable. | Your workflow code (or a “dispatcher” workflow) can query an API or DB for “next tasks” and **signal** or **start** activities; or you push task IDs into Temporal as workflow input. | **Activities** = one unit of work (e.g. “run exarp-go task_execute for T-123”). Worker runs activity code that shells out to `exarp-go -tool task_execute`. | **Strong fit:** One workflow “RunBacklogWaves”: for each wave (from exarp-go plan or API), start N activities (one per task_id); wait for wave to complete, then next wave. Activities call exarp-go; Temporal handles retries and observability. |
| **Prefect** | Python-native; **workers** pull **flow runs** from a server. | Flows can be scheduled (cron) or triggered by API; a flow can fetch “work” from Todo2/API and create child runs. | Each flow run can call exarp-go (subprocess or HTTP). | Use a scheduled flow “RunWaves” that (1) loads plan/waves (read file or call exarp-go), (2) for each wave, creates concurrent task runs that call `task_execute`; good if your runner is Python-friendly. |
| **Kestra** | **Workflows** run on **triggers** (schedule, webhook, event). | Schedule trigger (cron) or webhook; workflow tasks can call APIs or scripts to get task IDs. | **Tasks** in YAML: run a script, container, or HTTP call. | Scheduled workflow: trigger `@daily` → first task runs exarp-go to get waves/tasks (or reads plan file); next tasks run `task_execute` per ID (loop or parallel tasks). YAML-based, good for ops-heavy teams. |
| **Apache Airflow** | **Schedulers** schedule **DAGs**; workers execute tasks. | DAG triggered by schedule or manual/API; first task can query DB or call exarp-go to produce “task list” (XCom or downstream mapping). | Task = operator (Bash, Docker, etc.) that runs `exarp-go -tool task_execute`. | DAG “exarp_waves”: task “get_waves” returns list of (wave, task_ids); downstream tasks “execute_T_123” etc., with dependencies so wave N+1 runs after wave N. |
| **Dagster** | **Jobs** run on **schedules** or **sensors**; assets/ops as first-class. | Schedule or sensor triggers job; job can call an op that fetches work from Todo2/API. | Ops run your code (e.g. subprocess to exarp-go). | Similar to Airflow: one job “run_backlog” with op “get_waves” and ops “run_task_T_*”; dependencies by wave. |

**Summary:** Use **Temporal** for robust, long-running “run waves of tasks” with retries and visibility. Use **Kestra** for YAML-defined, schedule- or event-driven runs. Use **Prefect** or **Airflow**/Dagster if you prefer Python and DAGs.

---

## 3. Job queues (simple “pull and run”)

Lighter than full workflow engines: a **queue** of job payloads and **workers** that pull and run them. No built-in “waves”; you encode waves by what you push into the queue.

| System | How it runs | How it gets “tasks” | How it executes | Fit with exarp-go |
|--------|-------------|----------------------|-----------------|-------------------|
| **Redis + Bull (Node) / Celery (Python) / Sidekiq (Ruby)** | Workers pull from Redis (or DB). | A **producer** (cron job or CI step) pushes job payloads (e.g. `{ "task_id": "T-123" }`) into the queue; optionally one queue per wave so workers naturally run wave-by-wave. | Worker job runs `exarp-go -tool task_execute -args '{"task_id":"T-123"}'`. | **Producer:** Scheduled job or post-merge pipeline runs exarp-go to get waves/task IDs and enqueues one job per task_id (or per wave). Workers run continuously and execute. |
| **Kubernetes Jobs / CronJobs** | **CronJob** creates **Job** (one or more Pods) on a schedule. | CronJob spec runs a “dispatcher” container that (1) gets task list from exarp-go or DB, (2) creates one Kubernetes Job per task (or per wave). | Each Job’s container runs `task_execute` for one task_id. | Simple: CronJob every N hours → script gets backlog/waves → `kubectl create job ...` per task; or one CronJob per wave that runs a Pod with a script looping over task IDs. |

**Summary:** Use a **Redis-backed queue** if you already have workers and want “continuous pull”. Use **Kubernetes Jobs + CronJob** if you’re on K8s and want minimal new infrastructure.

---

## 4. Task runners (Make/CMake-like, but more capable)

These replace or wrap **Make/CMake/shell** for “run these tasks in order or in parallel”. They don’t “pull from Todo2” by default; you drive them from a script or from exarp-go.

| System | How it runs | How it gets “tasks” | How it executes | Fit with exarp-go |
|--------|-------------|----------------------|-----------------|-------------------|
| **Earthly** | Runs **targets** in containers; cache and parallelism. | You define targets; a “meta” target or script can be generated from exarp-go plan (e.g. one target per task_id that runs `task_execute`). | Each target runs in a container; can call exarp-go binary. | Generate Earthfile from waves (script that reads plan or calls exarp-go), then `earthly +wave0`, `+wave1`, etc., or `+task-T-123`. |
| **Nx** | Monorepo **task graph**; runs tasks in dependency order with cache. | Nx tasks are defined in code (e.g. “build lib A”); you could have a script that generates Nx config from Todo2/waves so “tasks” = exarp task IDs. | Task runner runs your command (e.g. exarp-go). | More natural for JS/TS monorepos; possible but not ideal unless your repo is already Nx. |
| **Task (taskfile.dev)** | YAML task runner; dependencies and parallelization. | Tasks are declared in YAML; you can have a task “run-wave-0” that calls a script which reads plan and runs `task_execute` for each ID. | `task run-wave-0` runs the script; script loops over IDs and calls exarp-go. | Lightweight: one Taskfile task per wave or “run-all-waves” that reads `.cursor/plans/...plan.md` and invokes exarp-go per task_id. |
| **Just** | Command runner (justfile). | Same idea: `just run-wave 0` runs a script that gets task IDs for wave 0 and runs `task_execute` for each. | Script invokes exarp-go. | Minimal; good for “I want one command that runs current wave from plan”. |

**Summary:** Use **Task** or **Just** to wrap “read waves from plan → run exarp-go per task_id” in a single command. Use **Earthly** if you want containerized, reproducible runs per task.

---

## 5. Cron / system schedulers (minimal)

Nothing “pulls” from a queue; they **run a single command** on a schedule. exarp-go does the rest.

| System | How it runs | How it gets “tasks” | How it executes | Fit with exarp-go |
|--------|-------------|----------------------|-----------------|-------------------|
| **cron** | Runs a command at fixed times. | None; you run one exarp-go invocation (e.g. `automation action=nightly`). | Single command per schedule. | `0 2 * * * /path/to/exarp-go -tool automation -args '{"action":"nightly"}'`. For waves you’d need a wrapper script that gets waves and loops `task_execute`. |
| **systemd timers** | Like cron, more flexible. | Same as cron. | One-shot or repeating service. | Service file runs exarp-go (automation or a script that runs waves); timer triggers it. |
| **launchd** (macOS) | Same idea on Mac. | Same as cron. | RunAtLoad, StartInterval, or calendar. | Plist that runs `exarp-go -tool automation -args '{"action":"daily"}'` or a wave runner script. |

**Summary:** Use **cron** or **systemd/launchd** when you only need “run exarp-go automation on a schedule”; add a **wrapper script** that reads plan/waves and calls `task_execute` in a loop if you want wave-by-wave execution.

---

## 6. Recommended mapping to your needs

| Need | Suggested options |
|------|--------------------|
| **“Like Jenkins but pull-based”** | Buildkite (agents pull jobs); define pipelines that get task IDs from exarp-go and run one job per task. |
| **“Run waves continuously on Kubernetes”** | Argo Workflows (CronWorkflow + DAG of steps calling exarp-go) or Tekton (CronJob + Pipeline with fan-out). |
| **“Durable workflow with retries and visibility”** | Temporal (workflow “RunBacklogWaves”, activities = `task_execute`). |
| **“YAML-defined, schedule + events”** | Kestra (schedule trigger → workflow that gets waves and runs script tasks calling exarp-go). |
| **“Simple queue: producer pushes, workers pull”** | Redis + Celery/Bull/Sidekiq; producer = cron or CI that enqueues task_ids from exarp-go. |
| **“Minimal: no new infra”** | cron/systemd + wrapper script that reads `.cursor/plans/...plan.md` or calls exarp-go to get waves and runs `task_execute` per ID. |
| **“Better Make: one command per wave”** | Task (taskfile.dev) or Just: one command “run-wave” that reads plan and invokes exarp-go for each task in that wave. |
| **“Distributed across many hosts”** | Buildkite (agent per host), GitLab CI (runner per host), Redis + Celery/Bull (worker per host), Temporal (worker per host), Argo/K8s Jobs (Pods across nodes). See **Distributed execution among multiple hosts** above. |

---

## 7. exarp-go integration points (reminder)

- **Get execution order / waves:**  
  `./bin/exarp-go -tool report -args '{"action":"plan"}'` → writes `.cursor/plans/<project>.plan.md` with waves; or `task_analysis action=execution_plan` for ordered task IDs.  
  Or read from Todo2 (e.g. `task_workflow` list + filter) and compute waves with your own script or a small tool.
- **Execute one task:**  
  `./bin/exarp-go -tool task_execute -args '{"task_id":"T-xxx","apply":true}'`
- **Run automation (no wave loop):**  
  `./bin/exarp-go -tool automation -args '{"action":"daily|nightly|sprint|discover"}'`
- **Claim before “assignment”:**  
  Use `task_workflow` or DB: claim task for agent, then run `task_execute` so the same task isn’t run by two runners.

---

## 8. References

- [PARALLEL_EXECUTION_FRAMEWORK.md](PARALLEL_EXECUTION_FRAMEWORK.md) — Waves, BacklogExecutionOrder, automation parallelism.
- [HARNESS_OPEN_SOURCE_EXARP_GO.md](HARNESS_OPEN_SOURCE_EXARP_GO.md) — Harness + exarp-go.
- [TASKS_AS_CI_AND_AUTOMATION.md](TASKS_AS_CI_AND_AUTOMATION.md) — What’s already CI vs exarp automation.
- [JENKINS_INTEGRATION.md](JENKINS_INTEGRATION.md) — Jenkins + exarp-go pattern.
- [CLI_MAKE_CI_USAGE.md](CLI_MAKE_CI_USAGE.md) — Invoking exarp-go from scripts/CI.
- **In this doc:** **Comparison table**, **Shell script vs alternatives**, **Go frameworks and libraries**, **Message queue integration**, **SQLite schema extensions**, **Locking**, **Context7**, and **Distributed execution among multiple hosts** (above).
- **Go libraries:** [Asynq](https://github.com/hibiken/asynq), [Machinery](https://github.com/RichardKnop/machinery), [gocraft/work](https://github.com/gocraft/work), [robfig/cron](https://github.com/robfig/cron).
- **Context7:** Use `resolve-library-id` + `query-docs` for asynq, machinery, robfig/cron, or Temporal to get API-level comparison; see [RESEARCH_HELPERS_REFERENCE.md](RESEARCH_HELPERS_REFERENCE.md).
- External: [Temporal Task Queues](https://docs.temporal.io/task-queue), [Buildkite Queues](https://buildkite.com/docs/pipelines/clusters/manage-queues), [Kestra Schedule Trigger](https://kestra.io/docs/workflow-components/triggers/schedule-trigger), [Argo CronWorkflows](https://argoproj.github.io/argo-workflows/cron-workflows/).
