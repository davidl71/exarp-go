package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/security"
)

// LintResult represents the result of a linting operation
type LintResult struct {
	Success bool                   `json:"success"`
	Output  string                 `json:"output,omitempty"`
	Errors  []LintError            `json:"errors,omitempty"`
	Fixed   bool                   `json:"fixed,omitempty"`
	Linter  string                 `json:"linter"`
	Raw     map[string]interface{} `json:"raw,omitempty"`
}

// LintError represents a single linting error
type LintError struct {
	File     string `json:"file"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
	Message  string `json:"message"`
	Rule     string `json:"rule,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// runLinter executes a linter command and returns the result
func runLinter(ctx context.Context, linter, path string, fix bool) (*LintResult, error) {
	// Set timeout from config
	ctx, cancel := context.WithTimeout(ctx, config.ToolTimeout("linting"))
	defer cancel()

	// Get project root for path validation
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		// Try to find project root
		var err error
		projectRoot, err = security.GetProjectRoot(".")
		if err != nil {
			// Fallback to current directory
			projectRoot, _ = os.Getwd()
		}
	}

	// Determine target path
	targetPath := path
	if targetPath == "" {
		targetPath = "."
	}

	// Validate and sanitize path to prevent directory traversal
	validatedPath, err := security.ValidatePathExists(targetPath, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	targetPath = validatedPath

	// Auto-detect linter based on file type if not specified
	if linter == "" || linter == "auto" {
		linter = detectLinter(targetPath)
	}

	// Auto-detect linter based on file type if not specified
	if linter == "" || linter == "auto" {
		linter = detectLinter(targetPath)
	}

	// Route to appropriate linter
	switch linter {
	case "golangci-lint", "golangcilint":
		return runGolangciLint(ctx, targetPath, fix)
	case "go-vet", "govet", "go vet":
		return runGoVet(ctx, targetPath)
	case "gofmt":
		return runGofmt(ctx, targetPath, fix)
	case "goimports":
		return runGoimports(ctx, targetPath, fix)
	case "markdownlint", "markdownlint-cli", "mdl", "markdown":
		return runMarkdownlint(ctx, targetPath, fix)
	case "shellcheck", "shfmt", "shell":
		return runShellcheck(ctx, targetPath, fix)
	default:
		return nil, fmt.Errorf("unsupported linter: %s (supported: golangci-lint, go-vet, gofmt, goimports, markdownlint, shellcheck)", linter)
	}
}

// runGolangciLint runs golangci-lint
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

	// Build command
	args := []string{"run", "--out-format=json"}
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

	// Parse JSON output
	var issues []struct {
		FromLinter  string   `json:"FromLinter"`
		Text        string   `json:"Text"`
		SourceLines []string `json:"SourceLines,omitempty"`
		Pos         struct {
			Filename string `json:"Filename"`
			Offset   int    `json:"Offset"`
			Line     int    `json:"Line"`
			Column   int    `json:"Column"`
		} `json:"Pos"`
	}

	var lintErrors []LintError
	if err := json.Unmarshal(output, &issues); err == nil {
		// Successfully parsed JSON
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
	} else {
		// Fallback: treat output as text
		if outputStr != "" {
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

// runGoVet runs go vet
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
		projectRoot, err = security.GetProjectRoot(path)
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

// runGofmt runs gofmt
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
		projectRoot, err = security.GetProjectRoot(path)
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

// runGoimports runs goimports
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
		projectRoot, err = security.GetProjectRoot(path)
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

// detectLinter automatically detects the appropriate linter based on file extension
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
	case ".md", ".markdown":
		return "markdownlint"
	case ".go":
		return "go-vet"
	case ".sh", ".bash":
		return "shellcheck"
	default:
		return "go-vet" // Default to Go linter
	}
}

// runMarkdownlint runs gomarklint (native Go markdown linter)
func runMarkdownlint(ctx context.Context, path string, fix bool) (*LintResult, error) {
	// Try gomarklint first (native Go tool)
	markdownlintCmd := "gomarklint"
	if _, err := exec.LookPath(markdownlintCmd); err != nil {
		// Fallback to markdownlint-cli (npm package)
		markdownlintCmd = "markdownlint-cli"
		if _, err := exec.LookPath(markdownlintCmd); err != nil {
			// Fallback to markdownlint (Ruby gem)
			markdownlintCmd = "markdownlint"
			if _, err := exec.LookPath(markdownlintCmd); err != nil {
				return &LintResult{
					Success: false,
					Linter:  "markdownlint",
					Output:  "No markdown linter found. Install gomarklint with: go install github.com/shinagawa-web/gomarklint@latest",
					Errors: []LintError{
						{
							Message:  "markdown linter binary not found in PATH",
							Severity: "error",
						},
					},
				}, nil
			}
		}
	}

	// Get project root for path validation
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		var err error
		projectRoot, err = security.GetProjectRoot(path)
		if err != nil {
			projectRoot, _ = os.Getwd()
		}
	}

	// Validate and sanitize path to prevent directory traversal
	absPath, relPath, err := security.ValidatePathWithinRoot(path, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Determine if path is a directory or file
	info, err := os.Stat(absPath)
	isDir := err == nil && info != nil && info.IsDir()

	// Exclude archive directory from linting
	if strings.Contains(absPath, "/archive/") || strings.Contains(relPath, "/archive/") {
		return &LintResult{
			Success: true,
			Linter:  "gomarklint",
			Output:  "Archive directory excluded from linting",
			Errors:  []LintError{},
		}, nil
	}

	// Build command - use JSON output for gomarklint, text for others
	args := []string{}
	useJSON := false

	if markdownlintCmd == "gomarklint" {
		// Use gomarklint with JSON output
		args = append(args, "--output=json")
		useJSON = true
	} else {
		// External tools (markdownlint-cli, markdownlint)
		if fix {
			args = append(args, "--fix")
		}
	}

	// relPath already calculated above for archive check

	// Add path(s) - all tools support files and directories
	if isDir {
		if markdownlintCmd == "gomarklint" {
			// gomarklint handles directories directly
			args = append(args, relPath)
		} else {
			// For external tools, use glob pattern
			args = append(args, filepath.Join(relPath, "**/*.md"))
		}
	} else {
		args = append(args, relPath)
	}

	cmd := exec.CommandContext(ctx, markdownlintCmd, args...)
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// All markdown linters return non-zero exit code when issues are found
	success := err == nil
	var lintErrors []LintError

	if !success && outputStr != "" {
		if useJSON {
			// Parse gomarklint JSON output
			// Structure: {"files": N, "lines": M, "errors": K, "details": {"file.md": [{"File": "...", "Line": N, "Message": "..."}]}}
			var jsonOutput struct {
				Details map[string][]struct {
					File    string `json:"File"`
					Line    int    `json:"Line"`
					Column  int    `json:"Column,omitempty"`
					Message string `json:"Message"`
					Rule    string `json:"Rule,omitempty"`
				} `json:"details"`
			}

			if err := json.Unmarshal(output, &jsonOutput); err == nil {
				// Successfully parsed JSON - iterate through files
				for fileName, issues := range jsonOutput.Details {
					for _, issue := range issues {
						lintErrors = append(lintErrors, LintError{
							File:     fileName,
							Line:     issue.Line,
							Column:   issue.Column,
							Message:  issue.Message,
							Rule:     issue.Rule,
							Severity: "warning",
						})
					}
				}
			} else {
				// JSON parsing failed, fall back to text parsing
				parseTextOutput(outputStr, &lintErrors)
			}
		} else {
			// Parse text output (for external tools)
			parseTextOutput(outputStr, &lintErrors)
		}
	}

	return &LintResult{
		Success: success,
		Linter:  "gomarklint",
		Output:  outputStr,
		Errors:  lintErrors,
		Fixed:   fix && success,
	}, nil
}

// parseTextOutput parses text-based markdownlint output
func parseTextOutput(outputStr string, lintErrors *[]LintError) {
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse format: path:line:column rule message
		// Example: docs/README.md:5:10 MD001/heading-increment Heading levels should only increment by one level at a time
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 3 {
			file := parts[0]
			lineCol := strings.Fields(parts[1])
			message := parts[2]

			lineNum := 0
			if len(lineCol) > 0 {
				fmt.Sscanf(lineCol[0], "%d", &lineNum)
			}

			// Extract rule name if present
			rule := ""
			messageParts := strings.Fields(message)
			if len(messageParts) > 0 && strings.Contains(messageParts[0], "/") {
				rule = messageParts[0]
				message = strings.Join(messageParts[1:], " ")
			}

			*lintErrors = append(*lintErrors, LintError{
				File:     file,
				Line:     lineNum,
				Message:  message,
				Rule:     rule,
				Severity: "warning",
			})
		} else {
			// Fallback: treat entire line as message
			*lintErrors = append(*lintErrors, LintError{
				Message:  line,
				Severity: "warning",
			})
		}
	}
}

// runShellcheck runs shellcheck (shell script linter)
func runShellcheck(ctx context.Context, path string, fix bool) (*LintResult, error) {
	// Try shellcheck first (comprehensive linter)
	shellcheckCmd := "shellcheck"
	if _, err := exec.LookPath(shellcheckCmd); err != nil {
		// Fallback to shfmt (formatter, can detect syntax errors)
		shellcheckCmd = "shfmt"
		if _, err := exec.LookPath(shellcheckCmd); err != nil {
			return &LintResult{
				Success: false,
				Linter:  "shellcheck",
				Output:  "No shell linter found. Install shellcheck with: brew install shellcheck or apt-get install shellcheck",
				Errors: []LintError{
					{
						Message:  "shell linter binary not found in PATH",
						Severity: "error",
					},
				},
			}, nil
		}
		// Use shfmt for formatting/checking
		return runShfmt(ctx, path, fix)
	}

	// Get project root for path validation
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		var err error
		projectRoot, err = security.GetProjectRoot(path)
		if err != nil {
			projectRoot, _ = os.Getwd()
		}
	}

	// Validate and sanitize path to prevent directory traversal
	absPath, relPath, err := security.ValidatePathWithinRoot(path, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Determine if path is a directory or file
	info, err := os.Stat(absPath)
	isDir := err == nil && info != nil && info.IsDir()

	// Build command - use JSON output for shellcheck
	args := []string{"--format=json"}
	if fix {
		// shellcheck doesn't have --fix, but we can note it
		// For actual fixing, would need shfmt
	}

	// Add path(s) - shellcheck supports files and directories
	if isDir {
		// For directories, find all .sh files
		args = append(args, filepath.Join(relPath, "*.sh"))
	} else {
		args = append(args, relPath)
	}

	cmd := exec.CommandContext(ctx, shellcheckCmd, args...)
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// shellcheck returns non-zero exit code when issues are found, but also outputs JSON
	// So we parse JSON regardless of exit code
	var lintErrors []LintError
	success := true // Will be set to false if we find errors

	if outputStr != "" {
		// Parse shellcheck JSON output
		// Format: [{"file":"path","line":N,"column":M,"level":"error|warning|info|style","code":2001,"message":"..."}]
		var jsonOutput []struct {
			File      string `json:"file"`
			Line      int    `json:"line"`
			EndLine   int    `json:"endLine,omitempty"`
			Column    int    `json:"column"`
			EndColumn int    `json:"endColumn,omitempty"`
			Level     string `json:"level"`
			Code      int    `json:"code"`
			Message   string `json:"message"`
		}

		if err := json.Unmarshal(output, &jsonOutput); err == nil && len(jsonOutput) > 0 {
			// Successfully parsed JSON
			success = false // Found issues
			for _, issue := range jsonOutput {
				severity := "warning"
				if issue.Level == "error" {
					severity = "error"
				} else if issue.Level == "info" {
					severity = "info"
				} else if issue.Level == "style" {
					severity = "warning"
				}

				// Convert code number to SC#### format
				code := fmt.Sprintf("SC%d", issue.Code)

				lintErrors = append(lintErrors, LintError{
					File:     issue.File,
					Line:     issue.Line,
					Column:   issue.Column,
					Message:  issue.Message,
					Rule:     code,
					Severity: severity,
				})
			}
		} else if err != nil {
			// JSON parsing failed, try to parse text output
			parseShellcheckTextOutput(outputStr, &lintErrors)
			if len(lintErrors) > 0 {
				success = false
			}
		}
	}

	return &LintResult{
		Success: success,
		Linter:  "shellcheck",
		Output:  outputStr,
		Errors:  lintErrors,
		Fixed:   false, // shellcheck doesn't auto-fix
	}, nil
}

// runShfmt runs shfmt (shell formatter, can detect syntax errors)
func runShfmt(ctx context.Context, path string, fix bool) (*LintResult, error) {
	// Check if shfmt is available
	if _, err := exec.LookPath("shfmt"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "shfmt",
			Output:  "shfmt not found. Install with: go install mvdan.cc/sh/v3/cmd/shfmt@latest",
			Errors: []LintError{
				{
					Message:  "shfmt binary not found in PATH",
					Severity: "error",
				},
			},
		}, nil
	}

	// Find project root
	var projectRoot string
	var absPath string

	if filepath.IsAbs(path) {
		absPath = path
	} else {
		if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
			absPath = filepath.Join(envRoot, path)
		} else {
			wd, _ := os.Getwd()
			absPath = filepath.Join(wd, path)
		}
	}

	// Walk up to find project root
	currentPath := absPath
	info, err := os.Stat(absPath)
	if err == nil && !info.IsDir() {
		currentPath = filepath.Dir(absPath)
	}
	for {
		if _, err := os.Stat(filepath.Join(currentPath, "go.mod")); err == nil {
			projectRoot = currentPath
			break
		}
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			projectRoot, _ = os.Getwd()
			break
		}
		currentPath = parent
	}

	relPath, err := filepath.Rel(projectRoot, absPath)
	if err != nil {
		relPath = path
	}

	// Build command
	args := []string{}
	if fix {
		args = append(args, "-w") // Write formatted output
	} else {
		args = append(args, "-d") // Diff mode (show what would change)
	}
	args = append(args, relPath)

	cmd := exec.CommandContext(ctx, "shfmt", args...)
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// shfmt returns non-zero if file needs formatting or has syntax errors
	success := err == nil
	var lintErrors []LintError

	if !success && outputStr != "" {
		// Parse shfmt output (diff format or error messages)
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Check for syntax errors
			if strings.Contains(line, "syntax error") || strings.Contains(line, "parse error") {
				lintErrors = append(lintErrors, LintError{
					File:     relPath,
					Message:  line,
					Severity: "error",
				})
			} else if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "+") {
				// Formatting differences
				lintErrors = append(lintErrors, LintError{
					File:     relPath,
					Message:  fmt.Sprintf("Formatting issue: %s", line),
					Severity: "warning",
				})
			}
		}
	}

	return &LintResult{
		Success: success,
		Linter:  "shfmt",
		Output:  outputStr,
		Errors:  lintErrors,
		Fixed:   fix && success,
	}, nil
}

// parseShellcheckTextOutput parses shellcheck text output
func parseShellcheckTextOutput(outputStr string, lintErrors *[]LintError) {
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse format: In path line N: column M: code message
		// Example: In scripts/check-go-health.sh line 121: column 24: SC2001 See if you can use ${variable//search/replace} instead.
		if strings.HasPrefix(line, "In ") {
			parts := strings.SplitN(line, ":", 4)
			if len(parts) >= 4 {
				file := strings.TrimPrefix(parts[0], "In ")
				file = strings.TrimSpace(file)

				lineNum := 0
				if strings.HasPrefix(parts[1], " line ") {
					fmt.Sscanf(parts[1], " line %d", &lineNum)
				}

				columnNum := 0
				if strings.HasPrefix(parts[2], " column ") {
					fmt.Sscanf(parts[2], " column %d", &columnNum)
				}

				codeAndMessage := strings.TrimSpace(parts[3])
				code := ""
				message := codeAndMessage

				// Extract code (e.g., "SC2001")
				fields := strings.Fields(codeAndMessage)
				if len(fields) > 0 && strings.HasPrefix(fields[0], "SC") {
					code = fields[0]
					message = strings.Join(fields[1:], " ")
				}

				*lintErrors = append(*lintErrors, LintError{
					File:     file,
					Line:     lineNum,
					Column:   columnNum,
					Message:  message,
					Rule:     code,
					Severity: "warning",
				})
			}
		} else {
			// Fallback: treat entire line as message
			*lintErrors = append(*lintErrors, LintError{
				Message:  line,
				Severity: "warning",
			})
		}
	}
}
