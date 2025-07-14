#!/bin/bash
set -euo pipefail

echo "=== Testing yq upgrade in build image ==="

# Test the exact same setup as Buildkite with the upgrade
docker run --rm -v "$(pwd)":/workdir --workdir /workdir \
  -e BUILDKITE_BUILD_CHECKOUT_PATH=/workdir \
  ghcr.io/theopenlane/build-image:latest sh -c '
    echo "Original yq version:"
    yq --version || echo "yq not found"

    # Install/upgrade yq like the fixed script does
    echo "Installing yq v4.45.4..."
    YQ_VERSION=4.45.4
    YQ_BINARY="yq_linux_amd64"
    wget -qO /tmp/yq https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/${YQ_BINARY}
    chmod +x /tmp/yq
    mv /tmp/yq /usr/local/bin/yq

    echo "New yq version: $(yq --version)"

    # Create test files that mimic the real scenario
    echo "core:
  server:
    listen: \":8080\"
    dev: true
  database:
    host: \"localhost\"
    port: 5432" > /tmp/test-source.yaml

    echo "core:
  server:
    listen: \":9090\"
    debug: false
  auth:
    enabled: true
helm:
  version: \"1.0.0\"
other:
  existing: \"keep\"" > /tmp/test-target.yaml

    echo ""
    echo "=== Testing the exact helm-automation logic ==="

    # Copy the exact logic from helm-automation.sh
    source="/tmp/test-source.yaml"
    target="/tmp/test-target.yaml"
    temp_merged="/tmp/temp-merged.yaml"

    echo "🔀 Merging with existing chart values..."
    # Copy target file as base
    cp "$target" "$temp_merged"

    # Extract core section from source and use it to replace target core section
    echo "📋 Replacing core section..."
    core_section=$(yq e ".core" "$source")
    echo "$core_section" > /tmp/core-section.yaml

    echo "Core section content:"
    cat /tmp/core-section.yaml

    echo ""
    echo "Running: yq e -i \".core = load(\"/tmp/core-section.yaml\")\" $temp_merged"
    if yq e -i ".core = load(\"/tmp/core-section.yaml\")" "$temp_merged"; then
        echo "✅ SUCCESS! Merge completed"
        echo ""
        echo "Final result:"
        cat "$temp_merged"
    else
        echo "❌ FAILED! Load function still not working"
        exit 1
    fi

    echo ""
    echo "=== Testing external secrets (if present) ==="
    if yq e ".externalSecrets" "$source" | grep -v "null" > /dev/null 2>&1; then
        echo "External secrets found, testing merge..."
        external_secrets_section=$(yq e ".externalSecrets" "$source")
        echo "$external_secrets_section" > /tmp/external-secrets-section.yaml
        yq e -i ".externalSecrets = load(\"/tmp/external-secrets-section.yaml\")" "$temp_merged"
        echo "✅ External secrets merge successful"
    else
        echo "No external secrets in test data"
    fi

    echo ""
    echo "=== Final merged result ==="
    cat "$temp_merged"
'