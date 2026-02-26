//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	fm "github.com/blacktop/go-foundationmodels"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/platform"
)

// handleStatusAction handles the "status" action (no prompt required).
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

	return framework.FormatResult(status, "")
}

// handleHardwareAction handles the "hardware" action (no prompt required).
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

	return framework.FormatResult(hardware, "")
}

// handleModelsAction handles the "models" action (no prompt required).
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

	return framework.FormatResult(models, "")
}
