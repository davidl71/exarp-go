#!/bin/bash
# Test estimation tool via CLI with Apple FM support

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${PROJECT_ROOT:-$SCRIPT_DIR}"
export GOCACHE="${GOCACHE:-$PROJECT_ROOT/.cache/go-build}"
export GOMODCACHE="${GOMODCACHE:-$PROJECT_ROOT/.cache/go-mod}"
mkdir -p "$GOCACHE" "$GOMODCACHE"

cd "$SCRIPT_DIR"

# Force TTY by using expect or by running in interactive shell
# For now, let's use a Go test program instead
cat > /tmp/test_estimation.go << 'EOF'
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/factory"
	"github.com/davidl71/exarp-go/internal/tools"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create server
	server, err := factory.NewServerFromConfig(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Register tools
	if err := tools.RegisterAllTools(server); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register tools: %v\n", err)
		os.Exit(1)
	}

	// Test estimation tool
	args := map[string]interface{}{
		"action":         "estimate",
		"name":           "CLI Test Task - Complex Migration",
		"details":        "Migrate a complex tool with multiple dependencies, semantic analysis, and hybrid approach. Requires understanding of architecture, implementing native Go code, integrating with existing systems, and comprehensive testing.",
		"priority":        "high",
		"tag_list":       []string{"test", "apple-fm", "migration"},
		"use_historical": true,
		"use_apple_fm":    true,
		"apple_fm_weight": 0.3,
	}

	argsBytes, _ := json.Marshal(args)
	
	fmt.Println("Testing estimation tool with Apple FM support...")
	fmt.Printf("Arguments: %s\n\n", string(argsBytes))
	
	ctx := context.Background()
	result, err := server.CallTool(ctx, "estimation", argsBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Result:")
	for i, content := range result {
		if len(result) > 1 {
			fmt.Printf("\n[%d] ", i+1)
		}
		fmt.Println(content.Text)
	}
}
EOF

cd "$PROJECT_ROOT"
go run /tmp/test_estimation.go 2>&1
