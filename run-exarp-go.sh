#!/bin/bash

###############################################################################
# Cursor MCP Server Wrapper
#
# Simple wrapper for exarp-go that builds if needed and runs the server.
# Designed for Cursor IDE STDIO transport.
#
# File watching is enabled by default - automatically rebuilds the server when
# source files change. Watch mode can be disabled via EXARP_WATCH=0 env var.
#
# Usage:
#   ./run-exarp-go.sh                    # Watch mode (auto-rebuild, default)
#   EXARP_WATCH=0 ./run-exarp-go.sh      # Single run (no watching)
###############################################################################

set -euo pipefail

# Configuration; respect PROJECT_ROOT if set by caller (e.g. AI agent)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${PROJECT_ROOT:-$SCRIPT_DIR}"
BINARY_NAME="exarp-go"
BINARY_PATH="$PROJECT_ROOT/bin/$BINARY_NAME"

# Watch mode (disable via EXARP_WATCH=0 environment variable)
# Default: enabled for automatic rebuilds during development
WATCH_MODE="${EXARP_WATCH:-1}"

# Directories to watch
WATCH_DIRS=(
    "cmd"
    "internal"
    "bridge"
)

# Build binary if it doesn't exist or is older than source files
needs_rebuild() {
    if [[ ! -f "$BINARY_PATH" ]]; then
        return 0 # Needs build
    fi

    # Check if any Go source file is newer than binary
    # Try Linux stat format first, fallback to macOS, then default to 0
    local binary_time=0
    if stat -c %Y "$BINARY_PATH" >/dev/null 2>&1; then
        # Linux format
        binary_time=$(stat -c %Y "$BINARY_PATH" 2>/dev/null || echo 0)
    elif stat -f %m "$BINARY_PATH" >/dev/null 2>&1; then
        # macOS format
        binary_time=$(stat -f %m "$BINARY_PATH" 2>/dev/null || echo 0)
    fi

    local newest_source=0
    # Try Linux stat format first
    if command -v stat >/dev/null 2>&1; then
        if stat -c %Y /dev/null >/dev/null 2>&1; then
            # Linux format
            newest_source=$(find "$PROJECT_ROOT/cmd" "$PROJECT_ROOT/internal" "$PROJECT_ROOT/bridge" \
                -name "*.go" -type f -exec stat -c %Y {} \; 2>/dev/null |
                sort -n | tail -1 || echo 0)
        elif stat -f %m /dev/null >/dev/null 2>&1; then
            # macOS format
            newest_source=$(find "$PROJECT_ROOT/cmd" "$PROJECT_ROOT/internal" "$PROJECT_ROOT/bridge" \
                -name "*.go" -type f -exec stat -f %m {} \; 2>/dev/null |
                sort -n | tail -1 || echo 0)
        fi
    fi

    if [[ "${newest_source:-0}" -gt "${binary_time:-0}" ]]; then
        return 0 # Needs build
    fi

    return 1 # Doesn't need build
}

# Build binary
build() {
    echo "[BUILD] Building $BINARY_NAME..." >&2
    mkdir -p "$PROJECT_ROOT/bin"

    if go build -o "$BINARY_PATH" ./cmd/server 2>&1; then
        chmod +x "$BINARY_PATH"
        echo "[BUILD] ✅ Build successful" >&2
        return 0
    else
        echo "[BUILD] ❌ Build failed" >&2
        return 1
    fi
}

# Get server PID (if running)
get_server_pid() {
    # For STDIO mode, we need to track the PID ourselves
    if [[ -f "$PROJECT_ROOT/.exarp-go.pid" ]]; then
        local pid=$(cat "$PROJECT_ROOT/.exarp-go.pid" 2>/dev/null || echo "")
        if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
            echo "$pid"
            return 0
        fi
    fi
    return 1
}

# Start server (background)
start_server() {
    # Stop existing server if running
    stop_server

    # Start new server
    "$BINARY_PATH" "$@" &
    local pid=$!
    echo "$pid" >"$PROJECT_ROOT/.exarp-go.pid"
    echo "[SERVER] Started (PID: $pid)" >&2

    # Wait a moment to check if it crashed
    sleep 0.5
    if ! kill -0 "$pid" 2>/dev/null; then
        echo "[SERVER] ❌ Server crashed immediately" >&2
        rm -f "$PROJECT_ROOT/.exarp-go.pid"
        return 1
    fi
}

# Stop server
stop_server() {
    local pid
    if pid=$(get_server_pid); then
        echo "[SERVER] Stopping (PID: $pid)..." >&2
        kill "$pid" 2>/dev/null || true
        # Wait for graceful shutdown
        sleep 0.5
        # Force kill if still running
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
        fi
        rm -f "$PROJECT_ROOT/.exarp-go.pid"
        echo "[SERVER] Stopped" >&2
    fi
}

# Watch files with fswatch (macOS)
watch_fswatch() {
    local watch_paths=()
    for dir in "${WATCH_DIRS[@]}"; do
        if [[ -d "$PROJECT_ROOT/$dir" ]]; then
            watch_paths+=("$PROJECT_ROOT/$dir")
        fi
    done
    # Also watch go.mod and go.sum
    watch_paths+=("$PROJECT_ROOT/go.mod")
    watch_paths+=("$PROJECT_ROOT/go.sum")

    echo "[WATCH] Starting file watcher (fswatch)..." >&2
    echo "[WATCH] Watching: ${WATCH_DIRS[*]}" >&2

    fswatch -o -r "${watch_paths[@]}" 2>/dev/null | while read -r; do
        echo "[WATCH] File change detected" >&2

        # Rebuild on change (but don't restart - Cursor manages the process)
        if build; then
            echo "[WATCH] ✅ Rebuild complete - restart Cursor to use new binary" >&2
        fi
    done
}

# Watch files with inotifywait (Linux)
watch_inotify() {
    local watch_paths=()
    for dir in "${WATCH_DIRS[@]}"; do
        if [[ -d "$PROJECT_ROOT/$dir" ]]; then
            watch_paths+=("$PROJECT_ROOT/$dir")
        fi
    done

    echo "[WATCH] Starting file watcher (inotifywait)..." >&2
    echo "[WATCH] Watching: ${WATCH_DIRS[*]}" >&2

    inotifywait -m -r -e modify,create,delete \
        --include '\.(go|py)$' \
        "${watch_paths[@]}" 2>/dev/null | while read -r directory event file; do
        # Debounce: ignore if file is being written (common with editors)
        sleep 0.2

        echo "[WATCH] File change detected: $file" >&2

        # Rebuild on change (but don't restart - Cursor manages the process)
        if build; then
            echo "[WATCH] ✅ Rebuild complete - restart Cursor to use new binary" >&2
        fi
    done
}

# Watch files with polling (fallback)
watch_polling() {
    echo "[WATCH] Starting polling-based file watcher..." >&2
    echo "[WATCH] Watching: ${WATCH_DIRS[*]}" >&2

    local last_check=0

    while true; do
        local current_check=0
        # Try Linux stat format first, then macOS
        if stat -c %Y /dev/null >/dev/null 2>&1; then
            # Linux format
            current_check=$(find "$PROJECT_ROOT/cmd" "$PROJECT_ROOT/internal" "$PROJECT_ROOT/bridge" \
                -name "*.go" -o -name "*.py" -type f \
                -exec stat -c %Y {} \; 2>/dev/null | sort -n | tail -1 || echo "0")
        elif stat -f %m /dev/null >/dev/null 2>&1; then
            # macOS format
            current_check=$(find "$PROJECT_ROOT/cmd" "$PROJECT_ROOT/internal" "$PROJECT_ROOT/bridge" \
                -name "*.go" -o -name "*.py" -type f \
                -exec stat -f %m {} \; 2>/dev/null | sort -n | tail -1 || echo "0")
        fi

        if [[ "${current_check:-0}" -gt "${last_check:-0}" ]] && [[ "${last_check:-0}" -gt 0 ]]; then
            echo "[WATCH] File change detected" >&2

            # Rebuild on change (but don't restart - Cursor manages the process)
            if build; then
                echo "[WATCH] ✅ Rebuild complete - restart Cursor to use new binary" >&2
            fi
        fi

        last_check=$current_check
        sleep 2
    done
}

# Watch files (auto-detect method)
watch_files() {
    echo "[WATCH] ⚠️  Watch mode enabled (rebuild only)" >&2
    echo "[WATCH] Note: For Cursor MCP, restart Cursor manually after rebuild" >&2
    echo "[WATCH] Server will NOT restart automatically in watch mode (STDIO constraint)" >&2

    if command -v fswatch &>/dev/null; then
        watch_fswatch "$@"
    elif command -v inotifywait &>/dev/null; then
        watch_inotify "$@"
    else
        echo "[WATCH] ⚠️  fswatch/inotifywait not found, using polling mode" >&2
        echo "[WATCH] Install fswatch (macOS: brew install fswatch) or inotifywait (Linux: apt-get install inotify-tools)" >&2
        watch_polling "$@"
    fi
}

# Cleanup function
cleanup() {
    echo "[CLEANUP] Shutting down..." >&2
    stop_server
    rm -f "$PROJECT_ROOT/.exarp-go.pid"
    exit 0
}

# Main
main() {
    # Change to project root
    cd "$PROJECT_ROOT"

    # Build if needed
    if needs_rebuild; then
        if ! build; then
            exit 1
        fi
    fi

    # Watch mode or single run
    if [[ "$WATCH_MODE" == "1" ]] || [[ "$WATCH_MODE" == "true" ]]; then
        # Watch mode: rebuild on changes, but don't auto-restart server
        # (Cursor MCP manages the server process via STDIO)
        # This mode is useful for development to auto-rebuild
        echo "[INFO] Watch mode: Auto-rebuilding on file changes..." >&2
        echo "[INFO] Note: For Cursor MCP, restart Cursor after rebuild completes" >&2
        echo "[INFO] Server will run in foreground for STDIO" >&2

        # Start file watcher in background
        watch_files "$@" &
        local watcher_pid=$!

        # Cleanup function for watch mode
        watch_cleanup() {
            echo "[CLEANUP] Shutting down watcher..." >&2
            kill "$watcher_pid" 2>/dev/null || true
            wait "$watcher_pid" 2>/dev/null || true
        }
        trap watch_cleanup SIGINT SIGTERM EXIT

        # Run server in foreground (STDIO for Cursor)
        # Don't use exec here so we can clean up the watcher
        "$BINARY_PATH" "$@"
        local exit_code=$?

        # Cleanup watcher
        watch_cleanup
        exit $exit_code
    else
        # Single run (foreground, STDIO) - use exec for clean process replacement
        exec "$BINARY_PATH" "$@"
    fi
}

# Run main
main "$@"
