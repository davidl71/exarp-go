# Estimation Tool Model Caching Analysis

**Date:** 2026-01-09  
**Issue:** Estimation tool appears to load MLX model on each call  
**Status:** Analysis Complete

---

## Problem Analysis

### Current Behavior

When calling the estimation tool multiple times, you see:
```
Fetching 11 files: 100%|██████████| 11/11 [00:00<00:00, 154823.30it/s]
```

This happens **every time** the estimation tool is called, even though the code has caching.

### Root Cause

**The Python bridge creates a new Python process for each call:**

```go
// internal/bridge/python.go:83
cmd := exec.CommandContext(ctx, "python3", bridgeScript, toolName, string(argsJSON))
```

**Impact:**
1. Each `ExecutePythonTool` call = new Python process
2. In-memory cache (`_MLX_MODEL_CACHE`) is lost between calls
3. Model must be loaded from disk to memory each time

### What's Actually Happening

1. **First call:**
   - Python process starts
   - MLX checks cache (empty)
   - Hugging Face downloads model files (if not on disk) → "Fetching 11 files"
   - Model loads from disk to memory
   - Model cached in `_MLX_MODEL_CACHE`
   - Process exits → **cache lost**

2. **Second call:**
   - **New Python process starts** (cache is empty again!)
   - Hugging Face files are on disk (cached by HF)
   - But model still needs to load from disk to memory
   - "Fetching 11 files" might be HF checking/validating files
   - Model loads again

### Code Evidence

**Model Cache (in-memory):**
```python
# project_management_automation/tools/mlx_integration.py:64-65
_MLX_MODEL_CACHE: dict[str, tuple[Any, Any]] = {}  # model_name -> (model_obj, tokenizer)
```

**Cache Check:**
```python
# mlx_integration.py:407-410
if model in _MLX_MODEL_CACHE:
    if verbose:
        logger.info(f"Using cached MLX model: {model}")
    model_obj, tokenizer = _MLX_MODEL_CACHE[model]
```

**Problem:** Cache only works within a single Python process. Each bridge call = new process.

---

## Solutions

### Solution 1: Persistent Python Process (Recommended)

**Approach:** Run Python as a long-lived daemon/server instead of spawning new processes.

**Implementation:**
```go
// internal/bridge/python_daemon.go
type PythonDaemon struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout io.ReadCloser
    mu     sync.Mutex
}

func (d *PythonDaemon) ExecuteTool(toolName string, args map[string]interface{}) (string, error) {
    // Send JSON-RPC request to persistent process
    // Model stays loaded in memory
}
```

**Benefits:**
- ✅ Model stays in memory between calls
- ✅ Much faster (no model loading overhead)
- ✅ Cache works as intended

**Challenges:**
- More complex implementation
- Need to handle process lifecycle
- Need to handle errors/restarts

### Solution 2: Hugging Face Disk Cache (Current - Partial)

**Current Status:**
- Hugging Face caches model files on disk
- Location: `~/.cache/huggingface/hub/`
- Files are reused, but model still loads from disk each time

**What "Fetching 11 files" means:**
- Hugging Face is checking/validating cached files
- Or downloading missing files
- This is faster than full download, but still has overhead

**Optimization:**
- Ensure HF cache is working properly
- Pre-download models if possible
- Use `HF_HOME` environment variable to control cache location

### Solution 3: Model Loading Service (Advanced)

**Approach:** Create a separate model loading service that persists.

**Implementation:**
- Separate Python process that loads models on startup
- Go code communicates with service via HTTP/Unix socket
- Models stay loaded in service process

**Benefits:**
- Models always ready
- Can share across multiple Go processes
- Better resource management

**Challenges:**
- Most complex solution
- Need service management
- Additional infrastructure

### Solution 4: Disable MLX for Estimation (Quick Fix)

**Approach:** Use statistical estimation only, skip MLX.

**Implementation:**
```python
# In estimation tool
use_mlx=False  # Skip MLX, use statistical only
```

**Benefits:**
- ✅ No model loading overhead
- ✅ Faster response times
- ✅ Still accurate for most tasks

**Trade-offs:**
- Less accurate for novel/complex tasks
- Loses semantic understanding benefits

---

## Recommended Approach

### Short Term (Quick Fix)

**Option A: Disable MLX for batch estimation**
- When estimating multiple tasks, disable MLX
- Use MLX only for single important estimates
- Reduces overhead significantly

**Option B: Pre-load model in separate call**
- Make one "warm-up" call to load model
- Then estimate multiple tasks quickly
- Model stays in memory during session

### Medium Term (Better Solution)

**Implement persistent Python process:**
1. Create `PythonDaemon` struct in Go
2. Start Python process on first call
3. Keep process alive for session
4. Reuse for all Python bridge calls
5. Shutdown on Go process exit

**Benefits:**
- Model caching works as intended
- Faster execution
- Better resource utilization

### Long Term (Best Solution)

**Migrate estimation to native Go:**
- Use Go HTTP client for MLX API (if available)
- Or use statistical estimation only
- Or use Apple Foundation Models (native Go)
- Eliminates Python bridge overhead entirely

---

## Performance Impact

### Current (New Process Each Time)
- Model load: ~2-5 seconds
- "Fetching files": ~0.5-2 seconds
- **Total overhead: ~2.5-7 seconds per call**

### With Persistent Process
- First call: ~2-5 seconds (load model)
- Subsequent calls: ~0.1-0.5 seconds (cached)
- **Savings: ~2-6.5 seconds per call**

### With Native Go (No MLX)
- Statistical only: ~0.01-0.1 seconds
- **Savings: ~2.5-7 seconds per call**

---

## Immediate Action Items

1. **Document current behavior** - Add note about model loading
2. **Add verbose logging** - Show when cache is used vs model loaded
3. **Consider disabling MLX** - For batch operations
4. **Plan persistent process** - Add to Phase 3 migration plan

---

## Code Changes Needed

### Option 1: Add Verbose Logging
```python
# mlx_integration.py
if model in _MLX_MODEL_CACHE:
    logger.info(f"✅ Using cached MLX model: {model} (fast)")
else:
    logger.info(f"⏳ Loading MLX model: {model} (first time, will cache)")
    # ... load model ...
    logger.info(f"✅ Model loaded and cached (future calls will be faster)")
```

### Option 2: Add Process ID Logging
```python
import os
logger.info(f"Python process PID: {os.getpid()}")
# This will show if it's a new process each time
```

### Option 3: Check Hugging Face Cache
```python
from huggingface_hub import snapshot_download
cache_dir = snapshot_download(model, local_files_only=True)
# This will show if files are already cached
```

---

## Conclusion

**Current Issue:**
- Model cache exists but doesn't work because each call = new Python process
- "Fetching 11 files" is Hugging Face checking/validating cached files
- Model still loads from disk to memory each time

**Recommended Solution:**
- **Short term:** Disable MLX for batch operations or accept overhead
- **Medium term:** Implement persistent Python process for model caching
- **Long term:** Migrate estimation to native Go (Phase 3)

**Impact:**
- Current: ~2.5-7 seconds overhead per estimation call
- With fix: ~0.1-0.5 seconds (20-50x faster)

---

**Next Steps:**
1. Add logging to confirm process behavior
2. Decide on short-term approach (disable MLX vs accept overhead)
3. Plan persistent process implementation
4. Consider native Go migration priority

