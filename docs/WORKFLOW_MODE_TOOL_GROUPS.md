# Workflow Mode Tool Groups

**Date:** 2026-01-11  
**Purpose:** Documentation for tool groups enable/disable functionality in the workflow_mode tool

---

## Overview

The `workflow_mode` tool provides functionality to enable and disable tool groups, allowing fine-grained control over which tools are available in different contexts. This feature helps reduce context size and focus AI agents on relevant tools for specific workflows.

## Available Tool Groups

The following tool groups can be enabled or disabled:

### Always Required (Cannot Be Disabled)

- **`core`** - Core system tools that are always available
- **`tool_catalog`** - Tool catalog and discovery tools (including `workflow_mode` itself)

### Optional Groups

- **`health`** - Project health monitoring tools (scorecard, overview, documentation health)
- **`tasks`** - Task management tools (alignment, duplicates, sync, approval)
- **`security`** - Security scanning and vulnerability assessment tools
- **`automation`** - Automation and workflow tools (daily automation, git hooks)
- **`config`** - Configuration management tools (cursor rules, ignore files)
- **`testing`** - Testing and quality assurance tools
- **`advisors`** - Advisor consultation tools
- **`memory`** - Memory and context management tools
- **`workflow`** - Workflow recommendation and mode inference tools
- **`prd`** - Product requirements document tools

## Usage

### Enable a Tool Group

To enable a specific tool group:

```json
{
  "action": "focus",
  "enable_group": "memory"
}
```

**Behavior:**
- Removes the group from `disabled_groups` if present
- Adds the group to `extra_groups` if not already present
- State is persisted to `.exarp/workflow_mode.json`

### Disable a Tool Group

To disable a specific tool group:

```json
{
  "action": "focus",
  "disable_group": "testing"
}
```

**Behavior:**
- Removes the group from `extra_groups` if present
- Adds the group to `disabled_groups` if not already present
- State is persisted to `.exarp/workflow_mode.json`
- **Error:** Cannot disable `core` or `tool_catalog` groups

### Check Current Status

To view the current state of enabled/disabled groups:

```json
{
  "action": "focus",
  "status": true
}
```

**Response includes:**
- Current workflow mode
- Extra groups (explicitly enabled)
- Disabled groups (explicitly disabled)
- Available modes and groups
- Last updated timestamp

## State Persistence

Tool group state is persisted in `.exarp/workflow_mode.json`:

```json
{
  "current_mode": "development",
  "extra_groups": ["memory", "workflow"],
  "disabled_groups": ["testing"],
  "last_updated": "2026-01-11T22:20:44+02:00"
}
```

**State Management:**
- State is loaded on server startup
- State is saved after each enable/disable operation
- State file is created automatically if it doesn't exist
- Directory `.exarp/` is created automatically if needed

## Workflow Mode Interaction

When switching workflow modes using the `mode` parameter:

```json
{
  "action": "focus",
  "mode": "security_review"
}
```

**Behavior:**
- Switches to the specified mode
- **Clears** `extra_groups` and `disabled_groups`
- Applies mode-specific default groups
- State is persisted

**Note:** Enable/disable operations are independent of mode switches. You can enable/disable groups without changing the mode, and the customizations persist until you switch modes.

## Restrictions

### Protected Groups

The following groups **cannot be disabled**:

- `core` - Required for basic system operation
- `tool_catalog` - Required for tool discovery (includes `workflow_mode` itself)

**Error Response:**
```json
{
  "success": false,
  "error": "cannot disable core group - always required"
}
```

### Case Sensitivity

Group names are **case-insensitive**. All of these are equivalent:

- `MEMORY`
- `memory`
- `Memory`

The system normalizes group names to lowercase internally.

## Examples

### Example 1: Enable Memory Tools

```json
{
  "action": "focus",
  "enable_group": "memory"
}
```

**Response:**
```json
{
  "success": true,
  "action": "group_enabled",
  "group": "memory",
  "mode": "development",
  "extra_groups": ["memory"],
  "disabled_groups": [],
  "available_groups": ["core", "tool_catalog", "health", "tasks", ...],
  "last_updated": "2026-01-11T22:20:44+02:00"
}
```

### Example 2: Disable Testing Tools

```json
{
  "action": "focus",
  "disable_group": "testing"
}
```

**Response:**
```json
{
  "success": true,
  "action": "group_disabled",
  "group": "testing",
  "mode": "development",
  "extra_groups": [],
  "disabled_groups": ["testing"],
  "available_groups": ["core", "tool_catalog", "health", "tasks", ...],
  "last_updated": "2026-01-11T22:20:44+02:00"
}
```

### Example 3: Attempt to Disable Core (Error)

```json
{
  "action": "focus",
  "disable_group": "core"
}
```

**Response:**
```json
{
  "success": false,
  "error": "cannot disable core group - always required"
}
```

### Example 4: Check Current Status

```json
{
  "action": "focus",
  "status": true
}
```

**Response:**
```json
{
  "mode": "development",
  "extra_groups": ["memory"],
  "disabled_groups": ["testing"],
  "last_updated": "2026-01-11T22:20:44+02:00",
  "available_modes": ["daily_checkin", "security_review", "task_management", ...],
  "available_groups": ["core", "tool_catalog", "health", "tasks", ...]
}
```

## Implementation Details

### Code Location

- **Implementation:** `internal/tools/workflow_mode.go`
- **State Management:** `WorkflowModeManager` struct
- **Functions:**
  - `enableGroup(group string) error` - Enables a tool group
  - `disableGroup(group string) error` - Disables a tool group
  - `setMode(mode string) (string, error)` - Sets workflow mode (clears groups)

### State File Location

- **Path:** `.exarp/workflow_mode.json` (relative to project root)
- **Format:** JSON
- **Permissions:** Created with `0644` permissions
- **Auto-creation:** Directory and file created automatically

### Integration with Workflow Modes

Tool groups work in conjunction with workflow modes:

1. **Mode Defaults:** Each mode has default groups (see workflow mode documentation)
2. **Extra Groups:** Groups added via `enable_group` are in addition to mode defaults
3. **Disabled Groups:** Groups in `disabled_groups` override mode defaults
4. **Mode Switch:** Switching modes clears `extra_groups` and `disabled_groups`

## Use Cases

### Use Case 1: Focused Development Session

Disable groups not needed for current work:

```json
{
  "action": "focus",
  "mode": "development",
  "disable_group": "prd",
  "disable_group": "advisors"
}
```

### Use Case 2: Add Memory Tools to Any Mode

Enable memory tools regardless of current mode:

```json
{
  "action": "focus",
  "enable_group": "memory"
}
```

### Use Case 3: Security Review Focus

Switch to security mode and disable non-security groups:

```json
{
  "action": "focus",
  "mode": "security_review",
  "disable_group": "testing",
  "disable_group": "automation"
}
```

## Best Practices

1. **Use Modes First:** Try workflow modes before manually enabling/disabling groups
2. **Preserve State:** Tool group state persists across server restarts
3. **Check Status:** Use `status: true` to see current state before making changes
4. **Mode Switches Reset:** Remember that switching modes clears custom groups
5. **Protected Groups:** Never try to disable `core` or `tool_catalog`

## Related Documentation

- `WORKFLOW_USAGE.md` - General workflow mode usage guide
- `docs/README.md` - Documentation index

## See Also

- Workflow mode documentation for mode-specific default groups
- Tool catalog for complete list of tools in each group
