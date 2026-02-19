# TUI / CLI / MCP Architecture Review

**Date:** 2026-02-19
**Scope:** `internal/cli/tui.go`, `tui3270.go`, `tui_test.go`, `child_agent.go`, `cli.go`, `mode.go`, MCP tool registry

---

## 1. Surface Inventory

| Surface | File(s) | Lines | Framework | Purpose |
|---------|---------|------:|-----------|---------|
| MCP tools | `internal/tools/registry.go` + handlers | ~1,960 | JSON-RPC 2.0 / stdio | AI agent capability API (24 tools) |
| CLI | `internal/cli/cli.go`, `task.go` | ~920 | flag + mcp-go-core | Terminal commands (`exarp-go task list`, etc.) |
| Bubbletea TUI | `internal/cli/tui.go` | 4,582 | charmbracelet/bubbletea + lipgloss | Interactive ANSI terminal UI |
| 3270 TUI | `internal/cli/tui3270.go` | 1,841 | racingmars/go3270 | IBM 3270 mainframe-style TN3270 server |
| Child agent | `internal/cli/child_agent.go` | 513 | stdlib | Shared Cursor CLI agent launcher |
| Mode detection | `internal/cli/mode.go` | 96 | -- | CLI/MCP mode and subcommand detection |
| Tests | `internal/cli/tui_test.go` | 490 | testing + bubbletea | Bubbletea TUI tests (13 cases) |

---

## 2. Feature Matrix

| Capability | MCP tools | CLI | Bubbletea TUI | 3270 TUI |
|---|:---:|:---:|:---:|:---:|
| Task CRUD | `task_workflow` | `task list/create/update/show` | View, select, detail | View, detail, edit |
| Task analysis | `task_analysis` (14 actions) | `-tool task_analysis` | `A` key | -- |
| Task search | `task_workflow` sub_action=list | `--status` flag | `/` search, 5 sort modes | `FIND`/`RESET` |
| Task execution | `task_execute` | `-tool task_execute` | -- | -- |
| Scorecard | `report` action=scorecard | `-tool report` | `p` key + run recs | Option 3 + run recs |
| Config | `generate_config` | `config init/export/convert` | `c` key (view/edit/save) | Option 2 (view only) |
| Handoffs | `session` action=handoff | `session handoff list/end` | `H` key (list/detail/close/approve) | Option 4 (list only) |
| Waves | `task_analysis` + `report` | `-tool task_analysis` | `w` key (expand/move/enqueue) | -- |
| Child agent | -- | `cursor agent` | `E`/`i`/`L` keys | Option 6 + `RUN` |
| Background jobs | -- | -- | `b` key (list/detail) | -- |
| Health/security/memory/LLM | 10+ tools | `-tool ...` | -- | -- |

---

## 3. Architecture: Current vs. Recommended

### Current: Three Independent Data Paths

```
MCP Tools ──→ internal/tools handlers ──→ database/config
CLI       ──→ handleTaskCommand       ──→ database/config
TUI       ──→ database.GetTask()      ──→ database/config
TUI3270   ──→ database.GetTask()      ──→ database/config
```

### Recommended: MCP Tools as Single Capability API

```
MCP Tools ──→ internal/tools handlers ──→ database/config
CLI       ──→ server.CallTool()       ──→ MCP Tools
TUI       ──→ server.CallTool()       ──→ MCP Tools
TUI3270   ──→ server.CallTool()       ──→ MCP Tools
```

---

## 4. Strengths

### Bubbletea TUI
- Feature-rich: 8 views (tasks, config, scorecard, handoffs, waves, jobs, analysis, help)
- Responsive layout (narrow/medium/wide), lipgloss styling
- Clean Elm architecture (Model/Update/View), async tea.Cmd
- 13 tests covering navigation, selection, resize, refresh, quit, loading

### 3270 TUI
- Authentic ISPF-style: command line, PF keys, fixed columns, SCROLL indicator
- Clean transaction pattern (`go3270.Tx` returning next transaction)
- Daemon mode with PID file and signal handling
- Full command parser (FIND, EDIT, VIEW, TOP, BOTTOM, RUN)

### Shared
- `child_agent.go` properly shared between both TUIs
- Both support the same core workflows (tasks, scorecard, handoffs, child agent)

---

## 5. Issues Found

### 5.1 Feature Creep Risk
- Both TUIs bypass MCP tools and call `database.*`, `tools.*`, `config.*` directly
- Bubbletea TUI has features no other surface has (jobs, wave move, queue enqueue)
- 3270 TUI reimplements logic independently (editor, recommendation runner)

### 5.2 Code Quality
- `tui.go` is 4,582 lines in one file (should be split into ~8 files)
- `model` struct has 50+ fields ("god object")
- `Update()` is ~1,200 lines of nested mode/key dispatch
- Mode is a bare string with 8+ values (no enum, no compile-time checks)
- 15+ message types defined inline (could be in separate file)

### 5.3 Duplication
- `truncatePad` (tui.go) and `t3270Pad` (tui3270.go): same algorithm
- `recommendationToCommand` and `recommendationToCommand3270`: same mapping
- `loadTasksForStatus` in tui3270.go, similar `loadTasks` cmd in tui.go
- Child agent menus duplicated between `childAgentMenuTransaction` and `handleCommand`

### 5.4 3270-Specific
- `extractMenuOption`: triple-fallback parsing could produce unexpected matches
- `showChildAgentResultTransaction`: message wrapping truncates at 2 lines (should use `splitIntoLines`)
- `daemonize`: doesn't truly fork (blocks on `serverFunc()`)
- Zero test coverage

### 5.5 Missing Abstractions
- No shared View interface between modes
- No command registry (keybindings and ISPF commands are independent)
- No typed tool-call adapter (raw JSON everywhere)
- No formal state machine for mode transitions

---

## 6. Recommended Abstractions (Priority Order)

| # | Abstraction | Effort | Impact |
|---|---|---|---|
| 1 | **View interface + per-view files** | Medium | Splits 4.5k file, enables testing |
| 2 | **Tool call adapter** (typed functions over `server.CallTool`) | Low | Single data path, eliminates direct DB calls |
| 3 | **Shared command registry** | Low | Eliminates 3-way duplication |
| 4 | **Per-view state structs** | Medium | Cleaner model, prevents field collision |
| 5 | **Mode transition state machine** | Low | Documents + enforces valid navigation |
| 6 | **3270 screen builder DSL** | Medium | Cleaner layout, responsive to terminal size |
| 7 | **Unified loading/error pattern** | Low | Removes boilerplate across views |
| 8 | **3270 test coverage** (pure functions) | Low | Catches regressions |

---

## 7. Feature Creep Prevention Rules

1. **MCP tools are the canonical capability layer** — all new business logic goes there first
2. **TUIs are presentation-only** — they call `server.CallTool()`, never `database.*` or `tools.*` directly
3. **New TUI feature checklist:**
   - Does the MCP tool already support it? If no, add it there first
   - Is it a presentation concern (sort, highlight, layout)? OK for TUI
   - Is it a new business capability (wave reorder, queue enqueue)? Must be MCP tool action
   - Does the other TUI need it too? Definitely a tool-layer feature
4. **CLI mirrors MCP** — `exarp-go task list` should call `server.CallTool("task_workflow")`, not custom handlers

---

## 8. Dead Code (Confirmed)

### High Confidence — Safe to Remove

| Location | Symbol | Reason |
|----------|--------|--------|
| `internal/cli/task.go:91-95` | `handleTaskList` | Legacy wrapper, zero callers |
| `internal/cli/task.go:189-193` | `handleTaskUpdate` | Legacy wrapper, zero callers |
| `internal/cli/task.go:242-246` | `handleTaskCreate` | Legacy wrapper, zero callers |
| `internal/tools/task_workflow_common.go:647-663` | `handleTaskWorkflowListPlaceholder` | Explicit placeholder + `var _ =` suppression |
| `internal/tools/task_workflow_common.go:637-644` | `extractTaskIDs` | Comment says "for approve file fallback" — never called |
| `internal/tools/tool_catalog.go:319-391` | `handleToolCatalogList` | Switch doesn't route "list"; action became a resource |
| `internal/cli/tui.go:89` | `colIDWide` | Defined, never referenced |
| `internal/cli/tui.go:94-95` | `colDescMed`, `colDescWide` | Defined, never referenced |

### Medium Confidence — Verify Before Removing

| Location | Symbol | Reason |
|----------|--------|--------|
| `internal/cli/cli.go:274` | `IsTTY()` (exported) | Only caller is inside `cli.go` — could be unexported |
| `internal/cli/cli.go:281` | `IsTTYStdout()` (exported) | Only called by `ColorEnabled()` in same file |
| `internal/cli/cli.go:287` | `ColorEnabled()` (exported) | Zero callers outside `cli.go` |
| `internal/cli/child_agent.go:95` | `AgentBinary()` (exported) | Zero callers; comment says "prefer agentCommand()" |
| `internal/tools/task_workflow_common.go:2378` | `normalizePriority` | Redundant wrapper around `NormalizePriority` |
| `internal/tools/normalization.go:175-202` | `IsPendingStatusNormalized`, `IsCompletedStatusNormalized`, `IsActiveStatusNormalized`, `IsReviewStatusNormalized` | Only used in `normalization_test.go`; production code uses different status checkers |

---

## 9. Maintainability Issues for AI Agents

### 9.1 Large Files (Exceed Context Windows)

| File | Lines | Problem |
|------|------:|---------|
| `internal/cli/tui.go` | 4,582 | All views, messages, styles, helpers in one file |
| `internal/tools/task_analysis_shared.go` | ~4,000 | All 14 analysis actions in one file |
| `internal/tools/task_workflow_common.go` | 2,775 | All workflow handlers in one file |
| `internal/tools/report.go` | 2,285 | Data fetching + markdown generation + formatting |
| `internal/tools/handlers.go` | 1,962 | Tool dispatch + argument extraction |

**Impact:** AI agents must read partial files, losing cross-function context. A 4,582-line file requires 2-3 reads and often misses related code.

**Fix:** Split each into 3-5 focused files by concern (views, messages, formatting, etc.)

### 9.2 God Object: `model` struct (tui.go:236)

50+ fields spanning 8 views. Any agent editing one view must load the entire struct definition to understand field names and avoid collisions.

**Fix:** Compose from per-view structs (`TaskModel`, `ConfigModel`, `WaveModel`, etc.)

### 9.3 Magic Strings (No Constants)

Status values (`"Todo"`, `"In Progress"`, `"Done"`, `"Review"`), priority values (`"low"`, `"medium"`, `"high"`), and backend names (`"fm"`, `"mlx"`, `"ollama"`) are scattered as string literals across 20+ files.

**Impact:** AI agents propose changes with typos (e.g., `"in_progress"` vs `"In Progress"`) because there's no canonical constant to reference.

**Fix:** Define constants in `internal/models/constants.go` and use them everywhere.

### 9.4 Multiple Data Paths (Same Operation)

Task loading happens via:
1. `database.ListTasks()` (direct DB)
2. `tools.LoadTodo2Tasks()` (JSON fallback)
3. `server.CallTool("task_workflow")` (MCP)
4. `handleTaskCommand(server, args)` (CLI)

An AI agent fixing a task-loading bug must identify which path the caller uses. Four paths = four places to check.

**Fix:** Route all through MCP tools (section 3 above).

### 9.5 Global State

`database.DB`, `CLIOutputOpts`, `ollamaHTTPClient`, `DefaultLocalAI`, `globalFileCache` — package-level mutable singletons. These create hidden dependencies that AI agents miss when reasoning about function behavior.

**Fix:** Pass as constructor parameters or context values.

### 9.6 Inconsistent Task ID Handling

Task ID parsing/validation exists in three places:
- `ParseTaskIDsFromParams` (task_workflow_common.go)
- `taskIDRegex` (plan_waves.go)
- `IsValidTaskID` (database/tasks.go)

**Fix:** Single `models.TaskID` type with `Parse()`, `Validate()`, `String()` methods.

---

## 10. Refactoring Priorities

### Tier 1 — Quick wins (1-2 hours each, high AI-agent impact)
1. Remove confirmed dead code (section 8, high confidence)
2. Extract magic strings into `models/constants.go`
3. Unexport or remove unused exported functions

### Tier 2 — Medium effort (half day each)
4. Split `tui.go` into per-view files
5. Add typed tool-call adapter layer
6. Centralize task ID handling

### Tier 3 — Larger refactors (1+ day each)
7. Route CLI/TUI through `server.CallTool()` instead of direct DB
8. Decompose `model` struct into per-view state
9. Replace global state with dependency injection
10. Split `task_analysis_shared.go` and `task_workflow_common.go`

---

## 11. References

- `internal/cli/tui.go` — Bubbletea TUI (4,582 lines)
- `internal/cli/tui3270.go` — 3270 TUI (1,841 lines)
- `internal/cli/child_agent.go` — Shared child agent launcher
- `internal/cli/tui_test.go` — TUI tests
- `internal/tools/registry.go` — MCP tool registry (24 tools)
- `docs/ISPF_PATTERNS_RESEARCH.md` — ISPF design patterns
- `docs/3270_TUI_IMPLEMENTATION.md` — 3270 implementation guide
- `docs/WORKFLOW_DSL_AND_MESSAGE_BUS.md` — Workflow DSL and message bus alignment
