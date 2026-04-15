# Python Bridge Process Pool Design

**Date:** 2026-01-13  
**Task:** T-1768325715211 - Performance audit: Python bridge subprocess overhead  
**Status:** Design Phase

---

## Problem Statement

**Current Issue:**
- Each `ExecutePythonTool` call creates a new Python subprocess
- Process creation overhead: ~50-200ms per call
- Python interpreter startup: ~100-300ms
- Module imports: ~200-500ms (first call)
- **Total overhead: ~350-1000ms per call** (before actual tool execution)

**Impact:**
- 32 calls to `ExecutePythonTool` across 7 files
- Model caching doesn't work (cache lost when process exits)
- Estimation tool loads MLX model on every call (~2-5 seconds overhead)
- In-memory caches are lost between calls

**Performance Goal:**
- Reduce overhead from ~350-1000ms to ~10-50ms per call
- Enable model caching (MLX, Ollama, etc.)
- Maintain backward compatibility

---

## Solution: Persistent Python Process Pool

### Architecture Overview

```
┌─────────────────────────────────────────┐
│  Go MCP Server                         │
│                                         │
│  ┌──────────────────────────────────┐  │
│  │  PythonBridgePool                │  │
│  │  - Manages persistent processes   │  │
│  │  - Health checking                │  │
│  │  - Automatic restart on errors    │  │
│  └──────────────────────────────────┘  │
│           │                             │
│           │ JSON-RPC over stdin/stdout  │
│           ▼                             │
│  ┌──────────────────────────────────┐  │
│  │  Python Daemon Process            │  │
│  │  - Persistent process              │  │
│  │  - Reads requests from stdin       │  │
│  │  - Writes responses to stdout     │  │
│  │  - Keeps models/caches in memory  │  │
│  └──────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

### Communication Protocol

**JSON-RPC 2.0 over stdin/stdout:**
- Each request: JSON object on single line
- Each response: JSON object on single line
- Line-delimited protocol (newline-separated)
- Request ID for correlation

**Request Format:**
```json
{
  "jsonrpc": "2.0",
  "id": "req-1234567890",
  "method": "execute_tool",
  "params": {
    "tool_name": "estimation",
    "args": {...}
  }
}
```

**Response Format:**
```json
{
  "jsonrpc": "2.0",
  "id": "req-1234567890",
  "result": "...",
  "error": null
}
```

---

## Implementation Plan

### Phase 1: Python Daemon Script

**File:** `bridge/execute_tool_daemon.py`

**Features:**
- Reads JSON-RPC requests from stdin (line-delimited)
- Executes tools using existing `execute_tool()` function
- Writes JSON-RPC responses to stdout
- Handles errors gracefully
- Keeps process alive between requests

**Key Functions:**
```python
def main():
    """Main daemon loop - reads requests from stdin, executes, writes to stdout."""
    for line in sys.stdin:
        request = json.loads(line)
        response = handle_request(request)
        print(json.dumps(response))
        sys.stdout.flush()
```

### Phase 2: Go Process Pool Manager

**File:** `internal/bridge/pool.go`

**Features:**
- Singleton pool manager
- Lazy initialization (start on first use)
- Health checking (ping/health check)
- Automatic restart on errors
- Idle timeout (restart after inactivity)
- Graceful shutdown

**Key Types:**
```go
type PythonProcessPool struct {
    mu       sync.RWMutex
    process  *exec.Cmd
    stdin    io.WriteCloser
    stdout   io.ReadCloser
    scanner  *bufio.Scanner
    lastUsed time.Time
    created  time.Time
    healthy  bool
}

func (p *PythonProcessPool) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error)
func (p *PythonProcessPool) ensureHealthy(ctx context.Context) error
func (p *PythonProcessPool) restart(ctx context.Context) error
```

### Phase 3: Integration

**File:** `internal/bridge/python.go`

**Changes:**
- Add pool-based execution path
- Fallback to subprocess if pool unavailable
- Feature flag to enable/disable pool
- Backward compatibility maintained

**Usage:**
```go
// Try pool first (if enabled)
if poolEnabled {
    result, err := pool.ExecuteTool(ctx, toolName, args)
    if err == nil {
        return result, nil
    }
    // Fall back to subprocess on error
}

// Fallback to subprocess (current behavior)
return executePythonToolSubprocess(ctx, toolName, args)
```

---

## Design Decisions

### 1. JSON-RPC Protocol

**Why JSON-RPC 2.0:**
- Standard protocol with request/response correlation
- Error handling built-in
- Easy to implement and debug
- Line-delimited for simple parsing

**Alternative Considered:** Binary protobuf
- **Rejected:** More complex, harder to debug, current protobuf support is optional

### 2. Line-Delimited Protocol

**Why:**
- Simple to parse with `bufio.Scanner`
- No need for length prefixes
- Easy to debug (readable JSON)
- Works well with stdin/stdout

**Alternative Considered:** Length-prefixed binary
- **Rejected:** More complex, harder to debug

### 3. Single Process vs Pool

**Why Single Process:**
- Simpler implementation
- Sufficient for current workload
- Easier error handling
- Can scale to pool later if needed

**Alternative Considered:** Multiple processes in pool
- **Deferred:** Can add later if needed for concurrent requests

### 4. Health Checking

**Strategy:**
- Ping request every N calls or after idle timeout
- Restart if process dead or unresponsive
- Check process state (running, exited)

**Timeout:**
- Idle timeout: 5 minutes (restart after inactivity)
- Request timeout: 30 seconds (same as current)
- Health check timeout: 5 seconds

### 5. Error Handling

**Strategy:**
- If pool fails, fallback to subprocess (backward compatible)
- Log errors but don't block execution
- Automatic restart on process death
- Graceful degradation

---

## Performance Expectations

### Current (Subprocess Per Call)
- Process creation: ~50-200ms
- Python startup: ~100-300ms
- Module imports: ~200-500ms (first call)
- **Total overhead: ~350-1000ms**

### With Persistent Process
- First call: ~350-1000ms (startup + first execution)
- Subsequent calls: ~10-50ms (just execution)
- **Savings: ~300-950ms per call (after first)**

### With Model Caching (MLX Example)
- First call: ~2-5 seconds (model load)
- Subsequent calls: ~0.1-0.5 seconds (cached)
- **Savings: ~1.5-4.5 seconds per call (after first)**

---

## Implementation Steps

1. ✅ **Research and Design** (Current)
   - Document architecture
   - Define protocol
   - Plan integration

2. ⬜ **Python Daemon Script**
   - Create `bridge/execute_tool_daemon.py`
   - Implement JSON-RPC handler
   - Test with manual stdin/stdout

3. ⬜ **Go Process Pool**
   - Create `internal/bridge/pool.go`
   - Implement process management
   - Add health checking

4. ⬜ **Integration**
   - Update `ExecutePythonTool` to use pool
   - Add fallback to subprocess
   - Add feature flag

5. ⬜ **Testing**
   - Unit tests for pool
   - Integration tests
   - Performance benchmarks

6. ⬜ **Documentation**
   - Update bridge documentation
   - Add usage examples
   - Document configuration

---

## Configuration

**Environment Variables:**
- `EXARP_PYTHON_POOL_ENABLED=true` - Enable process pool (default: true)
- `EXARP_PYTHON_POOL_IDLE_TIMEOUT=5m` - Idle timeout before restart
- `EXARP_PYTHON_POOL_HEALTH_CHECK_INTERVAL=10` - Health check every N calls

**Feature Flag:**
- Can be disabled for debugging
- Automatic fallback to subprocess if pool fails

---

## Backward Compatibility

**Guaranteed:**
- All existing tools work unchanged
- Fallback to subprocess if pool unavailable
- Same error messages and behavior
- No breaking changes

**Optional:**
- Pool can be disabled via environment variable
- Subprocess path always available as fallback

---

## Future Enhancements

1. **Process Pool (Multiple Processes)**
   - Support concurrent requests
   - Load balancing across processes
   - Better resource utilization

2. **Request Batching**
   - Batch multiple tool calls
   - Reduce round-trips
   - Improve throughput

3. **Metrics and Monitoring**
   - Track pool health
   - Monitor request latency
   - Alert on errors

4. **Graceful Shutdown**
   - Finish in-flight requests
   - Clean process termination
   - Resource cleanup

---

## References

- `docs/ESTIMATION_TOOL_MODEL_CACHING_ANALYSIS.md` - Problem analysis
- `project_management_automation/scripts/base/mcp_client.py` - Session pool pattern
- `internal/bridge/python.go` - Current implementation
- `bridge/execute_tool.py` - Python bridge script

---

## Next Steps

1. Implement Python daemon script
2. Implement Go process pool
3. Integrate with existing bridge
4. Test and benchmark
5. Document and deploy
