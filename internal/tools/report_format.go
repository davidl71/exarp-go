// report_format.go — report format helpers.
package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/proto"
)

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
