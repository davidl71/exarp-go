package platform

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// AppleFoundationModelsSupport represents the support status for Apple Foundation Models
type AppleFoundationModelsSupport struct {
	Supported         bool
	Reason            string
	OSVersion         string
	Architecture      string
	AppleIntelligence bool
}

// CheckAppleFoundationModelsSupport checks if the current platform supports Apple Foundation Models
func CheckAppleFoundationModelsSupport() AppleFoundationModelsSupport {
	result := AppleFoundationModelsSupport{
		Architecture: runtime.GOARCH,
	}

	// Check OS
	if runtime.GOOS != "darwin" {
		result.Reason = fmt.Sprintf("not macOS (current OS: %s)", runtime.GOOS)
		return result
	}

	// Check architecture (must be arm64 for Apple Silicon)
	if runtime.GOARCH != "arm64" {
		result.Reason = fmt.Sprintf("not Apple Silicon (current arch: %s)", runtime.GOARCH)
		return result
	}

	// Get macOS version
	osVersion, err := getMacOSVersion()
	if err != nil {
		result.Reason = fmt.Sprintf("failed to detect macOS version: %v", err)
		return result
	}
	result.OSVersion = osVersion

	// Check if macOS version is 26.0 or later (Tahoe)
	if !isMacOS26OrLater(osVersion) {
		result.Reason = fmt.Sprintf("macOS version too old (current: %s, required: 26.0+)", osVersion)
		return result
	}

	// Check Apple Intelligence availability (simplified check - would need actual API call)
	// For now, we assume if OS and arch are correct, Apple Intelligence might be available
	// The actual check would require calling into Foundation Models framework
	result.AppleIntelligence = true // Placeholder - actual check would use Foundation Models API

	result.Supported = true
	result.Reason = "platform supports Apple Foundation Models"

	return result
}

// getMacOSVersion gets the macOS version string
func getMacOSVersion() (string, error) {
	// Use sw_vers command to get macOS version
	cmd := exec.Command("sw_vers", "-productVersion")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run sw_vers: %w", err)
	}
	return strings.TrimSpace(out.String()), nil
}

// isMacOS26OrLater checks if the macOS version is 26.0 or later
func isMacOS26OrLater(version string) bool {
	// Parse version string (e.g., "26.3" or "26.0")
	parts := strings.Split(version, ".")
	if len(parts) < 1 {
		return false
	}

	// Check major version
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	return major >= 26
}
