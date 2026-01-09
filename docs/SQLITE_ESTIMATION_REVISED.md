# SQLite Migration - Revised Time Estimates

**Created:** 2026-01-09  
**Method:** Exarp-go estimation tool + complexity analysis  
**Status:** Updated Estimates

## Estimation Methodology

Estimates combine:
1. **Exarp-go estimation tool** - Historical data + MLX analysis
2. **Complexity analysis** - Based on task breakdown and requirements
3. **Realistic adjustments** - Accounting for testing, integration, and edge cases

## Main Migration Tasks

### sqlite-migration-1: Research & Schema Design

**Tool Estimate:** 1.2 hours  
**Adjusted Estimate:** **8-12 hours (1-1.5 days)**

**Reasoning:**
- Research SQLite Go best practices: 2-3 hours
- Design complete schema (7 tables): 2-3 hours
- Design indexes and constraints: 1-2 hours
- Migration strategy planning: 1-2 hours
- Documentation: 2-3 hours

**Confidence:** Medium (research tasks are variable)

---

### sqlite-migration-2: Database Infrastructure

**Tool Estimate:** 0.3 hours (seems too low)  
**Adjusted Estimate:** **16-24 hours (2-3 days)**

**Reasoning:**
- `sqlite.go` - Connection, initialization, pooling: 4-6 hours
- `schema.go` - Schema definitions, constants: 2-3 hours
- `migrations.go` - Migration system, version tracking: 4-6 hours
- `migrations/001_initial_schema.sql` - Complete SQL schema: 2-3 hours
- Testing infrastructure components: 4-6 hours

**Confidence:** High (well-defined scope)

---

### sqlite-migration-3: CRUD Operations

**Tool Estimate:** 1.2 hours (seems too low)  
**Adjusted Estimate:** **32-48 hours (4-6 days sequential, 3-4 days parallel)**

**Breakdown by Group:**

#### Group 1: Core CRUD (4 subtasks)
- **T3.1:** CreateTask - 4-6 hours
- **T3.2:** GetTask - 4-6 hours
- **T3.3:** UpdateTask - 6-8 hours (complex transaction logic)
- **T3.4:** DeleteTask - 2-3 hours
- **Total (Sequential):** 16-23 hours
- **Total (Parallel):** 8-10 hours (T3.1+T3.2 parallel, then T3.3, then T3.4)

#### Group 2: Query Operations (6 subtasks)
- **T3.5:** ListTasks - 6-8 hours (complex filtering)
- **T3.6-T3.10:** Query functions - 8-12 hours total (can be parallel)
  - T3.6: GetTasksByStatus - 1-2 hours
  - T3.7: GetTasksByTag - 2-3 hours
  - T3.8: GetTasksByPriority - 1-2 hours
  - T3.9: GetTasksByProject - 1-2 hours
  - T3.10: SearchTasks - 3-4 hours (FTS5 integration)
- **Total (Sequential):** 14-20 hours
- **Total (Parallel):** 8-10 hours (T3.5 first, then T3.6-T3.10 parallel)

#### Group 3: Relationships (4 subtasks)
- **T3.11:** Tag operations - 3-4 hours
- **T3.12:** Dependency operations - 4-6 hours (validation logic)
- **T3.13:** GetDependents - 2-3 hours
- **T3.14:** ValidateDependencyChain - 4-6 hours (graph algorithm)
- **Total (Sequential):** 13-19 hours
- **Total (Parallel):** 8-10 hours (T3.11+T3.12 parallel, then T3.13+T3.14)

#### Group 4: History & Comments (4 subtasks)
- **T3.15:** Change tracking - 2-3 hours
- **T3.16:** GetTaskHistory - 2-3 hours
- **T3.17:** Comment operations - 3-4 hours
- **T3.18:** Activity log - 3-4 hours
- **Total (Sequential):** 10-14 hours
- **Total (Parallel):** 5-6 hours (all can be parallel)

#### Group 5: Utilities (3 subtasks)
- **T3.19:** TaskFilters - 2-3 hours
- **T3.20:** Conversion helpers - 3-4 hours
- **T3.21:** Prepared statements - 3-4 hours
- **Total (Sequential):** 8-11 hours
- **Total (Parallel):** 4-5 hours (all can be parallel)

#### Group 6: Testing (5 test suites)
- **T3.22:** Core CRUD tests - 6-8 hours
- **T3.23:** Query tests - 6-8 hours
- **T3.24:** Relationship tests - 6-8 hours
- **T3.25:** History/comment tests - 3-4 hours
- **T3.26:** Integration tests - 6-8 hours
- **Total:** 27-36 hours (can overlap with implementation)

**Total Task 3 (Sequential):** 80-106 hours (10-13 days)  
**Total Task 3 (Parallel):** 48-64 hours (6-8 days)  
**With Testing Overlap:** **32-48 hours (4-6 days)**

**Confidence:** High (detailed breakdown available)

---

### sqlite-migration-4: Data Migration Tool

**Tool Estimate:** 1.2 hours (seems too low)  
**Adjusted Estimate:** **16-24 hours (2-3 days)**

**Reasoning:**
- JSON parsing logic: 3-4 hours
- Data validation logic: 3-4 hours
- Migration execution logic: 4-6 hours
- Rollback mechanism: 2-3 hours
- CLI tool interface: 2-3 hours
- Testing with real data: 2-4 hours

**Confidence:** Medium (migration tools can have edge cases)

---

### sqlite-migration-5: Code Simplification

**Tool Estimate:** 4 hours (historical match)  
**Adjusted Estimate:** **8-12 hours (1-1.5 days)**

**Reasoning:**
- Remove file locking calls: 2-3 hours
- Remove JSON cache usage: 2-3 hours
- Update error handling: 2-3 hours
- Update documentation: 1-2 hours
- Regression testing: 1-2 hours

**Confidence:** High (well-defined refactoring)

---

## Revised Timeline Summary

### Sequential Execution

| Task | Original Estimate | Revised Estimate | Confidence |
|------|------------------|-----------------|------------|
| Task 1: Research & Schema | 2-3 days | **1-1.5 days** | Medium |
| Task 2: Infrastructure | 3-4 days | **2-3 days** | High |
| Task 3: CRUD Operations | 5-6 days | **4-6 days** | High |
| Task 4: Migration Tool | 3-4 days | **2-3 days** | Medium |
| Task 5: Simplification | 2-3 days | **1-1.5 days** | High |
| **Total Sequential** | **15-20 days** | **10-15 days** | - |

### Parallel Execution

| Task | Sequential | Parallel | Savings |
|------|-----------|----------|---------|
| Task 1: Research | 1-1.5 days | 1-1.5 days | 0 days |
| Task 2: Infrastructure | 2-3 days | **1.5-2 days** | 0.5-1 day |
| Task 3: CRUD | 4-6 days | **3-4 days** | 1-2 days |
| Task 4: Migration | 2-3 days | **1.5-2 days** | 0.5-1 day |
| Task 5: Simplification | 1-1.5 days | 1-1.5 days | 0 days |
| **Total Parallel** | **10-15 days** | **8-11 days** | **2-4 days** |

### Best Case Timeline (Aggressive Parallelization)

- **Week 1:** Tasks 1-2 (3-4 days)
- **Week 2:** Task 3 (3-4 days)
- **Week 3:** Tasks 4-5 (2-3 days)
- **Total:** **8-11 days**

### Realistic Timeline (Moderate Parallelization)

- **Week 1:** Tasks 1-2 (3.5-4.5 days)
- **Week 2:** Task 3 (4-5 days)
- **Week 3:** Tasks 4-5 (2.5-3.5 days)
- **Total:** **10-13 days**

## Task 3 Detailed Estimates

### By Subtask Group

| Group | Subtasks | Sequential | Parallel | Savings |
|-------|----------|-----------|----------|---------|
| Group 1: Core CRUD | 4 | 16-23h (2-3d) | 8-10h (1-1.25d) | 1-1.75d |
| Group 2: Queries | 6 | 14-20h (1.75-2.5d) | 8-10h (1-1.25d) | 0.75-1.25d |
| Group 3: Relationships | 4 | 13-19h (1.5-2.5d) | 8-10h (1-1.25d) | 0.5-1.25d |
| Group 4: History | 4 | 10-14h (1.25-1.75d) | 5-6h (0.5-0.75d) | 0.75-1d |
| Group 5: Utilities | 3 | 8-11h (1-1.5d) | 4-5h (0.5d) | 0.5-1d |
| Group 6: Testing | 5 | 27-36h (3.5-4.5d) | 27-36h* (overlaps) | - |
| **Total** | **26** | **80-106h (10-13d)** | **48-64h (6-8d)** | **4-5d** |

*Testing overlaps with implementation, so doesn't add to total time

## Key Insights

1. **Task 3 is the largest** - 60-70% of total migration time
2. **Parallelization saves 4-5 days** on Task 3 alone
3. **Testing can overlap** - Write tests as you implement
4. **Tool estimates were too optimistic** - Adjusted based on complexity
5. **Realistic timeline: 10-13 days** with moderate parallelization

## Recommendations

1. **Focus parallelization on Task 3** - Biggest time savings opportunity
2. **Start testing early** - Write tests alongside implementation
3. **Use TDD for complex functions** - T3.3, T3.5, T3.14
4. **Batch similar operations** - All "GetByX" queries together
5. **Plan for integration time** - Add 10-20% buffer for integration issues

## Risk Factors

1. **SQLite learning curve** - If team is new to SQLite in Go
2. **Transaction complexity** - UpdateTask and dependency validation
3. **Migration edge cases** - Real data may have unexpected formats
4. **Performance optimization** - May need query tuning
5. **Integration issues** - Components developed in parallel may not integrate smoothly

**Recommended Buffer:** Add 20% to estimates for unknowns = **12-16 days total**

