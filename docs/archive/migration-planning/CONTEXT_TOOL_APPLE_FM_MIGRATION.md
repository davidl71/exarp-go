# Context Tool Migration: Native Go with Apple Foundation Models

**Date:** 2026-01-08  
**Status:** ✅ Complete  
**Purpose:** Migrate context tool's summarize action to use native Go with Apple Foundation Models

---

## Summary

Successfully migrated the `context` tool's `summarize` action to use native Go implementation with Apple Foundation Models, while maintaining Python bridge as fallback for other actions and unsupported platforms.

---

## Implementation Details

### Files Created

1. **`internal/tools/context_native.go`** (Apple platforms only)
   - Build tags: `//go:build darwin && arm64 && cgo`
   - Implements `summarizeWithAppleFM()` - Uses Apple Foundation Models for summarization
   - Implements `handleContextSummarizeNative()` - Native Go handler for summarize action
   - Supports all summarization levels: `brief`, `detailed`, `key_metrics`, `actionable`

2. **`internal/tools/context_native_nocgo.go`** (Non-Apple platforms)
   - Build tags: `//go:build !(darwin && arm64 && cgo)`
   - No-op implementation for platforms without Apple Foundation Models support
   - Ensures code compiles on all platforms

### Files Modified

1. **`internal/tools/handlers.go`**
   - Updated `handleContext()` to check for Apple FM availability
   - Uses native Go implementation when Apple FM is available
   - Falls back to Python bridge for:
     - Non-summarize actions (budget, batch)
     - When Apple FM is not available
     - When native implementation fails

---

## How It Works

### Execution Flow

```
User calls context(action="summarize", data="...")
    ↓
handleContext() checks action
    ↓
If action == "summarize":
    ↓
Check Apple FM availability
    ↓
If available:
    → handleContextSummarizeNative() (Native Go + Apple FM)
    → Returns summary using Apple Foundation Models
    ↓
If not available or fails:
    → bridge.ExecutePythonTool() (Python bridge fallback)
    → Returns summary using Python implementation
```

### Summarization Levels

The native implementation supports all 4 levels with optimized prompts:

1. **`brief`** - One concise sentence with key metrics
   - Temperature: 0.3
   - Prompt: "Summarize the following text in one concise sentence with key metrics"

2. **`detailed`** - Multiple paragraphs with categories
   - Temperature: 0.3
   - Prompt: "Summarize the following text in multiple paragraphs with categories"

3. **`key_metrics`** - Extract only numerical metrics
   - Temperature: 0.2 (more deterministic)
   - Prompt: "Extract only the numerical metrics and key numbers from the following text. Return as JSON"

4. **`actionable`** - Extract only actionable items
   - Temperature: 0.2 (more deterministic)
   - Prompt: "Extract only actionable items (recommendations, tasks, fixes) from the following text. Return as JSON"

---

## Benefits

### ✅ Privacy
- All summarization happens on-device
- No data sent to external services
- Perfect for sensitive project data

### ✅ Performance
- Native Go implementation (no Python bridge overhead)
- Direct Apple Foundation Models API calls
- Faster response times

### ✅ Reliability
- Graceful fallback to Python bridge
- Works on all platforms (Apple FM when available, Python otherwise)
- No breaking changes to existing functionality

### ✅ Maintainability
- Native Go code (easier to maintain than Python bridge)
- Clear separation of concerns
- Platform-specific code properly isolated with build tags

---

## Platform Support

### ✅ Apple Silicon Macs (macOS 26+)
- Uses Apple Foundation Models (native Go)
- Fast, on-device processing
- Privacy-focused

### ✅ Other Platforms
- Falls back to Python bridge
- Uses existing Python implementation
- No functionality loss

---

## Testing

### Manual Testing

```bash
# Test on Apple Silicon Mac (macOS 26+)
./bin/exarp-go -tool context -args '{"action":"summarize","data":"{\"status\":\"success\",\"score\":85,\"issues\":3}"}'

# Should use Apple FM if available
# Should fallback to Python bridge if not available
```

### Expected Behavior

1. **Apple Silicon Mac (macOS 26+) with Apple Intelligence:**
   - Uses native Go + Apple FM
   - Returns: `{"method": "apple_foundation_models", ...}`

2. **Other platforms or macOS < 26:**
   - Falls back to Python bridge
   - Returns: `{"method": "python", ...}` (or no method field)

---

## Code Structure

```
internal/tools/
├── context.go                    # Budget action (native Go, all platforms)
├── context_native.go             # Summarize action (Apple FM, darwin+arm64+cgo)
├── context_native_nocgo.go       # No-op (other platforms)
└── handlers.go                   # Updated handleContext() with routing logic
```

---

## Integration Points

### Apple Foundation Models Integration

- Uses `github.com/blacktop/go-foundationmodels` package
- Platform detection via `internal/platform/detection.go`
- Session management with proper cleanup (`defer sess.Release()`)

### Python Bridge Fallback

- Maintains compatibility with existing Python implementation
- Seamless fallback when Apple FM unavailable
- No changes to Python code required

---

## Future Enhancements

### Potential Improvements

1. **Batch Summarization:**
   - Could use Apple FM for batch operations
   - Currently uses Python bridge

2. **Tool-Aware Summarization:**
   - Could enhance prompts with tool_type hints
   - Better context understanding

3. **Streaming Support:**
   - When Apple FM adds streaming APIs
   - Real-time summary generation

---

## Migration Notes

### Breaking Changes
- **None** - Fully backward compatible
- All existing functionality preserved
- Python bridge still works for all actions

### Performance Impact
- **Positive** - Faster on Apple Silicon Macs
- **Neutral** - No impact on other platforms

### Dependencies
- **New:** `github.com/blacktop/go-foundationmodels` (already in go.mod)
- **Existing:** Python bridge (still used for fallback)

---

## Conclusion

✅ **Migration Complete**

The context tool now uses native Go with Apple Foundation Models for summarization on supported platforms, while maintaining full backward compatibility through Python bridge fallback. This provides:

- Better privacy (on-device processing)
- Better performance (native Go)
- Better maintainability (Go code vs Python bridge)
- Full compatibility (works everywhere)

The implementation is production-ready and follows the same patterns as other native Go tools in the codebase.

