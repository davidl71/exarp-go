# Protobuf Migration - Remaining Work

**Last Updated:** 2026-01-13  
**Status:** Tool handlers complete (27/27), Memory system migrated ✅

## ✅ Migrated Handlers (13)

1. ✅ **Memory** - `handleMemory` / `handleMemoryNative`
2. ✅ **Context** - `handleContext`
3. ✅ **Report** - `handleReport`
4. ✅ **Task Workflow** - `handleTaskWorkflow`
5. ✅ **Automation** - `handleAutomation`
6. ✅ **Testing** - `handleTesting`
7. ✅ **Lint** - `handleLint`
8. ✅ **Health** - `handleHealth`
9. ✅ **Security** - `handleSecurity`
10. ✅ **Infer Session Mode** - `handleInferSessionMode`
11. ✅ **Tool Catalog** - `handleToolCatalog`
12. ✅ **Workflow Mode** - `handleWorkflowMode`
13. ✅ **Estimation** - `handleEstimation`
14. ✅ **Session** - `handleSession`
15. ✅ **Git Tools** - `handleGitTools`
16. ✅ **Memory Maintenance** - `handleMemoryMaint`
17. ✅ **Task Analysis** - `handleTaskAnalysis`
18. ✅ **Task Discovery** - `handleTaskDiscovery`
19. ✅ **Ollama** - `handleOllama`
20. ✅ **MLX** - `handleMlx`
21. ✅ **Prompt Tracking** - `handlePromptTracking`
22. ✅ **Recommend** - `handleRecommend`
23. ✅ **Analyze Alignment** - `handleAnalyzeAlignment`
24. ✅ **Generate Config** - `handleGenerateConfig`
25. ✅ **Setup Hooks** - `handleSetupHooks`
26. ✅ **Check Attribution** - `handleCheckAttribution`
27. ✅ **Add External Tool Hints** - `handleAddExternalToolHints`

## ✅ Completed Systems

### 1. Tool Handler Arguments (27/27 handlers) ✅
- All handlers now support protobuf request parsing with JSON fallback
- Helper functions created in `internal/tools/protobuf_helpers.go`
- Backward compatible during transition period

### 2. Memory System File Format ✅
- `saveMemory()` now saves as `.pb` (protobuf binary)
- `LoadAllMemories()` auto-detects format (.pb first, .json fallback)
- Automatic migration from JSON to protobuf on load
- `deleteMemoryFile()` helper handles both formats
- Test helpers updated to use protobuf format

## ⏳ Remaining Simplification Opportunities

### High Priority

1. **Python Bridge Communication** - Infrastructure Added ⚠️
   - Go side: Creates protobuf `ToolRequest` (prepared)
   - Python side: Still uses JSON (backward compatible)
   - **Next Step:** Generate Python protobuf code from `bridge.proto`
   - **Status:** Infrastructure ready, waiting for Python protobuf code generation

### Medium Priority

2. **Context Summarization** - Pending
   - Complex JSON manipulation in batch operations
   - Can use `proto.ContextItem` for type safety
   - Files: `internal/tools/context.go`, `internal/tools/context_native.go`

3. **Report/Scorecard Data** - Pending
   - Nested `map[string]interface{}` structures
   - Can use `proto.ReportRequest` and `proto.ReportResponse`
   - Files: `internal/tools/report.go`, `internal/tools/report_mlx.go`
   - Schema: `TaskDiscoveryRequest` (exists in proto/tools.proto)
   - Helper needed: `ParseTaskDiscoveryRequest`, `TaskDiscoveryRequestToParams`

9. **Workflow Mode** - `handleWorkflowMode`
   - Schema: `WorkflowModeRequest` (exists in proto/tools.proto)
   - Helper needed: `ParseWorkflowModeRequest`, `WorkflowModeRequestToParams`

10. **Tool Catalog** - `handleToolCatalog`
    - Schema: `ToolCatalogRequest` (exists in proto/tools.proto)
    - Helper needed: `ParseToolCatalogRequest`, `ToolCatalogRequestToParams`

11. **Prompt Tracking** - `handlePromptTracking`
    - Schema: `PromptTrackingRequest` (exists in proto/tools.proto)
    - Helper needed: `ParsePromptTrackingRequest`, `PromptTrackingRequestToParams`

12. **Recommend** - `handleRecommend`
    - Schema: `RecommendRequest` (exists in proto/tools.proto)
    - Helper needed: `ParseRecommendRequest`, `RecommendRequestToParams`

### Lower Priority (Less Frequently Used)

13. **Analyze Alignment** - `handleAnalyzeAlignment`
    - Schema: `AnalyzeAlignmentRequest` (exists in proto/tools.proto)
    - Helper needed: `ParseAnalyzeAlignmentRequest`, `AnalyzeAlignmentRequestToParams`

14. **Generate Config** - `handleGenerateConfig`
    - Schema: `GenerateConfigRequest` (exists in proto/tools.proto)
    - Helper needed: `ParseGenerateConfigRequest`, `GenerateConfigRequestToParams`

15. **Setup Hooks** - `handleSetupHooks`
    - Schema: `SetupHooksRequest` (exists in proto/tools.proto)
    - Helper needed: `ParseSetupHooksRequest`, `SetupHooksRequestToParams`

16. **Check Attribution** - `handleCheckAttribution`
    - Schema: `CheckAttributionRequest` (exists in proto/tools.proto)
    - Helper needed: `ParseCheckAttributionRequest`, `CheckAttributionRequestToParams`

17. **Add External Tool Hints** - `handleAddExternalToolHints`
    - Schema: `AddExternalToolHintsRequest` (exists in proto/tools.proto)
    - Helper needed: `ParseAddExternalToolHintsRequest`, `AddExternalToolHintsRequestToParams`

18. **Ollama** - `handleOllama`
    - Schema: `OllamaRequest` (exists in proto/tools.proto)
    - Helper needed: `ParseOllamaRequest`, `OllamaRequestToParams`

19. **MLX** - `handleMlx`
    - Schema: `MlxRequest` (exists in proto/tools.proto)
    - Helper needed: `ParseMlxRequest`, `MlxRequestToParams`

20. **Infer Session Mode** - `handleInferSessionMode`
    - Schema: `InferSessionModeRequest` (exists in proto/tools.proto)
    - Helper needed: `ParseInferSessionModeRequest`, `InferSessionModeRequestToParams`

## 📋 Migration Pattern

For each remaining handler, follow this pattern:

### Step 1: Add Helper Functions

Add to `internal/tools/protobuf_helpers.go`:

```go
// Parse[ToolName]Request parses a [tool] request (protobuf or JSON)
func Parse[ToolName]Request(args json.RawMessage) (*proto.[ToolName]Request, map[string]interface{}, error) {
	var req proto.[ToolName]Request

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// [ToolName]RequestToParams converts a protobuf [ToolName]Request to params map
func [ToolName]RequestToParams(req *proto.[ToolName]Request) map[string]interface{} {
	params := make(map[string]interface{})
	// Convert all fields from req to params
	// Handle repeated fields (arrays) by converting to JSON strings
	// Handle optional fields with defaults
	return params
}
```

### Step 2: Update Handler

Update handler in `internal/tools/handlers.go`:

```go
func handle[ToolName](ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := Parse[ToolName]Request(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = [ToolName]RequestToParams(req)
		// Set defaults for protobuf request
	}

	// Continue with existing handler logic using params
	// ...
}
```

## 🎯 Quick Wins (Easiest to Migrate)

These handlers have simple parameter structures and can be migrated quickly:

1. **Health** - Simple action-based tool
2. **Security** - Simple action-based tool
3. **Infer Session Mode** - Very simple (no parameters)
4. **Tool Catalog** - Simple action-based tool
5. **Workflow Mode** - Simple action-based tool

## 📊 Progress Summary

- **Total Handlers:** 27
- **Migrated:** 27 (100%) ✅
- **Remaining:** 0 ✅
- **Schemas Created:** ✅ All 27 tool schemas exist in `proto/tools.proto`
- **Helper Functions:** ✅ All 27 helper function pairs created

## 🚀 Next Steps

1. **Batch migrate quick wins** (5 handlers: health, security, infer_session_mode, tool_catalog, workflow_mode)
2. **Migrate high-priority handlers** (5 handlers: estimation, session, git_tools, memory_maint, task_analysis)
3. **Migrate remaining handlers** (10 handlers)

Estimated time: 2-3 hours for all remaining handlers (following established pattern)
