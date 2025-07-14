#!/bin/bash
set -euo pipefail

echo "=== Testing Script Execution Locally ==="
echo "This mimics what Buildkite agent does before passing to Docker"

# Test 1: Check if script exists
echo ""
echo "Test 1: Checking if draft-pr-automation.sh exists..."
if [[ -f ".buildkite/draft-pr-automation.sh" ]]; then
    echo "✅ Script exists"
    ls -la .buildkite/draft-pr-automation.sh
else
    echo "❌ Script not found"
    exit 1
fi

# Test 2: Test chmod command
echo ""
echo "Test 2: Testing chmod command..."
chmod +x .buildkite/draft-pr-automation.sh
echo "✅ chmod completed"
ls -la .buildkite/draft-pr-automation.sh

# Test 3: Test script execution path
echo ""
echo "Test 3: Testing script path resolution..."
if [[ -x "./.buildkite/draft-pr-automation.sh" ]]; then
    echo "✅ Script is executable via ./.buildkite/draft-pr-automation.sh"
else
    echo "❌ Script not executable"
fi

# Test 4: Test with various path formats
echo ""
echo "Test 4: Testing different path formats..."
echo "Current directory: $(pwd)"
echo "Script with ./ prefix: $(ls -la ./.buildkite/draft-pr-automation.sh 2>/dev/null && echo "exists" || echo "not found")"
echo "Script without prefix: $(ls -la .buildkite/draft-pr-automation.sh 2>/dev/null && echo "exists" || echo "not found")"

# Test 5: Test the actual command
echo ""
echo "Test 5: Testing the exact command sequence..."
echo "Running: chmod +x .buildkite/draft-pr-automation.sh && ./.buildkite/draft-pr-automation.sh --help"
if chmod +x .buildkite/draft-pr-automation.sh && ./.buildkite/draft-pr-automation.sh --help 2>/dev/null; then
    echo "✅ Command sequence works"
else
    echo "❌ Command sequence failed with exit code: $?"
    echo "Let's check what the script outputs when run without args..."
    ./.buildkite/draft-pr-automation.sh || echo "Script exited with code: $?"
fi

echo ""
echo "=== Test Complete ==="