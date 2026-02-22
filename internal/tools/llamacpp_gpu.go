// llamacpp_gpu.go — GPU detection for llama.cpp inference (Metal, CUDA, ROCm, CPU fallback).
// No build tags — uses runtime/OS checks and environment variables only.
// exarp-tags: #feature #llamacpp
package tools

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

// GPUInfo describes the detected GPU backend and its capabilities.
type GPUInfo struct {
	Available  bool   `json:"available"`
	Backend    string `json:"backend"` // "metal", "cuda", "rocm", "cpu"
	DeviceName string `json:"device_name"`
	MemoryMB   int64  `json:"memory_mb"`
	Layers     int    `json:"layers"` // recommended GPU offload layers
}

// DetectGPU probes the system for GPU capabilities relevant to llama.cpp.
//
// Detection order:
//  1. darwin/arm64 → Metal (Apple Silicon GPU, always present).
//  2. CUDA_VISIBLE_DEVICES set → CUDA (NVIDIA GPU assumed).
//  3. ROC_ENABLE_PRE_VEGA / HSA_OVERRIDE_GFX_VERSION set → ROCm (AMD GPU).
//  4. Fallback → CPU only.
func DetectGPU() GPUInfo {
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		return detectMetal()
	}

	if cuda := os.Getenv("CUDA_VISIBLE_DEVICES"); cuda != "" {
		return detectCUDA(cuda)
	}

	if os.Getenv("ROC_ENABLE_PRE_VEGA") != "" || os.Getenv("HSA_OVERRIDE_GFX_VERSION") != "" {
		return GPUInfo{
			Available:  true,
			Backend:    "rocm",
			DeviceName: "AMD GPU (ROCm)",
			MemoryMB:   0,
			Layers:     32,
		}
	}

	return GPUInfo{
		Available:  false,
		Backend:    "cpu",
		DeviceName: cpuName(),
		MemoryMB:   0,
		Layers:     0,
	}
}

// RecommendGPULayers suggests how many transformer layers to offload to GPU
// based on model size and available GPU memory. Returns 0 if GPU memory is
// unknown or insufficient; -1 means "offload all layers".
func RecommendGPULayers(modelSizeMB int64, gpuMemoryMB int64) int {
	if gpuMemoryMB <= 0 || modelSizeMB <= 0 {
		return 0
	}

	ratio := float64(gpuMemoryMB) / float64(modelSizeMB)

	switch {
	case ratio >= 1.3:
		return -1 // all layers fit with room for KV cache
	case ratio >= 0.8:
		return 40
	case ratio >= 0.5:
		return 24
	case ratio >= 0.3:
		return 16
	default:
		return 0
	}
}

// detectMetal returns GPUInfo for Apple Silicon Metal.
func detectMetal() GPUInfo {
	mem := appleGPUMemoryMB()
	return GPUInfo{
		Available:  true,
		Backend:    "metal",
		DeviceName: "Apple Silicon GPU",
		MemoryMB:   mem,
		Layers:     -1, // Metal: offload all layers by default
	}
}

// appleGPUMemoryMB estimates shared GPU memory on Apple Silicon.
// Apple unified memory is shared; we approximate GPU-available as 75% of total RAM.
func appleGPUMemoryMB() int64 {
	if v := os.Getenv("EXARP_GPU_MEMORY_MB"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	// Conservative default for M1/M2/M3/M4 base models.
	return 12288 // 12 GB
}

// detectCUDA parses CUDA_VISIBLE_DEVICES and returns GPUInfo.
func detectCUDA(devices string) GPUInfo {
	devName := "NVIDIA GPU"
	if n := os.Getenv("NVIDIA_GPU_NAME"); n != "" {
		devName = n
	}

	var mem int64
	if v := os.Getenv("EXARP_GPU_MEMORY_MB"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			mem = n
		}
	}

	deviceCount := len(strings.Split(devices, ","))
	layers := 32
	if deviceCount > 1 {
		layers = 48
	}

	return GPUInfo{
		Available:  true,
		Backend:    "cuda",
		DeviceName: devName,
		MemoryMB:   mem,
		Layers:     layers,
	}
}

// cpuName returns a short description of the CPU for the "cpu" backend.
func cpuName() string {
	return runtime.GOARCH + " (" + runtime.GOOS + ")"
}
