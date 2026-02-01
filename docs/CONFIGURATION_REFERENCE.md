# Configuration Reference

**Date:** 2026-02-01  
**Status:** Reference  
**Related:** `CONFIGURATION_IMPLEMENTATION_PLAN.md`, `CONFIGURATION_PROTOBUF_INTEGRATION.md`

---

## Overview

exarp-go uses a protobuf-based configuration system. At runtime, only `.exarp/config.pb` is loaded. YAML is used for human editing via export/convert.

### File Format

| Format | Location | Purpose |
|--------|----------|---------|
| Protobuf (binary) | `.exarp/config.pb` | Runtime config (required for file-based config) |
| YAML | — | Export for editing only (not loaded at runtime) |

### Editing Workflow

```bash
# Export current config to YAML for editing
exarp-go config export yaml > config.yaml

# Edit config.yaml in your editor
vim config.yaml

# Convert YAML back to protobuf and save
exarp-go config convert yaml protobuf
```

### CLI Commands

| Command | Description |
|---------|-------------|
| `exarp-go config init` | Create `.exarp/config.pb` with defaults |
| `exarp-go config validate` | Validate configuration |
| `exarp-go config show [yaml\|json]` | Display current config |
| `exarp-go config export [yaml\|json\|protobuf]` | Export to format |
| `exarp-go config convert <from> <to>` | Convert between formats (yaml ↔ protobuf) |

---

## Parameter Reference

Values shown are defaults. Durations use Go format (e.g., `30m`, `60s`). Omitted keys inherit defaults.

### Top-Level

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `version` | string | `"1.0"` | Config schema version |

---

### timeouts

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `task_lock_lease` | duration | `30m` | Task lock duration |
| `task_lock_renewal` | duration | `20m` | When to renew lease before expiry |
| `stale_lock_threshold` | duration | `5m` | Consider lock stale after this |
| `tool_default` | duration | `60s` | Default tool timeout |
| `tool_scorecard` | duration | `60s` | Scorecard generation timeout |
| `tool_linting` | duration | `60s` | Linting timeout |
| `tool_testing` | duration | `300s` | Test execution timeout |
| `tool_report` | duration | `60s` | Report generation timeout |
| `ollama_download` | duration | `600s` | Model download timeout |
| `ollama_generate` | duration | `300s` | Text generation timeout |
| `http_client` | duration | `30s` | HTTP client timeout |
| `database_retry` | duration | `60s` | Database retry timeout |
| `context_summarize` | duration | `30s` | Context summarization timeout |
| `context_budget` | duration | `10s` | Budget calculation timeout |

---

### thresholds

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `similarity_threshold` | float | `0.85` | Duplicate task detection |
| `min_description_length` | int | `50` | Min task description length |
| `min_task_confidence` | float | `0.7` | Task confidence threshold |
| `min_coverage` | int | `80` | Minimum test coverage % |
| `min_test_confidence` | float | `0.7` | Test suggestion confidence |
| `min_estimation_confidence` | float | `0.7` | Task estimation confidence |
| `mlx_weight` | float | `0.3` | MLX model weight in estimation |
| `max_parallel_tasks` | int | `10` | Max concurrent tasks (automation) |
| `max_tasks_per_host` | int | `5` | Max tasks per agent/host |
| `max_automation_iterations` | int | `10` | Max automation loop iterations |
| `tokens_per_char` | float | `0.25` | Token estimation ratio |
| `default_context_budget` | int | `4000` | Default token budget |
| `context_reduction_threshold` | float | `0.5` | When to suggest summarization |
| `rate_limit_requests` | int | `100` | Requests per window |
| `rate_limit_window` | duration | `1m` | Rate limit window |
| `max_file_size` | int64 | `10485760` | Max file size (10MB) |
| `max_path_depth` | int | `20` | Max directory depth |

---

### tasks

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `default_status` | string | `"Todo"` | New task default status |
| `default_priority` | string | `"medium"` | New task default priority |
| `default_tags` | []string | `[]` | Default tags for new tasks |
| `status_workflow` | map | See below | Status → allowed next statuses |
| `stale_threshold_hours` | int | `2` | Hours before task considered stale |
| `auto_cleanup_enabled` | bool | `false` | Enable auto cleanup |
| `cleanup_dry_run` | bool | `true` | Dry run for cleanup |
| `id_format` | string | `"T-{epoch_milliseconds}"` | Task ID format |
| `id_prefix` | string | `"T-"` | Task ID prefix |
| `min_description_length` | int | `50` | Min description for clarity check |
| `require_description` | bool | `false` | Require description on create |
| `auto_clarify` | bool | `false` | Auto-run clarity check |

---

### database

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `sqlite_path` | string | `".todo2/todo2.db"` | SQLite database path |
| `json_fallback_path` | string | `".todo2/state.todo2.json"` | JSON fallback path |
| `backup_path` | string | `".todo2/backups"` | Backup directory |
| `max_connections` | int | `10` | Max DB connections |
| `connection_timeout` | duration | `30s` | Connection timeout |
| `query_timeout` | duration | `60s` | Query timeout |
| `retry_attempts` | int | `3` | Retry attempts |
| `retry_initial_delay` | duration | `100ms` | Initial retry delay |
| `retry_max_delay` | duration | `5s` | Max retry delay |
| `retry_multiplier` | float | `2.0` | Retry backoff multiplier |
| `auto_vacuum` | bool | `true` | SQLite auto vacuum |
| `wal_mode` | bool | `true` | WAL mode |
| `checkpoint_interval` | int | `1000` | Checkpoint interval |
| `backup_retention_days` | int | `30` | Backup retention |

---

### security

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `security.rate_limit.enabled` | bool | `true` | Enable rate limiting |
| `security.rate_limit.requests_per_window` | int | `100` | Requests per window |
| `security.rate_limit.window_duration` | duration | `1m` | Window duration |
| `security.rate_limit.burst_size` | int | `10` | Burst size |
| `security.path_validation.enabled` | bool | `true` | Enable path validation |
| `security.path_validation.allow_absolute_paths` | bool | `false` | Allow absolute paths |
| `security.path_validation.max_depth` | int | `20` | Max path depth |
| `security.path_validation.blocked_patterns` | []string | `["**/.git/**", ...]` | Blocked path patterns |
| `security.file_limits.max_file_size` | int64 | `10485760` | Max file size |
| `security.file_limits.max_files_per_operation` | int | `1000` | Max files per op |
| `security.access_control.enabled` | bool | `false` | Enable access control |
| `security.access_control.default_policy` | string | `"allow"` | Default policy |

---

### logging

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `level` | string | `"info"` | Log level |
| `tool_level` | string | `"info"` | Tool log level |
| `framework_level` | string | `"warn"` | Framework log level |
| `format` | string | `"json"` | Output format (json, text) |
| `include_timestamps` | bool | `true` | Include timestamps |
| `include_caller` | bool | `false` | Include caller info |
| `color_output` | bool | `true` | Colorized output |
| `log_dir` | string | `".exarp/logs"` | Log directory |
| `log_file` | string | `"exarp.log"` | Log file name |
| `session_log_dir` | string | `".exarp/logs/sessions"` | Session logs |
| `retention_days` | int | `30` | Log retention |
| `auto_cleanup` | bool | `true` | Auto cleanup old logs |
| `log_rotation.enabled` | bool | `true` | Enable rotation |
| `log_rotation.max_size` | int64 | `10485760` | Max size (10MB) |
| `log_rotation.max_files` | int | `10` | Max rotated files |
| `log_rotation.compress` | bool | `true` | Compress rotated |

---

### tools

Tool-specific overrides. Each tool has its own sub-section.

| Section | Key Parameters | Defaults |
|---------|----------------|----------|
| `tools.scorecard` | `default_scores`, `include_wisdom`, `output_format` | scores: 50 each, wisdom: true, format: text |
| `tools.report` | `default_format`, `include_metrics`, `include_recommendations` | format: text, metrics: true |
| `tools.linting` | `default_linter`, `auto_fix`, `timeout` | linter: auto, fix: false, timeout: 60s |
| `tools.testing` | `min_coverage`, `coverage_format`, `verbose` | coverage: 80, format: html |
| `tools.mlx` | `default_model`, `default_max_tokens`, `default_temperature` | Phi-3.5-mini, 512, 0.7 |
| `tools.ollama` | `default_model`, `default_host`, `default_context_size` | llama3.2, localhost:11434, 4096 |
| `tools.context` | `default_budget`, `tokens_per_char` | 4000, 0.25 |

---

### workflow

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `default_mode` | string | `"development"` | Default workflow mode |
| `auto_detect_mode` | bool | `true` | Auto-detect mode by time |
| `mode_suggestions.morning` | string | `"daily_checkin"` | Morning mode |
| `mode_suggestions.afternoon` | string | `"development"` | Afternoon mode |
| `mode_suggestions.evening` | string | `"review"` | Evening mode |
| `focus.enabled` | bool | `true` | Focus mode enabled |
| `focus.reduction_target` | int | `70` | Token reduction target % |
| `focus.preserve_core_tools` | bool | `true` | Preserve core tools |

---

### memory

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `categories` | []string | `[debug, research, ...]` | Memory categories |
| `storage_path` | string | `".exarp/memories"` | Memory storage path |
| `retention_days` | int | `90` | Memory retention |
| `auto_cleanup` | bool | `true` | Auto cleanup |
| `max_memories` | int | `1000` | Max memories |
| `consolidation.enabled` | bool | `true` | Enable consolidation |
| `consolidation.similarity_threshold` | float | `0.85` | Similarity for merge |
| `consolidation.frequency` | string | `"weekly"` | Consolidation frequency |

---

### project

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `name` | string | `""` | Project name (auto-detected) |
| `type` | string | `"library"` | library \| application \| service \| tool |
| `language` | string | `"go"` | Primary language |
| `root` | string | `"."` | Project root |
| `todo2_path` | string | `".todo2"` | Todo2 path |
| `exarp_path` | string | `".exarp"` | Exarp path |
| `features.sqlite_enabled` | bool | `true` | SQLite enabled |
| `features.json_fallback` | bool | `true` | JSON fallback |
| `features.python_bridge` | bool | `true` | Python bridge |
| `skip_checks` | []string | `[]` | Skip health checks |
| `custom_tools` | []string | `[]` | Custom tools |

---

## Examples by Project Type

### Minimal (Defaults Only)

No config file. exarp-go uses in-memory defaults. Suitable for quick tryouts.

```bash
# No .exarp/config.pb - everything uses defaults
exarp-go -tool report -args '{"action":"overview"}'
```

---

### Go Library (Light Customization)

```yaml
version: "1.0"

project:
  type: library
  language: go

thresholds:
  min_coverage: 90
  min_description_length: 100

tools:
  testing:
    min_coverage: 90
    coverage_format: html
  linting:
    default_linter: golangci-lint
    auto_fix: true
```

---

### Full Application (Broader Settings)

```yaml
version: "1.0"

project:
  type: application
  language: go
  name: my-app

timeouts:
  tool_testing: 600s
  ollama_generate: 120s

thresholds:
  min_coverage: 80
  max_automation_iterations: 20

tasks:
  default_priority: high
  default_tags: [app, production]
  stale_threshold_hours: 24
  auto_cleanup_enabled: true
  cleanup_dry_run: false

database:
  backup_retention_days: 14

logging:
  level: debug
  format: text
  color_output: true

tools:
  report:
    include_metrics: true
    include_recommendations: true
  ollama:
    default_model: llama3.2
    default_host: http://localhost:11434

memory:
  retention_days: 30
  max_memories: 500
```

---

### Strict CI (Short Timeouts, High Thresholds)

```yaml
version: "1.0"

project:
  type: tool
  language: go

timeouts:
  tool_default: 30s
  tool_testing: 120s
  tool_linting: 30s
  http_client: 10s

thresholds:
  min_coverage: 95
  min_task_confidence: 0.9
  min_description_length: 80
  max_parallel_tasks: 4
  max_automation_iterations: 5

tasks:
  require_description: true
  min_description_length: 80

security:
  rate_limit:
    enabled: true
    requests_per_window: 50
    window_duration: 1m
  path_validation:
    enabled: true
    allow_absolute_paths: false
    max_depth: 10

tools:
  testing:
    min_coverage: 95
    verbose: true
  linting:
    timeout: 30s
```

---

## Applying Examples

1. Save the YAML example to a file (e.g., `config.yaml`).
2. Convert to protobuf: `exarp-go config convert yaml protobuf` (reads stdin or `config.yaml`).
3. Or init first, then export, edit, convert:
   ```bash
   exarp-go config init
   exarp-go config export yaml > config.yaml
   # Edit config.yaml
   exarp-go config convert yaml protobuf
   ```
4. Validate: `exarp-go config validate`

---

## Related Documentation

- [CONFIGURATION_IMPLEMENTATION_PLAN.md](CONFIGURATION_IMPLEMENTATION_PLAN.md) — Implementation plan
- [CONFIGURATION_PROTOBUF_INTEGRATION.md](CONFIGURATION_PROTOBUF_INTEGRATION.md) — Protobuf integration
- [CONFIGURABLE_PARAMETERS_RECOMMENDATIONS.md](CONFIGURABLE_PARAMETERS_RECOMMENDATIONS.md) — Extended recommendations
