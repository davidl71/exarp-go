//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

// appleFMIfAvailable returns the Apple FM provider when built for darwin/arm64/cgo; used by the FM chain.
func appleFMIfAvailable() TextGenerator {
	return &appleFMProvider{}
}
