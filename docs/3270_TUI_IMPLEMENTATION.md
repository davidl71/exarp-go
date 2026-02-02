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

## Future Enhancements

Based on 3270BBS patterns, potential improvements:

1. **Task Editing**: Full CRUD operations for tasks
2. **Batch Operations**: Select multiple tasks for batch updates
3. **Search/Filter**: Search tasks by content, tags, or other criteria
4. **Auto-refresh**: Periodic task list updates (like bubbletea TUI)
5. **Color Support**: Use 3270 color attributes for better visual distinction
6. **Graphic Escape**: Support CP310 for enhanced box drawing (if client supports)
7. **Multi-session**: Support multiple virtual sessions (F23/F24 pattern from 3270BBS)
8. **Help System**: Context-sensitive help (PF1)

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
