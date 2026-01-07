# Native Go Markdown Linting Research

**Date:** 2026-01-07  
**Purpose:** Research native Go markdown linting libraries to replace external dependencies

---

## Executive Summary

**Recommended Solution:** `gomarklint` - A fast, lightweight, native Go markdown linter that can be integrated directly into exarp-go without external dependencies.

**Key Benefits:**
- ✅ **Native Go** - No npm/Ruby dependencies required
- ✅ **Fast** - 157 files, 52,000+ lines scanned in <50ms
- ✅ **CI-Friendly** - Designed for continuous integration
- ✅ **JSON Output** - Structured output for programmatic use
- ✅ **Configurable** - Supports `.gomarklint.json` configuration files

---

## Primary Option: gomarklint

### Repository
- **GitHub:** `github.com/shinagawa-web/gomarklint`
- **Language:** Pure Go
- **Status:** Active (as of 2026)

### Features

1. **Core Linting Capabilities:**
   - Heading level consistency checks
   - Duplicate heading detection
   - Missing trailing blank lines
   - Unclosed code blocks
   - YAML frontmatter handling
   - Broken external link detection (with `--enable-link-check`)

2. **Configuration:**
   - Configuration file: `.gomarklint.json`
   - Ignore patterns support (e.g., `**/CHANGELOG.md`)
   - Default options storage

3. **Output Formats:**
   - Standard text output
   - Structured JSON output (`--output=json`)
   - CI-friendly error format

4. **Performance:**
   - **Blazing fast:** 157 files, 52,000+ lines scanned in under 50ms
   - Minimal memory footprint
   - Efficient directory traversal

### Installation

```bash
# As a library (for integration)
go get github.com/shinagawa-web/gomarklint

# As a CLI tool (for testing)
go install github.com/shinagawa-web/gomarklint@latest
```

### Usage Examples

#### As a Library (Programmatic)

```go
package main

import (
    "fmt"
    "github.com/shinagawa-web/gomarklint"
)

func main() {
    // Lint a single file
    results, err := gomarklint.LintFile("README.md")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    // Process results
    for _, issue := range results.Issues {
        fmt.Printf("%s:%d:%d %s\n", 
            issue.File, issue.Line, issue.Column, issue.Message)
    }
}
```

#### As a CLI Tool

```bash
# Lint a single file
gomarklint README.md

# Lint a directory
gomarklint docs/

# JSON output
gomarklint --output=json docs/

# Enable link checking
gomarklint --enable-link-check docs/
```

### Integration into exarp-go

**Approach 1: Direct Library Integration (Recommended)**

```go
// internal/tools/linting.go

import (
    "github.com/shinagawa-web/gomarklint"
)

func runGomarklint(ctx context.Context, path string, fix bool) (*LintResult, error) {
    // Use gomarklint library directly
    results, err := gomarklint.LintPath(path)
    if err != nil {
        return &LintResult{
            Success: false,
            Linter:  "gomarklint",
            Output:  err.Error(),
            Errors:  []LintError{{Message: err.Error(), Severity: "error"}},
        }, nil
    }
    
    // Convert gomarklint results to LintResult
    var lintErrors []LintError
    for _, issue := range results.Issues {
        lintErrors = append(lintErrors, LintError{
            File:     issue.File,
            Line:     issue.Line,
            Column:   issue.Column,
            Message:  issue.Message,
            Rule:     issue.Rule,
            Severity: "warning",
        })
    }
    
    return &LintResult{
        Success: len(lintErrors) == 0,
        Linter:  "gomarklint",
        Output:  results.String(),
        Errors:  lintErrors,
    }, nil
}
```

**Approach 2: CLI Execution (Fallback)**

If the library API is not available, execute the CLI tool:

```go
func runGomarklint(ctx context.Context, path string, fix bool) (*LintResult, error) {
    // Check if gomarklint is available
    if _, err := exec.LookPath("gomarklint"); err != nil {
        return &LintResult{
            Success: false,
            Linter:  "gomarklint",
            Output:  "gomarklint not found. Install with: go install github.com/shinagawa-web/gomarklint@latest",
            Errors: []LintError{
                {Message: "gomarklint binary not found", Severity: "error"},
            },
        }, nil
    }
    
    // Execute gomarklint CLI
    args := []string{"--output=json"}
    if fix {
        args = append(args, "--fix")
    }
    args = append(args, path)
    
    cmd := exec.CommandContext(ctx, "gomarklint", args...)
    output, err := cmd.CombinedOutput()
    
    // Parse JSON output
    // ... (implementation)
}
```

---

## Comparison: Native Go vs External Tools

| Feature | gomarklint (Go) | markdownlint-cli (npm) | markdownlint (Ruby) |
|---------|----------------|------------------------|---------------------|
| **Language** | Go (native) | Node.js | Ruby |
| **Dependencies** | None (pure Go) | npm, Node.js | Ruby, gem |
| **Performance** | Very fast (<50ms) | Fast | Moderate |
| **Integration** | Direct library | CLI only | CLI only |
| **CI/CD** | Excellent | Good | Good |
| **Maintenance** | Go ecosystem | npm ecosystem | Ruby ecosystem |
| **Binary Size** | Small (Go binary) | Large (Node.js) | Moderate |

---

## Implementation Recommendations

### Phase 1: Add gomarklint Dependency

```bash
cd /Users/davidl/Projects/exarp-go
go get github.com/shinagawa-web/gomarklint
```

### Phase 2: Update Linting Implementation

1. **Replace `runMarkdownlint` function** to use gomarklint library
2. **Update auto-detection** to prefer gomarklint over external tools
3. **Add configuration support** for `.gomarklint.json`
4. **Maintain backward compatibility** with external tools as fallback

### Phase 3: Testing

1. Test with various markdown files
2. Test directory linting
3. Test JSON output parsing
4. Test error handling
5. Verify performance

---

## Configuration File Example

`.gomarklint.json`:

```json
{
  "ignore": [
    "**/CHANGELOG.md",
    "**/node_modules/**",
    "**/vendor/**"
  ],
  "enableLinkCheck": false,
  "rules": {
    "heading-increment": true,
    "no-duplicate-heading": true,
    "trailing-blank-line": true
  }
}
```

---

## Alternative Options (Not Recommended)

### 1. markdownlint-cli (npm)
- **Pros:** Feature-rich, widely used
- **Cons:** Requires Node.js/npm, external dependency
- **Status:** Not recommended for native Go integration

### 2. markdownlint (Ruby gem)
- **Pros:** Mature, feature-rich
- **Cons:** Requires Ruby, external dependency
- **Status:** Not recommended for native Go integration

### 3. Custom Implementation
- **Pros:** Full control, no dependencies
- **Cons:** Significant development effort, maintenance burden
- **Status:** Not recommended unless specific requirements

---

## Next Steps

1. ✅ **Research Complete** - gomarklint identified as best option
2. ⏳ **Add Dependency** - `go get github.com/shinagawa-web/gomarklint`
3. ⏳ **Update Implementation** - Replace `runMarkdownlint` with gomarklint library
4. ⏳ **Test Integration** - Verify functionality with real markdown files
5. ⏳ **Update Documentation** - Document markdown linting capabilities

---

## References

- **gomarklint GitHub:** `github.com/shinagawa-web/gomarklint`
- **Libraries.io:** https://libraries.io/go/github.com%2Fshinagawa-web%2Fgomarklint
- **markdownlint (npm):** https://github.com/DavidAnson/markdownlint

---

**Status:** ✅ Research complete - Ready for implementation

