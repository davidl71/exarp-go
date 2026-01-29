# Python Remaining – Plan and Task Mapping

**Date:** 2026-01-29  
**Context:** After removing prompt_discovery.py, prompts.py stub, unused imports, and broken templates usage in auto_primer.

---

## What’s Left in Python

### 1. Bridge (minimal surface)

| Item | Purpose | Safe to remove? |
|------|---------|-----------------|
| **bridge/execute_tool.py** | Routes tools to Python when Go delegates (only **mlx** generate uses it now) | No – required for mlx generate fallback |
| **bridge/execute_tool_daemon.py** | Daemon mode for tool execution | Optional – only if daemon path is retired |
| **bridge/context/** | Context tool (summarizer, tool) | No – still called when context tool uses bridge |
| **bridge/recommend/** | Recommend tool (model, workflow) | No – still called when recommend uses bridge |
| **bridge/agent_evaluation.py** | AgentScope evaluation (CI) | No – used by agentic CI |
| **bridge/statistics_bridge.py** | Statistics helpers via Go binary | Optional – if no Python callers |
| **bridge/proto/** | Protobuf for bridge protocol | No – used by execute_tool |

### 2. project_management_automation/ (library)

- **~100+ files** – tools, resources, scripts, utils.
- **Role:** Python implementations that `execute_tool.py` and automation scripts call.
- **resources/** – `context_primer.py`, `memories.py`, `session.py` (prompt_discovery and prompts stub removed).
- **tools/** – consolidated*.py, reporting, scorecard, overview, git, memory, session, etc.

**Safe to remove:** Nothing else at module level without auditing callers. Further removal = identify dead scripts/tools or migrate callers to Go.

### 3. Tests

- **tests/unit/python/test_execute_tool.py** – bridge tool routing (import path may need fixing).
- **tests/integration/bridge/** – bridge integration.
- **tests/fixtures/** – mocks and helpers.

---

## What Can Be Safely Removed (Next Steps)

1. **Dead scripts** – Audit `project_management_automation/scripts/` for scripts never run; remove or document as optional.
2. **statistics_bridge.py** – Remove only if no Python code calls it (grep for imports).
3. **execute_tool_daemon.py** – Remove only if nothing uses daemon mode (check Go/config).
4. **Ruff cleanup** – Apply `ruff check --fix` (or manual) for F401/F841 in `project_management_automation/` (see PYTHON_DEAD_CODE_REPORT.md).

---

## Plan for the Rest

| Area | Action |
|------|--------|
| **Bridge** | Keep for **mlx** (generate) and any context/recommend bridge path until those are native-only or deprecated. |
| **project_management_automation** | Treat as **library**; don’t delete en masse. Prefer: (1) stop calling from Go where possible, (2) migrate high-value tools to Go, (3) remove only proven-dead modules. |
| **Resources** | All 22 resources are **native Go**; no Python resource code needed. |
| **Prompts** | All prompts in **Go** (`internal/prompts`); Python manifest only in `context_primer.py` for discovery. |
| **Docs** | Keep PYTHON_REMAINING_ANALYSIS.md and NEXT_MIGRATION_STEPS.md updated as tools/resources are migrated. |

---

## Todo2 Tasks That Cover This

| Topic | Task / area |
|-------|----------------|
| **Python cleanup / removal** | "Remove unused code" (T-*), "Python Code Cleanup Epic", "Analyze Python code for safe removal", "Identify and document any leftover Python code...", "Remove leftover Python code files identified in audit" |
| **LEGACY_PROMPTS / scorecard** | "Replace or document LEGACY_PROMPTS_COUNT in Python scorecard/overview" – **done** (stub removed; scorecard/overview use 0). |
| **Prompt resources** | "Migrate prompt resources to native Go" – **done** (resources are native; prompt_discovery inlined in context_primer). |
| **Migration docs** | "Update migration docs for 4 removed Python fallbacks" |
| **Bridge usage** | "Document bridge usage", "Create Python code removal plan document" (PYTHON_DEAD_CODE_REPORT.md exists) |
| **Tool migration** | Many "Migrate X" tasks (ollama, lint, memory, session, etc.) – see NATIVE_GO_HANDLER_STATUS.md for current status. |
| **Resources migrated** | "All resources migrated to native Go (Phase 4)" – **done** (22 resources native). |

---

## Summary

- **Committed this session:** Removed prompt_discovery.py, prompts.py stub; inlined prompt manifest into context_primer; removed broken templates usage in auto_primer; dropped unused imports in bridge (agent_evaluation, execute_tool_daemon, statistics_bridge); updated scorecard/overview to use `prompts_count = 0`.
- **Left in Python:** Bridge (execute_tool, context, recommend, agent_evaluation, statistics_bridge, proto) and full `project_management_automation/` library.
- **Safe next removals:** Only after audit (dead scripts, unused statistics_bridge/daemon, or ruff F401/F841 in pma).
- **Plan:** Keep bridge for mlx/context/recommend until migrated or deprecated; treat pma as library; keep docs and Todo2 tasks above in sync.
