# Parallel Implementation Plan: mcp-go-core Extraction

**Date:** 2026-01-12  
**Tasks:** T-1768249093338, T-1768249096750, T-1768249100812  
**Status:** Planning

---

## Executive Summary

All three high-priority extraction tasks can be executed **in parallel** with minimal coordination. However, there is a logical dependency that should be considered for optimal workflow.

**Parallel Execution Groups:**
- **Group 1 (Independent):** Security Utilities, Common Types
- **Group 2 (After Common Types):** Framework Abstraction

---

## Task Analysis

### Task 1: Extract Framework Abstraction (T-1768249093338)
**Priority:** High  
**Estimated Duration:** 4-6 hours  
**Dependencies:** None (but benefits from Common Types being done first)

**Files:**
- `exarp-go/internal/framework/server.go`
- `exarp-go/internal/framework/factory.go`
- `exarp-go/internal/framework/adapters/gosdk/`

**Key Types Used:**
- `TextContent` - Currently defined in `server.go`
- `ToolSchema` - Currently defined in `server.go`
- `ToolInfo` - Currently defined in `server.go`

**Dependency Note:** Framework abstraction uses `TextContent`, `ToolSchema`, and `ToolInfo` types. These should ideally be extracted to `pkg/mcp/types/` first, then framework can import them.

---

### Task 2: Extract Common Types (T-1768249096750)
**Priority:** High  
**Estimated Duration:** 2-3 hours  
**Dependencies:** None

**Files:**
- `exarp-go/internal/framework/server.go` (TextContent, ToolSchema, ToolInfo)
- `devwisdom-go/internal/mcp/protocol.go` (similar types)

**Destination:**
- `mcp-go-core/pkg/mcp/types/common.go`

**Key Types:**
- `TextContent`
- `ToolSchema`
- `ToolInfo`

**Dependency Note:** This is a foundational task - other tasks can benefit from having these types available first.

---

### Task 3: Verify Security Utilities (T-1768249100812)
**Priority:** High  
**Estimated Duration:** 3-4 hours  
**Dependencies:** None

**Files:**
- ✅ Already extracted: `mcp-go-core/pkg/mcp/security/path.go`
- ⏳ To extract: `exarp-go/internal/security/access.go`
- ⏳ To extract: `exarp-go/internal/security/ratelimit.go`

**Dependency Note:** Completely independent - can run in parallel with everything.

---

## Recommended Parallel Execution Strategy

### Phase 1: Foundation (Parallel - 2-3 hours)

**Execute in parallel:**
1. **Task 2: Extract Common Types** (T-1768249096750)
   - Analyze type differences between projects
   - Create unified type definitions
   - Write tests
   - **Output:** `pkg/mcp/types/common.go` ready

2. **Task 3: Verify Security Utilities** (T-1768249100812)
   - Extract `access.go` and `ratelimit.go`
   - Write comprehensive tests
   - Document security best practices
   - **Output:** Complete security package

**Why parallel:**
- No dependencies between them
- Different codebases (types vs security)
- Can be done by different developers simultaneously

---

### Phase 2: Framework (After Phase 1 - 4-6 hours)

**Execute after Phase 1:**
3. **Task 1: Extract Framework Abstraction** (T-1768249093338)
   - Import types from `pkg/mcp/types` (created in Phase 1)
   - Extract framework files
   - Update imports to use shared types
   - Write tests
   - **Output:** Complete framework package using shared types

**Why sequential:**
- Framework uses `TextContent`, `ToolSchema`, `ToolInfo`
- Better to import from `pkg/mcp/types` than duplicate
- Cleaner architecture with proper dependencies

---

## Alternative: Fully Parallel Approach

If speed is critical and you're willing to handle minor refactoring:

**All 3 tasks can run in parallel:**
- Task 2 creates types in `pkg/mcp/types/`
- Task 1 creates framework with types inline (refactor later)
- Task 3 is completely independent

**Refactoring step after:**
- Move types from framework to `pkg/mcp/types/`
- Update framework imports
- Verify everything still works

**Trade-off:**
- ✅ Faster initial completion
- ❌ Requires refactoring step
- ❌ Risk of duplicate type definitions

---

## Detailed Work Breakdown

### Task 2: Common Types (Can start immediately)

**Steps:**
1. [30 min] Analyze types in both projects
   - Compare `exarp-go/internal/framework/server.go`
   - Compare `devwisdom-go/internal/mcp/protocol.go`
   - Document differences

2. [1 hour] Create unified types
   - Write `pkg/mcp/types/common.go`
   - Ensure JSON tags are consistent
   - Add validation if needed

3. [30 min] Write tests
   - Serialization/deserialization tests
   - JSON marshaling tests
   - Type validation tests

4. [30 min] Verify and document
   - Update README
   - Add examples
   - Verify build

**Total:** ~2.5 hours

---

### Task 3: Security Utilities (Can start immediately)

**Steps:**
1. [1 hour] Extract `access.go`
   - Copy from `exarp-go/internal/security/access.go`
   - Update imports
   - Verify functionality

2. [1 hour] Extract `ratelimit.go`
   - Copy from `exarp-go/internal/security/ratelimit.go`
   - Update imports
   - Verify functionality

3. [1 hour] Write comprehensive tests
   - Test access control functions
   - Test rate limiting functions
   - Integration tests

4. [30 min] Documentation
   - Document security best practices
   - Add usage examples
   - Update README

**Total:** ~3.5 hours

---

### Task 1: Framework Abstraction (Start after Task 2)

**Steps:**
1. [1 hour] Extract `server.go`
   - Copy interface definitions
   - **Import types from `pkg/mcp/types`** (from Task 2)
   - Update package name

2. [1 hour] Extract `factory.go`
   - Copy factory functions
   - Update imports
   - Verify dependencies

3. [1.5 hours] Extract `adapters/gosdk/`
   - Copy adapter implementation
   - Update imports
   - Verify SDK integration

4. [1 hour] Write tests
   - Interface tests
   - Factory tests
   - Adapter tests

5. [1 hour] Integration and verification
   - Update exarp-go to use shared library
   - Verify backward compatibility
   - Run all tests

**Total:** ~5.5 hours

---

## Resource Allocation

### Option 1: Single Developer (Sequential)
**Timeline:** ~11 hours total
- Phase 1: 3.5 hours (Task 2 + Task 3 in parallel)
- Phase 2: 5.5 hours (Task 1)
- **Total:** ~9 hours (with some parallel work)

### Option 2: Two Developers (Optimal Parallel)
**Timeline:** ~5.5 hours total
- Developer 1: Task 2 (2.5 hours) → Task 1 (5.5 hours) = 8 hours
- Developer 2: Task 3 (3.5 hours) → Help with Task 1 testing
- **Critical path:** 5.5 hours (Task 1)

### Option 3: Three Developers (Fully Parallel)
**Timeline:** ~5.5 hours total
- Developer 1: Task 2 (2.5 hours)
- Developer 2: Task 3 (3.5 hours)
- Developer 3: Task 1 (5.5 hours, starts after Task 2)
- **Critical path:** 5.5 hours (Task 1)

---

## Risk Mitigation

### Risk 1: Type Conflicts
**Mitigation:**
- Task 2 should complete before Task 1 starts
- If parallel, ensure types are extracted first
- Use clear naming conventions

### Risk 2: Import Path Issues
**Mitigation:**
- Use local replace in go.mod during development
- Test imports before committing
- Verify both projects can import

### Risk 3: Breaking Changes
**Mitigation:**
- Keep backward compatibility wrappers in exarp-go
- Gradual migration approach
- Comprehensive testing

---

## Success Criteria

### Task 2 (Common Types)
- [ ] Types defined in `pkg/mcp/types/common.go`
- [ ] All tests passing
- [ ] JSON serialization working
- [ ] Documentation complete

### Task 3 (Security Utilities)
- [ ] `access.go` extracted and tested
- [ ] `ratelimit.go` extracted and tested
- [ ] All security tests passing
- [ ] Documentation complete

### Task 1 (Framework Abstraction)
- [ ] Framework interfaces extracted
- [ ] Factory functions extracted
- [ ] Adapter implementation extracted
- [ ] Uses shared types from `pkg/mcp/types`
- [ ] All tests passing
- [ ] exarp-go integration verified

---

## Recommended Approach

**For fastest completion with clean architecture:**

1. **Start:** Task 2 and Task 3 in parallel (immediately)
2. **After Task 2 completes:** Start Task 1 (uses types from Task 2)
3. **After all complete:** Integration testing and documentation

**Timeline:**
- **Phase 1 (Parallel):** 0-3.5 hours
  - Task 2: 0-2.5 hours
  - Task 3: 0-3.5 hours
- **Phase 2 (Sequential):** 2.5-8 hours
  - Task 1: 2.5-8 hours (starts when Task 2 completes)
- **Total:** ~8 hours (with optimal parallelization)

---

## Next Steps

1. ✅ Review and approve this plan
2. ⏳ Assign tasks to developers (if multiple)
3. ⏳ Set up development branches
4. ⏳ Begin Phase 1 (Task 2 + Task 3 in parallel)
5. ⏳ Begin Phase 2 (Task 1) after Task 2 completes
6. ⏳ Integration testing
7. ⏳ Documentation updates

---

## Notes

- All tasks are high priority
- Task 2 should complete before Task 1 for clean architecture
- Task 3 is completely independent
- Consider using feature branches for each task
- Coordinate on import paths and package structure
