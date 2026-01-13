# Cursor Background Agents Setup Guide

## Why Background Agents Sidebar Isn't Visible

The Background Agents sidebar needs to be explicitly enabled in Cursor settings. Here's how to enable it:

## Step 1: Enable Background Agents Feature

### Method 1: Via Settings UI

1. **Open Settings:**
   - Press `Cmd+Shift+P` (macOS) or `Ctrl+Shift+P` (Windows/Linux)
   - Type "Settings (UI)" and press Enter
   - Or go to: Cursor → Settings (or File → Preferences → Settings)

2. **Navigate to Beta Features:**
   - In the settings sidebar, find "Beta Features"
   - Or search for "Background Agents" in the settings search

3. **Enable Background Agents:**
   - Find the toggle: **"Cursor: Background Agents"**
   - Turn it **ON**

4. **Restart Cursor:**
   - Quit Cursor completely (Cmd+Q / Ctrl+Q)
   - Reopen Cursor

### Method 2: Via Command Palette

1. **Open Command Palette:**
   - Press `Cmd+Shift+P` (macOS) or `Ctrl+Shift+P` (Windows/Linux)

2. **Enable Feature:**
   - Type: "Preferences: Open Settings (UI)"
   - Navigate to: Beta Features → Cursor: Background Agents
   - Toggle ON

## Step 2: Access Background Agents

Once enabled, you can access Background Agents in several ways:

### Method 1: Keyboard Shortcut
- Press `Ctrl+E` (Windows/Linux) or `Cmd+E` (macOS)
- This opens/closes the Background Agents sidebar

### Method 2: Sidebar Button
- Look for a sidebar toggle button (usually in top-right corner)
- Click to open Background Agents sidebar
- Or use: `Ctrl+Shift+S` / `Cmd+Shift+S`

### Method 3: View Menu
- Go to: View → Background Agents
- Or: View → Show Background Agents

## Step 3: Verify Background Agents Are Enabled

### Check Settings
1. Open Settings (Cmd+, / Ctrl+,)
2. Search for "background agents"
3. Verify "Cursor: Background Agents" is enabled

### Check Sidebar
1. Look for Background Agents icon in sidebar
2. Or use keyboard shortcut: `Ctrl+E` / `Cmd+E`
3. Sidebar should appear on the right side

## Requirements

### Cursor Version
- Background Agents require a recent version of Cursor
- Check version: Help → About Cursor
- Update if needed: Help → Check for Updates

### Subscription Plan
- **Background Agents require Cursor Pro plan**
- Check your plan: Settings → Account
- Upgrade if needed: [Cursor Pricing](https://www.trycursor.com/en/pricing)

## Troubleshooting

### Background Agents Option Not Found

**Issue:** Can't find "Cursor: Background Agents" in settings

**Possible Causes:**
1. **Outdated Cursor version** - Update Cursor
2. **Not on Pro plan** - Upgrade to Pro
3. **Feature not available yet** - Check release notes

**Solution:**
1. Update Cursor to latest version
2. Verify you're on Cursor Pro plan
3. Check Cursor release notes for Background Agents availability

### Sidebar Button Not Appearing

**Issue:** Sidebar button or Background Agents icon not visible

**Solution:**
1. Enable in Settings first (Beta Features → Cursor: Background Agents)
2. Restart Cursor
3. Try keyboard shortcut: `Ctrl+E` / `Cmd+E`
4. Check View menu for "Background Agents" option

### Keyboard Shortcut Not Working

**Issue:** `Ctrl+E` / `Cmd+E` doesn't open Background Agents

**Solution:**
1. Check if shortcut is bound: Settings → Keyboard Shortcuts
2. Search for "Background Agents"
3. Verify or set custom shortcut
4. Try: `Ctrl+Shift+S` / `Cmd+Shift+S` (sidebar toggle)

### Settings UI Changed

**Note:** Cursor's interface has changed in recent versions:
- The "Agent" and "Editor" buttons were replaced with a small arrows button
- Located in top-right corner
- Use this button or `Cmd+E` / `Ctrl+E` to switch views

## Alternative: Using Remote MCP Servers

If Background Agents aren't available or you prefer a different approach, you can use **SSH-based remote MCP servers** instead:

### Setup Remote MCP Server

```json
{
  "mcpServers": {
    "remote-exarp-go": {
      "command": "ssh",
      "args": [
        "davids-mac-mini.tailf62197.ts.net",
        "/Users/davidl/Projects/exarp-go/bin/exarp-go"
      ],
      "description": "Remote MCP server via SSH"
    }
  }
}
```

**Advantages:**
- Works with any Cursor plan
- Direct control over remote execution
- Use remote hardware (MLX, Ollama, etc.)
- No separate sidebar needed

See `docs/REMOTE_MCP_SERVERS.md` for complete setup instructions.

## Quick Enable Checklist

- [ ] Open Settings (Cmd+, / Ctrl+,)
- [ ] Navigate to Beta Features
- [ ] Enable "Cursor: Background Agents"
- [ ] Verify Pro plan subscription
- [ ] Restart Cursor
- [ ] Press `Ctrl+E` / `Cmd+E` to open sidebar
- [ ] Verify Background Agents sidebar appears

## Additional Resources

- [Cursor Background Agents Documentation](https://docs.cursor.com/en/background-agents)
- [Cursor Community Forum](https://forum.cursor.com)
- [Remote MCP Servers Guide](./REMOTE_MCP_SERVERS.md)

