# TUI Components Research - Bubble Tea v2

**Date:** 2026-03-12
**Purpose:** Evaluate native Bubble Tea v2 components vs third-party alternatives

## Executive Summary

Bubble Tea v2 (`charm.land/bubbles/v2`) now provides native components that replace previously researched third-party alternatives.

| Feature | Third-Party (Previously Researched) | Native v2 | Status |
|---------|-----------------------------------|------------|--------|
| Table with sorting | `evertras/bubble-table` | `bubbles/v2/table` | ✅ Native available |
| Text input | Built-in | `bubbles/v2/textinput` | ✅ Enhanced in v2 |
| Lists with filtering | Built-in | `bubbles/v2/list` | ✅ Enhanced in v2 |
| Spinner/loader | Built-in | `bubbles/v2/spinner` | ✅ Enhanced in v2 |
| Progress bars | Built-in | `bubbles/v2/progress` | ✅ Enhanced in v2 |
| Viewport/scrolling | Built-in | `bubbles/v2/viewport` | ✅ Enhanced in v2 |
| Terminal charts | `ntcharts` | None | ❌ Still needs third-party |

## Components Analysis

### 1. Table (`bubbles/v2/table`)

Native in v2 with:
- Sortable columns
- Keyboard navigation
- Custom styling via Lip Gloss
- Selection support

**Import:** `charm.land/bubbles/v2/table`

### 2. List (`bubbles/v2/list`)

Enhanced in v2:
- Fuzzy filtering built-in
- Custom item delegates
- Status bar support
- Pagination

**Import:** `charm.land/bubbles/v2/list`

### 3. Text Input (`bubbles/v2/textinput`)

Enhanced in v2:
- Real cursor support (opt-in)
- Password masking modes
- Suggestions/autocomplete
- Validation support

**Import:** `charm.land/bubbles/v2/textinput`

### 4. Spinner (`bubbles/v2/spinner`)

Enhanced in v2:
- Multiple spinner styles
- Custom messages

**Import:** `charm.land/bubbles/v2/spinner`

### 5. Viewport (`bubbles/v2/viewport`)

Enhanced in v2:
- Horizontal scrolling
- Line number gutter support
- Mouse wheel support

**Import:** `charm.land/bubbles/v2/viewport`

## Testing

For TUI testing, use **catwalk** (https://github.com/knz/catwalk):
- Native Bubble Tea testing library
- Unit-test friendly
- Tests model state transitions

## Recommendations

1. **Use native Bubbles v2** for all TUI components
2. **Add catwalk** for TUI testing (see TUI_CLI_TESTING_TOOLS.md)
3. **Skip third-party** table components - native is sufficient
4. **Consider third-party** only for terminal charts (no native option)

## References

- Bubbles v2: https://github.com/charmbracelet/bubbles
- Bubble Tea: https://github.com/charmbracelet/bubbletea
- Lip Gloss: https://github.com/charmbracelet/lipgloss
- Catwalk (testing): https://github.com/knz/catwalk
