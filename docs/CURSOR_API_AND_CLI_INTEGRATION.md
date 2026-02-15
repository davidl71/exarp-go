# Cursor API and CLI Integration

Summary of Cursor API/CLI documentation and suggested improvements to the exarp-go CLI to utilize them.

**References:** [Cursor APIs](https://cursor.com/docs/api) · [Cursor CLI](https://cursor.com/docs/cli/overview) · [Cloud Agents](https://cursor.com/docs/cloud-agent)

## Task IDs (Todo2)

Use `exarp-go task show T-<id>` to view details.

| Task ID | Summary |
|---------|--------|
| T-1771164528145 | Cursor CLI: add `cursor run` subcommand |
| T-1771164549623 | Session prime/handoff: Cursor CLI suggestion |
| T-1771164549862 | Automation: optional Cursor agent step |
| T-1771164550717 | MCP tool: cursor_cloud_agent (Cloud Agents API) |
| T-1771164551664 | Docs: Cursor integration help and CLI section |
| T-1771164552852 | Local AI task assignment (MLX, AFM, Ollama) |

---

## 1. Cursor API Overview

### Available APIs

| API | Description | Availability |
|-----|-------------|--------------|
| **Cloud Agents API** | Create/manage AI coding agents; launch agents with prompts, monitor status, add follow-ups, list/delete agents | **Beta, All Plans** |
| Admin API | Team members, settings, usage, spending | Enterprise |
| Analytics API | Usage, AI metrics, model usage | Enterprise |
| AI Code Tracking API | AI-generated code attribution | Enterprise |

### Authentication

- All APIs: **Basic Auth** with API key (key as username, empty password).
- **Cloud Agents API key:** [Cursor Dashboard → Integrations](https://cursor.com/dashboard?tab=integrations).
- Env var: `CURSOR_API_KEY` or flag `agent --api-key KEY`.

### Rate limits

- Cloud Agents API: standard rate limiting (per team).
- 429 response when exceeded; retry after backoff.

### Cloud Agents API capabilities (from docs)

- **Launch agents** – POST with prompt (+ optional repo/images); agent works on repo.
- **Monitor status** – CREATING / RUNNING / FINISHED / FAILED / CANCELLED.
- **Add follow-ups** – Send more instructions to a running agent.
- **List agents** – List cloud agents.
- **List repositories** – Repos the agent can use.
- **Delete agents** – Remove an agent.

Detailed endpoints: [Cloud Agents API](https://cursor.com/docs/cloud-agent/api/endpoints).

---

## 2. Cursor CLI (“agent” commands)

### Install

```bash
# macOS, Linux, WSL
curl https://cursor.com/install -fsS | bash
# Add to PATH: export PATH="$HOME/.local/bin:$PATH"
agent --version
```

### Main commands

| Command | Purpose |
|---------|---------|
| `agent` | Interactive session |
| `agent "prompt"` | Interactive with initial prompt |
| `agent -p "prompt"` | Non-interactive (scripts/CI) |
| `agent ls` | List previous chats |
| `agent resume` | Resume latest conversation |
| `agent --resume="chat-id"` | Resume specific chat |
| `agent mcp list` | List MCP servers |
| `agent mcp list-tools <server>` | List tools for a server |

### Modes

- `--mode=agent` (default) – full tool access.
- `--mode=plan` – design approach first.
- `--mode=ask` – read-only.

### Non-interactive (for automation)

```bash
agent -p "find and fix performance issues" --model "gpt-5.2"
agent -p "review these changes for security" --output-format text
export CURSOR_API_KEY=...   # or --api-key for CI
```

---

## 3. Suggested Improvements to exarp-go CLI

### 3.1 Cursor CLI wrapper subcommand (high value) — **Task:** T-1771164528145

**Idea:** Add a subcommand that invokes the Cursor CLI `agent` with a prompt derived from exarp-go context (e.g. current task or plan).

**Examples:**

```bash
# Run Cursor agent with prompt from task T-123
exarp-go cursor run T-123
# or
exarp-go cursor run --task T-123 --mode plan

# Run with custom prompt (still from project root so MCP tools apply)
exarp-go cursor run -p "Implement the proto refactor for task_workflow response types"

# Non-interactive (delegate to agent -p)
exarp-go cursor run -p "Review Todo2 backlog and suggest next task" --no-interactive
```

**Implementation sketch:**

- New top-level command: `cursor` (e.g. via `mcpcli.ParseArgs` → `case "cursor": handleCursorCommand(parsed)`).
- Subcommands: `run` (and later `cloud` if Cloud API is added).
- `cursor run [task-id]`:
  - If `task-id` given: load task from Todo2, build prompt from name + description.
  - Optional `--mode plan|ask|agent`, `-p "prompt"`, `--no-interactive`.
  - Resolve project root; exec `agent` (or `agent -p "..."`) in that directory so MCP (including exarp-go) is available.
- Detect `agent` on PATH; if missing, print install instructions and exit.
- Optional: read `CURSOR_API_KEY` from env for non-interactive runs (pass-through to `agent`).

**Files to touch:** `internal/cli/cli.go` (dispatch, new handler), new `internal/cli/cursor.go` (cursor run logic, exec `agent`).

---

### 3.2 Session prime / handoff: suggest Cursor CLI command (medium value) — **Task:** T-1771164549623

**Idea:** When session prime or handoff returns “suggested next tasks”, also return a ready-to-run Cursor CLI command for the first task so the user (or a script) can run it.

**Example addition to session prime / handoff JSON:**

```json
"cursor_cli_suggestion": "agent -p \"Work on T-123: Proto Task workflow response types\" --mode=plan"
```

**Implementation:**

- In `internal/tools/session.go`, when building prime or handoff response with `suggested_next` tasks, compute a single `cursor_cli_suggestion` string from the first suggested task (e.g. `agent -p "Work on <id>: <name>" --mode=plan`).
- Optional: add a `--mode` hint (plan vs agent) based on task complexity or tags.
- Document in session tool description that clients can use this for “run in Cursor” one-click or script integration.

---

### 3.3 Automation tool: optional “cursor agent” step (medium value) — **Task:** T-1771164549862

**Idea:** Allow automation (e.g. daily/sprint) to optionally delegate one step to the Cursor CLI agent with a generated prompt (e.g. “Review and suggest next actions for these tasks”).

**Example:**

```bash
exarp-go -tool automation -args '{"action":"daily","use_cursor_agent":true,"cursor_agent_prompt":"Review the backlog and suggest which task to do next"}'
```

**Implementation:**

- Add optional params to automation: `use_cursor_agent` (bool), `cursor_agent_prompt` (string, or default prompt).
- In `handleAutomationDaily` (or shared helper), if `use_cursor_agent` and `agent` on PATH:
  - Build prompt (e.g. include summary of tasks from report/task_workflow).
  - Run `agent -p "..."` in project root (non-interactive); capture stdout/stderr and attach to automation result as a step.
- If `agent` not found or Cursor not desired, skip step and continue (no hard dependency on Cursor).

---

### 3.4 New MCP tool: “cursor_cloud_agent” (optional, Cloud Agents API) — **Task:** T-1771164550717

**Idea:** For teams using Cursor Cloud Agents API, expose a tool that launches or queries cloud agents so Cursor can be driven from exarp-go (e.g. from automation or from another agent).

**Example tool schema:**

- `action`: `launch` | `status` | `list` | `follow_up` | `delete`
- `launch`: `prompt`, optional `repo`, optional `model`
- `status` / `follow_up` / `delete`: `agent_id`
- Auth: require `CURSOR_API_KEY` (or config) for API calls.

**Implementation:**

- New handler in `internal/tools/` that:
  - Reads API key from env or config.
  - Calls Cloud Agents API (https://api.cursor.com/... – exact base URL and paths from [Cloud Agents API docs](https://cursor.com/docs/cloud-agent/api/endpoints)).
  - Returns launch result, status, or list as structured JSON.
- Register as tool `cursor_cloud_agent` (or `cursor_agent`).
- Document in tool description: Beta, requires Cursor account and API key from Dashboard → Integrations.

---

### 3.5 Documentation and help (low effort) — **Task:** T-1771164551664

- **In `showUsage()` / help:** Add a “Cursor integration” section, e.g.:
  - “To run Cursor agent with task context: `exarp-go cursor run T-123` (when implemented)” or “Run Cursor agent from project root so it can use exarp-go MCP tools.”
- **Docs:** In `docs/CURSOR_MCP_SETUP.md` or a new “Cursor CLI + exarp-go” section:
  - Install Cursor CLI; run `agent` from project root so exarp-go MCP is available.
  - Optional: set `CURSOR_API_KEY` for non-interactive/CI use.
  - Reference this document for API/CLI summary and improvement roadmap.

### 3.6 Local AI task assignment (MLX, AFM, Ollama) — **Task:** T-1771164552852

**Idea:** Extend exarp-go so tasks can be assigned to or executed with **local** AI backends (MLX, Apple Foundation Models, Ollama) for estimation, summarization, or lightweight execution—without requiring Cursor or cloud agents.

**Options:**

1. **Task metadata:** Add optional `preferred_backend` (or `local_ai_backend`) to task metadata: `mlx` | `fm` | `ollama`. When present, tools that use LLMs (e.g. estimation, text_generate, report insights) can respect it.
2. **CLI / tool params:** For `task_workflow`, `estimation`, or a new “run task with local AI” flow, accept `--local-ai-backend mlx|fm|ollama` (or JSON param `local_ai_backend`) and pass it through to existing `text_generate` / `ollama` / `mlx` / `apple_foundation_models` tools.
3. **Lightweight execution:** Optional subcommand or tool, e.g. `exarp-go task run-with-ai T-123 --backend ollama`, that builds a prompt from the task (name + description), calls the chosen local backend via existing LLM tools, and returns or appends the result (e.g. as a comment or summary)—no code edits, just model output for planning/summarization.

**Implementation sketch:**

- **Task store:** Allow optional field on tasks (e.g. in Todo2 custom fields or tags like `#backend:ollama`) for `preferred_backend`. Document in task_workflow schema.
- **Estimation / text_generate:** When estimation (or other tools) invokes LLM, check task’s `preferred_backend` or explicit param and route to `ollama` / `mlx` / `apple_foundation_models` via existing handlers.
- **New flow (optional):** `task run-with-ai <task-id> [--backend mlx|fm|ollama]` or MCP tool `task_local_ai` that: loads task, builds prompt, calls `text_generate` (or ollama/mlx) with that backend, returns text. Keeps all execution local.

**Backends:** Reuse existing tools: `text_generate` (provider `fm`|`mlx`|`auto`), `ollama`, `mlx`, `apple_foundation_models` per [LLM tools rule](.cursor/rules/llm-tools.mdc). No new LLM integrations required—only wiring task context and backend preference into existing tools.

---

## 4. Priority summary

| Improvement | Task ID | Effort | Value | Note |
|-------------|---------|--------|--------|------|
| **cursor run** subcommand | T-1771164528145 | Medium | High | Direct “task → Cursor agent” flow; reuses existing task/plan context. |
| Session prime/handoff Cursor CLI suggestion | T-1771164549623 | Low | Medium | Better UX and scriptability. |
| Automation optional Cursor agent step | T-1771164549862 | Medium | Medium | Good for “daily/sprint + Cursor review” workflows. |
| **cursor_cloud_agent** MCP tool | T-1771164550717 | Medium–High | Medium | For teams using Cloud Agents API; depends on API stability (Beta). |
| Docs and help text | T-1771164551664 | Low | Low | Improves discoverability. |
| **Local AI task assignment (MLX, AFM, Ollama)** | T-1771164552852 | Medium | High | Assign/run tasks with local backends; extends existing ollama/mlx/afm tools. |

---

## 5. References

- [Cursor APIs Overview](https://cursor.com/docs/api)
- [Cursor CLI Overview](https://cursor.com/docs/cli/overview)
- [Cursor CLI Reference (parameters, auth)](https://cursor.com/docs/cli/reference/parameters)
- [Cloud Agents](https://cursor.com/docs/cloud-agent)
- [Cloud Agents API Endpoints](https://cursor.com/docs/cloud-agent/api/endpoints)
- [Cursor CLI MCP](https://cursor.com/docs/cli/mcp) – MCP server management from CLI
