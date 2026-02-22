# 3270 TUI Implementation for exarp-go

## Overview

This document describes the 3270 TUI (Terminal User Interface) implementation for exarp-go, inspired by the [go3270](https://github.com/racingmars/go3270) library and implementation patterns from [3270BBS](https://github.com/moshix/3270BBS).

## Features

The 3270 TUI provides a classic mainframe-style interface for task management:

- **Main Menu**: Navigation hub with options for Task List, Configuration, and Exit
- **Task List**: Browse tasks with cursor navigation (PF7/PF8 for up/down)
- **Task Details**: View full task information including description, tags, dependencies
- **Configuration**: Edit configuration settings (basic implementation)

## Architecture

### Key Components

1. **`tui3270.go`**: Main 3270 TUI implementation
   - `RunTUI3270()`: Starts the tn3270 server
   - `handle3270Connection()`: Handles individual client connections
   - Transaction handlers for each screen

2. **Transaction-Based Navigation**: Uses `go3270.RunTransactions()` pattern
   - `mainMenuTransaction`: Main menu screen
   - `taskListTransaction`: Task list with cursor navigation
   - `taskDetailTransaction`: Task detail view
   - `configTransaction`: Configuration editor

### State Management

The `tui3270State` struct maintains session state:
- Current task list and cursor position
- Selected task for detail view
- Project information
- Device info (code page, terminal capabilities)

## Usage

### Starting the Server

```bash
# Start on default port 3270
exarp-go tui3270

# Start on custom port
exarp-go tui3270 "" 2300

# Filter by status
exarp-go tui3270 "In Progress"

# Filter by status on custom port
exarp-go tui3270 "Todo" 2300
```

### Connecting with a 3270 Client

```bash
# Using x3270 (most common)
x3270 localhost:3270

# Using c3270 (command-line)
c3270 localhost:3270

# Using s3270 (scriptable)
s3270 localhost:3270
```

## Key Bindings

### Main Menu
- **1**: Navigate to Task List
- **2**: Navigate to Configuration
- **3** or **Enter** (empty): Exit
- **PF3**: Exit

### Task List
- **PF7**: Move cursor up
- **PF8**: Move cursor down
- **Enter** or **PF1**: View task details
- **PF3**: Back to main menu
- **PF12**: Cancel/Back

### Task Details
- **PF3**: Back to task list
- **PF12**: Cancel/Back

### Configuration
- **PF3**: Back to main menu

## Implementation Patterns

### Inspired by 3270BBS

1. **Transaction-Based Flow**: Uses `RunTransactions()` for clean screen-to-screen navigation
2. **Code Page Support**: Automatically detects and uses client code page (defaults to CP1047)
3. **Screen Layout**: 24x80 standard terminal size with proper field positioning
4. **Attention Keys**: PF keys for navigation and actions

### Code Page Handling

The implementation uses `devInfo.Codepage()` to ensure proper UTF-8 to EBCDIC translation:
- Automatically detects client code page during telnet negotiation
- Falls back to CP1047 if code page not detected
- Supports all single-byte EBCDIC code pages supported by x3270

## Registering as a 3270BBS Door

[3270BBS](https://github.com/moshix/3270BBS) supports remote host proxying — users connect to 3270BBS, select a remote host from a "Doors" menu, and get proxied to another TN3270 server. exarp-go can be registered as a door since it uses the same TN3270 protocol.

### Configuration

Add the following to your 3270BBS `tsu.cnf` configuration:

```ini
remote4=ExarpGo
remote4_description="exarp-go task management"
remote4_addr=localhost
remote4_port=3270
```

Adjust the `remote4_addr` to the hostname or IP where exarp-go is running. The port should match the `--port` flag passed to `exarp-go tui3270` (default: 3270).

### Requirements

- exarp-go must be running with `exarp-go tui3270` (or `exarp-go tui3270 --daemon`)
- 3270BBS and exarp-go can run on the same host (use different ports) or separate hosts
- Users return from the exarp-go door to 3270BBS by pressing **PA3**

### Community

The 3270BBS community is active at [Forum3270](https://www.moshix.tech). Publishing exarp-go as a door gives the 3270 TUI exposure beyond its current audience.

## Implemented Features

The following features have been implemented:

1. **Color-coded status/priority**: Task status and priority displayed with distinct 3270 colors
2. **Blue+Intense title bars**: All screen titles use mainframe-standard Blue+Intense headers
3. **Turquoise PF key lines**: Instructional PF key help lines in Turquoise
4. **ReverseVideo cursor row**: Selected row highlighted with reverse video
5. **PF9 status filter cycling**: Cycle through Todo/In Progress/Review/Done/All with PF9
6. **HandleScreen validation**: Task editor uses `HandleScreenAlt` with field validation rules
7. **Cursor-position row selection**: Click-to-select using Enter key + cursor position
8. **ISPF-style line commands**: Type S/E/D/I next to task rows for Select/Edit/Done/In Progress
9. **Green structural borders**: ASCII box-drawing borders around menu panels
10. **Terminal size adaptation**: Dynamic layout based on terminal dimensions (Model 2/3/4/5)
11. **Scorecard color coding**: Scores color-coded Green/Yellow/Red by threshold
12. **NoClear loading indicators**: Non-destructive "Loading..." overlays during data fetch
13. **Health/SDSF panel**: System health dashboard with git status, task counts, and server info
14. **Help system**: Context-sensitive help screen (PF1) with command and line command reference

## Future Enhancements

Potential improvements:

1. **Batch Operations**: Select multiple tasks for batch updates (multi-row line commands)
2. **Graphic Escape**: Support CP310 for enhanced box drawing (if client supports)
3. **Multi-session**: Support multiple virtual sessions (F23/F24 pattern from 3270BBS)
4. **Standalone mini-apps**: Git dashboard, log viewer, sprint board as separate `go3270.Tx` chains

## Dependencies

- `github.com/racingmars/go3270`: Core 3270 server library
- Standard exarp-go packages: `database`, `config`, `framework`, `tools`

## Testing

To test the 3270 TUI:

1. **Install a 3270 client**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install x3270
   
   # macOS
   brew install x3270
   ```

2. **Start the server**:
   ```bash
   exarp-go tui3270
   ```

3. **Connect with client**:
   ```bash
   x3270 localhost:3270
   ```

## Related projects

- **[3270BBS](https://github.com/moshix/3270BBS)** — Full BBS for IBM 3270 terminals (real and emulated). Same codebase runs **Forum3270**. Built with [go3270](https://github.com/racingmars/go3270). Clone: `https://github.com/moshix/3270BBS.git`

## References

- [go3270 Library](https://github.com/racingmars/go3270) - Core 3270 server library
- [3270BBS](https://github.com/moshix/3270BBS) - Full-featured BBS implementation using go3270
- [3270 Data Stream Reference](https://www.ibm.com/docs/en/zos/2.1.0?topic=streams-3270-data-stream) - IBM documentation
- [TN3270 RFC 1576](https://tools.ietf.org/html/rfc1576) - TN3270 Current Practices

## Notes

- The server listens on TCP port 3270 by default (configurable)
- Each client connection is handled in a separate goroutine
- Database initialization happens per-connection (could be optimized)
- Code page detection happens during telnet negotiation
- The implementation follows the recommended `RunTransactions()` pattern for larger applications
