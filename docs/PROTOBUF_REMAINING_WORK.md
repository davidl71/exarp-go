# Protobuf Migration - Remaining Work

**Last Updated:** 2026-01-13  
**Status:** 7 of 27 handlers migrated (26% complete)

## ✅ Migrated Handlers (7)

1. ✅ **Memory** - `handleMemory` / `handleMemoryNative`
2. ✅ **Context** - `handleContext`
3. ✅ **Report** - `handleReport`
4. ✅ **Task Workflow** - `handleTaskWorkflow`
5. ✅ **Automation** - `handleAutomation`
6. ✅ **Testing** - `handleTesting`
7. ✅ **Lint** - `handleLint`

## ⏳ Remaining Handlers (20)

### High Priority (Frequently Used)

1. **Estimation** - `handleEstimation`
   - Schema: `EstimationRequest` (exists in proto/tools.proto)
   - Helper needed: `ParseEstimationRequest`, `EstimationRequestToParams`

2. **Session** - `handleSession`
   - Schema: `SessionRequest` (exists in proto/tools.proto)
   - Helper needed: `ParseSessionRequest`, `SessionRequestToParams`

3. **Git Tools** - `handleGitTools`
   - Schema: `GitToolsRequest` (exists in proto/tools.proto)
   - Helper needed: `ParseGitToolsRequest`, `GitToolsRequestToParams`

4. **Health** - `handleHealth`
   - Schema: `HealthRequest` (exists in proto/tools.proto)
   - Helper needed: `ParseHealthRequest`, `HealthRequestToParams`

5. **Security** - `handleSecurity`
   - Schema: `SecurityRequest` (exists in proto/tools.proto)
   - Helper needed: `ParseSecurityRequest`, `SecurityRequestToParams`

### Medium Priority

6. **Memory Maintenance** - `handleMemoryMaint`
   - Schema: `MemoryMaintRequest` (exists in proto/tools.proto)
   - Helper needed: `ParseMemoryMaintRequest`, `MemoryMaintRequestToParams`

7. **Task Analysis** - `handleTaskAnalysis`
   - Schema: `TaskAnalysisRequest` (exists in proto/tools.proto)
   - Helper needed: `ParseTaskAnalysisRequest`, `TaskAnalysisRequestToParams`

8. **Task Discovery** - `handleTaskDiscovery`
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
- **Migrated:** 7 (26%)
- **Remaining:** 20 (74%)
- **Schemas Created:** ✅ All 27 tool schemas exist in `proto/tools.proto`
- **Helper Functions:** 7 created, 20 remaining

## 🚀 Next Steps

1. **Batch migrate quick wins** (5 handlers: health, security, infer_session_mode, tool_catalog, workflow_mode)
2. **Migrate high-priority handlers** (5 handlers: estimation, session, git_tools, memory_maint, task_analysis)
3. **Migrate remaining handlers** (10 handlers)

Estimated time: 2-3 hours for all remaining handlers (following established pattern)
