# Migration Plan: Python Statistics to gonum/stat

**Date:** 2026-01-09  
**Status:** 📋 Planning  
**Target:** Replace Python `statistics` module with `gonum.org/v1/gonum/stat`

---

## Executive Summary

This document outlines the step-by-step migration plan to replace Python's `statistics` module with gonum/stat in exarp-go estimation tools. The migration will reduce Python bridge dependencies, improve performance, and enable native Go statistical calculations.

**Scope:**
- 3 Python files requiring migration
- 17 statistics function calls
- Maintain backward compatibility during transition
- Zero-downtime migration strategy

**Benefits:**
- ✅ Native Go implementation (no Python bridge)
- ✅ Improved performance (compiled Go vs interpreted Python)
- ✅ Better percentile accuracy (interpolation vs manual calculation)
- ✅ Reduced dependencies

---

## Current State Analysis

### Files Requiring Migration

1. **`project_management_automation/tools/task_duration_estimator.py`**
   - 6 statistics calls
   - `get_statistics()` function (main endpoint)
   - Error analysis calculations

2. **`project_management_automation/tools/estimation_learner.py`**
   - 10 statistics calls
   - Multiple analysis functions (accuracy, tags, priority, patterns)

3. **`project_management_automation/tools/mlx_task_estimator.py`**
   - 1 statistics call
   - Adjustment averaging

### Statistics Functions Used

| Python Function | Count | Usage Context |
|----------------|-------|---------------|
| `statistics.mean()` | 12 | Error analysis, accuracy metrics, adjustments |
| `statistics.median()` | 2 | Descriptive statistics |
| `statistics.stdev()` | 2 | Variance analysis, pattern detection |
| Manual percentiles | 3 | p25, p75, p90 calculations |
| `statistics.StatisticsError` | 1 | Exception handling |

---

## Function Mapping Table

| Python Function | Gonum/Stat Equivalent | Migration Pattern |
|----------------|----------------------|-------------------|
| `statistics.mean(data)` | `stat.Mean(data, nil)` | Direct replacement |
| `statistics.median(data)` | `stat.Quantile(data, nil, 0.5)` | Use Quantile with p=0.5 |
| `statistics.stdev(data)` | `stat.StdDev(data, nil)` | Direct replacement (sample std dev) |
| `sorted_data[n//4]` (p25) | `stat.Quantile(data, nil, 0.25)` | More accurate interpolation |
| `sorted_data[3*n//4]` (p75) | `stat.Quantile(data, nil, 0.75)` | More accurate interpolation |
| `sorted_data[9*n//10]` (p90) | `stat.Quantile(data, nil, 0.90)` | More accurate interpolation |
| `statistics.StatisticsError` | `math.IsNaN()` check | Different error pattern |

---

## Migration Strategy

### Phase 1: Create Go Helper Functions

**Goal:** Create reusable Go helper functions that match Python API patterns

**Location:** `internal/tools/statistics.go` (new file)

**Helper Functions:**

```go
package tools

import (
    "math"
    "gonum.org/v1/gonum/stat"
)

// Mean calculates the arithmetic mean of a slice of floats
// Returns 0.0 for empty slice (matches Python behavior)
func Mean(data []float64) float64 {
    if len(data) == 0 {
        return 0.0
    }
    return stat.Mean(data, nil)
}

// Median calculates the median (50th percentile) of a slice of floats
// Returns 0.0 for empty slice
func Median(data []float64) float64 {
    if len(data) == 0 {
        return 0.0
    }
    return stat.Quantile(data, nil, 0.5)
}

// StdDev calculates the sample standard deviation
// Returns 0.0 for empty or single-element slices (matches Python behavior)
func StdDev(data []float64) float64 {
    if len(data) <= 1 {
        return 0.0
    }
    result := stat.StdDev(data, nil)
    if math.IsNaN(result) {
        return 0.0
    }
    return result
}

// Quantile calculates the specified quantile (0.0 to 1.0)
// Returns 0.0 for empty slice
func Quantile(data []float64, p float64) float64 {
    if len(data) == 0 {
        return 0.0
    }
    if p < 0.0 || p > 1.0 {
        return 0.0 // Invalid quantile
    }
    result := stat.Quantile(data, nil, p)
    if math.IsNaN(result) {
        return 0.0
    }
    return result
}

// Round rounds a float64 to specified decimal places
func Round(value float64, decimals int) float64 {
    multiplier := math.Pow(10, float64(decimals))
    return math.Round(value * multiplier) / multiplier
}
```

**Testing:** Unit tests for each helper function with edge cases

---

### Phase 2: Create Python Bridge Wrapper

**Goal:** Create Python-accessible wrapper that uses Go statistics

**Location:** `bridge/statistics_bridge.py` (new file)

**Implementation:**

```python
"""
Statistics bridge - uses Go gonum/stat via subprocess or direct import
Maintains Python API compatibility during migration
"""

import json
import subprocess
from typing import List, Optional

def mean(data: List[float]) -> float:
    """Calculate mean using Go gonum/stat."""
    if not data:
        return 0.0
    # Call Go helper function via bridge
    # Implementation depends on bridge architecture
    pass

def median(data: List[float]) -> float:
    """Calculate median using Go gonum/stat."""
    if not data:
        return 0.0
    # Call Go helper function via bridge
    pass

def stdev(data: List[float]) -> float:
    """Calculate standard deviation using Go gonum/stat."""
    if len(data) <= 1:
        return 0.0
    # Call Go helper function via bridge
    pass

def quantile(data: List[float], p: float) -> float:
    """Calculate quantile using Go gonum/stat."""
    if not data:
        return 0.0
    # Call Go helper function via bridge
    pass
```

**Note:** This phase may be skipped if migrating directly to Go implementations

---

### Phase 3: Migrate task_duration_estimator.py

**File:** `project_management_automation/tools/task_duration_estimator.py`

**Changes Required:**

1. **Replace `get_statistics()` function:**
   - Replace `statistics.mean()` → Go helper or bridge call
   - Replace `statistics.median()` → Go helper or bridge call
   - Replace `statistics.stdev()` → Go helper or bridge call
   - Replace manual percentiles → `stat.Quantile()` calls
   - Replace exception handling → NaN checks

**Before:**
```python
'mean': round(statistics.mean(actual_hours), 2),
'median': round(statistics.median(actual_hours), 2),
'stdev': round(statistics.stdev(actual_hours), 2) if len(actual_hours) > 1 else 0.0,
```

**After (using bridge):**
```python
from bridge.statistics_bridge import mean, median, stdev, quantile

'mean': round(mean(actual_hours), 2),
'median': round(median(actual_hours), 2),
'stdev': round(stdev(actual_hours), 2),
```

**After (direct Go - if migrating to Go implementation):**
```go
// In Go implementation
stats := map[string]interface{}{
    "mean": Round(Mean(actualHours), 2),
    "median": Round(Median(actualHours), 2),
    "stdev": Round(StdDev(actualHours), 2),
}
```

2. **Replace error analysis calculations:**
   - Lines 370-371: Replace `statistics.mean()` calls

3. **Replace exception handling:**
   - Line 376: Replace `statistics.StatisticsError` with validation checks

**Testing:**
- Unit tests comparing Python vs Go results
- Integration tests with real historical data
- Edge case tests (empty data, single values)

---

### Phase 4: Migrate estimation_learner.py

**File:** `project_management_automation/tools/estimation_learner.py`

**Changes Required:**

1. **`analyze_estimation_accuracy()` function:**
   - Replace 5 `statistics.mean()` calls (lines 111-115)
   - Replace 1 `statistics.median()` call (line 112)

2. **`_analyze_by_tag()` function:**
   - Replace 3 `statistics.mean()` calls (lines 174-176)

3. **`_analyze_by_priority()` function:**
   - Replace 3 `statistics.mean()` calls (lines 198-200)

4. **`_identify_patterns()` function:**
   - Replace 1 `statistics.mean()` call (line 222)
   - Replace 1 `statistics.stdev()` call (line 223)

**Testing:**
- Verify grouped analysis results match Python
- Test minimum sample size requirements
- Validate pattern detection accuracy

---

### Phase 5: Migrate mlx_task_estimator.py

**File:** `project_management_automation/tools/mlx_task_estimator.py`

**Changes Required:**

1. **`_apply_learned_adjustments()` function:**
   - Replace 1 `statistics.mean()` call (line 366)

**Testing:**
- Verify adjustment calculations match Python
- Test with various adjustment factor lists

---

## Code Examples

### Example 1: Basic Statistics Migration

**Python (Before):**
```python
import statistics

actual_hours = [2.5, 3.0, 4.5, 2.0, 3.5]
mean_hours = round(statistics.mean(actual_hours), 2)
median_hours = round(statistics.median(actual_hours), 2)
stdev_hours = round(statistics.stdev(actual_hours), 2) if len(actual_hours) > 1 else 0.0
```

**Go (After):**
```go
import (
    "gonum.org/v1/gonum/stat"
    "math"
)

actualHours := []float64{2.5, 3.0, 4.5, 2.0, 3.5}
meanHours := Round(Mean(actualHours), 2)
medianHours := Round(Median(actualHours), 2)
stdevHours := Round(StdDev(actualHours), 2)
```

### Example 2: Percentile Migration

**Python (Before):**
```python
sorted_hours = sorted(actual_hours)
p25 = round(sorted_hours[len(sorted_hours) // 4], 2)
p75 = round(sorted_hours[3 * len(sorted_hours) // 4], 2)
p90 = round(sorted_hours[9 * len(sorted_hours) // 10], 2)
```

**Go (After):**
```go
p25 := Round(Quantile(actualHours, 0.25), 2)
p75 := Round(Quantile(actualHours, 0.75), 2)
p90 := Round(Quantile(actualHours, 0.90), 2)
```

### Example 3: Error Handling Migration

**Python (Before):**
```python
try:
    stats = {
        'mean': round(statistics.mean(actual_hours), 2),
        # ...
    }
except statistics.StatisticsError as e:
    logger.warning(f"Statistics calculation failed: {e}")
    return json.dumps({'error': str(e)})
```

**Go (After):**
```go
if len(actualHours) == 0 {
    return json.Marshal(map[string]interface{}{
        "error": "no data available",
    })
}

mean := Mean(actualHours)
if math.IsNaN(mean) {
    return json.Marshal(map[string]interface{}{
        "error": "statistics calculation failed",
    })
}

stats := map[string]interface{}{
    "mean": Round(mean, 2),
    // ...
}
```

---

## Testing Strategy

### Unit Tests

**Test File:** `internal/tools/statistics_test.go`

**Test Cases:**

1. **Mean Tests:**
   - Empty slice → 0.0
   - Single value → value itself
   - Multiple values → correct mean
   - Negative values → correct mean
   - Large values → correct mean

2. **Median Tests:**
   - Empty slice → 0.0
   - Single value → value itself
   - Even number of values → average of middle two
   - Odd number of values → middle value
   - Unsorted data → correct median

3. **StdDev Tests:**
   - Empty slice → 0.0
   - Single value → 0.0
   - Two values → correct std dev
   - Multiple values → correct std dev
   - Constant values → 0.0

4. **Quantile Tests:**
   - Empty slice → 0.0
   - p25, p50, p75, p90 → correct values
   - Invalid p (< 0 or > 1) → 0.0
   - Edge cases (p=0.0, p=1.0)

### Integration Tests

**Test File:** `internal/tools/estimation_test.go`

**Test Cases:**

1. **Python vs Go Comparison:**
   - Run same data through Python and Go
   - Compare results (allow small floating-point differences)
   - Verify results match within tolerance (0.01)

2. **Real Historical Data:**
   - Load actual `.todo2/state.todo2.json`
   - Compare Python vs Go statistics output
   - Verify JSON structure matches

3. **Edge Cases:**
   - Empty historical data
   - Single task
   - All tasks with same duration
   - Extreme values

### Performance Tests

**Benchmark File:** `internal/tools/statistics_bench_test.go`

**Benchmarks:**

1. Mean calculation (1000 values)
2. Median calculation (1000 values)
3. StdDev calculation (1000 values)
4. Quantile calculation (1000 values)
5. Full statistics calculation (realistic data size)

**Expected Results:**
- Go implementation should be 2-5x faster than Python
- Memory usage should be lower

---

## Rollback Plan

### If Migration Fails

1. **Immediate Rollback:**
   - Revert Python file changes
   - Restore `import statistics` statements
   - Verify all tests pass

2. **Partial Rollback:**
   - Keep Go helper functions (useful for future)
   - Revert Python file changes
   - Document issues encountered

3. **Gradual Rollback:**
   - Migrate one file at a time
   - Test thoroughly before next file
   - Rollback individual files if issues found

### Rollback Triggers

- Test failures in CI/CD
- Performance degradation (>10% slower)
- Incorrect statistical results
- Production errors

---

## Performance Considerations

### Expected Improvements

1. **Execution Speed:**
   - Go compiled code: 2-5x faster than Python
   - No Python interpreter overhead
   - Better memory efficiency

2. **Memory Usage:**
   - Lower memory footprint (Go vs Python)
   - No Python object overhead
   - Efficient slice operations

3. **Startup Time:**
   - Faster module import (compiled vs interpreted)
   - No Python bridge overhead (if fully migrated)

### Performance Monitoring

1. **Before Migration:**
   - Baseline performance metrics
   - Average execution time
   - Memory usage

2. **During Migration:**
   - Compare Python vs Go results
   - Monitor performance improvements
   - Track any regressions

3. **After Migration:**
   - Verify performance improvements
   - Monitor production metrics
   - Optimize if needed

---

## Timeline and Risk Assessment

### Estimated Timeline

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 1: Go Helpers | 2-3 hours | None |
| Phase 2: Python Bridge (optional) | 1-2 hours | Phase 1 |
| Phase 3: task_duration_estimator | 2-3 hours | Phase 1 |
| Phase 4: estimation_learner | 3-4 hours | Phase 3 |
| Phase 5: mlx_task_estimator | 1 hour | Phase 4 |
| Testing & Validation | 4-6 hours | All phases |
| **Total** | **13-19 hours** | |

### Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| API incompatibility | Low | Medium | Thorough testing, helper functions |
| Precision differences | Medium | Low | Tolerance checks, rounding |
| Performance regression | Low | Medium | Benchmarking, monitoring |
| Edge case bugs | Medium | Medium | Comprehensive test coverage |
| Migration complexity | Low | High | Phased approach, rollback plan |

### Success Criteria

1. ✅ All tests pass (unit, integration, performance)
2. ✅ Results match Python within tolerance (0.01)
3. ✅ Performance improved or maintained
4. ✅ No production errors
5. ✅ Code review approved

---

## Migration Checklist

### Pre-Migration

- [ ] Review and approve migration plan
- [ ] Set up test environment
- [ ] Create baseline performance metrics
- [ ] Backup current code

### Phase 1: Go Helpers

- [ ] Create `internal/tools/statistics.go`
- [ ] Implement helper functions
- [ ] Write unit tests
- [ ] Verify tests pass

### Phase 2: Python Bridge (if needed)

- [ ] Create `bridge/statistics_bridge.py`
- [ ] Implement bridge functions
- [ ] Test bridge integration
- [ ] Document bridge usage

### Phase 3: task_duration_estimator

- [ ] Migrate `get_statistics()` function
- [ ] Replace error analysis calculations
- [ ] Update exception handling
- [ ] Run integration tests
- [ ] Compare Python vs Go results

### Phase 4: estimation_learner

- [ ] Migrate `analyze_estimation_accuracy()`
- [ ] Migrate `_analyze_by_tag()`
- [ ] Migrate `_analyze_by_priority()`
- [ ] Migrate `_identify_patterns()`
- [ ] Run integration tests
- [ ] Verify grouped analysis results

### Phase 5: mlx_task_estimator

- [ ] Migrate `_apply_learned_adjustments()`
- [ ] Run integration tests
- [ ] Verify adjustment calculations

### Post-Migration

- [ ] All tests pass
- [ ] Performance benchmarks meet targets
- [ ] Code review completed
- [ ] Documentation updated
- [ ] Production deployment
- [ ] Monitor for issues

---

## Conclusion

This migration plan provides a comprehensive roadmap for replacing Python statistics with gonum/stat. The phased approach minimizes risk, maintains backward compatibility, and enables gradual migration with testing at each step.

**Key Benefits:**
- Native Go implementation
- Improved performance
- Better percentile accuracy
- Reduced Python dependencies

**Next Steps:**
1. Review and approve this plan
2. Begin Phase 1 (Go helper functions)
3. Proceed with phased migration
4. Monitor and validate results

---

**Document Version:** 1.0  
**Last Updated:** 2026-01-09  
**Author:** AI Assistant  
**Status:** Ready for Review

