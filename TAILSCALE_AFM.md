# Tailscale Serve Configuration for AFM

`afm` is now accessible through your Tailscale network via Tailscale Serve.

## Access URLs

**Tailnet Only (Private):**
- `http://davids-mac-mini.tailf62197.ts.net/`
- `http://davids-mac-mini/`

**Note:** These URLs are only accessible to devices in your Tailscale tailnet.

## Current Configuration

```bash
# Status
tailscale serve status

# Current setup
http://davids-mac-mini.tailf62197.ts.net/
|-- / proxy http://localhost:9999
```

## Managing Tailscale Serve

### View Status
```bash
tailscale serve status
```

### Stop Serve
```bash
tailscale serve --http=80 off
```

### Restart Serve
```bash
# Stop
tailscale serve --http=80 off

# Start
tailscale serve --bg --http 80 http://localhost:9999
```

### Reset All Serve Configs
```bash
tailscale serve reset
```

## API Endpoints (via Tailscale)

Once configured, you can access all AFM endpoints through Tailscale:

### Health Check
```bash
curl http://davids-mac-mini.tailf62197.ts.net/health
```

### List Models
```bash
curl http://davids-mac-mini.tailf62197.ts.net/v1/models
```

### Chat Completions
```bash
curl -X POST http://davids-mac-mini.tailf62197.ts.net/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "foundation",
    "messages": [
      {"role": "user", "content": "Hello from Tailscale!"}
    ]
  }'
```

## Making it Public (Optional)

If you want to expose the service to the internet (not just your tailnet), use Tailscale Funnel:

```bash
# Enable funnel (makes it publicly accessible)
tailscale funnel --bg --http 80 http://localhost:9999

# Check funnel status
tailscale funnel status

# Disable funnel (back to tailnet only)
tailscale funnel --http=80 off
```

**⚠️ Security Warning:** Funnel makes your service publicly accessible on the internet. Only enable if you understand the security implications.

## Auto-Start on Boot

To make Tailscale serve start automatically, you can:

### Option 1: Add to LaunchAgent (Recommended)

Create a LaunchAgent that starts Tailscale serve after the afm service:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.tailscale.afm-serve</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>/opt/homebrew/bin/tailscale</string>
        <string>serve</string>
        <string>--bg</string>
        <string>--http</string>
        <string>80</string>
        <string>http://localhost:9999</string>
    </array>
    
    <key>RunAtLoad</key>
    <true/>
    
    <key>StartInterval</key>
    <integer>300</integer>
    
    <key>StandardOutPath</key>
    <string>/tmp/tailscale-serve.log</string>
    
    <key>StandardErrorPath</key>
    <string>/tmp/tailscale-serve.error.log</string>
</dict>
</plist>
```

### Option 2: Add to Shell Profile

Add to `~/.zshrc`:
```bash
# Start Tailscale serve for AFM (if not already running)
if ! tailscale serve status 2>&1 | grep -q "proxy http://localhost:9999"; then
    tailscale serve --bg --http 80 http://localhost:9999
fi
```

## Integration with exarp-go

You can now use the Tailscale URL in your Go code:

```go
// Use Tailscale URL instead of localhost
const afmURL = "http://davids-mac-mini.tailf62197.ts.net"

// Or make it configurable
afmURL := os.Getenv("AFM_URL")
if afmURL == "" {
    afmURL = "http://localhost:9999" // fallback to local
}

client := &http.Client{Timeout: 30 * time.Second}
resp, err := client.Post(
    afmURL + "/v1/chat/completions",
    "application/json",
    bytes.NewBuffer(jsonData),
)
```

## Troubleshooting

### Service Not Accessible
1. Check if afm service is running:
   ```bash
   curl http://localhost:9999/health
   ```

2. Check Tailscale serve status:
   ```bash
   tailscale serve status
   ```

3. Check Tailscale connection:
   ```bash
   tailscale status
   ```

### Port Already in Use
If port 80 is in use, use a different port:
```bash
tailscale serve --bg --http 8080 http://localhost:9999
```

Then access via: `http://davids-mac-mini.tailf62197.ts.net:8080`

### View Logs
```bash
# Tailscale serve logs
tail -f /tmp/tailscale-serve.log

# AFM service logs
tail -f /tmp/afm.log
```

## Security Notes

- **Tailnet Only:** By default, Tailscale serve is only accessible within your tailnet (private network)
- **No Authentication:** The AFM API itself doesn't require authentication - rely on Tailscale's network security
- **HTTPS:** Tailscale automatically provides HTTPS for funnel (public) access
- **Access Control:** Use Tailscale ACLs to control which devices/users can access the service

