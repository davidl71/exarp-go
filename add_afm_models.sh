#!/bin/bash
# Script to add 3 AFM model configurations for Cursor IDE
# This script helps verify and document the configurations

set -e

echo "🔧 AFM Model Configuration Helper for Cursor"
echo "=============================================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test function
test_endpoint() {
    local url=$1
    local name=$2
    
    echo -e "${BLUE}Testing $name...${NC}"
    if curl -s --max-time 5 "$url/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $name is accessible${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  $name is not accessible (may need to start service)${NC}"
        return 1
    fi
}

echo "📋 Three AFM Model Configurations:"
echo ""

# Configuration 1: Local
echo -e "${BLUE}1. AFM Local${NC}"
echo "   Name: Apple Foundation Models (Local)"
echo "   Base URL: http://localhost:9999/v1"
echo "   Model ID: foundation"
echo "   Use Case: Fast local access, development"
test_endpoint "http://localhost:9999" "Local AFM"
echo ""

# Configuration 2: Tailscale
echo -e "${BLUE}2. AFM Tailscale${NC}"
echo "   Name: Apple Foundation Models (Tailscale)"
echo "   Base URL: http://davids-mac-mini.tailf62197.ts.net/v1"
echo "   Model ID: foundation"
echo "   Use Case: Remote access from any Tailscale device"
test_endpoint "http://davids-mac-mini.tailf62197.ts.net" "Tailscale AFM"
echo ""

# Configuration 3: Local with Custom Port (if needed)
echo -e "${BLUE}3. AFM Local (Alternative)${NC}"
echo "   Name: Apple Foundation Models (Local Alt)"
echo "   Base URL: http://127.0.0.1:9999/v1"
echo "   Model ID: foundation"
echo "   Use Case: Alternative local access (127.0.0.1 instead of localhost)"
test_endpoint "http://127.0.0.1:9999" "Local AFM (127.0.0.1)"
echo ""

echo "📝 To add these in Cursor:"
echo ""
echo "1. Open Cursor Settings (Cmd + ,)"
echo "2. Go to 'Models' section"
echo "3. Click '+ Add Model' for each configuration above"
echo "4. Use the Base URL and Model ID shown above"
echo "5. Leave API Key empty (AFM doesn't require authentication)"
echo ""
echo "💡 Tip: You can add all 3 and switch between them as needed!"
echo ""

# Generate JSON config for reference
cat > /tmp/cursor_afm_models.json << 'EOF'
{
  "cursor_afm_models": [
    {
      "name": "Apple Foundation Models (Local)",
      "provider": "openai",
      "baseUrl": "http://localhost:9999/v1",
      "model": "foundation",
      "apiKey": "",
      "description": "Fast local access for development"
    },
    {
      "name": "Apple Foundation Models (Tailscale)",
      "provider": "openai",
      "baseUrl": "http://davids-mac-mini.tailf62197.ts.net/v1",
      "model": "foundation",
      "apiKey": "",
      "description": "Remote access via Tailscale network"
    },
    {
      "name": "Apple Foundation Models (Local Alt)",
      "provider": "openai",
      "baseUrl": "http://127.0.0.1:9999/v1",
      "model": "foundation",
      "apiKey": "",
      "description": "Alternative local access using 127.0.0.1"
    }
  ]
}
EOF

echo "📄 Configuration reference saved to: /tmp/cursor_afm_models.json"
echo ""
echo "✅ Done! Use the information above to add models in Cursor Settings."

