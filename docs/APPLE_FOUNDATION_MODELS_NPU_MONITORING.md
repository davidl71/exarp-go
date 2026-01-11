# Apple Foundation Models - NPU Monitoring Guide

**Date**: 2026-01-11  
**Status**: Testing Guide  
**Project**: exarp-go  
**Stream**: D - Activity Monitor Verification

---

## Overview

This guide documents the process for verifying that Apple Foundation Models integration properly utilizes the M4 NPU (Neural Processing Unit) and Neural Engine during API calls by monitoring Activity Monitor.

## Prerequisites

### Hardware Requirements
- **Mac**: M4 (or M3/M2/M1) Apple Silicon Mac
- **macOS**: 26.0+ (Tahoe) with Apple Intelligence enabled
- **Xcode**: 15.x or later (for Swift bridge)

### Software Requirements
- **exarp-go**: Built with Apple Foundation Models support
- **Swift Bridge**: Built for go-foundationmodels package
- **Activity Monitor**: Built-in macOS tool

## Setup Steps

### 1. Build Swift Bridge

```bash
# Build Swift bridge for go-foundationmodels
make build-swift-bridge
```

**Expected Output:**
```
✅ Swift bridge built in vendor directory
```

### 2. Build exarp-go with Apple FM Support

```bash
# Build with Apple Foundation Models support (CGO_ENABLED=1)
make build-apple-fm
```

**Expected Output:**
```
✅ Server built with Apple FM support: bin/exarp-go
```

### 3. Verify Platform Support

```bash
# Test platform detection
./bin/exarp-go -test apple_foundation_models
```

**Expected Output:**
```
✅ Test passed!
Platform supports Apple Foundation Models
```

## Testing Procedure

### Step 1: Open Activity Monitor

1. **Launch Activity Monitor**:
   ```bash
   open -a "Activity Monitor"
   ```

2. **Enable GPU History View**:
   - Go to **Window** → **GPU History**
   - Ensure GPU History window is visible and positioned for monitoring

3. **Observe Baseline**:
   - Note current GPU/NPU activity levels
   - Verify NPU graph is visible (should show separate NPU activity)

### Step 2: Prepare Test Script

Create a test script to run multiple API calls:

```bash
# test_apple_fm_npu.sh
#!/bin/bash

echo "Testing Apple Foundation Models NPU Utilization"
echo "================================================"

# Test 1: Simple generation
echo -e "\n[Test 1] Simple text generation..."
./bin/exarp-go -tool apple_foundation_models -args '{"action":"generate","prompt":"Write a short poem about artificial intelligence","max_tokens":100}'

sleep 5

# Test 2: Summarization
echo -e "\n[Test 2] Text summarization..."
./bin/exarp-go -tool apple_foundation_models -args '{"action":"summarize","prompt":"Apple Foundation Models provide on-device AI capabilities for Macs with Apple Silicon. They utilize the Neural Engine and NPU for efficient inference.","max_tokens":50}'

sleep 5

# Test 3: Classification
echo -e "\n[Test 3] Text classification..."
./bin/exarp-go -tool apple_foundation_models -args '{"action":"classify","prompt":"I love this product!","categories":"positive,negative,neutral","max_tokens":10}'

sleep 5

# Test 4: Longer generation
echo -e "\n[Test 4] Longer text generation..."
./bin/exarp-go -tool apple_foundation_models -args '{"action":"generate","prompt":"Explain how neural networks work in simple terms","max_tokens":256,"temperature":0.8}'

sleep 5

# Test 5: Creative mode
echo -e "\n[Test 5] Creative text generation..."
./bin/exarp-go -tool apple_foundation_models -args '{"action":"generate","prompt":"Write a creative story about a robot learning to paint","max_tokens":200,"temperature":0.9}'

echo -e "\n✅ All tests completed"
```

**Make executable:**
```bash
chmod +x test_apple_fm_npu.sh
```

### Step 3: Run Tests While Monitoring

1. **Start Activity Monitor GPU History** (already open from Step 1)

2. **Run test script**:
   ```bash
   ./test_apple_fm_npu.sh
   ```

3. **Observe NPU Activity**:
   - Watch for NPU spikes during each API call
   - Note the timing correlation between API calls and NPU activity
   - Verify Neural Engine activity (should show alongside NPU)

### Step 4: Alternative - Manual MCP Testing

If you prefer to test via MCP server:

1. **Start exarp-go MCP server**:
   ```bash
   ./bin/exarp-go
   ```

2. **Connect via MCP client** (e.g., Cursor IDE chat):
   - Use `apple_foundation_models` tool with various prompts
   - Monitor Activity Monitor during each call

**Example MCP Tool Calls:**

```json
{
  "name": "apple_foundation_models",
  "arguments": {
    "action": "generate",
    "prompt": "Explain quantum computing",
    "max_tokens": 150
  }
}
```

## Expected Results

### NPU Utilization Indicators

1. **NPU Activity Spikes**:
   - NPU graph should show significant activity during inference
   - Activity should correlate with API call execution time
   - NPU usage should drop after call completion

2. **Neural Engine Activity**:
   - Should show alongside NPU activity
   - Indicates proper hardware utilization
   - Activity should match inference duration

3. **CPU/GPU Activity**:
   - CPU activity should be minimal (models run on NPU)
   - GPU activity should be low (not using Metal for inference)
   - NPU handles the heavy lifting

### Success Criteria

✅ **NPU utilization confirmed if:**
- NPU activity spikes visible during API calls
- Activity correlates with inference timing
- Neural Engine shows concurrent activity
- CPU/GPU activity remains low
- Response times are fast (<2 seconds for typical prompts)

### Failure Indicators

❌ **Potential issues if:**
- No NPU activity during API calls
- High CPU usage instead of NPU
- API calls fail or timeout
- "Apple Foundation Models not supported" errors

## Troubleshooting

### Issue: No NPU Activity Visible

**Possible Causes:**
1. **Swift bridge not built**: Build with `make build-swift-bridge`
2. **CGO disabled**: Rebuild with `make build-apple-fm` (enables CGO)
3. **Apple Intelligence not enabled**: Check System Settings
4. **Wrong macOS version**: Requires macOS 26.0+

**Solutions:**
```bash
# Rebuild everything
make build-swift-bridge
make build-apple-fm

# Verify platform support
./bin/exarp-go -test apple_foundation_models
```

### Issue: Activity Monitor Not Showing GPU History

**Solution:**
- Ensure macOS version is recent enough (macOS 26.0+)
- Check Window menu for "GPU History" option
- Try restarting Activity Monitor

### Issue: API Calls Fail

**Check:**
1. Platform support: `./bin/exarp-go -test apple_foundation_models`
2. Build configuration: Ensure `CGO_ENABLED=1` during build
3. Swift bridge: Verify `libFMShim.a` exists in vendor directory
4. Apple Intelligence: Check System Settings

## Documentation Requirements

### What to Document

1. **Baseline Observations**:
   - NPU activity levels when idle
   - System resources before testing

2. **Test Results**:
   - For each API call:
     - Prompt used
     - NPU activity level (low/medium/high)
     - Duration of NPU spike
     - Response received
     - Any anomalies

3. **Screenshots** (if possible):
   - Activity Monitor GPU History during inference
   - NPU graph showing activity spike
   - CPU/GPU graphs for comparison

4. **Timing Data**:
   - API call duration
   - NPU utilization duration
   - Correlation between call start and NPU spike

### Example Documentation Format

```markdown
## Test Results - [Date]

### Test 1: Simple Generation
- **Prompt**: "Write a short poem about AI"
- **NPU Activity**: High spike during 1.2s inference
- **Neural Engine**: Active alongside NPU
- **CPU Usage**: Minimal (<5%)
- **Result**: ✅ Success - NPU utilized correctly

### Test 2: Summarization
- **Prompt**: [Text to summarize]
- **NPU Activity**: [Observations]
- **Result**: [Success/Failure]

[Continue for all tests...]
```

## Next Steps

After completing verification:

1. ✅ Document findings in task result comment
2. ✅ Note any issues or anomalies
3. ✅ Update this guide with any new insights
4. ✅ Share results with team

## References

- [Apple Foundation Models Integration](./APPLE_FOUNDATION_MODELS_INTEGRATION.md)
- [Apple Foundation Models Testing](./APPLE_FOUNDATION_MODELS_TESTING.md)
- [Apple Foundation Models Implementation](./APPLE_FOUNDATION_MODELS_IMPLEMENTATION.md)

---

**Status**: Ready for testing  
**Last Updated**: 2026-01-11
