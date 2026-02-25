// linting.go — MCP lint tool: LintResult/LintError types and runLinter dispatcher.
package tools

import (
	"context"
	"fmt"
	"os"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/security"
)

// LintResult represents the result of a linting operation.
type LintResult struct {
	Success bool                   `json:"success"`
	Output  string                 `json:"output,omitempty"`
	Errors  []LintError            `json:"errors,omitempty"`
	Fixed   bool                   `json:"fixed,omitempty"`
	Linter  string                 `json:"linter"`
	Raw     map[string]interface{} `json:"raw,omitempty"`
}

// LintError represents a single linting error.
type LintError struct {
	File     string `json:"file"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
	Message  string `json:"message"`
	Rule     string `json:"rule,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// runLinter executes a linter command and returns the result.
func runLinter(ctx context.Context, linter, path string, fix bool) (*LintResult, error) {
	// Set timeout from config
	ctx, cancel := context.WithTimeout(ctx, config.ToolTimeout("linting"))
	defer cancel()

	// Get project root for path validation
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		// Try to find project root
		var err error

		projectRoot, err = FindProjectRoot()
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

