package platform

import (
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	info := Detect()
	if info == nil {
		t.Fatal("Detect() returned nil")
	}
	if info.GOOS != runtime.GOOS {
		t.Errorf("info.GOOS = %q, want %q", info.GOOS, runtime.GOOS)
	}
	if info.GOARCH != runtime.GOARCH {
		t.Errorf("info.GOARCH = %q, want %q", info.GOARCH, runtime.GOARCH)
	}
}

func TestOS(t *testing.T) {
	os := OS()
	if os == OSUnknown {
		// Unknown OS is valid if runtime.GOOS is not one we recognize
		if runtime.GOOS == "windows" || runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			t.Errorf("OS() returned OSUnknown for known OS: %q", runtime.GOOS)
		}
	}
}

func TestArchitecture(t *testing.T) {
	arch := Architecture()
	if arch == ArchUnknown {
		// Unknown architecture is valid if runtime.GOARCH is not one we recognize
		knownArchs := []string{"amd64", "arm64", "386", "arm", "x86_64", "aarch64", "i386"}
		isKnown := false
		for _, known := range knownArchs {
			if runtime.GOARCH == known {
				isKnown = true
				break
			}
		}
		if isKnown {
			t.Errorf("Architecture() returned ArchUnknown for known architecture: %q", runtime.GOARCH)
		}
	}
}

func TestIsWindows(t *testing.T) {
	expected := runtime.GOOS == "windows"
	got := IsWindows()
	if got != expected {
		t.Errorf("IsWindows() = %v, want %v", got, expected)
	}
}

func TestIsLinux(t *testing.T) {
	expected := runtime.GOOS == "linux"
	got := IsLinux()
	if got != expected {
		t.Errorf("IsLinux() = %v, want %v", got, expected)
	}
}

func TestIsDarwin(t *testing.T) {
	expected := runtime.GOOS == "darwin"
	got := IsDarwin()
	if got != expected {
		t.Errorf("IsDarwin() = %v, want %v", got, expected)
	}
}

func TestIsUnix(t *testing.T) {
	expected := runtime.GOOS == "linux" || runtime.GOOS == "darwin"
	got := IsUnix()
	if got != expected {
		t.Errorf("IsUnix() = %v, want %v", got, expected)
	}
}

func TestIs64Bit(t *testing.T) {
	expected := runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64" || runtime.GOARCH == "x86_64" || runtime.GOARCH == "aarch64"
	got := Is64Bit()
	if got != expected {
		t.Errorf("Is64Bit() = %v, want %v", got, expected)
	}
}

func TestIs32Bit(t *testing.T) {
	expected := runtime.GOARCH == "386" || runtime.GOARCH == "arm" || runtime.GOARCH == "i386"
	got := Is32Bit()
	if got != expected {
		t.Errorf("Is32Bit() = %v, want %v", got, expected)
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "forward slashes",
			input:    "path/to/file",
			expected: "path/to/file",
		},
		{
			name:     "backslashes on Windows",
			input:    "path\\to\\file",
			expected: "path/to/file", // Normalized to forward slashes
		},
		{
			name:     "mixed slashes",
			input:    "path/to\\file",
			expected: "path/to/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePath(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPathSeparator(t *testing.T) {
	separator := PathSeparator()
	if IsWindows() {
		if separator != "\\" {
			t.Errorf("PathSeparator() = %q, want %q", separator, "\\")
		}
	} else {
		if separator != "/" {
			t.Errorf("PathSeparator() = %q, want %q", separator, "/")
		}
	}
}

func TestPathListSeparator(t *testing.T) {
	separator := PathListSeparator()
	if IsWindows() {
		if separator != ";" {
			t.Errorf("PathListSeparator() = %q, want %q", separator, ";")
		}
	} else {
		if separator != ":" {
			t.Errorf("PathListSeparator() = %q, want %q", separator, ":")
		}
	}
}

func TestPlatformInfo_String(t *testing.T) {
	info := Detect()
	str := info.String()
	if str == "" {
		t.Error("PlatformInfo.String() returned empty string")
	}
	// Should contain OS and Architecture
	if !contains(str, string(info.OS)) {
		t.Errorf("PlatformInfo.String() = %q, should contain OS %q", str, info.OS)
	}
	if !contains(str, string(info.Architecture)) {
		t.Errorf("PlatformInfo.String() = %q, should contain Architecture %q", str, info.Architecture)
	}
}

func TestPlatformInfo_IsCompatible(t *testing.T) {
	info := Detect()
	
	// Should be compatible with itself
	if !info.IsCompatible(info.OS, info.Architecture) {
		t.Errorf("PlatformInfo.IsCompatible(%q, %q) = false, want true", info.OS, info.Architecture)
	}
	
	// Should not be compatible with different OS
	if info.IsCompatible(OSUnknown, info.Architecture) && info.OS != OSUnknown {
		t.Errorf("PlatformInfo.IsCompatible(%q, %q) = true, want false", OSUnknown, info.Architecture)
	}
	
	// Should not be compatible with different architecture
	if info.IsCompatible(info.OS, ArchUnknown) && info.Architecture != ArchUnknown {
		t.Errorf("PlatformInfo.IsCompatible(%q, %q) = true, want false", info.OS, ArchUnknown)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
