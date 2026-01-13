# TUI/CLI Testing Tools Research

**Date:** 2026-01-12  
**Purpose:** Evaluate ttyrec and Go libraries for testing TUI/CLI applications in exarp-go

## Executive Summary

This document evaluates tools and libraries for testing terminal user interfaces (TUI) and command-line interfaces (CLI) in Go applications, specifically for exarp-go's CLI and TUI components.

**Key Findings:**
- **ttyrec**: Useful for recording/replaying terminal sessions, but requires external tool integration
- **catwalk**: Best fit for exarp-go's Bubble Tea TUI (already using Bubble Tea framework)
- **tmux-tui-testing**: Good for complex CLI interactions requiring real TTY
- **phoenix/testing**: Not applicable (exarp-go uses Bubble Tea, not Phoenix)

**Recommendation:** Use **catwalk** for TUI testing (Bubble Tea) and **standard Go testing** with output capture for CLI testing.

---

## Current State Analysis

### exarp-go CLI/TUI Architecture

**CLI Implementation:**
- Location: `internal/cli/cli.go`
- TTY Detection: Uses `golang.org/x/term.IsTerminal()` (line 200-202)
- Modes: Tool execution, interactive mode, task commands, TUI launch
- Testing: Basic CLI tests in Makefile (`test-cli-*` targets)

**TUI Implementation:**
- Location: `internal/cli/tui.go`
- Framework: **Bubble Tea** (`github.com/charmbracelet/bubbletea`)
- Features: Task management interface, config editing, 3270 TUI server
- Testing: No automated TUI tests currently

**Key Code References:**
```199:202:internal/cli/cli.go
// IsTTY checks if stdin is a terminal
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
```

```1:19:internal/cli/tui.go
package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/tools"
	"gopkg.in/yaml.v3"
)
```

---

## Tool Evaluation

### 1. ttyrec

**What it is:**
- Terminal recorder that captures all input/output with timing information
- Records terminal sessions to `.ttyrec` files
- Can replay sessions with `ttyplay`

**Pros:**
- ✅ Captures complete terminal interactions (input + output + timing)
- ✅ Useful for regression testing (record once, replay to verify)
- ✅ Works with any terminal application
- ✅ Preserves timing information for realistic replay

**Cons:**
- ❌ External tool dependency (not a Go library)
- ❌ Requires shell integration (`ttyrec command` wrapper)
- ❌ Binary format, harder to programmatically analyze
- ❌ Cross-platform concerns (installation varies)
- ❌ Not ideal for unit tests (better for integration/E2E)

**Integration Approach:**
```bash
# Record a session
ttyrec -e ./bin/exarp-go tui

# Replay for verification
ttyplay session.ttyrec
```

**Use Case:** Integration/E2E testing, not unit tests

**Recommendation:** ⚠️ **Consider for E2E testing**, not unit tests

---

### 2. catwalk (Bubble Tea Testing)

**What it is:**
- Testing library specifically for Bubble Tea TUI applications
- Simulates user inputs and verifies model states
- Supports testing style changes and keybindings

**Pros:**
- ✅ **Perfect fit** for exarp-go (already using Bubble Tea)
- ✅ Native Go library (no external dependencies)
- ✅ Unit-test friendly (fast, no real terminal needed)
- ✅ Can test model state transitions
- ✅ Can verify rendering output
- ✅ Supports keybinding testing

**Cons:**
- ❌ Only works with Bubble Tea (not a problem for exarp-go)
- ❌ Requires understanding of Bubble Tea model architecture

**Integration Example:**
```go
// Example test using catwalk
package cli_test

import (
    "testing"
    "github.com/knz/catwalk"
    "github.com/davidl71/exarp-go/internal/cli"
)

func TestTUIInitialState(t *testing.T) {
    model := cli.NewModel(server, "")
    
    // Simulate keypress
    updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
    
    // Verify state
    if updated.Cursor() != 1 {
        t.Errorf("Expected cursor at 1, got %d", updated.Cursor())
    }
}
```

**Use Case:** Unit testing Bubble Tea TUI models

**Recommendation:** ✅ **STRONGLY RECOMMENDED** for TUI testing

**Repository:** https://github.com/knz/catwalk

---

### 3. tmux-tui-testing (ttt)

**What it is:**
- Automated testing framework using tmux
- Simulates user interactions in real terminal sessions
- Executes test specifications in tmux panes

**Pros:**
- ✅ Works with real TTY (good for complex CLI interactions)
- ✅ Can test full terminal applications
- ✅ Supports complex interaction sequences
- ✅ Good for integration testing

**Cons:**
- ❌ Requires tmux (external dependency)
- ❌ Slower than unit tests (real terminal sessions)
- ❌ More complex setup
- ❌ May have cross-platform issues

**Integration Example:**
```go
// Example using tmux-tui-testing
import "github.com/SirMoM/tmux-tui-testing"

func TestCLIInteractive(t *testing.T) {
    spec := &ttt.Spec{
        Commands: []ttt.Command{
            {Send: "exarp-go -i"},
            {Wait: "exarp-go>"},
            {Send: "list"},
            {Wait: "Available tools"},
        },
    }
    
    if err := ttt.Run(spec); err != nil {
        t.Fatal(err)
    }
}
```

**Use Case:** Integration testing of complex CLI interactions

**Recommendation:** ⚠️ **Consider for integration tests**, not unit tests

**Repository:** https://pkg.go.dev/github.com/SirMoM/tmux-tui-testing

---

### 4. phoenix/testing

**What it is:**
- Testing package for Phoenix TUI framework
- Provides `NullTerminal` and `MockTerminal` helpers

**Pros:**
- ✅ Good testing utilities for TUI frameworks
- ✅ Mock terminal implementations

**Cons:**
- ❌ **Not applicable** - exarp-go uses Bubble Tea, not Phoenix
- ❌ Framework-specific (Phoenix only)

**Recommendation:** ❌ **NOT APPLICABLE** (wrong framework)

---

## Standard Go Testing Approaches

### CLI Testing (Current Approach)

**Current Implementation:**
```makefile
test-cli-list: build ## Test CLI list tools functionality
	@echo "$(BLUE)Testing CLI: list tools...$(NC)"
	@$(BINARY_PATH) -list > /dev/null 2>&1 && \
	 echo "$(GREEN)✅ CLI list command works$(NC)" || \
	 (echo "$(RED)❌ CLI list command failed$(NC)" && exit 1)
```

**Standard Go Testing Pattern:**
```go
func TestCLIListTools(t *testing.T) {
    // Override os.Args
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"exarp-go", "-list"}
    
    // Capture stdout
    oldStdout := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    
    // Run CLI
    err := cli.Run()
    w.Close()
    
    // Read output
    var buf bytes.Buffer
    io.Copy(&buf, r)
    os.Stdout = oldStdout
    
    // Verify
    if err != nil {
        t.Fatalf("CLI failed: %v", err)
    }
    if !strings.Contains(buf.String(), "Available tools") {
        t.Error("Expected 'Available tools' in output")
    }
}
```

**Pros:**
- ✅ No external dependencies
- ✅ Fast execution
- ✅ Standard Go testing patterns
- ✅ Easy to integrate with existing test infrastructure

**Recommendation:** ✅ **Continue using** for CLI unit tests

---

## Recommendations

### For TUI Testing

**Primary Recommendation: catwalk**
- Best fit for Bubble Tea framework
- Unit-test friendly
- No external dependencies
- Can test model state and rendering

**Implementation Plan:**
1. Add `catwalk` dependency: `go get github.com/knz/catwalk`
2. Create `internal/cli/tui_test.go`
3. Write tests for:
   - Model initialization
   - Key press handling
   - State transitions
   - Rendering output

### For CLI Testing

**Primary Recommendation: Standard Go Testing**
- Continue current approach
- Enhance with more comprehensive test cases
- Add output verification
- Test all CLI modes (tool execution, interactive, task commands)

**Enhancement Plan:**
1. Expand `internal/cli/cli_test.go`
2. Add tests for:
   - All CLI flags and modes
   - Error handling
   - Output formatting
   - TTY detection

### For Integration/E2E Testing

**Optional: ttyrec or tmux-tui-testing**
- Use for full end-to-end testing
- Record real user interactions
- Verify complete workflows
- Not required for unit tests

---

## Implementation Examples

### Example 1: TUI Test with catwalk

```go
package cli_test

import (
    "testing"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/davidl71/exarp-go/internal/cli"
    "github.com/davidl71/exarp-go/internal/framework"
)

func TestTUINavigation(t *testing.T) {
    // Setup
    server := setupMockServer(t)
    model := cli.NewModel(server, "")
    
    // Test initial state
    if model.Cursor() != 0 {
        t.Errorf("Expected cursor at 0, got %d", model.Cursor())
    }
    
    // Simulate down arrow
    updated, cmd := model.Update(tea.KeyMsg{
        Type: tea.KeyDown,
        Runes: []rune{'j'},
    })
    
    if cmd != nil {
        t.Error("Expected no command for navigation")
    }
    
    // Verify cursor moved
    if updated.Cursor() != 1 {
        t.Errorf("Expected cursor at 1, got %d", updated.Cursor())
    }
}
```

### Example 2: CLI Test with Standard Go Testing

```go
package cli_test

import (
    "bytes"
    "os"
    "strings"
    "testing"
    "github.com/davidl71/exarp-go/internal/cli"
)

func TestCLIListTools(t *testing.T) {
    // Save original args
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    
    // Set test args
    os.Args = []string{"exarp-go", "-list"}
    
    // Capture stdout
    oldStdout := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    
    // Run CLI
    err := cli.Run()
    w.Close()
    
    // Read output
    var buf bytes.Buffer
    buf.ReadFrom(r)
    os.Stdout = oldStdout
    
    // Verify
    if err != nil {
        t.Fatalf("CLI failed: %v", err)
    }
    
    output := buf.String()
    if !strings.Contains(output, "Available tools") {
        t.Error("Expected 'Available tools' in output")
    }
}
```

### Example 3: Integration Test with ttyrec (Optional)

```bash
#!/bin/bash
# test_tui_integration.sh

# Record TUI session
ttyrec -e ./bin/exarp-go tui <<EOF
# Simulate user interactions
q  # Quit
EOF

# Verify recording exists
if [ ! -f "ttyrecord" ]; then
    echo "Failed to record session"
    exit 1
fi

# Replay and verify
ttyplay ttyrecord
```

---

## Comparison Matrix

| Tool | Type | Framework | Unit Test | Integration | Complexity | Recommendation |
|------|------|-----------|-----------|-------------|------------|----------------|
| **catwalk** | Go Library | Bubble Tea | ✅ Excellent | ⚠️ Limited | Low | ✅ **Use for TUI** |
| **Standard Go** | Built-in | Any | ✅ Excellent | ⚠️ Limited | Low | ✅ **Use for CLI** |
| **tmux-tui-testing** | Go Library | Any | ❌ No | ✅ Excellent | Medium | ⚠️ **Optional E2E** |
| **ttyrec** | External Tool | Any | ❌ No | ✅ Excellent | Medium | ⚠️ **Optional E2E** |
| **phoenix/testing** | Go Library | Phoenix | ✅ Good | ⚠️ Limited | Low | ❌ **Not applicable** |

---

## Next Steps

1. **Immediate (TUI Testing):**
   - Add `catwalk` dependency
   - Create `internal/cli/tui_test.go`
   - Write initial TUI model tests

2. **Short-term (CLI Testing):**
   - Expand `internal/cli/cli_test.go`
   - Add comprehensive CLI mode tests
   - Improve output verification

3. **Long-term (Integration Testing):**
   - Consider ttyrec or tmux-tui-testing for E2E tests
   - Add CI integration for integration tests
   - Document testing patterns

---

## References

- **ttyrec**: https://github.com/mjording/ttyrec
- **catwalk**: https://github.com/knz/catwalk
- **tmux-tui-testing**: https://pkg.go.dev/github.com/SirMoM/tmux-tui-testing
- **phoenix/testing**: https://pkg.go.dev/github.com/phoenix-tui/phoenix/testing
- **Bubble Tea**: https://github.com/charmbracelet/bubbletea
- **Go Testing Guide**: https://pkg.go.dev/testing

---

## Conclusion

For exarp-go's TUI/CLI testing needs:

1. **TUI Testing**: Use **catwalk** (perfect fit for Bubble Tea)
2. **CLI Testing**: Use **standard Go testing** (current approach, enhance it)
3. **Integration Testing**: Consider **ttyrec** or **tmux-tui-testing** (optional, for E2E)

The combination of catwalk for TUI and standard Go testing for CLI provides comprehensive coverage without external tool dependencies for unit tests.
