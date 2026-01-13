//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"encoding/json"
	"fmt"

	fm "github.com/blacktop/go-foundationmodels"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/platform"
)

// handleStatusAction handles the "status" action (no prompt required)
func handleStatusAction() ([]framework.TextContent, error) {
	support := platform.CheckAppleFoundationModelsSupport()
	
	status := map[string]interface{}{
		"supported": support.Supported,
		"reason":    support.Reason,
		"platform":  "darwin",
		"arch":      "arm64",
	}

	if support.Supported {
		// Try to check model availability
		available := fm.CheckModelAvailability()
		status["model_available"] = available
	}

	result, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal status: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(result)},
	}, nil
}

// handleHardwareAction handles the "hardware" action (no prompt required)
func handleHardwareAction() ([]framework.TextContent, error) {
	support := platform.CheckAppleFoundationModelsSupport()
	
	hardware := map[string]interface{}{
		"supported": support.Supported,
		"reason":    support.Reason,
	}

	if support.Supported {
		// Get hardware information
		hardware["platform"] = "darwin"
		hardware["arch"] = "arm64"
		hardware["cgo_enabled"] = true
	}

	result, err := json.MarshalIndent(hardware, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hardware info: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(result)},
	}, nil
}

// handleModelsAction handles the "models" action (no prompt required)
func handleModelsAction() ([]framework.TextContent, error) {
	support := platform.CheckAppleFoundationModelsSupport()
	
	models := map[string]interface{}{
		"supported": support.Supported,
		"reason":    support.Reason,
	}

	if support.Supported {
		// Check default model availability
		available := fm.CheckModelAvailability()
		models["default_model_available"] = available
		models["note"] = "Use CheckModelAvailability() to check model availability"
	}

	result, err := json.MarshalIndent(models, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal models info: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(result)},
	}, nil
}
