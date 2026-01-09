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
	fastMode := true // Default to fast mode for CLI
	if len(os.Args) > 1 && os.Args[1] == "--full" {
		fastMode = false // Disable fast mode for full checks
	}
	opts := &tools.ScorecardOptions{FastMode: fastMode}
	scorecard, err := tools.GenerateGoScorecard(ctx, projectRoot, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result := tools.FormatGoScorecard(scorecard)
	fmt.Print(result)
}
