#!/bin/bash

###############################################################################
# Go Development Automation Script
# 
# Hot reload wrapper for exarp-go MCP server. Watches Go files, rebuilds,
# and restarts the server automatically. Designed for STDIO transport.
#
# Usage:
#   ./dev-go.sh [OPTIONS]
#
# Options:
#   --watch              Watch files and auto-reload server
#   --test               Auto-run tests on file changes
#   --build-only         Only rebuild, don't run server
#   --quiet              Suppress non-error output
#   --help, -h           Show this help message
###############################################################################

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"
BINARY_NAME="exarp-go"
BINARY_PATH="bin/$BINARY_NAME"
WATCH_FILES=false
AUTO_TEST=false
BUILD_ONLY=false
QUIET=false

# Files to watch
WATCH_PATTERNS=(
    "*.go"
    "go.mod"
    "go.sum"
)

# Directories to watch
WATCH_DIRS=(
    "cmd"
    "internal"
    "bridge"
)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Logging
log() {
    [[ "$QUIET" == true ]] && return
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%H:%M:%S')
    
    case "$level" in
        INFO)
            echo -e "${GREEN}[$timestamp] [INFO]${NC} $message"
            ;;
        WARN)
            echo -e "${YELLOW}[$timestamp] [WARN]${NC} $message"
            ;;
        ERROR)
            echo -e "${RED}[$timestamp] [ERROR]${NC} $message" >&2
            ;;
        TEST)
            echo -e "${CYAN}[$timestamp] [TEST]${NC} $message"
            ;;
        BUILD)
            echo -e "${BLUE}[$timestamp] [BUILD]${NC} $message"
            ;;
    esac
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --watch)
                WATCH_FILES=true
                shift
                ;;
            --test)
                AUTO_TEST=true
                shift
                ;;
            --build-only)
                BUILD_ONLY=true
                shift
                ;;
            --quiet)
                QUIET=true
                shift
                ;;
            --help|-h)
                cat <<EOF
Usage: $0 [OPTIONS]

Go development automation script for exarp-go MCP server

Options:
  --watch              Watch files and auto-reload server
  --test               Auto-run tests on file changes
  --build-only         Only rebuild, don't run server
  --quiet              Suppress non-error output
  --help, -h           Show this help message

Examples:
  $0 --watch                    # Auto-reload on changes
  $0 --watch --test             # Auto-reload + auto-test
  $0 --build-only                # Just rebuild binary
EOF
                exit 0
                ;;
            *)
                log ERROR "Unknown option: $1"
                exit 1
                ;;
        esac
    done
}

# Check dependencies
check_deps() {
    local missing=()
    
    if ! command -v go &> /dev/null; then
        missing+=("go")
    fi
    
    if [[ "$WATCH_FILES" == true ]] && ! command -v fswatch &> /dev/null && ! command -v inotifywait &> /dev/null; then
        log WARN "fswatch/inotifywait not found, using polling mode"
    fi
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        log ERROR "Missing dependencies: ${missing[*]}"
        log INFO "Install Go from https://go.dev/dl/"
        exit 1
    fi
}

# Build binary
build() {
    log BUILD "Building $BINARY_NAME..."
    
    # Create bin directory if it doesn't exist
    mkdir -p bin
    
    if go build -o "$BINARY_PATH" ./cmd/server 2>&1; then
        log BUILD "✅ Build successful: $BINARY_PATH"
        chmod +x "$BINARY_PATH"
        return 0
    else
        log ERROR "❌ Build failed"
        return 1
    fi
}

# Run tests
run_tests() {
    log TEST "Running Go tests..."
    
    if go test ./... -v 2>&1; then
        log TEST "✅ All tests passed"
        return 0
    else
        log ERROR "❌ Tests failed"
        return 1
    fi
}

# Quick test (faster)
quick_test() {
    log TEST "Running quick test..."
    
    if go test ./... -short 2>&1; then
        log TEST "✅ Quick test passed"
        return 0
    else
        log ERROR "❌ Quick test failed"
        return 1
    fi
}

# Start server
start_server() {
    if [[ ! -f "$BINARY_PATH" ]]; then
        log ERROR "Binary not found: $BINARY_PATH"
        log INFO "Run build first: go build -o $BINARY_PATH ./cmd/server"
        return 1
    fi
    
    log INFO "Starting server: $BINARY_PATH"
    
    # Start server in background, forwarding STDIO
    "$BINARY_PATH" &
    local pid=$!
    echo "$pid" > .server.pid
    log INFO "Server started (PID: $pid)"
    
    # Wait a moment to check if it's still running
    sleep 0.5
    if ! kill -0 "$pid" 2>/dev/null; then
        log ERROR "Server crashed immediately after start"
        rm -f .server.pid
        return 1
    fi
}

# Stop server
stop_server() {
    if [[ -f .server.pid ]]; then
        local pid=$(cat .server.pid 2>/dev/null || echo "")
        if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
            log INFO "Stopping server (PID: $pid)..."
            kill "$pid" 2>/dev/null || true
            # Wait for graceful shutdown
            sleep 0.5
            # Force kill if still running
            if kill -0 "$pid" 2>/dev/null; then
                kill -9 "$pid" 2>/dev/null || true
            fi
            rm -f .server.pid
            log INFO "Server stopped"
        fi
    fi
}

# Watch files with fswatch (macOS)
watch_fswatch() {
    log INFO "Starting file watcher (fswatch)..."
    
    local watch_paths=()
    for dir in "${WATCH_DIRS[@]}"; do
        if [[ -d "$PROJECT_ROOT/$dir" ]]; then
            watch_paths+=("$PROJECT_ROOT/$dir")
        fi
    done
    watch_paths+=("$PROJECT_ROOT/go.mod")
    watch_paths+=("$PROJECT_ROOT/go.sum")
    
    fswatch -o -r "${watch_paths[@]}" | while read -r; do
        log INFO "File change detected"
        
        if build; then
            if [[ "$AUTO_TEST" == true ]]; then
                quick_test
            fi
            
            if [[ "$WATCH_FILES" == true ]] && [[ "$BUILD_ONLY" != true ]]; then
                stop_server
                sleep 0.5
                start_server
            fi
        fi
    done
}

# Watch files with inotifywait (Linux)
watch_inotify() {
    log INFO "Starting file watcher (inotifywait)..."
    
    local watch_paths=()
    for dir in "${WATCH_DIRS[@]}"; do
        if [[ -d "$PROJECT_ROOT/$dir" ]]; then
            watch_paths+=("$PROJECT_ROOT/$dir")
        fi
    done
    watch_paths+=("$PROJECT_ROOT")
    
    inotifywait -m -r -e modify,create,delete \
        --include '\.(go|mod|sum)$' \
        "${watch_paths[@]}" 2>/dev/null | while read -r directory event file; do
        log INFO "File change detected: $file"
        
        if build; then
            if [[ "$AUTO_TEST" == true ]]; then
                quick_test
            fi
            
            if [[ "$WATCH_FILES" == true ]] && [[ "$BUILD_ONLY" != true ]]; then
                stop_server
                sleep 0.5
                start_server
            fi
        fi
    done
}

# Watch files with polling (fallback)
watch_polling() {
    log INFO "Starting polling-based file watcher..."
    
    local last_check=0
    
    while true; do
        local current_check=$(find "$PROJECT_ROOT" \
            \( -name "*.go" -o -name "go.mod" -o -name "go.sum" \) \
            -type f \
            -exec stat -f %m {} \; 2>/dev/null | sort -n | tail -1 || echo "0")
        
        if [[ $current_check -gt $last_check ]] && [[ $last_check -gt 0 ]]; then
            log INFO "File change detected"
            
            if build; then
                if [[ "$AUTO_TEST" == true ]]; then
                    quick_test
                fi
                
                if [[ "$WATCH_FILES" == true ]] && [[ "$BUILD_ONLY" != true ]]; then
                    stop_server
                    sleep 0.5
                    start_server
                fi
            fi
        fi
        
        last_check=$current_check
        sleep 2
    done
}

# Watch files (auto-detect method)
watch_files() {
    if command -v fswatch &> /dev/null; then
        watch_fswatch
    elif command -v inotifywait &> /dev/null; then
        watch_inotify
    else
        watch_polling
    fi
}

# Cleanup
cleanup() {
    log INFO "Shutting down..."
    stop_server
    rm -f .server.pid
    exit 0
}

# Main
main() {
    parse_args "$@"
    check_deps
    
    # Initial build
    if ! build; then
        log ERROR "Initial build failed"
        exit 1
    fi
    
    if [[ "$AUTO_TEST" == true ]]; then
        quick_test
    fi
    
    # Start server if not in build-only mode
    if [[ "$WATCH_FILES" == true ]] && [[ "$BUILD_ONLY" != true ]]; then
        start_server
    fi
    
    # Setup signal handlers
    trap cleanup SIGINT SIGTERM
    
    # Start watching if requested
    if [[ "$WATCH_FILES" == true ]]; then
        log INFO "Watching for changes... (Press Ctrl+C to stop)"
        log INFO "Watching directories: ${WATCH_DIRS[*]}"
        watch_files
    elif [[ "$BUILD_ONLY" == true ]]; then
        log INFO "Build complete. Exiting."
    else
        # Just run once
        if [[ "$AUTO_TEST" == true ]]; then
            quick_test
        fi
    fi
}

# Run main
main "$@"

