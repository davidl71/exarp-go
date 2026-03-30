## State management patterns (projection-first)

This note summarizes a pragmatic state management posture for exarp-go-style
operator tooling (CLI/TUI + automation + MCP tools/resources).

### 1) Keep one authoritative model

- Maintain a single **source of truth** for mutable state (e.g. TaskStore backed
  by SQLite).
- Treat everything else as a **projection** or **artifact**.

### 2) Events/messages drive updates

Adopt a message/update style (Elm/TEA) when complexity warrants it:

- **Message**: an input event (“update task”, “claim”, “start run”).
- **Update**: the handler that applies business rules and writes to the store.
- **View**: resources (`stdio://...`) and formatted outputs for UIs.

When complexity is low, a simple “App struct + event loop + explicit handlers”
is sufficient; the key is keeping mutations centralized.

### 3) Views are projections

Preferred “read” surfaces:

- MCP resources for structured projections (task lists, execution packs)
- report outputs (markdown/json) for human consumption
- local HTML artifacts for rich review UX (still a projection)

Views should be safe to delete and regenerate.

### 4) Effects are separate (optional)

If you need background work:

- schedule OS-native jobs (launchd/systemd) that call `automation`
- guard overlap via DB (idempotency/anti-concurrency)
- keep side effects visible via emitted artifacts/resources

### Recommended default posture

- **Projection-first** architecture: model (store) → projections (resources) → views (UI).
- Use TEA-style Message/Update separation for heavily interactive TUIs; keep
  light CLIs as thin wrappers over tool calls.

