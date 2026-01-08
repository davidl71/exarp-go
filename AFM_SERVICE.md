# AFM Service Management

`afm` is now configured as a launchd service and will start automatically on login.

## Service Status

**Service Name:** `com.maclocal.afm`  
**Status:** ✅ Running  
**Port:** 9999  
**Host:** 127.0.0.1

## Managing the Service

### Check Status
```bash
# Check if service is running
launchctl list | grep maclocal

# Get detailed service info
launchctl print gui/$(id -u)/com.maclocal.afm

# Test the API
curl http://localhost:9999/health
```

### Start Service
```bash
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.maclocal.afm.plist
```

### Stop Service
```bash
launchctl bootout gui/$(id -u)/com.maclocal.afm
```

### Restart Service
```bash
launchctl bootout gui/$(id -u)/com.maclocal.afm
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.maclocal.afm.plist
```

### View Logs
```bash
# Standard output
tail -f /tmp/afm.log

# Error output
tail -f /tmp/afm.error.log
```

## API Endpoints

Once the service is running, you can access:

- **Health Check:** `GET http://localhost:9999/health`
- **List Models:** `GET http://localhost:9999/v1/models`
- **Chat Completions:** `POST http://localhost:9999/v1/chat/completions`

## Example API Usage

```bash
# Health check
curl http://localhost:9999/health

# List available models
curl http://localhost:9999/v1/models

# Chat completion
curl -X POST http://localhost:9999/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "foundation",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ]
  }'
```

## Configuration

The service configuration is in:
`~/Library/LaunchAgents/com.maclocal.afm.plist`

To change the port or other settings, edit the plist file and restart the service:
```bash
# Edit the plist
nano ~/Library/LaunchAgents/com.maclocal.afm.plist

# Restart service
launchctl bootout gui/$(id -u)/com.maclocal.afm
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.maclocal.afm.plist
```

## Auto-Start

The service is configured with `RunAtLoad: true`, so it will automatically start:
- When you log in
- After system restarts
- If it crashes (KeepAlive: true)

## Integration with exarp-go

You can now use `afm` as an HTTP backend for your Apple Foundation Models tool:

```go
// Example: Use HTTP client instead of direct library
client := &http.Client{Timeout: 30 * time.Second}
resp, err := client.Post(
    "http://localhost:9999/v1/chat/completions",
    "application/json",
    bytes.NewBuffer(jsonData),
)
```

Or continue using the direct `go-foundationmodels` library (current approach).

