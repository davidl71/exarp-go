# Additional Configurable Parameters Recommendations

**Date:** 2026-01-13  
**Status:** Recommendations for Configuration System

---

## Executive Summary

This document extends the automation configuration analysis with recommendations for additional configurable parameters found throughout the exarp-go codebase. These parameters are currently hard-coded but would benefit from per-project configuration.

---

## Categories of Configurable Parameters

### 1. **Timeouts & Durations**

#### Current Hard-Coded Values
- **Task lock lease duration**: `30 * time.Minute` (multiple locations)
- **Stale lock detection**: `5 * time.Minute`
- **Tool execution timeouts**: `60 * time.Second` (scorecard, linting)
- **Ollama HTTP client timeout**: `600 * time.Second` (10 minutes for model downloads)
- **Database retry timeout**: `60 * time.Second`
- **Context timeout**: Various (60s, 300s, 600s)

#### Recommended Configuration
```yaml
# .exarp/config.yaml
timeouts:
  # Task management
  task_lock_lease: 30m          # Task lock duration
  task_lock_renewal: 20m        # When to renew lease (before expiry)
  stale_lock_threshold: 5m      # Consider lock stale after this
  
  # Tool execution
  tool_default: 60s              # Default tool timeout
  tool_scorecard: 60s           # Scorecard generation
  tool_linting: 60s             # Linting operations
  tool_testing: 300s            # Test execution
  tool_report: 60s              # Report generation
  
  # External services
  ollama_download: 600s          # Model downloads
  ollama_generate: 300s          # Text generation
  http_client: 30s              # General HTTP requests
  database_retry: 60s            # Database operation retry
  
  # Context management
  context_summarize: 30s         # Context summarization
  context_budget: 10s            # Budget calculation
```

### 2. **Thresholds & Limits**

#### Current Hard-Coded Values
- **Similarity threshold**: `0.85` (duplicate detection)
- **Confidence threshold**: `0.7` (testing, estimation)
- **Min coverage**: `80` (testing)
- **Min description length**: `50` characters (task clarity)
- **Max parallel tasks**: `10` (automation)
- **Max tasks per host**: `5` (automation)
- **Max iterations**: `10` (automation)
- **Rate limit**: `100 requests/minute` (security)
- **Token estimation**: `0.25 tokens/char` (context)

#### Recommended Configuration
```yaml
# .exarp/config.yaml
thresholds:
  # Task analysis
  similarity_threshold: 0.85           # Duplicate detection
  min_description_length: 50            # Task clarity check
  min_task_confidence: 0.7             # Task confidence threshold
  
  # Testing
  min_coverage: 80                     # Minimum test coverage %
  min_test_confidence: 0.7            # Test suggestion confidence
  
  # Estimation
  min_estimation_confidence: 0.7      # Task estimation confidence
  mlx_weight: 0.3                      # MLX model weight in estimation
  
  # Automation
  max_parallel_tasks: 10               # Max concurrent tasks
  max_tasks_per_host: 5                # Max tasks per agent/host
  max_automation_iterations: 10        # Max automation loop iterations
  
  # Context management
  tokens_per_char: 0.25                # Token estimation ratio
  default_context_budget: 4000         # Default token budget
  context_reduction_threshold: 0.5     # When to suggest summarization
  
  # Security
  rate_limit_requests: 100             # Requests per window
  rate_limit_window: 1m                # Rate limit window
  max_file_size: 10485760              # 10MB max file size
  max_path_depth: 20                   # Max directory depth
```

### 3. **Database Configuration**

#### Current Hard-Coded Values
- **Database path**: `.todo2/todo2.db` (relative to project root)
- **JSON fallback path**: `.todo2/state.todo2.json`
- **File permissions**: `0755` (directory creation)
- **Retry attempts**: Various (3-5 attempts)
- **Retry backoff**: Exponential (various durations)

#### Recommended Configuration
```yaml
# .exarp/config.yaml
database:
  # Paths
  sqlite_path: ".todo2/todo2.db"       # SQLite database location
  json_fallback_path: ".todo2/state.todo2.json"  # Fallback JSON
  backup_path: ".todo2/backups"       # Backup directory
  
  # Connection
  max_connections: 10                  # Max DB connections
  connection_timeout: 30s              # Connection timeout
  query_timeout: 60s                   # Query timeout
  
  # Retry logic
  retry_attempts: 3                   # Max retry attempts
  retry_initial_delay: 100ms          # Initial retry delay
  retry_max_delay: 5s                 # Max retry delay
  retry_multiplier: 2.0               # Exponential backoff multiplier
  
  # Maintenance
  auto_vacuum: true                    # Enable auto-vacuum
  wal_mode: true                       # Write-Ahead Logging
  checkpoint_interval: 1000            # WAL checkpoint interval
  backup_retention_days: 30            # Keep backups for N days
```

### 4. **Task Management Defaults**

#### Current Hard-Coded Values
- **Default status**: `"Todo"` (task creation)
- **Default priority**: `"medium"` (task creation)
- **Default status for approval**: `"Review"` → `"Todo"`
- **Stale threshold**: `2 hours` (task cleanup)
- **Task ID format**: `T-{epoch_milliseconds}`

#### Recommended Configuration
```yaml
# .exarp/config.yaml
tasks:
  # Defaults
  default_status: "Todo"
  default_priority: "medium"
  default_tags: []                     # Default tags for new tasks
  
  # Status transitions
  status_workflow:
    todo: ["In Progress", "Review"]
    "In Progress": ["Done", "Review", "Todo"]
    review: ["Todo", "In Progress", "Done"]
    done: []                            # Terminal state
  
  # Cleanup
  stale_threshold_hours: 2             # Consider task stale after N hours
  auto_cleanup_enabled: false           # Auto-cleanup stale tasks
  cleanup_dry_run: true                # Preview cleanup before applying
  
  # Task ID
  id_format: "T-{epoch_milliseconds}"  # Task ID format
  id_prefix: "T-"                      # Task ID prefix
  
  # Clarity checks
  min_description_length: 50           # Minimum description length
  require_description: false           # Require description for approval
  auto_clarify: false                  # Auto-request clarification
```

### 5. **Security Settings**

#### Current Hard-Coded Values
- **Rate limit**: `100 requests/minute`
- **Rate limit window**: `1 minute`
- **Path validation**: Basic (some tools missing)
- **File size limits**: Not enforced consistently

#### Recommended Configuration
```yaml
# .exarp/config.yaml
security:
  # Rate limiting
  rate_limit:
    enabled: true
    requests_per_window: 100
    window_duration: 1m
    burst_size: 10                     # Allow burst of N requests
  
  # Path validation
  path_validation:
    enabled: true
    allow_absolute_paths: false         # Only allow relative paths
    max_depth: 20                      # Max directory depth
    blocked_patterns:                  # Blocked path patterns
      - "**/.git/**"
      - "**/node_modules/**"
      - "**/vendor/**"
  
  # File operations
  file_limits:
    max_file_size: 10485760            # 10MB
    max_files_per_operation: 1000      # Max files in batch ops
    allowed_extensions:                # Whitelist (empty = all allowed)
      - ".go"
      - ".py"
      - ".md"
      - ".json"
      - ".yaml"
      - ".yml"
  
  # Access control
  access_control:
    enabled: false                     # Enable ACL checks
    default_policy: "allow"            # allow | deny
    restricted_tools: []               # Tools requiring special permission
```

### 6. **Logging & Output**

#### Current Hard-Coded Values
- **Log level**: Various (info, debug, error)
- **Output format**: JSON/text (varies by tool)
- **Log file location**: Various (session logs, etc.)
- **Log retention**: Not configured

#### Recommended Configuration
```yaml
# .exarp/config.yaml
logging:
  # Levels
  level: "info"                       # debug | info | warn | error
  tool_level: "info"                  # Tool-specific log level
  framework_level: "warn"             # Framework log level
  
  # Output
  format: "json"                      # json | text | structured
  include_timestamps: true
  include_caller: false                # Include file:line in logs
  color_output: true                   # Colorize terminal output
  
  # Files
  log_dir: ".exarp/logs"              # Log directory
  log_file: "exarp.log"               # Main log file
  session_log_dir: ".exarp/logs/sessions"  # Session logs
  log_rotation:
    enabled: true
    max_size: 10485760                 # 10MB per file
    max_files: 10                      # Keep N rotated files
    compress: true                     # Compress old logs
  
  # Retention
  retention_days: 30                  # Keep logs for N days
  auto_cleanup: true                   # Auto-cleanup old logs
```

### 7. **Tool-Specific Settings**

#### Current Hard-Coded Values
- **Scorecard default scores**: `50.0` (all metrics)
- **Report output format**: `"text"` (default)
- **Linting linter**: `"auto"` (auto-detect)
- **Testing framework**: `"auto"` (auto-detect)
- **MLX model**: `"mlx-community/Phi-3.5-mini-instruct-4bit"`
- **Ollama model**: `"llama3.2"` (default)
- **Context summarization level**: `"brief"` (default)

#### Recommended Configuration
```yaml
# .exarp/config.yaml
tools:
  # Scorecard
  scorecard:
    default_scores:
      security: 50.0
      testing: 50.0
      documentation: 50.0
      completion: 50.0
      alignment: 50.0
      clarity: 50.0
      cicd: 50.0
      dogfooding: 50.0
    include_wisdom: true               # Include advisor wisdom
    output_format: "text"              # text | json | html
  
  # Reporting
  report:
    default_format: "text"             # text | json | html | markdown
    default_output_path: ""            # Default output location
    include_metrics: true               # Include metrics in reports
    include_recommendations: true       # Include recommendations
  
  # Linting
  linting:
    default_linter: "auto"              # auto | gofmt | goimports | golangci-lint
    auto_fix: false                    # Auto-fix issues
    include_hints: true                 # Include tool hints
    timeout: 60s
  
  # Testing
  testing:
    default_framework: "auto"          # auto | go | pytest | jest
    min_coverage: 80                   # Minimum coverage %
    coverage_format: "html"            # html | text | json
    verbose: false                     # Verbose test output
  
  # MLX
  mlx:
    default_model: "mlx-community/Phi-3.5-mini-instruct-4bit"
    default_max_tokens: 512
    default_temperature: 0.7
    verbose: false
  
  # Ollama
  ollama:
    default_model: "llama3.2"
    default_host: "http://localhost:11434"
    default_context_size: 4096
    default_num_threads: 4
    default_num_gpu: 1
  
  # Context
  context:
    default_budget: 4000               # Default token budget
    default_summarization_level: "brief"  # brief | detailed | key_metrics | actionable
    tokens_per_char: 0.25              # Token estimation ratio
    include_raw: false                  # Include raw data in summaries
```

### 8. **Workflow Mode Settings**

#### Current Hard-Coded Values
- **Default mode**: `"development"` (session priming)
- **Mode suggestions**: Based on time of day
- **Tool groups**: Hard-coded per mode
- **Focus mode reduction**: `50-80%` (estimated)

#### Recommended Configuration
```yaml
# .exarp/config.yaml
workflow:
  # Defaults
  default_mode: "development"          # development | security_review | task_management
  auto_detect_mode: true                # Auto-detect based on time/context
  mode_suggestions:
    morning: "daily_checkin"           # Suggested mode for morning
    afternoon: "development"           # Suggested mode for afternoon
    evening: "review"                  # Suggested mode for evening
  
  # Mode-specific settings
  modes:
    development:
      enabled_tools: "*"               # All tools (or list specific)
      disabled_tools: []                # Disabled tools
      tool_limit: 0                     # 0 = no limit
    
    security_review:
      enabled_tools: ["security", "health", "linting"]
      disabled_tools: ["*"]             # Disable all others
      tool_limit: 10
    
    task_management:
      enabled_tools: ["task_*", "automation", "session"]
      disabled_tools: []
      tool_limit: 15
  
  # Focus mode
  focus:
    enabled: true
    reduction_target: 70               # Target % reduction in visible tools
    preserve_core_tools: true          # Always show core tools
```

### 9. **Memory & Session Settings**

#### Current Hard-Coded Values
- **Memory categories**: Hard-coded list
- **Session log location**: Hard-coded paths
- **Memory retention**: Not configured
- **Session limit**: `5` (recent tasks)

#### Recommended Configuration
```yaml
# .exarp/config.yaml
memory:
  # Categories
  categories:
    - debug
    - research
    - architecture
    - preference
    - insight
  
  # Storage
  storage_path: ".exarp/memory"        # Memory storage location
  session_log_path: ".exarp/logs/sessions"  # Session logs
  
  # Retention
  retention_days: 90                   # Keep memories for N days
  auto_cleanup: true                   # Auto-cleanup old memories
  max_memories: 1000                  # Max memories per category
  
  # Maintenance
  consolidation:
    enabled: true
    similarity_threshold: 0.85         # Merge similar memories
    frequency: "weekly"               # Consolidation frequency

session:
  # Priming
  auto_prime: true                     # Auto-prime on session start
  include_hints: true                  # Include tool hints
  include_tasks: true                  # Include recent tasks
  task_limit: 5                       # Max recent tasks to show
  
  # Handoffs
  enable_handoffs: true                # Enable handoff notes
  handoff_retention_days: 7            # Keep handoffs for N days
  
  # Tracking
  track_prompts: true                  # Track prompt usage
  prompt_log_path: ".exarp/logs/prompts"
```

### 10. **Project-Specific Settings**

#### Recommended Configuration
```yaml
# .exarp/config.yaml
project:
  # Identification
  name: ""                             # Project name (auto-detected)
  type: "library"                     # library | application | service | tool
  language: "go"                      # Primary language
  
  # Paths
  root: "."                            # Project root (auto-detected)
  todo2_path: ".todo2"                 # Todo2 directory
  exarp_path: ".exarp"                # Exarp config directory
  
  # Features
  features:
    sqlite_enabled: true               # Use SQLite database
    json_fallback: true                # Fallback to JSON if DB unavailable
    python_bridge: true               # Enable Python bridge
    mcp_servers: []                    # Additional MCP servers to use
  
  # Overrides
  skip_checks: []                      # Skip specific health checks
  custom_tools: []                     # Custom tool definitions
```

---

## Configuration File Structure

### Recommended File Organization

```
.exarp/
  ├── config.yaml              # Main configuration (all categories above)
  ├── automations.yaml         # Automation workflows (from previous doc)
  ├── tools.yaml               # Tool-specific overrides (optional)
  └── secrets.yaml              # Sensitive config (gitignored)
```

### Alternative: Single File

```
.exarp/
  └── config.yaml              # All configuration in one file
```

**Recommendation:** Start with single file (`config.yaml`), split later if needed.

---

## Implementation Priority

### Phase 1: High-Value Quick Wins
1. **Timeouts** - Easy to implement, high impact
2. **Thresholds** - Frequently adjusted, user-facing
3. **Task defaults** - Directly affects user workflow

### Phase 2: Medium Priority
4. **Database settings** - Important for performance
5. **Security settings** - Important for production
6. **Logging** - Useful for debugging

### Phase 3: Nice to Have
7. **Tool-specific settings** - Can use tool parameters for now
8. **Workflow modes** - Advanced feature
9. **Memory settings** - Lower priority

---

## Migration Strategy

### Step 1: Add Config Loading
- Create `internal/config/config.go` with all categories
- Load from `.exarp/config.yaml`
- Fallback to hard-coded defaults

### Step 2: Replace Hard-Coded Values
- One category at a time
- Start with timeouts (easiest)
- Test thoroughly before moving to next

### Step 3: CLI Tool
```bash
exarp-go config init          # Generate default config
exarp-go config validate       # Validate config
exarp-go config show           # Show current config
exarp-go config set <key> <value>  # Set config value
```

### Step 4: Documentation
- Configuration reference guide
- Examples for different project types
- Migration guide from hard-coded to config

---

## Benefits

1. **Flexibility** - Per-project customization
2. **Maintainability** - Single source of truth
3. **Testing** - Easy to test different configurations
4. **User Experience** - Non-developers can customize
5. **Documentation** - Config file is self-documenting
6. **Version Control** - Config changes tracked in git

---

## Questions for Discussion

1. **File structure**: Single `config.yaml` or split by category?
2. **Priority**: Which categories should be implemented first?
3. **Backward compatibility**: Keep all hard-coded defaults forever?
4. **Validation**: How strict should config validation be?
5. **CLI tool**: Essential for v1 or can wait?

---

## Related Documents

- `docs/AUTOMATION_CONFIGURATION_ANALYSIS.md` - Automation workflows
- `internal/config/config.go` - Current config structure
- `internal/tools/automation_native.go` - Automation implementation

---

## Summary

This document identifies **10 major categories** of configurable parameters:

1. ✅ **Timeouts & Durations** (15+ values)
2. ✅ **Thresholds & Limits** (20+ values)
3. ✅ **Database Configuration** (10+ values)
4. ✅ **Task Management Defaults** (10+ values)
5. ✅ **Security Settings** (15+ values)
6. ✅ **Logging & Output** (15+ values)
7. ✅ **Tool-Specific Settings** (30+ values)
8. ✅ **Workflow Mode Settings** (10+ values)
9. ✅ **Memory & Session Settings** (10+ values)
10. ✅ **Project-Specific Settings** (10+ values)

**Total: 150+ configurable parameters identified**

**Recommendation:** Implement in phases, starting with high-value quick wins (timeouts, thresholds, task defaults).
