# Review: #docs Tasks

**Date:** 2026-01-29  
**Scope:** All Todo2 tasks with tag `docs`.

---

## Summary

| Status      | Count |
|------------|-------|
| **Done**    | 54    |
| **In Progress** | 3  |
| **Todo**    | 124   |
| **Total**   | 181   |

---

## In Progress (3)

| ID | Content | Priority |
|----|--------|----------|
| T-1768319224557 | Review 6 remaining standalone tasks for epic assignment or completion | high |
| T-1768319355360 | Review dependencies between subtasks within epics | high |
| T-1768319664765 | Set priorities for all 7 epics and key subtasks | high |

These are planning/review tasks; not strictly “write docs” but tagged #docs (epic/planning docs).

---

## Todo – Clearly docs-focused (high signal)

Tasks that are explicitly about writing or updating documentation:

| ID | Content |
|----|--------|
| **T-1769531792537** | Update migration docs for 4 removed Python fallbacks |
| T-105 | Document gotoHuman API/tools |
| T-134 | Document form fields and usage |
| T-155 | Documentation updated |
| T-186 | Document bridge usage |
| T-223 | User documentation |
| T-238 | Update documentation |
| T-241 | Update migration status docs |
| T-262 | Automation tool documented as intentional Python bridge retention |
| T-264 | Hybrid patterns documented |
| T-324 | Update this document if needed |
| T-330 | Archive this policy document if superseded |
| T-360 | Documentation |
| T-392 | Update README |
| T-393 | Create migration guide |
| T-1768268676474 | Phase 1.7: Update Documentation |
| T-1768317554758 | T1.5.5: Update Documentation for Protobuf Integration |
| T-1768317926161 | … comprehensive documentation for the parallel migration workflow … |
| T-1768317926169 | … document any leftover Python code … |
| T-1768251822699 | Add Package-Level Documentation |

**Suggested next:**  
- **T-1769531792537** – Update migration docs for 4 removed Python fallbacks (direct follow-up to completed removals).  
- **T-186** – Document bridge usage (current bridge behavior is documented in NATIVE_GO_HANDLER_STATUS; could link or add a short “bridge usage” section).  
- **T-393** – Create migration guide (one place for “how to migrate” for contributors).

---

## Todo – Broader / tangentially docs (medium signal)

Tagged #docs but mix of implementation, verification, and docs:

- **Verification / testing:** T-344, T-347, T-348, T-351, T-354, T-357, T-373, T-399–T-404 (verify tools/prompts/resources, Cursor, README, etc.).
- **Implementation then document:** T-262, T-264 (document automation/hybrid patterns); T-247 (performance measured and documented); T-241 (migration status docs).
- **Epic/planning:** T-1768317926143 (scorecard), T-1768317926150 (testing/docs), T-1768317926177 (migration strategy), T-1768317926181 (testing/validation/cleanup), T-1768317926144 (cspell), T-1768317926149 (MLX integration), T-1768317926158 (MLX-enhanced estimates).

Consider: keep #docs only where the main deliverable is “documentation”; drop or add #testing / #migration for verification and implementation tasks.

---

## Done (54) – Sample

Representative completed #docs work:

- **T-1768163111562** – Document tool groups enable/disable functionality  
- **T-1768163517955** – Fix task_workflow documentation inconsistencies  
- **T-1768170934965** – Verify and improve scorecard resource native Go implementation  
- **T-1768249093338** – Extract Framework Abstraction to mcp-go-core  
- **T-1768268669031** – Phase 1.1: Create Config Package Structure  
- **T-1768326545827** – Consolidate logging systems  
- **T-1768326823408** – Investigate analyze_alignment JSON parsing error in git hooks  
- Plus many migration, config, and validation tasks that included or led to docs.

---

## Recommendations

1. **Do next (docs-only):**  
   - **T-1769531792537** – Update migration docs for 4 removed Python fallbacks.  
   - **T-393** – Create migration guide (if not already covered by NATIVE_GO_MIGRATION_PLAN / NEXT_MIGRATION_STEPS).

2. **Triage #docs tag:**  
   - Keep #docs for tasks whose primary outcome is “documentation” or “update doc X”.  
   - Remove #docs from pure verification (e.g. “Verify all 24 tools…”) or implementation-first tasks, or add more specific tags (#testing, #migration).

3. **In Progress:**  
   - The 3 In Progress tasks are epic/planning; close or move to Done when priorities and epic assignments are set and any related doc updates are done.

---

*Generated from Todo2 DB: tasks with tag `docs`.*
