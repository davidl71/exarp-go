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
	var _ TextGeneratorWithModel = (*gatewayTextGenerator)(nil)
}

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

func TestLocalAI_ImplementsTextGeneratorWithModel(t *testing.T) {
	var _ TextGeneratorWithModel = (*localaiTextGenerator)(nil)
}
