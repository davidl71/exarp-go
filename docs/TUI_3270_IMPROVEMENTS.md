# 3270 TUI Visual & Functional Improvements

Research based on [moshix/3270BBS](https://github.com/moshix/3270BBS) (Go, ~60k lines, `racingmars/go3270`), [moshix/minesweeper](https://github.com/moshix/minesweeper), and the [go3270 library](https://github.com/racingmars/go3270).

## Current State

The 3270 TUI (`internal/cli/tui3270*.go`) is functional but monochrome — all fields use default color with only `Intense: true` for differentiation. No box-drawing, no color coding, no terminal size adaptation.

**Screens:** Main menu, Task list, Task detail, Task editor, Scorecard, Handoffs, Help, Child agent menu/result.

**Files:** `tui3270.go` (state/server), `tui3270_transactions.go` (screens), `tui3270_menu.go` (menus), `tui3270_helpers.go` (commands/help).

---

## Part 1: Visual Improvements

### go3270 Color Palette (Available but Unused)

```go
go3270.Blue      // 0xf1 — titles, headers
go3270.Red       // 0xf2 — errors, high priority, danger
go3270.Pink      // 0xf3 — warnings, medium priority
go3270.Green     // 0xf4 — success, Done status, borders
go3270.Turquoise // 0xf5 — help text, PF key lines, metadata
go3270.Yellow    // 0xf6 — In Progress status, attention
go3270.White     // 0xf7 — neutral metadata, timestamps
```

### go3270 Highlighting (Available but Unused)

```go
go3270.ReverseVideo // 0xf2 — selected/cursor row
go3270.Underscore   // 0xf4 — section headers, separators
go3270.Blink        // 0xf1 — critical alerts (use sparingly)
```

### V1: Color-coded status and priority

Apply color to the Status and Priority columns in the task list.

| Status      | Color     | Rationale |
|-------------|-----------|-----------|
| Done        | Green     | Success/complete (universal) |
| In Progress | Yellow    | Active work, attention needed |
| Todo        | Turquoise | Queued, informational |
| Review      | Pink      | Needs human action |

| Priority | Color  | Rationale |
|----------|--------|-----------|
| high     | Red    | Urgent, critical |
| medium   | Yellow | Notable |
| low      | Green  | Low urgency |

**Effort:** Low — add `Color: go3270.Green` etc. to existing fields in `taskListTransaction`.

### V2: Blue+Intense title bars

All screen titles (row 1) use `Color: go3270.Blue, Intense: true` — the mainframe convention for panel headers.

**Current:** `{Row: 1, Col: 2, Content: "TASK LIST (...)", Intense: true}`
**After:** `{Row: 1, Col: 2, Content: "TASK LIST (...)", Intense: true, Color: go3270.Blue}`

**Effort:** Low — one-line change per screen (~8 screens).

### V3: Turquoise PF key lines

Bottom help/PF key lines use `Color: go3270.Turquoise` — the mainframe convention for instructional text.

**Effort:** Low — one-line change per screen.

### V4: ReverseVideo cursor row

The current cursor marker (`> `) in the task list should use `Highlighting: go3270.ReverseVideo` to make the selected row visually prominent (standard ISPF behavior).

**Effort:** Low — add highlighting to the cursor row's fields.

### V5: Green structural borders

Add ASCII box-drawing borders around the main menu options and section dividers. The minesweeper game demonstrates `+---+` grids in Green; apply the same pattern to menu panels and section separators.

**Example main menu:**
```
+----------------------------------+
| 1. Task List                     |
| 2. Configuration                 |
| 3. Scorecard                     |
| 4. Session handoffs              |
| 5. Exit                          |
| 6. Run in child agent            |
+----------------------------------+
```

**Effort:** Medium — construct border strings, add as Green fields.

### V6: AttributeOnly inline coloring

Use `AttributeOnly: true` to change color mid-line without starting a new field. This allows coloring the task ID differently from the content within a single row.

**Effort:** Medium — requires restructuring field layout in task list rows.

### V7: Terminal size adaptation

Currently `maxVisible` is hardcoded to 18 rows. Use `devInfo.AltDimensions()` (already available in state) to dynamically calculate visible rows, supporting Model 3 (32×80), Model 4 (43×80), and Model 5 (62×160).

```go
rows, cols := state.devInfo.AltDimensions()
maxVisible := rows - 6 // header(3) + status(1) + pfkeys(1) + margin(1)
```

**Effort:** Medium — affects task list, scorecard, handoff, and help screens.

### V8: Centered menu layouts

Center the main menu and child agent menu options horizontally, like 3270BBS and Minesweeper do. Currently everything is pinned to Col: 2 or Col: 4.

**Effort:** Low — calculate `(cols - textLen) / 2` for menu items.

---

## Part 2: Functional Improvements

### F1: Inline status shortcuts from task list (port from Bubbletea)

The Bubbletea TUI has `d`=Done, `i`=In Progress, `t`=Todo, `r`=Review shortcuts (in `tui_update_handlers.go`). The 3270 TUI has no equivalent — status changes require entering the editor (PF2) and manually typing the status.

**Approach:** Map PF keys to status changes from the task list:
- **PF4** = Mark Done
- **PF5** = Mark In Progress
- **PF6** = Mark Todo
- **PF10** = Mark Review

Update PF key help line accordingly. Uses `updateTaskFieldsViaMCP` (already available).

**Effort:** Medium — add PF key handlers in `taskListTransaction`, update help text.

### F2: ISPF-style line commands

3270 applications traditionally use **line commands** — single-letter prefixes typed next to each row. The task list has a writable `S` column but doesn't process it. Add:

| Command | Action |
|---------|--------|
| S       | Select (view detail) |
| E       | Edit task |
| D       | Mark Done |
| I       | Mark In Progress |
| /       | Show action menu for this task |

**Approach:** Make the `S` column a writable field per row (named `s_0`, `s_1`, ...). On Enter, scan all `s_*` fields for non-empty values and dispatch the matching action.

**Effort:** Medium-High — restructure task list row fields, add dispatch logic.

### F3: Status filter cycling (PF9 or Tab)

Currently switching between Todo/In Progress/Done/All requires typing commands (`TASKS`, `FIND`, etc.). Add a PF key (e.g. **PF9**) that cycles through status filters:

`Todo → In Progress → Review → Done → All → Todo → ...`

Display the current filter prominently in the title bar.

**Effort:** Low — add PF9 handler, cycle `state.status`, reload tasks.

### F4: Task creation from 3270 TUI

*Already exists as task T-1771456780882352000.*

Add a "NEW" command or PF key (e.g. **PF11**) that opens a task creation screen with fields for: Content (title), Priority (dropdown/input), Tags, Description. Uses `task_workflow` create action via MCP.

**Effort:** Medium — new transaction screen, MCP call.

### F5: Health/activity panel (SDSF-style)

Inspired by 3270BBS's SDSF panel. Show a live-ish dashboard with:
- Git status (branch, ahead/behind, dirty files)
- Build status (last `make b` result)
- Task statistics (Todo/In Progress/Done counts)
- Memory and disk usage
- Last scorecard score

Access via menu option 7 or command `SDSF` / `HEALTH`.

**Effort:** Medium-High — new transaction, calls `health` MCP tool.

### F6: Alternate screen size support

The go3270 library fully supports alternate screen sizes (32×80, 43×80, 62×160). Currently the TUI only uses the default 24×80. Pass `AltScreen: state.devInfo` to `ScreenOpts` when `devInfo.AltDimensions()` reports a larger screen.

Benefits: more visible task rows, wider content columns, room for side-by-side panels.

**Effort:** Medium — affects all `ShowScreenOpts` calls, needs dynamic layout.

### F7: Quick task status from detail view

The task detail screen currently only supports PF3=Back and PF2=Edit. Add PF4/PF5/PF6 for status changes directly from the detail view (same as task list).

**Effort:** Low — add PF key handlers to `taskDetailTransaction`.

### F8: Scorecard visual enhancement

The scorecard screen dumps plain text. Parse the scorecard data and render it with color-coded scores:
- Green for passing checks (>=80%)
- Yellow for warnings (50-79%)
- Red for failures (<50%)

**Effort:** Medium — parse scorecard structure, apply colors to individual fields.

---

## Implementation Priority

### Quick wins (Low effort, high visual impact)
1. **V1** — Color-coded status/priority
2. **V2** — Blue+Intense title bars
3. **V3** — Turquoise PF key lines
4. **V4** — ReverseVideo cursor row
5. **F3** — Status filter cycling (PF9)
6. **F7** — Quick status from detail view

### Medium effort, strong functional value
7. **F1** — Inline status shortcuts (PF4-6/10)
8. **V7** — Terminal size adaptation
9. **V5** — Green structural borders
10. **F4** — Task creation (T-1771456780882352000)
11. **F8** — Scorecard color coding

### Larger features
12. **F2** — ISPF-style line commands
13. **F5** — Health/SDSF panel
14. **V6** — AttributeOnly inline coloring
15. **F6** — Alternate screen size support

---

## Part 3: Integration & Reusable go3270 Functions

### 3270BBS as a Door Host

3270BBS supports **remote host proxying** — users connect to 3270BBS, select a remote host from a menu, and get proxied to another TN3270 server. Return via PA3.

**exarp-go as a 3270BBS door:**
```ini
# In 3270BBS tsu.cnf — add exarp-go as a remote host
remote4=ExarpGo
remote4_description="exarp-go task management"
remote4_addr=localhost
remote4_port=3270
```

Any user of 3270BBS can then access exarp-go's TUI from the 3270BBS "Doors" menu. This works because both use the same TN3270 protocol — exarp-go's `RunTUI3270` already listens on a TCP port and handles `NegotiateTelnet`.

**3270BBS community:** Active user base at [Forum3270](https://www.moshix.tech). Publishing exarp-go as a door could give the 3270 TUI exposure beyond its current audience.

### go3270 Functions Available but Not Used

Your 3270 TUI uses only 4 of the go3270 API surface. Here's what's available:

| Function/Feature | Used? | What it does | How to adopt |
|-----------------|-------|--------------|--------------|
| `ShowScreenOpts` | **Yes** | Low-level screen send+receive | Current approach |
| `RunTransactions` | **Yes** | Transaction-chain loop | Current approach |
| `NegotiateTelnet` | **Yes** | Connection setup | Current approach |
| `DevInfo.Codepage()` | **Yes** | Code page for encoding | Current approach |
| `HandleScreen` / `HandleScreenAlt` | No | **Higher-level loop with validation** — redisplays screen with error messages until all `Rules` pass, handles accept/exit keys automatically | Replace manual PF key dispatch in editor screens; eliminates boilerplate validation loops |
| `Rules` / `FieldRules` | No | **Field validation** — `MustChange`, `Validator` func, `Reset`, `ErrorText` | Use in task editor (status must be valid, priority must be valid) and task creation forms |
| `NonBlank` validator | No | Built-in "field not empty" check | Apply to required fields in editor/creation screens |
| `IsInteger` validator | No | Built-in integer check | Apply to numeric input fields (e.g. line numbers, recommendation #) |
| `Field.Color` | No | 7 colors (Blue/Red/Pink/Green/Turquoise/Yellow/White) | See Part 1 visual improvements |
| `Field.Highlighting` | No | ReverseVideo, Underscore, Blink | See V4 (cursor row) |
| `Field.AttributeOnly` | No | Change color/highlight mid-line without new field | See V6 (inline coloring) |
| `Field.NumericOnly` | No | Client-side numeric-only input restriction | Apply to score selection, line number inputs |
| `ScreenOpts.AltScreen` | No | **Alternate screen size** (32×80, 43×80, 62×160) | See V7/F6 (terminal size adaptation) |
| `ScreenOpts.NoClear` | No | **Overlay updates** — update parts of screen without clearing | Status bar updates, progress indicators, live refresh |
| `ScreenOpts.PostSendCallback` | No | **Post-send callback** — run function after screen sent, before blocking for response | Start background work (e.g. load data) while user reads the screen |
| `DevInfo.AltDimensions()` | No | Get actual terminal dimensions | Use for dynamic layout calculations |
| `DevInfo.SupportsGE()` | No | Check for code page 310 graphic escape support | Enable box-drawing characters (─│┌┐└┘├┤┬┴┼) when supported |
| `AIDtoString` | No | Human-readable AID key names | Use in error/debug messages: "PF3: unknown key" |

### Key Adoption Opportunities

#### 1. `HandleScreen` with validation (biggest win)

Your task editor (`taskEditorTransaction`) manually checks each response, has no validation, and doesn't redisplay with errors. `HandleScreen` does this automatically:

```go
// Current approach (manual, no validation):
response, err := go3270.ShowScreenOpts(screen, initialValues, conn, opts)
// ... manual PF key checking ...
// ... manual field reading ...
// ... no validation, user can type "xyz" as status ...

// With HandleScreen (automatic validation loop):
rules := go3270.Rules{
    "status": {
        MustChange: false,
        Validator: func(s string) bool {
            s = strings.TrimSpace(s)
            return s == "Todo" || s == "In Progress" || s == "Review" || s == "Done"
        },
    },
    "priority": {
        Validator: func(s string) bool {
            s = strings.TrimSpace(strings.ToLower(s))
            return s == "high" || s == "medium" || s == "low" || s == ""
        },
    },
}
response, err := go3270.HandleScreenAlt(
    screen, rules, values,
    []go3270.AID{go3270.AIDEnter, go3270.AIDPF15},  // accept keys
    []go3270.AID{go3270.AIDPF3},                      // exit keys
    "error_msg",                                       // error display field
    1, 12,                                             // cursor position
    conn, state.devInfo, state.devInfo.Codepage(),
)
// HandleScreen loops automatically until valid or user presses PF3
```

#### 2. `NoClear` for overlay/partial updates

The `ScreenOpts.NoClear` flag sends new field data without clearing the screen. This enables:
- **Status bar updates** — refresh just the bottom status line while task list stays
- **Progress indicators** — show "Loading..." then overlay with results
- **Live scorecard** — update scores in-place as checks complete

```go
// Overlay a status message without redrawing the whole screen
statusScreen := go3270.Screen{
    {Row: 22, Col: 2, Content: "Loading scorecard...", Color: go3270.Yellow},
}
go3270.ShowScreenOpts(statusScreen, nil, conn, go3270.ScreenOpts{
    NoClear:    true,
    NoResponse: true,
    Codepage:   devInfo.Codepage(),
})
// Screen stays, only row 22 updates — then send full screen when data is ready
```

#### 3. `PostSendCallback` for background loading

Start data loading *after* the screen is sent but *before* blocking for user input:

```go
opts := go3270.ScreenOpts{
    Codepage: devInfo.Codepage(),
    PostSendCallback: func(data any) error {
        // Pre-load scorecard data while user is reading the screen
        state := data.(*tui3270State)
        state.preloadedScorecard, _ = tools.GenerateGoScorecard(ctx, state.projectRoot, nil)
        return nil
    },
    CallbackData: state,
}
```

#### 4. Cursor position for row selection (Minesweeper pattern)

The Minesweeper game uses `resp.Row` and `resp.Col` to determine which cell the user clicked. Your task list could do the same — instead of PF7/PF8 scrolling, let users place the cursor on a task row and press Enter:

```go
// Convert cursor row to task index (like Minesweeper's coordinate mapping)
if resp.AID == go3270.AIDEnter {
    taskIdx := (resp.Row - headerRows) + state.listOffset
    if taskIdx >= 0 && taskIdx < len(state.tasks) {
        state.selectedTask = state.tasks[taskIdx]
        return state.taskDetailTransaction, state, nil
    }
}
```

This is more natural than PF7/PF8 — users just tab to a row and press Enter.

### Standalone 3270 Apps (Mini-Apps)

Following the Minesweeper pattern, you could build mini-apps that run as standalone TN3270 servers and are also accessible as "doors" from the main exarp-go TUI:

| Mini-App | Purpose | Pattern |
|----------|---------|---------|
| **Git Dashboard** | Interactive git log/diff viewer over 3270 | Like SDSF but for git |
| **Log Viewer** | Tail application logs in a 3270 scrollable view | 3270BBS has one |
| **Sprint Board** | Kanban-style board (columns = status) | Uses wide terminal |
| **CI Monitor** | Watch GitHub Actions workflow status | Poll and overlay with NoClear |

Each would be a separate `go3270.Tx` chain that the main TUI can dispatch to, or a standalone binary for direct connection.

---

## References

- **3270BBS** — https://github.com/moshix/3270BBS (Go, `racingmars/go3270`, binary-only)
  - 60k lines of Go, F23/F24 virtual sessions, built-in editor, SDSF panel, code page 310 box-drawing
- **Minesweeper** — https://github.com/moshix/minesweeper (Go, open source)
  - Clean example of color usage, centered layout, terminal size adaptation, box-drawing grids
- **go3270 library** — https://github.com/racingmars/go3270
  - Field colors (7 colors), highlighting (Blink/ReverseVideo/Underscore), `AttributeOnly`, `AltScreen`, code page support
- **h3270** — https://h3270.sourceforge.net (Java, regex-based layout engine for 3270-to-HTML)
  - Historical reference for 3270 screen transformation patterns
