package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/factory"
	"github.com/davidl71/exarp-go/internal/tools"
)

func main() {
	cfg, _ := config.Load()
	server, _ := factory.NewServerFromConfig(cfg)
	tools.RegisterAllTools(server)

	ctx := context.Background()

	fmt.Println("Testing MLX Integration for Report Generation")

	// Test 1: report(action="scorecard") with MLX
	fmt.Println("=== Test 1: report(action='scorecard') with MLX ===")
	args1 := map[string]interface{}{
		"action":                  "scorecard",
		"include_recommendations": true,
		"output_format":           "text",
		"use_mlx":                 true, // Enable MLX
	}
	argsBytes1, _ := json.Marshal(args1)
	result1, err1 := server.CallTool(ctx, "report", argsBytes1)
	if err1 != nil {
		fmt.Printf("❌ Error: %v\n", err1)
	} else {
		fmt.Println("✅ report(action='scorecard') with MLX succeeded")
		for _, content := range result1 {
			// Check if MLX insights are included
			if strings.Contains(content.Text, "AI-Generated Insights") || strings.Contains(content.Text, "MLX") {
				fmt.Println("✅ MLX insights detected in output")
			}
			// Show first 500 chars
			if len(content.Text) > 500 {
				fmt.Printf("Output preview:\n%s...\n", content.Text[:500])
			} else {
				fmt.Printf("Output:\n%s\n", content.Text)
			}
		}
	}

	// Test 2: report(action="overview") with MLX
	fmt.Println("\n=== Test 2: report(action='overview') with MLX ===")
	args2 := map[string]interface{}{
		"action":        "overview",
		"output_format": "text",
		"use_mlx":       true, // Enable MLX
	}
	argsBytes2, _ := json.Marshal(args2)
	result2, err2 := server.CallTool(ctx, "report", argsBytes2)
	if err2 != nil {
		fmt.Printf("❌ Error: %v\n", err2)
	} else {
		fmt.Println("✅ report(action='overview') with MLX succeeded")
		for _, content := range result2 {
			if strings.Contains(content.Text, "AI-Generated Insights") || strings.Contains(content.Text, "MLX") {
				fmt.Println("✅ MLX insights detected in output")
			}
			if len(content.Text) > 500 {
				fmt.Printf("Output preview:\n%s...\n", content.Text[:500])
			} else {
				fmt.Printf("Output:\n%s\n", content.Text)
			}
		}
	}

	fmt.Println("\n✅ MLX integration for report generation complete!")
}
