package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// defaultResearchTools returns default tool configs for research aggregator
func defaultResearchTools() []map[string]interface{} {
	return []map[string]interface{}{
		{"tool": "task_analysis", "action": "duplicates"},
		{"tool": "task_analysis", "action": "dependencies"},
		{"tool": "analyze_alignment", "action": "todo2"},
	}
}

// handleResearchAggregator runs multiple research tools and combines outputs (T-224).
func handleResearchAggregator(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}
	toolsList, _ := params["tools"].([]interface{})
	var toRun []map[string]interface{}
	if len(toolsList) == 0 {
		toRun = defaultResearchTools()
	} else {
		for _, t := range toolsList {
			switch v := t.(type) {
			case map[string]interface{}:
				if name, _ := v["tool"].(string); name != "" {
					toRun = append(toRun, v)
				}
			case string:
				toRun = append(toRun, map[string]interface{}{"tool": v, "action": defaultActionForTool(v)})
			}
		}
	}
	var sb strings.Builder
	sb.WriteString("# Research Result Aggregator\n\n")
	results := make([]map[string]interface{}, 0, len(toRun))
	for i, runParams := range toRun {
		toolName, _ := runParams["tool"].(string)
		if toolName == "" {
			continue
		}
		// Copy params excluding "tool" for the handler
		handlerParams := make(map[string]interface{})
		for k, v := range runParams {
			if k != "tool" {
				handlerParams[k] = v
			}
		}
		res := runDailyTask(ctx, toolName, handlerParams)
		results = append(results, map[string]interface{}{
			"tool":    toolName,
			"params":  handlerParams,
			"status":  res["status"],
			"summary": res["summary"],
		})
		sb.WriteString(fmt.Sprintf("## %d. %s\n", i+1, toolName))
		sb.WriteString(fmt.Sprintf("- Status: %v\n", res["status"]))
		if s, ok := res["summary"].(map[string]interface{}); ok {
			j, _ := json.Marshal(s)
			sb.WriteString("- Summary: " + string(j) + "\n")
		}
		sb.WriteString("\n")
	}
	combined := map[string]interface{}{
		"tools_run": len(results),
		"results":   results,
		"report":    sb.String(),
	}
	return response.FormatResult(combined, sb.String())
}

func defaultActionForTool(name string) string {
	switch name {
	case "task_analysis":
		return "duplicates"
	case "analyze_alignment":
		return "todo2"
	case "task_discovery":
		return "all"
	default:
		return ""
	}
}
