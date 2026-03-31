// protobuf_timeouts.go — TimeoutsConfig To/From protobuf.
package config

import (
	configpb "github.com/davidl71/exarp-go/proto"
)

func timeoutsToProtobuf(t *TimeoutsConfig) *configpb.TimeoutsConfig {
	return ptrToProto(t, func(t *TimeoutsConfig) *configpb.TimeoutsConfig {
		return &configpb.TimeoutsConfig{
			TaskLockLease:      durationToProto(t.TaskLockLease),
			TaskLockRenewal:    durationToProto(t.TaskLockRenewal),
			StaleLockThreshold: durationToProto(t.StaleLockThreshold),
			ToolDefault:        durationToProto(t.ToolDefault),
			ToolScorecard:      durationToProto(t.ToolScorecard),
			ToolLinting:        durationToProto(t.ToolLinting),
			ToolTesting:        durationToProto(t.ToolTesting),
			ToolReport:         durationToProto(t.ToolReport),
			OllamaDownload:     durationToProto(t.OllamaDownload),
			OllamaGenerate:     durationToProto(t.OllamaGenerate),
			HttpClient:         durationToProto(t.HTTPClient),
			DatabaseRetry:      durationToProto(t.DatabaseRetry),
			ContextSummarize:   durationToProto(t.ContextSummarize),
			ContextBudget:      durationToProto(t.ContextBudget),
		}
	})
}

func timeoutsFromProtobuf(pb *configpb.TimeoutsConfig) TimeoutsConfig {
	if pb == nil {
		return TimeoutsConfig{}
	}
	return TimeoutsConfig{
		TaskLockLease:      durationFromProto(pb.GetTaskLockLease()),
		TaskLockRenewal:    durationFromProto(pb.GetTaskLockRenewal()),
		StaleLockThreshold: durationFromProto(pb.GetStaleLockThreshold()),
		ToolDefault:        durationFromProto(pb.GetToolDefault()),
		ToolScorecard:      durationFromProto(pb.GetToolScorecard()),
		ToolLinting:        durationFromProto(pb.GetToolLinting()),
		ToolTesting:        durationFromProto(pb.GetToolTesting()),
		ToolReport:         durationFromProto(pb.GetToolReport()),
		OllamaDownload:     durationFromProto(pb.GetOllamaDownload()),
		OllamaGenerate:     durationFromProto(pb.GetOllamaGenerate()),
		HTTPClient:         durationFromProto(pb.GetHttpClient()),
		DatabaseRetry:      durationFromProto(pb.GetDatabaseRetry()),
		ContextSummarize:   durationFromProto(pb.GetContextSummarize()),
		ContextBudget:      durationFromProto(pb.GetContextBudget()),
	}
}
