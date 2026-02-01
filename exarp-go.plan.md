# Plan: github.com/davidl71/exarp-go

**Generated:** 2026-02-01

---

## 1. Purpose & Success Criteria

**Purpose:** MCP Server

**Success criteria:**
- Foundation Tools: complete
- Medium Complexity Tools: complete
- Complex Tools: complete
- Resources: complete
- Prompts: complete
- Overall health and task completion tracked in report/overview

## 2. Technical Foundation

- **Project type:** MCP Server
- **Tools:** 28 | **Prompts:** 34 | **Resources:** 21
- **Codebase:** 670 files (Go: 216)
- **Storage:** Todo2 (SQLite primary, JSON fallback)
- **Invariants:** Use Makefile targets; prefer report/task_workflow over direct file edits

### Critical Path

Longest dependency chain; completing these in order unblocks the most work.

1. T-1768319001463
2. T-1768268669031
3. T-1768268671999
4. T-1768268673151
5. T-1768268674114
6. T-1768268674677
7. T-1768268675515
8. T-1768268676474

## 3. Iterative Milestones

Each milestone is independently valuable. Check off as done.

- [ ] **Identify code that can be migrated from devwisdom-go to mcp-go-core** (T-1768325426665)
- [ ] **Migrate CLI to use mcp-go-core CLI utilities** (T-1768327631413)
- [ ] **Document gotoHuman API/tools** (T-105)
- [ ] **Test basic approval request flow** (T-106)
- [ ] **Create approval request helper function** (T-107)
- [ ] **Enhance `task_workflow` tool with approval action** (T-108)
- [ ] **Add approval request when task moves to Review** (T-109)
- [ ] **Test approval workflow end-to-end** (T-110)
- [ ] **Auto-sync Review tasks with gotoHuman** (T-111)
- [ ] **Handle approval/rejection responses** (T-112)

## 4. Open Questions

- *(Add open questions or decisions needed during implementation.)*
