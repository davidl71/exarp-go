// Package wisdom provides a public API for the devwisdom-go wisdom engine.
// This package re-exports the internal wisdom engine functionality for use by other modules.
package wisdom

import (
	"github.com/davidl71/devwisdom-go/internal/wisdom"
)

// Engine wraps the internal wisdom engine for public use.
type Engine struct {
	internal *wisdom.Engine
}

// NewEngine creates a new wisdom engine instance.
func NewEngine() *Engine {
	return &Engine{
		internal: wisdom.NewEngine(),
	}
}

// Initialize loads wisdom sources and configuration.
func (e *Engine) Initialize() error {
	return e.internal.Initialize()
}

// GetWisdom retrieves a wisdom quote based on score and source.
func (e *Engine) GetWisdom(score float64, source string) (*Quote, error) {
	internalQuote, err := e.internal.GetWisdom(score, source)
	if err != nil {
		return nil, err
	}
	return &Quote{
		Quote:         internalQuote.Quote,
		Source:        internalQuote.Source,
		Encouragement: internalQuote.Encouragement,
	}, nil
}

// GetAdvisors returns the advisor registry.
func (e *Engine) GetAdvisors() *AdvisorRegistry {
	return &AdvisorRegistry{
		internal: e.internal.GetAdvisors(),
	}
}

// ListSources returns all available wisdom source IDs.
func (e *Engine) ListSources() []string {
	return e.internal.ListSources()
}

// Quote represents a wisdom quote.
type Quote struct {
	Quote         string
	Source        string
	Encouragement string
}

// AdvisorRegistry wraps the internal advisor registry.
type AdvisorRegistry struct {
	internal *wisdom.AdvisorRegistry
}

// GetAdvisorForMetric returns advisor info for a metric.
func (r *AdvisorRegistry) GetAdvisorForMetric(metric string) (*AdvisorInfo, error) {
	internalInfo, err := r.internal.GetAdvisorForMetric(metric)
	if err != nil {
		return nil, err
	}
	return &AdvisorInfo{
		Advisor:   internalInfo.Advisor,
		Icon:      internalInfo.Icon,
		Rationale: internalInfo.Rationale,
	}, nil
}

// GetAdvisorForTool returns advisor info for a tool.
func (r *AdvisorRegistry) GetAdvisorForTool(tool string) (*AdvisorInfo, error) {
	internalInfo, err := r.internal.GetAdvisorForTool(tool)
	if err != nil {
		return nil, err
	}
	return &AdvisorInfo{
		Advisor:   internalInfo.Advisor,
		Icon:      internalInfo.Icon,
		Rationale: internalInfo.Rationale,
	}, nil
}

// GetAdvisorForStage returns advisor info for a stage.
func (r *AdvisorRegistry) GetAdvisorForStage(stage string) (*AdvisorInfo, error) {
	internalInfo, err := r.internal.GetAdvisorForStage(stage)
	if err != nil {
		return nil, err
	}
	return &AdvisorInfo{
		Advisor:   internalInfo.Advisor,
		Icon:      internalInfo.Icon,
		Rationale: internalInfo.Rationale,
	}, nil
}

// AdvisorInfo represents advisor metadata.
type AdvisorInfo struct {
	Advisor   string
	Icon      string
	Rationale string
}

// GetConsultationMode returns the consultation mode configuration for a score.
func GetConsultationMode(score float64) ConsultationModeConfig {
	internalConfig := wisdom.GetConsultationMode(score)
	return ConsultationModeConfig{
		Name:        internalConfig.Name,
		MinScore:    internalConfig.MinScore,
		MaxScore:    internalConfig.MaxScore,
		Frequency:   internalConfig.Frequency,
		Description: internalConfig.Description,
		Icon:        internalConfig.Icon,
	}
}

// ConsultationModeConfig represents configuration for a consultation mode.
type ConsultationModeConfig struct {
	Name        string
	MinScore    float64
	MaxScore    float64
	Frequency   string
	Description string
	Icon        string
}
