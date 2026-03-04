package tools

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestGatewaySupported(t *testing.T) {
	restore := os.Getenv("OPENAI_GATEWAY_BASE_URL")
	defer func() {
		if restore != "" {
			_ = os.Setenv("OPENAI_GATEWAY_BASE_URL", restore)
		} else {
			_ = os.Unsetenv("OPENAI_GATEWAY_BASE_URL")
		}
	}()

	_ = os.Unsetenv("OPENAI_GATEWAY_BASE_URL")
	if DefaultGatewayProvider().Supported() {
		t.Error("Supported() should be false when OPENAI_GATEWAY_BASE_URL is unset")
	}

	_ = os.Setenv("OPENAI_GATEWAY_BASE_URL", "http://localhost:6060")
	if !DefaultGatewayProvider().Supported() {
		t.Error("Supported() should be true when OPENAI_GATEWAY_BASE_URL is set")
	}

	_ = os.Setenv("OPENAI_GATEWAY_BASE_URL", "   ")
	if DefaultGatewayProvider().Supported() {
		t.Error("Supported() should be false when OPENAI_GATEWAY_BASE_URL is only whitespace")
	}
}

func TestGatewayGenerate_NoBaseURL(t *testing.T) {
	restore := os.Getenv("OPENAI_GATEWAY_BASE_URL")
	defer func() {
		if restore != "" {
			_ = os.Setenv("OPENAI_GATEWAY_BASE_URL", restore)
		} else {
			_ = os.Unsetenv("OPENAI_GATEWAY_BASE_URL")
		}
	}()
	_ = os.Unsetenv("OPENAI_GATEWAY_BASE_URL")

	_, err := DefaultGatewayProvider().Generate(context.Background(), "hi", 10, 0.5)
	if err == nil {
		t.Fatal("expected error when OPENAI_GATEWAY_BASE_URL is unset")
	}
	if err != nil && !strings.Contains(err.Error(), "OPENAI_GATEWAY_BASE_URL") {
		t.Errorf("error should mention OPENAI_GATEWAY_BASE_URL: %v", err)
	}
}

func TestGateway_ImplementsTextGeneratorWithModel(t *testing.T) {
	// Ensure gateway can be used when model param is passed (type assertion in text_generate).
	var _ TextGeneratorWithModel = (*gatewayTextGenerator)(nil)
}
