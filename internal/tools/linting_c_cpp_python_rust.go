// linting_c_cpp_python_rust.go — C/C++, Python, and Rust linters.
// C/C++: clang-tidy (primary), cppcheck (fallback), clang-format (fix).
// Python: ruff (primary), flake8 (fallback), pylint (fallback).
// Rust: cargo clippy (linter), rustfmt (formatter/fix).
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/davidl71/exarp-go/internal/projectroot"
	"github.com/davidl71/exarp-go/internal/security"
)

// ── C / C++ ──────────────────────────────────────────────────────────────────

// runClangTidy runs clang-tidy on a C/C++ file or directory.
// Falls back to cppcheck if clang-tidy is not available.
func runClangTidy(ctx context.Context, path string, fix bool) (*LintResult, error) {
	if _, err := exec.LookPath("clang-tidy"); err != nil {
		return runCppcheck(ctx, path, fix)
	}

	projectRoot, absPath, relPath, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	info, _ := os.Stat(absPath)
	isDir := info != nil && info.IsDir()

	var files []string
	if isDir {
		files, err = findFiles(absPath, []string{".c", ".cc", ".cpp", ".cxx"})
		if err != nil || len(files) == 0 {
			return &LintResult{Success: true, Linter: "clang-tidy", Output: "no C/C++ source files found"}, nil
		}
	} else {
		files = []string{relPath}
	}

	// Use compile_commands.json if present; otherwise pass --extra-arg to suppress errors.
	var baseArgs []string
	if _, err := os.Stat(filepath.Join(projectRoot, "compile_commands.json")); err == nil {
		baseArgs = []string{"-p", "."}
	} else {
		baseArgs = []string{"--extra-arg=-w"}
	}
	if fix {
		baseArgs = append(baseArgs, "--fix", "--fix-errors")
	}

	var allErrors []LintError
	var allOutput strings.Builder

	for _, f := range files {
		args := append(baseArgs, f) //nolint:gocritic
		cmd := exec.CommandContext(ctx, "clang-tidy", args...)
		cmd.Dir = projectRoot
		out, _ := cmd.CombinedOutput()
		allOutput.Write(out)
		allErrors = append(allErrors, parseGCCStyleOutput(string(out))...)
	}

	return &LintResult{
		Success: len(allErrors) == 0,
		Linter:  "clang-tidy",
		Output:  allOutput.String(),
		Errors:  allErrors,
		Fixed:   fix && len(allErrors) == 0,
	}, nil
}

// runCppcheck runs cppcheck on a C/C++ file or directory.
func runCppcheck(ctx context.Context, path string, _ bool) (*LintResult, error) {
	if _, err := exec.LookPath("cppcheck"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "cppcheck",
			Output:  "no C/C++ linter found. Install clang-tidy (brew install llvm) or cppcheck (brew install cppcheck)",
			Errors:  []LintError{{Message: "clang-tidy and cppcheck not found in PATH", Severity: "error"}},
		}, nil
	}

	projectRoot, absPath, relPath, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	info, _ := os.Stat(absPath)
	isDir := info != nil && info.IsDir()

	args := []string{"--enable=all", "--suppress=missingInclude", "--template=gcc", "--quiet"}
	if isDir {
		args = append(args, "--recursive", relPath)
	} else {
		args = append(args, relPath)
	}

	cmd := exec.CommandContext(ctx, "cppcheck", args...)
	cmd.Dir = projectRoot
	out, _ := cmd.CombinedOutput()
	outputStr := string(out)

	lintErrors := parseGCCStyleOutput(outputStr)

	return &LintResult{
		Success: len(lintErrors) == 0,
		Linter:  "cppcheck",
		Output:  outputStr,
		Errors:  lintErrors,
	}, nil
}

// runClangFormat runs clang-format on C/C++ files (format check or fix).
func runClangFormat(ctx context.Context, path string, fix bool) (*LintResult, error) {
	if _, err := exec.LookPath("clang-format"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "clang-format",
			Output:  "clang-format not found. Install with: brew install clang-format",
			Errors:  []LintError{{Message: "clang-format not found in PATH", Severity: "error"}},
		}, nil
	}

	projectRoot, absPath, _, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	info, _ := os.Stat(absPath)
	isDir := info != nil && info.IsDir()

	var files []string
	if isDir {
		files, err = findFiles(absPath, []string{".c", ".cc", ".cpp", ".cxx", ".h", ".hpp", ".hxx"})
		if err != nil || len(files) == 0 {
			return &LintResult{Success: true, Linter: "clang-format", Output: "no C/C++ files found"}, nil
		}
	} else {
		rel, _ := filepath.Rel(projectRoot, absPath)
		files = []string{rel}
	}

	var allErrors []LintError
	var allOutput strings.Builder

	for _, f := range files {
		var args []string
		if fix {
			args = []string{"-i", f}
		} else {
			args = []string{"--dry-run", "--Werror", f}
		}
		cmd := exec.CommandContext(ctx, "clang-format", args...)
		cmd.Dir = projectRoot
		out, runErr := cmd.CombinedOutput()
		allOutput.Write(out)
		if runErr != nil {
			allErrors = append(allErrors, LintError{File: f, Message: "formatting issue", Severity: "warning"})
		}
	}

	return &LintResult{
		Success: len(allErrors) == 0,
		Linter:  "clang-format",
		Output:  allOutput.String(),
		Errors:  allErrors,
		Fixed:   fix && len(allErrors) == 0,
	}, nil
}

// ── Python ───────────────────────────────────────────────────────────────────

// runRuff runs ruff (Python linter/formatter). Falls back to flake8, then pylint.
func runRuff(ctx context.Context, path string, fix bool) (*LintResult, error) {
	if _, err := exec.LookPath("ruff"); err != nil {
		if _, err2 := exec.LookPath("flake8"); err2 != nil {
			return runPylint(ctx, path)
		}
		return runFlake8(ctx, path)
	}

	projectRoot, absPath, relPath, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	info, _ := os.Stat(absPath)
	isDir := info != nil && info.IsDir()

	target := relPath
	if isDir && target == "" {
		target = "."
	}

	args := []string{"check", "--output-format=json"}
	if fix {
		args = append(args, "--fix")
	}
	args = append(args, target)

	cmd := exec.CommandContext(ctx, "ruff", args...)
	cmd.Dir = projectRoot
	out, _ := cmd.CombinedOutput()
	outputStr := string(out)

	// Parse ruff JSON output:
	// [{"code":"E501","message":"...","filename":"...","location":{"row":1,"column":1},"fix":null}]
	type ruffLocation struct {
		Row    int `json:"row"`
		Column int `json:"column"`
	}
	type ruffIssue struct {
		Code     string       `json:"code"`
		Message  string       `json:"message"`
		Filename string       `json:"filename"`
		Location ruffLocation `json:"location"`
	}

	var issues []ruffIssue
	var lintErrors []LintError

	if err := json.Unmarshal(out, &issues); err == nil {
		for _, issue := range issues {
			lintErrors = append(lintErrors, LintError{
				File:     issue.Filename,
				Line:     issue.Location.Row,
				Column:   issue.Location.Column,
				Message:  issue.Message,
				Rule:     issue.Code,
				Severity: "warning",
			})
		}
	} else {
		// Fallback: parse text output (file:line:col: code message)
		lintErrors = parsePythonTextOutput(outputStr)
	}

	return &LintResult{
		Success: len(lintErrors) == 0,
		Linter:  "ruff",
		Output:  outputStr,
		Errors:  lintErrors,
		Fixed:   fix && len(lintErrors) == 0,
	}, nil
}

// runFlake8 runs flake8 (Python linter, fallback).
func runFlake8(ctx context.Context, path string) (*LintResult, error) {
	if _, err := exec.LookPath("flake8"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "flake8",
			Output:  "flake8 not found. Install with: pip install flake8",
			Errors:  []LintError{{Message: "flake8 not found in PATH", Severity: "error"}},
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

	cmd := exec.CommandContext(ctx, "flake8", "--format=default", target)
	cmd.Dir = projectRoot
	out, _ := cmd.CombinedOutput()
	outputStr := string(out)

	lintErrors := parsePythonTextOutput(outputStr)

	return &LintResult{
		Success: len(lintErrors) == 0,
		Linter:  "flake8",
		Output:  outputStr,
		Errors:  lintErrors,
	}, nil
}

// runPylint runs pylint (Python linter, last fallback).
func runPylint(ctx context.Context, path string) (*LintResult, error) {
	if _, err := exec.LookPath("pylint"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "pylint",
			Output:  "no Python linter found. Install ruff (pip install ruff), flake8 (pip install flake8), or pylint (pip install pylint)",
			Errors:  []LintError{{Message: "ruff, flake8, and pylint not found in PATH", Severity: "error"}},
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

	// JSON output format.
	cmd := exec.CommandContext(ctx, "pylint", "--output-format=json", target)
	cmd.Dir = projectRoot
	out, _ := cmd.CombinedOutput()
	outputStr := string(out)

	// Parse pylint JSON:
	// [{"type":"error","module":"...","obj":"","line":N,"column":M,"path":"...","symbol":"...","message":"...","message-id":"E0001"}]
	type pylintIssue struct {
		Type      string `json:"type"`
		Line      int    `json:"line"`
		Column    int    `json:"column"`
		Path      string `json:"path"`
		Symbol    string `json:"symbol"`
		Message   string `json:"message"`
		MessageID string `json:"message-id"`
	}

	var issues []pylintIssue
	var lintErrors []LintError

	if err := json.Unmarshal(out, &issues); err == nil {
		for _, issue := range issues {
			severity := "warning"
			if issue.Type == "error" || issue.Type == "fatal" {
				severity = "error"
			} else if issue.Type == "convention" || issue.Type == "refactor" {
				severity = "info"
			}
			lintErrors = append(lintErrors, LintError{
				File:     issue.Path,
				Line:     issue.Line,
				Column:   issue.Column,
				Message:  issue.Message,
				Rule:     issue.MessageID,
				Severity: severity,
			})
		}
	} else {
		lintErrors = parsePythonTextOutput(outputStr)
	}

	return &LintResult{
		Success: len(lintErrors) == 0,
		Linter:  "pylint",
		Output:  outputStr,
		Errors:  lintErrors,
	}, nil
}

// ── Rust ─────────────────────────────────────────────────────────────────────

// runCargoClippy runs cargo clippy (Rust linter).
func runCargoClippy(ctx context.Context, path string, fix bool) (*LintResult, error) {
	if _, err := exec.LookPath("cargo"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "cargo-clippy",
			Output:  "cargo not found. Install Rust from https://rustup.rs/",
			Errors:  []LintError{{Message: "cargo not found in PATH", Severity: "error"}},
		}, nil
	}

	projectRoot, _, _, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	// Prefer the Cargo.toml directory if path points into a Rust project.
	rustRoot := findRustRoot(projectRoot)

	args := []string{"clippy", "--message-format=json", "--quiet"}
	if fix {
		args = append(args, "--fix", "--allow-dirty", "--allow-staged")
	}
	args = append(args, "--", "-D", "warnings")

	cmd := exec.CommandContext(ctx, "cargo", args...)
	cmd.Dir = rustRoot
	out, _ := cmd.CombinedOutput()
	outputStr := string(out)

	lintErrors := parseCargoJSON(outputStr)

	return &LintResult{
		Success: len(lintErrors) == 0,
		Linter:  "cargo-clippy",
		Output:  outputStr,
		Errors:  lintErrors,
		Fixed:   fix && len(lintErrors) == 0,
	}, nil
}

// runRustfmt runs rustfmt (Rust formatter).
func runRustfmt(ctx context.Context, path string, fix bool) (*LintResult, error) {
	if _, err := exec.LookPath("rustfmt"); err != nil {
		return &LintResult{
			Success: false,
			Linter:  "rustfmt",
			Output:  "rustfmt not found. Install with: rustup component add rustfmt",
			Errors:  []LintError{{Message: "rustfmt not found in PATH", Severity: "error"}},
		}, nil
	}

	projectRoot, absPath, _, err := resolveLintPaths(path)
	if err != nil {
		return nil, err
	}

	info, _ := os.Stat(absPath)
	isDir := info != nil && info.IsDir()

	var files []string
	if isDir {
		files, err = findFiles(absPath, []string{".rs"})
		if err != nil || len(files) == 0 {
			return &LintResult{Success: true, Linter: "rustfmt", Output: "no Rust files found"}, nil
		}
	} else {
		rel, _ := filepath.Rel(projectRoot, absPath)
		files = []string{rel}
	}

	var allErrors []LintError
	var allOutput strings.Builder

	for _, f := range files {
		args := []string{}
		if !fix {
			args = append(args, "--check")
		}
		args = append(args, f)
		cmd := exec.CommandContext(ctx, "rustfmt", args...)
		cmd.Dir = projectRoot
		out, runErr := cmd.CombinedOutput()
		allOutput.Write(out)
		if runErr != nil {
			allErrors = append(allErrors, LintError{File: f, Message: "formatting differs from rustfmt output", Severity: "warning"})
		}
	}

	return &LintResult{
		Success: len(allErrors) == 0,
		Linter:  "rustfmt",
		Output:  allOutput.String(),
		Errors:  allErrors,
		Fixed:   fix && len(allErrors) == 0,
	}, nil
}

// ── Shared helpers ────────────────────────────────────────────────────────────

// resolveLintPaths resolves project root, absolute path, and relative path for a lint target.
// Uses projectroot.Find() which checks PROJECT_ROOT env then walks up for markers.
func resolveLintPaths(path string) (projectRoot, absPath, relPath string, err error) {
	projectRoot, err = projectroot.Find()
	if err != nil {
		projectRoot, err = projectroot.FindFrom(path)
		if err != nil {
			projectRoot, _ = os.Getwd()
		}
	}

	absPath, relPath, err = security.ValidatePathWithinRoot(path, projectRoot)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid path: %w", err)
	}
	return projectRoot, absPath, relPath, nil
}

// findFiles walks dir recursively and returns relative paths for files matching exts.
func findFiles(dir string, exts []string) ([]string, error) {
	extSet := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		extSet[e] = struct{}{}
	}

	var files []string
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if _, ok := extSet[strings.ToLower(filepath.Ext(p))]; ok {
			rel, relErr := filepath.Rel(dir, p)
			if relErr == nil {
				files = append(files, rel)
			}
		}
		return nil
	})
	return files, err
}

// findRustRoot walks up from dir looking for Cargo.toml.
func findRustRoot(dir string) string {
	cur := dir
	for {
		if _, err := os.Stat(filepath.Join(cur, "Cargo.toml")); err == nil {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return dir
		}
		cur = parent
	}
}

// parseGCCStyleOutput parses GCC-compatible output (file:line:col: severity: message).
// Used by clang-tidy and cppcheck (--template=gcc).
func parseGCCStyleOutput(output string) []LintError {
	var errors []LintError
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "note:") {
			continue
		}
		// Format: file.cpp:10:5: warning: some message [rule]
		parts := strings.SplitN(line, ":", 5)
		if len(parts) < 4 {
			errors = append(errors, LintError{Message: line, Severity: "warning"})
			continue
		}
		file := parts[0]
		lineNum, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		colNum, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
		severityAndMsg := strings.TrimSpace(parts[3])
		msg := ""
		if len(parts) >= 5 {
			msg = strings.TrimSpace(parts[4])
		}

		severity := "warning"
		if strings.HasPrefix(severityAndMsg, "error") {
			severity = "error"
		} else if strings.HasPrefix(severityAndMsg, "note") || strings.HasPrefix(severityAndMsg, "info") {
			severity = "info"
		}

		// Extract rule from trailing [...] in message
		rule := ""
		if idx := strings.LastIndex(msg, "["); idx >= 0 && strings.HasSuffix(msg, "]") {
			rule = msg[idx+1 : len(msg)-1]
			msg = strings.TrimSpace(msg[:idx])
		}
		if msg == "" {
			msg = severityAndMsg
		}

		errors = append(errors, LintError{
			File:     file,
			Line:     lineNum,
			Column:   colNum,
			Message:  msg,
			Rule:     rule,
			Severity: severity,
		})
	}
	return errors
}

// parsePythonTextOutput parses flake8/ruff text output: file:line:col: CODE message
func parsePythonTextOutput(output string) []LintError {
	var errors []LintError
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// file.py:10:5: E501 line too long
		parts := strings.SplitN(line, ":", 4)
		if len(parts) < 4 {
			errors = append(errors, LintError{Message: line, Severity: "warning"})
			continue
		}
		file := parts[0]
		lineNum, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		colNum, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
		rest := strings.TrimSpace(parts[3])

		rule := ""
		msg := rest
		if fields := strings.Fields(rest); len(fields) > 1 {
			if len(fields[0]) > 0 && (fields[0][0] >= 'A' && fields[0][0] <= 'Z') {
				rule = fields[0]
				msg = strings.Join(fields[1:], " ")
			}
		}

		errors = append(errors, LintError{
			File:     file,
			Line:     lineNum,
			Column:   colNum,
			Message:  msg,
			Rule:     rule,
			Severity: "warning",
		})
	}
	return errors
}

// parseCargoJSON parses cargo clippy --message-format=json output (one JSON object per line).
func parseCargoJSON(output string) []LintError {
	var errors []LintError
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var msg struct {
			Reason  string `json:"reason"`
			Message *struct {
				Code    *struct{ Code string `json:"code"` } `json:"code"`
				Level   string                               `json:"level"`
				Message string                               `json:"message"`
				Spans   []struct {
					FileName    string `json:"file_name"`
					LineStart   int    `json:"line_start"`
					ColumnStart int    `json:"column_start"`
				} `json:"spans"`
			} `json:"message"`
		}

		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if msg.Reason != "compiler-message" || msg.Message == nil {
			continue
		}
		if msg.Message.Level == "note" || msg.Message.Level == "help" {
			continue
		}

		severity := msg.Message.Level
		if severity == "" {
			severity = "warning"
		}

		rule := ""
		if msg.Message.Code != nil {
			rule = msg.Message.Code.Code
		}

		file, lineNum, colNum := "", 0, 0
		if len(msg.Message.Spans) > 0 {
			s := msg.Message.Spans[0]
			file = s.FileName
			lineNum = s.LineStart
			colNum = s.ColumnStart
		}

		errors = append(errors, LintError{
			File:     file,
			Line:     lineNum,
			Column:   colNum,
			Message:  msg.Message.Message,
			Rule:     rule,
			Severity: severity,
		})
	}
	return errors
}
