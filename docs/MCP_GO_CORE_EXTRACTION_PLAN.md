# mcp-go-core Extraction Plan

**Date:** 2026-01-12  
**Status:** Planning Complete  
**Reference:** See `/home/dlowes/projects/mcp-go-core/PARALLEL_IMPLEMENTATION_PLAN.md` for detailed plan

---

## Tasks Created

Three high-priority tasks have been created for extracting shared code to `mcp-go-core`:

1. **T-1768249093338** - Extract Framework Abstraction to mcp-go-core
2. **T-1768249096750** - Extract Common Types to mcp-go-core  
3. **T-1768249100812** - Verify Security Utilities in mcp-go-core

---

## Parallel Execution Strategy

### Phase 1: Foundation (Parallel Execution)

**Execute simultaneously:**
- **Task 2** (Common Types) - 2.5 hours
- **Task 3** (Security Utilities) - 3.5 hours

**Why parallel:**
- No dependencies between them
- Different codebases
- Can be done by different developers

### Phase 2: Framework (After Phase 1)

**Execute after Task 2 completes:**
- **Task 1** (Framework Abstraction) - 5.5 hours

**Why sequential:**
- Framework uses `TextContent`, `ToolSchema`, `ToolInfo` types
- Should import from `pkg/mcp/types` (created in Task 2)
- Cleaner architecture with proper dependencies

---

## Timeline

**Optimal (2 developers):**
- **Phase 1:** 0-3.5 hours (parallel)
- **Phase 2:** 2.5-8 hours (Task 1 starts when Task 2 completes)
- **Total:** ~8 hours

**Single developer:**
- **Phase 1:** 0-3.5 hours (sequential: Task 2 then Task 3)
- **Phase 2:** 3.5-9 hours (Task 1)
- **Total:** ~9 hours

---

## Dependencies

```
Task 2 (Common Types)
    ↓
Task 1 (Framework Abstraction) - imports types from Task 2

Task 3 (Security Utilities) - independent
```

---

## Next Steps

1. ✅ Tasks created
2. ✅ Parallel plan documented
3. ⏳ Assign developers (if multiple)
4. ⏳ Begin Phase 1 (Task 2 + Task 3)
5. ⏳ Begin Phase 2 (Task 1) after Task 2
6. ⏳ Integration testing
7. ⏳ Documentation updates

---

## Detailed Plan

See `/home/dlowes/projects/mcp-go-core/PARALLEL_IMPLEMENTATION_PLAN.md` for:
- Detailed work breakdown
- Resource allocation options
- Risk mitigation strategies
- Success criteria
- Step-by-step instructions
