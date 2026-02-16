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

	// Test estimation tool with Apple FM
	args := map[string]interface{}{
		"action":          "estimate",
		"name":            "CLI Test Task - Complex Migration",
		"details":         "Migrate a complex tool with multiple dependencies, semantic analysis, and hybrid approach. Requires understanding of architecture, implementing native Go code, integrating with existing systems, and comprehensive testing.",
		"priority":        "high",
		"tag_list":        []string{"test", "apple-fm", "migration"},
		"use_historical":  true,
		"use_apple_fm":    true,
		"apple_fm_weight": 0.3,
	}

	argsBytes, _ := json.Marshal(args)

	fmt.Println("=== Testing Estimation Tool with Apple FM Support ===")
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

	// Parse and display key metrics
	var resultJSON map[string]interface{}
	if len(result) > 0 {
		if err := json.Unmarshal([]byte(result[0].Text), &resultJSON); err == nil {
			fmt.Println("\n=== Key Metrics ===")

			if method, ok := resultJSON["method"].(string); ok {
				fmt.Printf("Method: %s\n", method)
			}

			if hours, ok := resultJSON["estimate_hours"].(float64); ok {
				fmt.Printf("Estimate: %.1f hours\n", hours)
			}

			if confidence, ok := resultJSON["confidence"].(float64); ok {
				fmt.Printf("Confidence: %.1f%%\n", confidence*100)
			}

			if metadata, ok := resultJSON["metadata"].(map[string]interface{}); ok {
				if afmEst, ok := metadata["apple_fm_estimate"].(float64); ok {
					fmt.Printf("Apple FM Estimate: %.1f hours\n", afmEst)
				}

				if statEst, ok := metadata["statistical_estimate"].(float64); ok {
					fmt.Printf("Statistical Estimate: %.1f hours\n", statEst)
				}
			}
		}
	}
}
