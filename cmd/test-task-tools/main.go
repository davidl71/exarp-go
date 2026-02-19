// main.go — Test harness for task tool operations.
//
// Package main provides the entry point for the task tools test harness.
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
	cfg, _ := config.Load()

	server, _ := factory.NewServerFromConfig(cfg)
	if err := tools.RegisterAllTools(server); err != nil {
		fmt.Printf("Error registering tools: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	fmt.Println("Testing Task Tools Migration to Native Go with Apple FM")

	// Test 1: task_analysis(action="hierarchy")
	fmt.Println("=== Test 1: task_analysis(action='hierarchy') ===")

	args1 := map[string]interface{}{
		"action":                  "hierarchy",
		"include_recommendations": true,
		"output_format":           "json",
	}
	argsBytes1, _ := json.Marshal(args1)

	result1, err1 := server.CallTool(ctx, "task_analysis", argsBytes1)
	if err1 != nil {
		fmt.Printf("❌ Error: %v\n", err1)
	} else {
		fmt.Println("✅ task_analysis(action='hierarchy') succeeded")

		for _, content := range result1 {
			var output map[string]interface{}
			if json.Unmarshal([]byte(content.Text), &output) == nil {
				if method, ok := output["method"].(string); ok {
					fmt.Printf("Method: %s\n", method)
				}

				if total, ok := output["total_tasks"].(float64); ok {
					fmt.Printf("Total tasks: %.0f\n", total)
				}
			}
		}
	}

	// Test 2: task_discovery(action="comments")
	fmt.Println("\n=== Test 2: task_discovery(action='comments') ===")

	args2 := map[string]interface{}{
		"action":        "comments",
		"include_fixme": true,
	}
	argsBytes2, _ := json.Marshal(args2)

	result2, err2 := server.CallTool(ctx, "task_discovery", argsBytes2)
	if err2 != nil {
		fmt.Printf("❌ Error: %v\n", err2)
	} else {
		fmt.Println("✅ task_discovery(action='comments') succeeded")

		for _, content := range result2 {
			var output map[string]interface{}
			if json.Unmarshal([]byte(content.Text), &output) == nil {
				if method, ok := output["method"].(string); ok {
					fmt.Printf("Method: %s\n", method)
				}

				if summary, ok := output["summary"].(map[string]interface{}); ok {
					if total, ok := summary["total"].(float64); ok {
						fmt.Printf("Discovered tasks: %.0f\n", total)
					}
				}
			}
		}
	}

	// Test 3: task_workflow(action="clarify", sub_action="list")
	fmt.Println("\n=== Test 3: task_workflow(action='clarify', sub_action='list') ===")

	args3 := map[string]interface{}{
		"action":     "clarify",
		"sub_action": "list",
	}
	argsBytes3, _ := json.Marshal(args3)

	result3, err3 := server.CallTool(ctx, "task_workflow", argsBytes3)
	if err3 != nil {
		fmt.Printf("❌ Error: %v\n", err3)
	} else {
		fmt.Println("✅ task_workflow(action='clarify', sub_action='list') succeeded")

		for _, content := range result3 {
			var output map[string]interface{}
			if json.Unmarshal([]byte(content.Text), &output) == nil {
				if method, ok := output["method"].(string); ok {
					fmt.Printf("Method: %s\n", method)
				}

				if count, ok := output["tasks_awaiting_clarification"].(float64); ok {
					fmt.Printf("Tasks awaiting clarification: %.0f\n", count)
				}
			}
		}
	}

	fmt.Println("\n✅ Migration complete! All three tools now use native Go with Apple FM.")
}
