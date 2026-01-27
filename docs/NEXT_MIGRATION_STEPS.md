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

| Where | Change |
|-------|--------|
| `automation_native.go` ~328 (sprint) | `runDailyTaskPython(ctx, "report", overview)` → native `report` overview when implemented; then `runDailyTask(ctx, "report", ...)` |

## 3. Medium: shrink Python fallback surface

| Tool | Next step |
|------|-----------|
| **report** | Native overview for sprint automation |
| **analyze_alignment** | Native `action=prd` or document scope |
| **estimation** | Harden native, reduce fallback |
| **task_analysis** / **task_discovery** | Improve native or document bridge cases |

## 4. Deferred / intentional

- **task_workflow** `external=true` → bridge required (agentic-tools)
- **mlx** → bridge-only by design
- **lint** → hybrid by design (Go native, others bridge)
- memory semantic search, testing suggest/generate, security non-Go, recommend advisor

## 5. Docs

- Migration checklist (when to add tests, when to update toolsWithNoBridge)
- Hybrid pattern guide (when native-only vs native+fallback)
- Bridge call map (handlers, automation, task_workflow_common)

---

**Order:** 1) Test fixing/expansion → 2) report overview in automation → 3) analyze_alignment prd → 4) estimation/task_* hardening.
