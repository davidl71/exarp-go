# Next Native Go Migration Steps

**Date:** 2026-01-27  
**Source:** NATIVE_GO_MIGRATION_PLAN.md, PYTHON_FALLBACKS_SAFE_TO_REMOVE.md

---

## 1. High priority: testing and validation

- Unit tests for all native implementations
- Integration tests for tools/resources via MCP
- Regression tests for hybrid paths (native first, then bridge)
- Migration testing checklist in docs

## 2. Reduce automation Python use

| Where | Status |
|-------|--------|
| `automation_native.go` ~327 (sprint) | Done: uses `runDailyTask(ctx, "report", {"action": "overview"})` (native). |

## 3. Medium: shrink Python fallback surface

| Tool | Next step |
|------|-----------|
| **report** | Done: overview, briefing (native only, no fallback), scorecard (Go only; non-Go returns clear error). |
| **analyze_alignment** | Done: native `action=prd` (persona alignment) |
| **estimation** | Done: native only, no Python fallback. |
| **task_analysis** | Done: fully native (FM provider abstraction; no Python fallback). |
| **task_discovery** | Improve native or document bridge cases |

## 4. Deferred / intentional

- **task_workflow** `external=true` → bridge required (agentic-tools)
- **mlx** → bridge-only by design
- **lint** → hybrid by design (Go native, others bridge)
- memory semantic search, testing suggest/generate, security non-Go, recommend advisor

## 5. Docs

- **Migration testing checklist:** `docs/MIGRATION_TESTING_CHECKLIST.md` (when to add tests, when to update toolsWithNoBridge)
- Hybrid pattern guide (when native-only vs native+fallback)
- Bridge call map (handlers, automation, task_workflow_common)

---

**Order (current):** 1) Testing/validation checklist ✅; 2) report overview in automation ✅; 3) analyze_alignment prd ✅; 4) report briefing/scorecard + estimation shrink ✅ (2026-01-28); 5) task_analysis fully native ✅ (2026-01-28). **Next:** task_discovery improve native or document bridge cases; testing/validation (unit tests for native impls).
