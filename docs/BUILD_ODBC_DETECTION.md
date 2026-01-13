# ODBC Build Detection

## Overview

The build system now automatically detects whether `sql.h` is available and conditionally compiles ODBC support. This ensures that Apple Foundation Models (AFM) builds are not blocked by missing ODBC dependencies.

## Detection Logic

The Makefile checks for `sql.h` in the following locations (in order):

1. `$(brew --prefix unixodbc)/include/sql.h` - Homebrew unixodbc installation
2. `/opt/homebrew/include/sql.h` - Homebrew default location (Apple Silicon)
3. `/usr/local/include/sql.h` - Homebrew default location (Intel Mac)

## Build Behavior

### When `sql.h` is Found

- **Build with ODBC support**: Uses ODBC driver with full CGO support
- **Includes ODBC headers**: Sets `CGO_CFLAGS` and `CGO_LDFLAGS` appropriately
- **No build tags**: Compiles both ODBC and AFM support

### When `sql.h` is NOT Found

- **Build with AFM only**: Uses `-tags "no_odbc"` to exclude ODBC driver
- **AFM still works**: Apple Foundation Models support is not blocked
- **Clear messaging**: Build output indicates ODBC is disabled

## Build Targets Updated

The following build targets now include `sql.h` detection:

1. **`make build`** - Main build target
   - Checks for `sql.h` before building
   - Falls back to `no_odbc` tag if not found
   - Retries with `no_odbc` if build fails due to `sql.h` error

2. **`make build-debug`** - Debug build
   - Checks for `sql.h` before building
   - Uses `no_odbc` tag if not found

3. **`make build-race`** - Race detector build
   - Checks for `sql.h` before building
   - Uses `no_odbc` tag if not found

4. **`make build-apple-fm`** - Apple Foundation Models build
   - Already uses `no_odbc` tag by default
   - Explicitly excludes ODBC to avoid conflicts

## Error Handling

### Build Failure Recovery

If a build fails due to `sql.h` errors, the Makefile:

1. **Detects the error**: Checks build output for "sql.h" error messages
2. **Retries with no_odbc**: Automatically retries with `-tags "no_odbc"`
3. **Falls back to no-CGO**: If retry fails, falls back to `CGO_ENABLED=0` build
4. **Clear messaging**: Explains what happened and what's available

### Example Output

**When sql.h is found:**
```
Found sql.h at /opt/homebrew/include - Building with ODBC support
✅ Server built with CGO: bin/exarp-go (v...)
   ✅ ODBC support enabled
```

**When sql.h is NOT found:**
```
sql.h not found - Building with AFM support only (ODBC excluded)
   To enable ODBC: brew install unixodbc
✅ Server built with CGO: bin/exarp-go (v...)
   ⚠️  ODBC support disabled (sql.h not found)
```

**When build fails and retries:**
```
⚠️  Build failed due to missing sql.h - Retrying with no_odbc tag...
✅ Server built with CGO (AFM only): bin/exarp-go (v...)
   ⚠️  ODBC support disabled (sql.h not found)
```

## Installing ODBC Support

To enable ODBC support:

```bash
# macOS (Homebrew)
brew install unixodbc

# After installation, rebuild
make build
```

## Build Tags

The build system uses the following Go build tags:

- **`no_odbc`**: Excludes ODBC driver compilation
  - Used when `sql.h` is not found
  - Used by `build-apple-fm` target
  - Prevents ODBC-related compilation errors

- **Default (no tags)**: Includes ODBC driver
  - Used when `sql.h` is found
  - Full ODBC support available

## Code Organization

### ODBC Driver Files

- **`internal/database/driver_odbc.go`**: 
  - Build tag: `//go:build cgo && !no_odbc`
  - Compiled when CGO enabled AND `no_odbc` tag not set

- **`internal/database/driver_odbc_nocgo.go`**:
  - Build tag: `//go:build !cgo || no_odbc`
  - Compiled when CGO disabled OR `no_odbc` tag set
  - Provides stub implementation with clear error messages

## Testing

To test the build system:

1. **With ODBC installed:**
   ```bash
   make build
   # Should show: "✅ ODBC support enabled"
   ```

2. **Without ODBC (simulate):**
   ```bash
   # Temporarily rename sql.h
   sudo mv /opt/homebrew/include/sql.h /opt/homebrew/include/sql.h.bak
   make build
   # Should show: "⚠️  ODBC support disabled (sql.h not found)"
   # Restore: sudo mv /opt/homebrew/include/sql.h.bak /opt/homebrew/include/sql.h
   ```

3. **AFM build (always excludes ODBC):**
   ```bash
   make build-apple-fm
   # Should build successfully even without ODBC
   ```

## Benefits

1. **AFM Not Blocked**: Apple Foundation Models builds work without ODBC
2. **Automatic Detection**: No manual configuration needed
3. **Graceful Fallback**: Builds succeed even if ODBC unavailable
4. **Clear Messaging**: Users know what's available and what's not
5. **Optimal Builds**: Uses ODBC when available, excludes when not

## Related Documentation

- `docs/DATABASE_ABSTRACTION.md` - Database driver architecture
- `docs/INSTALL_AFM.md` - Apple Foundation Models installation
- `internal/database/driver_odbc.go` - ODBC driver implementation
