## Backlog Plan

### Current State

- Commit `ecbc27b` removed the deprecated `llamacpp` path and hardened SQLite task handling for legacy `NULL project_id` rows.
- Duplicate analysis currently reports 5 duplicate groups.
- Execution-plan analysis currently sees 2 actionable tasks with no blocking dependencies.

### Wave 0: Ready Now

1. `T-1774349837767010000` `Task agent_role metadata + model routing by role`
2. `T-1774349847557324000` `automation action=schedule with launchd (macOS) and overlap detection`

Reason:
- Both are dependency-free according to `task_analysis action=execution_plan`.
- `agent_role metadata + model routing` is higher leverage because it improves how exarp-go routes work before adding more automation.

### Wave 1: DB/Task-Store Cleanup

1. `T-1774327670844972000` `Add QueryContextDB helper to reduce boilerplate`
2. `T-1774327646312766000` `Convert GetTask to use sqlx`
3. `T-1774327651740736000` `Convert ListTasks to use sqlx.Select`
4. `T-1774327662714638000` `Convert tag_cache queries to sqlx`
5. `T-1774327657588812000` `Convert lock_monitoring queries to sqlx`

Reason:
- These tasks are one coherent refactor stream.
- The recent `NULL project_id` fix exposed that task-store correctness still depends on repeated hand-written SQL paths.
- Start with the shared helper, then move the highest-value task reads, then the lower-risk cache/monitoring paths.

### Wave 2: Validate and Prune

1. `T-1774340466624224000` `Test NOT NULL fix`
2. Review placeholder Todo tasks:
   - `T-1774344842224405000` `Visual`
   - `T-1774344842240369000` `Multi-Agent`
   - `T-1774344842255121000` `Scheduled`
   - `T-1774344842270815000` `Lazy`
   - `T-1774344837128726000` `Ledger/Continuity`

Reason:
- `Test NOT NULL fix` may still be valid, but it likely overlaps with the recent DB hardening and should be reviewed before implementation.
- The placeholder tasks are not actionable as written and should be expanded, split, or cancelled before they remain in the active backlog.

### Duplicate Cleanup

Current duplicate groups flagged by analysis:

1. `T-1774099861232879000` and `T-1774079470546533000`
2. `T-209` and `T-213`
3. `T-1774291426817117000` and `T-1774291425905206000`
4. `T-1773310482832226000` and `T-1773310480291859000`
5. `T-1773330033559952000` and `T-1773329378465490000`

Recommended action:
- Triage these before adding more medium/low-priority backlog items.
- Keep the better-scoped task in each pair and merge comments/history onto it.

### Recommended Next Action

Start with `T-1774349837767010000`, then either:

- continue directly into `T-1774349847557324000`, or
- if task-store robustness is still the priority, start the DB/sqlx refactor chain in Wave 1 immediately after `agent_role metadata + model routing`.
