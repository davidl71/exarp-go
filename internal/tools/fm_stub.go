//go:build !darwin || !arm64 || !cgo
// +build !darwin !arm64 !cgo

package tools

import "context"

// stubFMProvider implements FMProvider when Apple Foundation Models are not available.
type stubFMProvider struct{}

func (p *stubFMProvider) Supported() bool {
	return false
}

func (p *stubFMProvider) Generate(_ context.Context, _ string, _ int, _ float32) (string, error) {
	return "", ErrFMNotSupported
}

func init() {
	DefaultFM = &stubFMProvider{}
}
