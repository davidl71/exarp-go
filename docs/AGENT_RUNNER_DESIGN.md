# Agent runner abstraction (design)

**Status:** Design  
**Task:** T-1771252286533  
**Summary:** General abstraction for "run an agent in another task" so different runtimes (Cursor CLI, Cloud Agents API, sub-agent, MCP agent) can be plugged in behind one interface.

---

## 1. Goal

- **Single interface** for "run agent with prompt in project context" used by automation, task tools, and TUI.
- **Pluggable implementations:** Cursor CLI (`agent -p`), Cursor Cloud Agents API, future (e.g. sub-agent, other IDEs).
- **No hardcoding** of one runner in automation or handoff flows.

---

## 2. Interface (contract)

```go
// AgentRunner runs an agent with a prompt in a given project context.
// Implementations: Cursor CLI, Cloud Agents API, etc.
type AgentRunner interface {
    // Run runs the agent with the given prompt and options.
    // Returns result summary or error. Implementations may block until done (CLI) or return immediately (API).
    Run(ctx context.Context, opts RunOptions) (*RunResult, error)
}

type RunOptions struct {
    Prompt       string   // Required: prompt or task-derived prompt
    ProjectRoot  string   // Working directory / repo for the agent
    Mode         string   // Optional: "plan" | "ask" | "agent" (Cursor semantics)
    NonInteractive bool   // If true, run in batch (e.g. agent -p "..." ); if false, may open UI
    TaskID       string   // Optional: task ID for attribution/logging
}

type RunResult struct {
    Success   bool   // Whether the run completed successfully
    Runner    string // Implementation name, e.g. "cursor_cli", "cursor_cloud"
    Message   string // Human-readable summary or error
    ExitCode  int    // 0 on success; implementation-specific otherwise
}
```

- **Config:** Which runner to use is chosen by config or env (e.g. `EXARP_AGENT_RUNNER=cursor_cli` or `cursor_cloud`). Default: `cursor_cli` when `agent` is on PATH.
- **Existing code:** Current `cursor run` and child-agent launch (TUI, handoff) use `agentCommand()` and `exec` directly; a first step is to introduce the interface and a **Cursor CLI runner** that wraps that logic, then migrate callers to the runner.

---

## 3. Implementations

| Runner           | Description                    | When to use                          |
|------------------|--------------------------------|--------------------------------------|
| **cursor_cli**   | Exec `agent` (or EXARP_AGENT_CMD) from project root with `-p` or interactive | Default when Cursor CLI is installed; automation, TUI, handoff. |
| **cursor_cloud** | Cursor Cloud Agents API (launch agent with prompt, poll status) | CI/scripting without local Cursor; requires `CURSOR_API_KEY`. |
| *(future)*       | Sub-agent, other IDE CLI       | When other backends are added.       |

---

## 4. Integration points

- **CLI:** `exarp-go cursor run` continues to use the Cursor CLI runner (or the configured runner).
- **Automation:** Optional "run agent" step calls `AgentRunner.Run()` with generated prompt; runner chosen from config.
- **TUI / handoff:** "Run in child agent" uses the configured runner instead of hardcoded `agent` exec.
- **Task run-with-ai:** Local LLM flow (ollama/mlx/fm) remains separate; this abstraction is for *agent* runs (Cursor or Cloud), not for local model calls.

---

## 5. References

- [CURSOR_API_AND_CLI_INTEGRATION.md](CURSOR_API_AND_CLI_INTEGRATION.md) – Cursor CLI and Cloud API
- [GO_AI_ECOSYSTEM.md](GO_AI_ECOSYSTEM.md) – LLM/agent stack overview
- `internal/cli/cursor.go` – current `cursor run` and `agentCommand()`
- `internal/cli/child_agent.go` – current launch-from-TUI/handoff logic
