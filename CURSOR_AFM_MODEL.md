# Adding AFM as Custom Model in Cursor

This guide shows how to add the Apple Foundation Models (AFM) service as a custom model in Cursor IDE.

## Prerequisites

- ✅ AFM service running (via launchd service)
- ✅ Tailscale serve configured (optional, for remote access)
- ✅ Cursor IDE installed

## Configuration Steps

### Option 1: Via Cursor Settings UI (Recommended)

1. **Open Cursor Settings**
   - Press `Cmd + ,` (or `Ctrl + ,` on Windows/Linux)
   - Or: Click gear icon → Settings

2. **Navigate to Models**
   - In the settings sidebar, search for "Models"
   - Or go to: `Cursor Settings` → `Models`

3. **Add Custom Model**
   - Click "+ Add Model" or "Add Custom Model"
   - Fill in the following:
     - **Model Name:** `Apple Foundation Models` (or `AFM`)
     - **API Type:** `OpenAI Compatible`
     - **Base URL:** 
       - Local: `http://localhost:9999/v1`
       - Tailscale: `http://davids-mac-mini.tailf62197.ts.net/v1`
     - **Model ID:** `foundation`
     - **API Key:** (leave empty - AFM doesn't require authentication)

4. **Test Connection**
   - Click "Test Connection" to verify it works
   - If successful, click "Save"

5. **Select the Model**
   - In Cursor Chat or `Cmd/Ctrl + K`, use the model dropdown
   - Select "Apple Foundation Models" or "AFM"

### Option 2: Manual Configuration (Advanced)

If you prefer to configure manually via settings.json:

```json
{
  "cursor.models": {
    "custom": [
      {
        "name": "Apple Foundation Models",
        "provider": "openai",
        "baseUrl": "http://localhost:9999/v1",
        "model": "foundation",
        "apiKey": ""
      }
    ]
  }
}
```

Or for Tailscale access:

```json
{
  "cursor.models": {
    "custom": [
      {
        "name": "Apple Foundation Models (Tailscale)",
        "provider": "openai",
        "baseUrl": "http://davids-mac-mini.tailf62197.ts.net/v1",
        "model": "foundation",
        "apiKey": ""
      }
    ]
  }
}
```

## Verification

### Test the API Endpoint

Before adding to Cursor, verify the endpoint works:

```bash
# Health check
curl http://localhost:9999/health

# List models
curl http://localhost:9999/v1/models

# Test chat completion
curl -X POST http://localhost:9999/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "foundation",
    "messages": [
      {"role": "user", "content": "Hello, test message"}
    ]
  }'
```

### Test in Cursor

1. Open Cursor Chat (`Cmd/Ctrl + L`)
2. Select your custom "Apple Foundation Models" from the model dropdown
3. Send a test message
4. Verify you get a response from the local AFM service

## Configuration Options

### Local Access (Recommended for Development)

- **Base URL:** `http://localhost:9999/v1`
- **Pros:** Fast, no network dependency
- **Cons:** Only works on this machine

### Tailscale Access (Recommended for Remote)

- **Base URL:** `http://davids-mac-mini.tailf62197.ts.net/v1`
- **Pros:** Accessible from any device in your Tailscale network
- **Cons:** Requires Tailscale connection

### Both Configurations

You can add both configurations and switch between them:

```json
{
  "cursor.models": {
    "custom": [
      {
        "name": "AFM Local",
        "provider": "openai",
        "baseUrl": "http://localhost:9999/v1",
        "model": "foundation"
      },
      {
        "name": "AFM Tailscale",
        "provider": "openai",
        "baseUrl": "http://davids-mac-mini.tailf62197.ts.net/v1",
        "model": "foundation"
      }
    ]
  }
}
```

## Troubleshooting

### Model Not Appearing

1. **Check Service Status**
   ```bash
   # Check if AFM service is running
   launchctl list | grep maclocal
   curl http://localhost:9999/health
   ```

2. **Check Tailscale Serve** (if using Tailscale URL)
   ```bash
   tailscale serve status
   ```

3. **Verify Model ID**
   ```bash
   curl http://localhost:9999/v1/models
   # Should return: {"id": "foundation", ...}
   ```

### Connection Errors

1. **"Connection refused"**
   - AFM service might not be running
   - Check: `launchctl list | grep maclocal`
   - Restart: `launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.maclocal.afm.plist`

2. **"Model not found"**
   - Verify model ID is `foundation`
   - Check: `curl http://localhost:9999/v1/models`

3. **"Timeout"**
   - AFM might be processing a long request
   - Check logs: `tail -f /tmp/afm.log`

### Performance Issues

- **Slow responses:** Apple Foundation Models run on-device, so responses may be slower than cloud models
- **Resource usage:** Check Activity Monitor for CPU/Memory usage
- **Concurrent requests:** AFM may handle one request at a time

## Advanced Configuration

### Custom Instructions

You can set custom system instructions in the model configuration:

```json
{
  "cursor.models": {
    "custom": [
      {
        "name": "Apple Foundation Models",
        "provider": "openai",
        "baseUrl": "http://localhost:9999/v1",
        "model": "foundation",
        "systemPrompt": "You are a helpful coding assistant."
      }
    ]
  }
}
```

### Temperature and Other Parameters

Some Cursor configurations allow setting default parameters:

```json
{
  "cursor.models": {
    "custom": [
      {
        "name": "Apple Foundation Models",
        "provider": "openai",
        "baseUrl": "http://localhost:9999/v1",
        "model": "foundation",
        "temperature": 0.7,
        "maxTokens": 2000
      }
    ]
  }
}
```

## Integration with exarp-go

Once configured in Cursor, you can use AFM as an alternative to your direct `go-foundationmodels` integration:

- **Direct Library:** Fast, no HTTP overhead, integrated in Go code
- **HTTP API:** More flexible, can switch backends, works with any language

Both approaches are valid - choose based on your needs!

## Notes

- **Privacy:** All processing happens locally on your Mac - no data sent to external servers
- **Performance:** On-device models may be slower than cloud models but offer better privacy
- **Limitations:** Apple Foundation Models are optimized for on-device use and may have different capabilities than cloud models
- **Cost:** Free! No API costs for local models

