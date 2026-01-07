package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message"`
	Rule    string `json:"rule,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// runLinter executes a linter command and returns the result
func runLinter(ctx context.Context, linter, path string, fix bool) (*LintResult, error) {
	// Set default timeout
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Determine target path
	targetPath := path
	if targetPath == "" {
		targetPath = "."
	}

	// Normalize path
	if !filepath.IsAbs(targetPath) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		targetPath = filepath.Join(wd, targetPath)
	}

	// Check if path exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", targetPath)
	}

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
	default:
		return nil, fmt.Errorf("unsupported linter: %s (supported: golangci-lint, go-vet, gofmt, goimports, markdownlint)", linter)
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
					Message: "golangci-lint binary not found in PATH",
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
					Message: fmt.Sprintf("golangci-lint execution failed: %v", err),
					Severity: "error",
				},
			},
		}, nil
	}

	// Parse JSON output
	var issues []struct {
		FromLinter string `json:"FromLinter"`
		Text       string `json:"Text"`
		SourceLines []string `json:"SourceLines,omitempty"`
		Pos        struct {
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
						Message: line,
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
					Message: "go binary not found in PATH",
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
	args := []string{"vet"}
	
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
					Message: "gofmt binary not found in PATH",
					Severity: "error",
				},
			},
		}, nil
	}

	// Build command
	args := []string{"-d"}
	if fix {
		args = []string{"-w"}
	}
	
	// Add path
	relPath, err := filepath.Rel(".", path)
	if err != nil {
		relPath = path
	}
	if relPath != "." {
		args = append(args, relPath)
	} else {
		args = append(args, ".")
	}

	cmd := exec.CommandContext(ctx, "gofmt", args...)
	cmd.Dir = filepath.Dir(path)
	if cmd.Dir == "." {
		wd, _ := os.Getwd()
		cmd.Dir = wd
	}

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
					Message: "goimports binary not found in PATH",
					Severity: "error",
				},
			},
		}, nil
	}

	// Build command
	args := []string{"-d"}
	if fix {
		args = []string{"-w"}
	}
	
	// Add path
	relPath, err := filepath.Rel(".", path)
	if err != nil {
		relPath = path
	}
	if relPath != "." {
		args = append(args, relPath)
	} else {
		args = append(args, ".")
	}

	cmd := exec.CommandContext(ctx, "goimports", args...)
	cmd.Dir = filepath.Dir(path)
	if cmd.Dir == "." {
		wd, _ := os.Getwd()
		cmd.Dir = wd
	}

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
							Message: "markdown linter binary not found in PATH",
							Severity: "error",
						},
					},
				}, nil
			}
		}
	}

	// Determine if path is a directory or file
	info, err := os.Stat(path)
	isDir := err == nil && info.IsDir()

	// Find project root
	var projectRoot string
	var absPath string
	
	// Convert path to absolute
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
	
	// Walk up to find project root (directory with .gomarklint.json, .markdownlintrc, or go.mod)
	currentPath := absPath
	if !isDir {
		currentPath = filepath.Dir(absPath)
	}
	for {
		// Check for gomarklint config (preferred)
		if _, err := os.Stat(filepath.Join(currentPath, ".gomarklint.json")); err == nil {
			projectRoot = currentPath
			break
		}
		// Check for markdownlint config (fallback)
		if _, err := os.Stat(filepath.Join(currentPath, ".markdownlintrc")); err == nil {
			projectRoot = currentPath
			break
		}
		if _, err := os.Stat(filepath.Join(currentPath, ".markdownlint.json")); err == nil {
			projectRoot = currentPath
			break
		}
		// Check for go.mod (Go project root)
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
	
	// Get relative path from project root (needed for archive check)
	relPath, err := filepath.Rel(projectRoot, absPath)
	if err != nil {
		relPath = path
	}
	
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

