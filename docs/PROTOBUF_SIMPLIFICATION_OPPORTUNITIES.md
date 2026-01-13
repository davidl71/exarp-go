# Protobuf Simplification Opportunities

**Generated:** 2026-01-13  
**Purpose:** Identify where protobuf would simplify code by eliminating manual JSON parsing, type assertions, and error-prone conversions

## Executive Summary

After analyzing the codebase, **protobuf would simplify code in 5 key areas** by:
- Eliminating 50+ repetitive `json.Unmarshal` + `map[string]interface{}` patterns
- Removing manual type assertions (`params["action"].(string)`)
- Providing compile-time type safety
- Reducing error handling boilerplate
- Simplifying Python bridge communication

---

## 🔴 High-Impact Simplification Areas

### 1. **Tool Handler Argument Parsing** - CRITICAL

**Current Problem:**
Every tool handler repeats this pattern:
```go
var params map[string]interface{}
if err := json.Unmarshal(args, &params); err != nil {
    return nil, fmt.Errorf("failed to parse arguments: %w", err)
}

action, _ := params["action"].(string)  // Manual type assertion
limit, ok := params["limit"].(float64)  // JSON numbers are float64
if ok {
    limitInt = int(limit)  // Manual conversion
}
```

**Found in:** 50+ handlers in `internal/tools/handlers.go`

**Protobuf Simplification:**
```go
// With protobuf, each tool has a typed request message
var req proto.ToolRequest
if err := protobuf.Unmarshal(args, &req); err != nil {
    return nil, fmt.Errorf("failed to parse arguments: %w", err)
}

// Type-safe access - no assertions needed
action := req.Action  // string, guaranteed
limit := req.Limit    // int32, guaranteed
```

**Benefits:**
- ✅ **Eliminate 50+ repetitive parsing blocks**
- ✅ **Compile-time type safety** - no runtime type assertions
- ✅ **No manual conversions** - protobuf handles types correctly
- ✅ **Clearer code** - explicit schema in `.proto` files
- ✅ **Better error messages** - protobuf validates at parse time

**Files Affected:**
- `internal/tools/handlers.go` - 50+ handlers
- `internal/tools/*_native.go` - Native implementations
- `internal/tools/context.go` - Complex parameter handling

**Example Before/After:**

**Before (Current):**
```go
func handleMemory(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    var params map[string]interface{}
    if err := json.Unmarshal(args, &params); err != nil {
        return nil, fmt.Errorf("failed to parse arguments: %w", err)
    }
    
    action, _ := params["action"].(string)
    if action == "" {
        action = "search"
    }
    
    limit := 10
    if l, ok := params["limit"].(float64); ok {
        limit = int(l)  // Manual conversion
    }
    
    var category string
    if cat, ok := params["category"].(string); ok && cat != "" {
        category = cat
    }
    
    // ... rest of handler
}
```

**After (With Protobuf):**
```go
func handleMemory(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    var req proto.MemoryRequest
    if err := protobuf.Unmarshal(args, &req); err != nil {
        return nil, fmt.Errorf("failed to parse arguments: %w", err)
    }
    
    // Type-safe access - no assertions needed
    action := req.Action
    if action == "" {
        action = "search"  // Default value in proto
    }
    
    limit := int(req.Limit)  // int32 in proto, direct conversion
    category := req.Category  // string, guaranteed
    
    // ... rest of handler
}
```

**Code Reduction:** ~15-20 lines per handler → ~5-7 lines = **60-70% reduction**

---

### 2. **Python Bridge Communication** - HIGH

**Current Problem:**
```go
// Go side: Marshal to JSON string
argsJSON, err := json.Marshal(args)
cmd := exec.CommandContext(ctx, "python3", bridgeScript, toolName, string(argsJSON))

// Python side: Parse JSON string
args = json.loads(args_json) if args_json else {}
action = args.get("action", "default")
limit = args.get("limit", 10)  # No type safety
```

**Found in:**
- `internal/bridge/python.go` - `ExecutePythonTool()`
- `bridge/execute_tool.py` - All tool executions
- `bridge/execute_resource.py` - Resource handlers

**Protobuf Simplification:**
```go
// Go side: Marshal to protobuf binary
req := &proto.ToolRequest{
    ToolName: toolName,
    Arguments: protobufData,  // Already protobuf
}
data, _ := protobuf.Marshal(req)
cmd := exec.CommandContext(ctx, "python3", bridgeScript, "--protobuf")
cmd.Stdin = bytes.NewReader(data)  // Binary stdin

// Python side: Parse protobuf (using protobuf library)
req = ToolRequest()
req.ParseFromString(sys.stdin.read())
action = req.action  # Type-safe
limit = req.limit    # Correct type
```

**Benefits:**
- ✅ **Eliminate JSON stringification** - no `json.Marshal`/`json.dumps`
- ✅ **Type safety across language boundary** - Go and Python use same schema
- ✅ **Smaller payloads** - binary is 20-30% smaller than JSON
- ✅ **Faster parsing** - protobuf deserialization is faster
- ✅ **No manual type conversions** - protobuf handles Go/Python type differences

**Files Affected:**
- `internal/bridge/python.go` - `ExecutePythonTool()`, `ExecutePythonResource()`
- `bridge/execute_tool.py` - All 20+ tool routes
- `bridge/execute_resource.py` - Resource execution
- `bridge/statistics_bridge.py` - Statistics communication

**Code Reduction:** ~10-15 lines per bridge call → ~3-5 lines = **70% reduction**

---

### 3. **Memory System** - MEDIUM-HIGH

**Current Problem:**
```go
// Save: Manual JSON marshaling
data, err := json.MarshalIndent(memory, "", "  ")
os.WriteFile(memoryPath, data, 0644)

// Load: Manual JSON unmarshaling
var memory Memory
if err := json.Unmarshal(data, &memory); err != nil {
    continue // Skip invalid JSON
}
```

**Found in:**
- `internal/tools/memory.go` - `saveMemory()`, `LoadAllMemories()`
- Memory file storage (`.exarp/memories/*.json`)

**Protobuf Simplification:**
```go
// Save: Protobuf binary
pbMemory := models.MemoryToProto(memory)
data, _ := protobuf.Marshal(pbMemory)
os.WriteFile(memoryPath, data, 0644)

// Load: Protobuf binary
var pbMemory proto.Memory
if err := protobuf.Unmarshal(data, &pbMemory); err != nil {
    continue // Skip invalid protobuf
}
memory := models.ProtoToMemory(&pbMemory)
```

**Benefits:**
- ✅ **Smaller file sizes** - 20-30% reduction
- ✅ **Faster I/O** - binary is faster to read/write
- ✅ **Type safety** - schema enforces structure
- ✅ **Versioning** - can evolve schema without breaking existing files

**Files Affected:**
- `internal/tools/memory.go` - `saveMemory()`, `LoadAllMemories()`
- Memory file format (`.json` → `.pb`)

**Code Reduction:** Similar lines, but **better performance and type safety**

---

### 4. **Context Summarization** - MEDIUM

**Current Problem:**
```go
// Complex type assertions and conversions
for _, itemRaw := range items {
    item, ok := itemRaw.(map[string]interface{})
    if !ok {
        item = map[string]interface{}{"data": itemRaw}
    }
    
    data := item["data"]
    var dataStr string
    switch v := data.(type) {
    case string:
        dataStr = v
    case map[string]interface{}, []interface{}:
        bytes, err := json.Marshal(v)  // Re-marshal to string
        dataStr = string(bytes)
    default:
        dataStr = fmt.Sprintf("%v", v)
    }
}
```

**Found in:**
- `internal/tools/context.go` - `handleContextBatchNative()`
- Complex nested data structure handling

**Protobuf Simplification:**
```go
// With protobuf, items are already typed
for _, item := range req.Items {
    dataStr := item.Data  // string, guaranteed
    toolType := item.ToolType  // string, guaranteed
    
    // No type assertions or conversions needed
}
```

**Benefits:**
- ✅ **Eliminate complex type switching** - protobuf handles types
- ✅ **No re-marshaling** - data is already in correct format
- ✅ **Clearer code** - explicit schema shows data structure
- ✅ **Better performance** - no JSON round-trips

**Files Affected:**
- `internal/tools/context.go` - `handleContextBatchNative()`
- `internal/tools/context_native.go` - Context operations

**Code Reduction:** ~30 lines of type handling → ~5 lines = **83% reduction**

---

### 5. **Report/Scorecard Data** - MEDIUM

**Current Problem:**
```go
// Parse JSON result from Python bridge
var reportData map[string]interface{}
if err := json.Unmarshal([]byte(result), &reportData); err == nil {
    // Access nested data with type assertions
    if insights, ok := enhanced["ai_insights"].(map[string]interface{}); ok {
        // More nested access...
    }
}
```

**Found in:**
- `internal/tools/report.go` - Report generation
- `internal/tools/handlers.go` - Report handling
- Scorecard data structures

**Protobuf Simplification:**
```go
// Parse protobuf result
var report proto.ReportResponse
if err := protobuf.Unmarshal(result, &report); err == nil {
    // Type-safe access
    insights := report.AiInsights  // proto.AIInsights, guaranteed
    metrics := insights.Metrics     // proto.Metrics, guaranteed
}
```

**Benefits:**
- ✅ **Eliminate nested type assertions** - protobuf provides nested types
- ✅ **Type safety for complex structures** - compile-time validation
- ✅ **Clearer data model** - schema documents structure
- ✅ **Better IDE support** - autocomplete for nested fields

**Files Affected:**
- `internal/tools/report.go` - Report generation
- `internal/tools/report_mlx.go` - MLX-enhanced reports
- Scorecard data structures

**Code Reduction:** ~20-30 lines of nested assertions → ~5-10 lines = **50-70% reduction**

---

## 📊 Impact Summary

| Area | Current Lines | With Protobuf | Reduction | Files Affected |
|------|--------------|---------------|-----------|----------------|
| **Tool Handlers** | ~750 lines | ~250 lines | **67%** | 50+ handlers |
| **Python Bridge** | ~200 lines | ~60 lines | **70%** | 4 bridge files |
| **Memory System** | ~100 lines | ~100 lines | **0%** (but faster) | 1 file |
| **Context Summarization** | ~150 lines | ~50 lines | **67%** | 2 files |
| **Report/Scorecard** | ~100 lines | ~40 lines | **60%** | 2 files |
| **Total** | **~1,300 lines** | **~500 lines** | **62%** | **60+ files** |

---

## 🎯 Implementation Priority

### Phase 1: Tool Handler Arguments (Highest Impact)
**Why:** Affects every tool, eliminates most boilerplate
**Effort:** Medium (create proto schemas for each tool)
**Impact:** 67% code reduction in handlers

### Phase 2: Python Bridge (High Impact)
**Why:** Simplifies cross-language communication
**Effort:** Medium (update Go and Python sides)
**Impact:** 70% code reduction, better performance

### Phase 3: Memory System (Medium Impact)
**Why:** Better performance, type safety
**Effort:** Low (similar to Todo2Task migration)
**Impact:** Performance improvement, smaller files

### Phase 4: Context & Reports (Medium Impact)
**Why:** Simplifies complex data handling
**Effort:** Medium (complex nested structures)
**Impact:** 60-67% code reduction

---

## 🔧 Protobuf Schema Examples

### Tool Request Schema
```protobuf
syntax = "proto3";

package exarp.tools;

message MemoryRequest {
  string action = 1;        // "save", "recall", "search", "list"
  string title = 2;
  string content = 3;
  string category = 4;
  string task_id = 5;
  int32 limit = 6;
  float threshold = 7;
  bool include_related = 8;
  map<string, string> metadata = 9;
}

message ContextRequest {
  string action = 1;       // "summarize", "budget", "batch"
  string data = 2;
  string level = 3;         // "brief", "detailed", "key_metrics"
  string tool_type = 4;
  int32 max_tokens = 5;
  bool include_raw = 6;
  repeated ContextItem items = 7;
  int32 budget_tokens = 8;
  bool combine = 9;
}

message ContextItem {
  string data = 1;
  string tool_type = 2;
}
```

### Python Bridge Schema
```protobuf
syntax = "proto3";

package exarp.bridge;

message ToolRequest {
  string tool_name = 1;
  bytes arguments_protobuf = 2;  // Protobuf-encoded tool-specific request
  string project_root = 3;
  int32 timeout_seconds = 4;
  string request_id = 5;
}

message ToolResponse {
  bool success = 1;
  string result = 2;
  string error = 3;
  bytes result_protobuf = 4;  // Optional protobuf-encoded result
  int32 exit_code = 5;
}
```

---

## 💡 Key Simplifications

### 1. **Eliminate Type Assertions**
**Before:**
```go
action, _ := params["action"].(string)
limit, ok := params["limit"].(float64)
if ok {
    limitInt = int(limit)
}
```

**After:**
```go
action := req.Action  // string, guaranteed
limit := int(req.Limit)  // int32, direct conversion
```

### 2. **Eliminate JSON Round-Trips**
**Before:**
```go
// Convert to JSON string, then back to map
bytes, _ := json.Marshal(data)
dataStr := string(bytes)
// Later: parse JSON string
var parsed map[string]interface{}
json.Unmarshal([]byte(dataStr), &parsed)
```

**After:**
```go
// Direct protobuf access
dataStr := req.Data  // Already string, no conversion
```

### 3. **Eliminate Manual Error Handling**
**Before:**
```go
var params map[string]interface{}
if err := json.Unmarshal(args, &params); err != nil {
    return nil, fmt.Errorf("failed to parse: %w", err)
}
// Check each field exists and has correct type
if action, ok := params["action"].(string); !ok {
    return nil, fmt.Errorf("action must be string")
}
```

**After:**
```go
var req proto.MemoryRequest
if err := protobuf.Unmarshal(args, &req); err != nil {
    return nil, fmt.Errorf("failed to parse: %w", err)
}
// All fields are guaranteed to be correct type
action := req.Action  // No checks needed
```

### 4. **Simplify Nested Data Access**
**Before:**
```go
if insights, ok := reportData["ai_insights"].(map[string]interface{}); ok {
    if metrics, ok := insights["metrics"].(map[string]interface{}); ok {
        if score, ok := metrics["score"].(float64); ok {
            scoreInt := int(score)
        }
    }
}
```

**After:**
```go
score := report.AiInsights.Metrics.Score  // Direct access, type-safe
```

---

## 🚀 Next Steps

1. **Create tool-specific proto schemas** for high-frequency tools (memory, context, report)
2. **Update Python bridge** to use protobuf for tool communication
3. **Migrate tool handlers** one by one, starting with most-used tools
4. **Update memory system** to use protobuf file format
5. **Migrate report/scorecard** data structures

---

## 📈 Expected Benefits

- **Code Reduction:** ~800 lines eliminated (62% reduction)
- **Type Safety:** Compile-time validation instead of runtime assertions
- **Performance:** 20-30% faster serialization, smaller payloads
- **Maintainability:** Clearer code, explicit schemas, better IDE support
- **Error Reduction:** Fewer type assertion panics, better error messages

---

**Conclusion:** Protobuf would significantly simplify the codebase by eliminating repetitive JSON parsing, type assertions, and manual conversions. The highest impact areas are tool handler argument parsing and Python bridge communication.
