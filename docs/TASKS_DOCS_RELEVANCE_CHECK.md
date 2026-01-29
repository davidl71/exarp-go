# #docs Todo Tasks – Relevance Check

**Date:** 2026-01-29  
**Context:** 3 In Progress #docs tasks deleted. **#docs tag stripped** from 148 tasks (not-relevant); **30 Todo** tasks retain #docs (relevant + tangential).

---

## Summary

| Category | Count | Action |
|----------|-------|--------|
| **Relevant (docs is primary)** | ~22 | Keep #docs |
| **Tangential (doc + implementation)** | ~10 | Keep or add other tags |
| **Not relevant (implementation / test / verify / ops)** | ~91 | Consider removing #docs |

---

## Relevant – docs is the main deliverable (keep #docs)

| ID | Content |
|----|--------|
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
| T-1768317926161 | Comprehensive documentation for parallel migration workflow |
| T-1768317926169 | Document any leftover Python code |
| T-1768251822699 | Add Package-Level Documentation |
| T-1768223926189 | Create Python code removal plan document |
| T-245 | Prompts evaluated (Phase 5) – document Python bridge retention |
| T-247 | Performance improvements measured and documented |

---

## Tangential – doc is one outcome among others

| ID | Content | Note |
|----|--------|------|
| T-1768317926143 | Generate comprehensive project scorecard | Scorecard + docs |
| T-1768317926150 | Testing, performance, complete documentation | Mixed |
| T-1768317926177 | Comprehensive migration strategy | Strategy doc |
| T-1768317926181 | Testing, validation, cleanup of migration | Includes doc |
| T-395 | Code review of existing patterns (CodeLlama) | Review → doc |
| T-396 | Library documentation (Context7) | Doc |
| T-397 | Logical decomposition (Tractatus) | Doc/reasoning |
| T-398 | Latest patterns (Web Search) | Research → doc |

---

## Not relevant – implementation, test, verify, ops, UI (consider removing #docs)

**Approval / workflow (not docs):** T-108, T-117, T-118, T-125  
**Implementation (handlers, server, OAuth, etc.):** T-143–T-148, T-171, T-173, T-174, T-211–T-217, T-227, T-230, T-232, T-278, T-279, T-341–T-366, T-382–T-391, T-77  
**Migration/implementation:** T-1768316808114, T-1768316817909, T-1768316828486, T-1768317926186, T-1768317926188, T-1768318349855, T-1768325408714, T-1768325421830, T-1768325426665  
**Testing / verification:** T-1768170876574, T-1768317926183, T-1768320725711, T-1768325006421, T-306, T-314, T-344, T-347, T-348, T-351, T-354, T-357, T-366, T-373, T-399–T-404  
**Analysis / review (not doc-first):** T-1768223765685, T-1768312778714, T-1768321828916  
**Cleanup / ops / one-time:** T-1768251821268, T-297, T-311, T-405  
**Epic / audit / deployment:** T-1768318471624, T-1768325741814, T-394  
**UI / setup:** T-91, T-64, T-68  
**Lock/alerting/review (implementation or process):** T-320, T-321, T-322, T-325  

These are tagged #docs but their main deliverable is code, tests, verification, or process—not documentation. Removing #docs from them would make the #docs list reflect actual doc work.

---

## Recommendation

1. **Keep #docs** on the ~22 “Relevant” tasks (and optionally the ~10 “Tangential”).
2. **Remove #docs** from the ~91 “Not relevant” tasks if you want #docs to mean “primary deliverable is documentation.” (Removal would need to be done in Todo2 or via a bulk update; exarp-go may not support “remove tag” in one step.)
3. **Next doc tasks to do:** T-186 (Document bridge usage), T-393 (Create migration guide), T-241 (Update migration status docs).
