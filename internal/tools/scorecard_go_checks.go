// scorecard_go_checks.go — Go scorecard: per-file stats, filter candidates, health check functions.
package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"github.com/davidl71/exarp-go/internal/config"
)

// collectPerFileCodeStats returns per-file stats (path, lines, bytes, estimated tokens) for .go, _test.go, and bridge/tests .py.
// Used for large-file detection (split/refactor candidates) and for optional token-based analysis.
func collectPerFileCodeStats(projectRoot string) ([]FileSizeInfo, error) {
	var out []FileSizeInfo

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".venv" || info.Name() == "vendor" || info.Name() == ".git" || info.Name() == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, _ := filepath.Rel(projectRoot, path)
		if rel == "" || rel == "." {
			return nil
		}

		var add bool
		if strings.HasSuffix(path, ".go") {
			if skipGeneratedProtobufGo(projectRoot, path) {
				return nil
			}
			add = true
		} else if strings.HasSuffix(path, ".py") && (strings.Contains(path, "bridge/") || strings.Contains(path, "tests/")) {
			add = true
		}
		if !add {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable
		}

		lines := len(strings.Split(string(data), "\n"))
		bytes := len(data)
		tokens := int(float64(bytes) * config.TokensPerChar())

		out = append(out, FileSizeInfo{
			Path:            rel,
			Lines:           lines,
			Bytes:           bytes,
			EstimatedTokens: tokens,
		})
		return nil
	})

	return out, err
}

// filterLargeFileCandidates returns files that exceed the token or line threshold (split/refactor candidates).
func filterLargeFileCandidates(files []FileSizeInfo, tokenThreshold, lineThreshold int) []FileSizeInfo {
	var out []FileSizeInfo
	for _, f := range files {
		if f.EstimatedTokens >= tokenThreshold || f.Lines >= lineThreshold {
			out = append(out, f)
		}
	}
	return out
}

// countGoFilesWithBytes counts Go source files, lines, and total bytes (for token estimate).
func countGoFilesWithBytes(root string) (int, int, int, error) {
	var count, lines, bytes int

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip .venv, vendor, .git directories
			if info.Name() == ".venv" || info.Name() == "vendor" || info.Name() == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			if skipGeneratedProtobufGo(root, path) {
				return nil
			}
			count++
			data, err := os.ReadFile(path)
			if err == nil {
				lines += len(strings.Split(string(data), "\n"))
				bytes += len(data)
			}
		}

		return nil
	})

	return count, lines, bytes, err
}

// countGoTestFilesWithBytes counts Go test files, lines, and total bytes (for token estimate).
func countGoTestFilesWithBytes(root string) (int, int, int, error) {
	var count, lines, bytes int

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".venv" || info.Name() == "vendor" || info.Name() == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			count++

			data, err := os.ReadFile(path)
			if err == nil {
				lines += len(strings.Split(string(data), "\n"))
				bytes += len(data)
			}
		}

		return nil
	})

	return count, lines, bytes, err
}

// countPythonFilesWithBytes counts Python files and lines (bridge scripts only) and total bytes (for token estimate).
func countPythonFilesWithBytes(root string) (int, int, int, error) {
	var count, lines, bytes int

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".venv" || info.Name() == "__pycache__" || info.Name() == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, ".py") {
			// Only count bridge scripts and tests
			if strings.Contains(path, "bridge/") || strings.Contains(path, "tests/") {
				count++

				data, err := os.ReadFile(path)
				if err == nil {
					lines += len(strings.Split(string(data), "\n"))
					bytes += len(data)
				}
			}
		}

		return nil
	})

	return count, lines, bytes, err
}

// getGoModuleInfo gets Go module dependency count and version.
func getGoModuleInfo(ctx context.Context, root string) (int, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get module path and version
	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-f", "{{.Path}} {{.Version}}")
	cmd.Dir = root

	output, err := cmd.Output()
	if err != nil {
		return 0, "", err
	}

	parts := strings.Fields(string(output))
	version := "unknown"

	if len(parts) >= 2 {
		version = parts[1]
	}

	// Count dependencies
	cmd = exec.CommandContext(ctx, "go", "list", "-m", "all")
	cmd.Dir = root

	output, err = cmd.Output()
	if err != nil {
		return 0, version, nil // Non-fatal
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	deps := len(lines) - 1 // Subtract 1 for the main module

	return deps, version, nil
}

// getGoVersion gets the Go version.
func getGoVersion(ctx context.Context) (string, bool) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "version")

	output, err := cmd.Output()
	if err != nil {
		return "unknown", false
	}

	version := strings.TrimSpace(string(output))

	return version, true
}

// checkGoModTidy checks if go mod tidy passes.
func checkGoModTidy(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = root
	err := cmd.Run()

	return err == nil
}

// checkGoBuild checks if go build succeeds.
func checkGoBuild(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, config.ToolTimeout("scorecard"))
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "./...")
	cmd.Dir = root
	err := cmd.Run()

	return err == nil
}

// checkGoVet checks if go vet passes.
func checkGoVet(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "vet", "./...")
	cmd.Dir = root
	err := cmd.Run()

	return err == nil
}

// checkGoFmt checks if code is gofmt compliant.
func checkGoFmt(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var goFiles []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "vendor" || info.Name() == ".venv" {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, ".go") {
			if !skipGeneratedProtobufGo(root, path) {
				goFiles = append(goFiles, path)
			}
		}

		return nil
	})
	if err != nil {
		return false
	}

	if len(goFiles) == 0 {
		return true
	}

	const batchSize = 200
	for i := 0; i < len(goFiles); i += batchSize {
		end := i + batchSize
		if end > len(goFiles) {
			end = len(goFiles)
		}

		args := append([]string{"-l"}, goFiles[i:end]...)
		cmd := exec.CommandContext(ctx, "gofmt", args...)

		output, err := cmd.Output()
		if err != nil {
			return false
		}

		if strings.TrimSpace(string(output)) != "" {
			return false
		}
	}

	return true
}

// checkGolangciLintConfigured checks if golangci-lint is configured.
func checkGolangciLintConfigured(root string) bool {
	// Check for .golangci.yml or .golangci.yaml
	if _, err := os.Stat(filepath.Join(root, ".golangci.yml")); err == nil {
		return true
	}

	if _, err := os.Stat(filepath.Join(root, ".golangci.yaml")); err == nil {
		return true
	}

	return false
}

// checkGolangciLint checks if golangci-lint passes.
func checkGolangciLint(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, config.ToolTimeout("scorecard"))
	defer cancel()

	// Check if golangci-lint is available
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return false
	}

	cmd := exec.CommandContext(ctx, "golangci-lint", "run", "--timeout", "30s")
	cmd.Dir = root
	err := cmd.Run()

	return err == nil
}

// readCoverageFromFile parses coverage percentage from a coverage profile file.
// Returns 0.0 if the file does not exist or cannot be parsed. Used by both
// full mode (after go test) and fast mode (from prior coverage.out if present).
func readCoverageFromFile(ctx context.Context, root, filename string) float64 {
	coverPath := filepath.Join(root, filename)
	if _, err := os.Stat(coverPath); err != nil {
		return 0.0
	}
	coverCmd := exec.CommandContext(ctx, "go", "tool", "cover", "-func="+filename)
	coverCmd.Dir = root
	output, err := coverCmd.Output()
	if err != nil {
		return 0.0
	}
	lines := strings.Split(string(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "total:") {
			fields := strings.Fields(lines[i])
			if len(fields) > 0 {
				percentText := strings.TrimSuffix(fields[len(fields)-1], "%")
				if percent, parseErr := strconv.ParseFloat(percentText, 64); parseErr == nil {
					return percent
				}
			}
			break
		}
	}
	return 0.0
}

// checkGoTest checks if go test passes and gets coverage.
// Coverage is read from coverage.out even when tests fail, since go test may still
// write partial coverage for packages that ran before a failure.
func checkGoTest(ctx context.Context, root string) (bool, float64) {
	ctx, cancel := context.WithTimeout(ctx, config.ToolTimeout("testing"))
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "test", "./...", "-coverprofile=coverage.out", "-covermode=atomic", "-coverpkg=./internal/...")
	cmd.Dir = root
	err := cmd.Run()
	passes := err == nil

	coverage := readCoverageFromFile(ctx, root, "coverage.out")
	_ = os.Remove(filepath.Join(root, "coverage.out"))
	return passes, coverage
}

// checkGoVulncheck checks if govulncheck passes.
func checkGoVulncheck(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, config.ToolTimeout("scorecard"))
	defer cancel()

	// Check if govulncheck is available
	if _, err := exec.LookPath("govulncheck"); err != nil {
		return false // Not installed, not a failure
	}

	cmd := exec.CommandContext(ctx, "govulncheck", "./...")
	cmd.Dir = root
	err := cmd.Run()

	return err == nil
}

// checkPathBoundaryEnforcement checks if path boundary enforcement is implemented.
func checkPathBoundaryEnforcement(projectRoot string) bool {
	// Check if security/path.go exists and has ValidatePath function
	pathFile := filepath.Join(projectRoot, "internal", "security", "path.go")
	if _, err := os.Stat(pathFile); err != nil {
		return false
	}
	// Read file and check for ValidatePath function
	data, err := os.ReadFile(pathFile)
	if err != nil {
		return false
	}

	content := string(data)

	return strings.Contains(content, "func ValidatePath") && strings.Contains(content, "ValidatePathWithinRoot")
}

// checkRateLimiting checks if rate limiting is implemented.
func checkRateLimiting(projectRoot string) bool {
	// Check if security/ratelimit.go exists
	ratelimitFile := filepath.Join(projectRoot, "internal", "security", "ratelimit.go")
	if _, err := os.Stat(ratelimitFile); err != nil {
		return false
	}
	// Read file and check for RateLimiter
	data, err := os.ReadFile(ratelimitFile)
	if err != nil {
		return false
	}

	content := string(data)

	return strings.Contains(content, "type RateLimiter") && strings.Contains(content, "func Allow")
}

// checkAccessControl checks if access control is implemented.
func checkAccessControl(projectRoot string) bool {
	// Check if security/access.go exists
	accessFile := filepath.Join(projectRoot, "internal", "security", "access.go")
	if _, err := os.Stat(accessFile); err != nil {
		return false
	}
	// Read file and check for AccessControl
	data, err := os.ReadFile(accessFile)
	if err != nil {
		return false
	}

	content := string(data)

	return strings.Contains(content, "type AccessControl") && strings.Contains(content, "func CheckToolAccess")
}

