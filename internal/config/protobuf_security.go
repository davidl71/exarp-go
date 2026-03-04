// protobuf_security.go — SecurityConfig and sub-configs To/From protobuf.
package config

import (
	configpb "github.com/davidl71/exarp-go/proto"
)

func securityToProtobuf(s *SecurityConfig) *configpb.SecurityConfig {
	return ptrToProto(s, func(s *SecurityConfig) *configpb.SecurityConfig {
		return &configpb.SecurityConfig{
			RateLimit:      rateLimitToProtobuf(&s.RateLimit),
			PathValidation: pathValidationToProtobuf(&s.PathValidation),
			FileLimits:     fileLimitsToProtobuf(&s.FileLimits),
			AccessControl:  accessControlToProtobuf(&s.AccessControl),
		}
	})
}

func securityFromProtobuf(pb *configpb.SecurityConfig) SecurityConfig {
	if pb == nil {
		return SecurityConfig{}
	}
	return SecurityConfig{
		RateLimit:      rateLimitFromProtobuf(pb.GetRateLimit()),
		PathValidation: pathValidationFromProtobuf(pb.GetPathValidation()),
		FileLimits:     fileLimitsFromProtobuf(pb.GetFileLimits()),
		AccessControl:  accessControlFromProtobuf(pb.GetAccessControl()),
	}
}

func rateLimitToProtobuf(r *RateLimitConfig) *configpb.RateLimitConfig {
	return ptrToProto(r, func(r *RateLimitConfig) *configpb.RateLimitConfig {
		return &configpb.RateLimitConfig{
			Enabled:           r.Enabled,
			RequestsPerWindow: int32(r.RequestsPerWindow),
			WindowDuration:    durationToSeconds(r.WindowDuration),
			BurstSize:         int32(r.BurstSize),
		}
	})
}

func rateLimitFromProtobuf(pb *configpb.RateLimitConfig) RateLimitConfig {
	if pb == nil {
		return RateLimitConfig{}
	}
	return RateLimitConfig{
		Enabled:           pb.GetEnabled(),
		RequestsPerWindow: int(pb.GetRequestsPerWindow()),
		WindowDuration:    secondsToDuration(pb.GetWindowDuration()),
		BurstSize:         int(pb.GetBurstSize()),
	}
}

func pathValidationToProtobuf(p *PathValidationConfig) *configpb.PathValidationConfig {
	return ptrToProto(p, func(p *PathValidationConfig) *configpb.PathValidationConfig {
		return &configpb.PathValidationConfig{
			Enabled:            p.Enabled,
			AllowAbsolutePaths: p.AllowAbsolutePaths,
			MaxDepth:           int32(p.MaxDepth),
			BlockedPatterns:    p.BlockedPatterns,
		}
	})
}

func pathValidationFromProtobuf(pb *configpb.PathValidationConfig) PathValidationConfig {
	if pb == nil {
		return PathValidationConfig{}
	}
	return PathValidationConfig{
		Enabled:            pb.GetEnabled(),
		AllowAbsolutePaths: pb.GetAllowAbsolutePaths(),
		MaxDepth:           int(pb.GetMaxDepth()),
		BlockedPatterns:    pb.GetBlockedPatterns(),
	}
}

func fileLimitsToProtobuf(f *FileLimitsConfig) *configpb.FileLimitsConfig {
	return ptrToProto(f, func(f *FileLimitsConfig) *configpb.FileLimitsConfig {
		return &configpb.FileLimitsConfig{
			MaxFileSize:          f.MaxFileSize,
			MaxFilesPerOperation: int32(f.MaxFilesPerOperation),
			AllowedExtensions:    f.AllowedExtensions,
		}
	})
}

func fileLimitsFromProtobuf(pb *configpb.FileLimitsConfig) FileLimitsConfig {
	if pb == nil {
		return FileLimitsConfig{}
	}
	return FileLimitsConfig{
		MaxFileSize:          pb.GetMaxFileSize(),
		MaxFilesPerOperation: int(pb.GetMaxFilesPerOperation()),
		AllowedExtensions:    pb.GetAllowedExtensions(),
	}
}

func accessControlToProtobuf(a *AccessControlConfig) *configpb.AccessControlConfig {
	return ptrToProto(a, func(a *AccessControlConfig) *configpb.AccessControlConfig {
		return &configpb.AccessControlConfig{
			Enabled:         a.Enabled,
			DefaultPolicy:   a.DefaultPolicy,
			RestrictedTools: a.RestrictedTools,
		}
	})
}

func accessControlFromProtobuf(pb *configpb.AccessControlConfig) AccessControlConfig {
	if pb == nil {
		return AccessControlConfig{}
	}
	return AccessControlConfig{
		Enabled:         pb.GetEnabled(),
		DefaultPolicy:   pb.GetDefaultPolicy(),
		RestrictedTools: pb.GetRestrictedTools(),
	}
}
