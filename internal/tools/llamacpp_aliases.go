// llamacpp_aliases.go — Model alias resolution for llama.cpp.
//
// Two sources of aliases are merged at resolution time:
//  1. Built-in aliases (hardcoded, Ollama shorthand → canonical Ollama model name).
//  2. User aliases loaded from .exarp/llamacpp.json (optional, overrides built-ins).
//
// A resolved alias value is either:
//   - An Ollama model name (e.g. "llama3.2:latest") — passed to ResolveOllamaModelPath.
//   - An absolute or home-relative file path (starts with "/" or "~/") — used directly.
//
// Example .exarp/llamacpp.json:
//
//	{
//	  "aliases": {
//	    "my-llama": "llama3.2:latest",
//	    "local-7b": "/path/to/my-model.gguf",
//	    "deepseek": "deepseek-r1:7b"
//	  }
//	}
package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// builtinModelAliases are the hardcoded Ollama shorthand aliases shipped with exarp-go.
// Keys are shorthand names; values are canonical Ollama model names.
var builtinModelAliases = map[string]string{
	// llama family
	"llama3":  "llama3.2",
	"llama":   "llama3.2",
	"llama3b": "llama3.2:3b",
	"llama1b": "llama3.2:1b",
	// codellama
	"codellama": "codellama",
	"code":      "codellama",
	// phi family
	"phi":   "phi3",
	"phi3":  "phi3",
	"phi4":  "phi4",
	"phi35": "phi3.5",
	// mistral family
	"mistral":   "mistral",
	"mistral7b": "mistral:7b",
	// gemma family
	"gemma":  "gemma2",
	"gemma2": "gemma2",
	"gemma3": "gemma3",
	// qwen family
	"qwen":        "qwen2.5",
	"qwen2":       "qwen2.5",
	"qwen-coder":  "qwen2.5-coder",
	"qwen2-coder": "qwen2.5-coder",
	// deepseek family
	"deepseek":    "deepseek-r1",
	"deepseek-r1": "deepseek-r1",
	"deepseek-v3": "deepseek-v3",
	// other
	"smollm":    "smollm2",
	"tinyllama": "tinyllama",
	"vicuna":    "vicuna",
	"orca":      "orca-mini",
	"starcoder": "starcoder2",
}

// LlamaCppAliasConfig holds user-defined alias overrides and preload/warmup from .exarp/llamacpp.json.
type LlamaCppAliasConfig struct {
	// Aliases maps shorthand names to Ollama model names or absolute file paths.
	Aliases map[string]string `json:"aliases"`
	// Preload is a list of model names or paths to load on first use (when llamacpp is invoked).
	Preload []string `json:"preload"`
	// Warmup, when true, runs a short inference after each preload to keep the model hot.
	Warmup bool `json:"warmup"`
}

// llamacppConfigPath returns the path to .exarp/llamacpp.json.
func llamacppConfigPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".exarp", "llamacpp.json")
}

// loadLlamaCppAliasConfig reads user aliases from .exarp/llamacpp.json.
// Returns an empty config (not an error) if the file is absent or unreadable.
func loadLlamaCppAliasConfig(projectRoot string) LlamaCppAliasConfig {
	data, err := os.ReadFile(llamacppConfigPath(projectRoot))
	if err != nil {
		return LlamaCppAliasConfig{}
	}
	var cfg LlamaCppAliasConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return LlamaCppAliasConfig{}
	}
	return cfg
}

// mergedAliases returns built-in aliases merged with user overrides.
// User entries take precedence over built-ins.
func mergedAliases(projectRoot string) map[string]string {
	merged := make(map[string]string, len(builtinModelAliases))
	for k, v := range builtinModelAliases {
		merged[k] = v
	}
	user := loadLlamaCppAliasConfig(projectRoot)
	for k, v := range user.Aliases {
		merged[k] = v
	}
	return merged
}

// isFilePath returns true if v looks like an absolute or home-relative file path.
func isFilePath(v string) bool {
	return strings.HasPrefix(v, "/") || strings.HasPrefix(v, "~/") || strings.HasPrefix(v, "~\\")
}

// expandHome replaces a leading "~/" with the user's home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") && !strings.HasPrefix(path, "~\\") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}

// ResolveModelAlias resolves a model shorthand to its canonical Ollama model name.
// Returns the input unchanged if no alias exists.
// Uses only built-in aliases (no project root needed). For full resolution including
// user aliases and path aliases, use ResolveModelToPath.
func ResolveModelAlias(name string) string {
	base, tag := parseModelName(name)
	if canonical, ok := builtinModelAliases[base]; ok {
		if tag == "latest" {
			return canonical
		}
		return canonical + ":" + tag
	}
	return name
}

// ResolveModelToPath resolves a model name or alias to a GGUF file path.
//
// Resolution order:
//  1. Merge built-in + user aliases; look up name in merged map.
//  2. If resolved value is a file path (starts with "/" or "~/"), expand and return it.
//  3. Otherwise treat resolved value as an Ollama model name and call ResolveOllamaModelPath.
//  4. If no alias found, try ResolveOllamaModelPath directly on the input.
func ResolveModelToPath(name, projectRoot string) (string, error) {
	aliases := mergedAliases(projectRoot)

	base, tag := parseModelName(name)
	resolved := name
	if canonical, ok := aliases[base]; ok {
		if tag == "latest" {
			resolved = canonical
		} else {
			resolved = canonical + ":" + tag
		}
	}

	if isFilePath(resolved) {
		return expandHome(resolved), nil
	}

	return ResolveOllamaModelPath(resolved)
}

// ListAllAliases returns the full merged alias map (built-in + user) for display.
func ListAllAliases(projectRoot string) map[string]string {
	return mergedAliases(projectRoot)
}
