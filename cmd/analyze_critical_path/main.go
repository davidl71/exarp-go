package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/davidl71/exarp-go/internal/tools"
)

func main() {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Analyze critical path
	result, err := tools.AnalyzeCriticalPath(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print formatted output
	printCriticalPathAnalysis(result)
}

func printCriticalPathAnalysis(result map[string]interface{}) {
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println("CRITICAL PATH ANALYSIS")
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println()

	if total, ok := result["total_tasks"].(int); ok {
		fmt.Printf("Total Tasks: %d\n", total)
	}

	if hasCP, ok := result["has_critical_path"].(bool); ok && hasCP {
		if length, ok := result["critical_path_length"].(int); ok {
			fmt.Printf("Critical Path Length: %d tasks\n", length)
		}

		if maxLevel, ok := result["max_dependency_level"].(int); ok {
			fmt.Printf("Max Dependency Level: %d\n", maxLevel)
		}

		fmt.Println()
		fmt.Println("Critical Path Chain (Longest Dependency Path):")
		fmt.Println(strings.Repeat("-", 60))

		if path, ok := result["critical_path"].([]string); ok {
			if details, ok := result["critical_path_details"].([]map[string]interface{}); ok {
				for i, detail := range details {
					if i < len(path) {
						taskID := path[i]
						content, _ := detail["content"].(string)
						priority, _ := detail["priority"].(string)
						status, _ := detail["status"].(string)

						fmt.Printf("\n%d. %s", i+1, taskID)

						if content != "" {
							fmt.Printf(": %s", content)
						}

						fmt.Println()

						if priority != "" {
							fmt.Printf("   Priority: %s", priority)
						}

						if status != "" {
							if priority != "" {
								fmt.Printf(" | ")
							} else {
								fmt.Printf("   ")
							}

							fmt.Printf("Status: %s", status)
						}

						fmt.Println()

						if deps, ok := detail["dependencies"].([]interface{}); ok && len(deps) > 0 {
							depStrs := make([]string, 0, len(deps))

							for _, d := range deps {
								if depStr, ok := d.(string); ok {
									depStrs = append(depStrs, depStr)
								}
							}

							if len(depStrs) > 0 {
								fmt.Printf("   Depends on: %s\n", strings.Join(depStrs, ", "))
							}
						}

						if i < len(details)-1 {
							fmt.Println("   ↓")
						}
					}
				}
			} else {
				// Fallback to simple path
				fmt.Printf("  %s\n", strings.Join(path, " → "))
			}
		}

		fmt.Println()
		fmt.Println(strings.Repeat("-", 60))
		fmt.Println("\n💡 The critical path shows the longest dependency chain.")
		fmt.Println("   Tasks on this path determine the minimum project duration.")
	} else {
		fmt.Println()

		if hasCycles, ok := result["has_cycles"].(bool); ok && hasCycles {
			fmt.Println("❌ Cannot compute critical path: graph has cycles")

			if cycleCount, ok := result["cycle_count"].(int); ok {
				fmt.Printf("   Found %d cycle(s)\n", cycleCount)
			}

			if cycles, ok := result["cycles"].([][]string); ok && len(cycles) > 0 {
				fmt.Println("\nCycles found:")

				for i, cycle := range cycles {
					if i < 5 { // Show first 5 cycles
						fmt.Printf("  %d. %s\n", i+1, strings.Join(cycle, " → "))
					}
				}

				if len(cycles) > 5 {
					fmt.Printf("  ... and %d more\n", len(cycles)-5)
				}
			}
		} else {
			fmt.Println("ℹ️  No critical path found")
			fmt.Println("   Graph may be empty or all tasks are independent")
		}
	}

	fmt.Println()
}
