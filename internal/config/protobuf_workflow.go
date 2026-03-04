// protobuf_workflow.go — WorkflowConfig, ModeSuggestionsConfig, FocusConfig To/From protobuf.
package config

import (
	"fmt"

	configpb "github.com/davidl71/exarp-go/proto"
)

func workflowToProtobuf(w *WorkflowConfig) *configpb.WorkflowConfig {
	return ptrToProto(w, func(w *WorkflowConfig) *configpb.WorkflowConfig {
		modesJSON, _ := mapToJSON(w.Modes)
		return &configpb.WorkflowConfig{
			DefaultMode:     w.DefaultMode,
			AutoDetectMode:  w.AutoDetectMode,
			ModeSuggestions: modeSuggestionsToProtobuf(&w.ModeSuggestions),
			ModesJson:       modesJSON,
			Focus:           focusToProtobuf(&w.Focus),
		}
	})
}

func workflowFromProtobuf(pb *configpb.WorkflowConfig) (WorkflowConfig, error) {
	if pb == nil {
		return WorkflowConfig{}, nil
	}
	var modes map[string]ModeConfig
	if pb.GetModesJson() != "" {
		if err := jsonToMap(pb.GetModesJson(), &modes); err != nil {
			return WorkflowConfig{}, fmt.Errorf("failed to parse modes_json: %w", err)
		}
	}
	return WorkflowConfig{
		DefaultMode:     pb.GetDefaultMode(),
		AutoDetectMode:  pb.GetAutoDetectMode(),
		ModeSuggestions: modeSuggestionsFromProtobuf(pb.GetModeSuggestions()),
		Modes:           modes,
		Focus:           focusFromProtobuf(pb.GetFocus()),
	}, nil
}

func modeSuggestionsToProtobuf(m *ModeSuggestionsConfig) *configpb.ModeSuggestionsConfig {
	return ptrToProto(m, func(m *ModeSuggestionsConfig) *configpb.ModeSuggestionsConfig {
		return &configpb.ModeSuggestionsConfig{
			Morning:   m.Morning,
			Afternoon: m.Afternoon,
			Evening:   m.Evening,
		}
	})
}

func modeSuggestionsFromProtobuf(pb *configpb.ModeSuggestionsConfig) ModeSuggestionsConfig {
	if pb == nil {
		return ModeSuggestionsConfig{}
	}
	return ModeSuggestionsConfig{
		Morning:   pb.GetMorning(),
		Afternoon: pb.GetAfternoon(),
		Evening:   pb.GetEvening(),
	}
}

func focusToProtobuf(f *FocusConfig) *configpb.FocusConfig {
	return ptrToProto(f, func(f *FocusConfig) *configpb.FocusConfig {
		return &configpb.FocusConfig{
			Enabled:           f.Enabled,
			ReductionTarget:   int32(f.ReductionTarget),
			PreserveCoreTools: f.PreserveCoreTools,
		}
	})
}

func focusFromProtobuf(pb *configpb.FocusConfig) FocusConfig {
	if pb == nil {
		return FocusConfig{}
	}
	return FocusConfig{
		Enabled:           pb.GetEnabled(),
		ReductionTarget:   int(pb.GetReductionTarget()),
		PreserveCoreTools: pb.GetPreserveCoreTools(),
	}
}
