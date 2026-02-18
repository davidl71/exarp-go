# Cursor Plugin Submission Runbook

**Tag hints:** `#docs` `#cursor` `#plugin`

Step-by-step guide to validate and submit the exarp-go Cursor plugin. See [CURSOR_PLUGIN_PLAN.md](CURSOR_PLUGIN_PLAN.md) for context and [HUMAN_TASK_DEPENDENCIES.md](HUMAN_TASK_DEPENDENCIES.md) for task breakdown.

---

## 1. Pre-submission checklist

- [ ] Plugin manifest: `.cursor-plugin/plugin.json` with `name`, `displayName`, `description`, `version`, `license`, `keywords`
- [ ] MCP config: `mcp.json` at repo root with `exarp-go` server entry
- [ ] Binary builds: `make build` succeeds; `bin/exarp-go` exists and runs

---

## 2. Run validation

```bash
# Basic plugin structure (manifest + mcp.json)
make validate-plugin

# Or run script directly
bash scripts/validate-cursor-plugin.sh
```

If validation fails, fix reported issues before submitting.

---

## 3. Optional: cursor/plugin-template validation

The [cursor/plugin-template](https://github.com/cursor/plugin-template) uses a multi-plugin layout. exarp-go uses **single-plugin** (manifest at repo root). Our `validate-plugin` checks the structure we use.

To run the official template validator (for reference):

```bash
git clone --depth 1 https://github.com/cursor/plugin-template.git /tmp/plugin-template
cd /tmp/plugin-template
# Adapt paths — template expects plugins/*/ structure
node scripts/validate-template.mjs
```

Note: The template expects `.cursor-plugin/marketplace.json` and `plugins/*/`; exarp-go uses a single `.cursor-plugin/plugin.json` at root.

---

## 4. Submit to Cursor team

1. **Prepare repo:** Ensure plugin files are committed and pushed.
2. **Contact Cursor:** Per [Building Plugins](https://cursor.com/docs/plugins/building.md):
   - **Email:** kniparko@anysphere.com
   - **Slack:** If you have access to Cursor's Slack, message the plugins channel
3. **Include:**
   - Repository URL (e.g. `https://github.com/davidl71/exarp-go`)
   - Short description: "exarp-go MCP server for tasks, reports, session, automation"
   - Note: single-plugin layout; users must build `bin/exarp-go` or have it on PATH

---

## 5. After submission

- Wait for Cursor team response
- If approved, plugin will appear in [Cursor Marketplace](https://cursor.com/marketplace)
- Optional: Add MCP deeplink for one-click install per [MCP install links](https://cursor.com/docs/context/mcp/install-links.md)

---

## References

- [Cursor Plugins](https://cursor.com/docs/plugins)
- [Building Plugins](https://cursor.com/docs/plugins/building.md)
- [CURSOR_PLUGIN_PLAN.md](CURSOR_PLUGIN_PLAN.md) — plugin design and phases
