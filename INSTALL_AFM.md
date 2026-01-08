# Installing maclocal-api (afm) Locally

Quick guide to install `afm` (Apple Foundation Models CLI) for local testing.

## ✅ Recommended: Use the Install Script

The repository includes a comprehensive install script that handles everything:

```bash
# Clone the repository
git clone https://github.com/scouzi1966/maclocal-api.git
cd maclocal-api

# Run the install script (builds and installs to /usr/local/bin)
./install.sh
```

The script will:
- ✅ Validate Swift version requirements
- ✅ Build the binary in release mode
- ✅ Install to `/usr/local/bin/afm`
- ✅ Verify installation

## 🔨 Build Manually

If you prefer to build manually:

```bash
# Clone the repository
git clone https://github.com/scouzi1966/maclocal-api.git
cd maclocal-api

# Build release binary
make build
# or
swift build -c release --product afm

# The binary will be at: .build/release/afm

# Test it locally (no install needed)
.build/release/afm --help

# Or install manually
sudo cp .build/release/afm /usr/local/bin/afm
sudo chmod +x /usr/local/bin/afm
```

## 🍺 Homebrew Formula (Local)

I've created a local Homebrew formula for you. To use it:

### Option 1: Create a Git Repository for the Tap

```bash
# Create a proper git repo for the tap
cd ~/homebrew-tap
git init
git add Formula/afm.rb
git commit -m "Add afm formula"

# Add as a local tap (uses file:// protocol)
brew tap davidl71/homebrew-tap file://$HOME/homebrew-tap
brew install afm
```

### Option 2: Install Directly from Formula File

```bash
# Use brew's --formula flag with full path
brew install --formula ~/homebrew-tap/Formula/afm.rb
```

The formula file is located at: `~/homebrew-tap/Formula/afm.rb`

## 📦 Use Pre-built Binary (Fastest)

The repository includes pre-built binaries in the releases:

```bash
# Download and extract pre-built binary (v0.7.0)
cd /tmp
git clone https://github.com/scouzi1966/maclocal-api.git
cd maclocal-api

# Extract the binary from the release archive
tar -xzf afm-v0.7.0-arm64.tar.gz

# Move to your PATH
sudo mv afm /usr/local/bin/afm
sudo chmod +x /usr/local/bin/afm
```

## 🧪 Test Installation

After installation, test it:

```bash
# Show help
afm --help

# Test single prompt mode
afm -s "Hello, how are you?"

# Test server mode (in another terminal)
afm --port 9999

# Test API endpoint
curl http://localhost:9999/health
```

## 📋 Requirements

- **macOS**: 15.1+ (Sequoia) recommended for Apple Intelligence
- **Swift**: 6.2+ (you have 6.2.3 ✅)
- **Hardware**: Apple Silicon Mac (ARM64)
- **System Settings**: Apple Intelligence must be enabled

## 🎯 Quick Start Usage

```bash
# Start API server (default port 9999)
afm

# Single prompt (no server)
afm -s "Explain quantum computing"

# Pipe input
echo "What is AI?" | afm

# Custom port
afm --port 8080

# With verbose logging
afm --verbose
```

## 🔗 Integration with exarp-go

You could integrate `afm` as an alternative backend for your Apple Foundation Models tool:

```go
// Option: Use HTTP client to call afm API instead of direct library
// POST http://localhost:9999/v1/chat/completions
```

Or keep using the direct `go-foundationmodels` library (current approach).

