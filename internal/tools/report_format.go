// report_format.go — report format helpers.
package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/proto"
)

func formatOverviewText(data map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("======================================================================\n")
	sb.WriteString("  PROJECT OVERVIEW\n")
	sb.WriteString("======================================================================\n\n")

	// Project Info
	if project, ok := data["project"].(map[string]interface{}); ok {
		sb.WriteString("Project Information:\n")
		sb.WriteString(fmt.Sprintf("  Name:        %s\n", project["name"]))
		sb.WriteString(fmt.Sprintf("  Version:     %s\n", project["version"]))
		sb.WriteString(fmt.Sprintf("  Type:        %s\n", project["type"]))
		sb.WriteString(fmt.Sprintf("  Status:      %s\n", project["status"]))
		sb.WriteString("\n")
	}

	// Health Score
	if health, ok := data["health"].(map[string]interface{}); ok {
		sb.WriteString("Health Scorecard:\n")

		if score, ok := health["overall_score"].(float64); ok {
			sb.WriteString(fmt.Sprintf("  Overall Score: %.1f%%\n", score))
		}

		if ready, ok := health["production_ready"].(bool); ok {
			if ready {
				sb.WriteString("  Production Ready: YES ✅\n")
			} else {
				sb.WriteString("  Production Ready: NO ❌\n")
			}
		}

		sb.WriteString("\n")
	}

	// Tasks
	if tasks, ok := data["tasks"].(map[string]interface{}); ok {
		sb.WriteString("Task Status:\n")
		sb.WriteString(fmt.Sprintf("  Total:           %d\n", tasks["total"]))
		sb.WriteString(fmt.Sprintf("  Pending:        %d\n", tasks["pending"]))
		sb.WriteString(fmt.Sprintf("  Completed:      %d\n", tasks["completed"]))

		if rate, ok := tasks["completion_rate"].(float64); ok {
			sb.WriteString(fmt.Sprintf("  Completion:     %.1f%%\n", rate))
		}

		if hours, ok := tasks["remaining_hours"].(float64); ok {
			sb.WriteString(fmt.Sprintf("  Remaining Hours: %.1f\n", hours))
		}

		sb.WriteString("\n")
	}

	// Next Actions
	if actions, ok := data["next_actions"].([]map[string]interface{}); ok && len(actions) > 0 {
		sb.WriteString("Next Actions:\n")

		for i, action := range actions {
			if i >= 5 {
				break
			}

			sb.WriteString(fmt.Sprintf("  %d. %s (Priority: %s)\n", i+1, action["name"], action["priority"]))
		}

		sb.WriteString("\n")
	}

	// Planning (optional)
	if planning, ok := data["planning"].(map[string]interface{}); ok {
		if summary, ok := planning["critical_path_summary"].(string); ok && summary != "" {
			sb.WriteString("Critical Path: " + summary + "\n\n")
		}

		if summary, ok := planning["suggested_backlog_summary"].(string); ok && summary != "" {
			sb.WriteString("Suggested Backlog Order: " + summary + "\n\n")
		}
	}

	return sb.String()
}

// formatOverviewMarkdown formats overview as markdown.
func formatOverviewMarkdown(data map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("# Project Overview\n\n")

	// Project Info
	if project, ok := data["project"].(map[string]interface{}); ok {
		sb.WriteString("## Project Information\n\n")
		sb.WriteString(fmt.Sprintf("- **Name**: %s\n", project["name"]))
		sb.WriteString(fmt.Sprintf("- **Version**: %s\n", project["version"]))
		sb.WriteString(fmt.Sprintf("- **Type**: %s\n", project["type"]))
		sb.WriteString(fmt.Sprintf("- **Status**: %s\n\n", project["status"]))
	}

	// Health Score
	if health, ok := data["health"].(map[string]interface{}); ok {
		sb.WriteString("## Health Scorecard\n\n")

		if score, ok := health["overall_score"].(float64); ok {
			sb.WriteString(fmt.Sprintf("**Overall Score**: %.1f%%\n\n", score))
		}
	}

	// Tasks
	if tasks, ok := data["tasks"].(map[string]interface{}); ok {
		sb.WriteString("## Task Status\n\n")
		sb.WriteString(fmt.Sprintf("- **Total**: %d\n", tasks["total"]))
		sb.WriteString(fmt.Sprintf("- **Pending**: %d\n", tasks["pending"]))
		sb.WriteString(fmt.Sprintf("- **Completed**: %d\n", tasks["completed"]))

		if rate, ok := tasks["completion_rate"].(float64); ok {
			sb.WriteString(fmt.Sprintf("- **Completion Rate**: %.1f%%\n\n", rate))
		}
	}

	// Planning (optional)
	if planning, ok := data["planning"].(map[string]interface{}); ok {
		if summary, ok := planning["critical_path_summary"].(string); ok && summary != "" {
			sb.WriteString("## Planning\n\n")
			sb.WriteString("- **Critical Path**: " + summary + "\n\n")
		}

		if summary, ok := planning["suggested_backlog_summary"].(string); ok && summary != "" {
			sb.WriteString("- **Suggested Backlog Order**: " + summary + "\n\n")
		}
	}

	return sb.String()
}

// formatOverviewHTML formats overview as HTML.
func formatOverviewHTML(data map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	sb.WriteString("<title>Project Overview</title>\n")
	sb.WriteString("<style>body{font-family:Arial,sans-serif;margin:40px;}</style>\n")
	sb.WriteString("</head>\n<body>\n")
	sb.WriteString("<h1>Project Overview</h1>\n")

	// Project Info
	if project, ok := data["project"].(map[string]interface{}); ok {
		sb.WriteString("<h2>Project Information</h2>\n<ul>\n")
		sb.WriteString(fmt.Sprintf("<li><strong>Name</strong>: %s</li>\n", project["name"]))
		sb.WriteString(fmt.Sprintf("<li><strong>Version</strong>: %s</li>\n", project["version"]))
		sb.WriteString(fmt.Sprintf("<li><strong>Type</strong>: %s</li>\n", project["type"]))
		sb.WriteString("</ul>\n")
	}

	// Health Score
	if health, ok := data["health"].(map[string]interface{}); ok {
		sb.WriteString("<h2>Health Scorecard</h2>\n")

		if score, ok := health["overall_score"].(float64); ok {
			sb.WriteString(fmt.Sprintf("<p><strong>Overall Score</strong>: %.1f%%</p>\n", score))
		}
	}

	sb.WriteString("</body>\n</html>\n")

	return sb.String()
}

// formatOverviewTextProto formats overview from proto (type-safe, no map assertions).
func formatOverviewTextProto(pb *proto.ProjectOverviewData) string {
	if pb == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("======================================================================\n")
	sb.WriteString("  PROJECT OVERVIEW\n")
	sb.WriteString("======================================================================\n\n")

	if pb.Project != nil {
		sb.WriteString("Project Information:\n")
		sb.WriteString(fmt.Sprintf("  Name:        %s\n", pb.Project.Name))
		sb.WriteString(fmt.Sprintf("  Version:     %s\n", pb.Project.Version))
		sb.WriteString(fmt.Sprintf("  Type:        %s\n", pb.Project.Type))
		sb.WriteString(fmt.Sprintf("  Status:      %s\n", pb.Project.Status))
		sb.WriteString("\n")
	}

	if pb.Health != nil {
		sb.WriteString("Health Scorecard:\n")
		sb.WriteString(fmt.Sprintf("  Overall Score: %.1f%%\n", pb.Health.OverallScore))

		if pb.Health.ProductionReady {
			sb.WriteString("  Production Ready: YES ✅\n")
		} else {
			sb.WriteString("  Production Ready: NO ❌\n")
		}

		sb.WriteString("\n")
	}

	if pb.Tasks != nil {
		sb.WriteString("Task Status:\n")
		sb.WriteString(fmt.Sprintf("  Total:           %d\n", pb.Tasks.Total))
		sb.WriteString(fmt.Sprintf("  Pending:        %d\n", pb.Tasks.Pending))
		sb.WriteString(fmt.Sprintf("  Completed:      %d\n", pb.Tasks.Completed))
		sb.WriteString(fmt.Sprintf("  Completion:     %.1f%%\n", pb.Tasks.CompletionRate))
		sb.WriteString(fmt.Sprintf("  Remaining Hours: %.1f\n", pb.Tasks.RemainingHours))
		sb.WriteString("\n")
	}

	if len(pb.NextActions) > 0 {
		sb.WriteString("Next Actions:\n")

		for i, action := range pb.NextActions {
			if i >= 5 {
				break
			}

			sb.WriteString(fmt.Sprintf("  %d. %s (Priority: %s)\n", i+1, action.Name, action.Priority))
		}

		sb.WriteString("\n")
	}

	if pb.Planning != nil {
		if pb.Planning.CriticalPathSummary != "" {
			sb.WriteString("Critical Path: " + pb.Planning.CriticalPathSummary + "\n\n")
		}

		if pb.Planning.SuggestedBacklogSummary != "" {
			sb.WriteString("Suggested Backlog Order: " + pb.Planning.SuggestedBacklogSummary + "\n\n")
		}
	}

	return sb.String()
}

// GetOverviewText returns project overview as plain text for TUI/CLI display.
// It aggregates project data (health when Go project, tasks, codebase, etc.) and formats as text.
func GetOverviewText(ctx context.Context, projectRoot string) (string, error) {
	pb, err := aggregateProjectDataProto(ctx, projectRoot, false)
	if err != nil {
		return "", err
	}

	return formatOverviewTextProto(pb), nil
}

// formatOverviewMarkdownProto formats overview as markdown from proto.
func formatOverviewMarkdownProto(pb *proto.ProjectOverviewData) string {
	if pb == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("# Project Overview\n\n")

	if pb.Project != nil {
		sb.WriteString("## Project Information\n\n")
		sb.WriteString(fmt.Sprintf("- **Name**: %s\n", pb.Project.Name))
		sb.WriteString(fmt.Sprintf("- **Version**: %s\n", pb.Project.Version))
		sb.WriteString(fmt.Sprintf("- **Type**: %s\n", pb.Project.Type))
		sb.WriteString(fmt.Sprintf("- **Status**: %s\n\n", pb.Project.Status))
	}

	if pb.Health != nil {
		sb.WriteString("## Health Scorecard\n\n")
		sb.WriteString(fmt.Sprintf("**Overall Score**: %.1f%%\n\n", pb.Health.OverallScore))
	}

	if pb.Tasks != nil {
		sb.WriteString("## Task Status\n\n")
		sb.WriteString(fmt.Sprintf("- **Total**: %d\n", pb.Tasks.Total))
		sb.WriteString(fmt.Sprintf("- **Pending**: %d\n", pb.Tasks.Pending))
		sb.WriteString(fmt.Sprintf("- **Completed**: %d\n", pb.Tasks.Completed))
		sb.WriteString(fmt.Sprintf("- **Completion Rate**: %.1f%%\n\n", pb.Tasks.CompletionRate))
	}

	if pb.Planning != nil {
		if pb.Planning.CriticalPathSummary != "" {
			sb.WriteString("## Planning\n\n")
			sb.WriteString("- **Critical Path**: " + pb.Planning.CriticalPathSummary + "\n\n")
		}

		if pb.Planning.SuggestedBacklogSummary != "" {
			sb.WriteString("- **Suggested Backlog Order**: " + pb.Planning.SuggestedBacklogSummary + "\n\n")
		}
	}

	return sb.String()
}

// formatOverviewHTMLProto formats overview as HTML from proto.
func formatOverviewHTMLProto(pb *proto.ProjectOverviewData) string {
	if pb == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	sb.WriteString("<title>Project Overview</title>\n")
	sb.WriteString("<style>body{font-family:Arial,sans-serif;margin:40px;}</style>\n")
	sb.WriteString("</head>\n<body>\n")
	sb.WriteString("<h1>Project Overview</h1>\n")

	if pb.Project != nil {
		sb.WriteString("<h2>Project Information</h2>\n<ul>\n")
		sb.WriteString(fmt.Sprintf("<li><strong>Name</strong>: %s</li>\n", pb.Project.Name))
		sb.WriteString(fmt.Sprintf("<li><strong>Version</strong>: %s</li>\n", pb.Project.Version))
		sb.WriteString(fmt.Sprintf("<li><strong>Type</strong>: %s</li>\n", pb.Project.Type))
		sb.WriteString("</ul>\n")
	}

	if pb.Health != nil {
		sb.WriteString("<h2>Health Scorecard</h2>\n")
		sb.WriteString(fmt.Sprintf("<p><strong>Overall Score</strong>: %.1f%%</p>\n", pb.Health.OverallScore))
	}

	sb.WriteString("</body>\n</html>\n")

	return sb.String()
}

// getFloatParam safely extracts float64 from params.
func getFloatParam(params map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := params[key].(float64); ok {
		return val
	}

	return defaultValue
}
