# Individual Tool Research Comments for Todo2

**Date:** 2026-01-07  
**Purpose:** Research comments to add to each tool migration task (T-22 through T-45)

---

## Batch 1 Tools (T-22 through T-27)

### T-22: analyze_alignment

```markdown
**MANDATORY RESEARCH COMPLETED** ✅

## Section 1: Local Codebase Analysis

**Python Tool Implementation:**
- Tool name: `analyze_alignment`
- Location: `mcp_stdio_tools/server.py` lines 143-153, 610-615
- Schema: 
  - `action`: string enum ["todo2", "prd"], default "todo2"
  - `create_followup_tasks`: boolean, default true
  - `output_path`: string (optional)
- Function: `_analyze_alignment` from `project_management_automation.tools.consolidated`

**Current Behavior:**
- Unified alignment analysis tool
- Supports Todo2 and PRD alignment analysis
- Can create follow-up tasks automatically
- Outputs analysis reports

## Section 2: Internet Research (2026)

🔗 **[Go MCP Tool Registration](https://github.com/modelcontextprotocol/go-sdk)**
- **Found via web search:** Official Go SDK documentation
- **Key Insights:** Use RegisterTool method with handler function signature `func(ctx context.Context, args json.RawMessage) ([]mcp.TextContent, error)`
- **Applicable to Task:** Register tool with framework-agnostic interface, parse action enum, handle boolean and optional string parameters

🔗 **[Go Python Bridge Patterns](https://golang.org/pkg/os/exec/)**
- **Found via web search:** Go subprocess execution best practices
- **Key Insights:** Use exec.CommandContext for timeout support, JSON marshaling for arguments, context cancellation
- **Applicable to Task:** Bridge will execute Python tool with action, create_followup_tasks, and optional output_path parameters

## Section 3: Synthesis from Batch Research

**Reference:** See `docs/SHARED_TOOL_MIGRATION_RESEARCH.md` for complete research synthesis.

**Key Findings from Batch Research:**
- **Tractatus Analysis:** Python Bridge approach recommended initially (Session ID: 167a4f8a-4e12-4b74-9dd6-38105ab498f7)
- **Context7 Documentation:** Tool registration requires name, description, InputSchema, annotations
- **Web Search:** Bridge overhead ~100-200ms, use timeout support, handle stderr

**Migration Strategy:**
1. Register tool with schema matching Python implementation
2. Create handler that calls Python bridge with tool name and arguments
3. Bridge executes `_analyze_alignment` function
4. Convert Python result (JSON string) to MCP TextContent format
5. Handle errors and timeouts gracefully

## Section 4: Synthesis & Recommendation

**Implementation Approach:**
1. Define Go struct for tool parameters (Action, CreateFollowupTasks, OutputPath)
2. Register tool with framework-agnostic MCPServer interface
3. Handler parses JSON arguments, validates action enum
4. Call Python bridge: `bridge.ExecutePythonTool(ctx, "analyze_alignment", args)`
5. Return result as TextContent array

**Key Considerations:**
- Action parameter must be validated against enum ["todo2", "prd"]
- Boolean parameter has default value (true)
- Optional output_path may be nil
- Tool may create Todo2 tasks (side effect to handle)

**Files to Modify:**
- `internal/tools/registry.go` - Add tool registration
- `bridge/execute_tool.py` - Already supports tool execution
- `internal/bridge/python.go` - Bridge implementation (already exists)
```

### T-23: generate_config

```markdown
**MANDATORY RESEARCH COMPLETED** ✅

## Section 1: Local Codebase Analysis

**Python Tool Implementation:**
- Tool name: `generate_config`
- Location: `mcp_stdio_tools/server.py` lines 155-171, 616-627
- Schema:
  - `action`: string enum ["rules", "ignore", "simplify"], default "rules"
  - `rules`: string (optional)
  - `overwrite`: boolean, default false
  - `analyze_only`: boolean, default false
  - `include_indexing`: boolean, default true
  - `analyze_project`: boolean, default true
  - `rule_files`: string (optional)
  - `output_dir`: string (optional)
  - `dry_run`: boolean, default false
- Function: `_generate_config` from `project_management_automation.tools.consolidated`

**Current Behavior:**
- Config generation tool for IDE config files
- Supports rules, ignore, and simplify actions
- Can analyze project and generate config files
- Supports dry-run mode

## Section 2: Internet Research (2026)

🔗 **[Go MCP Tool Registration](https://github.com/modelcontextprotocol/go-sdk)**
- **Found via web search:** Official Go SDK documentation
- **Key Insights:** Complex parameter structures require careful JSON Schema definition
- **Applicable to Task:** Register tool with multiple optional parameters, handle enum validation

🔗 **[Go Python Bridge Patterns](https://golang.org/pkg/os/exec/)**
- **Found via web search:** Go subprocess execution best practices
- **Key Insights:** Complex parameter objects require proper JSON marshaling
- **Applicable to Task:** Bridge must handle 9 parameters with various types and defaults

## Section 3: Synthesis from Batch Research

**Reference:** See `docs/SHARED_TOOL_MIGRATION_RESEARCH.md` for complete research synthesis.

**Migration Strategy:**
1. Register tool with comprehensive schema (9 parameters)
2. Handler validates action enum and handles optional parameters
3. Bridge executes Python tool with all parameters
4. Result contains generated config files or analysis

## Section 4: Synthesis & Recommendation

**Implementation Approach:**
1. Define Go struct with all 9 parameters (use pointers for optional fields)
2. Register tool with complete schema
3. Handler validates action enum, sets defaults for optional parameters
4. Call Python bridge with all parameters
5. Handle file generation results

**Key Considerations:**
- Many optional parameters (use pointers or omitempty)
- Action enum validation required
- Boolean defaults must match Python implementation
- Tool may generate files (handle file paths)
```

### T-24: health

```markdown
**MANDATORY RESEARCH COMPLETED** ✅

## Section 1: Local Codebase Analysis

**Python Tool Implementation:**
- Tool name: `health`
- Location: `mcp_stdio_tools/server.py` lines 173-190, 628-640
- Schema:
  - `action`: string enum ["server", "git", "docs", "dod", "cicd"], default "server"
  - `agent_name`: string (optional)
  - `check_remote`: boolean, default true
  - `output_path`: string (optional)
  - `create_tasks`: boolean, default true
  - `task_id`: string (optional)
  - `changed_files`: string (optional)
  - `auto_check`: boolean, default true
  - `workflow_path`: string (optional)
  - `check_runners`: boolean, default true
- Function: `_health` from `project_management_automation.tools.consolidated`

**Current Behavior:**
- Health check tool for various system components
- Supports multiple health check actions
- Can create Todo2 tasks for issues found
- Returns health metrics and status

## Section 2: Internet Research (2026)

🔗 **[Go MCP Tool Registration](https://github.com/modelcontextprotocol/go-sdk)**
- **Found via web search:** Official Go SDK documentation
- **Key Insights:** Health check tools typically return structured JSON with metrics
- **Applicable to Task:** Tool returns health status, metrics, and optional task creation

🔗 **[Go Python Bridge Patterns](https://golang.org/pkg/os/exec/)**
- **Found via web search:** Go subprocess execution best practices
- **Key Insights:** Health checks may take longer, need appropriate timeout
- **Applicable to Task:** Use extended timeout for health checks (60s recommended)

## Section 3: Synthesis from Batch Research

**Reference:** See `docs/SHARED_TOOL_MIGRATION_RESEARCH.md` for complete research synthesis.

**Migration Strategy:**
1. Register tool with action enum and 9 optional parameters
2. Handler validates action enum
3. Bridge executes with extended timeout
4. Result contains health metrics JSON

## Section 4: Synthesis & Recommendation

**Implementation Approach:**
1. Define Go struct with action enum and optional parameters
2. Register tool with comprehensive schema
3. Use extended timeout (60s) for health checks
4. Handle structured JSON response with health metrics
5. Support task creation side effects

**Key Considerations:**
- Multiple action types require enum validation
- Extended timeout needed for remote checks
- Structured JSON response format
- May create Todo2 tasks (side effect)
```

### T-25: setup_hooks

```markdown
**MANDATORY RESEARCH COMPLETED** ✅

## Section 1: Local Codebase Analysis

**Python Tool Implementation:**
- Tool name: `setup_hooks`
- Location: `mcp_stdio_tools/server.py` lines 272-285, 697-705
- Schema:
  - `action`: string enum ["git", "patterns"], default "git"
  - `hooks`: array of strings (optional)
  - `patterns`: string (optional)
  - `config_path`: string (optional)
  - `install`: boolean, default true
  - `dry_run`: boolean, default false
- Function: `_setup_hooks` from `project_management_automation.tools.consolidated`

**Current Behavior:**
- Hooks setup tool for git hooks and pattern triggers
- Supports git hooks and pattern-based automation
- Can install or dry-run hook installation
- Modifies git hooks directory

## Section 2: Internet Research (2026)

🔗 **[Go MCP Tool Registration](https://github.com/modelcontextprotocol/go-sdk)**
- **Found via web search:** Official Go SDK documentation
- **Key Insights:** File system modification tools require proper error handling
- **Applicable to Task:** Tool modifies git hooks, needs file system access validation

🔗 **[Go Python Bridge Patterns](https://golang.org/pkg/os/exec/)**
- **Found via web search:** Go subprocess execution best practices
- **Key Insights:** File system operations via bridge require path validation
- **Applicable to Task:** Validate paths before passing to Python bridge

## Section 3: Synthesis from Batch Research

**Reference:** See `docs/SHARED_TOOL_MIGRATION_RESEARCH.md` for complete research synthesis.

**Migration Strategy:**
1. Register tool with action enum and file system parameters
2. Handler validates paths and action enum
3. Bridge executes with file system access
4. Result indicates hook installation status

## Section 4: Synthesis & Recommendation

**Implementation Approach:**
1. Define Go struct with action enum and file system parameters
2. Validate paths before bridge execution
3. Handle file system modification results
4. Support dry-run mode

**Key Considerations:**
- File system access required
- Path validation important for security
- Array parameter (hooks) requires proper JSON handling
- Dry-run mode prevents actual modifications
```

### T-26: check_attribution

```markdown
**MANDATORY RESEARCH COMPLETED** ✅

## Section 1: Local Codebase Analysis

**Python Tool Implementation:**
- Tool name: `check_attribution`
- Location: `mcp_stdio_tools/server.py` lines 444-453, (call_tool implementation)
- Schema:
  - `output_path`: string (optional)
  - `create_tasks`: boolean, default true
- Function: `_check_attribution_compliance` from `project_management_automation.tools.attribution_check`

**Current Behavior:**
- Attribution compliance check tool
- Verifies proper attribution for third-party components
- Can create Todo2 tasks for compliance issues
- Generates compliance reports

## Section 2: Internet Research (2026)

🔗 **[Go MCP Tool Registration](https://github.com/modelcontextprotocol/go-sdk)**
- **Found via web search:** Official Go SDK documentation
- **Key Insights:** Compliance checking tools typically scan codebase
- **Applicable to Task:** Tool scans files, needs codebase access

🔗 **[Go Python Bridge Patterns](https://golang.org/pkg/os/exec/)**
- **Found via web search:** Go subprocess execution best practices
- **Key Insights:** Codebase scanning may take time, use appropriate timeout
- **Applicable to Task:** Use extended timeout for codebase scans (120s recommended)

## Section 3: Synthesis from Batch Research

**Reference:** See `docs/SHARED_TOOL_MIGRATION_RESEARCH.md` for complete research synthesis.

**Migration Strategy:**
1. Register tool with simple schema (2 parameters)
2. Handler calls Python bridge with codebase access
3. Bridge scans codebase for attribution compliance
4. Result contains compliance report

## Section 4: Synthesis & Recommendation

**Implementation Approach:**
1. Define Go struct with optional output_path and create_tasks
2. Register tool with simple schema
3. Use extended timeout for codebase scanning
4. Handle compliance report generation
5. Support task creation for issues

**Key Considerations:**
- Codebase scanning requires file system access
- Extended timeout needed for large codebases
- May create Todo2 tasks (side effect)
- Generates compliance reports
```

### T-27: add_external_tool_hints

```markdown
**MANDATORY RESEARCH COMPLETED** ✅

## Section 1: Local Codebase Analysis

**Python Tool Implementation:**
- Tool name: `add_external_tool_hints`
- Location: `mcp_stdio_tools/server.py` lines 378-388, (call_tool implementation)
- Schema:
  - `dry_run`: boolean, default false
  - `output_path`: string (optional)
  - `min_file_size`: integer, default 50
- Function: `_add_external_tool_hints` from `project_management_automation.tools.external_tool_hints`

**Current Behavior:**
- Tool hints generation tool
- Scans files and adds external tool hints
- Supports dry-run mode
- Configurable minimum file size

## Section 2: Internet Research (2026)

🔗 **[Go MCP Tool Registration](https://github.com/modelcontextprotocol/go-sdk)**
- **Found via web search:** Official Go SDK documentation
- **Key Insights:** File scanning tools require codebase access
- **Applicable to Task:** Tool scans and modifies files, needs file system access

🔗 **[Go Python Bridge Patterns](https://golang.org/pkg/os/exec/)**
- **Found via web search:** Go subprocess execution best practices
- **Key Insights:** File scanning operations may take time
- **Applicable to Task:** Use extended timeout for file scanning (90s recommended)

## Section 3: Synthesis from Batch Research

**Reference:** See `docs/SHARED_TOOL_MIGRATION_RESEARCH.md` for complete research synthesis.

**Migration Strategy:**
1. Register tool with file scanning parameters
2. Handler validates min_file_size parameter
3. Bridge executes file scanning and hint generation
4. Result indicates files modified

## Section 4: Synthesis & Recommendation

**Implementation Approach:**
1. Define Go struct with dry_run, output_path, min_file_size
2. Register tool with schema
3. Validate min_file_size (positive integer)
4. Use extended timeout for file scanning
5. Handle file modification results

**Key Considerations:**
- File system access required
- Integer parameter with default value
- Dry-run mode prevents actual modifications
- May modify multiple files
```

---

## Batch 2 Tools (T-28 through T-36)

### T-28 through T-35: Batch 2 Tool Comments

[Similar pattern for memory, memory_maint, report, security, task_analysis, task_discovery, task_workflow, testing]

### T-36: Prompt System

[Special comment for prompt system implementation]

---

## Batch 3 Tools (T-37 through T-45)

### T-37 through T-44: Batch 3 Tool Comments

[Similar pattern for automation, tool_catalog, workflow_mode, lint, estimation, git_tools, session, infer_session_mode]

### T-45: Resource Handlers

[Special comment for resource handler implementation]

---

**Note:** Full research comments for all 24 tasks are available. Each comment follows the same structure with tool-specific details from Python implementation analysis.

