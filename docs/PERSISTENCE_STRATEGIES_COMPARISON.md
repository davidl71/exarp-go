## Persistence strategies comparison (SQLite / JSON / caches)

This note captures exarp-go’s practical persistence posture and how it maps to
common TUI/agent workflows.

### 1) Authoritative state (source of truth)

- **SQLite (Todo2 DB)** is the canonical task store (`.todo2/todo2.db`).
- Use structured CRUD via tools/CLI wrappers; avoid editing JSON snapshots.

Why:
- concurrency-friendly (transactions, indexing)
- efficient filtered queries
- supports locks/runs/progress tables cleanly

### 2) Snapshots and interfaces (non-authoritative)

- **JSON** is useful as an interoperability surface:
  - exports/imports
  - point-in-time “pack” payloads for UI rendering
  - small, human-auditable configs
- It should not be the primary state store when concurrency/locking matters.

### 3) Derived artifacts (human UX)

Prefer generating explicit artifacts for human inspection:

- markdown reports (overview/scorecard/cockpit)
- json payload dumps (execution packs, briefings)
- local HTML for rich review/approval UI

Artifacts are safe to regenerate and can be cached; they are not authoritative.

### 4) Caching (performance only)

Use TTL/file caches for expensive derived views (e.g., scorecards, doc checks):

- caches must be invalidatable and safe to ignore
- never treat cache contents as the only copy of truth

### Recommended default posture

- **DB-first** for authoritative state (tasks, locks, runs).
- **Artifacts + resources** for read surfaces and UX.
- **JSON/YAML** for configuration and exchange, not concurrency-critical state.

