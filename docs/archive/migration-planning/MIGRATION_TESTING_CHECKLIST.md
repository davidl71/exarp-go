# Native Go Migration Testing Checklist

**Date:** 2026-01-28  
**Purpose:** When to add/update tests and `toolsWithNoBridge` during native Go migration.

---

## When adding a new native implementation

1. **Unit test** the new handler (e.g. `TestHandleAlignmentPRD`, `TestHandleXNative`).
2. **Response shape:** Assert JSON keys the client expects (e.g. `success`, `data`, `status`).
3. **Edge cases:** Empty inputs, missing PROJECT_ROOT, invalid action.

## When removing a Python fallback (tool becomes native-only)

1. **Handlers:** Remove `bridge.ExecutePythonTool(ctx, "<tool>", params)`; return native result or error.
2. **Bridge:** Remove the `if tool_name == "<tool>":` branch and its import in `bridge/execute_tool.py`.
3. **Regression:** Add the tool to `toolsWithNoBridge` in `internal/tools/regression_test.go`.
4. **TestRegressionNativeOnlyTools:** Add the tool to the `want` list.
5. **TestRegressionFallbackBehavior:** For tools in `toolsWithNoBridge`, the test skips bridge comparison (no change needed if already skipping by map).
6. **Docs:** Update `docs/PYTHON_FALLBACKS_SAFE_TO_REMOVE.md` (safe-to-remove table and “not safe” list) and `docs/NEXT_MIGRATION_STEPS.md`.

## When keeping a hybrid (native first, bridge fallback)

1. **Regression:** Do *not* add to `toolsWithNoBridge`.
2. **Feature parity:** Add to `knownDifferences` in `TestRegressionFeatureParity` with a short reason (e.g. “Hybrid: native X; Python fallback when native fails”).
3. **Tests:** Prefer testing the native path; optionally compare to bridge when bridge is available (and skip when unavailable).

## Bridge call map (reference)

- **Handlers:** `internal/tools/handlers.go` — each handler that still calls `bridge.ExecutePythonTool`.
- **Automation:** `internal/tools/automation_native.go` — `runDailyTask` (native) vs `runDailyTaskPython` (bridge).
- **Task workflow:** `internal/tools/task_workflow_common.go` — bridge required when `external=true` (agentic-tools sync).

---

**See also:** `docs/NEXT_MIGRATION_STEPS.md`, `docs/PYTHON_FALLBACKS_SAFE_TO_REMOVE.md`
