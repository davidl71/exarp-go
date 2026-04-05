// Package tools: Report insight provider abstraction for AI-generated report/scorecard insights.
// Uses the FM chain (Apple → Ollama) when available.

package tools

import (
	"context"
)

// ReportInsightProvider generates long-form AI insights for report/scorecard content.
// ReportInsightProvider implements TextGenerator.
type ReportInsightProvider interface {
	TextGenerator
}

// defaultReportInsight is the shared default; set by init.
var defaultReportInsight ReportInsightProvider

func init() {
	defaultReportInsight = &fmReportInsight{}
}

// DefaultReportInsight returns the default report insight provider (FM chain).
func DefaultReportInsight() ReportInsightProvider {
	if defaultReportInsight == nil {
		return &fmReportInsight{}
	}

	return defaultReportInsight
}

// fmReportInsight uses the default FM provider (Apple → Ollama chain) only.
type fmReportInsight struct{}

func (c *fmReportInsight) Supported() bool {
	fm := DefaultFMProvider()
	return fm != nil && fm.Supported()
}

func (c *fmReportInsight) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	fm := DefaultFMProvider()
	if fm == nil || !fm.Supported() {
		return "", ErrFMNotSupported
	}
	return fm.Generate(ctx, prompt, maxTokens, temperature)
}
