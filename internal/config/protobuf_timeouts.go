// protobuf_timeouts.go — TimeoutsConfig To/From protobuf.
package config

import (
	configpb "github.com/davidl71/exarp-go/proto"
)

func timeoutsToProtobuf(t *TimeoutsConfig) *configpb.TimeoutsConfig {
	return ptrToProto(t, func(t *TimeoutsConfig) *configpb.TimeoutsConfig {
		return &configpb.TimeoutsConfig{
			TaskLockLease:      durationToSeconds(t.TaskLockLease),
			TaskLockRenewal:    durationToSeconds(t.TaskLockRenewal),
			StaleLockThreshold: durationToSeconds(t.StaleLockThreshold),
			ToolDefault:        durationToSeconds(t.ToolDefault),
			ToolScorecard:      durationToSeconds(t.ToolScorecard),
			ToolLinting:        durationToSeconds(t.ToolLinting),
			ToolTesting:        durationToSeconds(t.ToolTesting),
			ToolReport:         durationToSeconds(t.ToolReport),
			OllamaDownload:     durationToSeconds(t.OllamaDownload),
			OllamaGenerate:     durationToSeconds(t.OllamaGenerate),
			HttpClient:         durationToSeconds(t.HTTPClient),
			DatabaseRetry:      durationToSeconds(t.DatabaseRetry),
			ContextSummarize:   durationToSeconds(t.ContextSummarize),
			ContextBudget:      durationToSeconds(t.ContextBudget),
		}
	})
}

func timeoutsFromProtobuf(pb *configpb.TimeoutsConfig) TimeoutsConfig {
	if pb == nil {
		return TimeoutsConfig{}
	}
	return TimeoutsConfig{
		TaskLockLease:      secondsToDuration(pb.GetTaskLockLease()),
		TaskLockRenewal:    secondsToDuration(pb.GetTaskLockRenewal()),
		StaleLockThreshold: secondsToDuration(pb.GetStaleLockThreshold()),
		ToolDefault:        secondsToDuration(pb.GetToolDefault()),
		ToolScorecard:      secondsToDuration(pb.GetToolScorecard()),
		ToolLinting:        secondsToDuration(pb.GetToolLinting()),
		ToolTesting:        secondsToDuration(pb.GetToolTesting()),
		ToolReport:         secondsToDuration(pb.GetToolReport()),
		OllamaDownload:     secondsToDuration(pb.GetOllamaDownload()),
		OllamaGenerate:     secondsToDuration(pb.GetOllamaGenerate()),
		HTTPClient:         secondsToDuration(pb.GetHttpClient()),
		DatabaseRetry:      secondsToDuration(pb.GetDatabaseRetry()),
		ContextSummarize:   secondsToDuration(pb.GetContextSummarize()),
		ContextBudget:      secondsToDuration(pb.GetContextBudget()),
	}
}
