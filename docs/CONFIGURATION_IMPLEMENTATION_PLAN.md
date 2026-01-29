# Configuration System Implementation Plan

**Date:** 2026-01-13  
**Status:** Implementation Plan  
**Related:** `AUTOMATION_CONFIGURATION_ANALYSIS.md`, `CONFIGURABLE_PARAMETERS_RECOMMENDATIONS.md`

---

## Executive Summary

This document outlines the implementation plan for a comprehensive configuration system that replaces hard-coded values with per-project configuration. **Protobuf is mandatory** for file-based config: the runtime loads only `.exarp/config.pb`. YAML is supported for editing via `exarp-go config export yaml` and `exarp-go config convert yaml protobuf`.

---

## Goals

1. **Replace hard-coded values** with configurable settings (protobuf binary format)
2. **Maintain backward compatibility** with sensible defaults (no file = defaults)
3. **Enable per-project customization** without code changes
4. **Provide validation and error handling** for configuration
5. **Create CLI tools** for config management (init, validate, show, set, export, convert)

---

## Architecture

### Configuration File Structure (protobuf mandatory)

```
.exarp/
  └── config.pb      # Main configuration file (protobuf binary; mandatory for file-based config)
```

- **Runtime**: Only `.exarp/config.pb` is loaded. If only `config.yaml` exists, the loader returns an error instructing the user to run `exarp-go config convert yaml protobuf` or `exarp-go config init`.
- **Editing**: Use `exarp-go config export yaml` to emit YAML, edit, then `exarp-go config convert yaml protobuf` to write back.

### Configuration Schema

```yaml
version: "1.0"

# All configuration categories (see CONFIGURABLE_PARAMETERS_RECOMMENDATIONS.md)
automations: { ... }
timeouts: { ... }
thresholds: { ... }
database: { ... }
tasks: { ... }
security: { ... }
logging: { ... }
tools: { ... }
workflow: { ... }
memory: { ... }
project: { ... }
```

### Code Structure

```
internal/
  config/
    config.go              # MCP server config (Load); full config in loader
    defaults.go            # Default values (backward compatible)
    schema.go              # Configuration schema/types
    validation.go          # Config validation
    loader.go              # LoadConfig from .exarp/config.pb only; LoadConfigYAML for convert; WriteConfigToProtobufFile
    protobuf.go            # ToProtobuf / FromProtobuf
  tools/
    # Tools updated to use config instead of hard-coded values
```

---

## Implementation Phases

### Phase 1: Foundation (Week 1)

**Goal:** Create configuration infrastructure and implement high-value quick wins

#### Tasks:
1. ✅ Create config package structure
2. ✅ Implement YAML loading with defaults
3. ✅ Implement config validation
4. ✅ Implement timeouts configuration
5. ✅ Implement thresholds configuration
6. ✅ Implement task defaults configuration
7. ✅ Create CLI tool for config management
8. ✅ Update tools to use config values
9. ✅ Update TUI to use config values
10. ✅ Update documentation

**Deliverables:**
- `internal/config/` package with full structure
- `.exarp/config.yaml` support
- Timeouts, thresholds, and task defaults configurable
- CLI tool: `exarp-go config`

### Phase 1.5: Protobuf Integration ✅ (Protobuf mandatory)

**Goal:** Configuration is stored and loaded as protobuf only; YAML is for export/import only.

**Status:** ✅ Done (protobuf mandatory)  
**Runtime behavior:** `LoadConfig` loads only `.exarp/config.pb`. If only `config.yaml` exists, it returns an error. No config file = defaults.

#### Tasks:
1. ✅ Create protobuf conversion layer (Go structs ↔ protobuf messages)
2. ✅ Loader loads only protobuf; `LoadConfigYAML` for convert/import
3. ✅ CLI: init/set write `.exarp/config.pb`; export/convert for yaml ↔ protobuf
4. ✅ Schema synchronization validation (see `internal/config/protobuf_test.go`)
5. ✅ Documentation updated (this plan; protobuf mandatory)

**Deliverables:**
- `internal/config/protobuf.go` - Conversion functions
- `internal/config/loader.go` - Load from `.exarp/config.pb` only; `WriteConfigToProtobufFile`; `LoadConfigYAML` for convert
- `internal/cli/config.go` - init/set write config.pb; export/convert
- `docs/CONFIGURATION_PROTOBUF_INTEGRATION.md` - Detailed plan

**See:** `docs/CONFIGURATION_PROTOBUF_INTEGRATION.md` for complete implementation details

---

### Phase 2: Database & Security (Week 2)

**Goal:** Configure database and security settings

#### Tasks:
1. Implement database configuration
2. Implement security settings (rate limiting, path validation)
3. Update database code to use config
4. Update security code to use config
5. Add config validation for security settings

**Deliverables:**
- Database settings configurable
- Security settings configurable
- Rate limiting configurable

### Phase 3: Logging & Tools (Week 3)

**Goal:** Configure logging and tool-specific settings

#### Tasks:
1. Implement logging configuration
2. Implement tool-specific settings
3. Update logging code to use config
4. Update tool handlers to use config
5. Add tool config validation

**Deliverables:**
- Logging fully configurable
- Tool-specific settings configurable
- All tools respect configuration

### Phase 4: Workflow & Memory (Week 4)

**Goal:** Configure workflow modes and memory settings

#### Tasks:
1. Implement workflow mode configuration
2. Implement memory/session configuration
3. Update workflow mode code to use config
4. Update memory code to use config
5. Add project-specific settings

**Deliverables:**
- Workflow modes configurable
- Memory/session settings configurable
- Project-specific overrides working

### Phase 5: Automation Configuration (Week 5)

**Goal:** Make automation workflows fully configurable

#### Tasks:
1. Implement automation configuration loading
2. Refactor automation tool to use config
3. Update Makefile targets to use config
4. Update prompt templates to reference config
5. Add automation validation

**Deliverables:**
- All automations configurable via YAML
- Makefile targets use config
- Prompt templates reference config

### Phase 6: Testing & Documentation (Week 6)

**Goal:** Comprehensive testing and documentation

#### Tasks:
1. Write unit tests for config package
2. Write integration tests for config usage
3. Test backward compatibility (no config file)
4. Test with various config files
5. Update all documentation
6. Create migration guide
7. Create examples for different project types

**Deliverables:**
- Full test coverage
- Complete documentation
- Migration guide
- Example configurations

---

## Detailed Task Breakdown

### Phase 1 Tasks

#### T1.1: Create Config Package Structure
- Create `internal/config/` directory
- Create `config.go` with main Config struct
- Create `defaults.go` with all default values
- Create `schema.go` with configuration types
- Create `validation.go` with validation logic
- Create `loader.go` with YAML loading

**Files:**
- `internal/config/config.go`
- `internal/config/defaults.go`
- `internal/config/schema.go`
- `internal/config/validation.go`
- `internal/config/loader.go`

**Estimated:** 2 hours

#### T1.2: Implement YAML Loading
- Implement YAML file loading
- Implement default value merging
- Implement environment variable overrides
- Add error handling and validation
- Add config file location detection

**Files:**
- `internal/config/loader.go`

**Estimated:** 3 hours

#### T1.3: Implement Timeouts Configuration
- Define timeout configuration struct
- Add default timeout values
- Create timeout getter functions
- Update tools to use config timeouts
- Test timeout configuration

**Files:**
- `internal/config/schema.go` (timeouts section)
- `internal/config/defaults.go` (timeout defaults)
- `internal/tools/*.go` (update to use config)

**Estimated:** 4 hours

#### T1.4: Implement Thresholds Configuration
- Define threshold configuration struct
- Add default threshold values
- Create threshold getter functions
- Update tools to use config thresholds
- Test threshold configuration

**Files:**
- `internal/config/schema.go` (thresholds section)
- `internal/config/defaults.go` (threshold defaults)
- `internal/tools/*.go` (update to use config)

**Estimated:** 4 hours

#### T1.5: Implement Task Defaults Configuration
- Define task configuration struct
- Add default task values
- Create task getter functions
- Update task workflow to use config
- Test task configuration

**Files:**
- `internal/config/schema.go` (tasks section)
- `internal/config/defaults.go` (task defaults)
- `internal/tools/task_workflow*.go` (update to use config)

**Estimated:** 3 hours

#### T1.6: Create CLI Tool for Config
- Add `config` subcommand to CLI
- Implement `config init` (generate default config)
- Implement `config validate` (validate config file)
- Implement `config show` (display current config)
- Implement `config set <key> <value>` (set config value)
- Add help text and examples

**Files:**
- `internal/cli/config.go`
- `cmd/cli/main.go` (add config command)

**Estimated:** 4 hours

#### T1.7: Update Documentation
- Document configuration file format
- Document all configurable parameters
- Add examples for different project types
- Update README with config information

**Files:**
- `docs/CONFIGURATION_REFERENCE.md`
- `docs/CONFIGURATION_EXAMPLES.md`
- `README.md`

**Estimated:** 2 hours

**Phase 1 Total:** ~22 hours

---

## Implementation Details

### Configuration Loading Strategy

```go
// Load config with fallback to defaults
func LoadConfig(projectRoot string) (*Config, error) {
    // 1. Try to load .exarp/config.yaml
    // 2. If not found, use defaults
    // 3. Merge with environment variables
    // 4. Validate configuration
    // 5. Return config
}
```

### Default Values Strategy

```go
// All defaults match current hard-coded values
// Ensures backward compatibility
func GetDefaults() *Config {
    return &Config{
        Timeouts: TimeoutsConfig{
            TaskLockLease: 30 * time.Minute,
            ToolDefault: 60 * time.Second,
            // ... all current hard-coded values
        },
        // ... all other defaults
    }
}
```

### Validation Strategy

```go
// Validate configuration on load
func ValidateConfig(cfg *Config) error {
    // Check required fields
    // Check value ranges
    // Check file paths exist
    // Check dependencies
    return nil
}
```

---

## Testing Strategy

### Unit Tests
- Config loading with various YAML files
- Default value merging
- Environment variable overrides
- Validation logic
- Error handling

### Integration Tests
- Tools using config values
- Backward compatibility (no config file)
- Config file updates reflected in runtime
- CLI tool functionality

### Test Cases
1. **No config file** - Should use defaults
2. **Partial config** - Should merge with defaults
3. **Invalid config** - Should return clear errors
4. **Config updates** - Should reload on change (future)
5. **Environment overrides** - Should override file config

---

## Migration Path

### For Existing Projects

1. **No action required** - Defaults match current behavior
2. **Optional:** Run `exarp-go config init` to generate config file
3. **Optional:** Customize config file as needed
4. **Gradual:** Migrate hard-coded values to config over time

### For New Projects

1. Run `exarp-go config init` to create config file
2. Customize config for project needs
3. Commit config file to version control

---

## Success Criteria

### Phase 1 Complete When:
- ✅ Config package created and tested
- ✅ Timeouts configurable via YAML
- ✅ Thresholds configurable via YAML
- ✅ Task defaults configurable via YAML
- ✅ CLI tool functional
- ✅ Documentation complete
- ✅ Backward compatible (no breaking changes)

### Full Implementation Complete When:
- ✅ All 10 configuration categories implemented
- ✅ All hard-coded values replaced
- ✅ Full test coverage
- ✅ Complete documentation
- ✅ Migration guide available
- ✅ Example configs for different project types

---

## Risks & Mitigation

### Risk 1: Breaking Changes
**Mitigation:** All defaults match current hard-coded values. No config file = current behavior.

### Risk 2: Performance Impact
**Mitigation:** Load config once at startup, cache in memory. Lazy loading for rarely-used values.

### Risk 3: Config File Complexity
**Mitigation:** Start with single file, split later if needed. Good documentation and examples.

### Risk 4: Validation Errors
**Mitigation:** Clear error messages, validation on load, CLI tool for validation.

---

## Timeline

- **Week 1:** Phase 1 (Foundation) - 22 hours
- **Week 2:** Phase 2 (Database & Security) - 16 hours
- **Week 3:** Phase 3 (Logging & Tools) - 20 hours
- **Week 4:** Phase 4 (Workflow & Memory) - 16 hours
- **Week 5:** Phase 5 (Automation) - 20 hours
- **Week 6:** Phase 6 (Testing & Docs) - 16 hours

**Total:** ~110 hours (6 weeks)

---

## Next Steps

1. ✅ Create implementation plan (this document)
2. ✅ Create tasks in todo2 system
3. ✅ Complete Phase 1 (Foundation)
4. ✅ Complete Phase 1.5 (Protobuf Integration — see `CONFIGURATION_PROTOBUF_INTEGRATION.md`)
5. ⏳ Continue with Phase 2 (Database & Security)
6. ⏳ Continue with subsequent phases

---

## Related Documents

- `docs/AUTOMATION_CONFIGURATION_ANALYSIS.md` - Automation analysis
- `docs/CONFIGURABLE_PARAMETERS_RECOMMENDATIONS.md` - All parameters
- `docs/CONFIGURATION_PROTOBUF_INTEGRATION.md` - **NEW:** Protobuf integration plan (Phase 1.5)
- `docs/PROTOBUF_ANALYSIS.md` - Protobuf usage analysis
- `internal/config/config.go` - Current config (minimal)
- `internal/tools/automation_native.go` - Automation implementation
- `proto/config.proto` - Protobuf schema definition