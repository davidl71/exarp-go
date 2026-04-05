# Conflict policy (cross-tool standardization)

This document records the intended **`internal/conflictpolicy`** package and how it relates to existing detection and tools. It complements multi-agent detection in [`internal/tools/conflict_detection.go`](../internal/tools/conflict_detection.go) and content hashing in [`internal/models/task_hash.go`](../internal/models/task_hash.go).

## Problem

Several tools implement overlapping concepts:

- **Detection** (who collides?): task overlap, shared files, forbidden ownership — `conflict_detection.go`.
- **Equivalence** (same task body?): `models.ContentHash` / normalization — used by import and stores.
- **Policy** (what to do?): `import_on_conflict` (`fail` | `skip`), `git_tools` `conflict_strategy` (`newer` | `source` | `target`), ad hoc SQLite unique handling.

There is no single vocabulary or small set of helpers for **policy + structured incidents** across tools.

## Target package: `internal/conflictpolicy`

Location: **`exarp-go/internal/conflictpolicy/`** (new module; keep **`tools`** depending on **`models`** only — `conflictpolicy` may import `models`, not `tools`, to avoid cycles).

### 1. Modes (policy)

```go
type Mode string

const (
    ModeFail   Mode = "fail"   // abort or surface success:false + payload
    ModeSkip   Mode = "skip"   // omit conflicting side; continue
    ModeSource Mode = "source" // prefer incoming / "source" row
    ModeTarget Mode = "target" // prefer existing / "target" row
    ModeNewer  Mode = "newer"  // tool compares timestamps/version
)

func ParseMode(s string) (Mode, error)
```

- **`import_sqlite`** `import_on_conflict`: `fail` | `skip` → maps here.
- **`git_tools`** merge: `newer` | `source` | `target` → same enum; `newer` logic stays in git merge but strings normalize here.

### 2. Task content key (equivalence)

```go
// TaskContentKey = stable "same body text?" (content + long_description via models.ContentHash).
func TaskContentKey(t *models.Todo2Task) string
```

Replaces duplicated **`importTaskContentKey`** in `task_workflow_import_sqlite.go` once refactored.

### 3. Reason codes and incidents

```go
type Reason string

const (
    ReasonTaskIDContentMismatch Reason = "task_id_content_mismatch"
    ReasonSQLiteUnique          Reason = "sqlite_unique_violation"
    ReasonMultiAgentFileOverlap Reason = "multi_agent_file_overlap"
    ReasonMultiAgentDepOverlap  Reason = "multi_agent_dependency_overlap"
    ReasonForbiddenOwnership    Reason = "forbidden_ownership"
)

type Incident struct {
    Reason  Reason            `json:"reason"`
    TaskIDs []string          `json:"task_ids,omitempty"`
    Path    string            `json:"path,omitempty"`
    Detail  map[string]string `json:"detail,omitempty"`
}
```

Optional adapters: **`IncidentFromFileConflict`**, **`IncidentFromTaskOverlap`**, etc., mapping **`tools`** detection structs → **`Incident`** (lives in **`conflictpolicy`** or a thin **`tools/conflict_incidents.go`** if import cycles force it).

### 4. Decide helper (optional)

```go
func DecideInsert(mode Mode, targetExists, sameContent bool) (doInsert bool, skip Reason, err error)
```

Centralizes **import** and future batch-upsert “insert or skip?” rules.

### 5. Boundaries (what stays elsewhere)

| Concern | Package / file |
|--------|-----------------|
| Detect overlap / files / forbidden | `internal/tools/conflict_detection.go` |
| Hash algorithm | `internal/models/task_hash.go` |
| SQLite CRUD | `internal/database/` |
| Git branch merge apply | `internal/tools/git_tools_actions.go` |

## Implementation order (tasks)

1. **Scaffold `conflictpolicy`** — types, `ParseMode`, `TaskContentKey`, `Reason`, `Incident`, `DecideInsert`, unit tests (no `tools` import from `conflictpolicy`).
2. **Refactor `import_sqlite`** — use `conflictpolicy`; remove duplicate key helper / align SQLite unique handling with shared `Reason` if useful.
3. **Git tools** (optional) — resolve `conflict_strategy` via `ParseMode`; document mapping in git tool schema.
4. **Incidents + automation/session** (optional) — helpers to build `[]Incident` from `DetectConflicts` results; unify JSON shape for prime/automation consumers.

### Todo2 tracking (exarp-go project)

| Step | Task ID |
|------|---------|
| 1 Scaffold | `T-1775395137731895000` |
| 2 import_sqlite refactor | `T-1775395139651015000` (depends on 1) |
| 3 Git tools (optional) | `T-1775395140057192000` |
| 4 Incidents / session (optional) | `T-1775395140594959000` |

## References

- [`docs/ENUM_CANDIDATES.md`](./ENUM_CANDIDATES.md) — other fixed-vocabulary / `Parse*` candidates (import scan mode, split, handoff, comment type, etc.).
- [`docs/MULTI_AGENT_PLAN.md`](./MULTI_AGENT_PLAN.md) — detection integration points.
- [`docs/TASK_CONTENT_HASH_DESIGN.md`](./TASK_CONTENT_HASH_DESIGN.md) — content hash semantics.
