// protobuf_tools.go — ToolsConfig and tool sub-configs To/From protobuf.
package config

import (
	"fmt"

	configpb "github.com/davidl71/exarp-go/proto"
)

func toolsToProtobuf(t *ToolsConfig) *configpb.ToolsConfig {
	return ptrToProto(t, func(t *ToolsConfig) *configpb.ToolsConfig {
		return &configpb.ToolsConfig{
			Scorecard: scorecardToProtobuf(&t.Scorecard),
			Report:    reportToProtobuf(&t.Report),
			Linting:   lintingToProtobuf(&t.Linting),
			Testing:   testingToProtobuf(&t.Testing),
			Mlx:       mlxToProtobuf(&t.MLX),
			Ollama:    ollamaToProtobuf(&t.Ollama),
			Context:   contextToProtobuf(&t.Context),
		}
	})
}

func toolsFromProtobuf(pb *configpb.ToolsConfig) (ToolsConfig, error) {
	if pb == nil {
		return ToolsConfig{}, nil
	}
	scorecard, err := scorecardFromProtobuf(pb.GetScorecard())
	if err != nil {
		return ToolsConfig{}, fmt.Errorf("failed to convert scorecard config: %w", err)
	}
	return ToolsConfig{
		Scorecard: scorecard,
		Report:    reportFromProtobuf(pb.GetReport()),
		Linting:   lintingFromProtobuf(pb.GetLinting()),
		Testing:   testingFromProtobuf(pb.GetTesting()),
		MLX:       mlxFromProtobuf(pb.GetMlx()),
		Ollama:    ollamaFromProtobuf(pb.GetOllama()),
		Context:   contextFromProtobuf(pb.GetContext()),
	}, nil
}

func scorecardToProtobuf(s *ScorecardConfig) *configpb.ScorecardConfig {
	return ptrToProto(s, func(s *ScorecardConfig) *configpb.ScorecardConfig {
		defaultScoresJSON, _ := mapToJSON(s.DefaultScores)
		return &configpb.ScorecardConfig{
			DefaultScoresJson: defaultScoresJSON,
			IncludeWisdom:     s.IncludeWisdom,
			OutputFormat:      s.OutputFormat,
		}
	})
}

func scorecardFromProtobuf(pb *configpb.ScorecardConfig) (ScorecardConfig, error) {
	if pb == nil {
		return ScorecardConfig{}, nil
	}
	var defaultScores map[string]float64
	if pb.GetDefaultScoresJson() != "" {
		if err := jsonToMap(pb.GetDefaultScoresJson(), &defaultScores); err != nil {
			return ScorecardConfig{}, fmt.Errorf("failed to parse default_scores_json: %w", err)
		}
	}
	return ScorecardConfig{
		DefaultScores: defaultScores,
		IncludeWisdom: pb.GetIncludeWisdom(),
		OutputFormat:  pb.GetOutputFormat(),
	}, nil
}

func reportToProtobuf(r *ReportConfig) *configpb.ReportConfig {
	return ptrToProto(r, func(r *ReportConfig) *configpb.ReportConfig {
		return &configpb.ReportConfig{
			DefaultFormat:          r.DefaultFormat,
			DefaultOutputPath:      r.DefaultOutputPath,
			IncludeMetrics:         r.IncludeMetrics,
			IncludeRecommendations: r.IncludeRecommendations,
		}
	})
}

func reportFromProtobuf(pb *configpb.ReportConfig) ReportConfig {
	if pb == nil {
		return ReportConfig{}
	}
	return ReportConfig{
		DefaultFormat:          pb.GetDefaultFormat(),
		DefaultOutputPath:      pb.GetDefaultOutputPath(),
		IncludeMetrics:         pb.GetIncludeMetrics(),
		IncludeRecommendations: pb.GetIncludeRecommendations(),
	}
}

func lintingToProtobuf(l *LintingConfig) *configpb.LintingConfig {
	return ptrToProto(l, func(l *LintingConfig) *configpb.LintingConfig {
		return &configpb.LintingConfig{
			DefaultLinter: l.DefaultLinter,
			AutoFix:       l.AutoFix,
			IncludeHints:  l.IncludeHints,
			Timeout:       durationToProto(l.Timeout),
		}
	})
}

func lintingFromProtobuf(pb *configpb.LintingConfig) LintingConfig {
	if pb == nil {
		return LintingConfig{}
	}
	return LintingConfig{
		DefaultLinter: pb.GetDefaultLinter(),
		AutoFix:       pb.GetAutoFix(),
		IncludeHints:  pb.GetIncludeHints(),
		Timeout:       durationFromProto(pb.GetTimeout()),
	}
}

func testingToProtobuf(t *TestingConfig) *configpb.TestingConfig {
	return ptrToProto(t, func(t *TestingConfig) *configpb.TestingConfig {
		return &configpb.TestingConfig{
			DefaultFramework: t.DefaultFramework,
			MinCoverage:      int32(t.MinCoverage),
			CoverageFormat:   t.CoverageFormat,
			Verbose:          t.Verbose,
		}
	})
}

func testingFromProtobuf(pb *configpb.TestingConfig) TestingConfig {
	if pb == nil {
		return TestingConfig{}
	}
	return TestingConfig{
		DefaultFramework: pb.GetDefaultFramework(),
		MinCoverage:      int(pb.GetMinCoverage()),
		CoverageFormat:   pb.GetCoverageFormat(),
		Verbose:          pb.GetVerbose(),
	}
}

func mlxToProtobuf(m *MLXConfig) *configpb.MLXConfig {
	return ptrToProto(m, func(m *MLXConfig) *configpb.MLXConfig {
		return &configpb.MLXConfig{
			DefaultModel:       m.DefaultModel,
			DefaultMaxTokens:   int32(m.DefaultMaxTokens),
			DefaultTemperature: m.DefaultTemperature,
			Verbose:            m.Verbose,
		}
	})
}

func mlxFromProtobuf(pb *configpb.MLXConfig) MLXConfig {
	if pb == nil {
		return MLXConfig{}
	}
	return MLXConfig{
		DefaultModel:       pb.GetDefaultModel(),
		DefaultMaxTokens:   int(pb.GetDefaultMaxTokens()),
		DefaultTemperature: pb.GetDefaultTemperature(),
		Verbose:            pb.GetVerbose(),
	}
}

func ollamaToProtobuf(o *OllamaConfig) *configpb.OllamaConfig {
	return ptrToProto(o, func(o *OllamaConfig) *configpb.OllamaConfig {
		return &configpb.OllamaConfig{
			DefaultModel:       o.DefaultModel,
			DefaultHost:        o.DefaultHost,
			DefaultContextSize: int32(o.DefaultContextSize),
			DefaultNumThreads:  int32(o.DefaultNumThreads),
			DefaultNumGpu:      int32(o.DefaultNumGPU),
		}
	})
}

func ollamaFromProtobuf(pb *configpb.OllamaConfig) OllamaConfig {
	if pb == nil {
		return OllamaConfig{}
	}
	return OllamaConfig{
		DefaultModel:       pb.GetDefaultModel(),
		DefaultHost:        pb.GetDefaultHost(),
		DefaultContextSize: int(pb.GetDefaultContextSize()),
		DefaultNumThreads:  int(pb.GetDefaultNumThreads()),
		DefaultNumGPU:      int(pb.GetDefaultNumGpu()),
	}
}

func contextToProtobuf(c *ContextConfig) *configpb.ContextConfig {
	return ptrToProto(c, func(c *ContextConfig) *configpb.ContextConfig {
		return &configpb.ContextConfig{
			DefaultBudget:             int32(c.DefaultBudget),
			DefaultSummarizationLevel: c.DefaultSummarizationLevel,
			TokensPerChar:             c.TokensPerChar,
			IncludeRaw:                c.IncludeRaw,
		}
	})
}

func contextFromProtobuf(pb *configpb.ContextConfig) ContextConfig {
	if pb == nil {
		return ContextConfig{}
	}
	return ContextConfig{
		DefaultBudget:             int(pb.GetDefaultBudget()),
		DefaultSummarizationLevel: pb.GetDefaultSummarizationLevel(),
		TokensPerChar:             pb.GetTokensPerChar(),
		IncludeRaw:                pb.GetIncludeRaw(),
	}
}
