# Parallel Execution Plan - Research Tasks Priority

**Created:** 2026-01-10T18:24:21.963315
**Strategy:** Research-first parallel execution

**Syncing Todo2 from this plan:** Use `exarp-go -tool report -args '{"action":"update_waves_from_plan"}'` (or pass `plan_path` to a different markdown file). The action parses this document for wave sections and updates Todo2 task dependencies so wave 0 = no deps, wave N = depend on one task from wave N−1.

## Phase 1: Research Tasks (In Progress)

- **Status:** In Progress
- **Parallel Count:** 8 tasks
- **Sequential Time:** 23.2 hours
- **Parallel Time:** 2.9 hours (max duration)
- **Time Saved:** 20.3 hours (87.5%)

### Tasks Started:

1. **T-0** - Research task data structure and design task resources API
   - Priority: high | Est: 2.9h

2. **T-5** - Research common Go code patterns between exarp-go and devwisdom-go
   - Priority: high | Est: 2.9h

3. **T-6** - Analyze pros and cons of extracting common code to shared library
   - Priority: high | Est: 2.9h

4. **T-7** - Recommend shared library strategy with implementation approach
   - Priority: high | Est: 2.9h

5. **T-8** - Analyze MCP protocol implementations for unified approach
   - Priority: high | Est: 2.9h

6. **T-16** - T-16
   - Priority: high | Est: 2.9h

7. **T-17** - T-17
   - Priority: high | Est: 2.9h

8. **T-18** - T-18
   - Priority: high | Est: 2.9h

## Review of local commits (Wave 5 relevance)

**Reviewed:** 2026-02-18. Recent commits relevant to Wave 5 / docs: Wave 2 task tool enrichment (recommended_tools), Wave 3–4 and human task breakdown docs, generated plan with todos and Agents schema, Cursor hooks for auto session prime, Redis+Asynq queue (M1/M2), CLI compact JSON and --quiet, scorecard cache and CI Go module cache. No blocking changes; Wave 5 test fixes and doc tasks remain in Todo2.

---

## Wave 2 verification (Wave 5 plan)

**Verified:** 2026-02-18. Task tool enrichment (recommended_tools) is implemented and wired:

- **Metadata and task_workflow:** `task_workflow` create/update accepts `recommended_tools`; stored in task metadata (see [TASK_TOOL_ENRICHMENT_DESIGN.md](TASK_TOOL_ENRICHMENT_DESIGN.md)).
- **Task show:** Task detail includes `recommended_tools` from metadata (internal/tools/task_workflow_common.go, resources/tasks.go).
- **Session prime:** `suggested_next` items include `recommended_tools` (internal/tools/session.go).
- **CLI:** `exarp-go task update` and create support `--recommended-tools` (internal/cli/task.go).

Wave 2 plan tasks (T-1771172717418 and dependents) can be tracked in Todo2; implementation satisfies the design.

---

## Summary

- **Total Tasks Started:** 8
- **Total Sequential Time:** 23.2 hours
- **Total Parallel Time:** 2.9 hours
- **Total Time Saved:** 20.3 hours
- **Efficiency Gain:** 87.5%
