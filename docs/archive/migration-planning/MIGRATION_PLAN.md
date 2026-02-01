# Migration Plan: project-management-automation → exarp-go

**Date:** 2026-01-07  
**Status:** 📋 Planning Phase  
**Project:** Complete migration of remaining tools, prompts, and resources

---

## Executive Summary

This document outlines the comprehensive migration plan to move all remaining tools, prompts, and resources from `project-management-automation` (Python FastMCP) to `exarp-go` (Go MCP server).

**Current State:**
- ✅ **exarp-go**: 24 tools migrated, 15 prompts, 6 resources
- 📋 **project-management-automation**: ~100+ tools remaining, ~30+ prompts, ~10+ resources

**Goal:**
- Migrate all remaining functionality to exarp-go
- Maintain backward compatibility during migration
- Complete migration with full testing and validation

---

## Phase 1: Analysis and Inventory

### Objectives
1. Complete inventory of all tools in project-management-automation
2. Identify which tools are already migrated (24 tools)
3. Identify gaps and remaining tools to migrate
4. Document dependencies and relationships
5. Create migration priority matrix

### Deliverables
- `/Users/davidl/Projects/exarp-go/docs/MIGRATION_INVENTORY.md`
- Tool comparison matrix (Python vs Go)
- Dependency graph
- Priority assessment

### Key Tasks
- [ ] Analyze `project-management-automation/project_management_automation/server.py`
- [ ] Compare with `exarp-go/internal/tools/registry.go`
- [ ] Identify FastMCP dependencies vs standalone functions
- [ ] Map tool dependencies
- [ ] Create priority matrix (high/medium/low)

---

## Phase 2: Migration Strategy and Architecture

### Objectives
1. Design migration strategy (native Go vs Python bridge)
2. Create decision framework for each tool
3. Design architecture for new implementations
4. Plan dependency migration
5. Create testing strategy

### Deliverables
- `/Users/davidl/Projects/exarp-go/docs/MIGRATION_STRATEGY.md`
- `/Users/davidl/Projects/exarp-go/docs/MIGRATION_PATTERNS.md`
- Architecture decision records
- Testing plan

### Key Decisions
- **Native Go** vs **Python Bridge**: When to use each?
- **FastMCP Context**: How to handle tools requiring Context (elicit, interactive)?
- **Dependencies**: Python libs → Go alternatives or bridge?
- **Performance**: Optimization targets?

### Migration Patterns

#### Pattern 1: Native Go Implementation
- **Use when**: Simple logic, no complex Python deps
- **Example**: `lint` tool (Go linters only)
- **Benefits**: Performance, no Python bridge overhead

#### Pattern 2: Python Bridge
- **Use when**: Complex Python logic, heavy dependencies
- **Example**: `memory`, `security`, `automation` tools
- **Benefits**: Faster migration, preserve existing logic

#### Pattern 3: Hybrid Approach
- **Use when**: Tool has multiple components
- **Example**: `lint` tool (Go for Go linters, Python for others)
- **Benefits**: Best of both worlds

---

## Phase 3: Migrate High-Priority Tools (Batch 1)

### Objectives
1. Migrate 5-10 high-priority tools
2. Implement Go handlers and bridge scripts
3. Register tools in registry.go (Batch 4)
4. Integration testing
5. Documentation updates

### Selection Criteria
- **High Priority**: Frequently used, critical functionality
- **Medium Priority**: Important but less frequent
- **Low Priority**: Nice-to-have, can be deferred

### Migration Steps (per tool)
1. Analyze tool implementation in Python
2. Decide: Native Go vs Python Bridge
3. Implement Go handler (or bridge script)
4. Register in `internal/tools/registry.go`
5. Update `bridge/execute_tool.py` (if using bridge)
6. Write integration tests
7. Update documentation

### Deliverables
- Batch 4 tools in registry.go
- Integration tests for batch 1 tools
- Updated bridge scripts
- Documentation updates

---

## Phase 4: Migrate Remaining Tools (Batch 2)

### Objectives
1. Complete migration of all remaining tools
2. Full integration testing
3. No regressions
4. Complete documentation

### Approach
- Follow established patterns from Batch 1
- Parallel work where possible
- Comprehensive testing
- Incremental rollout

### Deliverables
- All tools migrated
- Complete integration test suite
- Updated README.md (tool count)
- Migration status report

---

## Phase 5: Migrate Prompts and Resources

### Objectives
1. Migrate all prompts (15+ prompts)
2. Migrate all resources (6+ resources)
3. Implement handlers
4. Integration testing

### Prompts
- **Current**: 15 prompts migrated
- **Remaining**: ~30+ prompts in project-management-automation
- **Pattern**: Python bridge for prompt retrieval

### Resources
- **Current**: 6 resources migrated
- **Remaining**: ~10+ resources in project-management-automation
- **Pattern**: Python bridge for resource retrieval

### Migration Steps
1. Analyze prompt/resource implementations
2. Update bridge scripts (`get_prompt.py`, `execute_resource.py`)
3. Register in Go handlers
4. Integration testing
5. Documentation

### Deliverables
- All prompts migrated
- All resources migrated
- Integration tests
- Updated documentation

---

## Phase 6: Testing, Validation, and Cleanup

### Objectives
1. Complete testing and validation
2. Performance validation
3. Documentation updates
4. Migration summary
5. Deprecation notices

### Testing Requirements
- [ ] Full integration test coverage
- [ ] MCP interface validation
- [ ] Tool/prompt/resource functionality
- [ ] Error handling validation
- [ ] Performance benchmarks (if applicable)

### Documentation
- [ ] Update README.md
- [ ] Create MIGRATION_COMPLETE.md
- [ ] Create MIGRATION_VALIDATION.md
- [ ] Update project-management-automation README (deprecation)

### Cleanup
- [ ] Remove unused code
- [ ] Update comments
- [ ] Code organization
- [ ] Deprecation notices

---

## Migration Statistics

### Current State (exarp-go)
- **Tools**: 24 migrated
- **Prompts**: 15 migrated
- **Resources**: 6 migrated

### Target State (project-management-automation)
- **Tools**: ~100+ tools (estimated)
- **Prompts**: ~30+ prompts (estimated)
- **Resources**: ~10+ resources (estimated)

### Migration Progress
- **Tools**: 24 / ~100+ (24%)
- **Prompts**: 15 / ~30+ (50%)
- **Resources**: 6 / ~10+ (60%)

---

## Risk Assessment

### High Risk
- **FastMCP Context Dependencies**: Tools requiring `elicit()` or interactive features
- **Complex Python Dependencies**: Heavy ML libraries (MLX, Ollama wrappers)
- **State Management**: Tools with complex state

### Medium Risk
- **Parameter Types**: Tools with unusual parameter structures
- **Return Types**: Complex return data structures
- **Error Handling**: Tools with unique error patterns

### Low Risk
- **Simple Tools**: Straightforward implementations
- **Standard Patterns**: Tools following common patterns
- **Well-Tested Tools**: Tools with existing test coverage

---

## Success Criteria

### Phase 1 ✅
- Complete inventory documented
- Gaps identified
- Priority matrix created

### Phase 2 ✅
- Migration strategy approved
- Architecture decisions documented
- Patterns established

### Phase 3 ✅
- Batch 1 tools migrated and tested
- Integration tests passing
- Documentation updated

### Phase 4 ✅
- All tools migrated
- Full test coverage
- No regressions

### Phase 5 ✅
- All prompts migrated
- All resources migrated
- Integration tests passing

### Phase 6 ✅
- Full validation complete
- Documentation updated
- Migration summary published
- Deprecation notices posted

---

## Timeline Estimate

- **Phase 1**: 2-3 days (analysis and inventory)
- **Phase 2**: 1-2 days (strategy and architecture)
- **Phase 3**: 1-2 weeks (high-priority tools, 5-10 tools)
- **Phase 4**: 2-3 weeks (remaining tools, ~50-70 tools)
- **Phase 5**: 1 week (prompts and resources)
- **Phase 6**: 1 week (testing, validation, cleanup)

**Total Estimate**: 6-8 weeks

---

## Next Steps

1. **Start Phase 1**: Complete analysis and inventory
2. **Review Strategy**: Validate migration approach
3. **Begin Migration**: Start with high-priority tools
4. **Iterate**: Continue migration in batches
5. **Validate**: Complete testing and validation

---

## References

- [exarp-go Migration Complete Summary](/Users/davidl/Projects/exarp-go/MIGRATION_FINAL_SUMMARY.md)
- [Bridge Analysis](/Users/davidl/Projects/exarp-go/docs/BRIDGE_ANALYSIS.md)
- [Project Management Automation README](/Users/davidl/Projects/project-management-automation/README.md)
- [exarp-go README](/Users/davidl/Projects/exarp-go/README.md)

---

**Last Updated**: 2026-01-07  
**Status**: Planning Phase  
**Next Review**: After Phase 1 completion

