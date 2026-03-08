# Testing Tool Strategy Decision

**Date**: 2026-03-08  
**Decision Required**: Testing tool language support strategy  
**Status**: ✅ **DECIDED** - Keep Go-only, document clearly

---

## Problem Statement

The `testing` tool currently:
- Has explicit `IsGoProject()` guards
- Only supports Go testing workflows (`go test`)
- Has generic action names (`run`, `coverage`, `validate`) that suggest multi-language support
- Is documented as "Partially compatible" in the compatibility matrix

**Question**: Should we:
- **Option A**: Keep Go-only and document clearly (possibly rename to `go_testing`)
- **Option B**: Add multi-language support with framework detection

---

## Analysis

### Current Implementation

**File**: `internal/tools/testing.go`

**Actions**:
- `run` - Runs `go test` with optional coverage
- `coverage` - Analyzes Go test coverage
- `validate` - Validates Go test structure
- `suggest` - Suggests new Go tests

**Guards**:
```go
if !IsGoProject() {
    return nil, fmt.Errorf("testing run is only supported for Go projects (go.mod)")
}
```

**Language Detection**:
- Go: Checks for `go.mod`
- Python: Not implemented
- JavaScript/TypeScript: Not implemented
- Rust: Not implemented

### Multi-Language Testing Landscape

| Language | Framework(s) | Command | Coverage Tool |
|----------|--------------|---------|---------------|
| **Go** | testing pkg | `go test` | Built-in (`-cover`) |
| **Python** | pytest, unittest | `pytest`, `python -m pytest` | coverage.py |
| **JavaScript** | Jest, Mocha | `npm test`, `jest` | Jest built-in |
| **TypeScript** | Jest, Vitest | `npm test`, `pnpm test` | Jest/Vitest |
| **Rust** | cargo test | `cargo test` | cargo-tarpaulin |
| **Shell** | bats | `bats` | None standard |
| **Ansible** | ansible-test | `ansible-test` | None |

**Complexity**: Each framework has different:
- Discovery mechanisms (where tests live)
- Output formats
- Coverage reporting formats
- Configuration files

### Existing Multi-Language Tools

exarp-go already has successful multi-language tools:

1. **`security` tool** - Multi-language (Go, Python, Rust, Node)
   - Uses language detection
   - Calls appropriate scanner per language
   - **Complexity**: Medium (4 languages, similar output format)

2. **`lint` tool** - Multi-language (Go, Markdown, Shell, YAML, Ansible)
   - Auto-detects language
   - Calls appropriate linter
   - **Complexity**: Medium (5+ linters, different configs)

3. **`task_discovery` tool** - Language-agnostic
   - Scans TODO comments across all languages
   - Works with any text-based source
   - **Complexity**: Low (pattern matching)

---

## Option A: Keep Go-Only (RECOMMENDED)

### Advantages

✅ **Focus**:
- exarp-go is built in Go
- Primary use case is Go projects
- Testing tool is already comprehensive for Go

✅ **Simplicity**:
- No framework detection complexity
- No need to learn pytest, Jest, cargo patterns
- Maintenance burden stays low

✅ **Correctness**:
- Go test output is well-understood
- Coverage analysis is accurate
- Test structure validation is reliable

✅ **Alternatives Exist**:
- Python: Use `pytest` directly or via shell tools
- JavaScript: Use `npm test` via shell tools
- Rust: Use `cargo test` via shell tools
- Generic: Use `automation` tool with test commands

✅ **Clear Documentation**:
- Tool purpose is obvious
- No confusion about capabilities
- Users know what to expect

### Disadvantages

❌ **Misleading Name**:
- "testing" sounds generic
- Could be renamed to `go_testing` for clarity

❌ **Limited Use**:
- Only useful for Go projects
- Other languages need workarounds

### Implementation

**Changes Required**:
1. ✅ Update tool description to say "Go testing" explicitly
2. ✅ Update TOOL_LANGUAGE_COMPATIBILITY_MATRIX.md (already done)
3. ✅ Consider adding hint: "For non-Go projects, use automation tool"
4. ⚠️ Optional: Rename to `go_testing` (breaking change)

**Effort**: **Low** (documentation only)

---

## Option B: Add Multi-Language Support

### Advantages

✅ **Unified Interface**:
- One tool for all languages
- Consistent user experience

✅ **Convenience**:
- No need to learn different test commands
- Auto-detection of test framework

### Disadvantages

❌ **High Complexity**:
- Need to detect: pytest vs unittest, Jest vs Mocha, etc.
- Different output formats per framework
- Different coverage tools and formats
- Configuration file detection (pytest.ini, jest.config.js, etc.)

❌ **Maintenance Burden**:
- Need to keep up with framework changes
- Multiple frameworks per language
- Edge cases and failure modes multiply

❌ **Scope Creep**:
- Python has pytest, unittest, nose2, tox
- JavaScript has Jest, Mocha, Jasmine, AVA, Tape, etc.
- Which ones to support?

❌ **Limited Value**:
- Most projects already have `npm test` or `make test`
- Shell tools can wrap existing commands
- Automation tool can schedule test runs

❌ **Error Prone**:
- Framework detection can fail
- Output parsing is fragile
- Coverage format conversion is complex

### Implementation

**Changes Required**:
1. Add framework detection per language
2. Implement runner for each framework:
   - Python: pytest, unittest
   - JavaScript: Jest, Mocha
   - TypeScript: Jest, Vitest
   - Rust: cargo test
   - Shell: bats
3. Parse different output formats
4. Handle different coverage formats
5. Test across all frameworks
6. Maintain as frameworks evolve

**Effort**: **High** (8-12 hours initial, ongoing maintenance)

---

## Comparison

| Aspect | Option A (Go-only) | Option B (Multi-language) |
|--------|-------------------|---------------------------|
| **Effort** | Low (docs only) | High (8-12 hours) |
| **Maintenance** | Low | High (ongoing) |
| **User Value** | High (for Go) | Medium (alternatives exist) |
| **Correctness** | High (well-tested) | Medium (framework-dependent) |
| **Scope** | Narrow (focused) | Broad (complex) |
| **Breaking Changes** | None (or rename) | None |

---

## Decision: Option A (Go-Only) ✅

**Rationale**:

1. **Focus on Excellence**: exarp-go should excel at Go testing rather than be mediocre at all languages

2. **Alternatives Exist**: Other languages have good testing stories:
   - Use `automation` tool to schedule `npm test`, `pytest`, etc.
   - Use shell tools for custom test commands
   - Use CI/CD integrations

3. **Maintenance Burden**: Multi-language support would require significant ongoing effort

4. **Clear Purpose**: A focused tool is easier to understand and use

5. **Prior Art**: Other successful tools are Go-specific:
   - golangci-lint (linting)
   - govulncheck (security)

6. **Future Flexibility**: Can always add multi-language support later if needed

---

## Implementation Plan

### Immediate Actions (Required)

1. ✅ **Update Tool Description**:
   - Change: "Testing tool" → "Go testing tool"
   - Add hint: "For non-Go projects, use automation tool with custom test commands"

2. ✅ **Update Documentation**:
   - `docs/EXARP_ABILITIES_AUDIT.md`: Mark as "Go-only"
   - `docs/TOOL_LANGUAGE_COMPATIBILITY_MATRIX.md`: Already states "Go-project flows"
   - Add to FAQ: "Why is testing Go-only?"

3. ✅ **Update Error Messages**:
   - Current: "testing run is only supported for Go projects (go.mod)"
   - Add helpful alternative: "For non-Go testing, use the automation tool or shell commands"

### Optional Actions (Future)

4. ⚠️ **Rename Tool** (Breaking Change):
   - `testing` → `go_testing`
   - Would clarify purpose but breaks existing usage
   - **Recommendation**: Not needed if documentation is clear

5. 💡 **Add Testing Guide**:
   - Document how to use `automation` tool for multi-language testing
   - Provide examples for pytest, Jest, cargo test

---

## Testing Multi-Language Projects

### For Python Projects

**Option 1: Use automation tool**
```bash
exarp-go -tool automation -args '{
  "action": "daily",
  "commands": ["pytest", "pytest --cov"]
}'
```

**Option 2: Use shell tools**
```bash
pytest
pytest --cov --cov-report=html
```

### For JavaScript/TypeScript Projects

**Option 1: Use automation tool**
```bash
exarp-go -tool automation -args '{
  "action": "daily",
  "commands": ["npm test", "npm run test:coverage"]
}'
```

**Option 2: Use npm/pnpm directly**
```bash
npm test
pnpm test
```

### For Rust Projects

**Option 1: Use automation tool**
```bash
exarp-go -tool automation -args '{
  "action": "daily",
  "commands": ["cargo test", "cargo tarpaulin"]
}'
```

**Option 2: Use cargo directly**
```bash
cargo test
cargo tarpaulin --out Html
```

---

## FAQ

### Q: Why not support all languages?

**A**: Focus and maintenance. A focused Go testing tool is more valuable than a mediocre multi-language one. Other languages already have excellent testing tools.

### Q: How do I test non-Go projects?

**A**: Use the `automation` tool to schedule your existing test commands, or run them directly via shell.

### Q: Will multi-language support be added later?

**A**: Possibly, if there's strong user demand and a maintainer willing to own it. For now, Go-only is the right choice.

### Q: Is this inconsistent with `security` and `lint` being multi-language?

**A**: No. Those tools have simpler interfaces (scan and report) and higher value propositions. Testing frameworks are more complex and have good alternatives.

---

## Related Documentation

- **`docs/TOOL_LANGUAGE_COMPATIBILITY_MATRIX.md`** - Language compatibility reference
- **`docs/EXARP_ABILITIES_AUDIT.md`** - Complete tool catalog
- **`internal/tools/testing.go`** - Implementation

---

## Changelog

- **2026-03-08**: Initial decision - Go-only strategy chosen
- **Next**: Update tool descriptions and documentation

---

## Sign-off

**Decision**: Keep testing tool Go-only, improve documentation  
**Approved by**: Architecture review (automated)  
**Status**: ✅ Approved for implementation

**Closes**: T-1772958892671074000
