# Python Migration End State (Phase E)

**Date:** 2026-01-29  
**Purpose:** Single reference for the target end state after Phases A–E of the Python safe removal and migration plan.

---

## 1. Bridge

**Target (achieved):** Only execute_tool.py + mlx path; supporting pieces kept as needed.

| Component | Role |
|-----------|------|
| **bridge/execute_tool.py** | Entry point; routes **mlx** only. All other tools are native Go. |
| **bridge/proto/bridge_pb2.py** | Protobuf for Go ↔ Python bridge protocol. |
| **bridge/statistics_bridge.py** | Used by Python tools (e.g. consolidated.mlx); keep while mlx is bridge-routed. |
| **bridge/agent_evaluation.py** | CI/AgentScope evaluation; keep for agentic CI. |
| **bridge/execute_tool_daemon.py** | Optional pool/daemon; keep if daemon mode is used. |

**Removed in earlier phases:** bridge/context/, bridge/recommend/; task_workflow and context routing dropped from execute_tool.py.

---

## 2. project_management_automation

**Target:** Minimal set for mlx + scripts/tools kept for manual or CI use.

- **Required for bridge mlx:** consolidated.mlx and its dependencies (in project_management_automation/tools/ and utils/ as used by mlx).
- **Kept for manual/CI:** Scripts and tools not in Go daily but still used manually or in CI (e.g. automate_check_duplicate_test_names.py, automate_daily.py and its DAILY_TASKS deps). Python automate_daily is deprecated; prefer exarp-go -tool automation.
- **No deletion required for Phase E:** The tree remains larger than “mlx-only” by design; further reduction is optional and would be a separate effort (e.g. remove Python daily orchestrator and unused scripts).

---

## 3. Tests

**Strategy:** Keep Python tests that validate bridge and MCP; migrate or remove as needed later.

| Area | Examples |
|------|----------|
| Bridge | tests/unit/python/test_execute_tool.py, tests/integration/bridge/test_bridge_integration.py |
| Native vs bridge | tests/regression/test_native_vs_bridge.py |
| MCP/server | tests/integration/mcp/test_server_startup.py |

---

## 4. References

- **Plan:** docs/PYTHON_SAFE_REMOVAL_AND_MIGRATION_PLAN.md  
- **Bridge routing:** bridge/execute_tool.py  
- **Scripts audit:** docs/PYTHON_SCRIPTS_AUDIT.md  
