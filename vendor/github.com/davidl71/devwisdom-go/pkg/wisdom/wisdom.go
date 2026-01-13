// Package wisdom provides a public API for the devwisdom-go wisdom engine.
// This package wraps the internal wisdom engine to make it accessible to other modules.
package wisdom

import (
	internal "github.com/davidl71/devwisdom-go/internal/wisdom"
)

// Engine is the main wisdom engine managing sources, advisors, and consultations.
// It provides thread-safe access to wisdom sources and advisor consultations.
// The engine must be initialized before use by calling Initialize().
type Engine = internal.Engine

// NewEngine creates a new wisdom engine instance.
// The engine is not initialized by default; call Initialize() before use.
func NewEngine() *Engine {
	return internal.NewEngine()
}

// Quote represents a wisdom quote with its metadata.
type Quote = internal.Quote

// Source represents a wisdom source with its quotes.
type Source = internal.Source

// AdvisorInfo represents information about an advisor.
type AdvisorInfo = internal.AdvisorInfo

// AdvisorRegistry manages advisors and their mappings.
type AdvisorRegistry = internal.AdvisorRegistry

// GetAeonLevel determines the aeon level from a score (0-100).
func GetAeonLevel(score float64) string {
	return internal.GetAeonLevel(score)
}

// ConsultationModeConfig represents configuration for a consultation mode.
type ConsultationModeConfig = internal.ConsultationModeConfig

// GetConsultationMode returns the consultation mode based on score.
func GetConsultationMode(score float64) ConsultationModeConfig {
	return internal.GetConsultationMode(score)
}
