// linting_php_latex.go — PHP and LaTeX linters.
// PHP: phpcs (primary), phpstan (static analysis), php-cs-fixer (formatter/fix).
// LaTeX: chktex (primary), lacheck (fallback).
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ── PHP ──────────────────────────────────────────────────────────────────────

// runPHPCS runs PHP_CodeSniffer. Falls back to phpstan if phpcs is not available.
func runPHPCS(ctx context.Context, path string, fix bool) (*LintResult, error) {
	if fix {
		if _, err := exec.LookPath("phpcbf"); err == nil {
			return runPHPCBF(ctx, path)
		}
		if _, err := exec.LookPath("php-cs-fixer"); err == nil {
			return runPHPCSFixer(ctx, path)
		}
	}

	if _, err := exec.LookPath("phpcs"); err != nil {
		return runPHPStan(ctx, path)
	}

	projectRoot, _, relPath, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	target := relPath
	if target == "" {
		target = "."
	}

	args := []string{"--report=json", target}
	cmd := exec.CommandContext(ctx, "phpcs", args...)
	cmd.Dir = projectRoot
	out, _ := cmd.CombinedOutput()
	outputStr := string(out)

	lintErrors := parsePHPCSJSON(outputStr)

	return &LintResult{
		Success: len(lintErrors) == 0,
		Linter:  "phpcs",
		Output:  outputStr,
		Errors:  lintErrors,
	}, nil
}

// runPHPCBF runs PHP Code Beautifier and Fixer (auto-fix companion to phpcs).
func runPHPCBF(ctx context.Context, path string) (*LintResult, error) {
	projectRoot, _, relPath, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	target := relPath
	if target == "" {
		target = "."
	}

	cmd := exec.CommandContext(ctx, "phpcbf", target)
	cmd.Dir = projectRoot
	out, runErr := cmd.CombinedOutput()

	// phpcbf exits 1 when it fixed files, 2 on error
	fixed := runErr == nil || (runErr != nil && cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 1)

	return &LintResult{
		Success: true,
		Linter:  "phpcbf",
		Output:  string(out),
		Fixed:   fixed,
	}, nil
}

// runPHPCSFixer runs php-cs-fixer (alternative PHP formatter).
func runPHPCSFixer(ctx context.Context, path string) (*LintResult, error) {
	projectRoot, _, relPath, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	target := relPath
	if target == "" {
		target = "."
	}

	args := []string{"fix", target, "--diff", "--allow-risky=yes"}
	cmd := exec.CommandContext(ctx, "php-cs-fixer", args...)
	cmd.Dir = projectRoot
	out, runErr := cmd.CombinedOutput()

	return &LintResult{
		Success: runErr == nil,
		Linter:  "php-cs-fixer",
		Output:  string(out),
		Fixed:   runErr == nil,
	}, nil
}

// runPHPStan runs PHPStan (static analysis, fallback when phpcs not available).
func runPHPStan(ctx context.Context, path string) (*LintResult, error) {
	if _, err := exec.LookPath("phpstan"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "phpstan",
			Output:  "no PHP linter found. Install phpcs (composer global require squizlabs/php_codesniffer) or phpstan (composer global require phpstan/phpstan)",
			Errors:  []LintError{{Message: "phpcs and phpstan not found in PATH", Severity: "error"}},
		}, nil
	}

	projectRoot, _, relPath, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	target := relPath
	if target == "" {
		target = "."
	}

	args := []string{"analyse", "--error-format=json", "--no-progress", target}
	cmd := exec.CommandContext(ctx, "phpstan", args...)
	cmd.Dir = projectRoot
	out, _ := cmd.CombinedOutput()
	outputStr := string(out)

	lintErrors := parsePHPStanJSON(outputStr)

	return &LintResult{
		Success: len(lintErrors) == 0,
		Linter:  "phpstan",
		Output:  outputStr,
		Errors:  lintErrors,
	}, nil
}

// ── LaTeX ────────────────────────────────────────────────────────────────────

// runChktex runs chktex (LaTeX linter). Falls back to lacheck.
func runChktex(ctx context.Context, path string, _ bool) (*LintResult, error) {
	if _, err := exec.LookPath("chktex"); err != nil {
		return runLacheck(ctx, path)
	}

	projectRoot, absPath, relPath, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	info, _ := os.Stat(absPath)
	isDir := info != nil && info.IsDir()

	var files []string
	if isDir {
		files, err = findFiles(absPath, []string{".tex", ".ltx", ".sty", ".cls"})
		if err != nil || len(files) == 0 {
			return &LintResult{Success: true, Linter: "chktex", Output: "no LaTeX files found"}, nil
		}
	} else {
		files = []string{relPath}
	}

	var allErrors []LintError
	var allOutput strings.Builder

	for _, f := range files {
		// -v0 = machine-readable, -q = quiet
		cmd := exec.CommandContext(ctx, "chktex", "-v0", "-q", f)
		cmd.Dir = projectRoot
		out, _ := cmd.CombinedOutput()
		allOutput.Write(out)
		allErrors = append(allErrors, parseChktexOutput(string(out), f)...)
	}

	return &LintResult{
		Success: len(allErrors) == 0,
		Linter:  "chktex",
		Output:  allOutput.String(),
		Errors:  allErrors,
	}, nil
}

// runLacheck runs lacheck (LaTeX linter, fallback).
func runLacheck(ctx context.Context, path string) (*LintResult, error) {
	if _, err := exec.LookPath("lacheck"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "lacheck",
			Output:  "no LaTeX linter found. Install chktex (brew install chktex / apt install chktex) or lacheck (brew install lacheck / apt install lacheck)",
			Errors:  []LintError{{Message: "chktex and lacheck not found in PATH", Severity: "error"}},
		}, nil
	}

	projectRoot, absPath, relPath, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	info, _ := os.Stat(absPath)
	isDir := info != nil && info.IsDir()

	var files []string
	if isDir {
		files, err = findFiles(absPath, []string{".tex", ".ltx"})
		if err != nil || len(files) == 0 {
			return &LintResult{Success: true, Linter: "lacheck", Output: "no LaTeX files found"}, nil
		}
	} else {
		files = []string{relPath}
	}

	var allErrors []LintError
	var allOutput strings.Builder

	for _, f := range files {
		cmd := exec.CommandContext(ctx, "lacheck", f)
		cmd.Dir = projectRoot
		out, _ := cmd.CombinedOutput()
		allOutput.Write(out)
		allErrors = append(allErrors, parseLacheckOutput(string(out))...)
	}

	return &LintResult{
		Success: len(allErrors) == 0,
		Linter:  "lacheck",
		Output:  allOutput.String(),
		Errors:  allErrors,
	}, nil
}

// ── Parsers ──────────────────────────────────────────────────────────────────

// parsePHPCSJSON parses phpcs --report=json output.
// {"totals":{"errors":1,...},"files":{"/path/to/file.php":{"errors":1,...,"messages":[{"message":"...","source":"...","severity":5,"fixable":true,"type":"ERROR","line":10,"column":1}]}}}
func parsePHPCSJSON(output string) []LintError {
	var result struct {
		Files map[string]struct {
			Messages []struct {
				Message  string `json:"message"`
				Source   string `json:"source"`
				Type     string `json:"type"`
				Line     int    `json:"line"`
				Column   int    `json:"column"`
				Severity int    `json:"severity"`
			} `json:"messages"`
		} `json:"files"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return parseGCCStyleOutput(output)
	}

	var errors []LintError
	for file, data := range result.Files {
		for _, msg := range data.Messages {
			severity := "warning"
			if msg.Type == "ERROR" {
				severity = "error"
			}
			errors = append(errors, LintError{
				File:     file,
				Line:     msg.Line,
				Column:   msg.Column,
				Message:  msg.Message,
				Rule:     msg.Source,
				Severity: severity,
			})
		}
	}
	return errors
}

// parsePHPStanJSON parses phpstan --error-format=json output.
// {"totals":{"errors":0,"file_errors":1},"files":{"/path/file.php":{"errors":1,"messages":["Error message"]}}}
func parsePHPStanJSON(output string) []LintError {
	var result struct {
		Files map[string]struct {
			Messages []string `json:"messages"`
		} `json:"files"`
		Errors []string `json:"errors"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return parseGCCStyleOutput(output)
	}

	var errors []LintError
	for file, data := range result.Files {
		for _, msg := range data.Messages {
			errors = append(errors, LintError{
				File:     file,
				Message:  msg,
				Severity: "error",
			})
		}
	}
	for _, msg := range result.Errors {
		errors = append(errors, LintError{
			Message:  msg,
			Severity: "error",
		})
	}
	return errors
}

// parseChktexOutput parses chktex -v0 output.
// Format: file.tex:line:col:severity:number:message
func parseChktexOutput(output, fallbackFile string) []LintError {
	var errors []LintError
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 6)
		if len(parts) < 5 {
			if line != "" {
				errors = append(errors, LintError{File: fallbackFile, Message: line, Severity: "warning"})
			}
			continue
		}
		file := parts[0]
		lineNum, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		colNum, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
		warnNum := strings.TrimSpace(parts[3])
		msg := strings.TrimSpace(parts[4])
		if len(parts) >= 6 {
			msg = strings.TrimSpace(parts[5])
		}

		severity := "warning"
		if strings.Contains(strings.ToLower(warnNum), "error") {
			severity = "error"
		}

		errors = append(errors, LintError{
			File:     file,
			Line:     lineNum,
			Column:   colNum,
			Message:  msg,
			Rule:     fmt.Sprintf("chktex-%s", warnNum),
			Severity: severity,
		})
	}
	return errors
}

// parseLacheckOutput parses lacheck output.
// Format: "file.tex", line N: message
func parseLacheckOutput(output string) []LintError {
	var errors []LintError
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse "file.tex", line N: message
		if strings.HasPrefix(line, "\"") {
			if idx := strings.Index(line, "\", line "); idx > 0 {
				file := line[1:idx]
				rest := line[idx+8:] // skip "\", line "
				if colonIdx := strings.Index(rest, ": "); colonIdx > 0 {
					lineNum, _ := strconv.Atoi(rest[:colonIdx])
					msg := rest[colonIdx+2:]
					errors = append(errors, LintError{
						File:     file,
						Line:     lineNum,
						Message:  msg,
						Severity: "warning",
					})
					continue
				}
			}
		}

		errors = append(errors, LintError{Message: line, Severity: "warning"})
	}
	return errors
}
