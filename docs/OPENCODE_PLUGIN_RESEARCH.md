# OpenCode Plugin Research: Patterns for exarp-go

Research date: 2026-03-24

## Summary

Reviewed 5 OpenCode plugins to identify patterns applicable to exarp-go:
- micode (288 stars)
- opencode-scheduler (217 stars)
- @plannotator/opencode (3.5k stars)
- opencode-skillful (237 stars)
- oh-my-openagent (43k stars)

---

## 1. Ledger/Continuity System (micode pattern)

**Source:** [vtemian/micode](https://github.com/vtemian/micode) - Brainstorm→Plan→Implement workflow

### Pattern Description
Session continuity via structured `CONTINUITY_*.md` files that persist context across sessions.

### Key Features
- `/ledger` command creates/updates continuity ledger
- Auto-compaction at 50% context usage
- Context injection into system prompt
- File ops tracking for deterministic logging

### Implementation Potential for exarp-go
```
thoughts/ledgers/CONTINUITY_{session}.md
```
Could store:
- Current task being worked
- Recent changes made
- Next steps planned
- Blockers encountered

### exarp-go Integration Points
- Session handoff enhancement (already exists)
- Auto-summary generation on context threshold
- Agent resume from last state

---

## 2. Visual Plan Review (plannotator pattern)

**Source:** [backnotprop/plannotator](https://github.com/backnotprop/plannotator) - Interactive plan annotation

### Pattern Description
Browser-based visual annotation of AI plans with inline comments, deletions, replacements.

### Key Features
- `submit_plan` tool opens browser UI
- Select text → annotate (delete/replace/comment)
- Plan diff on revision
- Obsidian integration for approved plans
- Runs locally (no network requests)

### Implementation Potential for exarp-go
- Task approval workflow with visual UI
- Inline comments on task implementation plans
- Diff view when agent revises after feedback

### exarp-go Integration Points
- Enhanced task approval (request_approval action)
- Planning document review
- handoff export with annotated plans

---

## 3. Multi-Agent Orchestration (oh-my-openagent pattern)

**Source:** [code-yeongyu/oh-my-openagent](https://github.com/code-yeongyu/oh-my-openagent) - Agent harness

### Pattern Description
Sisyphus orchestrates specialized agents: Prometheus (planner), Hephaestus (deep worker), Oracle (architecture), Librarian (search).

### Key Features
- Category-based delegation (visual-engineering, ultrabrain, deep, quick)
- Model fallback chains per agent
- Background parallel agents
- Hash-anchored edit tool (zero stale-line errors)
- Built-in MCPs (websearch, docs, GitHub)

### Agent Roles
| Agent | Role | Default Model |
|-------|------|---------------|
| Sisyphus | Main orchestrator | Claude Opus 4.6 |
| Prometheus | Strategic planner | GPT-5.4 |
| Hephaestus | Deep worker | GPT-5.3-codex |
| Oracle | Architecture/debug | GPT-5.4 |
| Librarian | Docs/code search | MiniMax M2.5 Free |

### Implementation Potential for exarp-go
- Task type → agent specialization mapping
- Multi-agent parallel task execution
- Model fallback for cost optimization
- Background task agents

### exarp-go Integration Points
- Task workflow agents (plan/execute/review)
- Concurrent multi-agent support (already in todos)
- Built-in search MCP for codebase queries

---

## 4. Scheduled Background Tasks (opencode-scheduler pattern)

**Source:** [different-ai/opencode-scheduler](https://github.com/different-ai/opencode-scheduler) - OS-native scheduling

### Pattern Description
Recurring job scheduling via launchd (macOS), systemd (Linux), or cron fallback.

### Key Features
- Natural language scheduling (`Schedule a daily job at 9am to...`)
- Scope isolation by workdir
- Job supervision (no overlap + optional timeout)
- Non-interactive scheduled runs (denies question prompts)

### Reliability Guarantees
- **No overlap**: Skip if previous run active
- **Non-interactive**: Forces `OPENCODE_PERMISSION=deny`
- **Optional timeout**: SIGTERM → SIGKILL

### Implementation Potential for exarp-go
- Recurring task checks/syncs
- Scheduled database cleanup
- Periodic health checks
- Automated reporting

### exarp-go Integration Points
- `exarp-go task sync --scheduled` for cron
- Periodic database vacuum
- Recurring report generation

---

## 5. Lazy Skill Loading (opencode-skillful pattern)

**Source:** [zenobi-us/opencode-skillful](https://github.com/zenobi-us/opencode-skillful) - On-demand skills

### Pattern Description
Skills discovered at startup but injected only when requested, reducing context bloat.

### Key Features
- `skill_find` - Discover skills by keyword
- `skill_use` - Load skill into chat context
- `skill_resource` - Read specific files from skills
- Per-model format (XML/JSON/Markdown)
- Security: Pre-indexed resources (no path traversal)

### Discovery Paths
1. `~/.config/opencode/skills/`
2. `.opencode/skills/` (project-local)

### Implementation Potential for exarp-go
- Task-specific skill libraries
- Project-scoped skill directories
- Skill-based tool categorization

### exarp-go Integration Points
- Tool skill documentation
- Project-specific command templates
- Skill discovery for task type

---

## Recommendations for exarp-go

### High Priority (Quick Wins)
1. **Ledger/Continuity** - Enhance existing session handoff with structured continuity files
2. **Visual Plan Review** - Add browser-based task approval UI

### Medium Priority (Core Enhancement)
3. **Multi-Agent Orchestration** - Implement concurrent task execution with agent specialization

### Low Priority (Future)
4. **Scheduled Tasks** - OS-native scheduling for recurring operations
5. **Lazy Skill Loading** - On-demand skill discovery system

---

## Related exarp-go Tasks
- T-1774162546797759000: Improve concurrent multi-agent support (related to #3)
- Session handoff enhancement (existing feature, could use #1)

---

## References
- https://github.com/vtemian/micode
- https://github.com/different-ai/opencode-scheduler
- https://github.com/backnotprop/plannotator
- https://github.com/zenobi-us/opencode-skillful
- https://github.com/code-yeongyu/oh-my-openagent
