# Next Migration Options

**Date**: 2026-01-09  
**Status**: After Automation Tool Migration

## Recently Completed

### ✅ Automation Tool (daily and discover actions)
- **Status**: Native Go ✅
- **Impact**: Daily action uses health docs (still Python bridge)
- **Next**: Health docs action would complete daily automation

---

## Current Migration Status

### Tools Already Native (Documentation May Be Outdated)

1. **security** ✅ **ALREADY NATIVE** (documentation outdated)
   - **Status**: Native Go for scan, alerts, report actions
   - **Implementation**: `internal/tools/security.go`
   - **Actions**: All 3 actions have native implementations
   - **Note**: Documentation says "Fully Python bridge" but it's actually native!

---

## Next Migration Candidates

### Option 1: Health Tool - docs Action (Recommended)

**Priority**: High (completes daily automation workflow)  
**Complexity**: Medium-High  
**Dependencies**: None (file operations)

**What it does**:
- Validates internal and external links in markdown files
- Validates documentation format
- Checks document currency (stale docs)
- Validates cross-references (uses NetworkX in Python, would need gonum in Go)
- Calculates health score
- Creates Todo2 tasks for issues

**Implementation approach**:
- Link validation: File system operations (straightforward in Go)
- Format validation: Runs external script (subprocess)
- Date currency: Regex parsing (straightforward)
- Cross-references: Would need gonum graph (complex, could defer)
- Health score: Simple calculation
- Task creation: Todo2 utilities exist

**Estimated effort**: 4-6 hours (simplified version without NetworkX cross-refs)

**Pros**:
- Completes daily automation migration
- High value (daily automation fully native)
- Mostly file operations (straightforward)

**Cons**:
- Cross-reference validation is complex (uses NetworkX)
- Could implement simplified version without NetworkX

---

### Option 2: Setup Hooks - patterns Action

**Priority**: Medium  
**Complexity**: Low-Medium  
**Dependencies**: None (file operations, pattern matching)

**What it does**:
- Sets up pattern-based triggers (file watchers)
- Configures file change detection
- May involve platform-specific implementations

**Estimated effort**: 2-3 hours

**Pros**:
- Simple (file operations)
- Low complexity

**Cons**:
- Lower priority than health docs
- Doesn't complete any workflow

---

### Option 3: Analyze Alignment - prd Action

**Priority**: Medium  
**Complexity**: Medium-High  
**Dependencies**: PRD parsing, alignment logic

**What it does**:
- Analyzes alignment between PRD and tasks
- More complex than todo2 alignment
- Requires PRD parsing and analysis

**Estimated effort**: 5-7 hours

**Pros**:
- Completes analyze_alignment tool migration

**Cons**:
- Lower priority than health docs
- More complex

---

## Recommendation

**Option 1: Health Tool - docs Action** (simplified version)

**Rationale**:
1. Completes daily automation migration (high value)
2. Mostly straightforward file operations
3. Can implement simplified version (skip NetworkX cross-refs for now)
4. High impact (daily automation is a core workflow)

**Implementation Plan**:
1. Implement link validation (file system)
2. Implement format validation (subprocess)
3. Implement date currency checks (regex)
4. Implement health score calculation
5. **Defer**: Cross-reference validation (can use Python bridge fallback)
6. **Defer**: Task creation (can use Python bridge fallback)

**Simplified Version**:
- Core functionality (links, format, currency)
- Skip NetworkX cross-reference validation (keep Python bridge for now)
- Skip task creation (keep Python bridge for now)
- Health score based on available checks

This gives us 80% of the value with 50% of the complexity.

---

## Alternative: Update Documentation

Before proceeding, we should:
1. Update `PYTHON_BRIDGE_DEPENDENCIES.md` to reflect security tool is already native
2. Verify which other tools might be incorrectly documented
