# ISPF Patterns Research for exarp-go 3270 TUI

## Overview

Research into ISPF (Interactive System Productivity Facility) patterns and moshix's implementations to enhance the 3270 TUI for exarp-go.

## Moshix's Mainframe Projects

### 1. **3270BBS** (Bulletin Board System)
- **Repository**: https://github.com/moshix/3270BBS
- **Language**: Go (60,000+ lines)
- **Features**:
  - Built-in editor for Topics, Posts, Notes, marketplace items, and Messages
  - Spell checker with mainframe terminology
  - Color tags support (`<<blue>>`, `<<red>>`, etc.)
  - Function key support (F2=Spell checker, F4=Delete line, etc.)
  - Editor commands: SAVE, CANCEL
  - Uses go3270 library by racingmars

### 2. **Proxy3270**
- **Language**: Go
- **Purpose**: 3270 terminal handling system
- **Potential patterns**: Terminal emulation, connection handling

### 3. **SPFPC**
- **Language**: Not Go (MS-DOS)
- **Purpose**: ISPF-like editor environment for MS-DOS
- **Note**: Not directly applicable but shows ISPF concepts

### 4. **MVS**
- **Repository**: https://github.com/moshix/mvs
- **Purpose**: Collection of utilities for MVS, z/OS, VM/SP, and z/VM
- **Includes**: Tools for ISPF panel development

## ISPF Concepts for exarp-go

### Editor Features (from 3270BBS)

**Function Keys:**
- F2 = Spell checker
- F4 = Delete current line
- F5 = Insert line below
- F6 = Insert line above
- F7 = Scroll up
- F8 = Scroll down
- F10 = Center current line
- F11 = Make centered box
- F13 = Centered box until next empty row
- F14 = Abandon Edit unsaved
- F15 = Save and exit editor
- SAVE = Save content
- CANCEL = Exit editor unsaved

**Editor Capabilities:**
- Line editing (insert, delete, modify)
- Multi-line content editing
- Spell checking
- Text formatting (centering, boxes)
- Scroll support for long content

### Panel Design Patterns

**Standard ISPF Panel Layout:**
```
┌─────────────────────────────────────────────────────────────┐
│ Panel Title (Row 1, centered, intense)                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Content area (Rows 2-21)                                    │
│                                                             │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│ Command line (Row 22)                                       │
│ PF keys help (Row 23)                                       │
└─────────────────────────────────────────────────────────────┘
```

**Key Elements:**
- Title bar (Row 1, intense/highlighted)
- Content area (Rows 2-21, scrollable)
- Command line (Row 22, for user input)
- PF key help (Row 23, shows available function keys)

### Navigation Patterns

**Menu Navigation:**
- Option numbers (1, 2, 3, etc.) for menu selection
- Enter key to select
- PF3 to go back/exit
- PF12 to cancel

**List Navigation:**
- PF7/PF8 for scroll up/down
- Enter to select item
- Cursor positioning with arrow keys (if supported)

**Detail View:**
- Display-only fields (read-only)
- Editable fields (Write: true)
- PF3 to return to list

## Potential Enhancements for exarp-go 3270 TUI

### 1. Task Editor

Add ISPF-like editor for task editing:

```go
// Task editor transaction
func (state *tui3270State) taskEditorTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
    // Load task content into editable fields
    screen := go3270.Screen{
        {Row: 1, Col: 2, Content: "EDIT TASK", Intense: true},
        {Row: 3, Col: 2, Content: "Task ID:", Intense: true},
        {Row: 3, Col: 12, Content: task.ID}, // Read-only
        {Row: 4, Col: 2, Content: "Status:", Intense: true},
        {Row: 4, Col: 12, Write: true, Name: "status", Content: task.Status},
        {Row: 5, Col: 2, Content: "Priority:", Intense: true},
        {Row: 5, Col: 12, Write: true, Name: "priority", Content: task.Priority},
        {Row: 7, Col: 2, Content: "Description:", Intense: true},
        // Multi-line editor for description
        {Row: 8, Col: 2, Write: true, Name: "desc_line1", Content: getLine(task.LongDescription, 0)},
        {Row: 9, Col: 2, Write: true, Name: "desc_line2", Content: getLine(task.LongDescription, 1)},
        // ... more lines
        {Row: 22, Col: 2, Write: true, Name: "command", Content: ""}, // Command line
        {Row: 23, Col: 2, Content: "PF3=Exit  PF15=Save  Enter=Execute Command"},
    }
    
    // Handle editor commands
    // SAVE, CANCEL, line editing commands
}
```

### 2. Enhanced List View

Add ISPF-style list with more features:

```go
// ISPF-style list with:
// - Scroll indicators (TOP, MORE, BOTTOM)
// - Line numbers
// - Selection markers
// - Filter/search capability
screen := go3270.Screen{
    {Row: 1, Col: 2, Content: "TASK LIST", Intense: true},
    {Row: 1, Col: 70, Content: "SCROLL: CSR", Intense: true}, // Scroll indicator
    // Task lines with line numbers
    {Row: 3, Col: 2, Content: " 1 > T-123 [Todo] [HIGH] Task description"},
    {Row: 4, Col: 2, Content: " 2   T-124 [Done] [MED]  Another task"},
    // ...
    {Row: 22, Col: 2, Write: true, Name: "command", Content: ""},
    {Row: 23, Col: 2, Content: "PF1=Help  PF3=Exit  PF7=Up  PF8=Down  Enter=Select"},
}
```

### 3. Command Line Interface

Add ISPF-style command line for advanced operations:

```go
// Command line at row 22
{Row: 22, Col: 2, Content: "Command ===>", Intense: true},
{Row: 22, Col: 15, Write: true, Name: "command", Content: ""},

// Support commands like:
// - FILTER status=Todo
// - SORT priority
// - FIND keyword
// - EDIT T-123
// - DELETE T-123
```

### 4. Panel Stacking

ISPF uses a panel stack for navigation history:

```go
type panelStack struct {
    panels []string // Transaction names
}

func (state *tui3270State) pushPanel(panelName string) {
    state.panelStack = append(state.panelStack, panelName)
}

func (state *tui3270State) popPanel() string {
    if len(state.panelStack) == 0 {
        return "main"
    }
    panel := state.panelStack[len(state.panelStack)-1]
    state.panelStack = state.panelStack[:len(state.panelStack)-1]
    return panel
}
```

### 5. Split Screen / Multi-Panel

ISPF supports split screens for side-by-side views:

```go
// Left panel: Task list
// Right panel: Task details
// Use alternate screen or screen regions
```

## Implementation Recommendations

### Phase 1: Basic Editor
1. Add task editor transaction
2. Support basic field editing (status, priority)
3. Multi-line description editor
4. Save/Cancel commands

### Phase 2: Enhanced Navigation
1. Panel stack for navigation history
2. Command line interface
3. Scroll indicators
4. Line numbers in lists

### Phase 3: Advanced Features
1. Spell checker integration
2. Text formatting (centering, boxes)
3. Search/filter commands
4. Batch operations

## References

- **3270BBS**: https://github.com/moshix/3270BBS
- **go3270 Library**: https://github.com/racingmars/go3270
- **ISPF Documentation**: IBM z/OS ISPF User's Guide
- **3270BBS Editor Features**: See 3270BBS README for editor function keys

## Notes

- 3270BBS uses code page 1047 as standard
- Editor supports graphic escape (CP310) for enhanced box drawing
- Function keys F1-F24 are standard in 3270 terminals
- Command line pattern is standard ISPF convention
- Panel stacking enables complex navigation flows
