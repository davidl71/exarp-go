# exarp-go CLI shortcuts

Shortcuts to avoid repeating long commands: environment variables, shell aliases, and Makefile targets.

## Environment variables

| Variable | Purpose | Example |
|----------|---------|--------|
| `REDIS_ADDR` | Redis for queue/worker (optional) | `127.0.0.1:6379` |
| `ASYNQ_QUEUE` | Asynq queue name (default: `default`) | `default` |
| `ASYNQ_CONCURRENCY` | Worker concurrency (default: 5) | `5` |
| `PROJECT_ROOT` | Override project root (some tools) | `/path/to/repo` |

Use a project-local env file (e.g. `.env` or `.env.local`) and `source` it, or export in your shell profile:

```bash
# Optional: point to local binary so 'exarp-go' uses project build
export PATH="$PWD/bin:$PATH"
# Optional: Redis for queue/worker
export REDIS_ADDR=127.0.0.1:6379
```

## Shell aliases (optional)

Add to `~/.zshrc` or `~/.bashrc` (adjust path to your exarp-go repo or `bin`):

```bash
# Use project binary; run from repo root or set EXARP_GO_ROOT
alias exarp='./bin/exarp-go'
alias exarp-task='./bin/exarp-go task'
alias exarp-queue='./bin/exarp-go queue'
alias exarp-worker='./bin/exarp-go worker'
```

From any directory (if you install exarp-go to PATH via `make install`):

```bash
alias exarp-task='exarp-go task'
alias exarp-queue='exarp-go queue'
alias exarp-worker='exarp-go worker'
```

## Makefile targets

Run from **project root**. `make build` is required first (or use `make queue-enqueue-wave` / `make queue-worker` which depend on `build`).

| Target | Usage |
|--------|--------|
| `make task-list` | List tasks (use `TASK_FLAGS="--status Todo"` to filter) |
| `make task-list-todo` | List Todo tasks |
| `make task-list-in-progress` | List In Progress tasks |
| `make task-update` | `make task-update TASK_ID=T-123 NEW_STATUS=Done` |
| `make task-create` | `make task-create TASK_NAME="Title" TASK_DESC="Description" TASK_PRIORITY=medium` |
| `make queue-enqueue-wave` | Enqueue wave 0 (set `WAVE=1` for wave 1). Requires `REDIS_ADDR`. |
| `make queue-worker` | Run Asynq worker. Requires `REDIS_ADDR`. |
| `make exarp-report-scorecard` | Run report scorecard (compact JSON) |
| `make exarp-report-overview` | Run report overview (compact JSON) |

Examples:

```bash
# Create a task without typing long CLI
make task-create TASK_NAME="Config and docs for Redis+Asynq (M3)" TASK_DESC="Document REDIS_ADDR in ALTERNATIVES doc." TASK_PRIORITY=medium

# Enqueue wave 0 (Redis must be running)
export REDIS_ADDR=127.0.0.1:6379
make queue-enqueue-wave
make queue-enqueue-wave WAVE=1

# Run worker in another terminal
export REDIS_ADDR=127.0.0.1:6379
make queue-worker
```

## See also

- [ASYNQ_FOLLOWUP_TASKS.md](ASYNQ_FOLLOWUP_TASKS.md) — Queue follow-up task list
- [.cursor/plans/redis-asynq-and-tui.plan.md](../.cursor/plans/redis-asynq-and-tui.plan.md) — Redis+Asynq plan
- `make help` — All Makefile targets
