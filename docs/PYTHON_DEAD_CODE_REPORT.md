# Python Dead Code Report

**Date:** 2026-01-27  
**Tool:** `uvx ruff check bridge/ project_management_automation/`  
**Scope:** Unused imports (F401), unused variables (F841), referenced-before-assignment (F823), and related issues.

---

## Summary

| Area | F401 (unused import) | F841 (unused var) | F823 / other | Applied fixes |
|------|----------------------|--------------------|--------------|---------------|
| **bridge/** | 8 | 1 | 0 | ✅ execute_tool.py cleaned |
| **project_management_automation/** | 50+ | 10+ | 1 (session.py) | Not applied (broader change) |

**Bridge changes applied:** Removed unused `base64`, `os`; dropped unused `context.tool.context` import (only `_context_unified` is used); fixed unused exception variable in protobuf parse fallback.

---

## Bridge – Findings and Fixes

### bridge/execute_tool.py ✅ Fixed

- **F401** `base64` imported but unused → **Removed**
- **F401** `os` imported but unused → **Removed**
- **F841** Line 65 `e` in `except Exception as e` never used → **Changed to `except Exception:`**
- **F401** `from context.tool import context as _context` unused (routing uses `_context_unified` only) → **Removed**; kept `from project_management_automation.tools.context_tool import context as _context_unified`.

### bridge/ – Remaining (low priority)

- **execute_resource.py**: Not present; resources are served by Go only (`internal/resources/`).
- **agent_evaluation.py**: ✅ Removed unused `os`. `agentscope.Agent`, `agentscope.Pipeline` unused (optional dep).
- **execute_tool_daemon.py**: ✅ Removed unused `os`, `pathlib.Path`.
- **statistics_bridge.py**: ✅ Removed unused `sys`.
- **context/tool.py**, **recommend/tool.py**: E402 (import not at top) – intentional for path setup.

---

## project_management_automation/ – Edits applied

- **auto_primer.py**: Removed broken `..resources.templates` usage in `get_session_context()` — `resources.templates` does not exist; the block that called `get_task_by_id` and `get_memories_by_category` was dead (would raise ImportError). Function now returns base auto-prime context only.

## project_management_automation/ – Summary (broader cleanup deferred)

Ruff reported 50+ F401 and 10+ F841 across tools and scripts. Notable:

- **F823** `project_management_automation/resources/session.py:265`: `datetime` referenced before assignment – inner `from datetime import datetime` at line 274 makes `datetime` local and shadows the top-level import; line 265 uses it before that. Fix: use the module-level `datetime` and remove the inner import, or restrict the inner import to a narrower block.
- **Unused imports** in consolidated_*.py: `asyncio`, `typing.Any`, `typing.Dict`, `typing.List` in several files.
- **Unused variables**: e.g. `task_desc`, `tasks_updated`, `status`, `source_tasks`, `chip`, `tags_lower`, `critical_count`.

A separate pass can clean these (e.g. `ruff check --fix` or manual edits) after validating tests.

---

## Regression tests (aligned with bridge)

- **Go:** `internal/tools/regression_test.go` — `toolsWithNoBridge` lists fully native tools (no bridge); `knownDifferences` documents hybrid tools (e.g. task_discovery: native first, bridge fallback). See `docs/MIGRATION_TESTING_CHECKLIST.md`.
- **Python:** `tests/unit/python/test_execute_tool.py` — `known_tools` lists tools the bridge actually routes; `task_analysis` removed (native-only, bridge does not route it).

---

## References

- **Safe fallback removal:** `docs/PYTHON_FALLBACKS_SAFE_TO_REMOVE.md`
- **Removal plan:** `docs/SAFE_PYTHON_REMOVAL_PLAN.md`, `docs/PYTHON_CODE_REMOVAL_ANALYSIS.md`
- **Ruff rules:** F401 unused import, F841 unused variable, F823 referenced before assignment, E402 module import not at top.

---

## How to re-run

```bash
uvx ruff check bridge/ project_management_automation/ --output-format=concise
# Auto-fix safe edits (imports/vars):
uvx ruff check bridge/ project_management_automation/ --fix --unsafe-fixes
```
