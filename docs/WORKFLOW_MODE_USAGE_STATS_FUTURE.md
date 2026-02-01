# Workflow Mode Usage Statistics — Future Implementation

**Status:** Documented for future implementation; no active Todo2 tasks.

**Tag hints:** `#feature` `#workflow_mode` `#tools`

This document describes how to fully implement per-tool usage statistics for the `workflow_mode` stats action. The current stats response includes the note: *"Stats show current mode state; per-tool usage counts are not tracked."* This doc is the reference for implementing real usage tracking when needed.

---

## 1. Purpose & Success Criteria

**Purpose:** Record every tool invocation (CLI and MCP), persist counts, and return `usage_by_tool` (and optional `usage_total`) from `workflow_mode` action=stats.

**Success criteria:**

- Every `CallTool` (CLI and MCP) is recorded without changing vendored mcp-go-core.
- Usage state is persisted (same file as mode state or separate `.exarp/workflow_usage.json`).
- `workflow_mode` action=stats returns `usage_by_tool` (tool name → count) and an updated note.
- Tests cover recording and stats response.

---

## 2. Technical Approach

### Single recording point

A **server wrapper** that implements `framework.MCPServer` and overrides only `CallTool`:

1. Call `RecordWorkflowUsage(name)` (or `getWorkflowManager().RecordUsage(name)`).
2. Delegate to `inner.CallTool(ctx, name, args)`.

No changes to vendor; all invocations (CLI and MCP) go through the same wrapped server.

### Where to wrap

- **MCP:** In `cmd/server/main.go`, after `factory.NewServerFromConfig(cfg)`, wrap with `tools.WrapServerWithUsageRecording(server)` before `RegisterAllTools(server)`.
- **CLI:** In `internal/cli/cli.go`, inside `setupServer()`, after creating the server, wrap before returning so all `-tool` and task flows are recorded.

### State and persistence

- **WorkflowModeManager** (in `internal/tools/workflow_mode.go`) gains:
  - `UsageCounts map[string]int` (tool name → total count).
  - Optional: `UsageByMode` or `UsageHistory` for “by mode” or “last 24h” later.
  - `RecordUsage(toolName string)` — increment and persist.
  - `GetUsageStats()` — return counts (and optional aggregates) for the stats response.
- **Persistence:** Either extend `.exarp/workflow_mode.json` with usage fields or use a separate `.exarp/workflow_usage.json` (with optional trimming by time/count).

---

## 3. Implementation Steps (Checklist)

When implementing, use this order:

| Step | Where | What |
|------|--------|------|
| 1 | `internal/tools/workflow_mode.go` | Add usage fields (`UsageCounts`, optional `UsageByMode` / `UsageHistory`), `RecordUsage(toolName)`, `GetUsageStats()`, load/save (same file or `.exarp/workflow_usage.json`). |
| 2 | `internal/tools/workflow_mode.go` or new `usage_recording.go` | Add `WrapServerWithUsageRecording(server framework.MCPServer) framework.MCPServer`: wrapper type that overrides `CallTool` to record then delegate. |
| 3 | `cmd/server/main.go` | After creating the server, wrap it with `WrapServerWithUsageRecording(server)` before `RegisterAllTools(server)`. |
| 4 | `internal/cli/cli.go` | In `setupServer()`, after creating the server, wrap it with `WrapServerWithUsageRecording(server)` before returning. |
| 5 | `internal/tools/workflow_mode.go` | In `handleWorkflowModeStats`, call `GetUsageStats()` and add `usage_by_tool` (and optional `usage_total`) to the result; update or remove the “not tracked” note. |
| 6 | Tests | Unit test: `RecordUsage` increments count; `GetUsageStats` returns `usage_by_tool`. Optional: test that `CallTool` through the wrapped server records; test that stats include `usage_by_tool` when usage exists. |

---

## 4. Optional Extensions

- **By mode:** Pass current mode into `RecordUsage` (e.g. from `getWorkflowManager().state.CurrentMode`) and store `UsageByMode`; expose in stats.
- **Last 24h:** Append to `UsageHistory` in `RecordUsage`, trim by time/count when loading/saving; add something like `usage_last_24h` in stats.
- **Avoid inflating workflow_mode:** In the wrapper, either record every call (including `workflow_mode` itself) or skip recording when `name == "workflow_mode"`.

---

## 5. References

- Current stats: `internal/tools/workflow_mode.go` — `handleWorkflowModeStats`
- Server creation: `cmd/server/main.go`, `internal/cli/cli.go` — `setupServer()`
- CallTool: `vendor/github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk/adapter.go` — `CallTool`
- Framework interface: `vendor/github.com/davidl71/mcp-go-core/pkg/mcp/framework/server.go` — `MCPServer`
- Plan file (high-level): `.cursor/plans/workflow-mode-usage-stats.plan.md`
