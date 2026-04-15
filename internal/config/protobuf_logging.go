// protobuf_logging.go — LoggingConfig and LogRotationConfig To/From protobuf.
package config

import (
	configpb "github.com/davidl71/exarp-go/proto"
)

func loggingToProtobuf(l *LoggingConfig) *configpb.LoggingConfig {
	return ptrToProto(l, func(l *LoggingConfig) *configpb.LoggingConfig {
		return &configpb.LoggingConfig{
			Level:             l.Level,
			ToolLevel:         l.ToolLevel,
			FrameworkLevel:   l.FrameworkLevel,
			Format:            l.Format,
			IncludeTimestamps: l.IncludeTimestamps,
			IncludeCaller:     l.IncludeCaller,
			ColorOutput:       l.ColorOutput,
			LogDir:            l.LogDir,
			LogFile:           l.LogFile,
			SessionLogDir:     l.SessionLogDir,
			LogRotation:       logRotationToProtobuf(&l.LogRotation),
			RetentionDays:     int32(l.RetentionDays),
			AutoCleanup:       l.AutoCleanup,
		}
	})
}

func loggingFromProtobuf(pb *configpb.LoggingConfig) LoggingConfig {
	if pb == nil {
		return LoggingConfig{}
	}
	return LoggingConfig{
		Level:             pb.GetLevel(),
		ToolLevel:         pb.GetToolLevel(),
		FrameworkLevel:   pb.GetFrameworkLevel(),
		Format:            pb.GetFormat(),
		IncludeTimestamps: pb.GetIncludeTimestamps(),
		IncludeCaller:     pb.GetIncludeCaller(),
		ColorOutput:       pb.GetColorOutput(),
		LogDir:            pb.GetLogDir(),
		LogFile:           pb.GetLogFile(),
		SessionLogDir:     pb.GetSessionLogDir(),
		LogRotation:       logRotationFromProtobuf(pb.GetLogRotation()),
		RetentionDays:     int(pb.GetRetentionDays()),
		AutoCleanup:       pb.GetAutoCleanup(),
	}
}

func logRotationToProtobuf(l *LogRotationConfig) *configpb.LogRotationConfig {
	return ptrToProto(l, func(l *LogRotationConfig) *configpb.LogRotationConfig {
		return &configpb.LogRotationConfig{
			Enabled:  l.Enabled,
			MaxSize:  l.MaxSize,
			MaxFiles: int32(l.MaxFiles),
			Compress: l.Compress,
		}
	})
}

func logRotationFromProtobuf(pb *configpb.LogRotationConfig) LogRotationConfig {
	if pb == nil {
		return LogRotationConfig{}
	}
	return LogRotationConfig{
		Enabled:  pb.GetEnabled(),
		MaxSize:  pb.GetMaxSize(),
		MaxFiles: int(pb.GetMaxFiles()),
		Compress: pb.GetCompress(),
	}
}
