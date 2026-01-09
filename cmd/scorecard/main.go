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

	// Use fast mode by default for CLI (skip expensive operations)
	opts := &tools.ScorecardOptions{FastMode: true}
	scorecard, err := tools.GenerateGoScorecard(ctx, projectRoot, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result := tools.FormatGoScorecard(scorecard)
	fmt.Print(result)
}
