# Big O Time Complexity Analysis

**Generated:** 2026-01-09  
**Purpose:** Comprehensive analysis of time complexity for all algorithms in the exarp-go codebase

---

## Executive Summary

This document provides a complete Big O time complexity analysis for all algorithms in the codebase. The analysis identifies:

- **Optimal algorithms** (O(1), O(log n), O(n)) - No action needed
- **Acceptable algorithms** (O(n log n)) - Monitor performance
- **Optimization opportunities** (O(n²) or worse) - Consider improvements

### Complexity Distribution

| Complexity | Count | Status | Action |
|------------|-------|--------|--------|
| O(1) | 5+ | ✅ Excellent | None |
| O(n) | 20+ | ✅ Good | None |
| O(log n) | 0 | ✅ Good | None |
| O(n log n) | 3 | ⚠️ Acceptable | Monitor |
| O(V + E) | 6 | ✅ Good | None |
| O(n²) | 2 | ⚠️ Needs Review | Optimize if needed |
| O(V²) | 1 | ⚠️ Needs Review | Optimize for large graphs |

---

## Graph Algorithms (`internal/tools/graph_helpers.go`)

### BuildTaskGraph()
**Complexity:** O(V + E)  
**Where:** V = number of tasks, E = number of dependencies  
**Analysis:**
- First pass: O(V) - Adds all tasks as nodes
- Second pass: O(E) - Adds all dependency edges
- **Status:** ✅ Optimal for graph construction

```go
// Lines 93-117: Two-pass construction
func BuildTaskGraph(tasks []Todo2Task) (*TaskGraph, error) {
    // First pass: O(V)
    for _, task := range tasks {
        tg.AddTask(task)
    }
    // Second pass: O(E)
    for _, task := range tasks {
        for _, depID := range task.Dependencies {
            tg.AddDependency(depID, task.ID)
        }
    }
}
```

### DetectCycles()
**Complexity:** O(V + E) worst case  
**Where:** V = vertices, E = edges  
**Analysis:**
- DFS traversal visits each node once: O(V)
- Each edge examined once: O(E)
- **Status:** ✅ Optimal for cycle detection

```go
// Lines 131-191: DFS-based cycle detection
func DetectCycles(tg *TaskGraph) [][]string {
    // DFS visits each node once: O(V)
    // Each edge examined once: O(E)
    // Total: O(V + E)
}
```

### FindCriticalPath()
**Complexity:** O(V + E)  
**Where:** V = vertices, E = edges  
**Analysis:**
- Topological sort: O(V + E)
- Dynamic programming longest path: O(V + E)
- **Status:** ✅ Optimal for DAG longest path

```go
// Lines 264-326: Topological sort + DP
func FindCriticalPath(tg *TaskGraph) ([]string, error) {
    // Topo sort: O(V + E)
    sorted, err := TopoSortTasks(tg)
    // DP longest path: O(V + E)
    // Process nodes in topological order
}
```

### GetTaskLevels()
**Complexity:** O(V + E) optimized, O(V²) worst case  
**Where:** V = vertices, E = edges  
**Analysis:**
- **Optimized path (acyclic graphs):** O(V + E) - Uses topological sort
- **Fallback path (cyclic graphs):** O(V²) - Iterative until convergence
- **Status:** ⚠️ **Optimization Opportunity** - Fallback can be slow for large cyclic graphs

```go
// Lines 330-370: Optimized with topological sort
func GetTaskLevels(tg *TaskGraph) map[string]int {
    // Check for cycles first
    hasCycles, err := HasCycles(tg)
    if !hasCycles {
        // O(V + E) - Optimal path
        sortedNodes, err := topo.Sort(tg.Graph)
        // Single pass level calculation
    } else {
        // O(V²) - Fallback for cyclic graphs
        return getTaskLevelsIterative(tg)
    }
}
```

### getTaskLevelsIterative()
**Complexity:** O(V²) worst case  
**Where:** V = vertices  
**Analysis:**
- Iterative approach until convergence
- Worst case: V iterations, each visiting V nodes
- **Status:** ⚠️ **Optimization Opportunity** - Consider memoization or better algorithm

```go
// Lines 373-415: Iterative convergence
func getTaskLevelsIterative(tg *TaskGraph) map[string]int {
    // Worst case: V iterations
    for changed {
        // Each iteration: O(V) nodes
        // Total: O(V²)
    }
}
```

**Recommendation:** For cyclic graphs with many nodes, consider:
- Using strongly connected components (SCC) analysis
- Memoization to avoid redundant calculations
- Early termination if levels stabilize quickly

---

## Statistics Functions (`internal/tools/statistics.go`)

### Mean()
**Complexity:** O(n)  
**Where:** n = number of data points  
**Analysis:**
- Single pass through data array
- **Status:** ✅ Optimal

```go
// Lines 12-21: Single pass
func Mean(data []float64) float64 {
    // O(n) - One pass through data
    return stat.Mean(data, nil)
}
```

### Median()
**Complexity:** O(n log n)  
**Where:** n = number of data points  
**Analysis:**
- Requires sorting: O(n log n)
- Quantile calculation: O(1) after sort
- **Status:** ⚠️ Acceptable - Sorting is necessary for median

```go
// Lines 25-38: Requires sorting
func Median(data []float64) float64 {
    sorted := make([]float64, len(data))
    copy(sorted, data)
    sort.Float64s(sorted)  // O(n log n)
    return stat.Quantile(0.5, stat.LinInterp, sorted, nil)
}
```

**Recommendation:** 
- For frequently called median calculations, consider maintaining sorted data
- Use selection algorithm (O(n)) if only median is needed, not full sort

### StdDev()
**Complexity:** O(n)  
**Where:** n = number of data points  
**Analysis:**
- Single pass through data
- **Status:** ✅ Optimal

```go
// Lines 42-51: Single pass
func StdDev(data []float64) float64 {
    // O(n) - One pass calculation
    return stat.StdDev(data, nil)
}
```

### Quantile()
**Complexity:** O(n log n)  
**Where:** n = number of data points  
**Analysis:**
- Requires sorting: O(n log n)
- Quantile lookup: O(1) after sort
- **Status:** ⚠️ Acceptable - Sorting is necessary

```go
// Lines 57-73: Requires sorting
func Quantile(data []float64, p float64) float64 {
    sorted := make([]float64, len(data))
    copy(sorted, data)
    sort.Float64s(sorted)  // O(n log n)
    return stat.Quantile(p, stat.LinInterp, sorted, nil)
}
```

**Recommendation:**
- If multiple quantiles needed, sort once and reuse
- Consider maintaining sorted data structure if quantiles are frequently calculated

---

## Task Analysis Functions (`internal/tools/task_analysis_shared.go`)

### findDuplicateTasks()
**Complexity:** O(n²)  
**Where:** n = number of tasks  
**Analysis:**
- Nested loops comparing every task with every other task
- Similarity calculation for each pair
- **Status:** ⚠️ **Optimization Opportunity** - Consider hash-based or clustering approaches

```go
// Lines 314-334: Nested loops
func findDuplicateTasks(tasks []Todo2Task, threshold float64) [][]string {
    for i, task1 := range tasks {  // O(n)
        for j := i + 1; j < len(tasks); j++ {  // O(n)
            task2 := tasks[j]
            similarity := calculateSimilarity(task1, task2)  // O(1) or O(m)
            // ...
        }
    }
    // Total: O(n²)
}
```

**Recommendation:**
- For large task sets (>1000 tasks), consider:
  - Locality-sensitive hashing (LSH) for approximate duplicate detection
  - Clustering algorithms to group similar tasks first
  - Parallel processing for similarity calculations
  - Early termination if similarity is clearly below threshold

### mergeDuplicateTasks()
**Complexity:** O(n × m)  
**Where:** n = number of duplicate groups, m = average tasks per group  
**Analysis:**
- Loops through duplicate groups
- Merges tags and dependencies for each group
- **Status:** ✅ Acceptable - Linear in output size

```go
// Lines 383-443: Merge operations
func mergeDuplicateTasks(tasks []Todo2Task, duplicates [][]string) []Todo2Task {
    // O(n) groups × O(m) operations per group
    // = O(n × m) where m is average group size
}
```

### analyzeTags()
**Complexity:** O(n × t)  
**Where:** n = number of tasks, t = average tags per task  
**Analysis:**
- Single pass through tasks
- For each task, processes all tags
- **Status:** ✅ Optimal - Linear in input size

```go
// Lines 454-481: Single pass with tag processing
func analyzeTags(tasks []Todo2Task) TagAnalysis {
    for _, task := range tasks {  // O(n)
        for _, tag := range task.Tags {  // O(t) average
            // Process tag
        }
    }
    // Total: O(n × t)
}
```

### findMissingDependencies()
**Complexity:** O(n × d)  
**Where:** n = number of tasks, d = average dependencies per task  
**Analysis:**
- Builds task map: O(n)
- Checks each dependency: O(n × d)
- **Status:** ✅ Optimal - Linear in input size

```go
// Lines 579-600: Dependency validation
func findMissingDependencies(tasks []Todo2Task, tg *TaskGraph) []map[string]interface{} {
    // Build map: O(n)
    // Check dependencies: O(n × d)
    // Total: O(n × d)
}
```

---

## Database Operations (`internal/database/tasks.go`)

### CreateTask()
**Complexity:** O(t + d)  
**Where:** t = number of tags, d = number of dependencies  
**Analysis:**
- Single task insert: O(1)
- Tag inserts: O(t)
- Dependency inserts: O(d)
- **Status:** ✅ Optimal - Linear in task metadata size

```go
// Lines 25-100: Transaction with multiple inserts
func CreateTask(task *Todo2Task) error {
    // Insert task: O(1)
    // Insert tags: O(t)
    for _, tag := range task.Tags {
        tx.Exec(...)
    }
    // Insert dependencies: O(d)
    for _, depID := range task.Dependencies {
        tx.Exec(...)
    }
    // Total: O(t + d)
}
```

**Note:** SQL operations are typically O(1) with proper indexes, but we count the number of operations.

### Query Operations
**Complexity:** O(n) for full scans, O(log n) with indexes  
**Where:** n = number of rows  
**Analysis:**
- Full table scans: O(n)
- Indexed lookups: O(log n)
- **Status:** ✅ Acceptable - Depends on database indexing

**Recommendation:**
- Ensure proper indexes on frequently queried columns:
  - `tasks.status`
  - `tasks.priority`
  - `task_tags.tag`
  - `task_dependencies.depends_on_id`

---

## Summary of Optimization Opportunities

### High Priority

1. **`getTaskLevelsIterative()` - O(V²)**
   - **Impact:** Can be slow for large cyclic dependency graphs
   - **Recommendation:** Use SCC analysis or memoization
   - **Effort:** Medium

2. **`findDuplicateTasks()` - O(n²)**
   - **Impact:** Quadratic growth with task count
   - **Recommendation:** Consider LSH or clustering for large datasets
   - **Effort:** High

### Medium Priority

3. **`Median()` and `Quantile()` - O(n log n)**
   - **Impact:** Sorting overhead for large datasets
   - **Recommendation:** Maintain sorted data or use selection algorithms
   - **Effort:** Low-Medium

### Low Priority

4. **Database query optimization**
   - **Impact:** Ensure proper indexing for common queries
   - **Recommendation:** Review and add indexes as needed
   - **Effort:** Low

---

## Complexity Reference Guide

### Quick Identification Rules

| Pattern | Complexity | Example |
|---------|------------|---------|
| No loops | O(1) | Array access, map lookup |
| Single loop | O(n) | Iterating through array |
| Nested loops | O(n²) | Comparing all pairs |
| Sorting | O(n log n) | `sort.Float64s()` |
| Graph traversal | O(V + E) | DFS, BFS |
| Recursive (halving) | O(log n) | Binary search |
| Recursive (all subsets) | O(2ⁿ) | Subset generation |

### Best Practices

1. **Profile before optimizing** - Measure actual performance, not just complexity
2. **Consider input size** - O(n²) may be fine for small n (< 100)
3. **Cache expensive operations** - Reuse sorted data, memoize results
4. **Use appropriate data structures** - Maps for O(1) lookups, sets for membership
5. **Parallelize when possible** - Independent operations can run concurrently

---

## References

- [Big O Cheat Sheet](https://www.freecodecamp.org/news/big-o-cheat-sheet-time-complexity-chart/)
- [gonum/graph Documentation](https://pkg.go.dev/gonum.org/v1/gonum/graph)
- [gonum/stat Documentation](https://pkg.go.dev/gonum.org/v1/gonum/stat)

---

## Optimization Recommendations

### High Priority Optimizations

#### 1. Optimize `getTaskLevelsIterative()` - Reduce O(V²) to O(V + E)

**Current Implementation:**
- **Complexity:** O(V²) worst case
- **Location:** `internal/tools/graph_helpers.go:373-415`
- **Issue:** Iterative convergence can be slow for large cyclic graphs

**Recommended Approach:**
1. **Use Strongly Connected Components (SCC) Analysis**
   - Identify cyclic components separately
   - Apply topological sort within each SCC
   - Combine results efficiently

2. **Add Memoization**
   - Cache level calculations for nodes already processed
   - Avoid redundant iterations

3. **Early Termination**
   - Stop when levels stabilize (no changes in iteration)
   - Add maximum iteration limit to prevent infinite loops

**Expected Improvement:**
- **Best case:** O(V + E) - When cycles are small
- **Average case:** O(V + E × C) where C is average cycle size
- **Worst case:** Still O(V²) but with better constants

**Implementation Priority:** High  
**Effort:** Medium  
**Impact:** Significant for large dependency graphs (> 100 tasks)

---

#### 2. Optimize `findDuplicateTasks()` - Reduce O(n²) to O(n log n) or Better

**Current Implementation:**
- **Complexity:** O(n²)
- **Location:** `internal/tools/task_analysis_shared.go:314-334`
- **Issue:** Compares every task with every other task

**Recommended Approaches (Choose based on use case):**

**Option A: Early Termination (Quick Win)**
- Add similarity threshold check before full calculation
- Skip pairs that are clearly different
- **Complexity:** Still O(n²) but with better average case

**Option B: Locality-Sensitive Hashing (LSH)**
- Use LSH to group similar tasks first
- Only compare tasks within same hash buckets
- **Complexity:** O(n) average case, O(n²) worst case
- **Best for:** Large datasets (> 1000 tasks)

**Option C: Clustering Pre-processing**
- Use k-means or similar to group tasks
- Compare within clusters only
- **Complexity:** O(n log n) with clustering + O(k²) comparisons
- **Best for:** Very large datasets (> 5000 tasks)

**Option D: Parallel Processing**
- Process similarity calculations in parallel
- Use goroutines for independent comparisons
- **Complexity:** Still O(n²) but faster wall-clock time
- **Best for:** Multi-core systems

**Recommended Strategy:**
1. **Immediate:** Add early termination (Option A) - Low effort, good improvement
2. **Short-term:** Add parallel processing (Option D) - Medium effort, significant speedup
3. **Long-term:** Consider LSH (Option B) if dataset grows large - High effort, best scalability

**Implementation Priority:** High  
**Effort:** Low (Option A) to High (Option B)  
**Impact:** Significant for large task sets (> 500 tasks)

---

### Medium Priority Optimizations

#### 3. Optimize `Median()` and `Quantile()` - Reduce O(n log n) Sorting Overhead

**Current Implementation:**
- **Complexity:** O(n log n) due to sorting
- **Location:** `internal/tools/statistics.go:25-38, 57-73`

**Recommended Approaches:**

**Option A: Maintain Sorted Data**
- Keep data sorted if multiple quantiles needed
- Sort once, query multiple times
- **Complexity:** O(n log n) once + O(1) per query

**Option B: Use Selection Algorithm**
- For median only: Use quickselect (O(n) average)
- For single quantile: Use selection algorithm
- **Complexity:** O(n) average case

**Option C: Incremental Updates**
- If data changes frequently, use incremental sorting
- Maintain sorted order as data is added/removed
- **Complexity:** O(log n) per update

**Recommended Strategy:**
- **For single median/quantile:** Use selection algorithm (Option B)
- **For multiple quantiles:** Maintain sorted data (Option A)
- **For streaming data:** Use incremental approach (Option C)

**Implementation Priority:** Medium  
**Effort:** Low-Medium  
**Impact:** Moderate - Only matters for large datasets or frequent calls

---

### Low Priority Optimizations

#### 4. Database Query Optimization

**Current State:**
- Queries may use full table scans
- Index coverage needs review

**Recommended Actions:**
1. **Add Indexes:**
   ```sql
   CREATE INDEX idx_tasks_status ON tasks(status);
   CREATE INDEX idx_tasks_priority ON tasks(priority);
   CREATE INDEX idx_task_tags_tag ON task_tags(tag);
   CREATE INDEX idx_task_dependencies_depends_on ON task_dependencies(depends_on_id);
   ```

2. **Query Analysis:**
   - Use `EXPLAIN QUERY PLAN` to analyze query performance
   - Identify slow queries
   - Optimize based on actual usage patterns

**Implementation Priority:** Low  
**Effort:** Low  
**Impact:** Moderate - Improves query performance

---

## Implementation Roadmap

### Phase 1: Quick Wins (1-2 days)
1. ✅ Add early termination to `findDuplicateTasks()`
2. ✅ Add maximum iteration limit to `getTaskLevelsIterative()`
3. ✅ Add database indexes

### Phase 2: Medium Improvements (3-5 days)
1. ⬜ Add parallel processing to `findDuplicateTasks()`
2. ⬜ Use selection algorithm for single median/quantile
3. ⬜ Add memoization to `getTaskLevelsIterative()`

### Phase 3: Advanced Optimizations (1-2 weeks)
1. ⬜ Implement SCC analysis for cyclic graphs
2. ⬜ Consider LSH for duplicate detection (if needed)
3. ⬜ Implement incremental sorting for statistics

---

## Performance Testing Recommendations

Before and after optimization, measure:

1. **Benchmark Functions:**
   - `getTaskLevelsIterative()` with various graph sizes (10, 100, 1000 tasks)
   - `findDuplicateTasks()` with various task counts (100, 500, 1000, 5000)
   - `Median()` and `Quantile()` with various data sizes

2. **Metrics to Track:**
   - Execution time (wall-clock and CPU)
   - Memory usage
   - Scalability (how performance degrades with size)

3. **Test Cases:**
   - Small datasets (< 100 items) - Should be fast
   - Medium datasets (100-1000 items) - Should be acceptable
   - Large datasets (> 1000 items) - Should complete in reasonable time

---

**Last Updated:** 2026-01-09  
**Next Review:** When adding new algorithms or when performance issues are identified

