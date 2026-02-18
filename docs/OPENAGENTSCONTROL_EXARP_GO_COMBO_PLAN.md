# Plan: OpenAgentsControl + exarp-go Combo

**Tag hints:** `#docs` `#integration` `#opencode` `#workflow`

**Status:** Draft  
**Created:** 2026-02-17

---

## 1. Purpose & Success Criteria

**Purpose:** Define how [OpenAgentsControl (OAC)](https://github.com/darrenhinde/OpenAgentsControl) and exarp-go work together so users get **plan-first, approval-gated agent workflows** (OAC) plus **task lifecycle, reporting, and session context** (exarp-go) in one setup.

**Success criteria:**

- Clear split of responsibilities: OAC = agent workflow (plan → approve → execute); exarp-go = task store, reports, session, Cursor/MCP integration.
- One doc and optional Todo2 tasks track the combo; no duplicate task stores or conflicting approval models.
- Users can run OAC (OpenCode CLI or Claude Code plugin) with exarp-go as the task/report/session backend via MCP or CLI.

---

## 2. Role Split

| Concern | OpenAgentsControl (OAC) | exarp-go |
|--------|--------------------------|----------|
| **Agent workflow** | Plan → propose → human approve → execute. ContextScout, MVI context, approval gates. | Not in scope. |
| **Task store** | Can delegate “what to do” to external systems. | **Primary:** Todo2 (SQLite/JSON), task_workflow, task_analysis, dependencies. |
| **Session / handoff** | Not in scope. | **Primary:** session (prime, handoff, assignee), Cursor hints. |
| **Reporting / scorecard** | Not in scope. | **Primary:** report, scorecard, health, security. |
| **Where it runs** | OpenCode CLI, Claude Code plugin (beta). | MCP server (Cursor, OpenCode), CLI, TUI, HTTP API. |
| **Context** | `.opencode/context/` (patterns, standards). | Cursor rules/skills, `.todo2/`, plans, vizvibe.mmd. |

**Combo idea:** OAC drives *how* the AI plans and executes (with your patterns and approval). exarp-go holds *what* is to be done, *who* is doing it, and *how the project looks* (tasks, reports, session).

---

## 3. Technical Foundation

- **OAC:** Built on [OpenCode](https://opencode.ai/). Supports MCP servers in config. Agents are markdown; context in `.opencode/context/`. Multi-language (TypeScript, Python, Go, Rust).
- **exarp-go:** Go MCP server; tools for task_workflow, task_analysis, report, session, health, security, automation. Integrates with OpenCode via MCP (see [OPENCODE_INTEGRATION.md](OPENCODE_INTEGRATION.md)).
- **Integration points:**
  1. **MCP:** Add exarp-go as an OpenCode/OAC MCP server so agents can call `task_workflow`, `report`, `session`, etc.
  2. **CLI:** From OpenCode terminal or scripts, run `exarp-go task list`, `exarp-go task sync`, `exarp-go task update`, `exarp-go report`, etc.
  3. **Context:** OAC context (patterns) stays in `.opencode/`. exarp-go context (tasks, plans) stays in `.todo2/`, `.cursor/`, `vizvibe.mmd`. No overlap.

---

## 4. Implementation Phases

**Tag hints:** `#docs` `#integration`

### Phase 1: Document the combo (this doc)

- [x] Create combo plan (this file).
- [x] Add a short “OAC + exarp-go” section to README or docs index that links here and to OPENCODE_INTEGRATION.md.
- [ ] Optional: add `docs/OPENAGENTSCONTROL_EXARP_GO_COMBO_PLAN.md` to vizvibe.mmd as a future or current node.

### Phase 2: OpenCode/OAC config and verification

- [x] Document in this doc or OPENCODE_INTEGRATION.md: add exarp-go as MCP server when using OAC (same config as “OpenCode + exarp-go”).
- [x] Optional: add an example `opencode.json` snippet under `docs/` or in this plan that enables exarp-go for an OAC project.
- [ ] Verify: run OAC (e.g. `opencode --agent OpenAgent`) in a repo that has exarp-go MCP configured; confirm tools (task list, report) are callable by the agent.

### Phase 3: Workflow narrative (how a user runs both)

- [x] Document a single “day in the life” flow: e.g. “Start with exarp-go session prime / task list → run OAC for a feature (plan → approve → execute) → update Todo2 via exarp-go task update / sync → report or handoff via exarp-go.”
- [ ] Optional: add OAC to docs that compare tools (e.g. when to use OAC vs exarp-go vs Cursor-only).

**Day in the life: OAC + exarp-go**

1. **Start:** Run exarp-go session prime (or `exarp-go task list --status Todo`) to see backlog and context. Optionally run `exarp-go report overview` or `exarp-go task list --order execution` for dependency order.
2. **Pick work:** Choose a Todo2 task (e.g. from list or suggested-tasks). Optionally move it to In Progress via `exarp-go task update T-xxx --new-status "In Progress"` or let OAC/agent do it via MCP.
3. **Run OAC:** Start OpenCode/OAC (e.g. `opencode --agent OpenAgent`). The agent uses exarp-go MCP to read tasks/reports. Run OAC’s plan → approve → execute flow for the feature or ticket.
4. **Update Todo2:** When the feature or task is done, update status via exarp-go: `exarp-go task update T-xxx --new-status Done` or via MCP `task_workflow` approve/update. Run `exarp-go task sync` if you use both DB and JSON.
5. **Report or handoff:** Run `exarp-go report scorecard` or `exarp-go report briefing` for a snapshot; use `exarp-go -tool session -args '{"action":"handoff",...}'` to leave notes for the next session or machine.

Same `.todo2` and `PROJECT_ROOT` = shared backlog; exarp-go holds tasks and reports, OAC/OpenCode drives the agent workflow.

---

## 5. Open Questions

- Whether OAC’s TaskManager or external-scout flows should **create or update** Todo2 tasks via exarp-go (e.g. `task_workflow` create/update) so one backlog lives in Todo2.
- Whether to add a small “OAC + exarp-go” quickstart (install OAC, add exarp-go MCP, run one task list + one report from the agent).
- How Claude Code plugin (OAC beta) discovers or configures exarp-go (if at all).

---

## 6. Out of Scope / Deferred

- exarp-go does **not** implement OAC-style approval gates or ContextScout; that remains in OAC/OpenCode.
- exarp-go does **not** ship or install OAC; users install OAC separately and optionally attach exarp-go as MCP/CLI.
- Deferred: automated tests that run “OAC agent calls exarp-go tool” (would require OpenCode/OAC in CI).

---

## 7. Oh My OpenCode + exarp-go

[Oh My OpenCode](https://github.com/code-yeongyu/oh-my-opencode) is an **OpenCode plugin** (“the best agent harness”) that provides multi-agent orchestration, LSP/AST tools, background agents, and strong defaults (Sisyphus, Hephaestus, Oracle, Librarian, Explore, etc.). It is a **different style** from OAC: **autonomy and speed** rather than approval gates.

| Aspect | Oh My OpenCode | OpenAgentsControl (OAC) |
|--------|----------------|--------------------------|
| **Execution** | Autonomous; “ultrawork” / `ulw`; agents run until done; Todo enforcer | Propose → human approve → execute |
| **Best for** | Speed, parallel agents, “boulder until done,” multi-model orchestration | Control, repeatability, approval gates, team patterns |
| **exarp-go** | Same MCP/CLI integration as OAC: add exarp-go to OpenCode MCP config; agents can call task_workflow, report, session, etc. | Same |

**Integration:** Add exarp-go as an MCP server in the same OpenCode config used by Oh My OpenCode (see [OPENCODE_INTEGRATION.md](OPENCODE_INTEGRATION.md)). Sisyphus or other agents can then list tasks, update status, run reports, and use session prime/handoff. No change to exarp-go; only the harness (Oh My OpenCode vs OAC vs vanilla OpenCode) differs.

**When to use which:**

- **OAC + exarp-go:** You want plan-first, approval before execution, and pattern control; exarp-go holds the backlog and reports.
- **Oh My OpenCode + exarp-go:** You want autonomous execution, parallel agents, and “don’t stop until done”; exarp-go still holds the backlog and reports.

---

## 8. Agentic + exarp-go

[Agentic](https://github.com/Cluster444/agentic) is an **agentic workflow and context-engineering tool** for OpenCode: phased workflow (Research → Plan → Execute → Commit → Review), “thoughts” directory for context, and slash commands (`/ticket`, `/research`, `/plan`, `/execute`, `/commit`, `/review`). Install with `agentic pull` or `agentic pull -g`; commands deploy into `.opencode/` or global config.

**Combo:** Use **Agentic** for the structured workflow and context; use **exarp-go** (as MCP in the same OpenCode config) for *what* to work on (task list, task update, sync) and reporting (report, scorecard, session prime/handoff). Example: pick a Todo2 task via exarp-go, run Agentic’s `/ticket` then `/research`/`/plan`/`/execute` for that ticket; when done, update the task in exarp-go. Add exarp-go as MCP per [OPENCODE_INTEGRATION.md](OPENCODE_INTEGRATION.md).

---

## 9. Other OpenCode ecosystem

| Project | Role | exarp-go |
|--------|------|----------|
| [OpenWork](https://github.com/different-ai/openwork) | Desktop app (Claude Cowork–style); host/client mode, sessions, permissions, skills manager | Add as MCP in OpenCode config used by OpenWork |
| [opencode-obsidian](https://github.com/mtymek/opencode-obsidian) | Obsidian plugin: embed OpenCode in sidebar | Add exarp-go as MCP; embedded OpenCode can call task/report/session tools |
| [OCX](https://github.com/kdcokenny/ocx) | OpenCode extension manager; profiles, registries, `ocx add` | exarp-go is not an OCX component; add exarp-go MCP entry manually in config |
| [opencode-workspace](https://github.com/kdcokenny/opencode-workspace) | Bundle (plugins + MCPs + agents + skills); install via `ocx add kdco/workspace` | Add exarp-go as extra MCP in same config |
| [opencode-background-agents](https://github.com/kdcokenny/opencode-background-agents) | Async delegation; results persisted to disk | Same MCP config; agents can call exarp-go for tasks/reports |
| [opencode-notify](https://github.com/kdcokenny/opencode-notify) | Native OS notifications when tasks complete | UX only; use with exarp-go MCP for backlog/reports |
| [opencode-skillful](https://github.com/zenobi-us/opencode-skillful) | Lazy-loaded skills (skill_find, skill_use, skill_resource); *archived* | Complementary; exarp-go = task/report layer |

---

## 10. References

- [OpenAgentsControl (OAC)](https://github.com/darrenhinde/OpenAgentsControl) — plan-first agents, approval gates, context system, built for OpenCode.
- [Oh My OpenCode](https://github.com/code-yeongyu/oh-my-opencode) — OpenCode plugin for multi-agent orchestration, autonomy, LSP/MCP, Sisyphus/Hephaestus.
- [Agentic](https://github.com/Cluster444/agentic) — agentic workflow and context engineering for OpenCode (research → plan → execute → commit → review).
- [exarp-go OpenCode integration](OPENCODE_INTEGRATION.md) — MCP, CLI, HTTP API for exarp-go with OpenCode.
- [OpenCode](https://opencode.ai/) — open-source AI coding framework; OAC, Oh My OpenCode, Agentic, OpenWork, and others extend or host it.
