# Feature Requests

Tracked enhancement requests for exarp-go. Add new entries at the top with a short id for cross-references.

---

## Cursor Skills: tool that writes SKILL.md (create_skill / generate_config)

**Status:** Open  
**Id:** `cursor-skills-write-tool`  
**See also:** `skills/README.md` (bundled skills, future enhancement note)

### Summary

Add an exarp-go **tool** (e.g. `create_skill` or an option in `generate_config`) that writes Cursor Agent Skill files (SKILL.md) into the user's Cursor skills path (`~/.cursor/skills/` or `.cursor/skills/`), so users can install or update exarp-go–related skills via the MCP server instead of copying by hand.

### Current behavior

- exarp-go ships **bundled skills** in `skills/` (use-exarp-tools, task-workflow, report-scorecard). Users install by copying those directories into `~/.cursor/skills/` or `.cursor/skills/`.
- No tool exists to create or update SKILL.md from exarp-go.

### Requested behavior

1. **Tool that writes SKILL.md**
   - New tool (e.g. `create_skill`) or extended `generate_config` that:
     - Accepts skill name, description, body (or path to template).
     - Writes a SKILL.md (and optionally a directory) to a target path.
   - Target path: user-configurable or inferred (e.g. `~/.cursor/skills/<name>/` or project `.cursor/skills/<name>/`).
   - Optional: “install bundled exarp-go skills” action that copies from repo `skills/` into the user's chosen skills path.

2. **Integration**
   - Document in `skills/README.md` when the tool is available.
   - Reuse or align with `generate_config` patterns (e.g. `output_dir`, `dry_run`) if implemented as an extension.

### Use case

- Users who already use exarp-go as MCP want to add or update Cursor skills without leaving the IDE or copying files manually.
- Project maintainers want to offer “install exarp-go skills” from the tool instead of instructing copy-paste.

---

## Multi-language and per-language scorecard support

**Status:** Open  
**Id:** `scorecard-multilang`  
**See also:** `internal/tools/handlers.go` (comment in `handleReport` for `action=scorecard`)

### Summary

`report(action=scorecard)` and `stdio://scorecard` should support **any** project: multi-language codebases (C++, Python, Rust, TypeScript, Swift, Go, etc.) and optional **per-language** scorecards, instead of being limited to Go-only or Python/exarp-go–style projects.

### Current behavior

- **Go projects:** Native `GenerateGoScorecard` in `internal/tools/scorecard_go.go` (go.mod, Go tests, coverage, etc.). Works well for Go-only repos.
- **Non-Go projects:** Fallback to Python bridge → `generate_project_scorecard` in `project_management_automation/tools/project_scorecard.py`. That implementation is:
  - **Python-centric** (e.g. `*.py`, `project_management_automation/tools`, `prompts.PROMPTS`).
  - **exarp-go/pma–oriented** (assumes a certain layout and tooling).
- **Per-language scorecards:** Not supported. No breakdown by C++, Rust, TypeScript, etc.

For multi-language or arbitrary projects, scorecards are often **incomplete or skewed** (e.g. only Python counted, wrong “tools/prompts” logic, no C++/Rust/TS metrics).

### Requested behavior

1. **Language-agnostic / multi-language scorecard**
   - Detect project languages from the repo (e.g. `go.mod`, `Cargo.toml`, `CMakeLists.txt`, `package.json`, `*.py`, `requirements.txt`, etc.).
   - Collect metrics per detected language (files, tests, config, CI) and produce an **overall** scorecard that aggregates them.
   - Avoid hardcoding exarp-go or project_management_automation layout.

2. **Optional per-language scorecards**
   - Support a mode or parameter to emit **one scorecard per language** (e.g. `PROJECT_SCORECARD_CPP.md`, `PROJECT_SCORECARD_PY.md`), with:
     - Source/test file counts and patterns.
     - Config and CI presence.
     - Framework and purpose (e.g. “C++: Core logic, Catch2; Python: integration, pytest”).
   - Allow a **config file** (e.g. `scorecard_languages.json`) to define:
     - `source_patterns`, `test_patterns`, `include_dirs`, `exclude_dirs`, `config_markers`, `ci_workflows`, `framework`, `purpose`, `status` per language.

3. **Unified entrypoint**
   - Keep `report(action=scorecard)` and `stdio://scorecard` as the main entrypoints; extend parameters, e.g.:
     - `per_language` (bool), `languages` (list or “all”), `config` (path to `scorecard_languages.json`), `output_dir` for per-language files.

### Use case

- **ib_box_spread_full_universal** (and similar multi-language projects): C++20, Python, Rust, TypeScript, Swift. Users want:
  - One overall scorecard for the repo.
  - Optional per-language scorecards (C++, Python, Rust, TypeScript, Swift) with correct patterns and CI awareness, without maintaining a separate project-specific scorecard script.

### Implementation notes

- A possible approach is to add a **language-agnostic** scorecard implementation (native Go or evolved Python) that:
  - Scans the project root for language markers and optional `scorecard_languages.json`.
  - Applies patterns and exclusions per language to count source/test files, config, and CI.
  - Outputs overall and, if requested, per-language markdown (and optionally JSON).
- The existing `GenerateGoScorecard` can remain for Go projects; the new logic can be used when `!IsGoProject()` or when `per_language`/multi-language is requested.
- Reference: some projects already use an ad-hoc `scorecard_languages.json` and `generate_project_scorecard.py` with `--per-language`; the goal is to absorb that into exarp-go so one tool serves “any scorecard.”

---

## How to add a feature request

1. Copy the block above (from `##` to `---`).
2. Replace the title, status, id, summary, current/requested behavior, use case, and implementation notes.
3. Insert it after this “How to add” section and before older entries.
