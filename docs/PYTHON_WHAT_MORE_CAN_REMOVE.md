# What More Can We Remove? (Post Phase E)

**Date:** 2026-01-29  
**Context:** Phases A–E are done. Bridge routes **mlx** only; Go automation is primary. This doc answers what Python/bridge code can be removed next and under what conditions.

---

## 1. Do Not Remove (Required)

| Item | Reason |
|------|--------|
| **bridge/execute_tool.py** | Entry point; routes mlx. Go calls it when native mlx generate is unavailable. |
| **bridge/proto/bridge_pb2.py** | Protobuf for Go ↔ Python bridge. |
| **project_management_automation/tools/consolidated.py** | Re-exports; bridge imports `mlx` from here. |
| **project_management_automation/tools/consolidated_ai.py** | Defines `mlx()` used by bridge. |
| **project_management_automation/tools/mlx_integration.py** | MLX status/hardware/models/generate; used by consolidated_ai.mlx. |
| **project_management_automation/utils/** (used by mlx_integration) | e.g. project_root, logging; mlx_integration may use them. |
| **bridge/statistics_bridge.py** | Imported by task_duration_estimator.py, which is used by estimation, automation, task_clarity, mlx_task_estimator, coreml_task_estimator. Removing it would break those tools for direct Python/manual use. |

---

## 2. Optional Removals (Only If You Accept the Trade-off)

### 2.1 Python daily orchestrator only (no other scripts) ✅ Done 2026-01-29

**Removed:** `project_management_automation/scripts/automate_daily.py` and `project_management_automation/tools/daily_automation.py`. `consolidated_automation.py` action=daily now returns “Daily automation is native Go only; use exarp-go -tool automation”.

**Effect:** Python daily is no longer available; use `exarp-go -tool automation -args '{"action":"daily"}'` only. All other scripts (e.g. duplicate_test_names) remain for manual/CI use.

---

### 2.2 Bridge daemon (execute_tool_daemon.py)

**Remove:** `bridge/execute_tool_daemon.py`.

**Condition:** Only if you also disable or remove the **Go bridge pool** that uses it (`internal/bridge/pool.go` uses `execute_tool_daemon.py`). If the pool is disabled, Go will fall back to one-shot `execute_tool.py` per call.

**Effect:** No persistent Python daemon for bridge calls; more process spawns when mlx is invoked via bridge.

**Trade-off:** Simpler bridge, slightly higher latency for repeated mlx bridge calls.

---

### 2.3 AgentScope CI script (agent_evaluation.py)

**Remove:** `bridge/agent_evaluation.py`.

**Condition:** Only if you remove or replace the AgentScope step in **agentic CI** (`.github/workflows/agentic-ci.yml` references `python bridge/agent_evaluation.py`).

**Effect:** AgentScope-based evaluation in CI would need another implementation or be dropped.

**Trade-off:** One less Python script in the repo; CI behavior changes.

---

### 2.4 context_tool stub

**Remove:** `project_management_automation/tools/context_tool.py`.

**Condition:** Accept that any **direct Python** caller that does `from project_management_automation.tools.context_tool import context` (or similar) will get an ImportError. The bridge already does not route `context`; it returns “Unknown tool”.

**Effect:** Slightly smaller tree; direct Python users of context must use exarp-go or another path.

**Trade-off:** Low impact if no one calls context from Python directly.

---

## 3. What We Should Not Remove (Without a Bigger Change)

- **Individual automate_*.py scripts** (e.g. `automate_docs_health_v2.py`, `automate_todo2_alignment_v2.py`): Each is imported by a tool module (e.g. `docs_health.py`, `todo2_alignment.py`). Removing a script alone would break that tool. To remove scripts you must either remove or rewrite the tool that imports it (and any caller of that tool).
- **consolidated_*.py modules other than consolidated_ai**: The bridge only imports `mlx` from `consolidated`, but `consolidated.py` re-exports from all consolidated_* modules. Removing one of those modules would break the `consolidated` re-exports and any direct Python caller that uses those tools. So “minimal mlx-only” would require a **slim consolidated** that only imports from consolidated_ai (and its mlx chain); that’s a refactor, not a simple deletion.
- **statistics_bridge.py**: Used by `task_duration_estimator.py`, which is used by estimation, automation, task_clarity, mlx_task_estimator, coreml_task_estimator. Required for those tools to keep working from Python.

---

## 4. Summary

| Goal | What to remove | Condition |
|------|----------------|-----------|
| **Stop Python daily entirely** | `automate_daily.py` + `daily_automation.py` | ✅ Done 2026-01-29. Use Go automation only. |
| **Drop bridge daemon** | `execute_tool_daemon.py` | Disable/remove Go pool in `internal/bridge/pool.go`. |
| **Drop AgentScope from CI** | `agent_evaluation.py` | Change/remove AgentScope step in agentic-ci.yml. |
| **Trim direct Python context** | `context_tool.py` | Accept breaking direct Python callers of context tool. |
| **Further reduction** | Many scripts + their tool callers | Bigger refactor: e.g. “mlx-only” consolidated + remove or replace tools that use the scripts. |

**Done:** **(2.1) Python daily orchestrator** removed 2026-01-29. Daily automation is exarp-go only; duplicate_test_names and other checks stay available via manual scripts/tools.

---

## 5. References

- **Plan:** docs/PYTHON_SAFE_REMOVAL_AND_MIGRATION_PLAN.md  
- **End state:** docs/PYTHON_MIGRATION_END_STATE.md  
- **Scripts audit:** docs/PYTHON_SCRIPTS_AUDIT.md  
