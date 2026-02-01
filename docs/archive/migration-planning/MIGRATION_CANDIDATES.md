# Native Go Migration Candidates

**Date:** 2026-01-07  
**Analysis:** Prompts and Resources Migration Assessment

---

## Executive Summary

**Excellent Candidates for Native Go Migration:**
- ✅ **Prompts (17 total)** - **EASIEST** - Just static string templates
- ⚠️ **Resources (6 total)** - **MIXED** - Some easy, some complex

**Current Status:**
- Prompts: 0/17 native (0%) | 17/17 Python bridge (100%)
- Resources: 0/6 native (0%) | 6/6 Python bridge (100%)

---

## 1. Prompts - ⭐ EXCELLENT CANDIDATE

### Current Implementation

**What Python Does:**
```python
# bridge/get_prompt.py
# Just imports string constants and returns them
prompt_map = {
    "align": TASK_ALIGNMENT_ANALYSIS,  # String constant
    "discover": TASK_DISCOVERY,        # String constant
    # ... 15 more
}
return prompt_map.get(prompt_name)  # Just returns the string
```

**Complexity:** ⭐ **TRIVIAL** - Just string retrieval

### Migration Assessment

**Effort:** 🟢 **VERY LOW** (2-4 hours)
- Copy 17 prompt strings from Python to Go
- Create a simple map in Go
- Remove Python bridge dependency

**Benefits:**
- ✅ **Eliminates Python dependency** for prompts
- ✅ **Faster** - No subprocess overhead
- ✅ **Simpler** - No JSON parsing
- ✅ **More reliable** - No Python import errors
- ✅ **Better error handling** - Direct Go errors

**Implementation:**
```go
// internal/prompts/templates.go
package prompts

var promptTemplates = map[string]string{
    "align": `Analyze Todo2 task alignment with project goals.
    
This prompt will:
1. Evaluate task alignment with project objectives
2. Identify misaligned or out-of-scope tasks
...`,
    "discover": `Discover tasks from TODO comments...`,
    // ... 15 more
}

func GetPromptTemplate(name string) (string, error) {
    template, ok := promptTemplates[name]
    if !ok {
        return "", fmt.Errorf("unknown prompt: %s", name)
    }
    return template, nil
}
```

**Migration Steps:**
1. Extract all 17 prompt strings from Python
2. Create `internal/prompts/templates.go` with string map
3. Update `createPromptHandler()` to use Go map instead of Python bridge
4. Remove `bridge.GetPythonPrompt()` calls
5. Test all prompts
6. Delete `bridge/get_prompt.py` (optional cleanup)

**Risk:** 🟢 **VERY LOW** - Just string constants, no logic

---

## 2. Resources - ⚠️ MIXED CANDIDATES

### 2.1 Scorecard Resource - ⭐ GOOD CANDIDATE

**Current:** `stdio://scorecard` → Python `generate_project_scorecard()`

**Assessment:**
- ✅ **Native Go implementation EXISTS** (`internal/tools/scorecard_go.go`)
- ✅ Already used by `report` tool for Go projects
- ⚠️ Python version has more features (non-Go projects, more metrics)

**Effort:** 🟡 **LOW-MEDIUM** (2-3 hours)
- Use existing `GenerateGoScorecard()` function
- Add fallback to Python for non-Go projects (optional)
- Or enhance Go version to match Python features

**Implementation:**
```go
func handleScorecard(ctx context.Context, uri string) ([]byte, string, error) {
    // Check if Go project
    if isGoProject() {
        scorecard, err := scorecard.GenerateGoScorecard(projectRoot)
        if err != nil {
            return nil, "", err
        }
        json, _ := json.Marshal(scorecard)
        return json, "application/json", nil
    }
    
    // Fallback to Python for non-Go projects
    return bridge.ExecutePythonResource(ctx, uri)
}
```

**Risk:** 🟡 **LOW** - Go version already exists and works

---

### 2.2 Memory Resources - ⚠️ COMPLEX

**Current:** 5 memory resources use Python `get_memories_*_resource()` functions

**What Python Does:**
```python
# project_management_automation/resources/memories.py
def get_memories_resource(limit: int = 50) -> str:
    memories = _load_all_memories()[:limit]  # Loads JSON files from .exarp/memories/
    # Calculates statistics
    # Returns JSON string
```

**Complexity:** 🟡 **MEDIUM** - File I/O, JSON parsing, filtering, statistics

**Assessment:**
- ⚠️ **File-based storage** - Reads from `.exarp/memories/*.json`
- ⚠️ **JSON parsing** - Go can handle this easily
- ⚠️ **Filtering logic** - Category, task, date filtering
- ⚠️ **Statistics calculation** - Category counts, totals

**Effort:** 🟡 **MEDIUM** (4-6 hours per resource)
- Implement file reading from `.exarp/memories/`
- Implement JSON parsing
- Implement filtering logic
- Implement statistics calculation

**Resources to Migrate:**
1. `stdio://memories` - All memories (easiest)
2. `stdio://memories/category/{category}` - Filter by category
3. `stdio://memories/task/{task_id}` - Filter by task
4. `stdio://memories/recent` - Filter by date (last 24h)
5. `stdio://memories/session/{date}` - Filter by session date

**Implementation Pattern:**
```go
// internal/resources/memories.go
package resources

type Memory struct {
    ID       string                 `json:"id"`
    Title    string                 `json:"title"`
    Content  string                 `json:"content"`
    Category string                 `json:"category"`
    TaskID   string                 `json:"task_id,omitempty"`
    Date     string                 `json:"date"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func loadAllMemories(projectRoot string) ([]Memory, error) {
    memoriesDir := filepath.Join(projectRoot, ".exarp", "memories")
    // Read all .json files
    // Parse JSON
    // Return slice
}

func getMemoriesResource(limit int) ([]byte, error) {
    memories, err := loadAllMemories(getProjectRoot())
    if err != nil {
        return nil, err
    }
    
    // Calculate statistics
    categories := make(map[string]int)
    for _, m := range memories {
        categories[m.Category]++
    }
    
    result := map[string]interface{}{
        "memories": memories[:limit],
        "total": len(memories),
        "returned": min(limit, len(memories)),
        "categories": categories,
    }
    
    return json.Marshal(result)
}
```

**Risk:** 🟡 **MEDIUM** - Need to match Python behavior exactly

**Recommendation:**
- ✅ **Migrate `stdio://memories` first** (simplest)
- ⚠️ **Migrate others incrementally** (one at a time)
- ✅ **Keep Python as fallback** during migration

---

## 3. Migration Priority Matrix

| Component | Effort | Benefit | Risk | Priority | Recommendation |
|-----------|--------|---------|------|----------|----------------|
| **Prompts (all 17)** | 🟢 Very Low | 🟢 High | 🟢 Very Low | ⭐⭐⭐ **HIGHEST** | **Migrate immediately** |
| **Scorecard Resource** | 🟡 Low | 🟢 High | 🟢 Low | ⭐⭐ **HIGH** | **Migrate soon** |
| **Memories Resource (all)** | 🟡 Medium | 🟡 Medium | 🟡 Medium | ⭐ **MEDIUM** | **Migrate incrementally** |

---

## 4. Recommended Migration Order

### Phase 1: Prompts (Easiest Win) ⭐
**Effort:** 2-4 hours  
**Impact:** Eliminates Python dependency for 17 prompts

1. Extract all prompt strings
2. Create `internal/prompts/templates.go`
3. Update handlers to use Go map
4. Test all prompts
5. Remove Python bridge dependency

### Phase 2: Scorecard Resource
**Effort:** 2-3 hours  
**Impact:** Faster scorecard generation for Go projects

1. Use existing `GenerateGoScorecard()`
2. Update `handleScorecard()` to use Go version
3. Add Python fallback for non-Go projects (optional)
4. Test

### Phase 3: Memory Resources (Incremental)
**Effort:** 4-6 hours per resource  
**Impact:** Faster memory access, no Python dependency

1. Start with `stdio://memories` (simplest)
2. Add `stdio://memories/category/{category}`
3. Add `stdio://memories/task/{task_id}`
4. Add `stdio://memories/recent`
5. Add `stdio://memories/session/{date}`

---

## 5. Implementation Details

### 5.1 Prompts Migration

**File Structure:**
```
internal/prompts/
  ├── registry.go      (existing - update handlers)
  ├── templates.go     (new - prompt string map)
  └── templates_test.go (new - test all prompts)
```

**Key Changes:**
```go
// internal/prompts/registry.go
func createPromptHandler(promptName string) framework.PromptHandler {
    return func(ctx context.Context, args map[string]interface{}) (string, error) {
        // OLD: promptText, err := bridge.GetPythonPrompt(ctx, promptName)
        // NEW:
        promptText, err := GetPromptTemplate(promptName)
        if err != nil {
            return "", fmt.Errorf("failed to get prompt %s: %w", promptName, err)
        }
        
        // Add template substitution (from review)
        if len(args) > 0 {
            promptText = substituteTemplate(promptText, args)
        }
        
        return promptText, nil
    }
}
```

### 5.2 Scorecard Resource Migration

**Key Changes:**
```go
// internal/resources/handlers.go
func handleScorecard(ctx context.Context, uri string) ([]byte, string, error) {
    projectRoot := getProjectRoot()
    
    // Check if Go project
    if isGoProject(projectRoot) {
        scorecard, err := scorecard.GenerateGoScorecard(projectRoot)
        if err != nil {
            // Fallback to Python on error
            return bridge.ExecutePythonResource(ctx, uri)
        }
        
        jsonData, err := json.Marshal(scorecard)
        if err != nil {
            return nil, "", fmt.Errorf("failed to marshal scorecard: %w", err)
        }
        
        return jsonData, "application/json", nil
    }
    
    // Non-Go project - use Python
    return bridge.ExecutePythonResource(ctx, uri)
}
```

### 5.3 Memory Resources Migration

**File Structure:**
```
internal/resources/
  ├── handlers.go      (existing - update handlers)
  ├── memories.go      (new - memory loading/filtering)
  └── memories_test.go (new - test memory operations)
```

**Key Functions:**
```go
// internal/resources/memories.go
func loadAllMemories(projectRoot string) ([]Memory, error)
func filterByCategory(memories []Memory, category string) []Memory
func filterByTask(memories []Memory, taskID string) []Memory
func filterRecent(memories []Memory, hours int) []Memory
func filterBySession(memories []Memory, date string) []Memory
func calculateStatistics(memories []Memory) map[string]interface{}
```

---

## 6. Testing Strategy

### Prompts
- ✅ Test all 17 prompts return correct strings
- ✅ Test template substitution with various args
- ✅ Test error handling for unknown prompts
- ✅ Test with empty args map

### Scorecard
- ✅ Test Go project scorecard generation
- ✅ Test non-Go project fallback to Python
- ✅ Test error handling
- ✅ Compare Go vs Python output (should match)

### Memory Resources
- ✅ Test loading all memories
- ✅ Test filtering by category
- ✅ Test filtering by task
- ✅ Test filtering by date
- ✅ Test statistics calculation
- ✅ Test with empty memories directory
- ✅ Test with invalid JSON files

---

## 7. Benefits Summary

### Prompts Migration
- ✅ **Eliminates Python dependency** (17 prompts)
- ✅ **Faster** - No subprocess overhead (~10ms → <1ms)
- ✅ **Simpler** - Direct Go map lookup
- ✅ **More reliable** - No Python import errors
- ✅ **Enables template substitution** (from review)

### Scorecard Migration
- ✅ **Faster** for Go projects
- ✅ **Reduces Python dependency**
- ✅ **Uses existing Go implementation**

### Memory Resources Migration
- ✅ **Faster** memory access
- ✅ **Eliminates Python dependency** (5 resources)
- ✅ **Better error handling** in Go
- ✅ **Type safety** with Go structs

---

## 8. Risks & Mitigation

### Prompts
- **Risk:** Copy-paste errors when extracting strings
- **Mitigation:** Automated extraction script, comprehensive tests

### Scorecard
- **Risk:** Go version missing features from Python version
- **Mitigation:** Feature comparison, keep Python fallback

### Memory Resources
- **Risk:** Behavior differences from Python version
- **Mitigation:** Side-by-side testing, keep Python fallback during migration

---

## 9. Estimated Timeline

| Phase | Component | Effort | Total |
|-------|-----------|--------|-------|
| **Phase 1** | Prompts (17) | 2-4 hours | 2-4 hours |
| **Phase 2** | Scorecard Resource | 2-3 hours | 4-7 hours |
| **Phase 3** | Memory Resources (5) | 4-6 hours each | 24-36 hours |
| **TOTAL** | | | **30-47 hours** |

**Recommended Approach:**
- Start with Phase 1 (quick win)
- Then Phase 2 (medium effort)
- Phase 3 can be done incrementally over time

---

## 10. Conclusion

**Top Priority:**
1. ⭐⭐⭐ **Prompts** - Easiest, highest impact, lowest risk
2. ⭐⭐ **Scorecard** - Good candidate, existing Go code
3. ⭐ **Memory Resources** - Medium complexity, can be incremental

**Next Steps:**
1. Create migration tasks for Phase 1 (Prompts)
2. Extract prompt strings from Python
3. Implement Go templates
4. Test and verify
5. Remove Python bridge dependency

---

**Analysis Complete** ✅

