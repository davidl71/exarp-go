// protobuf_tasks.go — TasksConfig To/From protobuf.
package config

import (
	"fmt"

	configpb "github.com/davidl71/exarp-go/proto"
)

func tasksToProtobuf(t *TasksConfig) *configpb.TasksConfig {
	return ptrToProto(t, func(t *TasksConfig) *configpb.TasksConfig {
		statusWorkflowJSON, _ := mapToJSON(t.StatusWorkflow)
		return &configpb.TasksConfig{
			DefaultStatus:        t.DefaultStatus,
			DefaultPriority:      t.DefaultPriority,
			DefaultTags:          t.DefaultTags,
			StatusWorkflowJson:   statusWorkflowJSON,
			StaleThresholdHours:  int32(t.StaleThresholdHours),
			AutoCleanupEnabled:   t.AutoCleanupEnabled,
			CleanupDryRun:        t.CleanupDryRun,
			IdFormat:             t.IDFormat,
			IdPrefix:             t.IDPrefix,
			MinDescriptionLength: int32(t.MinDescriptionLength),
			RequireDescription:   t.RequireDescription,
			AutoClarify:          t.AutoClarify,
		}
	})
}

func tasksFromProtobuf(pb *configpb.TasksConfig) (TasksConfig, error) {
	if pb == nil {
		return TasksConfig{}, nil
	}
	var statusWorkflow map[string][]string
	if pb.GetStatusWorkflowJson() != "" {
		if err := jsonToMap(pb.GetStatusWorkflowJson(), &statusWorkflow); err != nil {
			return TasksConfig{}, fmt.Errorf("failed to parse status_workflow_json: %w", err)
		}
	}
	return TasksConfig{
		DefaultStatus:        pb.GetDefaultStatus(),
		DefaultPriority:      pb.GetDefaultPriority(),
		DefaultTags:          pb.GetDefaultTags(),
		StatusWorkflow:       statusWorkflow,
		StaleThresholdHours:  int(pb.GetStaleThresholdHours()),
		AutoCleanupEnabled:   pb.GetAutoCleanupEnabled(),
		CleanupDryRun:        pb.GetCleanupDryRun(),
		IDFormat:             pb.GetIdFormat(),
		IDPrefix:             pb.GetIdPrefix(),
		MinDescriptionLength: int(pb.GetMinDescriptionLength()),
		RequireDescription:   pb.GetRequireDescription(),
		AutoClarify:          pb.GetAutoClarify(),
	}, nil
}
