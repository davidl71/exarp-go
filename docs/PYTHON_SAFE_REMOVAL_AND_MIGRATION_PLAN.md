# Python Safe Removal and Migration Plan

**Date:** 2026-01-29  
**Purpose:** What can be safely removed now; plan to migrate or deprecate the rest.

---

## 1. Safe to Remove Now

### 1.1 `bridge/context/` ✅ Removed this session

| Item | Reason |
|------|--------|
| **`bridge/context/`** (__init__.py, summarizer.py, tool.py) | **Unused.** Go uses only native context (handleContext → summarize/budget/batch). execute_tool.py routes "context" to project_management_automation.tools.context_tool, not to bridge/context. No code imports bridge.context. |

**Follow-up:** Remove the "context" branch from execute_tool.py (so bridge returns "Unknown tool" for context). Update project_management_automation/tools/context_tool.py to return a clear JSON error: "Context tool is native Go only; use exarp-go -tool context." so any direct Python caller gets a consistent message.

### 1.2 Already removed (prior sessions)

- **`bridge/recommend/`** — Removed; recommend is full native Go; bridge never routed to it.
- **Python `tool_count_health`** — Removed from DAILY_TASKS and todo2 alignment fallback; Go health action=tools used instead.

---

## 2. What Remains (Do Not Remove Yet)

### 2.1 Bridge (required for mlx and optional task_workflow path)

| Path | Used by Go? | Used by bridge script? | Safe to remove? |
|------|-------------|------------------------|-----------------|
| **bridge/execute_tool.py** | Yes (only for mlx when native unavailable) | — | No |
| **bridge/execute_tool_daemon.py** | Optional pool | — | No (used if daemon mode) |
| **bridge/context/** | No | No | ✅ Removed |
| **bridge/agent_evaluation.py** | No | No (CI/AgentScope) | No |
| **bridge/statistics_bridge.py** | No | Yes (Python tools import it) | No |
| **bridge/proto/bridge_pb2.py** | Yes (Go can send protobuf) | Yes | No |

After removing bridge/context, bridge still has: execute_tool, execute_tool_daemon, agent_evaluation, statistics_bridge, proto.

### 2.2 Bridge routing (execute_tool.py)

| tool_name | Handler | Go uses bridge? | Remove branch? |
|-----------|---------|-----------------|----------------|
| **task_workflow** | _task_workflow (consolidated) | No (Go native only) | Optional: remove → "Unknown tool" from bridge |
| **mlx** | _mlx (consolidated) | Yes (when native generate unavailable) | No |
| **context** | _context_unified (context_tool) | No (Go native only) | **Yes** (and remove bridge/context) |

### 2.3 project_management_automation/ (100+ files)

- **Required for bridge:** consolidated (task_workflow, mlx), context_tool (stub after removal).
- **Required for Python automation:** scripts/automate_daily.py, tools/daily_automation.py, and all scripts/tools referenced by DAILY_TASKS or by tools that bridge might call.
- **Not safe to remove in isolation:** Any script or tool that is imported by another Python module or by execute_tool.py. See PYTHON_SCRIPTS_AUDIT.md.

---

## 3. Migration Plan for the Rest

### Phase A — Done / This session

1. Remove **bridge/context/** and drop context from bridge routing; make context_tool return native-only message.
2. Document remaining Python and this plan.

### Phase B — Shrink bridge surface ✅ Done 2026-01-29

1. **Remove task_workflow from bridge routing** ✅  
   Removed task_workflow branch and import from execute_tool.py. Bridge now only imports and routes **mlx**; task_workflow returns "Unknown tool" from bridge.

2. **Keep mlx bridge path**  
   Go calls the bridge for mlx when native generate is unavailable. Keep consolidated.mlx and bridge routing for "mlx" until mlx generate is fully native or deprecated.

### Phase C — Python automation path (larger decision) ✅ Done 2026-01-29

1. **Deprecate Python automation** ✅  
   - Document: "For daily/sprint automation, use exarp-go -tool automation (Go). Python automate_daily is deprecated."  
   - Added to README (Automation section), AGILE_DAILY_STANDUP_IMPLEMENTATION.md, and this plan.  
   - Do not delete project_management_automation/ yet; keep for mlx and any direct Python tool use.

2. **Optional: remove Python daily orchestrator** (deferred)  
   - Stop recommending `python -m project_management_automation.scripts.automate_daily` — done in docs.  
   - Removing automate_daily.py and its callers (e.g. daily_automation.py) would allow removing many automate_*.py scripts and their tool implementations in one go.  
   - **Recommendation:** Defer until Go automation fully replaces all desired Python daily tasks and duplicate_test_names is either migrated to Go or dropped.

### Phase D — Duplicate test names and one-offs ✅ Done 2026-01-29

1. **duplicate_test_names** — **Decision: (b) Drop from daily, run manually if needed.**  
   - Go daily automation does not include duplicate_test_names (stays at 8 tasks).  
   - No native Go check implemented; Python script remains available for manual or CI use.  
   - To run manually: `uv run python -m project_management_automation.scripts.automate_check_duplicate_test_names` (or `python3 -m ...`).  
   - Optional future: (a) implement native Go check and add to automation daily, or (c) keep Python-only.

2. **Other Python-only tools**  
   - Tools in consolidated or project_management_automation that are never called by Go (e.g. docs_health, todo2_alignment when invoked via bridge): **leave as-is** for direct Python use.  
   - Prefer **exarp-go** for tools that have Go equivalents (see NATIVE_GO_HANDLER_STATUS.md). No bridge routing changes in Phase D.

### Phase E — End state (optional long term)

- **Bridge:** Only execute_tool.py + mlx path (+ proto, statistics_bridge, agent_evaluation as needed).
- **project_management_automation:** Minimal set for mlx (consolidated.mlx and its deps) + any scripts/tools we choose to keep for manual or CI use.
- **Tests:** Keep Python tests that validate bridge and MCP; migrate or remove as needed.

---

## 4. Summary Table

| Action | Item | When |
|--------|------|------|
| **Remove** | bridge/context/ | This session |
| **Update** | execute_tool.py: drop "context" branch | This session |
| **Update** | context_tool.py: return native-only message | This session |
| **Done** | execute_tool.py: drop "task_workflow" branch | Phase B ✅ |
| **Keep** | bridge + consolidated.mlx, statistics_bridge, proto, agent_evaluation | Until mlx generate is native or deprecated |
| **Deprecate (doc only)** | Python automate_daily | Phase C ✅ |
| **Done** | duplicate_test_names: drop from daily, run manually (Phase D ✅) | Phase D ✅ |

---

## 5. References

- **Bridge routing:** bridge/execute_tool.py  
- **Go handler status:** docs/NATIVE_GO_HANDLER_STATUS.md  
- **Scripts audit:** docs/PYTHON_SCRIPTS_AUDIT.md  
- **Remaining analysis:** docs/PYTHON_REMAINING_ANALYSIS.md  
