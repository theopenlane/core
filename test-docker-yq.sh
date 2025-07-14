#!/bin/bash
set -euo pipefail

echo "=== Testing yq load() function in build image ==="

# Test the exact same setup as Buildkite
docker run --rm -v "$(pwd)":/workdir --workdir /workdir \
  -e BUILDKITE_BUILD_CHECKOUT_PATH=/workdir \
  ghcr.io/theopenlane/build-image:latest sh -c '
    # Install yq like the script does
    if ! command -v yq >/dev/null 2>&1; then
        echo "Installing yq..."
        YQ_VERSION=4.45.4
        YQ_BINARY="yq_linux_amd64"
        wget -qO /tmp/yq https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/${YQ_BINARY}
        chmod +x /tmp/yq
        mv /tmp/yq /usr/local/bin/yq
    fi

    echo "yq version: $(yq --version)"

    # Create test files
    echo "core:
  server:
    listen: \":8080\"
    dev: true" > /tmp/test-source.yaml

    echo "core:
  server:
    listen: \":9090\"
helm:
  version: \"1.0.0\"" > /tmp/test-target.yaml

    echo "Testing basic extraction:"
    yq e ".core" /tmp/test-source.yaml

    echo ""
    echo "Testing load function:"
    yq e ".core" /tmp/test-source.yaml > /tmp/core-section.yaml

    echo "Core section file content:"
    cat /tmp/core-section.yaml

    echo ""
    echo "Testing load() function:"
    if yq e -i ".core = load(\"/tmp/core-section.yaml\")" /tmp/test-target.yaml; then
        echo "SUCCESS: load() function works"
        echo "Result:"
        cat /tmp/test-target.yaml
    else
        echo "FAILED: load() function does not work"
        echo "Using fallback approach:"
        # Remove load function completely
        cp /tmp/test-target.yaml /tmp/test-target-clean.yaml
        yq e -i "del(.core)" /tmp/test-target-clean.yaml
        echo "core:" >> /tmp/test-target-clean.yaml
        sed "s/^/  /" /tmp/core-section.yaml >> /tmp/test-target-clean.yaml
        echo "Fallback result:"
        cat /tmp/test-target-clean.yaml
    fi
'