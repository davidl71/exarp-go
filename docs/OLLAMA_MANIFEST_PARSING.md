# Ollama Manifest Parsing Strategy for llama.cpp Integration

**Tag hints:** `#design` `#llamacpp` `#ollama-integration`

Design for discovering GGUF model files from Ollama's local storage so llama.cpp can load them directly—enabling "serverless" inference using models already downloaded by Ollama, without the Ollama HTTP API.

---

## 1. Ollama Storage Directory Structure

Ollama stores models in a two-part layout inspired by OCI/Docker:

```
~/.ollama/models/                    # or OLLAMA_MODELS env override
├── blobs/
│   └── sha256-<hash>                # raw GGUF/binary blobs (content-addressed)
└── manifests/
    └── registry.ollama.ai/
        └── library/
            └── <model>/             # e.g., llama3.2, phi3, codellama
                └── <tag>            # e.g., latest, 1b, 3b, mini
```

**Example layout (macOS/Linux):**

```
~/.ollama/models/
├── blobs/
│   ├── sha256-633fc5be925f9a484b61d6f9b9a78021eeb462100bd557309f01ba84cac26adf
│   ├── sha256-542b217f179c7825eeb5bca3c77d2b75ed05bafbd3451d9188891a60a85337c6
│   └── sha256-23291dc44752bac878bf46ab0f2b8daf75c710060f80f1a351151c7be2f5ee0f
└── manifests/
    └── registry.ollama.ai/
        └── library/
            ├── phi3/
            │   ├── latest
            │   └── mini
            ├── llama3.2/
            │   ├── latest
            │   └── 1b
            └── codellama/
                └── 7b-instruct-q4_0
```

**Base path:** Default `~/.ollama/models`; override via `OLLAMA_MODELS` env var.

---

## 2. Manifest JSON Format

Manifests follow OCI Image Manifest v2 (`application/vnd.docker.distribution.manifest.v2+json`). Each manifest file (e.g. `manifests/registry.ollama.ai/library/phi3/mini`) is a JSON object:

```json
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
  "config": {
    "mediaType": "application/vnd.docker.container.image.v1+json",
    "digest": "sha256:23291dc44752bac878bf46ab0f2b8daf75c710060f80f1a351151c7be2f5ee0f",
    "size": 483
  },
  "layers": [
    {
      "mediaType": "application/vnd.ollama.image.model",
      "digest": "sha256:633fc5be925f9a484b61d6f9b9a78021eeb462100bd557309f01ba84cac26adf",
      "size": 2176177120
    },
    {
      "mediaType": "application/vnd.ollama.image.license",
      "digest": "sha256:fa8235e5b48faca34e3ca98cf4f694ef08bd216d28b58071a1f85b1d50cb814d",
      "size": 1084
    },
    {
      "mediaType": "application/vnd.ollama.image.template",
      "digest": "sha256:542b217f179c7825eeb5bca3c77d2b75ed05bafbd3451d9188891a60a85337c6",
      "size": 148
    },
    {
      "mediaType": "application/vnd.ollama.image.params",
      "digest": "sha256:8dde1baf1db03d318a2ab076ae363318357dff487bdd8c1703a29886611e581f",
      "size": 78
    }
  ]
}
```

**Relevant layer media types:**

| mediaType | Purpose |
|-----------|---------|
| `application/vnd.ollama.image.model` | **GGUF model weights** — use this for llama.cpp |
| `application/vnd.ollama.image.license` | License text |
| `application/vnd.ollama.image.template` | Prompt templates |
| `application/vnd.ollama.image.params` | Default generation params |

---

## 3. Extracting GGUF Blob Path from Manifest Digest

The GGUF layer uses digest format `sha256:<hex>`. The blob file on disk is named:

```
sha256-<hex>
```

**Conversion:** Replace the colon in `sha256:<hex>` with a hyphen → `sha256-<hex>`.

**Example:**
- Digest: `sha256:633fc5be925f9a484b61d6f9b9a78021eeb462100bd557309f01ba84cac26adf`
- Blob path: `{basePath}/blobs/sha256-633fc5be925f9a484b61d6f9b9a78021eeb462100bd557309f01ba84cac26adf`

**Algorithm:**

1. Find the layer with `mediaType == "application/vnd.ollama.image.model"`.
2. Extract `layer.digest` (e.g. `sha256:633fc5be...`).
3. Convert to blob filename: `strings.Replace(digest, ":", "-", 1)`.
4. Resolve path: `filepath.Join(basePath, "blobs", blobFilename)`.

---

## 4. Model Name → GGUF Path Resolution Algorithm

**Input:** Model name (e.g. `phi3`, `phi3:mini`, `llama3.2:1b`).

**Output:** Absolute path to GGUF blob file, or error.

```
1. Resolve basePath:
   - If OLLAMA_MODELS set → use it
   - Else → ~/.ollama/models (expand home dir)

2. Parse model name:
   - Split on ":" → (baseName, tag)
   - If no ":", tag = "latest"

3. Build manifest path:
   path = basePath/manifests/registry.ollama.ai/library/{baseName}/{tag}

4. Read manifest JSON from path

5. Find layer where mediaType == "application/vnd.ollama.image.model"

6. Convert digest to blob filename: sha256:xxx → sha256-xxx

7. Resolve blob path:
   blobPath = basePath/blobs/{blobFilename}

8. Verify blob file exists and is readable

9. Return blobPath
```

**Name normalization:**
- `phi3` → `phi3:latest`
- `phi3:mini` → `phi3:mini`
- For custom registries, `library/phi3:7b` → baseName `library/phi3`, tag `7b` (v1: library only)

---

## 5. Caching Strategy for Manifest Data

**Goals:** Avoid repeated filesystem scans and manifest reads; invalidate when Ollama adds/removes models.

**Approach: TTL-based in-memory cache**

| Cache entry | Key | Value | TTL |
|-------------|-----|-------|-----|
| Manifest content | `{basePath}:{model}:{tag}` | parsed manifest → GGUF digest | 5 minutes |
| Model list | `{basePath}:__list__` | `[]OllamaModel` | 2 minutes |

**Invalidation:**
- TTL expiry (primary)
- Optional: watch `manifests/` mtime; keep simple for v1

**Implementation notes:**
- Use `sync.Map` or small LRU for manifest cache
- `DiscoverOllamaModels` populates list cache; `ResolveModelPath` uses manifest cache
- Cache key includes `basePath` so `OLLAMA_MODELS` changes are respected

**No persistent cache** for v1—in-memory only.

---

## 6. Error Handling

| Scenario | Behavior |
|----------|----------|
| **Base path missing** | Return error: `ollama models directory not found: %s` |
| **Manifest file missing** | Return error: `model not found: %s (manifest missing)` |
| **Manifest JSON invalid** | Return error: `invalid manifest for %s: %w` |
| **No GGUF layer** | Return error: `model %s has no GGUF layer` |
| **Blob file missing** | Return error: `model blob not found: %s (run: ollama pull %s)` |
| **Blob not readable** | Return error: `cannot read model blob: %w` |
| **Ollama server running** | **No conflict** — we only read files. Concurrent load by both llama.cpp and Ollama may cause issues; document as limitation. |

---

## 7. API Design (Go Implementation)

### Package

Suggested: `internal/ollama/manifest.go` or `internal/tools/ollama_manifest.go`.

### Types

```go
// OllamaModel represents a discovered Ollama model.
type OllamaModel struct {
    Name   string // e.g. "phi3:mini", "llama3.2:1b"
    Tag    string // e.g. "mini", "latest"
    Base   string // base model name, e.g. "phi3"
    Size   int64  // GGUF blob size in bytes (from manifest)
    Digest string // full digest, e.g. "sha256:633fc5be..."
}

// OllamaManifest is the parsed manifest structure (minimal).
type OllamaManifest struct {
    SchemaVersion int             `json:"schemaVersion"`
    Config        *ManifestLayer  `json:"config"`
    Layers        []ManifestLayer `json:"layers"`
}

type ManifestLayer struct {
    MediaType string `json:"mediaType"`
    Digest    string `json:"digest"`
    Size      int64  `json:"size"`
}
```

### Functions

```go
// DiscoverOllamaModels scans the Ollama models directory and returns all
// models with GGUF layers. basePath defaults to ~/.ollama/models if empty.
func DiscoverOllamaModels(basePath string) ([]OllamaModel, error)

// ResolveModelPath returns the absolute path to the GGUF blob for the
// given model name. Name can be "model" or "model:tag"; tag defaults to "latest".
func ResolveModelPath(basePath, name string) (string, error)
```

### Discovery Algorithm (DiscoverOllamaModels)

```
1. basePath = resolveBasePath(basePath)
2. manifestsDir = basePath/manifests/registry.ollama.ai/library
3. For each modelDir in os.ReadDir(manifestsDir):
     For each tagFile in os.ReadDir(modelDir):
       manifestPath := manifestsDir/modelDir.Name/tagFile.Name
       manifest := readManifest(manifestPath)
       layer := findLayer(manifest, "application/vnd.ollama.image.model")
       if layer != nil:
         models = append(models, OllamaModel{Name: modelDir.Name+":"+tagFile.Name, ...})
4. Return models
```

### Usage Example

```go
models, err := ollama.DiscoverOllamaModels("")
if err != nil {
    return nil, err
}
path, err := ollama.ResolveModelPath("", "phi3:mini")
if err != nil {
    return err
}
// path = "/Users/me/.ollama/models/blobs/sha256-633fc5be..."
// Pass to go-skynet llama Load(path)
```

### Dependencies

- `os`, `path/filepath`, `encoding/json`, `strings`
- `os.UserHomeDir` for `~` expansion

---

## References

- [LLAMACPP_EVALUATION.md](LLAMACPP_EVALUATION.md) — go-skynet recommendation, model path strategy
- [Inside Ollama's Model Storage](https://medium.com/@dewasheesh.rana/inside-ollamas-model-storage-understanding-blobs-and-manifests-06f1620dd0b2)
- Ollama registry: `https://registry.ollama.ai/v2/library/{model}/manifests/{tag}`
