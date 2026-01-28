//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

package tools

// appleFMIfAvailable returns nil when Apple FM is not available; used by the FM chain.
func appleFMIfAvailable() TextGenerator {
	return nil
}
