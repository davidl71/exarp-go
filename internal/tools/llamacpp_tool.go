// llamacpp_tool.go — MCP "llamacpp" tool: local GGUF model inference via llama.cpp bindings.
// Actions: status, models, generate, load, unload.
// Delegates to llamacpp_provider.go (CGO) or returns "not available" via llamacpp_nocgo.go.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/davidl71/exarp-go/internal/framework"
)

// preloadDone tracks project roots for which preload has already run (once per root).
var preloadMu sync.Mutex
var preloadDone = make(map[string]bool)

// RunLlamaCppPreloadIfConfigured loads models listed in .exarp/llamacpp.json "preload" on first use.
// If "warmup" is true, runs a short inference after each load. Safe to call multiple times; runs at most once per projectRoot.
func RunLlamaCppPreloadIfConfigured(projectRoot string) {
	if projectRoot == "" {
		return
	}
	preloadMu.Lock()
	if preloadDone[projectRoot] {
		preloadMu.Unlock()
		return
	}
	preloadDone[projectRoot] = true
	preloadMu.Unlock()

	cfg := loadLlamaCppAliasConfig(projectRoot)
	if len(cfg.Preload) == 0 {
		return
	}
	ctx := context.Background()
	for _, name := range cfg.Preload {
		path, err := ResolveModelToPath(strings.TrimSpace(name), projectRoot)
		if err != nil {
			continue // skip unresolvable entries
		}
		if err := LlamaCppLoad(path); err != nil {
			continue // skip load failures (e.g. missing file)
		}
		if cfg.Warmup {
			if p := DefaultLlamaCppProvider(); p != nil && p.Supported() {
				_, _ = p.Generate(ctx, "Hi", 5, 0)
			}
		}
	}
}

// handleLlamaCppTool dispatches llamacpp actions.
func handleLlamaCppTool(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	action, _ := params["action"].(string)
	if action == "" {
		action = "status"
	}

	switch action {
	case "status":
		return handleLlamaCppStatus()
	case "models":
		return handleLlamaCppModels()
	case "generate":
		return handleLlamaCppGenerate(ctx, params)
	case "load":
		return handleLlamaCppLoad(params)
	case "unload":
		return handleLlamaCppUnload(params)
	default:
		return nil, fmt.Errorf("unknown llamacpp action: %q", action)
	}
}

func handleLlamaCppStatus() ([]framework.TextContent, error) {
	projectRoot, _ := FindProjectRoot()
	RunLlamaCppPreloadIfConfigured(projectRoot)

	provider := DefaultLlamaCppProvider()
	gpu := DetectGPU()

	available := provider != nil && provider.Supported()
	mem := GetProcessMemoryInfo()
	memoryMB := mem.HeapSysMB
	if mem.RSSMB > 0 {
		memoryMB = mem.RSSMB
	}
	status := map[string]interface{}{
		"available":    available,
		"model_loaded": LlamaCppModelPath(),
		"manager":      ModelManagerStats(),
		"gpu": map[string]interface{}{
			"available":   gpu.Available,
			"backend":     gpu.Backend,
			"device_name": gpu.DeviceName,
			"memory_mb":   gpu.MemoryMB,
			"layers":      gpu.Layers,
		},
		"process_memory": map[string]interface{}{
			"memory_mb":   memoryMB,
			"heap_alloc_mb": mem.HeapAllocMB,
			"heap_sys_mb":   mem.HeapSysMB,
			"rss_mb":        mem.RSSMB,
			"limit_mb":      mem.LimitMB,
		},
		"platform": runtime.GOOS + "/" + runtime.GOARCH,
	}
	if mem.Warning != "" {
		status["process_memory"].(map[string]interface{})["warning"] = mem.Warning
	}

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling status: %w", err)
	}
	return []framework.TextContent{{Type: "text", Text: string(data)}}, nil
}

func handleLlamaCppModels() ([]framework.TextContent, error) {
	projectRoot, _ := FindProjectRoot()

	models, err := DiscoverOllamaModels()
	if err != nil {
		return nil, fmt.Errorf("discovering models: %w", err)
	}

	modelList := make([]map[string]interface{}, 0, len(models))
	for _, m := range models {
		modelList = append(modelList, map[string]interface{}{
			"name":      m.Name,
			"base":      m.Base,
			"tag":       m.Tag,
			"gguf_path": m.GGUFPath,
			"size":      m.Size,
			"family":    m.Family,
		})
	}

	aliases := ListAllAliases(projectRoot)

	result := map[string]interface{}{
		"models":  modelList,
		"aliases": aliases,
	}
	if len(models) == 0 {
		result["hint"] = "No GGUF models found in Ollama storage. Run: ollama pull <model>"
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling models: %w", err)
	}
	return []framework.TextContent{{Type: "text", Text: string(data)}}, nil
}

func handleLlamaCppGenerate(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	provider := DefaultLlamaCppProvider()
	if provider == nil || !provider.Supported() {
		return nil, fmt.Errorf("llamacpp provider not available (build with -tags llamacpp,cgo or set LLAMACPP_MODEL_PATH)")
	}

	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required for llamacpp generate")
	}

	maxTokens := 512
	if v, ok := params["max_tokens"].(float64); ok && v > 0 {
		maxTokens = int(v)
	}

	// Optionally truncate prompt to fit context window (context management)
	if contextSize, ok := params["context_size"].(float64); ok && contextSize > 0 {
		ctxSize := int(contextSize)
		if ctxSize > maxTokens {
			promptBudget := ctxSize - maxTokens
			prompt = TruncateToTokenBudget(prompt, promptBudget, 0)
		}
	} else if contextSize, ok := params["context_size"].(int); ok && contextSize > maxTokens {
		promptBudget := contextSize - maxTokens
		prompt = TruncateToTokenBudget(prompt, promptBudget, 0)
	}

	temperature := float32(0.7)
	if v, ok := params["temperature"].(float64); ok && v >= 0 {
		temperature = float32(v)
	}

	text, err := provider.Generate(ctx, prompt, maxTokens, temperature)
	if err != nil {
		return nil, fmt.Errorf("llamacpp generate failed: %w", err)
	}

	return []framework.TextContent{{Type: "text", Text: text}}, nil
}

func handleLlamaCppLoad(params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, _ := FindProjectRoot()

	modelPath, _ := params["model_path"].(string)
	if modelPath == "" {
		model, _ := params["model"].(string)
		if model != "" {
			resolved, err := ResolveModelToPath(model, projectRoot)
			if err != nil {
				return nil, fmt.Errorf("resolving model %q: %w", model, err)
			}
			modelPath = resolved
		}
	}

	if modelPath == "" {
		return nil, fmt.Errorf("model_path or model name is required for llamacpp load")
	}

	if err := LlamaCppLoad(modelPath); err != nil {
		return nil, fmt.Errorf("llamacpp load failed: %w", err)
	}
	return []framework.TextContent{{Type: "text", Text: fmt.Sprintf("Model loaded: %s", modelPath)}}, nil
}

func handleLlamaCppUnload(_ map[string]interface{}) ([]framework.TextContent, error) {
	LlamaCppUnload()
	return []framework.TextContent{{Type: "text", Text: "Model unloaded"}}, nil
}

// registerLlamaCppTool registers the "llamacpp" MCP tool with appropriate schema.
func registerLlamaCppTool(server framework.MCPServer) error {
	return server.RegisterTool(
		"llamacpp",
		"[HINT: action=status|models|generate|load|unload. Local GGUF model inference via llama.cpp. Use for direct GGUF model loading from Ollama blobs. Related: ollama, text_generate.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"status", "models", "generate", "load", "unload"},
					"default": "status",
				},
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Text prompt for generation (required for generate action)",
				},
				"model_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to GGUF model file (for load action)",
				},
				"model": map[string]interface{}{
					"type":        "string",
					"description": "Ollama model name to resolve to GGUF path (e.g. llama3.2:latest)",
				},
				"max_tokens": map[string]interface{}{
					"type":    "integer",
					"default": 512,
				},
				"temperature": map[string]interface{}{
					"type":    "number",
					"default": 0.7,
				},
				"gpu_layers": map[string]interface{}{
					"type":        "integer",
					"default":     -1,
					"description": "Number of layers to offload to GPU (-1 = all)",
				},
				"context_size": map[string]interface{}{
					"type":        "integer",
					"default":     2048,
					"description": "Context window size in tokens",
				},
			},
		},
		handleLlamaCppTool,
	)
}
