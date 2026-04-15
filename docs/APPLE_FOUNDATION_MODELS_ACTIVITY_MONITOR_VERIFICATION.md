# Apple Foundation Models - Activity Monitor Verification Guide

**Date**: 2026-01-12  
**Task**: AFM-1767880652340-3  
**Purpose**: Verify Neural Engine (NPU) utilization during Apple Foundation Models API calls

## Overview

This guide describes how to monitor Activity Monitor during Apple Foundation Models API calls to verify that the Neural Engine (NPU) is being utilized on Apple M4 devices.

## Prerequisites

1. **macOS 26.0+ (Tahoe)** - Required for Apple Foundation Models
2. **Apple Silicon (arm64)** - M1, M2, M3, or M4 chip
3. **Apple Intelligence enabled** - Required for Foundation Models API
4. **Xcode installed** - Required for Swift bridge compilation
5. **Swift bridge built** - Run `make build-swift-bridge` first

## Setup

### 1. Build with Apple Foundation Models Support

```bash
cd /Users/davidl/Projects/exarp-go

# Build Swift bridge (if not already built)
make build-swift-bridge

# Build server with Apple FM support
make build-apple-fm
```

### 2. Start the MCP Server

```bash
# Run the server
./bin/exarp-go
```

### 3. Open Activity Monitor

1. Open **Activity Monitor** (Applications → Utilities)
2. **Window → GPU History** (or press `Cmd+4`)
3. Keep Activity Monitor visible during testing

## Verification Process

### Step 1: Prepare Test Query

Use a simple test query for Apple Foundation Models:

```json
{
  "action": "generate",
  "prompt": "Write a brief summary of machine learning fundamentals.",
  "max_tokens": 100
}
```

### Step 2: Monitor Activity Monitor

**Before API call:**
- Note baseline GPU/CPU usage
- Note memory usage

**During API call:**
1. **Run the Apple Foundation Models tool** via MCP client (e.g., Cursor)
2. **Observe Activity Monitor**:
   - **GPU History**: Should show activity (Neural Engine may appear as GPU)
   - **CPU Usage**: Should be relatively low (coordination only)
   - **Memory Usage**: May increase slightly (model loading)

**After API call:**
- GPU/CPU usage should return to baseline
- Memory may remain elevated if model stays loaded

### Step 3: Expected Behavior

**Neural Engine (NPU) Usage:**
- Neural Engine activity may appear as GPU activity in Activity Monitor
- This is expected - both use unified memory architecture
- Activity should be visible during inference

**Performance Indicators:**
- **Low CPU usage** (<30%) - Indicates hardware acceleration
- **GPU activity** during inference - May indicate NPU/GPU usage
- **Fast response times** - Hardware acceleration working

### Step 4: Verification Checklist

- [ ] GPU History shows activity during API call
- [ ] CPU usage remains relatively low (<50%)
- [ ] API calls complete successfully
- [ ] Response times are reasonable (<5 seconds for simple queries)
- [ ] No errors in server logs

## Test Script

Create a test script to automate the API call (manual Activity Monitor monitoring still required):

```bash
#!/bin/bash
# test_apple_fm_verification.sh

echo "Testing Apple Foundation Models API..."
echo "Open Activity Monitor → GPU History before running this script"
echo "Press Enter to continue..."
read

# Run a test query via the tool
# (This would be done via MCP client in practice)
echo "Running test query..."
# Test query would go here

echo "Check Activity Monitor for GPU/Neural Engine activity"
```

## Troubleshooting

### No Activity in Activity Monitor

**Possible causes:**
1. Swift bridge not built - Run `make build-swift-bridge`
2. Platform not supported - Verify macOS version and architecture
3. Apple Intelligence not enabled - Check system settings
4. Tool not working - Check server logs for errors

**Solutions:**
```bash
# Verify build
make build-apple-fm

# Check platform support
go test ./internal/platform -v

# Check server logs
./bin/exarp-go 2>&1 | grep -i "foundation\|apple"
```

### High CPU Usage

If CPU usage is high (>70%), it may indicate:
- Software fallback instead of hardware acceleration
- Platform not fully supported
- Check server logs for platform detection errors

### API Calls Fail

If API calls fail:
1. Check platform support: `go test ./internal/platform -v`
2. Verify Swift bridge: `ls vendor/github.com/blacktop/go-foundationmodels/libFMShim.a`
3. Check server logs for detailed error messages
4. Verify Apple Intelligence is enabled in system settings

## Notes

### Neural Engine Detection

**Important:** Neural Engine activity may appear as GPU activity in Activity Monitor. This is expected behavior because:

1. **Unified Memory Architecture**: Neural Engine and GPU share memory
2. **Activity Monitor Display**: May group NPU activity with GPU activity
3. **Hardware Integration**: Both use Metal framework for communication

**What to look for:**
- GPU activity during inference (may include NPU)
- Low CPU usage (indicates hardware acceleration)
- Fast response times (hardware acceleration working)

### Performance Benchmarks

Expected performance on M4:
- **Response time**: 1-5 seconds for simple queries
- **CPU usage**: <30% during inference
- **Memory usage**: <1GB additional for model
- **Throughput**: Varies based on query complexity

## Summary

To verify Neural Engine utilization:

1. ✅ Build server with Apple FM support
2. ✅ Open Activity Monitor (GPU History)
3. ✅ Run Apple Foundation Models API calls
4. ✅ Observe GPU activity during inference
5. ✅ Verify low CPU usage (<50%)
6. ✅ Confirm successful API responses

**Status**: Verification process documented. Manual testing required to complete verification.
