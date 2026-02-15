# Protobuf and TUI/TUI3270 — Future Improvements

*Document for future improvement. Current decision: no change to TUI3270; it keeps using the database and go3270 as today.*

## Current behavior

### TUI3270 data sources

- **Task list:** Loaded **directly from the database** via `loadTasksForStatus` → `database.GetTasksByStatus` / `database.Todo2Task`. The 3270 TUI does **not** call `task_workflow` for the list; it reads from the DB.
- **Session:** The TUI calls `server.CallTool(ctx, "session", argsBytes)` (e.g. in `internal/cli/tui3270.go`). The result is whatever the server returns (text/JSON). Proto already defines the session response shape (e.g. `SessionPrimeResult`).
- **Rendering:** Screens are built with `go3270.Screen` (rows, fields, attributes). The tn3270 protocol and go3270 library are unchanged by protobuf.

### Protobuf’s role today

- Proto defines **tool request/response messages** in `proto/tools.proto`. Tool handlers build proto, then convert to map/JSON for MCP and CLI.
- The TUI3270 task list path does **not** use those tool responses; it uses `database.Todo2Task` only.

---

## Where protobuf can help (future options)

### 1. No change (current decision)

- TUI3270 continues to use the DB for task list and go3270 for rendering.
- Session (and any other tool the TUI calls) already benefits from a stable proto-backed response schema when the TUI parses JSON.

### 2. TUI screen / display contract in proto (optional)

- Add messages in `proto/tools.proto` for TUI display, e.g.:
  - `TUI3270TaskListScreen` or similar: rows, column definitions, cursor index, scroll state.
- One place builds this from `database.Todo2Task` (or from a task list response); the 3270 code only renders from the proto.
- **Benefits:** Single definition of “task list for display”; easier to reuse for another client (e.g. web TUI or a future remote 3270 client).
- **Files:** `proto/tools.proto`, `internal/cli/tui3270.go` (build screen from proto or from existing DB load + proto marshal).

### 3. Remote TUI / wire format (future)

- If the 3270 client is ever split from the server (custom protocol over the network instead of in-process tn3270), proto could be the wire format for:
  - Screen updates (e.g. list of fields/lines),
  - Commands (e.g. scroll, select, run).
- Not required for the current in-process, tn3270-based TUI.

---

## What protobuf does not replace

- **3270 protocol and rendering:** tn3270 and `go3270.Screen` stay as-is.
- **Current task list path:** Refactoring tool responses to proto (e.g. `TaskListResponse`) does not change how the 3270 task list is filled unless the TUI is refactored to call `task_workflow` and parse that response instead of using the DB.

---

## References

- [.cursor/plans/protobuf-execution.plan.md](../.cursor/plans/protobuf-execution.plan.md) — Proto execution plan and waves
- [internal/cli/tui3270.go](../internal/cli/tui3270.go) — 3270 TUI implementation; `loadTasksForStatus`, `taskListTransaction`, session `CallTool`
- [proto/tools.proto](../proto/tools.proto) — Tool request/response messages
