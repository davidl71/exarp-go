# Python Scripts Audit (project_management_automation/scripts/)

**Date:** 2026-01-29  
**Purpose:** Identify dead scripts and what calls them.

---

## Summary

**No scripts are completely dead.** Every `automate_*.py` script is imported by at least one tool in `project_management_automation/tools/`. exarp-go’s **Go automation does not run any of these scripts** (it uses native tools). They are only used when:

1. A Python tool is invoked via the bridge (e.g. `docs_health`, `todo2_alignment`), or  
2. Someone runs the Python daily orchestrator: `python3 -m project_management_automation.scripts.automate_daily`.

---

## Script → Caller Map

| Script | Imported by (tool) | In automate_daily.py DAILY_TASKS? |
|--------|--------------------|-----------------------------------|
| `automate_attribution_check.py` | `tools/attribution_check.py` | No |
| `automate_automation_opportunities.py` | `tools/automation_opportunities.py` | No |
| `automate_check_duplicate_test_names.py` | None (only run by `automate_daily.py`) | Yes (run as sub-script) |
| `automate_daily.py` | `tools/daily_automation.py` | — (orchestrator) |
| `automate_dependency_security.py` | `tools/dependency_security.py` | Yes |
| `automate_docs_health_v2.py` | `tools/docs_health.py` | Yes |
| `automate_external_tool_hints.py` | `tools/external_tool_hints.py` | No (one-time; removed from daily) |
| `automate_run_tests.py` | `tools/run_tests.py` | No |
| `automate_sprint.py` | `tools/sprint_automation.py` | No |
| `automate_stale_task_cleanup.py` | `tools/stale_task_cleanup.py` | Yes |
| `automate_test_coverage.py` | `tools/test_coverage.py` | No |
| `automate_todo_sync.py` | `tools/todo_sync.py` | No |
| `automate_todo2_alignment_v2.py` | `tools/todo2_alignment.py` | Yes |
| `automate_todo2_duplicate_detection.py` | `tools/duplicate_detection.py` | Yes |

---

## How exarp-go Runs Automation

- **Makefile:** `make sprint`, `make sprint-start`, `make pre-sprint` → call **Go binary** `exarp-go -tool automation -args '{"action":"sprint",...}'`.
- **Go automation** (`internal/tools/automation_native.go`): runs **native tools** only (e.g. `analyze_alignment`, `task_analysis`, `health`, `report`, `memory_maint`). It does **not** invoke any Python script or the bridge for daily/sprint.

So the Python scripts are **bypassed** when using exarp-go’s normal automation path; they are **only** used when the Python tools (or the Python daily orchestrator) are run.

---

## Safe to Remove?

- **None** of the scripts can be removed without also removing or rewriting the tool that imports them (or that orchestration path in `automate_daily.py`).
- **Optional reduction:** If we explicitly drop the “Python automation” path (no bridge for these tools, no `automate_daily` usage), we could then remove scripts and their callers together; that’s a larger change and should be a deliberate decision.

---

## Recommendation

- **Do not delete any script** for now; treat them as library code for the Python tools.
- **Document:** Go automation = native only; Python scripts = used only when Python tools or `automate_daily` are run.
- **Optional later:** Add a single “Python automation deprecated” note and a task to migrate remaining Python tool callers to Go, then remove scripts and tools together.
