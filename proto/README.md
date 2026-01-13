# Protocol Buffers Schemas

This directory contains Protocol Buffer (protobuf) schema definitions for exarp-go.

## Overview

Protocol Buffers provide:
- **Performance**: 2-10x faster serialization than JSON
- **Size**: 30-50% smaller payloads
- **Type Safety**: Compile-time validation
- **Versioning**: Backward/forward compatibility

## Schema Files

### `todo2.proto`
Defines the Todo2Task message structure for task management.

**Messages:**
- `Todo2Task`: Individual task with all metadata
- `Todo2State`: Complete state file structure

**Usage:**
- Task serialization/deserialization
- Database storage
- Task workflow operations

### `bridge.proto`
Defines messages for Python bridge communication.

**Messages:**
- `ToolRequest`: Request to execute a Python tool
- `ToolResponse`: Response from Python tool execution
- `ResourceRequest`: Request to fetch a resource
- `ResourceResponse`: Response with resource content

**Usage:**
- Python bridge subprocess communication
- Tool execution requests/responses
- Resource handler communication

### `config.proto`
Defines configuration structure (simplified version).

**Messages:**
- `FullConfig`: Complete configuration structure
- Various nested config messages (TimeoutsConfig, ThresholdsConfig, etc.)

**Usage:**
- Configuration serialization
- Config file storage
- Configuration management tools

## Building

### Prerequisites

Install protobuf compiler and Go plugin:

```bash
# macOS
brew install protobuf

# Linux
apt-get install protobuf-compiler

# Install Go plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

### Generate Go Code

```bash
# Using Makefile (recommended)
make proto

# Or manually
protoc --go_out=. --go_opt=paths=source_relative \
  proto/todo2.proto \
  proto/bridge.proto \
  proto/config.proto
```

Generated code will be in `proto/todo2/`, `proto/bridge/`, `proto/config/` directories.

## Field Numbering

Field numbers are reserved for future additions:
- **1-15**: Most frequently used fields (1-byte encoding)
- **16-2047**: Regular fields (2-byte encoding)
- **2048+**: Reserved for future use

**Important**: Never reuse field numbers. If a field is removed, mark it as `reserved` to prevent accidental reuse.

## Versioning Strategy

### Backward Compatibility
- New fields are always added at the end
- Removed fields are marked as `reserved`
- Field numbers are never reused

### Migration
- Old code can read new messages (unknown fields ignored)
- New code can read old messages (missing fields use defaults)
- Use optional fields for new additions when possible

## Best Practices

1. **Use proto3 syntax** (current standard)
2. **Reserve field numbers** for removed fields
3. **Document all messages** with comments
4. **Use appropriate types** (int64 for timestamps, string for IDs)
5. **Keep messages focused** (avoid overly large messages)

## Testing

Validate proto files:

```bash
# Check syntax
protoc --proto_path=proto proto/*.proto

# Generate and verify Go code compiles
make proto
go build ./proto/...
```

## References

- [Protocol Buffers Guide](https://protobuf.dev)
- [Go Protobuf Documentation](https://pkg.go.dev/google.golang.org/protobuf)
- [Buf Documentation](https://buf.build/docs)
