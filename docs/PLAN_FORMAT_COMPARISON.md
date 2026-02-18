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

## Frontmatter `todos` section (to-dos)

YAML frontmatter in `.plan.md` files may include a **`todos`** array so Cursor and exarp-go can show progress and sync with Todo2.

### Schema

```yaml
todos:
  - id: <string>       # Task or step identifier (see conventions below)
    content: <string>  # Short label; Cursor displays this
    status: <string>   # pending | in_progress | completed (or review, done — normalized by plan_sync)
```

- **`id`** — Unique key for the item. Two conventions:
  - **Todo2 task ID** (`T-<digits>`): Use when the plan item maps to a Todo2 task. Required for `task_workflow` `sync_from_plan` / `sync_plan_status` to create/update Todo2 and to match milestone checkboxes `(T-ID)`.
  - **Wave or slug** (e.g. `wave1-cursor-run`): Use in execution/wave plans where each item is a wave or group; sync will create Todo2 tasks with that id if you run sync_from_plan (optional).
- **`content`** — Human-readable label; often truncated to ~120 chars in generated plans.
- **`status`** — Allowed values (exarp-go normalizes when writing back):
  - `pending` — not started (Todo2: Todo)
  - `in_progress` — in progress (Todo2: In Progress, Review)
  - `completed` or `done` — done (Todo2: Done)

Parsing: `internal/tools/plan_sync.go` (`PlanTodo`, `parsePlanFile`). Status mapping: `todo2StatusToPlanStatus` / `cursorStatusToTodo2`.

### Which exarp-go generators emit `todos`

| Generator | Emits `todos`? | Notes |
|-----------|----------------|--------|
| **report** `action=plan` | Yes | Main project plan; id = Todo2 task ID, content from task name, status from Todo2. Also emits `waves` in frontmatter. |
| **report** `action=parallel_execution_plan` / **task_analysis** `action=execution_plan`, `output_format=subagents_plan` | No | Writes `parallel-execution-subagents.plan.md` with only `name`, `overview`, `isProject`, `status`. No `todos` array. |
| **task_analysis** `action=execution_plan` (save to `.plan.md`) | No | `formatExecutionPlanMarkdown` writes body-only markdown (waves table, full order); no YAML frontmatter. |

So only the main **report(action=plan)** generator produces Cursor-buildable frontmatter including `todos`. Hand-written plans (e.g. docs execution plans with wave-based ids) should add `todos` manually if you want Cursor Build and plan sync to use them. Plans like `docs_tasks_execution_plan_*.plan.md` use wave-style ids (`wave1-cursor-run`, etc.) and `status: completed`/`pending`; that format is valid and is not generated by exarp-go (add/update by hand or script).

---

## "Referenced by" and Agents section (body)

Generated plans include a **body** block (not frontmatter) that lists who uses the plan and an **Agents** table.

### Schema

- **Referenced by:** One line in the plan body:
  ```markdown
  **Referenced by:** [label1](path1), [label2](path2), ...
  ```
  - Comma-separated markdown links to agents (e.g. `.cursor/agents/<name>.md`) and/or rules (e.g. `.cursor/rules/<name>.mdc`) that reference this plan.
  - Main plan generator emits **3 references**: [wave-task-runner](.cursor/agents/wave-task-runner.md), [wave-verifier](.cursor/agents/wave-verifier.md), [plan-execution](.cursor/rules/plan-execution.mdc).
  - Scorecard improvement plans emit a placeholder: *(none by default; add agents in `.cursor/agents/` or rules in `.cursor/rules/` to reference this plan)*.

- **Agents:** A `## Agents` section with a table:
  ```markdown
  ## Agents

  | Agent | Role |
  |-------|------|
  | [name](path) | Short role description. |
  ```
  - Main plan: two rows (wave-task-runner, wave-verifier). Scorecard plans: one placeholder row suggesting adding an agent.

There is no YAML frontmatter field for "Referenced by" or agents; both are markdown body only. Source: `internal/tools/report.go` (`generatePlanMarkdown`, `generateScorecardDimensionPlan`).

---

## Cursor-buildable behavior

- **Location:** Default output is **`.cursor/plans/<project-slug>.plan.md`** so Cursor discovers the plan and shows the Build option (plans in project root may not get Build).
- **Frontmatter** `name`, `overview`, `todos` (task IDs or wave slugs), `isProject: true`, **`status: draft`** — Cursor plan UI and agents can read these; `status: draft` for new plans; Cursor may set `status: built` when you build from the plan.
- **Checkboxes** in **Iterative Milestones** — Cursor can track completion; `todos` in frontmatter lists task IDs for tooling.
- **Plan ↔ Todo2 sync:** Use `task_workflow` with `action=sync_from_plan` (or `sync_plan_status`) and `planning_doc=<path to .plan.md>`. Parses frontmatter and milestone checkboxes, creates/updates Todo2 tasks, and optionally writes the plan file back so checkboxes and frontmatter `todos[].status` match Todo2 (bidirectional). Params: `planning_doc` (required), `dry_run` (optional), `write_plan` (optional, default true).
- **Tables and mermaid** — Same structure as reference plans so Cursor can parse sections and execution order.

Rebuild exarp-go and run `report(action="plan")` to generate a plan in `.cursor/plans/` with this format. Reference: **~/.cursor/plans/configuration_tasks_plan_983edd1f.plan.md**.

### Build button does nothing (Cursor IDE)

Cursor has known issues where the **Build** button on a plan does nothing or stops working:

- **Frontmatter stripped on save:** Saving the plan from Cursor (or certain editors) can remove YAML fields Cursor doesn’t recognize (e.g. `status: draft`, `last_updated`, `tag_hints`, `waves`). Build may only work when `status: draft` (and optionally other expected fields) are present in the frontmatter.
- **Plan corruption:** Saving plans to the repo or locally has been reported to corrupt the file and break Build ([forum](https://forum.cursor.com/t/plans-corrupt-when-save-it-to-your-repo-breaking-build-button-from-working/149598)).

**Workarounds:**

1. **Restore frontmatter:** Ensure the plan has at least `name`, `overview`, `todos`, `isProject: true`, and **`status: draft`** in the YAML block. Regenerating with exarp-go restores full frontmatter: `./bin/exarp-go -tool report -args '{"action":"plan"}'`.
2. **Repair without full regenerate:** To restore frontmatter and the “## 3. Iterative Milestones” checkboxes without overwriting the rest of the body (e.g. after Cursor stripped fields), use: `report(action="plan", repair=true, plan_path=".cursor/plans/exarp-go.plan.md")` or `exarp-go -tool report -args '{"action":"plan","repair":true,"plan_path":".cursor/plans/exarp-go.plan.md"}'`.
3. **Avoid saving from Cursor:** Edit the plan in a plain editor or only regenerate via exarp-go so Cursor doesn’t rewrite and strip fields.
4. **Build via chat:** Ask in Cursor chat to “build this plan” or “execute the plan” instead of relying on the Build button.
5. **Restart Cursor** after fixing the plan file so it re-parses frontmatter.
