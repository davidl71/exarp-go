# Tool to Resource Conversion Plan

**Date:** 2026-01-09  
**Objective:** Convert simple read-only tools to MCP resources for better API semantics and caching

## Analysis Summary

Identified **3 tools** that should be converted to resources:

1. ✅ **`server_status`** → `stdio://server/status` resource (HIGH PRIORITY)
2. ✅ **`list_models`** → `stdio://models` resource (HIGH PRIORITY)  
3. ✅ **`tool_catalog` (list action)** → `stdio://tools` resource (MEDIUM PRIORITY)

## Rationale

- **Resources** = data retrieval (read-only, queryable data)
- **Tools** = actions (operations, inputs, state changes)
- **Benefits:** Clearer API semantics, simpler client code, better caching, more idiomatic MCP usage

## Implementation Tasks

### Task 1: Research Resource Implementation Patterns
- Review existing resource handlers (scorecard.go, memories.go, tasks.go)
- Understand URI template parsing patterns
- Document resource handler structure
- Identify any differences from tool handlers

### Task 2: Convert server_status to stdio://server/status resource
- Create `internal/resources/server.go` with `handleServerStatus` handler
- Register resource in `internal/resources/handlers.go`
- Remove tool registration from `internal/tools/registry.go` (line 1478-1489)
- Update tests in `internal/resources/handlers_test.go`
- Update expected resource count (17 → 18)

### Task 3: Convert list_models to stdio://models resource
- Create `internal/resources/models.go` with `handleModels` handler
- Move MODEL_CATALOG constant from `internal/tools/recommend.go`
- Register resource in `internal/resources/handlers.go`
- Remove tool registration from `internal/tools/registry.go` (registerBatch4Tools)
- Update tests
- Update expected resource count (18 → 19)

### Task 4: Convert tool_catalog (list) to stdio://tools resource
- Create `internal/resources/tools.go` with `handleTools` handler
- Move getToolCatalog() and filtering logic from `internal/tools/tool_catalog.go`
- Support URI templates: `stdio://tools` and `stdio://tools/{category}`
- Register resources in `internal/resources/handlers.go`
- Keep `tool_catalog` tool but remove "list" action (keep only "help")
- Update tests
- Update expected resource count (19 → 21, two resources for tools)

### Task 5: Update Documentation and Tests
- Update `cmd/sanity-check/main.go` resource count (17 → 21)
- Update documentation mentioning these tools
- Add migration notes
- Update any examples or usage docs

## Files to Modify

### Create New Files:
- `internal/resources/server.go`
- `internal/resources/models.go`
- `internal/resources/tools.go`

### Modify Existing Files:
- `internal/resources/handlers.go` - Add registrations
- `internal/resources/handlers_test.go` - Add tests
- `internal/tools/registry.go` - Remove tool registrations
- `internal/tools/tool_catalog.go` - Remove list action, keep help
- `internal/tools/recommend.go` - Move MODEL_CATALOG or keep reference
- `cmd/sanity-check/main.go` - Update counts
- `docs/PROMPTS_RESOURCES_REVIEW.md` - Update resource count

## Implementation Notes

### server_status Conversion
- Very simple - just returns static JSON
- No parameters needed
- Can reuse existing handler code directly

### list_models Conversion  
- Returns static MODEL_CATALOG array
- No parameters needed
- Recommend tool still needs access to MODEL_CATALOG, so either:
  - Move constant to shared package, OR
  - Keep constant in tools package and import it in resources

### tool_catalog Conversion
- More complex - supports filtering
- Need URI template support: `stdio://tools/{category}`
- Consider query parameters vs URI templates (MCP prefers URI templates)
- Keep "help" action in tool since it performs analysis

## Testing Strategy

1. Test resource handlers return correct JSON format
2. Test URI template parsing (for tools/{category})
3. Test resource registration counts
4. Verify tools are removed from tool registry
5. Test backward compatibility (if any clients use these tools)

## Migration Strategy

- **Deprecation Period:** Mark tools as deprecated in documentation
- **Parallel Support:** Consider keeping tools for one version, then removing
- **Documentation:** Update all examples and docs to use resources instead

## Priority Order

1. **server_status** (simplest, good proof of concept)
2. **list_models** (simple, builds confidence)
3. **tool_catalog** (most complex, do last)
