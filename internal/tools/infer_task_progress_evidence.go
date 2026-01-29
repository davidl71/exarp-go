package tools

import (
	"os"
	"path/filepath"
	"strings"
)

// CodebaseEvidence holds paths and content snippets gathered from the project root
// for use in heuristic task-completion inference.
type CodebaseEvidence struct {
	// Paths is the list of relative file paths (under project root) that were scanned.
	Paths []string
	// Snippets maps relative path to the first few lines of content (for matching).
	Snippets map[string]string
}

const (
	// defaultSnippetLines is the number of lines to read per file for evidence.
	defaultSnippetLines = 10
	// maxFileBytesForSnippet caps file read size when collecting snippets.
	maxFileBytesForSnippet = 64 * 1024
)

// skipDir names that should not be walked (same as task_discovery).
var skipDirNames = map[string]bool{
	".git": true, "node_modules": true, "__pycache__": true, ".venv": true,
	"vendor": true, ".idea": true, ".vscode": true, "dist": true, "build": true, "target": true,
}

// GatherEvidence walks the project root up to scanDepth (1-5), restricted by file extensions,
// and collects relative paths and small content snippets for task-completion inference.
func GatherEvidence(projectRoot string, scanDepth int, extensions []string) (*CodebaseEvidence, error) {
	if scanDepth < 1 {
		scanDepth = 1
	}
	if scanDepth > 5 {
		scanDepth = 5
	}
	extSet := make(map[string]bool)
	for _, e := range extensions {
		e = strings.TrimSpace(strings.ToLower(e))
		if e != "" && !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		if e != "" {
			extSet[e] = true
		}
	}
	if len(extSet) == 0 {
		extSet[".go"] = true
		extSet[".py"] = true
		extSet[".ts"] = true
		extSet[".tsx"] = true
		extSet[".js"] = true
		extSet[".jsx"] = true
		extSet[".java"] = true
		extSet[".cs"] = true
		extSet[".rs"] = true
	}

	rootAbs, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, err
	}

	evidence := &CodebaseEvidence{
		Paths:   make([]string, 0),
		Snippets: make(map[string]string),
	}

	err = filepath.Walk(rootAbs, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if skipDirNames[base] {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(rootAbs, path)
			if rel == "." {
				return nil
			}
			segments := strings.Count(rel, string(filepath.Separator)) + 1
			if segments >= scanDepth {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !extSet[ext] {
			return nil
		}

		rel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return nil
		}
		segments := strings.Count(rel, string(filepath.Separator)) + 1
		if segments > scanDepth {
			return nil
		}

		evidence.Paths = append(evidence.Paths, rel)
		snippet := readSnippet(path)
		if snippet != "" {
			evidence.Snippets[rel] = snippet
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return evidence, nil
}

func readSnippet(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.Size() > maxFileBytesForSnippet {
		return ""
	}
	buf := make([]byte, info.Size())
	n, _ := f.Read(buf)
	if n == 0 {
		return ""
	}
	lines := strings.Split(string(buf[:n]), "\n")
	if len(lines) > defaultSnippetLines {
		lines = lines[:defaultSnippetLines]
	}
	return strings.TrimSpace(strings.Join(lines, " "))
}
