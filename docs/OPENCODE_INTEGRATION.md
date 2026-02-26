# exarp-go and OpenCode Integration

OpenCode is an open-source AI coding agent (TUI, desktop, IDE). exarp-go can interact with it in three ways: **MCP**, **CLI**, and **HTTP API**. **OpenAgentsControl (OAC)** and other OpenCode-based tools use the same MCP config: add exarp-go under `mcp` in your OpenCode config file (see [OPENAGENTSCONTROL_EXARP_GO_COMBO_PLAN.md](OPENAGENTSCONTROL_EXARP_GO_COMBO_PLAN.md)).

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

### Verification (Phase 2 checklist)

**Validation:** Consider the integration validated when the steps below are done. No separate “validation” build is required; the checklist is the acceptance criteria.

To confirm exarp-go works with OpenCode MCP:

1. **Config:** Add exarp-go to `mcp` in `~/.config/opencode/opencode.json` or project `opencode.json` with `command` and `environment.PROJECT_ROOT` (see [docs/opencode-exarp-go.example.json](opencode-exarp-go.example.json)).
2. **Run OpenCode:** Start OpenCode (e.g. `opencode` or `opencode --agent OpenAgent`) in a repo that has a `.todo2` (or run once so exarp-go can create it).
3. **List tools:** In the agent session, ask to list MCP tools or run a command that uses exarp-go; confirm `task_workflow`, `report`, `session`, `health`, etc. appear.
4. **Call a tool:** Ask the agent to run a task list or report (e.g. “list my Todo2 tasks” or “give me a project overview”). Confirm the agent can invoke the tool and you see task or report output.
5. **Optional CLI from terminal:** In OpenCode’s terminal run `exarp-go task list --status Todo` (with `PROJECT_ROOT` set or from project root). Use `--quiet` or `--json` for script-friendly output.

If any step fails, check binary path, `PROJECT_ROOT`, and OpenCode MCP logs.

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

---

## 4. Plugin (lifecycle hooks + slash commands)

exarp-go ships with an OpenCode plugin (`.opencode/plugins/exarp-go.ts`) that adds lifecycle hooks on top of the MCP integration. The plugin is loaded automatically when OpenCode starts in this project.

### What the plugin provides

| Feature | Hook | Description |
|---------|------|-------------|
| **Auto `PROJECT_ROOT`** | `shell.env` | Injects `PROJECT_ROOT` into every shell command so exarp-go always knows the project root. |
| **System prompt context** | `experimental.chat.system.transform` | Injects current task state into every LLM system prompt — the LLM always knows your tasks without being asked. Cached with 30s TTL; invalidated on `todo.updated` and `session.created` events. |
| **Compaction context** | `experimental.session.compacting` | Injects current task state into compaction summaries so task context survives context resets. |
| **macOS notifications** | `session.idle` event | Sends a desktop notification when a session goes idle. |
| **Slash commands** | `config` hook | Registers `/tasks`, `/prime`, `/scorecard`, `/health` commands. |

### Slash commands

| Command | Description |
|---------|-------------|
| `/tasks` | List current tasks grouped by status |
| `/prime` | Prime session with project context, hints, and handoffs |
| `/scorecard` | Generate and display the project scorecard |
| `/health` | Run project health checks (tools, docs) |

### Setup

The plugin auto-loads from `.opencode/plugins/`. Dependencies are installed by OpenCode (Bun) at startup from `.opencode/package.json`.

If you want to use a custom exarp-go binary path, set `EXARP_GO_BINARY`:

```bash
EXARP_GO_BINARY=/path/to/bin/exarp-go opencode
```

### Plugin + MCP together

The plugin complements MCP — it does not replace it. MCP provides the 35+ tools (task_workflow, report, session, etc.) that the LLM calls directly. The plugin adds environment setup, lifecycle hooks, context persistence across compactions, and convenience slash commands.

---

## 5. MLX + OpenCode (Local Models)

Use local MLX models with OpenCode for on-device inference. Combined with exarp-go MCP, you get tasks, reports, and session prime alongside local LLM planning.

### Setup

1. **Install [OpenCode](https://opencode.ai/):**
   ```bash
   curl -fsSL https://opencode.ai/install | bash
   ```

2. **Install [mlx-lm](https://github.com/ml-explore/mlx-lm):**
   ```bash
   pip install mlx-lm
   ```

3. **Configure OpenCode** — Add the MLX provider and exarp-go MCP to `~/.config/opencode/opencode.json`:
   ```json
   {
     "$schema": "https://opencode.ai/config.json",
     "provider": {
       "mlx": {
         "npm": "@ai-sdk/openai-compatible",
         "name": "MLX (local)",
         "options": { "baseURL": "http://127.0.0.1:8080/v1" },
         "models": {
           "mlx-community/Qwen2.5-Coder-7B-Instruct-8bit": { "name": "Qwen 2.5 Coder" }
         }
       }
     },
     "mcp": {
       "exarp-go": {
         "type": "local",
         "command": ["/path/to/bin/exarp-go"],
         "enabled": true,
         "environment": { "PROJECT_ROOT": "/path/to/your/project" },
         "timeout": 10000
       }
     }
   }
   ```

4. **Start the MLX server:**
   ```bash
   mlx_lm.server
   ```

5. **Connect in OpenCode TUI:** Run `opencode`, then `/connect` → select MLX → API key `none` → choose model.

### Known limitations

- **Tool calling:** Until [mlx-lm#711](https://github.com/ml-explore/mlx-lm/pull/711) lands, tool calling in `mlx_lm.server` is limited or broken.
- **Model support:** Tool-calling support varies by model. Check mlx-lm issues for your model.
- **exarp-go tools:** exarp-go MCP works independently of the LLM provider; tasks, report, and session tools are available regardless of MLX status.

Reference: [OpenCode with MLX gist](https://gist.github.com/awni/93a973a0cf5fb539b2ce1f37ec4a9989).

### Skip LSP: MCP is sufficient

For OpenCode and code intelligence, **MCP (exarp-go) is sufficient** — you do not need a separate LSP (Language Server Protocol) setup. exarp-go provides tools, prompts, and resources; OpenCode’s built-in language support plus exarp-go’s `task_workflow`, `report`, `session`, and `health` tools cover task management and project context. Add LSP only if you need editor-specific diagnostics or completions beyond what OpenCode already offers.

---

## 6. OpenCode and Cursor (ACP)

OpenCode [supports the Agent Client Protocol (ACP)](https://opencode.ai/docs/acp/): you run `opencode acp` and your editor talks to OpenCode over JSON-RPC via stdio. Editors that support ACP (Zed, JetBrains, Avante.nvim, CodeCompanion.nvim) can add OpenCode as an agent server; see [OpenCode ACP docs](https://opencode.ai/docs/acp/) for their config snippets.

### Can Cursor use OpenCode via ACP?

**Cursor does not currently support configuring an external ACP agent** the way Zed does (e.g. `agent_servers` → `command: "opencode", args: ["acp"]`). So you cannot “point Cursor at OpenCode” in Cursor settings and have Cursor’s chat use OpenCode as the backend.

The existing **Cursor ↔ ACP** projects do the **reverse**:

- **[cursor-acp](https://github.com/roshan-c/cursor-acp)** and **[cursor-agent-acp](https://github.com/blowmage/cursor-agent-acp-npm)** expose the **Cursor CLI agent** over ACP. That lets ACP clients (Zed, JetBrains, Emacs, Neovim) use **Cursor** as their agent, not the other way around.

### Practical integration today

| Goal | Approach |
|------|----------|
| **Use OpenCode as the agent** | Use an ACP-capable editor (Zed, JetBrains, Neovim). In that editor, add OpenCode as the ACP agent: `command: "opencode", args: ["acp"]` (see [OpenCode ACP](https://opencode.ai/docs/acp/)). Add exarp-go as MCP in OpenCode’s config so OpenCode can call task_workflow, report, session. |
| **Use Cursor and OpenCode on the same project** | Use **Cursor** for editing + Cursor’s built-in agent + exarp-go MCP. Use **OpenCode** (TUI or inside Zed/JetBrains/Neovim via ACP) in the same repo with exarp-go MCP. Same `.todo2` / `PROJECT_ROOT` = shared tasks and reports. |
| **Use Cursor’s agent from another editor** | Install [cursor-acp](https://github.com/roshan-c/cursor-acp) or [cursor-agent-acp](https://github.com/blowmage/cursor-agent-acp-npm), then in Zed (or another ACP client) add “Cursor” as an agent that runs the adapter. |

### If Cursor adds ACP agent support later

If Cursor gains a setting like “external agent server” (similar to Zed’s `agent_servers`), the config would look like:

```json
{
  "agent_servers": {
    "OpenCode": {
      "command": "opencode",
      "args": ["acp"]
    }
  }
}
```

Until then, use an ACP-capable editor when you want OpenCode as the agent, and use Cursor with its own agent + exarp-go MCP when you want Cursor; both can share exarp-go for tasks and reports.
