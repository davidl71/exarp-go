//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"context"

	"github.com/davidl71/exarp-go/internal/platform"
)

// appleFMProvider implements FMProvider using Apple Foundation Models (go-foundationmodels).
// Reuses GenerateWithOptions from apple_foundation.go.
type appleFMProvider struct{}

func (p *appleFMProvider) Supported() bool {
	return platform.CheckAppleFoundationModelsSupport().Supported
}

func (p *appleFMProvider) Generate(_ context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	if !p.Supported() {
		return "", ErrFMNotSupported
	}

	return GenerateWithOptions(prompt, maxTokens, temperature)
}
