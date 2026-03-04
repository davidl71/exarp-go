// protobuf_thresholds.go — ThresholdsConfig To/From protobuf.
package config

import (
	configpb "github.com/davidl71/exarp-go/proto"
)

func thresholdsToProtobuf(t *ThresholdsConfig) *configpb.ThresholdsConfig {
	return ptrToProto(t, func(t *ThresholdsConfig) *configpb.ThresholdsConfig {
		return &configpb.ThresholdsConfig{
			SimilarityThreshold:       t.SimilarityThreshold,
			MinDescriptionLength:      int32(t.MinDescriptionLength),
			MinTaskConfidence:         t.MinTaskConfidence,
			MinCoverage:               int32(t.MinCoverage),
			MinTestConfidence:         t.MinTestConfidence,
			MinEstimationConfidence:   t.MinEstimationConfidence,
			MlxWeight:                 t.MLXWeight,
			MaxParallelTasks:          int32(t.MaxParallelTasks),
			MaxTasksPerHost:           int32(t.MaxTasksPerHost),
			MaxTasksPerWave:           int32(t.MaxTasksPerWave),
			MaxAutomationIterations:   int32(t.MaxAutomationIterations),
			TokensPerChar:             t.TokensPerChar,
			DefaultContextBudget:      int32(t.DefaultContextBudget),
			ContextReductionThreshold: t.ContextReductionThreshold,
			RateLimitRequests:         int32(t.RateLimitRequests),
			RateLimitWindow:           durationToSeconds(t.RateLimitWindow),
			MaxFileSize:               t.MaxFileSize,
			MaxPathDepth:              int32(t.MaxPathDepth),
		}
	})
}

func thresholdsFromProtobuf(pb *configpb.ThresholdsConfig) ThresholdsConfig {
	if pb == nil {
		return ThresholdsConfig{}
	}
	return ThresholdsConfig{
		SimilarityThreshold:       pb.GetSimilarityThreshold(),
		MinDescriptionLength:      int(pb.GetMinDescriptionLength()),
		MinTaskConfidence:         pb.GetMinTaskConfidence(),
		MinCoverage:               int(pb.GetMinCoverage()),
		MinTestConfidence:         pb.GetMinTestConfidence(),
		MinEstimationConfidence:   pb.GetMinEstimationConfidence(),
		MLXWeight:                 pb.GetMlxWeight(),
		MaxParallelTasks:          int(pb.GetMaxParallelTasks()),
		MaxTasksPerHost:           int(pb.GetMaxTasksPerHost()),
		MaxTasksPerWave:           int(pb.GetMaxTasksPerWave()),
		MaxAutomationIterations:   int(pb.GetMaxAutomationIterations()),
		TokensPerChar:             pb.GetTokensPerChar(),
		DefaultContextBudget:      int(pb.GetDefaultContextBudget()),
		ContextReductionThreshold: pb.GetContextReductionThreshold(),
		RateLimitRequests:         int(pb.GetRateLimitRequests()),
		RateLimitWindow:           secondsToDuration(pb.GetRateLimitWindow()),
		MaxFileSize:               pb.GetMaxFileSize(),
		MaxPathDepth:              int(pb.GetMaxPathDepth()),
	}
}
