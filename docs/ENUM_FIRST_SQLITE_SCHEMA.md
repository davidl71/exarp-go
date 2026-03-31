## Enum-first SQLite schema (tasks + future enum fields)

This repo is moving toward **enum-first** internal storage for categorical fields, while keeping
string columns during a client compatibility window.

### Why enum-first helps

- **Faster filters**: integer comparisons + smaller indexes.
- **Fewer normalization paths**: avoid repeated `NormalizeStatus*` / `ParseTask*` work on hot paths.
- **Safety**: `CHECK` constraints prevent “string drift” and typos from entering the DB.
- **Clearer intent**: the schema encodes the allowed state space explicitly.

### Current enum-first fields (tasks)

Migration `migrations/016_recreate_tasks_with_enum_and_timeints.sql` introduces:

- `tasks.status_enum` and `tasks.priority_enum` as `INTEGER`
- `tasks.created_ts`, `tasks.last_modified_ts`, `tasks.completed_at_ts` as unix seconds `INTEGER`
- `CHECK` constraints that enforce **string ↔ enum** consistency:
  - `status` ∈ {Todo, In Progress, Review, Done, Blocked, Cancelled} and matches `status_enum`
  - `priority` ∈ {'', low, medium, high, critical} and matches `priority_enum`

### Querying guidance (internal/database)

Prefer enum/int columns in new query paths:

- **Statuses**
  - `WHERE status_enum = ?` (single)
  - `WHERE status_enum IN (?, ?, ...)` (multi)
- **Priority**
  - `WHERE priority_enum = ?`
- **Hot-path open backlog**
  - Use the partial index (Todo/In Progress): `status_enum IN (1, 2)`

Keep string columns for:
- compatibility output
- ad-hoc inspection and human readability

### “More enums” follow-through: required logic changes

If we add enum columns for other TEXT categorical fields, we should also:

1) **Write both forms**
   - Always populate `*_enum` and keep the string column canonical.

2) **Read enum-first**
   - Prefer enum columns in WHERE clauses and in in-memory comparisons.

3) **Add the right indexes**
   - `CREATE INDEX ... ON <table>(<enum_col>)`
   - partial indexes for common hot subsets.

### Next enum candidates (recommended)

- `task_comments.comment_type` → `comment_type_enum`
- `task_activities.activity_type` → `activity_type_enum`
- `tasks.metadata_format` → `metadata_format_enum` (json/protobuf)
- `task_execution_runs.status` → `status_enum` (execution cockpit)
- `task_verifications.kind` → `kind_enum`

### References

- SQLite generated columns: `https://sqlite.org/gencol.html`
- SQLite CHECK constraints: `https://www2.sqlite.org/lang_createtable.html#ckconst`
- SQLite partial indexes: `https://sqlite.org/partialindex.html`

