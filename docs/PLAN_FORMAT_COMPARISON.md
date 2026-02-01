# Plan Format Comparison: exarp-go Generated vs configuration_tasks_plan

This compares the **exarp-go generated plan** (report action=plan) to the format used in **configuration_tasks_plan_983edd1f.plan.md** (Cursor/plans style with frontmatter and tables).

**Reference:** `~/.cursor/plans/configuration_tasks_plan_983edd1f.plan.md`  
**Generated:** Run `report(action="plan")` after building exarp-go to get the new format.

---

## Alignment: Generated plans now match reference format (Cursor-buildable)

As of the latest generator, exarp-go plans include:

| Aspect | exarp-go generated (new) | configuration_tasks_plan (reference) |
|--------|--------------------------|--------------------------------------|
| **Frontmatter** | YAML: `name`, `overview`, `todos` (task IDs), `isProject: true` | YAML: `name`, `overview`, `todos`, `isProject` |
| **Title** | `# <display name> Plan` + **Generated:** date | `# <Name> Plan` |
| **Scope** | **## Scope** paragraph (from purpose/overview) | **## Scope** paragraph |
| **Section dividers** | `---` between every major section | `---` between every major section |
| **Sections** | 1. Technical Foundation (+ critical path)<br>2. Backlog Tasks (table)<br>3. Iterative Milestones (checkboxes)<br>4. Recommended Execution Order (mermaid + list + Parallel)<br>5. Open Questions<br>6. Out-of-Scope / Deferred | 1. Epic phases (tables)<br>2. MCP/CLI tasks (table)<br>3. Verification (table)<br>4. Recommended Execution Order (mermaid + list)<br>5. Out-of-Scope / Deferred |
| **Tasks** | Table: \| Task \| Priority \| Description \| + checklist `- [ ] **Name** (T-xxx)` | Tables: \| Task \| Priority \| Description \| |
| **Execution order** | Mermaid flowchart (critical path chain) + numbered list + **Parallel:** items | Mermaid flowchart + numbered list + **Parallel:** items |
| **Out-of-scope** | **## 6. Out-of-Scope / Deferred** (low-priority backlog not in next actions, or placeholder) | **## 5. Out-of-Scope / Deferred** |
| **Content source** | Project info, Todo2 next actions, critical path, risks | Hand-written; task IDs and descriptions curated |

---

## Remaining differences

- **Epic/phase tables:** Reference plans can have multiple themed tables (Phase | Goal | Key Files; Task | Priority | Description by theme). Generated plan has one **Backlog Tasks** table (Task | Priority | Description) from Todo2 next actions.
- **Content:** Reference plans are hand-curated (epic IDs, references, phase goals). Generated plan is auto-filled from Todo2, critical path, and project metrics — same structure, different data source.
- **Out-of-scope:** Generated plan fills Out-of-Scope from low-priority backlog tasks not in next actions, or a placeholder; reference plans list hand-picked deferred items.

---

## Cursor-buildable behavior

- **Location:** Default output is **`.cursor/plans/<project-slug>.plan.md`** so Cursor discovers the plan and shows the Build option (plans in project root may not get Build).
- **Frontmatter** `name`, `overview`, `todos` (task IDs), `isProject: true`, **`status: draft`** — Cursor plan UI and agents can read these; `status: draft` for new plans; Cursor may set `status: built` when you build from the plan.
- **Checkboxes** in **Iterative Milestones** — Cursor can track completion; `todos` in frontmatter lists task IDs for tooling.
- **Plan ↔ Todo2 sync:** Use `task_workflow` with `action=sync_from_plan` (or `sync_plan_status`) and `planning_doc=<path to .plan.md>`. Parses frontmatter and milestone checkboxes, creates/updates Todo2 tasks, and optionally writes the plan file back so checkboxes and frontmatter `todos[].status` match Todo2 (bidirectional). Params: `planning_doc` (required), `dry_run` (optional), `write_plan` (optional, default true).
- **Tables and mermaid** — Same structure as reference plans so Cursor can parse sections and execution order.

Rebuild exarp-go and run `report(action="plan")` to generate a plan in `.cursor/plans/` with this format. Reference: **~/.cursor/plans/configuration_tasks_plan_983edd1f.plan.md**.
