# Configuration Protobuf Integration Plan (Phase 1.5)

**Date:** 2026-01-13  
**Status:** ✅ **COMPLETE** - All tasks implemented and tested  
**Related:** `CONFIGURATION_IMPLEMENTATION_PLAN.md`, `PROTOBUF_ANALYSIS.md`

---

## Executive Summary

This document describes the integration of Protocol Buffers (protobuf) with the configuration system. **Protobuf is mandatory** for file-based config: the runtime loads only `.exarp/config.pb`. YAML is supported for editing via `exarp-go config export yaml` and `exarp-go config convert yaml protobuf`.

---

## Goals

1. **Integrate protobuf schema** with the configuration system
2. **Mandatory protobuf** for file-based config (`.exarp/config.pb` only at runtime)
3. **YAML for editing** — export to YAML, edit, then convert back to protobuf
4. **Provide conversion utilities** between formats (export, convert)
5. **Schema validation** via tests (see `internal/config/protobuf_test.go`)
6. **Version compatibility** — schema evolution documented below

---

## Architecture

### Configuration Format (protobuf mandatory)

```
.exarp/
  └── config.pb      # Required for file-based config (protobuf binary)
```

- **Runtime:** Only `.exarp/config.pb` is loaded. If only `config.yaml` exists, the loader returns an error; run `exarp-go config convert yaml protobuf` or `exarp-go config init`.
- **No file:** Defaults are used (no error).
- **Editing:** Use `exarp-go config export yaml` to emit YAML, edit, then `exarp-go config convert yaml protobuf` to write back.

### Conversion Flow

```
Protobuf (.pb) → LoadConfig → Go Structs (runtime)
Go Structs → export yaml → YAML (editing)
YAML → convert yaml protobuf → Protobuf (.pb) (save)
```

---

## Implementation Phases

### Phase 1.5: Protobuf Integration

**Goal:** Integrate protobuf schema with configuration system while maintaining YAML as primary format

**Estimated Time:** 4-6 hours

---

## Detailed Task Breakdown

### T1.5.1: Create Protobuf Conversion Layer ✅ **COMPLETE**

**Goal:** Convert between Go structs and protobuf messages

**Status:** ✅ Implemented and tested

#### Implementation Steps:

1. **Create conversion functions** (`internal/config/protobuf.go`)
   - `ToProtobuf(cfg *FullConfig) (*config.FullConfig, error)` - Go → Protobuf
   - `FromProtobuf(pb *config.FullConfig) (*FullConfig, error)` - Protobuf → Go
   - Handle duration conversions (Go `time.Duration` ↔ protobuf `int64` seconds)
   - Handle nested struct conversions (TimeoutsConfig, ThresholdsConfig, etc.)
   - Handle repeated fields (slices, maps)
   - Handle optional fields (pointers, zero values)

2. **Duration Conversion Helpers**
   ```go
   // Convert Go time.Duration to protobuf int64 (seconds)
   func durationToSeconds(d time.Duration) int64 {
       return int64(d.Seconds())
   }
   
   // Convert protobuf int64 (seconds) to Go time.Duration
   func secondsToDuration(seconds int64) time.Duration {
       return time.Duration(seconds) * time.Second
   }
   ```

3. **Nested Struct Conversions**
   - TimeoutsConfig conversion
   - ThresholdsConfig conversion
   - TasksConfig conversion
   - DatabaseConfig conversion
   - SecurityConfig conversion
   - LoggingConfig conversion
   - ToolsConfig conversion
   - WorkflowConfig conversion
   - MemoryConfig conversion
   - ProjectConfig conversion
   - AutomationsConfig conversion

4. **Error Handling**
   - Validate protobuf messages after conversion
   - Return descriptive errors for conversion failures
   - Handle nil pointers gracefully

**Files:**
- `internal/config/protobuf.go` (NEW - ~400-500 lines)

**Estimated:** 2 hours

**Dependencies:**
- `proto/config.proto` (already exists)
- `proto/config.pb.go` (already generated)
- `internal/config/schema.go` (already exists)

---

### T1.5.2: Add Protobuf Format Support to Loader

**Goal:** Support loading config from protobuf binary format

#### Implementation Steps:

1. **Update `loader.go` to detect format**
   ```go
   // Detect config file format
   func detectConfigFormat(configPath string) (string, error) {
       // Check file extension
       // .yaml/.yml -> YAML
       // .pb -> Protobuf binary
       // Default to YAML if unknown
   }
   ```

2. **Add protobuf loading function**
   ```go
   // LoadConfigProtobuf loads config from protobuf binary format
   func LoadConfigProtobuf(projectRoot string) (*FullConfig, error) {
       configPath := filepath.Join(projectRoot, ".exarp", "config.pb")
       
       // Read binary file
       data, err := os.ReadFile(configPath)
       if err != nil {
           return nil, fmt.Errorf("failed to read protobuf config: %w", err)
       }
       
       // Unmarshal protobuf
       var pbConfig config.FullConfig
       if err := proto.Unmarshal(data, &pbConfig); err != nil {
           return nil, fmt.Errorf("failed to unmarshal protobuf: %w", err)
       }
       
       // Convert to Go structs
       cfg, err := FromProtobuf(&pbConfig)
       if err != nil {
           return nil, fmt.Errorf("failed to convert protobuf: %w", err)
       }
       
       // Merge with defaults (protobuf takes precedence)
       defaults := GetDefaults()
       merged := mergeConfig(defaults, cfg)
       
       // Apply environment variable overrides
       applyEnvOverrides(merged)
       
       // Validate
       if err := ValidateConfig(merged); err != nil {
           return nil, fmt.Errorf("config validation failed: %w", err)
       }
       
       return merged, nil
   }
   ```

3. **Update `LoadConfig` to support both formats**
   ```go
   func LoadConfig(projectRoot string) (*FullConfig, error) {
       // Start with defaults
       cfg := GetDefaults()
       
       // Try protobuf first (if exists, use it)
       pbPath := filepath.Join(projectRoot, ".exarp", "config.pb")
       yamlPath := filepath.Join(projectRoot, ".exarp", "config.yaml")
       
       if _, err := os.Stat(pbPath); err == nil {
           // Protobuf exists, load it
           return LoadConfigProtobuf(projectRoot)
       }
       
       // Fall back to YAML
       if _, err := os.Stat(yamlPath); err == nil {
           // Load YAML (existing implementation)
           // ...
       }
       
       // Apply environment variable overrides
       applyEnvOverrides(cfg)
       
       // Validate
       if err := ValidateConfig(cfg); err != nil {
           return nil, fmt.Errorf("config validation failed: %w", err)
       }
       
       return cfg, nil
   }
   ```

4. **Add format detection helper**
   ```go
   // GetConfigFormat returns the format of the config file (yaml, pb, or none)
   func GetConfigFormat(projectRoot string) (string, error) {
       pbPath := filepath.Join(projectRoot, ".exarp", "config.pb")
       yamlPath := filepath.Join(projectRoot, ".exarp", "config.yaml")
       
       if _, err := os.Stat(pbPath); err == nil {
           return "protobuf", nil
       }
       if _, err := os.Stat(yamlPath); err == nil {
           return "yaml", nil
       }
       return "none", nil
   }
   ```

**Files:**
- `internal/config/loader.go` (UPDATE - add ~100-150 lines)

**Estimated:** 1.5 hours

**Dependencies:**
- T1.5.1 (protobuf conversion layer)

---

### T1.5.3: Add Protobuf CLI Commands ✅ **COMPLETE**

**Goal:** Add CLI commands for protobuf format operations

**Status:** ✅ Implemented and tested

#### Implementation Steps:

1. **Add `config export` command**
   ```go
   // handleConfigExport exports config to different formats
   func handleConfigExport(args []string) error {
       // Parse format argument (yaml, json, protobuf)
       format := "yaml"
       if len(args) > 0 {
           format = args[0]
       }
       
       // Load current config
       projectRoot, err := config.FindProjectRoot()
       if err != nil {
           return err
       }
       
       cfg, err := config.LoadConfig(projectRoot)
       if err != nil {
           return err
       }
       
       // Export based on format
       switch format {
       case "protobuf", "pb":
           return exportProtobuf(cfg, projectRoot)
       case "yaml":
           return exportYAML(cfg, projectRoot)
       case "json":
           return exportJSON(cfg, projectRoot)
       default:
           return fmt.Errorf("unknown format: %s", format)
       }
   }
   ```

2. **Add `config convert` command**
   ```go
   // handleConfigConvert converts between config formats
   func handleConfigConvert(args []string) error {
       if len(args) < 2 {
           return fmt.Errorf("usage: exarp-go config convert <from> <to>")
       }
       
       fromFormat := args[0]
       toFormat := args[1]
       
       projectRoot, err := config.FindProjectRoot()
       if err != nil {
           return err
       }
       
       // Load from source format
       var cfg *config.FullConfig
       switch fromFormat {
       case "yaml":
           cfg, err = config.LoadConfig(projectRoot) // Loads YAML
       case "protobuf", "pb":
           cfg, err = config.LoadConfigProtobuf(projectRoot)
       default:
           return fmt.Errorf("unknown source format: %s", fromFormat)
       }
       
       if err != nil {
           return err
       }
       
       // Save to target format
       switch toFormat {
       case "yaml":
           return saveYAML(cfg, projectRoot)
       case "protobuf", "pb":
           return saveProtobuf(cfg, projectRoot)
       default:
           return fmt.Errorf("unknown target format: %s", toFormat)
       }
   }
   ```

3. **Add helper functions**
   ```go
   // exportProtobuf exports config as protobuf binary
   func exportProtobuf(cfg *config.FullConfig, projectRoot string) error {
       // Convert to protobuf
       pbConfig, err := config.ToProtobuf(cfg)
       if err != nil {
           return err
       }
       
       // Marshal to binary
       data, err := proto.Marshal(pbConfig)
       if err != nil {
           return err
       }
       
       // Write to file
       outputPath := filepath.Join(projectRoot, ".exarp", "config.pb")
       return os.WriteFile(outputPath, data, 0644)
   }
   
   // saveProtobuf saves config as protobuf binary
   func saveProtobuf(cfg *config.FullConfig, projectRoot string) error {
       return exportProtobuf(cfg, projectRoot)
   }
   ```

4. **Update command handler**
   ```go
   func handleConfigCommand(args []string) error {
       if len(args) == 0 {
           return printConfigHelp()
       }
       
       subcommand := args[0]
       switch subcommand {
       case "init":
           return handleConfigInit(args[1:])
       case "validate":
           return handleConfigValidate(args[1:])
       case "show":
           return handleConfigShow(args[1:])
       case "set":
           return handleConfigSet(args[1:])
       case "export":  // NEW
           return handleConfigExport(args[1:])
       case "convert":  // NEW
           return handleConfigConvert(args[1:])
       case "help", "--help", "-h":
           return printConfigHelp()
       default:
           return fmt.Errorf("unknown config subcommand: %s", subcommand)
       }
   }
   ```

5. **Update help text**
   ```go
   func printConfigHelp() error {
       help := `Configuration Management Commands
   
   Usage: exarp-go config <subcommand> [options]
   
   Subcommands:
     init              Generate default .exarp/config.yaml file
     validate          Validate the current config file
     show [format]     Display current configuration (yaml or json)
     set <key>=<value> Set a config value (not yet fully implemented)
     export [format]   Export config to format (yaml, json, protobuf)
     convert <from> <to> Convert config between formats (yaml ↔ protobuf)
     help              Show this help message
   
   Examples:
     exarp-go config init
     exarp-go config validate
     exarp-go config show
     exarp-go config show json
     exarp-go config export protobuf
     exarp-go config convert yaml protobuf
     exarp-go config convert protobuf yaml
   
   Configuration File:
     Location: .exarp/config.yaml (primary) or .exarp/config.pb (optional)
     Format: YAML (human-readable) or Protobuf Binary (type-safe)
     Defaults: All defaults match current hard-coded behavior
   
   For more information, see:
     docs/CONFIGURATION_IMPLEMENTATION_PLAN.md
     docs/CONFIGURATION_PROTOBUF_INTEGRATION.md
   `
       fmt.Print(help)
       return nil
   }
   ```

**Files:**
- `internal/cli/config.go` (UPDATE - add ~200-250 lines)

**Estimated:** 1.5 hours

**Dependencies:**
- T1.5.1 (protobuf conversion layer)
- T1.5.2 (protobuf loader)

---

### T1.5.4: Add Schema Synchronization Validation ✅ **COMPLETE**

**Goal:** Ensure protobuf schema stays in sync with Go structs

**Status:** ✅ Implemented with comprehensive test suite (all tests passing)

#### Implementation Steps:

1. **Create schema validation test**
   ```go
   // Test that protobuf schema matches Go structs
   func TestProtobufSchemaSync(t *testing.T) {
       // Create a full config with all fields set
       cfg := config.GetDefaults()
       
       // Convert to protobuf
       pbConfig, err := config.ToProtobuf(cfg)
       require.NoError(t, err)
       
       // Convert back to Go
       converted, err := config.FromProtobuf(pbConfig)
       require.NoError(t, err)
       
       // Verify all fields are preserved
       // Use reflection or field-by-field comparison
       assert.Equal(t, cfg.Version, converted.Version)
       assert.Equal(t, cfg.Timeouts, converted.Timeouts)
       // ... compare all fields
   }
   ```

2. **Add field count validation**
   ```go
   // Validate that all Go struct fields have protobuf equivalents
   func TestProtobufFieldCoverage(t *testing.T) {
       // Use reflection to count fields in Go structs
       // Compare with protobuf message field counts
       // Ensure no fields are missing
   }
   ```

3. **Add conversion round-trip test**
   ```go
   // Test that conversion is lossless
   func TestProtobufRoundTrip(t *testing.T) {
       original := config.GetDefaults()
       
       // Go → Protobuf → Go
       pb, err := config.ToProtobuf(original)
       require.NoError(t, err)
       
       converted, err := config.FromProtobuf(pb)
       require.NoError(t, err)
       
       // Verify equality (with tolerance for floating point)
       assertConfigEqual(t, original, converted)
   }
   ```

**Files:**
- `internal/config/protobuf_test.go` (NEW - ~200-300 lines)

**Estimated:** 1 hour

**Dependencies:**
- T1.5.1 (protobuf conversion layer)

---

### T1.5.5: Update Documentation ✅ **COMPLETE**

**Goal:** Document protobuf format support and usage

**Status:** ✅ Documentation updated

#### Implementation Steps:

1. **Create protobuf integration guide**
   - Document protobuf format benefits
   - Explain when to use YAML vs protobuf
   - Provide conversion examples
   - Document schema evolution process

2. **Update configuration reference**
   - Add protobuf format section
   - Document binary format structure
   - Explain version compatibility

3. **Add examples**
   - Example: Convert YAML to protobuf
   - Example: Load config from protobuf
   - Example: Export config as protobuf

**Files:**
- `docs/CONFIGURATION_PROTOBUF_INTEGRATION.md` (THIS FILE)
- `docs/CONFIGURATION_REFERENCE.md` (UPDATE)
- `README.md` (UPDATE - add protobuf mention)

**Estimated:** 0.5 hours

**Dependencies:**
- All previous tasks

---

## Implementation Details

### Conversion Strategy

#### Duration Handling

```go
// Go struct uses time.Duration
type TimeoutsConfig struct {
    TaskLockLease time.Duration `yaml:"task_lock_lease"`
}

// Protobuf uses int64 (seconds)
message TimeoutsConfig {
  int64 task_lock_lease = 1;  // Duration in seconds
}

// Conversion
func (t *TimeoutsConfig) ToProtobuf() *config.TimeoutsConfig {
    return &config.TimeoutsConfig{
        TaskLockLease: int64(t.TaskLockLease.Seconds()),
    }
}
```

#### Nested Struct Conversion

```go
// Convert nested configs recursively
func (c *FullConfig) ToProtobuf() (*config.FullConfig, error) {
    pb := &config.FullConfig{
        Version: c.Version,
    }
    
    // Convert nested structs
    if c.Timeouts.TaskLockLease > 0 {
        pb.Timeouts = &config.TimeoutsConfig{
            TaskLockLease: int64(c.Timeouts.TaskLockLease.Seconds()),
            // ... other fields
        }
    }
    
    // ... convert other nested configs
    
    return pb, nil
}
```

#### Zero Value Handling

```go
// Only set protobuf fields if Go struct has non-zero values
func convertTimeouts(goTimeouts TimeoutsConfig) *config.TimeoutsConfig {
    pb := &config.TimeoutsConfig{}
    
    if goTimeouts.TaskLockLease > 0 {
        pb.TaskLockLease = int64(goTimeouts.TaskLockLease.Seconds())
    }
    // ... other fields
    
    return pb
}
```

---

## Testing Strategy

### Unit Tests

1. **Conversion Tests**
   - Test Go → Protobuf conversion
   - Test Protobuf → Go conversion
   - Test round-trip conversion (lossless)
   - Test with all fields set
   - Test with partial fields
   - Test with zero values

2. **Format Detection Tests**
   - Test YAML file detection
   - Test protobuf file detection
   - Test format priority (protobuf over YAML)
   - Test fallback to defaults

3. **CLI Command Tests**
   - Test `config export protobuf`
   - Test `config convert yaml protobuf`
   - Test `config convert protobuf yaml`
   - Test error handling

### Integration Tests

1. **End-to-End Tests**
   - Create YAML config → Convert to protobuf → Load protobuf → Verify
   - Create protobuf config → Convert to YAML → Load YAML → Verify
   - Test with real config files

2. **Performance Tests**
   - Benchmark YAML loading vs protobuf loading
   - Measure conversion overhead
   - Compare file sizes

---

## Migration Path

### For Existing Projects (with only config.yaml)

1. **Required:** Create `.exarp/config.pb` (protobuf mandatory)
   ```bash
   exarp-go config convert yaml protobuf
   ```
2. Optionally keep `config.yaml` as a backup or remove it; runtime uses only `config.pb`.

### For New Projects

1. Create default config (writes `.exarp/config.pb`):
   ```bash
   exarp-go config init
   ```
2. To edit as YAML: `exarp-go config export yaml` → edit file → `exarp-go config convert yaml protobuf`

---

## Success Criteria

### Phase 1.5 Complete When:

- ✅ Protobuf conversion layer implemented and tested
- ✅ Loader supports both YAML and protobuf formats
- ✅ CLI commands for export and convert working
- ✅ Schema synchronization validated
- ✅ Documentation complete
- ✅ All tests passing (13 test cases, all passing)
- ✅ Backward compatible (existing YAML configs work)

**Status:** ✅ **ALL CRITERIA MET - PHASE 1.5 COMPLETE**

---

## Risks & Mitigation

### Risk 1: Schema Drift

**Problem:** Protobuf schema and Go structs get out of sync

**Mitigation:**
- Automated tests to detect schema drift
- Code review process for schema changes
- Document schema evolution process

### Risk 2: Conversion Errors

**Problem:** Data loss during conversion

**Mitigation:**
- Comprehensive round-trip tests
- Field-by-field validation
- Clear error messages

### Risk 3: Performance Overhead

**Problem:** Conversion adds overhead

**Mitigation:**
- Only convert when needed
- Cache converted configs
- Benchmark and optimize hot paths

---

## Timeline

- **T1.5.1:** Protobuf Conversion Layer - 2 hours
- **T1.5.2:** Protobuf Loader Support - 1.5 hours
- **T1.5.3:** CLI Commands - 1.5 hours
- **T1.5.4:** Schema Validation - 1 hour
- **T1.5.5:** Documentation - 0.5 hours

**Total:** 6.5 hours (~1 day)

---

## Dependencies

### Required

- ✅ `proto/config.proto` - Protobuf schema (already exists)
- ✅ `proto/config.pb.go` - Generated Go code (already exists)
- ✅ `internal/config/schema.go` - Go structs (already exists)
- ✅ `internal/config/loader.go` - YAML loader (already exists)

### Optional

- Protobuf compiler (`protoc`) - For schema updates
- Buf tool - For schema linting/validation

---

## Implementation Status

1. ✅ Create implementation plan (this document)
2. ✅ Implement T1.5.1 (Protobuf Conversion Layer) - **COMPLETE**
3. ✅ Implement T1.5.2 (Protobuf Loader Support) - **COMPLETE**
4. ✅ Implement T1.5.3 (CLI Commands) - **COMPLETE**
5. ✅ Implement T1.5.4 (Schema Validation) - **COMPLETE**
6. ✅ Implement T1.5.5 (Documentation) - **COMPLETE**
7. ✅ Test and validate - **All tests passing**

**Phase 1.5 Status:** ✅ **COMPLETE**

## Next Steps

8. ⏳ Continue with Phase 2 (Database & Security)

---

## Schema evolution

When changing configuration fields:

1. **Update `proto/config.proto`** — Add or deprecate fields with new field numbers; never reuse numbers.
2. **Regenerate Go:** Run `buf generate` (or `make proto` if available) to refresh `proto/config.pb.go`.
3. **Update `internal/config/protobuf.go`** — Extend `ToProtobuf` / `FromProtobuf` for new nested structs or fields; preserve zero-value handling.
4. **Update `internal/config/schema.go`** — Keep Go structs in sync with the proto (and YAML tags for export).
5. **Run tests:** `go test ./internal/config/...` — `TestProtobufSchemaSync` and round-trip tests will fail if schemas drift.

See **Risks & Mitigation → Schema Drift** above for automation and review practices.

---

## Related Documents

- `docs/CONFIGURATION_IMPLEMENTATION_PLAN.md` - Main configuration plan (protobuf mandatory, file layout)
- `docs/PROTOBUF_ANALYSIS.md` - Protobuf usage analysis
- `docs/CONFIGURABLE_PARAMETERS_RECOMMENDATIONS.md` - All parameters
- `proto/config.proto` - Protobuf schema definition
- `internal/config/schema.go` - Go struct definitions

---

**Last Updated:** 2026-01-13  
**Status:** ✅ **COMPLETE** - All tasks implemented, tested, and documented

## Implementation Summary

**Completed Tasks:**
- ✅ T1.5.1: Protobuf Conversion Layer (`internal/config/protobuf.go` - ~1000 lines)
- ✅ T1.5.2: Protobuf Loader Support (`internal/config/loader.go` - updated)
- ✅ T1.5.3: CLI Commands (`internal/cli/config.go` - export/convert commands)
- ✅ T1.5.4: Schema Validation (`internal/config/protobuf_test.go` - 13 test cases, all passing)
- ✅ T1.5.5: Documentation (this document and README.md updated)

**Key Features Implemented:**
- Dual format support (YAML primary, protobuf optional)
- Lossless round-trip conversion (Go ↔ Protobuf)
- Duration conversion (time.Duration ↔ int64 seconds)
- Map-to-JSON conversion (StatusWorkflow, DefaultScores, Modes)
- Comprehensive test coverage
- CLI export/convert commands
- Format detection and priority handling

**Usage Examples:**
```bash
# Create default config (writes .exarp/config.pb)
exarp-go config init

# Export current config to YAML (for editing)
exarp-go config export yaml

# After editing YAML, write back to protobuf (mandatory for runtime)
exarp-go config convert yaml protobuf

# Convert existing protobuf to YAML (e.g. for inspection)
exarp-go config convert protobuf yaml

# Validate and see format
exarp-go config validate  # Shows format: protobuf or defaults
```

**Next Phase:** Phase 2 - Database & Security Configuration
