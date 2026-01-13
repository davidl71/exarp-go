# mcp-go-core Integration Plan for exarp-go

**Date:** 2026-01-12  
**Status:** ✅ **COMPLETE**  
**Goal:** Replace local implementations with shared mcp-go-core library

---

## Overview

Integrate `mcp-go-core` shared library into `exarp-go` to:
- Reduce code duplication
- Improve maintainability
- Share improvements across projects (exarp-go, devwisdom-go)
- Standardize MCP infrastructure

---

## Components to Replace

### 1. Framework Abstraction ✅ Ready
- **Current:** `internal/framework/server.go`, `internal/framework/factory.go`
- **Replace with:** `github.com/davidl71/mcp-go-core/pkg/mcp/framework`
- **Status:** mcp-go-core has identical interface
- **Files affected:** ~50+ files

### 2. Security Utilities ✅ Ready
- **Current:** `internal/security/path.go`, `internal/security/access.go`, `internal/security/ratelimit.go`
- **Replace with:** `github.com/davidl71/mcp-go-core/pkg/mcp/security`
- **Status:** mcp-go-core has identical implementations
- **Files affected:** ~30+ files

### 3. Factory Functions ✅ Ready
- **Current:** `internal/factory/server.go`
- **Replace with:** `github.com/davidl71/mcp-go-core/pkg/mcp/factory`
- **Status:** mcp-go-core has factory with same API
- **Files affected:** `cmd/server/main.go`, `internal/cli/cli.go`

### 4. Common Types ✅ Ready
- **Current:** Types in `internal/framework/server.go` (TextContent, ToolSchema, ToolInfo)
- **Replace with:** `github.com/davidl71/mcp-go-core/pkg/mcp/types`
- **Status:** mcp-go-core has identical types
- **Files affected:** All files using these types

### 5. Go SDK Adapter ✅ Ready
- **Current:** `internal/framework/adapters/gosdk/adapter.go`
- **Replace with:** `github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk`
- **Status:** mcp-go-core has identical adapter
- **Files affected:** Factory usage

---

## Integration Strategy

### Option 1: Direct Replacement (Recommended)
- Replace imports directly
- Remove local implementations
- Fastest approach, cleanest result

### Option 2: Compatibility Layer (Gradual)
- Create thin wrappers in `internal/` packages
- Gradually migrate to direct imports
- More gradual, safer for large codebase

**Recommendation:** **Option 1** - Direct replacement
- Types are identical (verified)
- Interfaces are identical (verified)
- Can test incrementally
- Cleaner long-term result

---

## Integration Steps

### Phase 1: Setup (5 minutes)
1. ✅ Add local path replacement to `go.mod`
2. ✅ Add mcp-go-core dependency
3. ✅ Verify dependency resolves

### Phase 2: Types Migration (15 minutes)
1. Update `internal/framework/server.go` to use `mcp-go-core/pkg/mcp/types`
2. Update all files using `framework.TextContent` → `types.TextContent`
3. Update all files using `framework.ToolSchema` → `types.ToolSchema`
4. Update all files using `framework.ToolInfo` → `types.ToolInfo`
5. Test build

### Phase 3: Framework Migration (30 minutes)
1. Update `internal/framework/server.go` to re-export from mcp-go-core
2. Or replace imports directly: `internal/framework` → `mcp-go-core/pkg/mcp/framework`
3. Update factory to use `mcp-go-core/pkg/mcp/factory`
4. Update adapter imports
5. Test build

### Phase 4: Security Migration (20 minutes)
1. Replace `internal/security` imports with `mcp-go-core/pkg/mcp/security`
2. Update all files using security utilities
3. Test build

### Phase 5: Cleanup (10 minutes)
1. Remove local implementations (or keep as deprecated wrappers)
2. Update documentation
3. Run full test suite
4. Verify MCP server still works

---

## File-by-File Migration Plan

### High Priority (Core Files)

1. **`cmd/server/main.go`**
   - Replace `internal/factory` → `mcp-go-core/pkg/mcp/factory`
   - Replace `internal/framework` → `mcp-go-core/pkg/mcp/framework`
   - Test: Server starts correctly

2. **`internal/framework/server.go`**
   - Option A: Re-export from mcp-go-core (compatibility)
   - Option B: Remove, update all imports (clean)
   - **Recommendation:** Option B (direct imports)

3. **`internal/factory/server.go`**
   - Replace with `mcp-go-core/pkg/mcp/factory`
   - Or remove if using mcp-go-core factory directly

4. **`internal/framework/adapters/gosdk/adapter.go`**
   - Replace with `mcp-go-core/pkg/mcp/framework/adapters/gosdk`
   - Or remove if using mcp-go-core adapter directly

### Medium Priority (Tool Files)

5. **All files in `internal/tools/`**
   - Update `framework.TextContent` → `types.TextContent`
   - Update `framework.ToolSchema` → `types.ToolSchema`
   - Update imports: `internal/framework` → `mcp-go-core/pkg/mcp/framework` or `mcp-go-core/pkg/mcp/types`

6. **All files using `internal/security`**
   - Update imports: `internal/security` → `mcp-go-core/pkg/mcp/security`
   - Functions are identical, no code changes needed

---

## Testing Strategy

### Build Tests
1. ✅ `go build ./cmd/server` - Server builds
2. ✅ `go build ./internal/...` - All packages build
3. ✅ `go test ./...` - All tests pass

### Integration Tests
1. ✅ Server starts without errors
2. ✅ Tools register correctly
3. ✅ Tool execution works
4. ✅ Security utilities work (path validation, etc.)

### MCP Protocol Tests
1. ✅ Initialize handshake works
2. ✅ Tools list works
3. ✅ Tool calls work
4. ✅ Resources work
5. ✅ Prompts work

---

## Rollback Plan

If issues arise:
1. Revert `go.mod` changes
2. Restore local implementations
3. Investigate issues
4. Fix and retry

---

## Success Criteria

- [ ] All imports updated to use mcp-go-core
- [ ] All tests passing
- [ ] Server builds successfully
- [ ] MCP server works correctly
- [ ] No functionality regressions
- [ ] Documentation updated

---

## Estimated Time

- **Setup:** 5 minutes
- **Types Migration:** 15 minutes
- **Framework Migration:** 30 minutes
- **Security Migration:** 20 minutes
- **Cleanup:** 10 minutes
- **Testing:** 20 minutes
- **Total:** ~100 minutes (~1.5 hours)

---

## Next Steps

1. ✅ Add dependency (local path replacement)
2. ⏳ Start with types migration (lowest risk)
3. ⏳ Migrate framework abstraction
4. ⏳ Migrate security utilities
5. ⏳ Clean up local implementations
6. ⏳ Test thoroughly

---

**Status:** Ready to begin integration
