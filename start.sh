#!/bin/bash

###############################################################################
# One startup script to rule them all — exarp-go
#
# Single entry point for run, watch, dev, build, and test. Safe for Cursor MCP
# (STDIO): default is foreground server with build-if-needed; watch only when
# stdin is a TTY and EXARP_WATCH=1.
#
# Usage:
#   ./start.sh              # Run MCP server (build if needed), STDIO-safe
#   ./start.sh server       # Same as default
#   ./start.sh watch        # Watch files, rebuild on change, run server
#   ./start.sh dev          # Watch + run tests on change + server
#   ./start.sh build        # Build only (make b)
#   ./start.sh test         # Run tests (make test)
#   ./start.sh help         # Show this help
#
# Env:
#   PROJECT_ROOT   Override repo root (e.g. for MCP)
#   EXARP_WATCH=1  Enable watch when TTY (default: 1 for TTY, 0 for pipe)
#   EXARP_DEBUG_LOG=1  Log stderr to .cursor/exarp-go-mcp-debug.log
###############################################################################

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Use script dir when PROJECT_ROOT unset or literal Cursor placeholder (so MCP always starts)
if [[ -z "${PROJECT_ROOT}" || "${PROJECT_ROOT}" == "{{PROJECT_ROOT}}" ]]; then
  PROJECT_ROOT="$SCRIPT_DIR"
fi
export GOCACHE="${GOCACHE:-$PROJECT_ROOT/.cache/go-build}"
export GOMODCACHE="${GOMODCACHE:-$PROJECT_ROOT/.cache/go-mod}"
mkdir -p "$GOCACHE" "$GOMODCACHE"
BINARY_NAME="exarp-go"
BINARY_PATH="$PROJECT_ROOT/bin/$BINARY_NAME"

WATCH_DIRS=( cmd internal bridge )

# Optional debug log
if [[ "${EXARP_DEBUG_LOG:-0}" == "1" ]] && [[ -n "$PROJECT_ROOT" ]]; then
    EXARP_DEBUG_FILE="${PROJECT_ROOT}/.cursor/exarp-go-mcp-debug.log"
    exec 2> >(tee -a "$EXARP_DEBUG_FILE" >&2)
fi

# Subcommand (default: server)
SUBCMD="${1:-}"
if [[ "$SUBCMD" == "server" || "$SUBCMD" == "mcp" ]]; then
    SUBCMD=""
    shift
fi

show_help() {
    cat <<EOF
Usage: $0 [COMMAND] [OPTIONS]

Commands:
  (none) or server   Run MCP server (build if needed). STDIO-safe for Cursor.
  watch              Watch files, rebuild on change, run server in foreground.
  dev                Watch + run tests on change + server in foreground.
  build              Build only (make b).
  test               Run tests (make test).
  help               Show this help.

Environment:
  PROJECT_ROOT       Repo root (default: script directory).
  EXARP_WATCH=1      Enable watch when stdin is a TTY (default: 1).
  EXARP_DEBUG_LOG=1  Append stderr to .cursor/exarp-go-mcp-debug.log.

Examples:
  $0                 # Run server (e.g. for Cursor MCP)
  $0 watch           # Dev: auto-rebuild and run
  $0 dev             # Dev: auto-rebuild, test, and run
  $0 build           # Just build
  $0 test            # Just test
EOF
}

# Build using Makefile when available
do_build() {
    echo "[BUILD] Building $BINARY_NAME..." >&2
    mkdir -p "$PROJECT_ROOT/bin"
    (
        cd "$PROJECT_ROOT"
        if [[ -f Makefile ]]; then
            make b
        else
            go build -o "$BINARY_PATH" ./cmd/server
        fi
    )
    chmod +x "$BINARY_PATH"
    echo "[BUILD] ✅ Build successful" >&2
}

needs_rebuild() {
    [[ ! -f "$BINARY_PATH" ]] && return 0
    local binary_time=0 newest_source=0
    if stat -c %Y "$BINARY_PATH" >/dev/null 2>&1; then
        binary_time=$(stat -c %Y "$BINARY_PATH" 2>/dev/null || echo 0)
    elif stat -f %m "$BINARY_PATH" >/dev/null 2>&1; then
        binary_time=$(stat -f %m "$BINARY_PATH" 2>/dev/null || echo 0)
    fi
    if stat -c %Y /dev/null >/dev/null 2>&1; then
        newest_source=$(find "$PROJECT_ROOT/cmd" "$PROJECT_ROOT/internal" "$PROJECT_ROOT/bridge" \
            -name "*.go" -type f -exec stat -c %Y {} \; 2>/dev/null | sort -n | tail -1 || echo 0)
    elif stat -f %m /dev/null >/dev/null 2>&1; then
        newest_source=$(find "$PROJECT_ROOT/cmd" "$PROJECT_ROOT/internal" "$PROJECT_ROOT/bridge" \
            -name "*.go" -type f -exec stat -f %m {} \; 2>/dev/null | sort -n | tail -1 || echo 0)
    fi
    newest_source=${newest_source:-0}
    [[ "$newest_source" =~ ^[0-9]+$ ]] || newest_source=0
    [[ "$binary_time" =~ ^[0-9]+$ ]] || binary_time=0
    [[ "${newest_source}" -gt "${binary_time}" ]]
}

# Watch: rebuild on change (no auto-restart; user restarts Cursor)
watch_loop() {
    local watcher_pid
    if command -v fswatch &>/dev/null; then
        (
            cd "$PROJECT_ROOT"
            fswatch -o -r cmd internal bridge go.mod go.sum 2>/dev/null | while read -r; do
                echo "[WATCH] File change detected" >&2
                do_build && echo "[WATCH] ✅ Rebuild complete - restart Cursor to use new binary" >&2
            done
        ) &
        watcher_pid=$!
    elif command -v inotifywait &>/dev/null; then
        (
            cd "$PROJECT_ROOT"
            inotifywait -m -r -e modify,create,delete --include '\.(go|mod|sum)$' cmd internal bridge go.mod go.sum 2>/dev/null | while read -r _ _ _; do
                sleep 0.2
                echo "[WATCH] File change detected" >&2
                do_build && echo "[WATCH] ✅ Rebuild complete - restart Cursor to use new binary" >&2
            done
        ) &
        watcher_pid=$!
    else
        echo "[WATCH] No fswatch/inotifywait; install for watch (e.g. brew install fswatch)" >&2
        return 0
    fi
    trap "kill $watcher_pid 2>/dev/null || true; exit 0" SIGINT SIGTERM
}

# Run server in foreground (STDIO for MCP)
run_server() {
    cd "$PROJECT_ROOT"
    if [[ ! -f "$BINARY_PATH" ]] || needs_rebuild; then
        do_build
    fi
    exec "$BINARY_PATH" "$@"
}

# --- Commands ---
case "$SUBCMD" in
    help|-h|--help)
        show_help
        exit 0
        ;;
    build)
        cd "$PROJECT_ROOT"
        do_build
        exit 0
        ;;
    test)
        cd "$PROJECT_ROOT"
        if [[ -f Makefile ]]; then
            make test
        else
            go test ./...
        fi
        exit 0
        ;;
    watch)
        shift
        cd "$PROJECT_ROOT"
        if [[ ! -f "$BINARY_PATH" ]] || needs_rebuild; then
            do_build
        fi
        echo "[INFO] Watch mode: auto-rebuilding on file changes. Restart Cursor after rebuild." >&2
        watch_loop
        exec "$BINARY_PATH" "$@"
        ;;
    dev)
        shift
        cd "$PROJECT_ROOT"
        if [[ ! -f "$BINARY_PATH" ]] || needs_rebuild; then
            do_build
        fi
        if [[ -f Makefile ]]; then
            make test
        else
            go test ./... -short
        fi
        echo "[INFO] Dev mode: watch + test on change, server in foreground." >&2
        watch_loop
        exec "$BINARY_PATH" "$@"
        ;;
    "")
        # Default: run server. If TTY and EXARP_WATCH=1, run watch in background + server foreground
        cd "$PROJECT_ROOT"
        if [[ -t 0 ]] && [[ "${EXARP_WATCH:-1}" == "1" ]]; then
            if [[ ! -f "$BINARY_PATH" ]] || needs_rebuild; then
                do_build
            fi
            echo "[INFO] Watch enabled (EXARP_WATCH=1). Rebuild on change; restart Cursor after rebuild." >&2
            watch_loop
        fi
        run_server "$@"
        ;;
    *)
        echo "Unknown command: $SUBCMD" >&2
        show_help >&2
        exit 1
        ;;
esac
