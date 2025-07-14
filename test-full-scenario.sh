#!/bin/bash
set -euo pipefail

echo "=== Full CI scenario test ==="

# Create a test directory and clone both repos
TEST_DIR="/tmp/buildkite-test-$(date +%s)"
mkdir -p "$TEST_DIR"
echo "Test directory: $TEST_DIR"

# Test the exact same setup as Buildkite
docker run --rm -v "$(pwd)":/workdir -v "$TEST_DIR":/test --workdir /workdir \
  -e BUILDKITE_BUILD_CHECKOUT_PATH=/workdir \
  -e HELM_CHART_REPO=https://github.com/theopenlane/openlane-infra.git \
  -e HELM_CHART_PATH=charts/openlane \
  -e BUILDKITE_BUILD_NUMBER=9999 \
  ghcr.io/theopenlane/build-image:latest sh -c '
    echo "=== Testing full scenario in build container ==="

    # Check if we have the real files needed
    if [[ ! -f "/workdir/config/helm-values.yaml" ]]; then
        echo "⚠️  No helm-values.yaml found, creating test file"
        mkdir -p /workdir/config
        cat > /workdir/config/helm-values.yaml << EOF
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
EOF
    fi

    # Source the actual helm-automation.sh functions
    echo "Sourcing helm-automation.sh..."
    source /workdir/.buildkite/helm-automation.sh

    # Test the merge_helm_values function with real repos
    echo ""
    echo "=== Testing merge_helm_values function ==="

    # Create temp directory for the test
    work_temp=$(mktemp -d)
    cd "$work_temp"

    # Clone the actual openlane-infra repo
    echo "Cloning openlane-infra..."
    if git clone "$HELM_CHART_REPO" .; then
        echo "✅ Successfully cloned openlane-infra"

        # Check if target values.yaml exists
        if [[ -f "charts/openlane/values.yaml" ]]; then
            echo "✅ Found target values.yaml"

            # Run the actual merge function
            echo ""
            echo "=== Running merge_helm_values function ==="
            if merge_helm_values "/workdir/config/helm-values.yaml" "charts/openlane/values.yaml" "Test merge"; then
                echo "✅ merge_helm_values function completed successfully"

                # Show the result
                echo ""
                echo "=== Merged values.yaml preview ==="
                head -50 charts/openlane/values.yaml
            else
                echo "❌ merge_helm_values function failed"
                exit 1
            fi
        else
            echo "❌ Target values.yaml not found in charts/openlane/"
            ls -la charts/openlane/
            exit 1
        fi
    else
        echo "❌ Failed to clone openlane-infra"
        exit 1
    fi

    echo ""
    echo "=== Test completed successfully ==="
'

# Clean up
echo "Cleaning up test directory: $TEST_DIR"
rm -rf "$TEST_DIR"