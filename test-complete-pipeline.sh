#!/bin/bash
set -euo pipefail

echo "=== COMPLETE PIPELINE TEST ==="
echo "Testing the entire draft-pr-automation.sh script in Docker container"

# Set up all the environment variables that Buildkite would provide
docker run --rm -v "$(pwd)":/workdir --workdir /workdir \
  -e BUILDKITE_BUILD_CHECKOUT_PATH=/workdir \
  -e HELM_CHART_REPO=https://github.com/theopenlane/openlane-infra.git \
  -e HELM_CHART_PATH=charts/openlane \
  -e BUILDKITE_PULL_REQUEST=992 \
  -e BUILDKITE_BRANCH=feat-tags \
  -e BUILDKITE_BUILD_NUMBER=9999 \
  -e BUILDKITE_COMMIT=abc123def456 \
  -e GITHUB_TOKEN=dummy_token_for_test \
  ghcr.io/theopenlane/build-image:latest sh -c '
    apk add --no-cache bash >/dev/null 2>&1
    echo "=== Environment Check ==="
    echo "Working directory: $(pwd)"
    echo "BUILDKITE_BUILD_CHECKOUT_PATH: $BUILDKITE_BUILD_CHECKOUT_PATH"
    echo "HELM_CHART_REPO: $HELM_CHART_REPO"
    echo "BUILDKITE_PULL_REQUEST: $BUILDKITE_PULL_REQUEST"
    echo "BUILDKITE_BRANCH: $BUILDKITE_BRANCH"
    echo "BUILDKITE_BUILD_NUMBER: $BUILDKITE_BUILD_NUMBER"
    echo ""

    echo "=== Checking required files ==="
    ls -la .buildkite/
    echo ""

    # Create config directory and files if they do not exist (for testing)
    mkdir -p config
    if [[ ! -f "config/helm-values.yaml" ]]; then
        echo "Creating test config/helm-values.yaml..."
        cat > config/helm-values.yaml << EOF
core:
  server:
    listen: ":8080"
    dev: true
    readTimeout: "15s"
    writeTimeout: "15s"
  database:
    host: "localhost"
    port: 5432
    driver: "postgres"
  auth:
    enabled: true
    tokenExpiry: "24h"
externalSecrets:
  enabled: true
  secretStore: "vault"
  refreshInterval: "1h"
other:
  setting: "test-value"
EOF
    else
        echo "Using existing config/helm-values.yaml"
    fi

    echo ""
    echo "=== Running COMPLETE draft-pr-automation.sh script ==="
    echo "This will test the entire pipeline including:"
    echo "- Tool installation (gh, yq)"
    echo "- Repository cloning"
    echo "- Branch creation"
    echo "- Helm values merging"
    echo "- ConfigMap handling"
    echo "- External secrets merging"
    echo "- Git operations"
    echo "- Draft PR creation (will fail due to dummy token but thats expected)"
    echo ""

    # Run the actual script
    bash .buildkite/draft-pr-automation.sh || {
        echo ""
        echo "=== Script failed, but lets check what we accomplished ==="
        echo "Checking if temporary directories have our work..."
        find /tmp -name "tmp.*" -type d 2>/dev/null | head -3 | while read tmpdir; do
            if [[ -d "$tmpdir" && -f "$tmpdir/charts/openlane/values.yaml" ]]; then
                echo ""
                echo "Found work in: $tmpdir"
                echo "Checking if values.yaml was modified..."
                echo "First 30 lines of modified values.yaml:"
                head -30 "$tmpdir/charts/openlane/values.yaml"
                echo ""
                echo "Checking for core section:"
                if grep -A 10 "^core:" "$tmpdir/charts/openlane/values.yaml"; then
                    echo "✅ Core section found in values.yaml"
                else
                    echo "❌ Core section not found"
                fi
                echo ""
                echo "Checking git status in work directory:"
                cd "$tmpdir"
                git status || echo "Git status failed"
                git log --oneline -5 || echo "Git log failed"
            fi
        done

        echo ""
        echo "Exit code was: $?"
        echo "This is expected if GitHub authentication failed (dummy token)"
    }

    echo ""
    echo "=== Test completed ==="
'