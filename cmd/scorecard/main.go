// main.go — Standalone scorecard runner for project health metrics.
//
// Package main provides the entry point for the scorecard tool.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidl71/exarp-go/internal/tools"
)

func main() {
	ctx := context.Background()

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	projectRoot := filepath.Clean(wd)

	// Check for --full flag to disable fast mode
	fastMode := len(os.Args) <= 1 || os.Args[1] != "--full" // Default to fast mode for CLI

	opts := &tools.ScorecardOptions{FastMode: fastMode}

	scorecard, err := tools.GenerateGoScorecard(ctx, projectRoot, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result := tools.FormatGoScorecard(scorecard)
	fmt.Print(result)
}
