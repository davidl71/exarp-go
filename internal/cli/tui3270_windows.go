//go:build windows

// tui3270_windows.go — Windows has no setsid detach; parent always runs the server in-process.
package cli

// tryDetachTUI3270 is a no-op on Windows. The caller falls back to foreground server startup.
func tryDetachTUI3270(pidFileFlag string) (bool, error) {
	_ = pidFileFlag

	return false, nil
}
