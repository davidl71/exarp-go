package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/internal/tools"
)

func handleAgentBriefing(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}
	data, err := tools.BuildAgentBriefingData(ctx, projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to build agent briefing: %w", err)
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal agent briefing: %w", err)
	}
	return jsonData, "application/json", nil
}

func handleAgentExecutionPack(ctx context.Context, uri string) ([]byte, string, error) {
	taskID, err := parseURIVariableByIndexWithValidation(uri, 4, "task ID", "stdio://agent/task/{task_id}/execution-pack")
	if err != nil {
		return nil, "", err
	}
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}
	jsonData, err := tools.BuildTaskExecutionPackJSON(ctx, projectRoot, taskID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to build execution pack: %w", err)
	}
	return jsonData, "application/json", nil
}

func handleAgentAlerts(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}
	data, err := tools.BuildAgentAlertsData(ctx, projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to build agent alerts: %w", err)
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal agent alerts: %w", err)
	}
	return jsonData, "application/json", nil
}
