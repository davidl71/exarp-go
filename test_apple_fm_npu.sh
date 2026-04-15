#!/bin/bash

# Apple Foundation Models NPU Monitoring Test Script
# Tests multiple API calls while monitoring Activity Monitor GPU History

set -e

BINARY_PATH="./bin/exarp-go"
if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ Error: $BINARY_PATH not found"
    echo "   Build first with: make build-apple-fm"
    exit 1
fi

echo "Testing Apple Foundation Models NPU Utilization"
echo "================================================"
echo ""
echo "⚠️  IMPORTANT: Have Activity Monitor GPU History window open!"
echo "   Window → GPU History"
echo ""
read -p "Press Enter to start tests..."

# Test 1: Simple generation
echo -e "\n[Test 1/5] Simple text generation..."
echo "   Prompt: 'Write a short poem about artificial intelligence'"
echo "   Monitoring NPU activity..."
"$BINARY_PATH" -tool apple_foundation_models -args '{"action":"generate","prompt":"Write a short poem about artificial intelligence","max_tokens":100}' || echo "   ⚠️  Test 1 failed"
sleep 3

# Test 2: Summarization
echo -e "\n[Test 2/5] Text summarization..."
echo "   Prompt: [Summarizing text about Apple Foundation Models]"
echo "   Monitoring NPU activity..."
"$BINARY_PATH" -tool apple_foundation_models -args '{"action":"summarize","prompt":"Apple Foundation Models provide on-device AI capabilities for Macs with Apple Silicon. They utilize the Neural Engine and NPU for efficient inference. This enables privacy-focused AI processing without sending data to external servers.","max_tokens":50}' || echo "   ⚠️  Test 2 failed"
sleep 3

# Test 3: Classification
echo -e "\n[Test 3/5] Text classification..."
echo "   Prompt: 'I love this product!'"
echo "   Categories: positive, negative, neutral"
echo "   Monitoring NPU activity..."
"$BINARY_PATH" -tool apple_foundation_models -args '{"action":"classify","prompt":"I love this product!","categories":"positive,negative,neutral","max_tokens":10}' || echo "   ⚠️  Test 3 failed"
sleep 3

# Test 4: Longer generation
echo -e "\n[Test 4/5] Longer text generation..."
echo "   Prompt: 'Explain how neural networks work in simple terms'"
echo "   Temperature: 0.8 (more creative)"
echo "   Monitoring NPU activity..."
"$BINARY_PATH" -tool apple_foundation_models -args '{"action":"generate","prompt":"Explain how neural networks work in simple terms","max_tokens":256,"temperature":0.8}' || echo "   ⚠️  Test 4 failed"
sleep 3

# Test 5: Creative mode
echo -e "\n[Test 5/5] Creative text generation..."
echo "   Prompt: 'Write a creative story about a robot learning to paint'"
echo "   Temperature: 0.9 (high creativity)"
echo "   Monitoring NPU activity..."
"$BINARY_PATH" -tool apple_foundation_models -args '{"action":"generate","prompt":"Write a creative story about a robot learning to paint","max_tokens":200,"temperature":0.9}' || echo "   ⚠️  Test 5 failed"

echo -e "\n✅ All tests completed"
echo ""
echo "📊 Check Activity Monitor GPU History for:"
echo "   - NPU activity spikes during each test"
echo "   - Neural Engine activity"
echo "   - Low CPU/GPU usage (NPU doing the work)"
echo ""
echo "📝 Document your observations in the task result comment"
