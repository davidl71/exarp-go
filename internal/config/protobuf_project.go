// protobuf_project.go — ProjectConfig and FeaturesConfig To/From protobuf.
package config

import (
	configpb "github.com/davidl71/exarp-go/proto"
)

func projectToProtobuf(p *ProjectConfig) *configpb.ProjectConfig {
	return ptrToProto(p, func(p *ProjectConfig) *configpb.ProjectConfig {
		return &configpb.ProjectConfig{
			Name:                     p.Name,
			Type:                     p.Type,
			Language:                 p.Language,
			Root:                     p.Root,
			Todo2Path:                p.Todo2Path,
			ExarpPath:                p.ExarpPath,
			Features:                 featuresToProtobuf(&p.Features),
			SkipChecks:               p.SkipChecks,
			CustomTools:              p.CustomTools,
			TaskDiscoveryIgnorePaths: p.TaskDiscoveryIgnorePaths,
		}
	})
}

func projectFromProtobuf(pb *configpb.ProjectConfig) ProjectConfig {
	if pb == nil {
		return ProjectConfig{}
	}
	return ProjectConfig{
		Name:                     pb.GetName(),
		Type:                     pb.GetType(),
		Language:                 pb.GetLanguage(),
		Root:                     pb.GetRoot(),
		Todo2Path:                pb.GetTodo2Path(),
		ExarpPath:                pb.GetExarpPath(),
		Features:                 featuresFromProtobuf(pb.GetFeatures()),
		SkipChecks:               pb.GetSkipChecks(),
		CustomTools:              pb.GetCustomTools(),
		TaskDiscoveryIgnorePaths: pb.GetTaskDiscoveryIgnorePaths(),
	}
}

func featuresToProtobuf(f *FeaturesConfig) *configpb.FeaturesConfig {
	return ptrToProto(f, func(f *FeaturesConfig) *configpb.FeaturesConfig {
		return &configpb.FeaturesConfig{
			SqliteEnabled: f.SQLiteEnabled,
			JsonFallback:  f.JSONFallback,
			McpServers:    f.MCPServers,
		}
	})
}

func featuresFromProtobuf(pb *configpb.FeaturesConfig) FeaturesConfig {
	if pb == nil {
		return FeaturesConfig{}
	}
	return FeaturesConfig{
		SQLiteEnabled: pb.GetSqliteEnabled(),
		JSONFallback:  pb.GetJsonFallback(),
		MCPServers:    pb.GetMcpServers(),
	}
}
