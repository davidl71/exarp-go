# Python Migration End State (Phase E)

**Date:** 2026-01-29  
**Purpose:** Single reference for the target end state after Phases A–E of the Python safe removal and migration plan.

---

## 1. Bridge

**Target (achieved):** Only execute_tool.py + mlx path; supporting pieces kept as needed.

| Component | Role |
|-----------|------|
| **bridge/execute_tool.py** | Entry point; routes **mlx** only; imports consolidated_mlx_only so only mlx path loads. |
| **bridge/proto/bridge_pb2.py** | Protobuf for Go ↔ Python bridge protocol. |
| ~~bridge/statistics_bridge.py~~ | **Moved 2026-01-29** to project_management_automation/tools/statistics_bridge.py; bridge folder is MLX-only. |
| ~~bridge/agent_evaluation.py~~ | **Removed 2026-01-29;** AgentScope job removed from agentic-ci.yml. |
| ~~bridge/execute_tool_daemon.py~~ | **Removed 2026-01-29;** pool removed; one-shot execute_tool.py only. |

**Removed in earlier phases:** bridge/context/, bridge/recommend/; task_workflow and context routing dropped from execute_tool.py.

---

## 2. project_management_automation

**Target:** Minimal set for mlx + scripts/tools kept for manual or CI use.

- **Required for bridge mlx:** consolidated_mlx_only.py → consolidated_ai.mlx → mlx_integration.py (and their deps). Bridge no longer loads other consolidated_* modules. **2026-01-29:** All other Python (scripts, resources, utils, non-MLX tools) removed; see docs/PYTHON_DEPRECATED_MLX_ONLY.md.
- **Kept for manual/CI:** Scripts and tools not in Go daily but still used manually or in CI (e.g. automate_check_duplicate_test_names.py). Python automate_daily and daily_automation removed 2026-01-29; use exarp-go -tool automation for daily.
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
