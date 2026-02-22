// llamacpp_ollama.go — Ollama model discovery for llama.cpp integration.
// Reads Ollama's local storage (~/.ollama/models or OLLAMA_MODELS env) to discover
// GGUF model blobs without the Ollama HTTP API. Used by llamacpp_provider to resolve
// model names to file paths for direct llama.cpp loading.
// See docs/OLLAMA_MANIFEST_PARSING.md for the storage layout and manifest format.
package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ollamaGGUFMediaType = "application/vnd.ollama.image.model"

// OllamaLocalModel represents a GGUF model discovered from Ollama's local storage.
// Distinct from OllamaModel (ollama_native.go) which represents an HTTP API response.
type OllamaLocalModel struct {
	Name     string // full name, e.g. "llama3.2:latest"
	Base     string // base model name, e.g. "llama3.2"
	Tag      string // tag, e.g. "latest", "1b"
	GGUFPath string // absolute path to the GGUF blob file
	Size     int64  // GGUF blob size in bytes
	Digest   string // sha256 digest from manifest
	Family   string // model family if inferrable from name
}

// OllamaManifest represents a parsed Ollama manifest file (OCI Image Manifest v2).
type OllamaManifest struct {
	SchemaVersion int           `json:"schemaVersion"`
	MediaType     string        `json:"mediaType"`
	Config        OllamaLayer   `json:"config"`
	Layers        []OllamaLayer `json:"layers"`
}

// OllamaLayer represents a single layer entry in an Ollama manifest.
type OllamaLayer struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

// modelAliases maps shorthand model names to canonical Ollama model names.
var modelAliases = map[string]string{
	"llama3":    "llama3.2",
	"llama":     "llama3.2",
	"codellama": "codellama",
	"phi":       "phi3",
	"mistral":   "mistral",
	"gemma":     "gemma2",
	"qwen":      "qwen2.5",
}

// modelFamilies maps base model name prefixes to family names for classification.
var modelFamilies = map[string]string{
	"llama":     "llama",
	"codellama": "llama",
	"phi":       "phi",
	"mistral":   "mistral",
	"gemma":     "gemma",
	"qwen":      "qwen",
	"deepseek":  "deepseek",
	"starcoder": "starcoder",
	"vicuna":    "llama",
}

// getOllamaModelsDir returns the Ollama models directory. Uses OLLAMA_MODELS env
// if set, otherwise defaults to ~/.ollama/models.
func getOllamaModelsDir() string {
	if dir := os.Getenv("OLLAMA_MODELS"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/tmp", ".ollama", "models")
	}
	return filepath.Join(home, ".ollama", "models")
}

// DiscoverOllamaModels scans Ollama's local storage and returns all models
// that have a GGUF layer. Returns an empty slice (not an error) if the Ollama
// models directory doesn't exist.
func DiscoverOllamaModels() ([]OllamaLocalModel, error) {
	basePath := getOllamaModelsDir()
	libraryDir := filepath.Join(basePath, "manifests", "registry.ollama.ai", "library")

	entries, err := os.ReadDir(libraryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading ollama manifests directory: %w", err)
	}

	var models []OllamaLocalModel
	for _, modelEntry := range entries {
		if !modelEntry.IsDir() {
			continue
		}
		modelName := modelEntry.Name()
		modelDir := filepath.Join(libraryDir, modelName)

		tagEntries, err := os.ReadDir(modelDir)
		if err != nil {
			continue
		}

		for _, tagEntry := range tagEntries {
			if tagEntry.IsDir() {
				continue
			}
			tag := tagEntry.Name()
			manifestPath := filepath.Join(modelDir, tag)

			manifest, err := parseOllamaManifest(manifestPath)
			if err != nil {
				continue
			}

			layer := findGGUFLayer(manifest)
			if layer == nil {
				continue
			}

			ggufPath := digestToBlobPath(basePath, layer.Digest)

			if _, err := os.Stat(ggufPath); err != nil {
				continue
			}

			models = append(models, OllamaLocalModel{
				Name:     modelName + ":" + tag,
				Base:     modelName,
				Tag:      tag,
				GGUFPath: ggufPath,
				Size:     layer.Size,
				Digest:   layer.Digest,
				Family:   inferModelFamily(modelName),
			})
		}
	}

	return models, nil
}

// ResolveOllamaModelPath resolves a model name to its GGUF blob path on disk.
// Accepts "model" (defaults to :latest) or "model:tag". Applies alias resolution.
func ResolveOllamaModelPath(name string) (string, error) {
	basePath := getOllamaModelsDir()

	modelName, tag := parseModelName(ResolveModelAlias(name))

	manifestPath := filepath.Join(
		basePath, "manifests", "registry.ollama.ai", "library", modelName, tag,
	)

	manifest, err := parseOllamaManifest(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("model not found: %s (no manifest at %s)", name, manifestPath)
		}
		return "", fmt.Errorf("invalid manifest for %s: %w", name, err)
	}

	layer := findGGUFLayer(manifest)
	if layer == nil {
		return "", fmt.Errorf("model %s has no GGUF layer", name)
	}

	blobPath := digestToBlobPath(basePath, layer.Digest)

	if _, err := os.Stat(blobPath); err != nil {
		return "", fmt.Errorf("model blob not found: %s (run: ollama pull %s)", blobPath, name)
	}

	return blobPath, nil
}

// ResolveModelAlias resolves a model alias to its canonical name.
// Returns the input unchanged if no alias exists.
func ResolveModelAlias(name string) string {
	base, tag := parseModelName(name)
	if canonical, ok := modelAliases[base]; ok {
		if tag == "latest" {
			return canonical
		}
		return canonical + ":" + tag
	}
	return name
}

// parseOllamaManifest reads and parses an Ollama manifest JSON file.
func parseOllamaManifest(manifestPath string) (*OllamaManifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var manifest OllamaManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parsing manifest %s: %w", manifestPath, err)
	}
	return &manifest, nil
}

// digestToBlobPath converts a digest like "sha256:abc123" to the blob file path
// under basePath. Ollama stores blobs as "sha256-abc123" (colon replaced with hyphen).
func digestToBlobPath(basePath, digest string) string {
	blobName := strings.Replace(digest, ":", "-", 1)
	return filepath.Join(basePath, "blobs", blobName)
}

// findGGUFLayer returns the first layer with the GGUF model media type, or nil.
func findGGUFLayer(manifest *OllamaManifest) *OllamaLayer {
	for i := range manifest.Layers {
		if manifest.Layers[i].MediaType == ollamaGGUFMediaType {
			return &manifest.Layers[i]
		}
	}
	return nil
}

// parseModelName splits "model:tag" into (model, tag). Defaults tag to "latest".
func parseModelName(name string) (string, string) {
	if idx := strings.IndexByte(name, ':'); idx >= 0 {
		return name[:idx], name[idx+1:]
	}
	return name, "latest"
}

// inferModelFamily returns a family name based on the model's base name prefix.
func inferModelFamily(baseName string) string {
	lower := strings.ToLower(baseName)
	for prefix, family := range modelFamilies {
		if strings.HasPrefix(lower, prefix) {
			return family
		}
	}
	return ""
}
