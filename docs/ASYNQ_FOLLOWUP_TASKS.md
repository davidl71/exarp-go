# Asynq Integration — Follow-up Tasks

Follow-up work for the Redis+Asynq producer (M1) and worker (M2) implemented in `internal/queue` and CLI (`queue enqueue-wave`, `worker`). Create these in Todo2 when your DB is ready (e.g. after fixing migration 8 / `assigned_to` if needed).

Reference: [.cursor/plans/redis-asynq-and-tui.plan.md](../.cursor/plans/redis-asynq-and-tui.plan.md)

---

## M3 — Config and docs (medium)

**Name:** Config and docs for Redis+Asynq (M3)

**Description:** Document `REDIS_ADDR`, `ASYNQ_QUEUE`, `ASYNQ_CONCURRENCY` in `docs/ALTERNATIVES_ITERATIVE_CI_SCRIPTS.md`. Add a "Redis + Asynq integration" subsection with producer/worker usage, lock flow (claim → task_execute → release), and pointer to the plan. Optional: Makefile targets (e.g. `make run-asynq-worker`) or README section.

**Tags:** config, docs  
**Priority:** medium

---

## M4 — TUI queue-aware (medium)

**Name:** TUI queue-aware: Enqueue wave or queue status (M4)

**Description:** Keep TUI waves view as-is. If `REDIS_ADDR` is set: add a "Queue: enabled" indicator and/or "Enqueue current wave" action that calls the producer for the current wave. Alternatively, add a queue/jobs view showing enqueued/running counts (e.g. via Asynq inspector). Ensure TUI and plan stay source of truth; enqueue uses same waves data the TUI shows.

**Tags:** feature, tui  
**Priority:** medium  
**Dependencies:** (conceptually on producer; already implemented)

---

## M5 — Optional dispatcher (low)

**Name:** Optional: dispatcher to enqueue waves on schedule (M5)

**Description:** Cron or scheduler runs a small dispatcher that gets the execution plan (or reads `.cursor/plans/...plan.md`), then enqueues one job per task_id for the next wave(s). Enables "run backlog on schedule" without running workers on the same machine.

**Tags:** feature  
**Priority:** low  
**Dependencies:** Producer + worker (done)

---

## Implementation polish

### Vendor sync (medium)

**Name:** Vendor sync for Asynq and new deps

**Description:** Run `go mod vendor` to include `github.com/hibiken/asynq` and transitive deps in `vendor/` so `make build` with `-mod=vendor` succeeds. Resolve any `vendor/modules.txt` inconsistencies.

**Tags:** build, config  
**Priority:** medium

---

### Unit tests for internal/queue (medium)

**Name:** Unit tests for internal/queue

**Description:** Test `queue` package: `ConfigFromEnv`, `Config.Enabled()`; `EnqueueTask`/`EnqueueWave` when disabled (no-op); payload marshal/unmarshal; optional integration test with real Redis (skip when `REDIS_ADDR` unset).

**Tags:** testing  
**Priority:** medium

---

## Create in Todo2

After fixing the DB (e.g. migration 8 / `assigned_to`), run from project root:

```bash
# M3
go run ./cmd/server task create "Config and docs for Redis+Asynq (M3)" \
  --description "Document REDIS_ADDR, ASYNQ_QUEUE, ASYNQ_CONCURRENCY in ALTERNATIVES_ITERATIVE_CI_SCRIPTS.md; add Redis+Asynq subsection. Plan M3." \
  --priority medium

# M4
go run ./cmd/server task create "TUI queue-aware: Enqueue wave or queue status (M4)" \
  --description "If REDIS_ADDR set: queue indicator and/or Enqueue current wave in TUI. Plan M4." \
  --priority medium

# M5
go run ./cmd/server task create "Optional: dispatcher to enqueue waves on schedule (M5)" \
  --description "Cron/scheduler dispatcher: get plan, enqueue next wave(s). Plan M5." \
  --priority low

# Vendor
go run ./cmd/server task create "Vendor sync for Asynq and new deps" \
  --description "go mod vendor to include hibiken/asynq so make build with -mod=vendor succeeds." \
  --priority medium

# Tests
go run ./cmd/server task create "Unit tests for internal/queue" \
  --description "Test queue: ConfigFromEnv, EnqueueTask/EnqueueWave when disabled, payload marshal/unmarshal." \
  --priority medium
```

Or use the **task_workflow** MCP tool with `action=create` and the same names/descriptions.
