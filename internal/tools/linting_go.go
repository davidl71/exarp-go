// linting_go.go — Go linters: golangci-lint, go-vet, gofmt, goimports.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davidl71/exarp-go/internal/projectroot"
	"github.com/davidl71/exarp-go/internal/security"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runGolangciLint runs golangci-lint.
func runGolangciLint(ctx context.Context, path string, fix bool) (*LintResult, error) {
	// Check if golangci-lint is available
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "golangci-lint",
			Output:  "golangci-lint not found. Install it from https://golangci-lint.run/",
			Errors: []LintError{
				{
					Message:  "golangci-lint binary not found in PATH",
					Severity: "error",
				},
			},
		}, nil
	}

	// Find project root - prefer PROJECT_ROOT env var, then look for go.mod
	var projectRoot string

	var absPath string

	// Try PROJECT_ROOT environment variable first (set by MCP server)
	if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
		if _, err := os.Stat(filepath.Join(envRoot, "go.mod")); err == nil {
			projectRoot = envRoot
		}
	}

	// If PROJECT_ROOT not set or invalid, find by walking up from path
	if projectRoot == "" {
		// Convert path to absolute
		if filepath.IsAbs(path) {
			absPath = path
		} else {
			// Try PROJECT_ROOT first, then fall back to Getwd
			if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
				absPath = filepath.Join(envRoot, path)
			} else {
				wd, _ := os.Getwd()
				absPath = filepath.Join(wd, path)
			}
		}

		// Walk up from the path to find go.mod
		currentPath := absPath

		for {
			if _, err := os.Stat(filepath.Join(currentPath, "go.mod")); err == nil {
				projectRoot = currentPath
				break
			}

			parent := filepath.Dir(currentPath)
			if parent == currentPath {
				// Reached filesystem root, use current working directory
				projectRoot, _ = os.Getwd()
				break
			}

			currentPath = parent
		}
	} else {
		// PROJECT_ROOT is set, use it to resolve relative paths
		if filepath.IsAbs(path) {
			absPath = path
		} else {
			absPath = filepath.Join(projectRoot, path)
		}
	}

	// Determine if path is a directory
	pathInfo, err := os.Stat(absPath)
	isDir := err == nil && pathInfo.IsDir()

	// Build command (v2 uses --output.json.path instead of --out-format=json)
	args := []string{"run", "--output.json.path=stdout"}
	if fix {
		args = append(args, "--fix")
	}

	// Get relative path from project root
	relPath, err := filepath.Rel(projectRoot, absPath)
	if err != nil {
		// Fallback: use path as-is if we can't get relative
		relPath = path
	}

	// Add path with ... for recursive directory traversal
	if relPath == "." || relPath == "" {
		args = append(args, "./...")
	} else if isDir {
		// For directories, append /... for recursive search
		if !strings.HasSuffix(relPath, "/...") {
			args = append(args, relPath+"/...")
		} else {
			args = append(args, relPath)
		}
	} else {
		// For files, use as-is
		args = append(args, relPath)
	}

	cmd := exec.CommandContext(ctx, "golangci-lint", args...)
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// golangci-lint returns non-zero exit code when issues are found
	// This is expected behavior, not an error
	if err != nil && !strings.Contains(err.Error(), "exit status") {
		return &LintResult{
			Success: false,
			Linter:  "golangci-lint",
			Output:  outputStr,
			Errors: []LintError{
				{
					Message:  fmt.Sprintf("golangci-lint execution failed: %v", err),
					Severity: "error",
				},
			},
		}, nil
	}

	// Parse JSON output (v2: {"Issues":[...]}, v1: [...])
	type issuePos struct {
		Filename string `json:"Filename"`
		Offset   int    `json:"Offset"`
		Line     int    `json:"Line"`
		Column   int    `json:"Column"`
	}

	type issueStruct struct {
		FromLinter  string   `json:"FromLinter"`
		Text        string   `json:"Text"`
		SourceLines []string `json:"SourceLines,omitempty"`
		Pos         issuePos `json:"Pos"`
	}

	var lintErrors []LintError

	var issues []issueStruct
	if err := json.Unmarshal(output, &struct {
		Issues *[]issueStruct `json:"Issues"`
	}{Issues: &issues}); err != nil || len(issues) == 0 {
		// Fall back to v1 flat array format
		_ = json.Unmarshal(output, &issues)
	}

	for _, issue := range issues {
		lintErrors = append(lintErrors, LintError{
			File:     issue.Pos.Filename,
			Line:     issue.Pos.Line,
			Column:   issue.Pos.Column,
			Message:  issue.Text,
			Rule:     issue.FromLinter,
			Severity: "warning",
		})
	}
	// Fallback: treat output as text only if JSON parsing produced nothing and output doesn't look like JSON
	if len(lintErrors) == 0 && outputStr != "" && !strings.Contains(outputStr, `"Issues"`) && !strings.Contains(outputStr, `"FromLinter"`) {
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				lintErrors = append(lintErrors, LintError{
					Message:  line,
					Severity: "warning",
				})
			}
		}
	}

	success := len(lintErrors) == 0

	return &LintResult{
		Success: success,
		Linter:  "golangci-lint",
		Output:  outputStr,
		Errors:  lintErrors,
		Fixed:   fix && success,
	}, nil
}

// runGoVet runs go vet.
func runGoVet(ctx context.Context, path string) (*LintResult, error) {
	// Check if go is available
	if _, err := exec.LookPath("go"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "go-vet",
			Output:  "go command not found. Go must be installed.",
			Errors: []LintError{
				{
					Message:  "go binary not found in PATH",
					Severity: "error",
				},
			},
		}, nil
	}

	// Get project root for path validation
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		// Try to find project root
		var err error

		projectRoot, err = projectroot.FindFrom(path)
		if err != nil {
			// Fallback to current directory
			projectRoot, _ = os.Getwd()
		}
	}

	// Validate and sanitize path to prevent directory traversal
	_, relPath, err := security.ValidatePathWithinRoot(path, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Determine if path is a directory (validate again to get absPath)
	absPath, err := security.ValidatePath(path, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	pathInfo, err := os.Stat(absPath)
	isDir := err == nil && pathInfo != nil && pathInfo.IsDir()

	// Build command
	args := []string{"vet"}

	// Add path with ... for recursive directory traversal
	if relPath == "." || relPath == "" {
		args = append(args, "./...")
	} else if isDir {
		// For directories, append /... for recursive search
		if !strings.HasSuffix(relPath, "/...") {
			args = append(args, "./"+relPath+"/...")
		} else {
			args = append(args, "./"+relPath)
		}
	} else {
		// For files, use as-is
		args = append(args, "./"+relPath)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// go vet returns non-zero exit code when issues are found
	success := err == nil

	var lintErrors []LintError

	if !success && outputStr != "" {
		// Parse go vet output (format: filename:line:column: message)
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Try to parse go vet format
			parts := strings.SplitN(line, ":", 4)
			if len(parts) >= 4 {
				lintErrors = append(lintErrors, LintError{
					File:     parts[0],
					Message:  parts[3],
					Severity: "warning",
				})
			} else {
				lintErrors = append(lintErrors, LintError{
					Message:  line,
					Severity: "warning",
				})
			}
		}
	}

	return &LintResult{
		Success: success,
		Linter:  "go-vet",
		Output:  outputStr,
		Errors:  lintErrors,
	}, nil
}

// runGofmt runs gofmt.
func runGofmt(ctx context.Context, path string, fix bool) (*LintResult, error) {
	// Check if gofmt is available
	if _, err := exec.LookPath("gofmt"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "gofmt",
			Output:  "gofmt not found. Go must be installed.",
			Errors: []LintError{
				{
					Message:  "gofmt binary not found in PATH",
					Severity: "error",
				},
			},
		}, nil
	}

	// Get project root for path validation
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		var err error

		projectRoot, err = projectroot.FindFrom(path)
		if err != nil {
			projectRoot, _ = os.Getwd()
		}
	}

	// Validate and sanitize path to prevent directory traversal
	_, relPath, err := security.ValidatePathWithinRoot(path, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Build command
	args := []string{"-d"}
	if fix {
		args = []string{"-w"}
	}

	// Add path
	if relPath != "." && relPath != "" {
		args = append(args, relPath)
	} else {
		args = append(args, ".")
	}

	cmd := exec.CommandContext(ctx, "gofmt", args...)
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// gofmt returns non-zero exit code when formatting issues are found
	success := err == nil && (outputStr == "" || !strings.Contains(outputStr, "diff"))

	var lintErrors []LintError

	if !success && outputStr != "" {
		// Parse gofmt diff output
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "diff ") {
				continue
			}

			if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
				continue
			}

			if strings.HasPrefix(line, "@@") {
				// Extract file and line info
				parts := strings.Fields(line)
				if len(parts) > 0 {
					lintErrors = append(lintErrors, LintError{
						Message:  fmt.Sprintf("Formatting issue at %s", parts[0]),
						Severity: "warning",
					})
				}
			} else if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "+") {
				// Formatting change
				lintErrors = append(lintErrors, LintError{
					Message:  line,
					Severity: "info",
				})
			}
		}
	}

	return &LintResult{
		Success: success,
		Linter:  "gofmt",
		Output:  outputStr,
		Errors:  lintErrors,
		Fixed:   fix && success,
	}, nil
}

// runGoimports runs goimports.
func runGoimports(ctx context.Context, path string, fix bool) (*LintResult, error) {
	// Check if goimports is available
	if _, err := exec.LookPath("goimports"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "goimports",
			Output:  "goimports not found. Install it with: go install golang.org/x/tools/cmd/goimports@latest",
			Errors: []LintError{
				{
					Message:  "goimports binary not found in PATH",
					Severity: "error",
				},
			},
		}, nil
	}

	// Get project root for path validation
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		var err error

		projectRoot, err = projectroot.FindFrom(path)
		if err != nil {
			projectRoot, _ = os.Getwd()
		}
	}

	// Validate and sanitize path to prevent directory traversal
	_, relPath, err := security.ValidatePathWithinRoot(path, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Build command
	args := []string{"-d"}
	if fix {
		args = []string{"-w"}
	}

	// Add path
	if relPath != "." && relPath != "" {
		args = append(args, relPath)
	} else {
		args = append(args, ".")
	}

	cmd := exec.CommandContext(ctx, "goimports", args...)
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// goimports returns non-zero exit code when import issues are found
	success := err == nil && (outputStr == "" || !strings.Contains(outputStr, "diff"))

	var lintErrors []LintError

	if !success && outputStr != "" {
		// Parse goimports diff output (similar to gofmt)
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "diff ") {
				continue
			}

			if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
				continue
			}

			if strings.HasPrefix(line, "@@") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					lintErrors = append(lintErrors, LintError{
						Message:  fmt.Sprintf("Import issue at %s", parts[0]),
						Severity: "warning",
					})
				}
			} else if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "+") {
				lintErrors = append(lintErrors, LintError{
					Message:  line,
					Severity: "info",
				})
			}
		}
	}

	return &LintResult{
		Success: success,
		Linter:  "goimports",
		Output:  outputStr,
		Errors:  lintErrors,
		Fixed:   fix && success,
	}, nil
}

// runDeadcode runs deadcode.
func runDeadcode(ctx context.Context, path string) (*LintResult, error) {
	// Check if deadcode is available
	if _, err := exec.LookPath("deadcode"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "deadcode",
			Output:  "deadcode not found. Install it with: go install golang.org/x/tools/cmd/deadcode@latest",
			Errors: []LintError{
				{
					Message:  "deadcode binary not found in PATH",
					Severity: "error",
				},
			},
		}, nil
	}

	// Get project root for path validation
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		var err error

		projectRoot, err = projectroot.FindFrom(path)
		if err != nil {
			projectRoot, _ = os.Getwd()
		}
	}

	// Validate and sanitize path to prevent directory traversal
	absPath, relPath, err := security.ValidatePathWithinRoot(path, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	pathInfo, err := os.Stat(absPath)
	isDir := err == nil && pathInfo != nil && pathInfo.IsDir()

	// Build command - deadcode takes packages as args
	args := []string{}
	if relPath == "." || relPath == "" {
		args = append(args, "./...")
	} else if isDir {
		if !strings.HasSuffix(relPath, "/...") {
			args = append(args, "./"+relPath+"/...")
		} else {
			args = append(args, "./"+relPath)
		}
	} else {
		// For files, get the package
		args = append(args, "./"+filepath.Dir(relPath)+"/...")
	}

	cmd := exec.CommandContext(ctx, "deadcode", args...)
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// deadcode returns non-zero when issues found
	success := err == nil || strings.Contains(err.Error(), "exit status")

	var lintErrors []LintError

	if outputStr != "" {
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Parse deadcode output: "package/path:file.go:line: unused function Foo"
			parts := strings.SplitN(line, ":", 4)
			if len(parts) >= 3 {
				lintErrors = append(lintErrors, LintError{
					File:     parts[1],
					Message:  strings.Join(parts[2:], ":"),
					Severity: "warning",
				})
			} else {
				lintErrors = append(lintErrors, LintError{
					Message:  line,
					Severity: "warning",
				})
			}
		}
	}

	return &LintResult{
		Success: success,
		Linter:  "deadcode",
		Output:  outputStr,
		Errors:  lintErrors,
	}, nil
}

// detectLinter automatically detects the appropriate linter based on file extension.
func detectLinter(path string) string {
	// Check if path is a directory or file
	info, err := os.Stat(path)
	if err != nil {
		return "go-vet" // Default to Go linter
	}

	// If it's a directory, default to Go linter
	if info.IsDir() {
		return "go-vet"
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go-vet"
	case ".md", ".markdown":
		return "markdownlint"
	case ".sh", ".bash":
		return "shellcheck"
	case ".c", ".cc", ".cpp", ".cxx", ".h", ".hpp", ".hxx":
		return "clang-tidy"
	case ".py":
		return "ruff"
	case ".rs":
		return "clippy"
	default:
		return "go-vet" // Default to Go linter
	}
}
