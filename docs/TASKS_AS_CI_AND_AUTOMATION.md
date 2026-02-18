# Tasks as CI Scripts and Exarp Automation

**Purpose:** Identify Todo2 tasks that are better implemented as CI steps or exarp automation (daily/nightly/sprint) so they run automatically instead of as one-off manual tasks.

**Last updated:** 2026-01-29

---

## 1. Already automation / CI

| What | Where |
|------|--------|
| Tool smoke tests | Agentic CI: `-test tool_workflow`, `automation`, `testing` |
| Todo2 sync | Agentic CI: `task_workflow` sync |
| Report overview | Agentic CI: `report` overview |
| Alignment + duplicates | `make pre-sprint` → `task_analysis duplicates` + `analyze_alignment todo2` |
| Daily/sprint automation | `automation` tool (daily/nightly/sprint) + `make sprint`, `sprint-start` |
| Task completion inference | `make check-tasks`, `make update-completed-tasks` |
| Task sanity (epochs, dup IDs, invalid status, missing deps) | `make task-sanity-check` (task_workflow sanity_check) |
| Pre-push | Git hook: `analyze_alignment todo2` + security scan |
| Pre-commit | Git hook: build + health docs + security scan |

**Implication:** Any Todo2 item that only says “run alignment” or “run duplicate detection” is already covered; use `make pre-sprint` or the automation tool instead of creating a one-off task.

---

## 2. Todo2 tasks that should be CI or exarp automation

| Task / theme | Recommendation |
|--------------|----------------|
| **Verify all 24 tools work via Go server** | **CI.** Extend agentic-ci (or sanity-check) to run all tools and fail if any crash. |
| **Verify all 6 resources / all 8 prompts** | **CI.** Add a step that lists and accesses each resource/prompt; fail if any error. |
| **Verify project_overview and consolidated_reporting** | **CI.** Ongoing guarantee: run `report` overview/scorecard in CI (already done for overview). |
| **Run duplicate detection on epoch-based tasks** | **Exarp automation.** Use `make pre-sprint` or automation daily/sprint; do not create a one-off task. |
| **Alerting system for stale lock patterns** | **Exarp automation.** Nightly job that runs `DetectStaleLocks` and reports (implemented in automation nightly). |
| **Review new epoch-based tasks for relevance** | **Exarp + human.** Automate the report (list epoch-based tasks + status); keep review as human step. |
| **Run migration tool (dry-run first)** | **CI (optional).** Add `make migrate-dry-run` to CI so migration script is validated on each run. |
| **Verify `bin/exarp-go` exists and is executable** | **CI.** Already true in agentic-ci (build + run). Drop as a task or document “CI enforces this.” |
| **Task sanity check** | **CI.** Run `make task-sanity-check` in agentic-ci so invalid Todo2 state blocks merge. |

---

## 3. Implemented CI and automation changes

### 3.1 Agentic CI (`.github/workflows/agentic-ci.yml`)

- **Todo2 task sanity check:** Run `make task-sanity-check` in the agent-validation job so that invalid Todo2 state (epoch dates, empty content, invalid status, duplicate IDs, missing deps) fails the pipeline.
- **Optional:** `make migrate-dry-run` can be added in a separate step when migration tool is built (e.g. only on main/develop) to validate the migration script.

### 3.2 Nightly automation (stale lock alerting)

- **Stale lock check** added to the automation tool’s **nightly** action.
- Nightly runs `database.DetectStaleLocks(ctx, 5*time.Minute)` and appends a summary to the nightly result (expired count, near-expiry count, stale count).
- If the database is not available (e.g. no Todo2 DB), the step is reported as skipped or soft-fail so nightly still completes.

### 3.3 Documentation

- This document: `docs/TASKS_AS_CI_AND_AUTOMATION.md` — plan and mapping of tasks to CI vs automation.

### 3.4 One-off tasks removed (addressed by CI/automation)

The following one-off tasks were **removed** from both the Todo2 database (`.todo2/todo2.db`) and the Todo2 state file (`.todo2/state.todo2.json`) so they no longer appear in the backlog. The work is covered by CI or exarp automation.

| Task ID | Task name | How addressed |
|---------|-----------|----------------|
| ~~T-1769716109198~~ | Verify project_overview and consolidated_reporting with Go project | Removed. Ongoing: agentic-ci runs `report` overview. |
| ~~T-1768318259581~~ | Run duplicate detection on epoch-based tasks for consolidation | Removed. Use `make pre-sprint` or automation daily/sprint (`task_analysis` duplicates). |
| ~~T-321~~ | Alerting system for stale lock patterns | Removed. Nightly automation runs `DetectStaleLocks` and reports counts. See §3.2. |
| ~~T-401~~ | Verify all 24 tools work via Go server | Removed. Agentic-ci runs tool tests and task sanity check. |
| ~~T-402~~ | Verify all 8 prompts work via Go server | Removed. Approach in §2; optional CI step can be added later. |
| ~~T-403~~ | Verify all 6 resources work via Go server | Removed. Approach in §2; optional CI step can be added later. |

**Removal:** Use **batch delete** (one call) for speed: `task_workflow` with `action=delete` and `task_ids` comma-separated. See skill `.cursor/skills/task-cleanup/SKILL.md`.

### 3.5 Performance tasks removed

Performance-audit and performance-only tasks were removed from both the database and Todo2 so they no longer clutter the backlog. If they reappear (e.g. after a restore or sync from another source), remove them again with **one** batch delete:

```bash
./bin/exarp-go -tool task_workflow -args '{"action":"delete","task_ids":"T-1768325753706,T-1768325747897,T-1768325741814,T-1768325734895,T-1768325728818,T-1768325715211,T-247,T-177,T-226,T-81,T-166,T-359,T-222,T-200,T-1768317407961"}'
```

Or use the **task-cleanup** skill (one MCP call with the same `task_ids`).

| Task ID | Task name |
|---------|-----------|
| T-1768325753706 | Performance audit: Concurrent tool execution |
| T-1768325747897 | Performance audit: Memory allocation optimization |
| T-1768325741814 | Performance audit: Profile-Guided Optimization (PGO) |
| T-1768325734895 | Performance audit: Database query optimization |
| T-1768325728818 | Performance audit: File I/O caching |
| T-1768325715211 | Performance audit: Python bridge subprocess overhead |
| T-247 | Performance improvements measured and documented |
| T-177 | Create baseline performance metrics |
| T-226 | Add agent performance tracking |
| T-81 | Performance metrics (lock contention, wait times) |
| T-166 | Performance benchmarks (if applicable) |
| T-359 | Performance optimization |
| T-222 | Performance benchmarks |
| T-200 | Performance benchmarks meet targets |
| T-1768317407961 | Measure performance improvements for protobuf serialization |

### 3.6 Done task pruning

To clean the backlog by removing **Done** tasks:

```bash
# Preview (dry run) – legacy IDs only (T-XX where XX < 1000000)
make task-prune-done DRY_RUN=1

# Prune legacy Done tasks
make task-prune-done

# Prune all Done tasks
make task-prune-done PRUNE_WAVE=all

# Prune by tag (tag tasks first, then delete)
make task-prune-done PRUNE_TAG=prune DRY_RUN=1
make task-prune-done PRUNE_TAG=prune
```

**Prune waves:**
- `PRUNE_WAVE=legacy` (default): Remove Done tasks with legacy sequential IDs (T-1 … T-999999).
- `PRUNE_WAVE=all`: Remove all Done tasks.

**Prune by tag:**
- `PRUNE_TAG=<tag>`: Remove tasks with the given tag. Tag tasks (e.g. via `exarp-go task update T-123` or TUI), then prune by tag. Overrides `PRUNE_WAVE`.

---

## 4. How to use this

- **Before creating a “verify X” or “run Y check” task:** Check this doc and the table in §2. If it’s better as CI or exarp automation, add the step to agentic-ci or to daily/nightly/sprint instead of creating a Todo2 task.
- **Recurring checks:** Prefer CI (agentic-ci) or exarp (automation + Makefile) so they run automatically; use Todo2 only for one-off or human-review work.
- **Stale locks:** Rely on nightly automation’s stale lock report; no separate manual Todo2 task needed for “check stale locks.”

---

## 5. References

- Agentic CI: `.github/workflows/agentic-ci.yml`
- Automation tool: `internal/tools/automation_native.go` (daily, nightly, sprint)
- Task sanity: `make task-sanity-check` → `task_workflow` action `sanity_check`
- Stale locks: `internal/database/lock_monitoring.go` (`DetectStaleLocks`, `CleanupExpiredLocksWithReport`)
- Makefile: `pre-sprint`, `sprint`, `check-tasks`, `update-completed-tasks`, `migrate-dry-run`
