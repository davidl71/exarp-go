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

	scorecard, err := tools.GenerateGoScorecard(ctx, projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result := tools.FormatGoScorecard(scorecard)
	fmt.Print(result)
}

