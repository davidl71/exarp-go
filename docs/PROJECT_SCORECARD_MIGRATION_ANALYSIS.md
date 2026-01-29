# Project Scorecard Migration to Native Go

**Date:** 2026-01-29  
**Context:** stdio://scorecard and report(action=scorecard) are already native Go for Go projects. This doc identifies what can be migrated so Python callers use Go when possible.

---

## Migration status

| Item | Status | Notes |
|------|--------|-------|
| **1.** project_overview._get_health_metrics() → Go | ✅ **Done** | Go projects call `exarp-go -tool report -args '{"action":"scorecard","output_format":"json"}'`; else Python. |
| **2.** consolidated_reporting.report(scorecard) → Go | ✅ **Done** | Go projects delegate to exarp-go; else Python or "scorecard only for Go" message. |
| **3.** project_scorecard.py shrink | ✅ **Done** | Thin wrapper delegates to exarp-go for Go projects; non-Go returns clear error. |
| **4.** LEGACY_PROMPTS_COUNT in Python | ⏳ **Optional follow-up** | Keep import or replace with constant (e.g. `0`) for non-Go/overview metrics. See Todo2 task. |

---

## Current state

### Go (already native)

| Component | Location | Notes |
|-----------|----------|------|
| **stdio://scorecard** | `internal/resources/scorecard.go` | Uses `tools.GenerateGoScorecard`; returns JSON error for non-Go. |
| **report(action=scorecard)** | `internal/tools/handlers.go` | Uses `GenerateGoScorecard`, `GoScorecardToMap`, MLX/wisdom enhancement. Go projects only. |
| **Scorecard generation** | `internal/tools/scorecard_go.go` | `GenerateGoScorecard`, `GoScorecardResult`, `FormatGoScorecard`. |
| **Scorecard helpers** | `internal/tools/scorecard_mlx.go` | `GoScorecardToMap`, `ExtractBlockers`, `FormatGoScorecardWithMLX`, wisdom. |

### Python (delegates to Go for Go projects)

| Consumer | File | What it does |
|----------|------|----------------|
| **project_overview** | `project_management_automation/tools/project_overview.py` | `_get_health_metrics()` calls exarp-go for Go projects; else Python `generate_project_scorecard`. |
| **consolidated_reporting** | `project_management_automation/tools/consolidated_reporting.py` | `report(action="scorecard")` delegates to exarp-go for Go projects; else Python or "Go only" message. |
| **project_scorecard** | `project_management_automation/tools/project_scorecard.py` | Thin wrapper: delegates to exarp-go for Go projects; non-Go returns clear error. |

---

## What was migrated (1–3 done)

### 1. **project_overview._get_health_metrics() → use Go for Go projects** ✅

- **Before:** Always called Python `generate_project_scorecard("json", False, None)`.
- **After:** If `go.mod` exists, call `exarp-go -tool report -args '{"action":"scorecard","output_format":"json"}'` and parse JSON; else keep Python path.
- **Benefit:** Overview uses same numbers as stdio://scorecard for Go projects.

### 2. **consolidated_reporting.report(action="scorecard") → delegate to Go for Go projects** ✅

- **Before:** Always called Python `generate_project_scorecard(...)`.
- **After:** If Go project, call exarp-go and return tool output; else keep Python or "scorecard only for Go projects" message.
- **Benefit:** Scripts and Python MCP callers get native Go scorecard for Go projects.

### 3. **project_scorecard.py shrink** ✅

- **Before:** ~1400 lines; full `generate_project_scorecard`, wisdom_client, metric logic.
- **After:** Thin wrapper: delegates to exarp-go for Go repos; non-Go returns clear error. Most Python logic removed.

### 4. **prompts.py and project_scorecard/project_overview** (optional follow-up)

- `project_scorecard.py` and `project_overview.py` import `LEGACY_PROMPTS_COUNT` from `prompts.py` for metrics.
- Options: keep import for backward compatibility, or replace with a constant (e.g. `0`) in Python that still runs for non-Go or overview metrics. Tracked as Todo2 task.

---

## Recommended order (1–3 done)

1. ✅ **Implement (1):** Done. `project_overview.py` calls exarp-go for Go projects; else Python.
2. ✅ **Implement (2):** Done. `consolidated_reporting.py` delegates to exarp-go for Go projects; else Python or "Go only" message.
3. ✅ **Shrink (3):** Done. `project_scorecard.py` is a thin wrapper delegating to exarp-go for Go projects.

---

## Verification and edge cases

- **CLI stdout:** `exarp-go -tool report -args '{"action":"scorecard","output_format":"json"}'` prints `Executing tool: ...`, `Arguments: ...`, and `Result:\n` before the JSON. Python callers (`_call_go_scorecard_json`, consolidated_reporting) extract JSON by taking the content after `Result:\n` before parsing.
- **Verification:** `_get_health_metrics` and `report(action="scorecard", output_format="json")` were tested in a Go project; both use exarp-go and return expected `overall_score`, `production_ready`, `blockers`, `scores`.

---

## Non-goals (for this migration)

- **Porting full project_overview to Go** – not required for “scorecard-related” migration; overview can stay Python and only consume Go scorecard when available.
- **Porting project_scorecard.py line-by-line to Go** – Go already has `GenerateGoScorecard`; the migration is “call Go from Python” for Go projects, not reimplement Python in Go.

---

## References

- Go scorecard: `internal/tools/scorecard_go.go`, `internal/tools/scorecard_mlx.go`
- Report handler: `internal/tools/handlers.go` (handleReport, action=scorecard)
- Resource: `internal/resources/scorecard.go` (stdio://scorecard)
- Python consumers: `project_management_automation/tools/project_overview.py`, `project_management_automation/tools/consolidated_reporting.py`, `project_management_automation/tools/project_scorecard.py`
