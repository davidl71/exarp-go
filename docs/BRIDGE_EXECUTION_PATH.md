# Bridge Execution Path

**Date:** 2026-01-29  
**Purpose:** How the Go → Python bridge is invoked and what it runs.

---

## 1. When Go Calls the Bridge

**Only one tool uses the bridge from production Go code:**

| Caller | File | When |
|--------|------|------|
| **mlx** | `internal/tools/mlx_invoke.go` → `InvokeMLXTool()` | When native MLX is unavailable **or** when native succeeds but `action == "generate"` and native returned an error (fallback to bridge for generate). |
| **mlx** | `internal/tools/handlers.go` | `handleMlx` uses native first; on failure for generate (or when native not available), calls `bridge.ExecutePythonTool(ctx, "mlx", params)`. |

**Tests only (not production tool routing):**

- `internal/tools/regression_test.go` — calls `ExecutePythonTool` for health and other tools to compare native vs bridge; not used for normal tool handling.

---

## 2. Go → Python Flow

```
1. Handler (e.g. handleMlx in handlers.go)
   → InvokeMLXTool(ctx, params) in mlx_invoke.go

2. mlx_invoke.go
   → If MLXNativeAvailable(): try handleMlxNative(ctx, params)
   → If native fails or action is "generate" and native errored: bridge.ExecutePythonTool(ctx, "mlx", params)

3. internal/bridge/python.go
   → ExecutePythonTool(ctx, toolName, args)
   → If pool enabled: pool.ExecuteTool(ctx, toolName, args)  [persistent Python process]
   → Else: executePythonToolSubprocess(ctx, toolName, args)

4. Subprocess path (executePythonToolSubprocess):
   → workspaceRoot = getWorkspaceRoot()  (PROJECT_ROOT or executable-relative)
   → bridgeScript = workspaceRoot + "/bridge/execute_tool.py"
   → Build proto.ToolRequest{ ToolName, ArgumentsJson, ProjectRoot, ... }
   → exec: uv run python bridge/execute_tool.py <toolName> --protobuf
   → stdin: protobuf-serialized ToolRequest
   → stdout: protobuf ToolResponse (or JSON on fallback)
   → On protobuf parse failure: retry with JSON args (uv run python bridge/execute_tool.py <toolName> <argsJSON>)
   → Return resp.Result (string) or error
```

---

## 3. Python Side: bridge/execute_tool.py

**Entry:** Script is run as `python execute_tool.py <tool_name> [--protobuf] [args_json]`.

1. **Parse input:** Protobuf (stdin) or JSON (argv). Sets `tool_name`, `args`.
2. **Path setup:** `PROJECT_ROOT = parent of bridge/`, `sys.path.insert(PROJECT_ROOT)`, `sys.path.insert(BRIDGE_ROOT)`.
3. **Imports:** Only:
   - `project_management_automation.tools.consolidated`: `mlx`
4. **Routing:**
   - `if tool_name == "mlx":` → `_mlx(**args)` (Python consolidated).
   - **Else** → "Unknown tool" (JSON or protobuf error), exit 1.

**No longer routed:** task_workflow (Phase B), context, recommend. All other tools are native Go only and never reach the bridge.

5. **Output:** Result dict/string → JSON; if protobuf mode, wrap in `ToolResponse` and write binary to stdout.

---

## 4. Summary

| Step | Location | What happens |
|------|----------|----------------|
| 1 | handlers.go | handleMlx (and any other handler that might call bridge) |
| 2 | mlx_invoke.go | InvokeMLXTool: native first, then `bridge.ExecutePythonTool(ctx, "mlx", params)` |
| 3 | internal/bridge/python.go | ExecutePythonTool: pool or subprocess; build ToolRequest; exec `uv run python bridge/execute_tool.py <tool> --protobuf` |
| 4 | bridge/execute_tool.py | Parse request; route by tool_name (task_workflow, mlx); call consolidated; return JSON/protobuf |
| 5 | project_management_automation.tools.consolidated | task_workflow(), mlx() — actual Python implementations |

**Tools that ever reach the bridge from Go:** only **mlx** (and only when native is unavailable or generate fails). **task_workflow** and **context** are not called from Go; bridge no longer routes them (Phase B / context removal). Direct invocation of the bridge script with tool=task_workflow or tool=context returns "Unknown tool".
