## Integration patterns (API / Shell / MCP)

This note summarizes practical patterns for connecting UIs (Cursor/OpenCode/TUI),
shell automation, and structured tools/resources in **exarp-go**.

### Core principle: one integration spine

- **UI surfaces** (Cursor chat, OpenCode hooks, CLI, TUI) should be thin.
- **All state mutation and domain logic** should live in **MCP tools** (and the
  underlying DB/task store).
- **Resources** (`stdio://...`) are the read-only projection layer for UIs.

In practice:

```
UI / Hook / Script
  → exarp-go CLI (optional convenience)
  → MCP tool call (structured JSON)
  → DB + domain logic
  → Resource(s) / formatted result(s) back to UI
```

### When to prefer MCP tools vs CLI vs “API”

- **MCP tool**: best default for structured operations (task CRUD, reports,
  scheduling, health). Works well over **stdio** in IDEs and in local scripts.
- **CLI subcommand**: best for operator convenience wrappers, especially when
  the output is a **local artifact** (e.g., `task review` generating HTML and
  opening it in a browser) or when you want shell-friendly ergonomics.
- **HTTP API**: only when you truly need multi-client, remote access, or a
  shared service boundary. stdio remains the simplest default for local work.

### Browser artifact pattern (rich review without server complexity)

When you need rich interaction (approve/reject, copy/paste commands, annotate),
a **local browser artifact** is often cheaper than building that UX in the TUI:

- CLI generates deterministic HTML in a temp dir (or project docs/)
- HTML renders JSON payloads (execution packs, reports)
- Buttons emit copyable `exarp-go -tool ... -args '{...}'` commands

This keeps the “UI” disposable and ensures the *real* logic lives in tools.

### Scheduling pattern: OS-native schedule + overlap guard

For recurring background jobs:

- `automation action=schedule` installs a **launchd/systemd** timer that runs:
  `exarp-go -tool automation -args '{"action":"<target_action>", ...}'`
- Use a DB-backed overlap guard (e.g., `automation_runs`) so schedules are
  **idempotent** and don’t overlap even if timers drift or the machine wakes.

### Recommended defaults for exarp-go

- Use **stdio** transport by default (IDE-local workflows).
- Prefer **resources** for read surfaces and “ambient awareness”.
- Keep “one-off” human workflows as **CLI** wrappers over tool calls.
- Add HTTP only when you have a concrete remote/multi-client need.

