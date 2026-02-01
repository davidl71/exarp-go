---
name: Cursor Plan Mode Improvements
overview: "Improve exarp-go to better integrate with Cursor Plan Mode: file references, bidirectional sync, tag hints, and plan discoverability."
todos:
  - id: T-1769980664971
    content: Add file/code references to plan milestones
    status: pending
  - id: T-1769980677401
    content: Align Scope section with Purpose & Success Criteria
    status: pending
  - id: T-1769980679071
    content: Add tag_hints to generated plan frontmatter
    status: pending
  - id: T-1769980680612
    content: Add plan sync from file (Cursor to Todo2)
    status: pending
  - id: T-1769980682108
    content: Update existing plans without overwriting hand-edits
    status: pending
  - id: T-1769980690895
    content: Update report tool MCP descriptor for plan actions
    status: pending
  - id: T-1769980692361
    content: Session prime - suggest Plan Mode for complex tasks
    status: pending
  - id: T-1769980693841
    content: Sync plan status back to Todo2 when built
    status: pending
isProject: false
status: draft
tag_hints: ["planning", "feature", "report"]
---

# Cursor Plan Mode Improvements Plan

**Created:** 2026-02-01

## Scope

Improve exarp-go's integration with Cursor Plan Mode so generated plans align with Cursor's expectations (Purpose & Success Criteria, file references, tag hints) and support bidirectional sync between Cursor plans and Todo2.

---

## 1. Technical Foundation

- **Source:** Analysis of Cursor docs (cursor.com/learn/creating-plans, cursor.com/blog/plan-mode)
- **Current:** exarp-go `report(action=plan)` generates `.cursor/plans/<slug>.plan.md` with Todo2 data
- **Gap:** Missing file references, tag hints; no sync from Cursor edits back to Todo2

---

## 2. Backlog Tasks

| Task | Priority | Description |
|------|----------|-------------|
| **T-1769980664971** | medium | Add file/code references to plan milestones (use task_analysis or task_discovery) |
| **T-1769980677401** | low | Align Scope section with "Purpose & Success Criteria" (Cursor Creating Plans) |
| **T-1769980679071** | low | Add tag_hints to generated plan frontmatter from Todo2 tags |
| **T-1769980680612** | high | Add plan sync from file (Cursor → Todo2) - bidirectional |
| **T-1769980682108** | medium | Update existing plans without overwriting hand-edits (merge strategy) |
| **T-1769980690895** | low | Update report tool MCP descriptor for plan, scorecard_plans |
| **T-1769980692361** | low | Session prime: suggest Plan Mode for complex tasks |
| **T-1769980693841** | medium | Sync plan status back to Todo2 when built |

---

## 3. Iterative Milestones

Each milestone is independently valuable. Check off as done.

- [ ] **Add file/code references to plan milestones** (T-1769980664971)
- [ ] **Align Scope section with Purpose & Success Criteria** (T-1769980677401)
- [ ] **Add tag_hints to generated plan frontmatter** (T-1769980679071)
- [ ] **Add plan sync from file (Cursor to Todo2)** (T-1769980680612)
- [ ] **Update existing plans without overwriting hand-edits** (T-1769980682108)
- [ ] **Update report tool MCP descriptor for plan actions** (T-1769980690895)
- [ ] **Session prime: suggest Plan Mode for complex tasks** (T-1769980692361)
- [ ] **Sync plan status back to Todo2 when built** (T-1769980693841)

---

## 4. Recommended Execution Order

**Quick wins (low effort):**
1. T-1769980677401 — Align Scope with Purpose & Success Criteria
2. T-1769980679071 — Add tag_hints to frontmatter
3. T-1769980690895 — Update report MCP descriptor
4. T-1769980692361 — Session prime Plan Mode hint

**Medium effort:**
5. T-1769980664971 — Add file references to milestones
6. T-1769980682108 — Update existing plans (merge)
7. T-1769980693841 — Sync plan status to Todo2

**Higher impact:**
8. T-1769980680612 — Bidirectional plan sync (Cursor ↔ Todo2)

---

## 5. Open Questions

- Should plan sync support `.agentplan` JSON format (Cursor 2.2+) in addition to `.plan.md`?

---

## 6. Out-of-Scope / Deferred

- None

---

*Generated for Cursor Plan Mode integration. Regenerate with `report(action="plan")` when Todo2 changes.*
