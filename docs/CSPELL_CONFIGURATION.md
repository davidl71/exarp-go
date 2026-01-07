# CSpell Configuration for Go Projects

**Date:** 2026-01-07  
**Status:** ✅ Configured

---

## Overview

CSpell (Code Spell Checker) is configured for this Go project with:
- ✅ Built-in Go dictionary (enabled by default)
- ✅ Project-specific terms dictionary
- ✅ Appropriate ignore patterns

---

## Configuration Files

### `.cspell.json`
Main configuration file with:
- Go language dictionary enabled
- Project-specific words list
- Ignore patterns for build artifacts and archives

### `.cspell/project-words.txt`
Project-specific terms including:
- Project names (exarp, exarp-go)
- Tool names (gomarklint, markdownlint)
- Technical terms (MCPServer, jsonrpc, stdio)
- Go package names (jsonschema, uritemplate, etc.)

---

## Go Dictionary Support

CSpell includes a **built-in Go dictionary** (`@cspell/dict-golang`) that covers:
- ✅ Go keywords (func, type, const, var, etc.)
- ✅ Built-in library names (fmt, os, io, etc.)
- ✅ Standard library packages (net/http, encoding/json, etc.)
- ✅ Go 1.12 and earlier features

**No additional configuration needed** - the Go dictionary is enabled by default.

---

## Project-Specific Terms

### Project Names
- `exarp` - Project name
- `exarp-go` - Go implementation
- `exarp_pma` - Python implementation

### Tools & Utilities
- `gomarklint` - Native Go markdown linter
- `markdownlint` - Markdown linter
- `golangci-lint` - Go linter
- `goimports` - Go import formatter
- `gofmt` - Go formatter
- `govet` - Go vet tool

### MCP & Protocols
- `MCPServer` - MCP server interface
- `jsonrpc` - JSON-RPC protocol
- `stdio` - Standard I/O
- `MCP` - Model Context Protocol
- `FastMCP` - FastMCP framework

### Go Packages
- `jsonschema` - JSON schema package
- `uritemplate` - URI template package
- `modelcontextprotocol` - MCP Go SDK
- `shinagawa` - gomarklint author
- `bmatcuk` - doublestar author
- `spf13` - cobra/pflag packages

### Technical Terms
- `stdin`, `stdout`, `stderr` - Standard streams
- `subprocess` - Process execution
- `filepath` - File path handling
- `unmarshal`, `marshal` - JSON operations
- `lookpath` - Executable lookup

---

## Ignore Patterns

The following are excluded from spell checking:
- `**/node_modules/**` - Node.js dependencies
- `**/vendor/**` - Go vendor directory
- `**/.git/**` - Git directory
- `**/bin/**` - Build artifacts
- `**/coverage/**` - Test coverage
- `**/archive/**` - Archived documentation
- `**/.cursor/**` - Cursor IDE files
- `**/go.sum` - Go checksum file
- `**/*.lock` - Lock files

---

## Usage

### Install CSpell
```bash
npm install -g cspell
```

### Check Spelling
```bash
# Check all files
cspell "**/*.{go,md}"

# Check specific file
cspell README.md

# Check with config
cspell --config .cspell.json "**/*.go"
```

### Add New Words
1. Add to `.cspell/project-words.txt` (one word per line)
2. Or add to `words` array in `.cspell.json`

### VS Code Integration
If using VS Code, install the "Code Spell Checker" extension. It will automatically use `.cspell.json`.

---

## Go-Specific Considerations

### Built-in Dictionary
The Go dictionary (`golang`) is enabled by default and covers:
- Standard Go keywords
- Standard library packages
- Common Go patterns

### Custom Terms Needed
You may need to add:
- **Project-specific names** (exarp, Todo2)
- **Third-party packages** (modelcontextprotocol, jsonschema)
- **Tool names** (gomarklint, golangci-lint)
- **Domain terms** (MCPServer, jsonrpc)

---

## Maintenance

### Adding New Terms
1. **Project-specific:** Add to `.cspell/project-words.txt`
2. **Temporary:** Add to `words` array in `.cspell.json`
3. **Common Go terms:** Usually covered by built-in dictionary

### Updating Configuration
- Review `.cspell.json` periodically
- Add new ignore patterns as needed
- Update project words as project evolves

---

## References

- **CSpell Docs:** https://cspell.org/
- **Go Dictionary:** https://www.npmjs.com/package/@cspell/dict-golang
- **Custom Dictionaries:** https://cspell.org/docs/dictionaries/custom-dictionaries/

---

**Status:** ✅ Configured and ready to use

