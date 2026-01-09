package platform

import (
	"runtime"
	"testing"
)

func TestCheckAppleFoundationModelsSupport(t *testing.T) {
	result := CheckAppleFoundationModelsSupport()

	// Log the result for debugging
	t.Logf("Supported: %v", result.Supported)
	t.Logf("Reason: %s", result.Reason)
	t.Logf("OS Version: %s", result.OSVersion)
	t.Logf("Architecture: %s", result.Architecture)
	t.Logf("Apple Intelligence: %v", result.AppleIntelligence)

	// Basic sanity checks
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		// On Apple Silicon macOS, should have OS version
		if result.OSVersion == "" && result.Supported {
			t.Error("OS version should be detected on macOS")
		}
	} else {
		// On non-Apple platforms, should not be supported
		if result.Supported {
			t.Errorf("Should not be supported on %s/%s", runtime.GOOS, runtime.GOARCH)
		}
	}
}

func TestIsMacOS26OrLater(t *testing.T) {
	tests := []struct {
		version  string
		expected bool
	}{
		{"26.0", true},
		{"26.3", true},
		{"27.0", true},
		{"25.9", false},
		{"24.0", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := isMacOS26OrLater(tt.version)
			if result != tt.expected {
				t.Errorf("isMacOS26OrLater(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}
