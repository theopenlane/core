#!/bin/bash
set -euo pipefail

echo "=== Testing yq commands locally ==="

# Clean up any existing test directory
rm -rf ~/tmp/yq-test
mkdir -p ~/tmp/yq-test
cd ~/tmp/yq-test

echo "📁 Created test directory: $(pwd)"

# Create test files
cat > test-source.yaml << 'EOF'
core:
  server:
    listen: ":8080"
    dev: true
  database:
    host: "localhost"
    port: 5432
externalSecrets:
  enabled: true
  secretStore: "vault"
other:
  setting: "value"
EOF

cat > test-target.yaml << 'EOF'
core:
  server:
    listen: ":9090"
    debug: false
  auth:
    enabled: true
helm:
  version: "1.0.0"
other:
  existing: "keep"
EOF

echo "📝 Created test YAML files"

# Test 1: Basic extraction
echo ""
echo "Test 1: Extract core section"
echo "Command: yq e '.core' test-source.yaml"
echo "Result:"
yq e '.core' test-source.yaml
echo "✅ Basic extraction works"

# Test 2: Save to temp file and test load function
echo ""
echo "Test 2: Test load function (current approach)"
yq e '.core' test-source.yaml > /tmp/core-values.yaml
echo "Command: yq e '.core = load(\"/tmp/core-values.yaml\")' test-target.yaml"
echo "Result:"
if yq e '.core = load("/tmp/core-values.yaml")' test-target.yaml 2>/dev/null; then
    echo "✅ Load function works"
else
    echo "❌ Load function failed - this explains the Buildkite error"
    echo ""
    echo "Test 3: Try eval-all approach instead"
    echo "Command: yq eval-all 'select(fileIndex == 0) * {\"core\": (select(fileIndex == 1))}' test-target.yaml /tmp/core-values.yaml"
    echo "Result:"
    yq eval-all 'select(fileIndex == 0) * {"core": (select(fileIndex == 1))}' test-target.yaml /tmp/core-values.yaml
    echo "✅ eval-all approach works"
fi

# Test 4: External secrets
echo ""
echo "Test 4: External secrets handling"
echo "Command: yq e '.externalSecrets' test-source.yaml"
echo "Result:"
yq e '.externalSecrets' test-source.yaml
echo ""
echo "Check for null:"
if yq e '.externalSecrets' test-source.yaml | grep -v "null" > /dev/null 2>&1; then
    echo "✅ External secrets detected (not null)"

    # Test external secrets merge
    yq e '.externalSecrets' test-source.yaml > /tmp/external-secrets.yaml
    echo "Testing external secrets merge with eval-all:"
    yq eval-all 'select(fileIndex == 0) * {"core": (select(fileIndex == 1))}' test-target.yaml /tmp/core-values.yaml > temp-merged.yaml
    yq eval-all 'select(fileIndex == 0) * {"externalSecrets": (select(fileIndex == 1))}' temp-merged.yaml /tmp/external-secrets.yaml
    echo "✅ External secrets merge works"
else
    echo "No external secrets found"
fi

# Test 5: Version check
echo ""
echo "Test 5: yq version information"
echo "Command: yq --version"
yq --version

echo ""
echo "=== Summary ==="
echo "If load() function failed, the Buildkite error is due to yq version differences."
echo "The eval-all approach should work across yq versions."

# Clean up
cd - > /dev/null
rm -rf ~/tmp/yq-test

echo "🧹 Cleaned up test files"