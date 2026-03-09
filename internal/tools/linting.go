// linting.go — MCP lint tool: LintResult/LintError types and runLinter dispatcher.
package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/davidl71/exarp-go/internal/config"
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

	targetPath := path
	if targetPath == "" {
		targetPath = "."
	}

	if linter == "" || linter == "auto" {
		if isDirectoryPath(targetPath) {
			return runAutoLinter(ctx, targetPath, fix)
		}
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
	case "deadcode":
		return runDeadcode(ctx, targetPath)
	case "markdownlint", "markdownlint-cli", "mdl", "markdown":
		return runMarkdownlint(ctx, targetPath, fix)
	case "shellcheck", "shfmt", "shell":
		return runShellcheck(ctx, targetPath, fix)
	// C / C++
	case "clang-tidy", "cppcheck", "c", "cpp", "c++":
		return runClangTidy(ctx, targetPath, fix)
	case "clang-format":
		return runClangFormat(ctx, targetPath, fix)
	// Python
	case "ruff", "flake8", "pylint", "python":
		return runRuff(ctx, targetPath, fix)
	// Rust
	case "clippy", "cargo-clippy", "cargo clippy", "rust":
		return runCargoClippy(ctx, targetPath, fix)
	case "rustfmt":
		return runRustfmt(ctx, targetPath, fix)
	// PHP
	case "phpcs", "phpstan", "php-cs-fixer", "phpcbf", "php":
		return runPHPCS(ctx, targetPath, fix)
	// LaTeX
	case "chktex", "lacheck", "latex", "tex":
		return runChktex(ctx, targetPath, fix)
	default:
		return nil, fmt.Errorf("unsupported linter: %s (supported: golangci-lint, go-vet, gofmt, goimports, deadcode, markdownlint, shellcheck, clang-tidy, cppcheck, clang-format, ruff, flake8, pylint, clippy, rustfmt, phpcs, phpstan, php-cs-fixer, chktex, lacheck)", linter)
	}
}

func runAutoLinter(ctx context.Context, path string, fix bool) (*LintResult, error) {
	linters, err := detectLintersForPath(path)
	if err != nil {
		return nil, err
	}
	if len(linters) == 0 {
		linters = []string{"go-vet"}
	}

	combined := &LintResult{
		Success: true,
		Linter:  "auto",
		Raw: map[string]interface{}{
			"selected_linters": linters,
		},
	}

	outputs := make([]map[string]interface{}, 0, len(linters))
	for _, selected := range linters {
		result, err := runLinter(ctx, selected, path, fix)
		if err != nil {
			return nil, err
		}

		combined.Success = combined.Success && result.Success
		combined.Errors = append(combined.Errors, result.Errors...)
		if result.Output != "" {
			if combined.Output != "" {
				combined.Output += "\n"
			}
			combined.Output += fmt.Sprintf("[%s]\n%s", result.Linter, result.Output)
		}

		outputs = append(outputs, map[string]interface{}{
			"linter":  result.Linter,
			"success": result.Success,
			"errors":  len(result.Errors),
		})
	}

	combined.Raw["results"] = outputs
	return combined, nil
}

func isDirectoryPath(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func detectLintersForPath(path string) ([]string, error) {
	linters := make([]string, 0, 4)

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{detectLinter(path)}, nil
	}

	hasGo := false
	hasMarkdown := false
	hasShell := false

	err = filepath.WalkDir(path, func(current string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "bin", "build", "dist", "out", ".cache", ".todo2", ".exarp", ".task", "node_modules":
				if current != path {
					return filepath.SkipDir
				}
			case "archive":
				if filepath.Base(filepath.Dir(current)) == "docs" {
					return filepath.SkipDir
				}
			case "plans":
				if filepath.Base(filepath.Dir(current)) == ".cursor" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		switch filepath.Ext(current) {
		case ".go":
			hasGo = true
		case ".md", ".markdown":
			hasMarkdown = true
		case ".sh", ".bash":
			hasShell = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if hasGo {
		linters = append(linters, preferredGoAutoLinter())
	}
	if hasMarkdown {
		linters = append(linters, "markdownlint")
	}
	if hasShell {
		linters = append(linters, "shellcheck")
	}

	return slices.Compact(linters), nil
}

func preferredGoAutoLinter() string {
	if _, err := exec.LookPath("golangci-lint"); err == nil {
		return "golangci-lint"
	}
	return "go-vet"
}
