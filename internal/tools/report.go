// report.go — MCP "report" tool: overview, briefing, PRD handlers.
package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
	"github.com/spf13/cast"
	"golang.org/x/sync/singleflight"
)

var reportOverviewFlight singleflight.Group

// handleReportOverview handles the overview action for report tool.
func handleReportOverview(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	outputFormat := "text"
	if format := strings.TrimSpace(cast.ToString(params["output_format"])); format != "" {
		outputFormat = format
	}

	outputPath := DefaultReportOutputPath(projectRoot, "PROJECT_OVERVIEW.md", params)
	includePlanning := cast.ToBool(params["include_planning"])

	flightKey := fmt.Sprintf("overview:%s:%v", projectRoot, includePlanning)
	v, err, _ := reportOverviewFlight.Do(flightKey, func() (interface{}, error) {
		return aggregateProjectDataProto(ctx, projectRoot, includePlanning)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate project data: %w", err)
	}
	overviewProto := v.(*proto.ProjectOverviewData)

	// Format output based on requested format (use proto for type-safe formatting)
	var formattedOutput string

	switch outputFormat {
	case "json":
		overviewMap := ProtoToProjectOverviewData(overviewProto)
		AddTokenEstimateToResult(overviewMap)
		compact := cast.ToBool(params["compact"])
		contents, err := FormatResultOptionalCompact(overviewMap, outputPath, compact)
		if err != nil {
			return nil, fmt.Errorf("failed to format JSON: %w", err)
		}

		return contents, nil
	case "markdown":
		formattedOutput = formatOverviewMarkdownProto(overviewProto)
	case "html":
		formattedOutput = formatOverviewHTMLProto(overviewProto)
	default:
		formattedOutput = formatOverviewTextProto(overviewProto)
	}

	// Save to file if requested
	if outputPath != "" {
		if dir := filepath.Dir(outputPath); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create output directory: %w", err)
			}
		}
		if err := os.WriteFile(outputPath, []byte(formattedOutput), 0644); err != nil {
			return nil, fmt.Errorf("failed to write output file: %w", err)
		}

		formattedOutput += fmt.Sprintf("\n\n[Report saved to: %s]", outputPath)
	}

	return []framework.TextContent{
		{Type: "text", Text: formattedOutput},
	}, nil
}

// handleReportBriefing handles the briefing action for report tool.
// Uses proto internally (BuildBriefingDataProto) for type-safe briefing data.
func handleReportBriefing(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	var score float64

	if sc, ok := params["score"].(float64); ok {
		score = sc
	} else if sc, ok := params["score"].(int); ok {
		score = float64(sc)
	} else {
		score = 50.0 // Default score
	}

	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	engine, err := getWisdomEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize wisdom engine: %w", err)
	}

	// Build briefing from proto (type-safe)
	briefingProto := BuildBriefingDataProto(engine, score)
	briefingMap := BriefingDataToMap(briefingProto)
	AddTokenEstimateToResult(briefingMap)
	compact := cast.ToBool(params["compact"])
	return FormatResultOptionalCompact(briefingMap, "", compact)
}

// handleReportPRD handles the prd action for report tool.
func handleReportPRD(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	projectName := cast.ToString(params["project_name"])
	if projectName == "" {
		// Try to get from go.mod or default
		projectName = filepath.Base(projectRoot)
	}

	includeArchitecture := true
	if _, ok := params["include_architecture"]; ok {
		includeArchitecture = cast.ToBool(params["include_architecture"])
	}

	includeMetrics := true
	if _, ok := params["include_metrics"]; ok {
		includeMetrics = cast.ToBool(params["include_metrics"])
	}

	includeTasks := true
	if _, ok := params["include_tasks"]; ok {
		includeTasks = cast.ToBool(params["include_tasks"])
	}

	outputPath := DefaultReportOutputPath(projectRoot, "PRD.md", params)

	// Generate PRD
	prd, err := generatePRD(ctx, projectRoot, projectName, includeArchitecture, includeMetrics, includeTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PRD: %w", err)
	}

	// Save to file if requested
	if outputPath != "" {
		if dir := filepath.Dir(outputPath); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create output directory: %w", err)
			}
		}
		if err := os.WriteFile(outputPath, []byte(prd), 0644); err != nil {
			return nil, fmt.Errorf("failed to write PRD file: %w", err)
		}

		prd += fmt.Sprintf("\n\n[PRD saved to: %s]", outputPath)
	}

	return []framework.TextContent{
		{Type: "text", Text: prd},
	}, nil
}

// handleReportPlan generates a Cursor-style plan file with .plan.md suffix (Purpose, Technical Foundation, Iterative Milestones, Open Questions).
// See https://cursor.com/learn/creating-plans. Default output is .cursor/plans/<project-slug>.plan.md so Cursor discovers it and shows Build.
