# Automation Configuration Analysis

**Date:** 2026-01-13  
**Status:** Analysis & Proposal

---

## Executive Summary

This document analyzes all hard-coded automations in the exarp-go codebase and proposes a per-project configuration-based approach to simplify and customize automation workflows.

**Key Finding:** Multiple automations are hard-coded across Go code, Makefile, and prompt templates, making them difficult to customize per project.

---

## Hard-Coded Automations Identified

### 1. **Automation Tool (`internal/tools/automation_native.go`)**

#### Daily Automation (Lines 35-149)
**Hard-coded tasks:**
```go
// Task 1: analyze_alignment (todo2 action)
runDailyTask(ctx, "analyze_alignment", map[string]interface{}{"action": "todo2"})

// Task 2: task_analysis (duplicates action)
runDailyTask(ctx, "task_analysis", map[string]interface{}{"action": "duplicates"})

// Task 3: health (docs action)
runDailyTask(ctx, "health", map[string]interface{}{"action": "docs"})
```

**Issues:**
- Fixed 3-task sequence
- No way to skip tasks
- No project-specific customization
- Tool parameters are hard-coded

#### Nightly Automation (Lines 151-266)
**Hard-coded tasks:**
```go
// Task 1: memory_maint (consolidate action)
runDailyTaskPython(ctx, "memory_maint", map[string]interface{}{"action": "consolidate"})

// Task 2: task_analysis (tags action)
runDailyTask(ctx, "task_analysis", map[string]interface{}{"action": "tags"})

// Task 3: health (server action)
runDailyTask(ctx, "health", map[string]interface{}{"action": "server"})
```

**Issues:**
- Same problems as daily automation
- Mix of native Go and Python bridge calls (hard-coded)

#### Sprint Automation (Lines 268-383)
**Hard-coded tasks:**
```go
// Task 1: analyze_alignment (todo2 action)
runDailyTask(ctx, "analyze_alignment", map[string]interface{}{"action": "todo2"})

// Task 2: task_analysis (hierarchy action)
runDailyTask(ctx, "task_analysis", map[string]interface{}{"action": "hierarchy"})

// Task 3: report (overview action)
runDailyTaskPython(ctx, "report", map[string]interface{}{"action": "overview"})
```

**Issues:**
- Fixed sprint workflow
- No project-specific variations
- Cannot adapt to different sprint methodologies

### 2. **Makefile Targets (`Makefile`)**

#### Sprint Start (Lines 584-597)
```makefile
sprint-start:
	$(BINARY_PATH) -tool automation -args '{"action":"sprint","run_analysis_tools":true,...}'
```

**Hard-coded parameters:**
- `run_analysis_tools=true`
- `run_testing_tools=false`
- `extract_subtasks=true`
- `auto_approve=true`
- `max_iterations=1`

#### Sprint End (Lines 599-614)
```makefile
sprint-end:
	$(MAKE) test-coverage
	$(BINARY_PATH) -tool health -args '{"action":"docs"}'
	$(BINARY_PATH) -tool security -args '{"action":"report"}'
```

**Hard-coded sequence:**
1. Test coverage (always runs)
2. Docs health check (always runs)
3. Security report (always runs)

#### Pre-Sprint (Lines 616-631)
```makefile
pre-sprint:
	$(BINARY_PATH) -tool task_analysis -args '{"action":"duplicates"}'
	$(BINARY_PATH) -tool analyze_alignment -args '{"action":"todo2"}'
	$(BINARY_PATH) -tool health -args '{"action":"docs"}'
```

**Hard-coded 3-step sequence**

### 3. **Prompt Templates (`internal/prompts/templates.go`)**

#### Daily Check-in (Lines 173-190)
**Hard-coded workflow:**
```
1. health(action="server")
2. report(action="scorecard")
3. consult_advisor(stage="daily_checkin")
4. task_workflow(action="clarify", sub_action="list")
5. health(action="git")
```

#### Sprint Start (Lines 192-201)
**Hard-coded workflow:**
```
1. task_analysis(action="duplicates")
2. analyze_alignment(action="todo2")
3. task_workflow(action="approve")
4. task_workflow(action="clarify", sub_action="list")
5. consult_advisor(stage="planning")
```

#### Sprint End (Lines 203-212)
**Hard-coded workflow:**
```
1. testing(action="run", coverage=True)
2. testing(action="coverage")
3. health(action="docs")
4. security(action="report")
5. consult_advisor(stage="review")
```

#### Pre-Sprint (Lines 214-221)
**Hard-coded workflow:**
```
1. task_analysis(action="duplicates")
2. analyze_alignment(action="todo2")
3. health(action="docs")
```

#### Post-Implementation (Lines 223-230)
**Hard-coded workflow:**
```
1. health(action="docs")
2. security(action="report")
3. run_automation(action="discover")
```

### 4. **Python Automation Scripts**

Located in `project_management_automation/scripts/`:
- `automate_daily.py` - Hard-coded daily tasks
- `automate_sprint.py` - Hard-coded sprint tasks
- `automate_stale_task_cleanup.py` - Hard-coded cleanup logic
- `automate_todo2_alignment_v2.py` - Hard-coded alignment checks

---

## Problems with Current Approach

### 1. **No Project Customization**
- Every project gets the same automation workflows
- Cannot adapt to different project types (library vs. app vs. service)
- Cannot skip irrelevant tasks (e.g., docs check for internal tools)

### 2. **Maintenance Burden**
- Changes require code modifications
- Multiple places to update (Go code, Makefile, prompts)
- Risk of inconsistencies across locations

### 3. **Testing Complexity**
- Hard to test different automation configurations
- Cannot easily experiment with workflow variations
- Difficult to A/B test automation strategies

### 4. **User Experience**
- Users cannot customize workflows without code changes
- No way to disable unwanted automations
- Cannot add project-specific tasks

### 5. **Documentation Drift**
- Prompt templates describe workflows that may not match implementation
- Makefile targets may diverge from Go code
- Multiple sources of truth

---

## Proposed Solution: Per-Project Configuration

### Architecture Overview

```
.exarp/
  └── automations.yaml    # Per-project automation configuration
```

### Configuration Schema

```yaml
# .exarp/automations.yaml
version: "1.0"

# Default automation workflows
automations:
  daily:
    enabled: true
    tasks:
      - tool: analyze_alignment
        params:
          action: todo2
        enabled: true
        timeout: 300s
      
      - tool: task_analysis
        params:
          action: duplicates
        enabled: true
        timeout: 180s
      
      - tool: health
        params:
          action: docs
        enabled: true
        timeout: 120s
        condition: "project_type != 'internal'"
    
    parallel: false  # Run tasks sequentially
    continue_on_error: true
    notify_on_failure: false

  nightly:
    enabled: true
    schedule: "0 2 * * *"  # 2 AM daily
    tasks:
      - tool: memory_maint
        params:
          action: consolidate
        enabled: true
        bridge: python  # Specify bridge type
      
      - tool: task_analysis
        params:
          action: tags
        enabled: true
      
      - tool: health
        params:
          action: server
        enabled: true
    
    parallel: false
    continue_on_error: true

  sprint_start:
    enabled: true
    tasks:
      - tool: task_analysis
        params:
          action: duplicates
        enabled: true
      
      - tool: analyze_alignment
        params:
          action: todo2
        enabled: true
      
      - tool: task_workflow
        params:
          action: approve
        enabled: true
      
      - tool: task_workflow
        params:
          action: clarify
          sub_action: list
        enabled: true
      
      - tool: consult_advisor
        params:
          stage: planning
        enabled: true
        mcp_server: devwisdom  # Specify MCP server
    
    parallel: false
    continue_on_error: false  # Fail fast for sprint start

  sprint_end:
    enabled: true
    tasks:
      - tool: testing
        params:
          action: run
          coverage: true
        enabled: true
      
      - tool: testing
        params:
          action: coverage
        enabled: true
      
      - tool: health
        params:
          action: docs
        enabled: true
      
      - tool: security
        params:
          action: report
        enabled: true
      
      - tool: consult_advisor
        params:
          stage: review
        enabled: true
        mcp_server: devwisdom
    
    parallel: true  # Can run in parallel
    continue_on_error: false

  pre_sprint:
    enabled: true
    tasks:
      - tool: task_analysis
        params:
          action: duplicates
        enabled: true
      
      - tool: analyze_alignment
        params:
          action: todo2
        enabled: true
      
      - tool: health
        params:
          action: docs
        enabled: true
    
    parallel: false
    continue_on_error: true

  post_impl:
    enabled: true
    tasks:
      - tool: health
        params:
          action: docs
        enabled: true
      
      - tool: security
        params:
          action: report
        enabled: true
      
      - tool: automation
        params:
          action: discover
        enabled: true
    
    parallel: false
    continue_on_error: true

# Project-specific overrides
project_overrides:
  # Example: Skip docs check for internal tools
  - automation: daily
    task: health
    condition: "project_type == 'internal'"
    action: disable
  
  # Example: Add custom task for specific project
  - automation: daily
    task:
      tool: custom_check
      params:
        action: validate
    action: prepend  # Add at beginning

# Global settings
settings:
  default_timeout: 300s
  max_parallel_tasks: 5
  retry_failed_tasks: false
  log_level: info
  output_format: json
```

### Implementation Plan

#### Phase 1: Configuration Loading
1. **Create config package** (`internal/config/automations.go`)
   - Load `.exarp/automations.yaml`
   - Validate schema
   - Merge with defaults
   - Support environment variable overrides

2. **Default configuration** (embedded in binary)
   - Current hard-coded workflows as defaults
   - Backward compatible
   - Can be overridden per project

#### Phase 2: Refactor Automation Tool
1. **Update `handleAutomationNative`**
   - Load configuration from `.exarp/automations.yaml`
   - Iterate over configured tasks instead of hard-coded list
   - Support task conditions (enable/disable based on project type)
   - Support parallel execution where configured

2. **Task execution engine**
   - Generic task runner
   - Support native Go, Python bridge, and MCP server calls
   - Timeout handling
   - Error handling with `continue_on_error` option

#### Phase 3: Update Makefile
1. **Makefile targets read from config**
   ```makefile
   sprint-start:
   	@$(BINARY_PATH) -tool automation -args '{"action":"sprint_start","use_config":true}'
   ```

2. **Or use config directly**
   ```makefile
   sprint-start:
   	@$(BINARY_PATH) -tool automation -args '{"action":"sprint_start"}'
   	# Tool automatically loads .exarp/automations.yaml
   ```

#### Phase 4: Update Prompt Templates
1. **Dynamic prompt generation**
   - Load automation config
   - Generate prompt text from configured tasks
   - Keep templates as fallback

2. **Or reference config in prompts**
   ```
   Use automation(action="sprint_start") to run configured sprint start workflow.
   See .exarp/automations.yaml for customization.
   ```

#### Phase 5: CLI Tool for Config Management
```bash
# Generate default config
exarp-go automation config init

# Validate config
exarp-go automation config validate

# Test automation (dry run)
exarp-go automation test daily

# List configured automations
exarp-go automation list
```

---

## Benefits

### 1. **Flexibility**
- ✅ Per-project customization
- ✅ Enable/disable tasks
- ✅ Custom task parameters
- ✅ Project-specific conditions

### 2. **Maintainability**
- ✅ Single source of truth (YAML config)
- ✅ Version controlled with project
- ✅ Easy to review and modify
- ✅ No code changes needed

### 3. **Testing**
- ✅ Test different configurations
- ✅ Easy to experiment
- ✅ Share configs across projects
- ✅ A/B test workflows

### 4. **User Experience**
- ✅ Non-developers can customize
- ✅ Clear documentation in YAML
- ✅ Validation and error messages
- ✅ Preview before running

### 5. **Backward Compatibility**
- ✅ Default config matches current behavior
- ✅ Existing code continues to work
- ✅ Gradual migration path
- ✅ Can still use hard-coded defaults if no config

---

## Migration Path

### Step 1: Add Config Support (Non-Breaking)
- Add config loading with fallback to defaults
- Keep existing hard-coded code as defaults
- Test with and without config files

### Step 2: Generate Default Configs
- CLI tool to generate `.exarp/automations.yaml` from current defaults
- Users can customize from there

### Step 3: Update Documentation
- Document configuration format
- Provide examples for different project types
- Migration guide for existing projects

### Step 4: Deprecate Hard-Coded Values (Optional)
- After config is stable, mark hard-coded values as deprecated
- Encourage config usage
- Eventually remove hard-coded defaults (long-term)

---

## Example Use Cases

### 1. Internal Tool Project
```yaml
automations:
  daily:
    tasks:
      - tool: analyze_alignment
        enabled: true
      - tool: task_analysis
        enabled: true
      - tool: health
        enabled: false  # Skip docs check for internal tools
```

### 2. Library Project
```yaml
automations:
  daily:
    tasks:
      - tool: analyze_alignment
        enabled: true
      - tool: health
        params:
          action: docs
        enabled: true  # Docs are important for libraries
      - tool: testing
        params:
          action: coverage
        enabled: true  # Coverage is critical for libraries
```

### 3. Service Project
```yaml
automations:
  daily:
    tasks:
      - tool: health
        params:
          action: server
        enabled: true  # Server health is critical
      - tool: security
        params:
          action: scan
        enabled: true  # Security is critical for services
```

### 4. Custom Workflow
```yaml
automations:
  weekly_review:
    enabled: true
    tasks:
      - tool: report
        params:
          action: scorecard
        enabled: true
      - tool: report
        params:
          action: overview
        enabled: true
      - tool: consult_advisor
        params:
          stage: review
        enabled: true
```

---

## Implementation Considerations

### 1. **Configuration Location**
- **Option A:** `.exarp/automations.yaml` (project root)
- **Option B:** `exarp.automations.yaml` (project root)
- **Option C:** `config/automations.yaml` (configurable path)

**Recommendation:** Option A (`.exarp/automations.yaml`) - consistent with `.todo2/` pattern

### 2. **Configuration Format**
- **YAML:** Human-readable, easy to edit
- **JSON:** Machine-friendly, but less readable
- **TOML:** Alternative, but less common

**Recommendation:** YAML (most user-friendly)

### 3. **Validation**
- Schema validation (JSON Schema or custom)
- Runtime validation (check tool exists, params valid)
- Helpful error messages

### 4. **Performance**
- Cache loaded config
- Lazy loading (only when automation runs)
- Fast path for default config (no file I/O)

### 5. **Security**
- Validate tool names (prevent arbitrary execution)
- Sandbox parameters (no code injection)
- Path validation for file operations

---

## Next Steps

1. **Review & Approval**
   - Review this proposal
   - Get feedback on configuration schema
   - Decide on implementation priority

2. **Prototype**
   - Implement config loading
   - Refactor one automation (e.g., daily) to use config
   - Test with sample projects

3. **Full Implementation**
   - Complete all automations
   - Update Makefile
   - Update prompt templates
   - Add CLI tool

4. **Documentation**
   - Configuration guide
   - Examples for different project types
   - Migration guide

5. **Testing**
   - Unit tests for config loading
   - Integration tests for automations
   - Test with various configurations

---

## Questions for Discussion

1. **Configuration location:** `.exarp/automations.yaml` or alternative?
2. **Backward compatibility:** Keep hard-coded defaults forever or deprecate?
3. **Priority:** Which automation should be migrated first?
4. **CLI tool:** Essential for v1 or can wait?
5. **Validation:** How strict should parameter validation be?

---

## Related Files

- `internal/tools/automation_native.go` - Current implementation
- `Makefile` - Hard-coded targets
- `internal/prompts/templates.go` - Hard-coded workflows
- `project_management_automation/scripts/` - Python automation scripts

---

## References

- Current automation implementation: `internal/tools/automation_native.go`
- Makefile automation targets: `Makefile` (lines 584-675)
- Prompt templates: `internal/prompts/templates.go`
