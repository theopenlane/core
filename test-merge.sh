#!/bin/bash
set -euo pipefail

# Test script to simulate the draft PR automation merge logic
echo "=== Testing Helm Values Merge Logic ==="

# Setup directories
CORE_DIR="/Users/manderson/core"
INFRA_DIR="/Users/manderson/openlane-infra"
CHART_DIR="$INFRA_DIR/charts/openlane"

# Source files from core
SOURCE_VALUES="$CORE_DIR/config/helm-values.yaml"
SOURCE_CONFIGMAP="$CORE_DIR/config/configmap.yaml"
SOURCE_EXTERNAL_SECRETS="$CORE_DIR/config/external-secrets"

# Target files in infra
TARGET_VALUES="$CHART_DIR/values.yaml"
TARGET_CONFIGMAP="$CHART_DIR/templates/core-configmap.yaml"
TARGET_EXTERNAL_SECRETS="$CHART_DIR/templates/external-secrets"

echo "📋 Source files:"
echo "  - Values: $SOURCE_VALUES"
echo "  - ConfigMap: $SOURCE_CONFIGMAP"
echo "  - External Secrets: $SOURCE_EXTERNAL_SECRETS"

echo "🎯 Target files:"
echo "  - Values: $TARGET_VALUES"
echo "  - ConfigMap: $TARGET_CONFIGMAP"
echo "  - External Secrets: $TARGET_EXTERNAL_SECRETS"

# Backup original target files outside templates directory to avoid Helm lint issues
echo "💾 Creating backups..."
cp "$TARGET_VALUES" "/tmp/values.yaml.backup"
if [[ -f "$TARGET_CONFIGMAP" ]]; then
    cp "$TARGET_CONFIGMAP" "/tmp/core-configmap.yaml.backup"
fi
if [[ -d "$TARGET_EXTERNAL_SECRETS" ]]; then
    cp -r "$TARGET_EXTERNAL_SECRETS" "/tmp/external-secrets.backup"
fi

# Test the merge function
merge_helm_values() {
  local source="$1"
  local target="$2"
  local description="$3"

  if [[ ! -f "$source" ]]; then
    echo "  ⚠️  Source file not found: $source"
    return 1
  fi

  if [[ ! -f "$target" ]]; then
    echo "  ⚠️  Target file not found: $target"
    return 1
  fi

  local temp_merged=$(mktemp)

  echo "  🔀 Merging $description..."
  # Copy target file as base
  cp "$target" "$temp_merged"

  # Merge the source config directly under the openlane section in target
  echo "  📋 Merging source config under openlane section..."
  if ! yq e -i '.openlane *= load("'"$source"'")' "$temp_merged"; then
    echo "  ❌ Direct merge under openlane failed, trying alternative approach"
    # Ensure openlane section exists
    yq e -i '.openlane //= {}' "$temp_merged"
    # Merge individual sections
    for section in $(yq e 'keys | .[]' "$source"); do
      echo "  📋 Merging $section section..."
      yq e -i ".openlane.$section = load(\"$source\").$section" "$temp_merged"
    done
  else
    echo "  ✅ Merge under openlane section completed successfully"
  fi

  # Replace target with merged content
  mv "$temp_merged" "$target"
  return 0
}

# Perform the merge
echo "🔄 Testing merge operation..."
if merge_helm_values "$SOURCE_VALUES" "$TARGET_VALUES" "Helm values.yaml"; then
  echo "✅ Values merge completed"
else
  echo "❌ Values merge failed"
  exit 1
fi

# Copy configmap
echo "📝 Copying ConfigMap..."
if [[ -f "$SOURCE_CONFIGMAP" ]]; then
  mkdir -p "$(dirname "$TARGET_CONFIGMAP")"
  cp "$SOURCE_CONFIGMAP" "$TARGET_CONFIGMAP"
  echo "✅ ConfigMap copied"
else
  echo "⚠️  Source ConfigMap not found"
fi

# Copy external secrets
echo "🔐 Copying External Secrets..."
if [[ -d "$SOURCE_EXTERNAL_SECRETS" ]]; then
  rm -rf "$TARGET_EXTERNAL_SECRETS"
  mkdir -p "$(dirname "$TARGET_EXTERNAL_SECRETS")"
  cp -r "$SOURCE_EXTERNAL_SECRETS" "$TARGET_EXTERNAL_SECRETS"
  echo "✅ External Secrets copied"
else
  echo "⚠️  Source External Secrets not found"
fi

# Debug: Check if config sections were merged under openlane
echo "🔍 Checking if config sections were merged under openlane..."
echo "📋 Checking openlane.server section:"
yq e '.openlane.server' "$TARGET_VALUES" | head -5
echo "📋 Checking openlane.auth section:"
yq e '.openlane.auth' "$TARGET_VALUES" | head -5
echo "📋 Checking openlane.externalSecrets section:"
yq e '.openlane.externalSecrets' "$TARGET_VALUES" | head -5

# Test with helm lint
echo "🔍 Running Helm lint..."
cd "$INFRA_DIR"
if task lint; then
  echo "🎉 Helm lint passed!"
else
  echo "💥 Helm lint failed!"
  echo ""
  echo "🔧 Restoring backups..."
  cp "/tmp/values.yaml.backup" "$TARGET_VALUES"
  if [[ -f "/tmp/core-configmap.yaml.backup" ]]; then
    cp "/tmp/core-configmap.yaml.backup" "$TARGET_CONFIGMAP"
  fi
  if [[ -d "/tmp/external-secrets.backup" ]]; then
    rm -rf "$TARGET_EXTERNAL_SECRETS"
    cp -r "/tmp/external-secrets.backup" "$TARGET_EXTERNAL_SECRETS"
  fi
  exit 1
fi

echo ""
echo "✅ Test completed successfully!"
echo "💡 To restore originals: "
echo "  cp /tmp/values.yaml.backup $TARGET_VALUES"
echo "  cp /tmp/core-configmap.yaml.backup $TARGET_CONFIGMAP"
echo "  cp -r /tmp/external-secrets.backup $TARGET_EXTERNAL_SECRETS"