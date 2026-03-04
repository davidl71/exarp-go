// protobuf_memory.go — MemoryConfig and ConsolidationConfig To/From protobuf.
package config

import (
	configpb "github.com/davidl71/exarp-go/proto"
)

func memoryToProtobuf(m *MemoryConfig) *configpb.MemoryConfig {
	return ptrToProto(m, func(m *MemoryConfig) *configpb.MemoryConfig {
		return &configpb.MemoryConfig{
			Categories:     m.Categories,
			StoragePath:    m.StoragePath,
			SessionLogPath: m.SessionLogPath,
			RetentionDays:  int32(m.RetentionDays),
			AutoCleanup:    m.AutoCleanup,
			MaxMemories:    int32(m.MaxMemories),
			Consolidation:  consolidationToProtobuf(&m.Consolidation),
		}
	})
}

func memoryFromProtobuf(pb *configpb.MemoryConfig) MemoryConfig {
	if pb == nil {
		return MemoryConfig{}
	}
	return MemoryConfig{
		Categories:     pb.GetCategories(),
		StoragePath:    pb.GetStoragePath(),
		SessionLogPath: pb.GetSessionLogPath(),
		RetentionDays:  int(pb.GetRetentionDays()),
		AutoCleanup:    pb.GetAutoCleanup(),
		MaxMemories:    int(pb.GetMaxMemories()),
		Consolidation:  consolidationFromProtobuf(pb.GetConsolidation()),
	}
}

func consolidationToProtobuf(c *ConsolidationConfig) *configpb.ConsolidationConfig {
	return ptrToProto(c, func(c *ConsolidationConfig) *configpb.ConsolidationConfig {
		return &configpb.ConsolidationConfig{
			Enabled:             c.Enabled,
			SimilarityThreshold: c.SimilarityThreshold,
			Frequency:           c.Frequency,
		}
	})
}

func consolidationFromProtobuf(pb *configpb.ConsolidationConfig) ConsolidationConfig {
	if pb == nil {
		return ConsolidationConfig{}
	}
	return ConsolidationConfig{
		Enabled:             pb.GetEnabled(),
		SimilarityThreshold: pb.GetSimilarityThreshold(),
		Frequency:           pb.GetFrequency(),
	}
}
