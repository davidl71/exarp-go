package tools

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestLocalAISupported(t *testing.T) {
	restore := os.Getenv("LOCALAI_BASE_URL")
	defer func() {
		if restore != "" {
			_ = os.Setenv("LOCALAI_BASE_URL", restore)
		} else {
			_ = os.Unsetenv("LOCALAI_BASE_URL")
		}
	}()

	_ = os.Unsetenv("LOCALAI_BASE_URL")
	if DefaultLocalAIProvider().Supported() {
		t.Error("Supported() should be false when LOCALAI_BASE_URL is unset")
	}

	_ = os.Setenv("LOCALAI_BASE_URL", "http://localhost:8080")
	if !DefaultLocalAIProvider().Supported() {
		t.Error("Supported() should be true when LOCALAI_BASE_URL is set")
	}

	_ = os.Setenv("LOCALAI_BASE_URL", "   ")
	if DefaultLocalAIProvider().Supported() {
		t.Error("Supported() should be false when LOCALAI_BASE_URL is only whitespace")
	}
}

func TestLocalAIGenerate_NoBaseURL(t *testing.T) {
	restore := os.Getenv("LOCALAI_BASE_URL")
	defer func() {
		if restore != "" {
			_ = os.Setenv("LOCALAI_BASE_URL", restore)
		} else {
			_ = os.Unsetenv("LOCALAI_BASE_URL")
		}
	}()
	_ = os.Unsetenv("LOCALAI_BASE_URL")

	_, err := DefaultLocalAIProvider().Generate(context.Background(), "hi", 10, 0.5)
	if err == nil {
		t.Fatal("expected error when LOCALAI_BASE_URL is unset")
	}
	if err != nil && !strings.Contains(err.Error(), "LOCALAI_BASE_URL") {
		t.Errorf("error should mention LOCALAI_BASE_URL: %v", err)
	}
}
