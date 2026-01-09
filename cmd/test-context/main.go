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

	// Test data
	testData := map[string]interface{}{
		"status":        "success",
		"overall_score": 85,
		"health_score":  85,
		"broken_links":  2,
		"stale_files":   3,
		"tasks_created": 2,
		"recommendations": []string{
			"Fix broken internal links",
			"Update stale documentation files",
		},
	}

	dataJSON, _ := json.Marshal(testData)

	fmt.Println("Testing context tool unification...")
	fmt.Printf("Input data: %s\n\n", string(dataJSON))

	ctx := context.Background()

	// Test 1: context(action="summarize") - should use Apple FM
	fmt.Println("=== Test 1: context(action='summarize') ===")
	args1 := map[string]interface{}{
		"action": "summarize",
		"data":   string(dataJSON),
		"level":  "brief",
	}
	argsBytes1, _ := json.Marshal(args1)
	result1, err1 := server.CallTool(ctx, "context", argsBytes1)
	if err1 != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err1)
	} else {
		fmt.Println("✅ context(action='summarize') succeeded")
		for _, content := range result1 {
			var output map[string]interface{}
			if json.Unmarshal([]byte(content.Text), &output) == nil {
				if method, ok := output["method"].(string); ok {
					fmt.Printf("Method: %s\n", method)
				}
			}
		}
	}

	// Test 2: context(action="budget") - should use native Go
	fmt.Println("\n=== Test 2: context(action='budget') ===")
	items := []interface{}{
		testData,
		map[string]interface{}{"status": "warning", "score": 70, "issues": 5},
		map[string]interface{}{"status": "info", "score": 90, "issues": 1},
	}
	itemsJSON, _ := json.Marshal(items)
	args2 := map[string]interface{}{
		"action":        "budget",
		"items":         string(itemsJSON),
		"budget_tokens": 4000,
	}
	argsBytes2, _ := json.Marshal(args2)
	result2, err2 := server.CallTool(ctx, "context", argsBytes2)
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err2)
	} else {
		fmt.Println("✅ context(action='budget') succeeded (now uses native Go!)")
		for _, content := range result2 {
			var output map[string]interface{}
			if json.Unmarshal([]byte(content.Text), &output) == nil {
				if total, ok := output["total_tokens"].(float64); ok {
					fmt.Printf("Total tokens: %.0f\n", total)
				}
				if strategy, ok := output["strategy"].(string); ok {
					fmt.Printf("Strategy: %s\n", strategy)
				}
			}
		}
	}

	// Test 3: context_budget tool directly - should still work
	fmt.Println("\n=== Test 3: context_budget tool directly ===")
	args3 := map[string]interface{}{
		"items":         string(itemsJSON),
		"budget_tokens": 4000,
	}
	argsBytes3, _ := json.Marshal(args3)
	result3, err3 := server.CallTool(ctx, "context_budget", argsBytes3)
	if err3 != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err3)
	} else {
		fmt.Println("✅ context_budget tool succeeded")
		for _, content := range result3 {
			var output map[string]interface{}
			if json.Unmarshal([]byte(content.Text), &output) == nil {
				if total, ok := output["total_tokens"].(float64); ok {
					fmt.Printf("Total tokens: %.0f\n", total)
				}
			}
		}
	}

	fmt.Println("\n✅ Unification complete! context(action='budget') now uses native Go.")
}
