# llama.cpp vs Ollama HTTP: Benchmark Template

**Tag hints:** `#llamacpp` `#performance`

Methodology and results template for comparing exarp-go's direct llama.cpp integration against Ollama's HTTP API.

---

## Test Methodology

### Variables Held Constant

- **Model**: Same GGUF file (e.g., Mistral 7B Q4_K_M)
- **Prompt**: Identical prompt text per test case
- **Parameters**: Same temperature, max_tokens, top_p
- **Hardware**: Same machine, same GPU
- **Warm-up**: 3 warm-up runs discarded before measurement

### Test Cases

| Test | Prompt | max_tokens | Purpose |
|------|--------|------------|---------|
| Short completion | "Write a haiku about coding" | 50 | Latency-sensitive |
| Medium generation | "Explain quicksort in 200 words" | 256 | Typical use |
| Long generation | "Write a detailed tutorial on..." | 1024 | Throughput |
| Classification | "Classify: positive/negative/neutral" | 10 | Minimal output |

### Metrics

| Metric | Unit | How Measured |
|--------|------|--------------|
| Time to first token (TTFT) | ms | Start of request → first token emitted |
| Total latency | ms | Start of request → last token |
| Tokens per second | tok/s | Total tokens / generation time |
| Peak RSS | MB | Memory usage during generation |
| Cold start latency | ms | First request after process start |

---

## Expected Advantages: llama.cpp Direct

| Factor | llamacpp | Ollama HTTP | Advantage |
|--------|----------|-------------|-----------|
| HTTP overhead | None | ~1-5ms per request | llamacpp |
| Server startup | None (in-process) | ~2-5s cold start | llamacpp |
| Memory sharing | Shares process memory | Separate process | llamacpp |
| Model loading | Cached in ModelManager | Cached in Ollama server | Comparable |
| Streaming | Direct callback | HTTP SSE chunking | llamacpp |
| GPU access | Direct Metal/CUDA | Via Ollama's Metal/CUDA | Comparable |

---

## Results

> **Status:** Template — populate after running benchmarks.

### Hardware

- **Machine**: (fill in)
- **CPU**: (fill in)
- **GPU**: (fill in)
- **RAM**: (fill in)
- **OS**: (fill in)

### Model: Mistral 7B Q4_K_M

| Test | Metric | llamacpp | Ollama HTTP | Difference |
|------|--------|----------|-------------|------------|
| Short | TTFT (ms) | — | — | — |
| Short | Total (ms) | — | — | — |
| Short | tok/s | — | — | — |
| Medium | TTFT (ms) | — | — | — |
| Medium | Total (ms) | — | — | — |
| Medium | tok/s | — | — | — |
| Long | TTFT (ms) | — | — | — |
| Long | Total (ms) | — | — | — |
| Long | tok/s | — | — | — |
| Classify | TTFT (ms) | — | — | — |
| Classify | Total (ms) | — | — | — |

### Memory

| Metric | llamacpp | Ollama HTTP |
|--------|----------|-------------|
| Peak RSS (idle) | — | — |
| Peak RSS (1 model loaded) | — | — |
| Peak RSS (during generation) | — | — |

### Cold Start

| Metric | llamacpp | Ollama HTTP |
|--------|----------|-------------|
| Process start → ready (ms) | — | — |
| First model load (ms) | — | — |
| First token after load (ms) | — | — |

---

## Running Benchmarks

### llamacpp

```bash
# Build with llamacpp tag
CGO_ENABLED=1 go build -tags llamacpp -o bin/exarp-go ./cmd/server

# Run benchmark (when implemented)
./bin/exarp-go -tool llamacpp -args '{
  "action": "generate",
  "model": "mistral-7b-q4_k_m.gguf",
  "prompt": "Write a haiku about coding",
  "max_tokens": 50
}'
```

### Ollama

```bash
# Ensure Ollama is running
ollama serve &

# Pull the same model
ollama pull mistral:7b-instruct-v0.3-q4_K_M

# Run via exarp-go
./bin/exarp-go -tool ollama -args '{
  "action": "generate",
  "model": "mistral:7b-instruct-v0.3-q4_K_M",
  "prompt": "Write a haiku about coding"
}'
```

---

## See Also

- `docs/LLAMACPP_BUILD_REQUIREMENTS.md` — Build prerequisites
- `docs/LLM_EXPOSURE_OPPORTUNITIES.md` — LLM abstraction overview
