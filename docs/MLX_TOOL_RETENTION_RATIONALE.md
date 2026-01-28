# MLX Tool Retention Rationale

**Date:** 2026-01-12  
**Status:** ✅ Research Complete - Intentional Python Bridge Retention  
**Task:** T-1768170891804

---

## Executive Summary

After comprehensive research, the `mlx` tool will **remain Python bridge only** as an intentional architectural decision. No Go bindings are available for MLX, and the project already uses a Go-native alternative (`go-foundationmodels`) for core AI features.

---

## Research Findings

### 1. MLX Go Bindings Availability

**Search Results:**
- ❌ **No Official Go Bindings**: MLX does not provide official Go bindings
- ❌ **No Community Bindings**: No community-created Go bindings found on GitHub or elsewhere
- ❌ **No CGO Wrappers**: No existing CGO wrappers for MLX Metal framework integration

**Technical Barriers:**
- Would require CGO to interface with Metal Performance Shaders
- Would need Metal framework integration (Apple-specific)
- Significant development effort for optional feature
- No community interest or momentum for Go bindings

### 2. Go-Native Alternative: Apple Foundation Models

**Library:** `github.com/blacktop/go-foundationmodels v0.1.8`

**Status in exarp-go:**
- ✅ **Already Integrated**: Used for core AI features
- ✅ **Native Go**: No Python bridge required
- ✅ **Platform Support**: macOS arm64 with CGO (conditional compilation)

**Current Usage:**
- `estimation` tool - Task duration estimation with Apple FM
- `context` tool - Text summarization with Apple FM
- `task_analysis` tool - Hierarchy analysis with Apple FM
- `task_workflow` tool - Clarification question generation with Apple FM
- `apple_foundation_models` tool - Direct Apple Intelligence API access

**Comparison:**

| Feature | MLX (Python) | Apple Foundation Models (Go) |
|---------|--------------|------------------------------|
| **Language** | Python | Go (native) |
| **Platform** | Apple Silicon | Apple Silicon |
| **API Type** | ML model execution | Apple Intelligence APIs |
| **Use Case** | Run local ML models | On-device AI services |
| **Integration** | Python bridge required | Native Go (no bridge) |
| **Status** | Optional enhancement | Core functionality |
| **Purpose** | Report enhancements | Core AI features |

### 3. MLX Current Usage

**Location:** `internal/tools/handlers.go:790-804`

**Functionality:**
- Actions: `status`, `hardware`, `models`, `generate`
- Used by: `report` tool for optional AI-generated insights
- Purpose: Enhance reports with MLX-generated analysis
- Status: Optional - reports work without MLX

**Implementation:**
```go
// NOTE: MLX remains Python bridge only by design.
// Rationale: No Go bindings available for MLX (Apple's ML library).
// Creating Go bindings would require CGO and Metal framework integration,
// which is high effort for optional feature. MLX is used for optional
// report enhancements, not critical functionality.
func handleMlx(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    result, err := bridge.ExecutePythonTool(ctx, "mlx", params)
    // ...
}
```

**Report insights abstraction (2026-01-28):** Report/scorecard AI insights no longer call the bridge directly. They use `DefaultReportInsight()` (see `internal/tools/insight_provider.go`), which tries MLX via the bridge first, then falls back to DefaultFM (Apple FM) when MLX is unavailable. Report code does not depend on the bridge or MLX by name. See `docs/MLX_ABSTRACTION_OPPORTUNITIES.md`.

---

## Decision Rationale

### ✅ Keep MLX as Python Bridge Only

**Reasons:**
1. **No Go Bindings Available**: Confirmed via multiple web searches - no official or community bindings
2. **High Migration Effort**: Would require CGO + Metal framework integration (significant development)
3. **Optional Feature**: MLX is used for optional report enhancements, not critical functionality
4. **Go Alternative Exists**: Apple Foundation Models already provides Go-native AI capabilities
5. **Different Use Cases**: MLX (model execution) vs Apple FM (on-device AI APIs) serve different purposes
6. **Python Ecosystem Mature**: MLX Python ecosystem is well-developed with extensive model support
7. **Bridge Works Well**: Python bridge provides sufficient functionality for optional enhancements
8. **Low Priority**: MLX is not a core dependency, reports work without it

### ❌ Do Not Create Go Bindings

**Reasons:**
1. **High Development Effort**: CGO + Metal integration would require significant work
2. **No Community Support**: No existing bindings or community interest
3. **Low Benefit**: Optional feature doesn't justify high migration effort
4. **Alternative Available**: Go-native Apple Foundation Models already covers core AI needs
5. **Maintenance Burden**: Custom bindings would require ongoing maintenance

---

## Impact Assessment

### Positive Impacts

1. ✅ **No Migration Effort**: No work required to keep Python bridge
2. ✅ **Mature Ecosystem**: Benefits from MLX Python ecosystem
3. ✅ **Stability**: Python bridge implementation is stable
4. ✅ **Focus**: Resources can focus on higher-priority migrations
5. ✅ **Go Alternative**: Apple Foundation Models provides native Go AI capabilities

### Negative Impacts

1. ⚠️ **Python Dependency**: Keeps Python bridge dependency (minimal impact - optional feature)
2. ⚠️ **Platform-Specific**: Only works on Apple Silicon (not a concern - MLX requirement)
3. ⚠️ **Performance**: Python bridge has overhead (acceptable for optional enhancement)

---

## Documentation Status

### ✅ Completed Documentation

1. ✅ **MLX_EVALUATION_2026-01-12.md**: Comprehensive evaluation document
2. ✅ **MLX_TOOL_RETENTION_RATIONALE.md**: This document (retention rationale)
3. ✅ **Code Comments**: `internal/tools/handlers.go:790` - Intentional retention documented
4. ✅ **Migration Documents**: Updated with MLX as intentional Python bridge retention

### Migration Status Documents Updated

- ✅ `MIGRATION_STATUS_CURRENT.md`: MLX documented as intentional retention
- ✅ `PYTHON_BRIDGE_DEPENDENCIES.md`: MLX rationale documented
- ✅ `NATIVE_GO_MIGRATION_PLAN.md`: MLX marked as intentional retention
- ✅ `MIGRATION_CHECKLIST.md`: MLX marked as intentional retention (no migration needed)

---

## Conclusion

**MLX tool will remain Python bridge only** as an intentional architectural decision. This is the correct approach because:

1. ✅ No Go bindings are available (confirmed via research)
2. ✅ Creating bindings would be high effort with low benefit
3. ✅ Go-native alternative (Apple Foundation Models) already provides core AI features
4. ✅ MLX is optional enhancement, not critical functionality
5. ✅ Python bridge works well for this use case
6. ✅ Resources are better spent on higher-priority migrations

**Status:** ✅ **Research Complete - Documented as Intentional Retention**

**Next Review:** If Go bindings become available or requirements change

---

**Last Updated:** 2026-01-12  
**Task:** T-1768170891804 - Research and document MLX tool Go bindings or retention rationale
