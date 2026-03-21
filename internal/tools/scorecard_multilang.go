// scorecard_multilang.go — C++, Python, Rust detection and health checks for multi-language scorecard.
package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// LangHealth holds per-language health check results and score.
type LangHealth struct {
	Lang            string   `json:"lang"`                // "cpp", "python", "rust", "go", "typescript", "swift"
	LangRoot        string   `json:"lang_root,omitempty"` // subdir where detected, e.g. "native", "agents/go"
	Detected        bool     `json:"detected"`
	BuildPasses     bool     `json:"build_passes"`
	TestPasses      bool     `json:"test_passes"`
	LintPasses      bool     `json:"lint_passes"`
	FmtPasses       bool     `json:"fmt_passes"`
	Score           float64  `json:"score"`
	Recommendations []string `json:"recommendations,omitempty"`
	FileCount       int      `json:"file_count"`
}

// defaultLangDirs are subdirs (relative to project root) to search for each language. "" = repo root.
var defaultLangDirs = []string{"", "native", "python", "agents", "agents/go", "agents/backend", "web", "cursor-extension", "ios", "desktop", "ondevice"}

const multilangTimeout = 60 * time.Second

// findLangRoot returns the first subdir in dirs where hasMarker(dir) and hasSource(dir) are true.
// dir is the absolute path filepath.Join(projectRoot, subdir).
func findLangRoot(projectRoot string, dirs []string, hasMarker, hasSource func(dir string) bool) (bool, string) {
	for _, sub := range dirs {
		dir := projectRoot
		if sub != "" {
			dir = filepath.Join(projectRoot, sub)
		}
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			continue
		}
		if hasMarker(dir) && hasSource(dir) {
			return true, sub
		}
	}
	return false, ""
}

// DetectCppProject returns true if the project has C++ (CMakeLists.txt or Makefile + .cpp/.h) in root or common subdirs.
func DetectCppProject(projectRoot string) bool {
	found, _ := DetectCppProjectRoot(projectRoot)
	return found
}

// DetectCppProjectRoot returns true and the subdir (e.g. "native") where C++ was detected.
func DetectCppProjectRoot(projectRoot string) (bool, string) {
	hasMarker := func(dir string) bool {
		_, errCMake := os.Stat(filepath.Join(dir, "CMakeLists.txt"))
		_, errMake := os.Stat(filepath.Join(dir, "Makefile"))
		return errCMake == nil || errMake == nil
	}
	hasSource := func(dir string) bool { return countCppFilesIn(dir) > 0 }
	return findLangRoot(projectRoot, defaultLangDirs, hasMarker, hasSource)
}

// DetectPythonProject returns true if the project has Python in root or common subdirs.
func DetectPythonProject(projectRoot string) bool {
	found, _ := DetectPythonProjectRoot(projectRoot)
	return found
}

// DetectPythonProjectRoot returns true and the subdir where Python was detected.
func DetectPythonProjectRoot(projectRoot string) (bool, string) {
	hasMarker := func(dir string) bool {
		_, p := os.Stat(filepath.Join(dir, "pyproject.toml"))
		_, r := os.Stat(filepath.Join(dir, "requirements.txt"))
		_, s := os.Stat(filepath.Join(dir, "setup.py"))
		return p == nil || r == nil || s == nil
	}
	hasSource := func(dir string) bool { return countPythonFilesIn(dir) > 0 }
	return findLangRoot(projectRoot, defaultLangDirs, hasMarker, hasSource)
}

// DetectRustProject returns true if the project has Rust (Cargo.toml + .rs) in root or common subdirs.
func DetectRustProject(projectRoot string) bool {
	found, _ := DetectRustProjectRoot(projectRoot)
	return found
}

// DetectRustProjectRoot returns true and the subdir where Rust was detected.
func DetectRustProjectRoot(projectRoot string) (bool, string) {
	hasMarker := func(dir string) bool {
		_, err := os.Stat(filepath.Join(dir, "Cargo.toml"))
		return err == nil
	}
	hasSource := func(dir string) bool { return countRustFilesIn(dir) > 0 }
	return findLangRoot(projectRoot, defaultLangDirs, hasMarker, hasSource)
}

// DetectGoProject returns true if the project has Go (go.mod + .go) in root or common subdirs.
func DetectGoProject(projectRoot string) bool {
	found, _ := DetectGoProjectRoot(projectRoot)
	return found
}

// DetectGoProjectRoot returns true and the subdir where Go was detected (e.g. "agents/go").
func DetectGoProjectRoot(projectRoot string) (bool, string) {
	hasMarker := func(dir string) bool {
		_, err := os.Stat(filepath.Join(dir, "go.mod"))
		return err == nil
	}
	hasSource := func(dir string) bool { return countGoFilesIn(dir) > 0 }
	return findLangRoot(projectRoot, defaultLangDirs, hasMarker, hasSource)
}

// DetectTypeScriptProject returns true if the project has TypeScript/JS (package.json + .ts/.tsx/.js) in root or common subdirs.
func DetectTypeScriptProject(projectRoot string) bool {
	found, _ := DetectTypeScriptProjectRoot(projectRoot)
	return found
}

// DetectTypeScriptProjectRoot returns true and the subdir where TypeScript was detected.
func DetectTypeScriptProjectRoot(projectRoot string) (bool, string) {
	hasMarker := func(dir string) bool {
		_, err := os.Stat(filepath.Join(dir, "package.json"))
		return err == nil
	}
	hasSource := func(dir string) bool { return countTypeScriptFilesIn(dir) > 0 }
	return findLangRoot(projectRoot, defaultLangDirs, hasMarker, hasSource)
}

// DetectSwiftProject returns true if the project has Swift (Package.swift or .xcodeproj + .swift) in root or common subdirs.
func DetectSwiftProject(projectRoot string) bool {
	found, _ := DetectSwiftProjectRoot(projectRoot)
	return found
}

// DetectSwiftProjectRoot returns true and the subdir where Swift was detected.
func DetectSwiftProjectRoot(projectRoot string) (bool, string) {
	hasMarker := func(dir string) bool {
		_, pkg := os.Stat(filepath.Join(dir, "Package.swift"))
		if pkg == nil {
			return true
		}
		// Check for .xcodeproj in this dir
		entries, err := os.ReadDir(dir)
		if err != nil {
			return false
		}
		for _, e := range entries {
			if e.IsDir() && strings.HasSuffix(e.Name(), ".xcodeproj") {
				return true
			}
		}
		return false
	}
	hasSource := func(dir string) bool { return countSwiftFilesIn(dir) > 0 }
	return findLangRoot(projectRoot, defaultLangDirs, hasMarker, hasSource)
}

func isSkipDir(path, projectRoot string) bool {
	rel, err := filepath.Rel(projectRoot, path)
	if err != nil {
		return false
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for _, p := range parts {
		switch p {
		case ".git", "vendor", "node_modules", "build", "third_party", "__pycache__", "target", ".venv", "venv":
			return true
		}
		if strings.HasPrefix(p, ".") && p != "." && p != ".." {
			return true
		}
	}
	return false
}

func walkSkipDir(path string, info os.FileInfo, projectRoot string) error {
	if info.IsDir() && isSkipDir(path, projectRoot) {
		return filepath.SkipDir
	}
	return nil
}

// countCppFilesIn counts C/C++ files under dir (absolute path). Used for detection and per-dir file count.
func countCppFilesIn(dir string) int {
	var n int
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if err := walkSkipDir(path, info, dir); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".cpp", ".cc", ".cxx", ".c", ".h", ".hpp", ".hxx":
			n++
		}
		return nil
	})
	return n
}

func countPythonFilesIn(dir string) int {
	var n int
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if err := walkSkipDir(path, info, dir); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(path), ".py") {
			n++
		}
		return nil
	})
	return n
}

func countRustFilesIn(dir string) int {
	var n int
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if err := walkSkipDir(path, info, dir); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(path), ".rs") {
			n++
		}
		return nil
	})
	return n
}

func countGoFilesIn(dir string) int {
	var n int
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if err := walkSkipDir(path, info, dir); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(path), ".go") {
			n++
		}
		return nil
	})
	return n
}

func countTypeScriptFilesIn(dir string) int {
	var n int
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if err := walkSkipDir(path, info, dir); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".ts", ".tsx", ".js", ".jsx":
			n++
		}
		return nil
	})
	return n
}

func countSwiftFilesIn(dir string) int {
	var n int
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if err := walkSkipDir(path, info, dir); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(path), ".swift") {
			n++
		}
		return nil
	})
	return n
}

func workDir(projectRoot, langRoot string) string {
	if langRoot == "" {
		return projectRoot
	}
	return filepath.Join(projectRoot, langRoot)
}

func runInDir(ctx context.Context, projectRoot string, name string, args ...string) (passed bool) {
	ctx, cancel := context.WithTimeout(ctx, multilangTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = projectRoot
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil
}

// CollectCppHealth runs C++ build/test from the given language root and returns health + score.
func CollectCppHealth(ctx context.Context, projectRoot, langRoot string) LangHealth {
	dir := workDir(projectRoot, langRoot)
	h := LangHealth{Lang: "cpp", Detected: true, LangRoot: langRoot}
	h.FileCount = countCppFilesIn(dir)

	// Prefer CMake; fallback to Make
	hasCMake, _ := os.Stat(filepath.Join(dir, "CMakeLists.txt"))
	hasMake, _ := os.Stat(filepath.Join(dir, "Makefile"))

	if hasCMake == nil {
		buildDir := filepath.Join(dir, "build")
		_ = os.MkdirAll(buildDir, 0755)
		h.BuildPasses = runInDir(ctx, dir, "cmake", "-S", ".", "-B", "build", "-G", "Ninja") ||
			runInDir(ctx, dir, "cmake", "-S", ".", "-B", "build")
		if h.BuildPasses {
			h.BuildPasses = runInDir(ctx, dir, "cmake", "--build", "build")
		}
		if h.BuildPasses {
			h.TestPasses = runInDir(ctx, dir, "ctest", "--test-dir", "build", "--output-on-failure")
		}
		if !h.BuildPasses {
			h.Recommendations = append(h.Recommendations, "Fix C++ build (cmake --build build)")
		}
		if !h.TestPasses && h.BuildPasses {
			h.Recommendations = append(h.Recommendations, "Fix C++ tests (ctest --test-dir build)")
		}
	} else if hasMake == nil {
		h.BuildPasses = runInDir(ctx, dir, "make", "-j")
		if h.BuildPasses {
			h.TestPasses = runInDir(ctx, dir, "make", "test") || runInDir(ctx, dir, "ctest", "--output-on-failure")
		}
		if !h.BuildPasses {
			h.Recommendations = append(h.Recommendations, "Fix C++ build (make)")
		}
		if !h.TestPasses && h.BuildPasses {
			h.Recommendations = append(h.Recommendations, "Fix C++ tests (make test or ctest)")
		}
	}

	// Optional: clang-format not run (would require gathering file list)
	h.Score = multilangScore(h.BuildPasses, h.TestPasses, h.LintPasses, h.FmtPasses)
	return h
}

// CollectPythonHealth runs Python install/test from the given language root.
func CollectPythonHealth(ctx context.Context, projectRoot, langRoot string) LangHealth {
	dir := workDir(projectRoot, langRoot)
	h := LangHealth{Lang: "python", Detected: true, LangRoot: langRoot}
	h.FileCount = countPythonFilesIn(dir)

	// Prefer uv, then pip
	h.BuildPasses = runInDir(ctx, dir, "uv", "sync") ||
		runInDir(ctx, dir, "pip", "install", "-e", ".") ||
		runInDir(ctx, dir, "pip", "install", "-r", "requirements.txt")
	// If no lockfile/req, "build" is N/A; assume pass if we have .py files
	if !h.BuildPasses && h.FileCount > 0 {
		h.BuildPasses = true
	}

	h.TestPasses = runInDir(ctx, dir, "uv", "run", "pytest", "--tb=no", "-q") ||
		runInDir(ctx, dir, "python", "-m", "pytest", "--tb=no", "-q") ||
		runInDir(ctx, dir, "pytest", "--tb=no", "-q")
	if !h.TestPasses && h.BuildPasses {
		h.Recommendations = append(h.Recommendations, "Fix Python tests (pytest)")
	}

	h.LintPasses = runInDir(ctx, dir, "uv", "run", "ruff", "check", ".") ||
		runInDir(ctx, dir, "ruff", "check", ".")
	if !h.LintPasses {
		if _, err := exec.LookPath("ruff"); err == nil {
			h.Recommendations = append(h.Recommendations, "Fix Python lint (ruff check)")
		}
	}

	h.Score = multilangScore(h.BuildPasses, h.TestPasses, h.LintPasses, h.FmtPasses)
	return h
}

// CollectRustHealth runs cargo build/test from the given language root.
func CollectRustHealth(ctx context.Context, projectRoot, langRoot string) LangHealth {
	dir := workDir(projectRoot, langRoot)
	h := LangHealth{Lang: "rust", Detected: true, LangRoot: langRoot}
	h.FileCount = countRustFilesIn(dir)

	h.BuildPasses = runInDir(ctx, dir, "cargo", "build", "--quiet")
	if !h.BuildPasses {
		h.Recommendations = append(h.Recommendations, "Fix Rust build (cargo build)")
	}

	if h.BuildPasses {
		h.TestPasses = runInDir(ctx, dir, "cargo", "test", "--quiet")
		if !h.TestPasses {
			h.Recommendations = append(h.Recommendations, "Fix Rust tests (cargo test)")
		}
	}

	h.LintPasses = runInDir(ctx, dir, "cargo", "clippy", "--quiet", "--", "-D", "warnings")
	if !h.LintPasses && h.BuildPasses {
		h.Recommendations = append(h.Recommendations, "Fix Rust lint (cargo clippy)")
	}

	h.FmtPasses = runInDir(ctx, dir, "cargo", "fmt", "--", "--check")
	if !h.FmtPasses && h.BuildPasses {
		h.Recommendations = append(h.Recommendations, "Run cargo fmt")
	}

	h.Score = multilangScore(h.BuildPasses, h.TestPasses, h.LintPasses, h.FmtPasses)
	return h
}

// CollectGoHealth runs go build/test from the given language root (e.g. agents/go).
func CollectGoHealth(ctx context.Context, projectRoot, langRoot string) LangHealth {
	dir := workDir(projectRoot, langRoot)
	h := LangHealth{Lang: "go", Detected: true, LangRoot: langRoot}
	h.FileCount = countGoFilesIn(dir)

	h.BuildPasses = runInDir(ctx, dir, "go", "build", "./...")
	if !h.BuildPasses {
		h.Recommendations = append(h.Recommendations, "Fix Go build (go build ./...)")
	}
	if h.BuildPasses {
		h.TestPasses = runInDir(ctx, dir, "go", "test", "./...")
		if !h.TestPasses {
			h.Recommendations = append(h.Recommendations, "Fix Go tests (go test ./...)")
		}
	}
	h.LintPasses = runInDir(ctx, dir, "go", "vet", "./...")
	h.FmtPasses = runInDir(ctx, dir, "gofmt", "-l", ".") // exit 0 = no files need formatting
	h.Score = multilangScore(h.BuildPasses, h.TestPasses, h.LintPasses, h.FmtPasses)
	return h
}

// CollectTypeScriptHealth runs npm/yarn install and build/test from the given language root.
func CollectTypeScriptHealth(ctx context.Context, projectRoot, langRoot string) LangHealth {
	dir := workDir(projectRoot, langRoot)
	h := LangHealth{Lang: "typescript", Detected: true, LangRoot: langRoot}
	h.FileCount = countTypeScriptFilesIn(dir)

	h.BuildPasses = runInDir(ctx, dir, "npm", "ci") ||
		runInDir(ctx, dir, "npm", "install") ||
		runInDir(ctx, dir, "yarn", "install")
	if !h.BuildPasses && h.FileCount > 0 {
		h.BuildPasses = true
	}
	if h.BuildPasses {
		h.BuildPasses = runInDir(ctx, dir, "npm", "run", "build") ||
			runInDir(ctx, dir, "npx", "tsc", "--noEmit")
		if !h.BuildPasses {
			h.Recommendations = append(h.Recommendations, "Fix TypeScript build (npm run build or tsc --noEmit)")
		}
	}
	h.TestPasses = runInDir(ctx, dir, "npm", "test", "--", "--passWithNoTests") ||
		runInDir(ctx, dir, "npm", "run", "test")
	h.LintPasses = runInDir(ctx, dir, "npm", "run", "lint") ||
		runInDir(ctx, dir, "npx", "eslint", ".", "--max-warnings", "0")
	h.Score = multilangScore(h.BuildPasses, h.TestPasses, h.LintPasses, h.FmtPasses)
	return h
}

// CollectSwiftHealth runs swift build/test from the given language root (e.g. ios, desktop).
func CollectSwiftHealth(ctx context.Context, projectRoot, langRoot string) LangHealth {
	dir := workDir(projectRoot, langRoot)
	h := LangHealth{Lang: "swift", Detected: true, LangRoot: langRoot}
	h.FileCount = countSwiftFilesIn(dir)

	// Prefer SwiftPM
	h.BuildPasses = runInDir(ctx, dir, "swift", "build")
	if !h.BuildPasses {
		// xcodebuild if .xcodeproj present
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if e.IsDir() && strings.HasSuffix(e.Name(), ".xcodeproj") {
				h.BuildPasses = runInDir(ctx, dir, "xcodebuild", "-scheme", strings.TrimSuffix(e.Name(), ".xcodeproj"), "-destination", "generic/platform=iOS", "build")
				break
			}
		}
	}
	if h.BuildPasses {
		h.TestPasses = runInDir(ctx, dir, "swift", "test")
	}
	if !h.BuildPasses {
		h.Recommendations = append(h.Recommendations, "Fix Swift build (swift build or xcodebuild)")
	}
	h.Score = multilangScore(h.BuildPasses, h.TestPasses, h.LintPasses, h.FmtPasses)
	return h
}

func multilangScore(build, test, lint, fmt bool) float64 {
	score := 0.0
	if build {
		score += 40
	}
	if test {
		score += 30
	}
	if lint {
		score += 15
	}
	if fmt {
		score += 15
	}
	return score
}

// CollectMultilangHealth runs detection and health for C++, Python, Rust, Go, TypeScript, Swift.
func CollectMultilangHealth(ctx context.Context, projectRoot string) []LangHealth {
	var out []LangHealth
	if found, root := DetectCppProjectRoot(projectRoot); found {
		out = append(out, CollectCppHealth(ctx, projectRoot, root))
	}
	if found, root := DetectPythonProjectRoot(projectRoot); found {
		out = append(out, CollectPythonHealth(ctx, projectRoot, root))
	}
	if found, root := DetectRustProjectRoot(projectRoot); found {
		out = append(out, CollectRustHealth(ctx, projectRoot, root))
	}
	if found, root := DetectGoProjectRoot(projectRoot); found {
		out = append(out, CollectGoHealth(ctx, projectRoot, root))
	}
	if found, root := DetectTypeScriptProjectRoot(projectRoot); found {
		out = append(out, CollectTypeScriptHealth(ctx, projectRoot, root))
	}
	if found, root := DetectSwiftProjectRoot(projectRoot); found {
		out = append(out, CollectSwiftHealth(ctx, projectRoot, root))
	}
	return out
}

// MultilangCacheSignature returns a stable string for cache key: detected languages, their roots, and file counts.
// When files are added/removed, the signature changes so the cache is invalidated.
func MultilangCacheSignature(projectRoot string) string {
	var parts []string
	if found, root := DetectCppProjectRoot(projectRoot); found {
		parts = append(parts, "cpp:"+root+":"+strconv.Itoa(countCppFilesIn(workDir(projectRoot, root))))
	}
	if found, root := DetectPythonProjectRoot(projectRoot); found {
		parts = append(parts, "py:"+root+":"+strconv.Itoa(countPythonFilesIn(workDir(projectRoot, root))))
	}
	if found, root := DetectRustProjectRoot(projectRoot); found {
		parts = append(parts, "rs:"+root+":"+strconv.Itoa(countRustFilesIn(workDir(projectRoot, root))))
	}
	if found, root := DetectGoProjectRoot(projectRoot); found {
		parts = append(parts, "go:"+root+":"+strconv.Itoa(countGoFilesIn(workDir(projectRoot, root))))
	}
	if found, root := DetectTypeScriptProjectRoot(projectRoot); found {
		parts = append(parts, "ts:"+root+":"+strconv.Itoa(countTypeScriptFilesIn(workDir(projectRoot, root))))
	}
	if found, root := DetectSwiftProjectRoot(projectRoot); found {
		parts = append(parts, "swift:"+root+":"+strconv.Itoa(countSwiftFilesIn(workDir(projectRoot, root))))
	}
	return strings.Join(parts, ",")
}

// FormatMultilangSections returns formatted text for the given language health results.
func FormatMultilangSections(langs []LangHealth) string {
	if len(langs) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, h := range langs {
		if !h.Detected {
			continue
		}
		title := strings.ToUpper(h.Lang)
		if h.LangRoot != "" {
			title += " (" + h.LangRoot + ")"
		}
		sb.WriteString(fmt.Sprintf("  ── %s ──\n", title))
		sb.WriteString(fmt.Sprintf("    Files:     %d\n", h.FileCount))
		sb.WriteString(fmt.Sprintf("    Score:     %.1f%%\n", h.Score))
		sb.WriteString(fmt.Sprintf("    Build:     %s\n", checkMark(h.BuildPasses)))
		sb.WriteString(fmt.Sprintf("    Tests:     %s\n", checkMark(h.TestPasses)))
		sb.WriteString(fmt.Sprintf("    Lint:      %s\n", checkMark(h.LintPasses)))
		sb.WriteString(fmt.Sprintf("    Format:    %s\n", checkMark(h.FmtPasses)))
		if len(h.Recommendations) > 0 {
			for _, r := range h.Recommendations {
				sb.WriteString(fmt.Sprintf("    • %s\n", r))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
