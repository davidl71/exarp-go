#!/usr/bin/env bash
# build-llamacpp.sh — Clone go-skynet/go-llama.cpp and build libbinding.a.
# Metal is enabled automatically on macOS/arm64; CPU-only elsewhere.
# Usage: ./scripts/build-llamacpp.sh [--clean]
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LLAMA_DIR="${REPO_ROOT}/go-llama.cpp"
LLAMA_REPO="https://github.com/go-skynet/go-llama.cpp"

info()  { printf '\033[1;34m==>\033[0m %s\n' "$*"; }
warn()  { printf '\033[1;33mWARN:\033[0m %s\n' "$*"; }
error() { printf '\033[1;31mERROR:\033[0m %s\n' "$*" >&2; }

if [[ "${1:-}" == "--clean" ]]; then
    info "Cleaning go-llama.cpp build artifacts..."
    if [[ -d "$LLAMA_DIR" ]]; then
        make -C "$LLAMA_DIR" clean 2>/dev/null || true
    fi
    info "Clean complete."
    exit 0
fi

# --- Step 1: Clone if missing ---
if [[ ! -d "$LLAMA_DIR" ]]; then
    info "Cloning go-llama.cpp..."
    git clone --recurse-submodules "$LLAMA_REPO" "$LLAMA_DIR"
else
    info "go-llama.cpp already present at $LLAMA_DIR"
fi

# --- Step 2: Detect platform ---
OS="$(uname -s)"
ARCH="$(uname -m)"
BUILD_ARGS=()

if [[ "$OS" == "Darwin" && "$ARCH" == "arm64" ]]; then
    info "Detected macOS/arm64 — enabling Metal acceleration"
    BUILD_ARGS+=("BUILD_TYPE=metal")
elif [[ "$OS" == "Linux" ]]; then
    if command -v nvcc &>/dev/null; then
        info "Detected Linux with CUDA toolkit — enabling CUDA"
        BUILD_ARGS+=("BUILD_TYPE=cublas")
    else
        info "Detected Linux (CPU only)"
    fi
else
    info "Detected $OS/$ARCH — building CPU only"
fi

# --- Step 3: Build libbinding.a ---
info "Building libbinding.a (this may take a few minutes)..."
cd "$LLAMA_DIR"
make libbinding.a "${BUILD_ARGS[@]}" 2>&1 | tail -5

if [[ -f "$LLAMA_DIR/libbinding.a" ]]; then
    info "Build successful: $LLAMA_DIR/libbinding.a"
    ls -lh "$LLAMA_DIR/libbinding.a"
else
    error "libbinding.a not found after build"
    exit 1
fi

# --- Step 4: Print build instructions ---
echo ""
info "Next steps:"
echo "  1. Build exarp-go with llamacpp support:"
echo "     CGO_ENABLED=1 go build -tags llamacpp -o bin/exarp-go ./cmd/server"
echo ""
echo "  2. Set model path:"
echo "     export LLAMACPP_MODEL_PATH=/path/to/model.gguf"
echo "     # or set LLAMACPP_MODEL_DIR for a directory of models"
echo ""
echo "  3. Run:"
echo "     ./bin/exarp-go"
