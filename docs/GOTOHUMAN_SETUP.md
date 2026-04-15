# gotoHuman Setup Guide

## Quick Start

### 1. Get API Key

1. Visit gotoHuman platform (check their website/documentation)
2. Sign up or log in to your account
3. Navigate to API settings
4. Generate a new API key
5. Copy the API key

### 2. Configure API Key

**Option A: Environment Variable (System-wide)**

Add to your shell profile (`~/.zshrc` or `~/.bashrc`):
```bash
export GOTOHUMAN_API_KEY="your-api-key-here"
```

Then restart your terminal or run:
```bash
source ~/.zshrc  # or ~/.bashrc
```

**Option B: MCP Configuration (Project-specific)**

Update `.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "gotohuman": {
      "command": "uvx",
      "args": [...],
      "env": {
        "GOTOHUMAN_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

**⚠️ Security Note:** Don't commit API keys to git! Use environment variables or `.cursor/mcp.json` (which should be in `.gitignore`).

### 3. Restart Cursor

After setting the API key:
1. Quit Cursor completely (Cmd+Q)
2. Reopen Cursor
3. Verify gotoHuman server loads without errors

### 4. Test Configuration

In Cursor chat, try:
```
@gotoHuman list-forms
```

If successful, you should see a list of available forms.

## Troubleshooting

### "API key is required" Error

**Solution:** Ensure API key is set:
- Check environment variable: `echo $GOTOHUMAN_API_KEY`
- Check MCP config: Verify `env.GOTOHUMAN_API_KEY` in `.cursor/mcp.json`
- Restart Cursor after setting API key

### Server Not Loading

**Solution:**
1. Check Cursor MCP logs (Settings → Features → Model Context Protocol)
2. Verify `uvx` is installed: `which uvx`
3. Verify npm is available: `which npx`
4. Test wrapper manually: `uvx mcpower-proxy==0.0.87 --help`

### Tools Not Available

**Solution:**
1. Verify API key is correct
2. Check gotoHuman account status
3. Verify internet connection (gotoHuman may require online access)
4. Check gotoHuman service status

## Next Steps

Once API key is configured:
1. Discover available forms
2. Test approval request
3. Integrate with Todo2 workflow

---

**Last Updated:** 2026-01-07

