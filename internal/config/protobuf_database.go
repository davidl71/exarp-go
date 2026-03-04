// protobuf_database.go — DatabaseConfig To/From protobuf.
package config

import (
	configpb "github.com/davidl71/exarp-go/proto"
)

func databaseToProtobuf(d *DatabaseConfig) *configpb.DatabaseConfig {
	return ptrToProto(d, func(d *DatabaseConfig) *configpb.DatabaseConfig {
		return &configpb.DatabaseConfig{
			SqlitePath:          d.SQLitePath,
			JsonFallbackPath:    d.JSONFallbackPath,
			BackupPath:          d.BackupPath,
			MaxConnections:      int32(d.MaxConnections),
			ConnectionTimeout:   durationToSeconds(d.ConnectionTimeout),
			QueryTimeout:        durationToSeconds(d.QueryTimeout),
			RetryAttempts:       int32(d.RetryAttempts),
			RetryInitialDelay:   durationToSeconds(d.RetryInitialDelay),
			RetryMaxDelay:       durationToSeconds(d.RetryMaxDelay),
			RetryMultiplier:     d.RetryMultiplier,
			AutoVacuum:          d.AutoVacuum,
			WalMode:             d.WALMode,
			CheckpointInterval:  int32(d.CheckpointInterval),
			BackupRetentionDays: int32(d.BackupRetentionDays),
		}
	})
}

func databaseFromProtobuf(pb *configpb.DatabaseConfig) DatabaseConfig {
	if pb == nil {
		return DatabaseConfig{}
	}
	return DatabaseConfig{
		SQLitePath:          pb.GetSqlitePath(),
		JSONFallbackPath:    pb.GetJsonFallbackPath(),
		BackupPath:          pb.GetBackupPath(),
		MaxConnections:      int(pb.GetMaxConnections()),
		ConnectionTimeout:   secondsToDuration(pb.GetConnectionTimeout()),
		QueryTimeout:        secondsToDuration(pb.GetQueryTimeout()),
		RetryAttempts:       int(pb.GetRetryAttempts()),
		RetryInitialDelay:   secondsToDuration(pb.GetRetryInitialDelay()),
		RetryMaxDelay:       secondsToDuration(pb.GetRetryMaxDelay()),
		RetryMultiplier:     pb.GetRetryMultiplier(),
		AutoVacuum:          pb.GetAutoVacuum(),
		WALMode:             pb.GetWalMode(),
		CheckpointInterval:  int(pb.GetCheckpointInterval()),
		BackupRetentionDays: int(pb.GetBackupRetentionDays()),
	}
}
