# Scorecard Auto-Fix Investigation

**Date:** 2026-02-22
**Status:** Research / Documentation
**Task:** T-1771535382892408000
**Tag hints:** `#quality` `#refactor`

---

## Current Auto-Fix Capabilities

`make scorecard-fix` runs three targets in sequence:

```makefile
scorecard-fix: go-mod-tidy fmt lint-fix
```

| Target | What it fixes |
|--------|---------------|
| `go-mod-tidy` | Missing/unused module dependencies in go.mod/go.sum |
| `fmt` | Go formatting (gofmt/goimports via exarp-go) |
| `lint-fix` | golangci-lint auto-fixable issues (unused imports, simple style) |

After running, re-verify with `make scorecard-full`.

## Scorecard Dimensions and Fix Potential

The scorecard (`internal/tools/scorecard_go.go`) calculates scores across these dimensions:

### Currently Auto-Fixable

| Dimension | Check | Auto-fix? | How |
|-----------|-------|-----------|-----|
| Code Quality | `GoFmtCompliant` | Yes | `make fmt` |
| Code Quality | `GoModTidyPasses` | Yes | `make tidy` |
| Code Quality | `GoLintPasses` | Partial | `make lint-fix` (only some lint rules) |
| Code Quality | `GoVetPasses` | No | Requires manual review |

### Potentially Auto-Fixable (Not Yet Implemented)

| Item | Feasibility | Approach |
|------|-------------|----------|
| Missing file comments | High | `make lint-nav-fix` already suggests fixes; could be automated with a script that prepends `// filename.go — ...` based on package/filename |
| Magic string literals | Medium | Replace `"Todo"`, `"In Progress"`, `"Done"`, `"Review"` with `models.StatusTodo` etc. via `sed` or Go AST rewriting |
| Missing package doc comments | Medium | Generate `// Package X ...` from directory name and existing file comments |
| Import ordering | High | Already handled by `goimports` in `make fmt` |

### Cannot Be Auto-Fixed

| Dimension | Check | Why |
|-----------|-------|-----|
| Testing | `GoTestPasses` | Requires fixing test logic |
| Testing | `GoTestCoverage` | Requires writing new tests |
| Security | `GoVulnCheckPasses` | Requires dependency updates/evaluation |
| Security | `PathBoundaryEnforcement` | Architectural decision |
| Security | `RateLimiting` | Requires implementation |
| Documentation | `ReadmeExists` | Content requires human judgment |
| Documentation | `DocsFileCount` | Requires writing meaningful docs |
| Architecture | Large file splits | Requires understanding of code structure |
| Build | `GoBuildPasses` | Requires fixing compilation errors |

## Recommendation: Additional Auto-Fix Rules

### Priority 1: File Comment Auto-Fix (High Value, Low Risk)

Add a `make lint-nav-fix-auto` target that:
1. Scans `.go` files > 50 lines without a file comment before line 5
2. Generates a comment from the filename: `// filename.go — [Package] implementation.`
3. Inserts it before the `package` line

This would improve the AI navigability score with zero risk of breaking code.

### Priority 2: Magic String Replacement (Medium Value, Medium Risk)

Create a `make fix-magic-strings` target using Go AST rewriting:
1. Find bare `"Todo"`, `"In Progress"`, `"Done"`, `"Review"` in `internal/tools/*.go`
2. Replace with `models.StatusTodo`, `models.StatusInProgress`, `models.StatusDone`, `models.StatusReview`
3. Add `import "github.com/davidl71/exarp-go/internal/models"` if missing

This requires AST rewriting to avoid breaking strings in comments, test data, or user-facing messages.

### Priority 3: Enhanced scorecard-fix Target

Extend `make scorecard-fix` to include the new auto-fixers:

```makefile
scorecard-fix: go-mod-tidy fmt lint-fix lint-nav-fix-auto fix-magic-strings
```

## Summary

| Category | Current | Potential | Effort |
|----------|---------|-----------|--------|
| Formatting | Fully auto-fixed | — | — |
| Module deps | Fully auto-fixed | — | — |
| Lint issues | Partially auto-fixed | — | — |
| File comments | Manual | Auto-fixable | 2-3h |
| Magic strings | Manual | Auto-fixable | 4-6h |
| Test coverage | Manual | Cannot auto-fix | — |
| Security | Manual | Cannot auto-fix | — |

## References

- `Makefile` — `scorecard-fix`, `lint-nav`, `lint-nav-fix` targets
- `internal/tools/scorecard_go.go` — Scorecard calculation and dimensions
- `.cursor/rules/ai-navigability.mdc` — File comment and magic string rules
