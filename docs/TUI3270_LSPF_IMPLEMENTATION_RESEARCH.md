# tui3270 Implementation Research (lspf + 3270BBS)

## Goal

Capture reusable patterns from:
- `daniel64/lspf` (ISPF dialogue manager, C++/ncurses)
- `moshix/3270BBS` (3270 terminal UX, go3270-based)
- `racingmars/go3270` (library used by both exarp-go and 3270BBS)

...and map them onto exarp-go's current `tui3270` implementation so the next work can be planned as concrete, file-backed changes.

**Long-term direction:** exarp-go should behave like a native “APP” in the 3270 terminal ecosystem — consistent PF keys, screen layout, command bar, session management, and help conventions that match existing mainframe tooling.

## Sources

- External references:
  - `https://github.com/daniel64/lspf`
  - `https://github.com/moshix/3270BBS`
  - `https://github.com/racingmars/go3270`
- Local implementation:
  - `internal/cli/tui3270.go`
  - `internal/cli/tui3270_menu.go`
  - `internal/cli/tui3270_helpers.go`
  - `internal/cli/tui3270_screen_*.go`

## What lspf contributes

The strongest reusable ideas from `lspf` are not raw rendering details but interaction structure:

1. **Dedicated screen zones**
   - main content area
   - status/OIA line
   - command/swap bar
2. **First-class command line**
   - commands are a primary interaction mode, not an afterthought
3. **Field-centric panels**
   - protected vs writable fields are explicit
4. **Current-row pointer semantics**
   - lists are driven by a stable “current row” model
5. **PF-key command mapping**
   - action dispatch is standardized instead of screen-local ad hoc branching
6. **Scrollable regions with explicit indicators**
   - TOP/MORE/BOTTOM/CSR-style feedback
7. **Panel/layout separation from application actions**
   - screens describe layout; commands/services mutate state

## Current exarp-go tui3270 mapping

### 1. Core state and navigation

**Existing files**
- `internal/cli/tui3270.go`

**Already present**
- `tui3270State`
- `sessionStack` with `pushSession` / `popSession`
- transaction-based screen navigation via `go3270.RunTransactions`

**Assessment**
- exarp-go already approximates lspf's panel stack and swap behavior well
- the current state object is the right place to formalize command bar state, OIA state, and current-row semantics across all screens

### 2. Command bar

**Existing files**
- `internal/cli/tui3270_menu.go`
- `internal/cli/tui3270_screen_tasklist.go`
- `internal/cli/tui3270_helpers.go`

**Already present**
- `Command ===>` field on menu and task list
- `state.command`
- `handleCommand(...)` flow in helpers

**Gap**
- command bar behavior is screen-specific rather than standardized
- there is no shared command-bar rendering/helper layer reused by all screens
- command discoverability and completion/retrieve/history are limited

### 3. Status/OIA line

**Existing files**
- `internal/cli/tui3270_helpers.go`
- `internal/cli/tui3270_screen_tasklist.go`
- screen-specific files that append PF-key lines/status rows

**Already present**
- status line helpers: `t3270StatusRow`, `t3270PFRow`, `showLoadingOverlay`
- task list shows counters, filter state, cursor position

**Gap**
- no unified OIA/status model shared across screens
- status content is hand-built per screen
- command status, mode status, transient action result, and screen identity are not normalized

### 4. PF-key mapping

**Existing files**
- `internal/cli/tui3270_menu.go`
- `internal/cli/tui3270_screen_tasklist.go`
- `internal/cli/tui3270_screen_taskdetail.go`
- other `tui3270_screen_*.go` files

**Already present**
- strong PF usage: PF1 Help, PF3 Back, PF7/8 scroll, PF9 filter, PF11 swap, PF12 cancel
- detail screen uses PF4/PF5/PF6/PF10 for state changes

**Gap**
- PF handling is duplicated across screens
- no central registry/table describing allowed PF actions by screen
- PF help strings can drift from actual behavior

### 5. Current-row pointer and table/list behavior

**Existing files**
- `internal/cli/tui3270_screen_tasklist.go`
- `internal/cli/tui3270_helpers.go`

**Already present**
- `state.cursor`
- `state.listOffset`
- reverse-video current row
- line-command column (`S/E/D/I`)

**Gap**
- current-row semantics are mostly task-list-specific
- other list-like screens could share a common CRP/list window abstraction
- scroll/window calculations are duplicated per screen instead of being a reusable model

### 6. Scrollable areas

**Existing files**
- `internal/cli/tui3270_screen_tasklist.go`
- `internal/cli/tui3270_helpers.go`
- `internal/cli/tui3270_screen_taskdetail.go`

**Already present**
- task list uses `SCROLL ===>` plus `CS/MORE/BOTTOM/TOP`
- helper row math exists (`t3270MaxVisible`, `t3270ContentMaxRow`)

**Gap**
- detail screens still behave like static views instead of reusable scrollable regions
- long content handling is split into ad hoc truncation/line splitting
- scroll indicators are not consistently shared across all long screens

### 7. Screen layout separation

**Existing files**
- `internal/cli/tui3270_screen_*.go`
- `internal/cli/tui3270_helpers.go`

**Already present**
- good per-screen file separation
- helper functions for colors, rows, validators, formatting

**Gap**
- no small layout framework for common 3270 screen regions
- repeated hand assembly of title, command row, status row, PF row
- screen files still combine layout composition with action dispatch

## Recommended implementation themes

### Theme 1: Standardized 3270 chrome

Create shared builders/helpers for:
- title row
- command row
- OIA/status row
- PF help row

**Primary files**
- `internal/cli/tui3270_helpers.go`
- all `internal/cli/tui3270_screen_*.go`

### Theme 2: Central PF/action tables

Introduce declarative screen action metadata so:
- PF help text is generated
- handler availability is centralized
- screens cannot drift between displayed PF keys and actual behavior

**Primary files**
- new helper or mapping file under `internal/cli/`
- `tui3270_screen_tasklist.go`
- `tui3270_screen_taskdetail.go`
- `tui3270_menu.go`

### Theme 3: Reusable list window / CRP model

Abstract:
- current row
- top visible row
- page size
- top/bottom/more indicators

Then reuse it across:
- task list
- handoffs
- sprintboard
- git dashboard
- health lists where applicable

**Primary files**
- `internal/cli/tui3270.go`
- `internal/cli/tui3270_screen_tasklist.go`
- `internal/cli/tui3270_screen_handoff.go`
- `internal/cli/tui3270_screen_sprintboard.go`

### Theme 4: First-class command bar services

Evolve `handleCommand(...)` into a command subsystem with:
- normalized command scope
- clearer per-screen command support
- future command retrieve/history support
- future completion/help support

**Primary files**
- `internal/cli/tui3270_helpers.go`
- `internal/cli/tui3270_menu.go`
- `internal/cli/tui3270_screen_tasklist.go`

### Theme 5: Scrollable detail areas

Move detail and long-text screens toward explicit scroll-region behavior instead of static truncation.

**Primary files**
- `internal/cli/tui3270_screen_taskdetail.go`
- `internal/cli/tui3270_screen_scorecard.go`
- `internal/cli/tui3270_screen_health.go`
- `internal/cli/tui3270_screen_gitdashboard.go`

## Suggested phased plan

### Phase 1: Shared 3270 chrome
- add a shared screen-chrome helper layer
- standardize title/command/status/PF rows
- normalize OIA/status content

### Phase 2: PF/action metadata
- define screen-local PF maps via shared metadata
- generate PF help rows from metadata
- align handlers with displayed keys

### Phase 3: Reusable CRP/list window model
- extract cursor/listOffset/page calculations
- reuse in task list and at least one other list-like screen

### Phase 4: Command bar consolidation
- unify command parsing/dispatch behavior
- make command availability screen-aware
- prepare for retrieve/history/completion

### Phase 5: Scrollable detail regions
- implement explicit scroll state for long detail screens
- add consistent indicators and PF support

## Recommendation

Do **not** attempt split-screen/app-stacking parity with lspf first.

The highest-value next step is to standardize the **3270 chrome + PF/action metadata + CRP/list window model**. Those changes fit exarp-go's current architecture and improve consistency without requiring a large framework rewrite.
