# JSONL Audit Trail for Todo2 Task Changes

**Date:** 2026-02-26
**Status:** Design / Research
**Source:** OpenWork (different-ai/openwork) pattern analysis

---

## Problem

Todo2 tracks task state in SQLite (primary) and JSON (fallback), but lacks a general-purpose audit log for significant actions. When debugging task state issues or understanding "what happened," we rely on Git history and database timestamps — neither gives a complete, queryable timeline.

## Proposed Solution

Append-only JSONL file that records significant task lifecycle events.

### Format

Each line is a self-contained JSON object:

```json
{"ts":"2026-02-26T15:30:00Z","event":"task.created","task_id":"T-123","agent":"general-host-456","data":{"content":"Add feature X","priority":"high"}}
{"ts":"2026-02-26T15:31:00Z","event":"task.status_changed","task_id":"T-123","agent":"general-host-456","data":{"from":"Todo","to":"In Progress"}}
{"ts":"2026-02-26T16:00:00Z","event":"task.lock_acquired","task_id":"T-123","agent":"backend-host-789","data":{"lease_duration":"30m"}}
```

### Event Types

| Event | When | Data |
|-------|------|------|
| `task.created` | CreateTask | content, priority, tags |
| `task.status_changed` | UpdateTask (status differs) | from, to |
| `task.updated` | UpdateTask (non-status fields) | changed_fields |
| `task.deleted` | DeleteTask | content (for recovery) |
| `task.lock_acquired` | ClaimTaskForAgent | agent, lease_duration |
| `task.lock_released` | ReleaseTask | agent |
| `task.lock_expired` | CleanupExpiredLocks | agent, expired_at |

### File Location

`.todo2/audit.jsonl` — co-located with the database.

### Properties

- **Append-only**: no mutations, no deletions (except rotation)
- **Schema-free**: new event types can be added without migration
- **Grep-friendly**: `grep "task.deleted" .todo2/audit.jsonl`
- **Zero migration**: just start writing; old installations get the file on first event
- **Rotation**: optional size-based rotation (e.g., 10MB → archive and start new file)

## Implementation Plan

1. Add `audit.go` in `internal/database/` with `AppendAuditEvent(event AuditEvent) error`
2. Call from `CreateTask`, `UpdateTask`, `DeleteTask`, `ClaimTaskForAgent`, `ReleaseTask`
3. File open/close per write (safe for concurrent agents; no long-lived handle)
4. Optional: `AuditQuery(filter)` for reading back events

## Trade-offs

| Pro | Con |
|-----|-----|
| Zero-cost when not queried | Disk usage grows over time |
| No schema migration needed | Not indexed (linear scan) |
| Works across agents/processes | Needs rotation for long-running projects |
| Git-trackable if desired | Duplicate of some DB info |

## Alternatives Considered

- **SQLite audit table**: More queryable but requires migration, couples tightly to DB schema
- **Git commit messages**: Already available but not structured or queryable
- **Event sourcing**: Full event sourcing would replace the current state-based model — too invasive

## Decision

Implement as a lightweight, opt-in feature. Default: enabled. Can be disabled via config (`database.audit_log: false`).

## References

- OpenWork JSONL pattern: `docs/research/OPENCODE_PLUGIN_PATTERNS.md` §3.2
- Current task CRUD: `internal/database/tasks_crud.go`
- Agent locking: `internal/database/tasks_lock.go`
