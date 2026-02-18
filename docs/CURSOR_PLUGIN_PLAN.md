# Plan: exarp-go as a Cursor Plugin

**Tag hints:** `#docs` `#planning` `#cursor` `#plugin`

**Status:** Draft  
**Created:** 2026-02-17

---

## 1. Purpose & Success Criteria

**Purpose:** Package exarp-go as a [Cursor Plugin](https://cursor.com/docs/plugins) so users can install it from the Cursor Marketplace (or via MCP deeplink) and get the exarp-go MCP server, optional rules, and optional skills in one bundle.

**Success criteria:**

- A planning doc and sketch exist (this file).
- A plugin layout (`.cursor-plugin/` + manifest) is defined; optional: starter files in-repo or in a separate plugin repo.
- Todo2 tasks track: manifest authoring, validation, and submission.
- Users can install “exarp-go” as a plugin (project or user scope) and have the MCP server (and any bundled rules/skills) available without manual MCP config.

**Context:** Cursor 2.5 (Feb 2026) introduced [Plugins and the Marketplace](https://cursor.com/changelog/2-5). Plugins bundle rules, skills, agents, commands, MCP servers, and hooks ([Cursor Plugins docs](https://cursor.com/docs/plugins)). exarp-go is already an MCP server; packaging it as a plugin improves discoverability and one-click install.

---

## 2. Plugin Structure Sketch

### 2.1 Manifest (`.cursor-plugin/plugin.json`)

Required fields per Cursor’s plugin model:

| Field         | Example value |
|---------------|----------------|
| `name`        | `exarp-go` (lowercase kebab-case, unique) |
| `displayName` | `exarp-go` |
| `author`      | Your / org name |
| `description` | Task lifecycle, reports, session, and automation MCP server for Cursor and OpenCode. |
| `keywords`    | `["tasks", "todo2", "report", "session", "mcp", "automation"]` |
| `license`     | Same as repo (e.g. MIT) |
| `version`     | Semver, e.g. `1.0.0` |

Optional: `repository`, `homepage`, etc. if the docs support them.

### 2.2 Directory Layout (single-plugin)

```
exarp-go/                          # or a dedicated plugin repo
├── .cursor-plugin/
│   └── plugin.json                # manifest (above)
├── mcp/                           # or embedded in plugin.json / config dir
│   └── exarp-go.json              # MCP server config (command, args, env)
├── rules/                         # optional: bundle key .mdc rules
│   └── (optional rules that reference exarp-go)
├── skills/                        # optional: skills that use exarp-go tools
│   └── (optional)
└── ...
```

**MCP config in plugin:** The plugin can ship an MCP entry that points to the user’s `exarp-go` binary (e.g. `bin/exarp-go` in the project where the plugin is installed, or a well-known path). Cursor merges plugin MCP config into the user’s MCP list. Use `{{PROJECT_ROOT}}` (or equivalent) so the binary path is resolved per project if desired.

**Binary distribution:** Plugins are Git repos; the exarp-go binary is typically built by the user or installed separately. The plugin config points at that binary (e.g. `$HOME/.local/bin/exarp-go` or `./bin/exarp-go` when the plugin is used inside the exarp-go repo). Alternatively, document “install exarp-go binary first” in the plugin description.

### 2.3 Multi-plugin vs single-plugin

- **Single-plugin:** One `.cursor-plugin/plugin.json` at repo root; plugin contents (mcp/, rules/, skills/) at repo root. Fits “exarp-go repo is the plugin” or “exarp-go-plugin” repo that references the binary.
- **Multi-plugin:** Use `.cursor-plugin/marketplace.json` at root and one `plugin.json` per plugin under `plugins/*/`. Use if we later ship multiple plugins (e.g. exarp-go + a “exarp-go rules pack”).

Start with **single-plugin** for exarp-go.

---

## 3. Implementation Phases

### Phase 1: Document and sketch (this doc)

- [x] Create planning doc with purpose, sketch, and phases.
- [x] Sketch manifest fields and directory layout.
- [ ] Add a short “Cursor Plugin” subsection to README or docs index that links here.

### Phase 2: Manifest and config

- [x] Add `.cursor-plugin/plugin.json` (in-repo or in a `cursor-plugin/` subdir for clarity).
- [x] Add MCP server config (`mcp.json` at repo root) that points at exarp-go binary with `{{PROJECT_ROOT}}/bin/exarp-go` and `PROJECT_ROOT` env. For projects other than exarp-go repo, user must install binary and adjust path.
- [ ] Optional: bundle 1–2 rules (e.g. task-workflow, session) as part of the plugin.

### Phase 3: Validate and submit

- [x] Run `make validate-plugin` for local structure validation (see [CURSOR_PLUGIN_SUBMISSION.md](CURSOR_PLUGIN_SUBMISSION.md))
- [ ] Submit plugin to Cursor team for marketplace (per [Building Plugins](https://cursor.com/docs/plugins/building.md)).
- [ ] Optional: Add MCP deeplink for one-click install ([MCP install links](https://cursor.com/docs/context/mcp/install-links.md)).

---

## 4. Open Questions

- Whether the plugin lives in the main exarp-go repo (e.g. `.cursor-plugin/` at root) or in a separate “exarp-go-cursor-plugin” repo that references the binary.
- How to resolve binary path across platforms (Makefile/build produces `bin/exarp-go`; user may install to `$HOME/.local/bin` or use `go install`).
- Whether to bundle any rules/skills by default or keep the plugin MCP-only.

---

## 5. References

- [Cursor Plugins](https://cursor.com/docs/plugins) — What plugins contain, marketplace, installing, managing.
- [Building Plugins](https://cursor.com/docs/plugins/building.md) — Manifest, directory structure, submission.
- [Cursor Changelog 2.5](https://cursor.com/changelog/2-5) — Plugins, sandbox controls, async subagents.
- [Cursor Marketplace](https://cursor.com/marketplace) — Browse and install plugins.
- [MCP install links](https://cursor.com/docs/context/mcp/install-links.md) — Deeplinks for MCP config.
- [cursor/plugin-template](https://github.com/cursor/plugin-template) — Official template (simple + advanced).
- [CURSOR_MCP_SETUP.md](CURSOR_MCP_SETUP.md) — Current manual MCP setup for exarp-go.
