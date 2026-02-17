# exarp-go and OpenCode Integration

OpenCode is an open-source AI coding agent (TUI, desktop, IDE). exarp-go can interact with it in three ways: **MCP**, **CLI**, and **HTTP API**.

---

## 1. MCP server (recommended)

OpenCode supports [MCP servers](https://opencode.ai/docs/configure/mcp-servers) in its config. Add exarp-go as a **local** MCP server so OpenCode’s agent can call exarp-go tools (tasks, reports, session, health, etc.).

### Config file location

OpenCode loads config from (later overrides earlier):

| Source | Location |
|--------|----------|
| **Global** | `~/.config/opencode/opencode.json` (or `opencode.jsonc`) on macOS, Linux, and Windows |
| **Project** | `opencode.json` in the project root |
| **Custom path** | Set `OPENCODE_CONFIG` to a file path |
| **Custom dir** | Set `OPENCODE_CONFIG_DIR` for agents, commands, modes, plugins |

Use the global file for exarp-go MCP if you want it in every project; use a project `opencode.json` to limit it to that repo. See [OpenCode Config](https://opencode.ai/docs/config/).

### Config

In your OpenCode config file, add:

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/absolute/path/to/exarp-go/bin/exarp-go"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/path/to/your/project"
      },
      "timeout": 10000
    }
  }
}
```

- **command**: Path to the exarp-go binary. Use an absolute path, or ensure `exarp-go` is on `PATH` and use `["exarp-go"]`.
- **environment.PROJECT_ROOT**: Project root (where `.todo2` and `.exarp` live). OpenCode may support a variable for “current project”; if so, use that. Otherwise set it per project or use a wrapper script that sets `PROJECT_ROOT` from the current directory.
- **timeout**: Optional; exarp-go tool calls can be slow (e.g. report). 10000 ms or higher is reasonable.

### Behavior

- When OpenCode starts the server, it runs the `command` with the given `environment`.
- exarp-go detects non-TTY stdin and runs in **MCP server mode** (JSON-RPC 2.0 on stdin/stdout).
- OpenCode then discovers and invokes exarp-go tools (e.g. `task_workflow`, `report`, `session`, `health`, `testing`).

### CLI management (if available)

If your OpenCode version supports it, you can use:

- `opencode mcp add` – add exarp-go
- `opencode mcp list` – list servers
- `opencode mcp debug exarp-go` – test connection

---

## 2. CLI (terminal / scripts)

From OpenCode’s integrated terminal (or any shell), you can run exarp-go as a normal CLI. OpenCode can then use the output or drive behavior via shell commands.

### From project root

```bash
cd /path/to/project
export PROJECT_ROOT="$PWD"   # optional if exarp-go finds .todo2/.exarp from cwd
exarp-go task list --status Todo
exarp-go task show T-123
exarp-go task update T-123 --new-status "In Progress"
```

### Tool invocation

```bash
exarp-go -tool task_workflow -args '{"action":"sync","sub_action":"list","status":"Todo"}'
exarp-go -tool report -args '{"action":"overview","include_metrics":true}'
exarp-go -tool session -args '{"action":"prime","include_tasks":true,"include_hints":true}'
```

OpenCode can run these in Plan or Build mode (e.g. “run exarp-go task list and summarize”) or you can add custom commands that wrap exarp-go.

---

## 3. HTTP API

If you run exarp-go in serve mode, any client (including scripts or tools that OpenCode can call) can use the REST API.

### Start server

```bash
cd /path/to/project
exarp-go -serve :8080
```

### Endpoints (from `internal/api/server.go`)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/tasks` | List tasks |
| GET | `/api/session/prime` | Session prime |
| GET | `/api/report/overview` | Report overview |
| GET | `/api/report/scorecard` | Report scorecard |
| POST | `/api/tools/{name}` | Call any MCP tool by name (body: JSON args) |

Example:

```bash
curl -s http://localhost:8080/api/report/overview
curl -X POST http://localhost:8080/api/tools/task_workflow \
  -H "Content-Type: application/json" \
  -d '{"action":"sync","sub_action":"list","status":"Todo"}'
```

OpenCode doesn’t speak HTTP to exarp-go by default; this is useful if you add a custom tool or script that calls the API (e.g. “get project status” → `curl .../api/report/overview`).

---

## Summary

| Method | Use case |
|--------|----------|
| **MCP** | OpenCode agent uses exarp-go tools directly (tasks, report, session, health, etc.). |
| **CLI** | Run exarp-go from OpenCode’s terminal or from custom commands/scripts. |
| **HTTP** | Scripts or custom tools call exarp-go’s REST API while `exarp-go -serve` is running. |

For tight integration (agent calling tasks, reports, session prime), configure exarp-go as an MCP server in OpenCode’s `mcp` config and set `PROJECT_ROOT` to the project you’re working in.
