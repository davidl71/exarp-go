# Prompts: Migration and Obsolescence Plan

**Created:** 2026-01-29  
**Purpose:** Which Python prompts are obsolete (can be removed), which are already migrated to Go (Python can be deprecated), and which need migration to Go.

**2026-01-29:** Steps 1–3 done. All 16 “need migration” prompts added to Go. **Python prompt stub removed:** `project_management_automation/prompts.py` deleted; `project_scorecard` and `project_overview` use `prompts_count = 0` (prompts are in exarp-go). Prompt manifest lives in `context_primer.py` only.

---

## Summary

| Category | Count | Action |
|----------|--------|--------|
| **Migrated** (in Go; Python can be removed) | 18 | Deprecate/remove from Python |
| **Need migration** (add to Go) | 16 | Add to `internal/prompts/templates.go` + registry |
| **Obsolete** (remove, do not migrate) | 5 | Remove from Python only |
| **Python PROMPTS dict total** | 38 keys | |
| **Go promptTemplates total** | 19 keys | |

---

## 1. Migrated — Python can be removed/deprecated

These prompts exist in **Go** (`internal/prompts/templates.go`) with equivalent or updated content. The Python entries are redundant; remove them from `project_management_automation/prompts.py` (and from the `PROMPTS` dict) or mark deprecated.

| Python PROMPTS key | Python constant | Go key | Notes |
|--------------------|-----------------|--------|--------|
| `task_alignment` | TASK_ALIGNMENT_ANALYSIS | `align` | Same intent, Go uses analyze_alignment tool |
| `duplicate_cleanup` | DUPLICATE_TASK_CLEANUP | `dups` | Same intent |
| `task_sync` | TASK_SYNC | `sync` | Same intent |
| `task_discovery` | TASK_DISCOVERY | `discover` | Go uses task_discovery(action=…); content aligned |
| `security_scan_all` | SECURITY_SCAN_ALL | `scan` | One Go prompt covers all languages via tool params |
| `pre_sprint_cleanup` | PRE_SPRINT_CLEANUP | `pre_sprint` | Same workflow |
| `post_implementation_review` | POST_IMPLEMENTATION_REVIEW | `post_impl` | Same workflow |
| `daily_checkin` | DAILY_CHECKIN | `daily_checkin` | Same |
| `sprint_start` | SPRINT_START | `sprint_start` | Same |
| `sprint_end` | SPRINT_END | `sprint_end` | Same |
| `project_scorecard` | PROJECT_SCORECARD | `scorecard` | Same |
| `project_overview` | PROJECT_OVERVIEW | `overview` | Same |
| `project_dashboard` | PROJECT_DASHBOARD | `dashboard` | Same |
| `memory_system` | MEMORY_SYSTEM | `remember` | Same |
| `config_generation` | CONFIG_GENERATION | `config` | Same |
| `mode_suggestion` | MODE_SUGGESTION | `mode` | Same (Agent vs Ask) |
| `context_management` | CONTEXT_MANAGEMENT | `context` | Same |

**Go-only (no Python key):** `task_update` — task status update workflow; Python has no direct equivalent key in PROMPTS. No Python removal needed for this.

**Action:** Remove the 17 Python constants and their PROMPTS entries above (or replace with a stub that points to “use stdio://prompts/{go_key}”). Update any Python code that imports by name (e.g. `from prompts import PROMPTS` → use count from Go or a small manifest if needed for project_scorecard.py / project_overview.py).

---

## 2. Obsolete — Remove from Python, do not migrate

These are redundant, thin variants, or superseded. Remove from Python only; do not add to Go.

| Python PROMPTS key | Python constant | Reason |
|--------------------|-----------------|--------|
| `doc_quick_check` | DOCUMENTATION_QUICK_CHECK | Thin variant of doc_health_check (create_tasks=False). One “docs” prompt with a sentence on the flag is enough; Go can add a single `docs` prompt later if needed. |
| `security_scan_python` | SECURITY_SCAN_PYTHON | Redundant: Go `scan` and Python `security_scan_all` already cover “all languages”; tool param `languages=["python"]` suffices. |
| `security_scan_rust` | SECURITY_SCAN_RUST | Same as above for Rust. |
| `automation_high_value` | AUTOMATION_HIGH_VALUE | Thin variant of automation_discovery (min_value_score=0.8). One automation discovery prompt with optional param is enough. |
| `mode_select` | MODE_SELECT | Long “mode-aware workflow selection” guide. Overlaps with MODE_SUGGESTION / Go `mode`. Keeping one mode prompt is enough; this one can be dropped. |

**Action:** Delete these 5 constants and their 5 PROMPTS entries from `prompts.py`. Do not add to Go.

---

## 3. Need migration — Add to Go

These exist only in Python and should be added to Go so that MCP clients (and Cursor) get them via exarp-go. Order below is a suggested migration sequence.

### 3.1 Documentation (2)

| Python key | Suggested Go key | Notes |
|------------|------------------|--------|
| `doc_health_check` | `docs` | Documentation health check; create_tasks optional. |
| *(obsolete: doc_quick_check — do not migrate)* | — | |

### 3.2 Automation (2)

| Python key | Suggested Go key | Notes |
|------------|------------------|--------|
| `automation_discovery` | `automation_discover` | Discover automation opportunities; reference automation(action="discover"). |
| *(obsolete: automation_high_value — do not migrate)* | — | |

### 3.3 Workflow (4)

| Python key | Suggested Go key | Notes |
|------------|------------------|--------|
| `weekly_maintenance` | `weekly_maintenance` | Weekly maintenance workflow (docs, dups, security, sync). |
| `task_review` | `task_review` | Backlog hygiene: duplicates, alignment, clarify, hierarchy, approve. |
| `project_health` | `project_health` | Full health check: server, docs, testing, security, cicd, alignment. |
| `automation_setup` | `automation_setup` | One-time setup: setup_hooks(git), setup_hooks(patterns), cron. |

### 3.4 Wisdom / Advisor (2)

| Python key | Suggested Go key | Notes |
|------------|------------------|--------|
| `advisor_consult` | `advisor_consult` | Consult advisor by metric/stage/tool; devwisdom integration. |
| `advisor_briefing` | `advisor_briefing` | Morning briefing from report(briefing) and scorecard. |

### 3.5 Personas (8)

| Python key | Suggested Go key | Notes |
|------------|------------------|--------|
| `persona_developer` | `persona_developer` | Developer daily workflow (checkin, commit, PR, debugging). |
| `persona_project_manager` | `persona_project_manager` | PM: standup, sprint planning/retro, status report. |
| `persona_code_reviewer` | `persona_code_reviewer` | Code review workflow and checklist. |
| `persona_executive` | `persona_executive` | Executive: weekly check, monthly review, dashboard. |
| `persona_security` | `persona_security` | Security engineer: daily/weekly scan, audit. |
| `persona_architect` | `persona_architect` | Architect: architecture review, tech debt. |
| `persona_qa` | `persona_qa` | QA: testing status, sprint testing review. |
| `persona_tech_writer` | `persona_tech_writer` | Tech writer: doc health, targets. |

**Action:** Add the 16 prompts above (1 docs + 1 automation + 4 workflow + 2 advisor + 8 persona) to `internal/prompts/templates.go` and register them in `internal/prompts/registry.go`. Use Python constant text as the source; adjust tool names to exarp-go (e.g. `report(action="scorecard")`, `health(action="docs")`) where needed. After migration, treat the corresponding Python entries as “migrated” and remove/deprecate them per Section 1.

---

## 4. Python consumers to update

Before or after removing prompts from Python, update:

| Consumer | Change |
|----------|--------|
| `project_management_automation/tools/project_scorecard.py` | Uses `from prompts import PROMPTS` for `prompts_count`. Prefer: read prompt count from exarp-go (e.g. sanity-check or a small manifest) or hardcode expected count for “Python legacy” path. |
| `project_management_automation/tools/project_overview.py` | Same as above for PROMPTS. |
| `project_management_automation/resources/prompt_discovery.py` | PROMPT_MODE_MAPPING and PROMPT_PERSONA_MAPPING reference Python PROMPTS keys. Either: (a) keep a minimal manifest of prompt names/categories for discovery only, or (b) migrate discovery to Go (e.g. resource or tool that returns prompt list by mode/persona). |

---

## 5. Implementation order

1. **Obsolete:** Remove the 5 obsolete Python constants and their PROMPTS entries.
2. **Migrated:** Remove or deprecate the 17 migrated Python constants and PROMPTS entries; update project_scorecard.py and project_overview.py to not depend on full PROMPTS for count (or use manifest).
3. **Need migration:** Add to Go in batches:
   - Batch 1: `docs`, `automation_discover`, `weekly_maintenance`, `task_review`, `project_health`, `automation_setup` (6).
   - Batch 2: `advisor_consult`, `advisor_briefing` (2).
   - Batch 3: All 8 `persona_*` prompts (8).
4. **Discovery:** Decide whether to keep a minimal Python manifest for prompt_discovery or implement mode/persona prompt list in Go and retire Python discovery.

---

## 6. Go registration and sanity-check

- After adding prompts, update `internal/prompts/registry.go` with the new names and short descriptions.
- Update `cmd/sanity-check/main.go` (and any docs) that hardcode expected prompt count (e.g. 18 → 19 then to 37 if all 18 “need migration” are added).

---

## 7. Reference: Python PROMPTS keys (full list)

All 38 keys in `prompts.PROMPTS`:

- **Documentation:** doc_health_check, doc_quick_check  
- **Tasks:** task_alignment, duplicate_cleanup, task_sync, task_discovery  
- **Security:** security_scan_all, security_scan_python, security_scan_rust  
- **Automation:** automation_discovery, automation_high_value  
- **Workflow:** pre_sprint_cleanup, post_implementation_review, weekly_maintenance, daily_checkin, sprint_start, sprint_end, task_review, project_health, automation_setup  
- **Reports:** project_scorecard, project_overview, project_dashboard  
- **Wisdom:** advisor_consult, advisor_briefing  
- **Memory/Config:** memory_system, config_generation  
- **Mode/Context:** mode_suggestion, mode_select, context_management  
- **Personas:** persona_developer, persona_project_manager, persona_code_reviewer, persona_executive, persona_security, persona_architect, persona_qa, persona_tech_writer  

Go keys today (19): align, discover, config, scan, scorecard, overview, dashboard, remember, daily_checkin, sprint_start, sprint_end, pre_sprint, post_impl, sync, dups, context, mode, task_update.
