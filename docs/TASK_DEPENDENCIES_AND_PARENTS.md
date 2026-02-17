# Task dependencies and parent-child (Feb 2026)

Summary of a pass over task titles to add missing **dependencies** (blocking order) and **parent_id** (hierarchy) so waves and execution order reflect reality.

## Changes applied

### Parent: Local AI task assignment (T-1771164552852)

Set **parent_id** for:

- **T-1771172293634** – Verify task_workflow local_ai_backend for all relevant actions  
- **T-1771171137798** – Proto: add local_ai_backend to EstimationRequest  

(Other Local AI children already had this parent: T-1771171136408, T-1771171134703, T-1771171361363, T-1771171132414.)

### Parent: Unit tests for all components – Phase 6 (T-1771171348644)

Set **parent_id** for:

- **T-1771171357749** – Unit tests for task_execute (RunTaskExecutionFlow with mocked model)  
- **T-1771171360013** – Unit tests for execution_apply (ParseExecutionResponse, ApplyChanges)  

(T-1771171359038 was already a child.)

### Parent: Performance benchmarks – Phase 6 (T-1771171349874)

Set **parent_id** for:

- **T-1771171362298** – Benchmark RunTaskExecutionFlow  
- **T-1771171363644** – Benchmark RefinePromptLoop (single iteration)  

(T-1771171361363 was already a child.)

### Parent: Enhancements Epic (T-1768318471624)

Set **parent_id** for:

- **T-1771172723091** – Task tool enrichment: include recommended_tools in task show  
- **T-1771172721983** – Task tool enrichment: include recommended_tools in session prime  
- **T-1771172717418** – Task tool enrichment: support recommended_tools…  
- **T-1771172725704** – Task tool enrichment: CLI --recommended-tools  
- **T-1771172724534** – Task tool enrichment: tag-based enrich_tool_hints  
- **T-1771172735887** – Task tool enrichment: optional config for tag-to-tools  
- **T-1771252276374** – Document AI/LLM stack in main docs  
- **T-1771252272139** – Optional LocalAI backend (OpenAI-compatible)  
- **T-1771252280227** – Optional stdio://llm/status resource  
- **T-1771164549623** – Session prime/handoff: Cursor CLI suggestion  

### Parent: Testing & Validation Epic (T-1768318471621)

Set **parent_id** for the Fix* test tasks (same epic as Stream 5):

- **T-1771278405185** – Fix database package tests  
- **T-1771278411878** – Fix tools tests  
- **T-1771280369024605000** – Fix internal/tools tests  
- **T-1771278414077** – Fix CLI integration  

### Dependencies added (Phase / Wave / Plan–Test–Verify)

**Phase 1.x and T1.5.x** – Already had correct chains; no change.

**Wave 2:** T-1771246229289 → T-1771245936020; T-1771246233626 → T-1771246229289.

**mcp-go-core:** T-1771245936020 → T-1771245926712; T-1771245933129 → T-1771245936020.

**Fix test chain:** T-1771278411878, T-1771280369024605000, T-1771278414077 → T-1771278405185.

**Logical connections:** Plan-related tasks mostly Done. Testing & Validation Epic groups Stream 5 and Fix* tasks. Verify tasks (Wave 2, Verify local_ai_backend) chained or under correct epics.

---

## Optional dependency suggestions (not applied)

- **Database test fixes** – T-1771245897653 (comment tests) and T-1771245901197 (lock tests) could depend on T-1771278405185 (Fix database package tests) if you want “fix package tests first, then comment/lock”.
- **Jenkins** – T-1771253068151, T-1771253063737, T-1771253072479, T-1771253067039, T-1771253070385 could be given a parent “Jenkins/CI epic” if you add one.

## Task analysis: suggest_dependencies

The **task_analysis** tool has an action **suggest_dependencies** that infers possible dependencies from task content and (optionally) from planning docs. It does not apply changes.

**From content:** Phase N.M → Phase N.(M-1); T1.5.M → T1.5.(M-1); "Verify Wave N" / "Wave N verification" → Fix gosdk / Verify gosdk; "Fix tools/CLI/internal tests" → "Fix database package tests".

**From planning docs:** With `include_planning_docs: true`, scans `.cursor/plans` and `docs` for `*plan*.md`, and extracts:
- Explicit lines: "Depends on: T-XXX", "dependencies: [T-XXX]", "After: T-XXX"
- Milestone order: earlier checkbox `(T-ID)` in the doc suggested as dependency for the next.

**Usage:**
```bash
exarp-go -tool task_analysis -args '{"action":"suggest_dependencies"}'
exarp-go -tool task_analysis -args '{"action":"suggest_dependencies","include_planning_docs":true}'
```
Then apply any suggestion with `task_workflow` update (task_ids + dependencies).

---

## How to update in future

- **Parent (hierarchy):**  
  `exarp-go -tool task_workflow -args '{"action":"update","task_ids":"T-XXX,T-YYY","parent_id":"T-EPIC"}'`  
- **Dependencies (blocking):**  
  `exarp-go -tool task_workflow -args '{"action":"update","task_ids":"T-YYY","dependencies":["T-XXX"]}'`  
  Note: `dependencies` **replaces** the task’s current dependency list; include existing deps if you only want to add one.

## References

- TUI “Waves” view uses dependency order and parent_id (see `internal/cli/tui.go`, `BacklogExecutionOrder`).  
- `internal/tools/graph_helpers.go`: `BuildTaskGraph` uses both `Dependencies` and `parent_id` for ordering.  
- Task workflow skill: `.cursor/skills/task-workflow/SKILL.md`.
