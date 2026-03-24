---
name: database-maintenance
description: Explicit SQLite database maintenance for exarp-go. Use when the user asks to inspect database size/free space, run WAL checkpoint, run VACUUM, run ANALYZE, diagnose SQLite file growth, or verify database maintenance health. Prefer `health action=database` over ad hoc PRAGMA commands or direct sqlite3 shell usage.
---

# Database Maintenance

Use the `health` tool with `action=database` for explicit SQLite maintenance and status.

## Operations

| Need | Tool call |
|------|-----------|
| Inspect database status | `health` with `action=database`, `operation=status` |
| Run WAL checkpoint | `health` with `action=database`, `operation=checkpoint`, optional `checkpoint_mode=PASSIVE|FULL|RESTART|TRUNCATE` |
| Reclaim free space | `health` with `action=database`, `operation=vacuum` |
| Refresh planner statistics | `health` with `action=database`, `operation=analyze` |

## Rules

1. Treat full maintenance as explicit. Do not assume normal task CRUD should checkpoint, vacuum, or sync stores.
2. Start with `operation=status` before `vacuum` unless the user already asked for a specific maintenance action.
3. Use `checkpoint_mode=TRUNCATE` when the goal is to shrink a WAL file explicitly.
4. Report before/after size signals when you run `vacuum`:
   `estimated_db_bytes`, `estimated_free_bytes`, `page_count`, `freelist_count`.
5. If config docs mention `auto_vacuum` or `checkpoint_interval`, do not assume they are active runtime automation unless the code shows they are wired.

## Examples

- `./bin/exarp-go -tool health -args '{"action":"database","operation":"status"}'`
- `./bin/exarp-go -tool health -args '{"action":"database","operation":"checkpoint","checkpoint_mode":"TRUNCATE"}'`
- `./bin/exarp-go -tool health -args '{"action":"database","operation":"vacuum"}'`
- `./bin/exarp-go -tool health -args '{"action":"database","operation":"analyze"}'`
