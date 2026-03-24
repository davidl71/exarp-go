# Performance Optimization Research

## Status: 2026-03-23

## Executive Summary

Three concrete optimization opportunities identified with ~30% scorecard walk time reduction potential.

---

## 1. scorecard_multilang.go: Six Separate filepath.Walk Calls

**Current State:**
- `countCppFilesIn()` - walks dir counting `.cpp,.cc,.cxx,.c,.h,.hpp,.hxx`
- `countPythonFilesIn()` - walks dir counting `.py`
- `countRustFilesIn()` - walks dir counting `.rs`
- `countGoFilesIn()` - walks dir counting `.go`
- `countTypeScriptFilesIn()` - walks dir counting `.ts,.tsx,.js,.jsx`
- `countSwiftFilesIn()` - walks dir counting `.swift`

Each function does `filepath.Walk(dir, callback)` with identical `walkSkipDir` filtering.

**Reference Implementation:**
`sscorecard_go_checks.go:212` - `collectAllFileStats()` consolidates 5 walks into 1.

**Proposed Fix:**
Single `collectLangFileStats(dir string) map[string]int` that counts all langs in one walk:

```go
func collectLangFileStats(dir string) map[string]int {
    stats := map[string]int{"cpp": 0, "python": 0, "rust": 0, "go": 0, "typescript": 0, "swift": 0}
    filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil { return nil }
        if err := walkSkipDir(path, info, dir); err != nil { return err }
        if info.IsDir() { return nil }
        ext := strings.ToLower(filepath.Ext(path))
        // ... switch on ext, increment counters
        return nil
    })
    return stats
}
```

**Effort:** Medium | **Impact:** High

---

## 2. testing.go: Three Separate filepath.Walk for Validation

**Current State:**
- `validateGoTests()` - walks `testPath` collecting `*_test.go`
- `validatePyTests()` - walks `searchPath` collecting `test_*.py` and `*_test.py`
- `validateCargoTests()` - walks `searchPath` collecting `*.rs`

Each with its own `filepath.Walk`.

**Proposed Fix:**
Single `collectTestFiles(projectRoot, testPath) map[string][]string` returning per-extension file lists in one walk.

**Effort:** Low | **Impact:** Medium

---

## 3. task_workflow_maintenance.go: N+1 Delete Pattern

**Current State (line 516-519):**
```go
for _, task := range tasksToRemove {
    if err := database.DeleteTask(context.Background(), task.ID); err == nil {
        removedIDs = append(removedIDs, task.ID)
    }
}
```

**Proposed Fix:**
Batch DELETE with transaction:
```go
if len(taskIDs) > 0 {
    placeholders := make([]string, len(taskIDs))
    args := make([]interface{}, len(taskIDs))
    for i, id := range taskIDs {
        placeholders[i] = "?"
        args[i] = id
    }
    _, err = tx.ExecContext(ctx, `DELETE FROM tasks WHERE id IN (`+strings.Join(placeholders,",")+`)`, args...)
}
```

**Effort:** Low | **Impact:** Medium

---

## Verification

After each optimization:
```bash
make test
make lint
```

---

## Priority Order

1. **scorecard_multilang.go** - Highest impact, established pattern to follow
2. **testing.go** - Lower effort, moderate impact
3. **task_workflow_maintenance.go** - Lower impact, simple fix
