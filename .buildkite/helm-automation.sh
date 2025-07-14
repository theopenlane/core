#!/bin/bash
set -euo pipefail

#
# EXECUTION CONTEXT: Docker container (with git, gh, docker, jq, buildkite-agent)
# REQUIRED TOOLS: git, gh, docker, jq, buildkite-agent
# ASSUMPTIONS: GitHub token available, Docker daemon accessible
#

# Unified Helm automation script for updating charts from core config changes
# Consolidates functionality from helmpr.sh, pr.sh, and helm-values-pr.sh

# Check required tools are available in container
check_required_tools() {
    local missing_tools=()

    for tool in git gh docker jq buildkite-agent; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done

    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        echo "❌ Missing required tools: ${missing_tools[*]}"
        echo "This script must run in a container with these tools installed"
        echo "Expected to run via Buildkite pipeline with ghcr.io/theopenlane/build-image:latest"
        exit 1
    fi
}

# Verify environment variables are set
check_environment() {
    local missing_vars=()

    for var in HELM_CHART_REPO BUILDKITE_BUILD_NUMBER; do
        if [[ -z "${!var:-}" ]]; then
            missing_vars+=("$var")
        fi
    done

    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        echo "❌ Missing required environment variables: ${missing_vars[*]}"
        exit 1
    fi
}

# Run checks
check_required_tools
check_environment

YQ_VERSION=${YQ_VERSION:-4.9.6}
repo="${HELM_CHART_REPO}"
chart_dir="${HELM_CHART_PATH:-charts/openlane}"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "=== Helm Chart Automation ==="
echo "Repository: $repo"
echo "Chart directory: $chart_dir"
echo "Build: ${BUILDKITE_BUILD_NUMBER}"
echo "Tools verified: ✅"

# Clone the target repository
echo "Cloning repository..."
if ! git clone "$repo" "$work"; then
  echo "❌ Failed to clone $repo" >&2
  exit 1
fi

cd "$work"
branch="update-helm-config-${BUILDKITE_BUILD_NUMBER}"

echo "Creating branch: $branch"
git checkout -b "$branch"

# Track what changes we make
changes_made=false
change_summary=""

# Import slack utility functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/slack-utils.sh"

# Function to update YAML values using yq (from pr.sh functionality)
function update_yaml_values() {
  local file_path="$1"
  local changes_prefix="BUILDKITE_PLUGIN_YAML_UPDATE_VALUES"
  local metadata_prefix="BUILDKITE_PLUGIN_YAML_UPDATE_VALUES_FROM_METADATA"

  echo "Checking for YAML updates in $file_path"

  # Loop through environment variables for updates
  while read -r line; do
    if [[ "$line" == *"$metadata_prefix"* ]]; then
      field_name=$(echo "$line" | cut -d '=' -f 2 | sed "s/^${metadata_prefix}//")
      metadata_key=$(echo "$line" | cut -d '=' -f 3-)
      field_value=$(buildkite-agent meta-data get "$metadata_key" 2>/dev/null || echo "")
    elif [[ "$line" == *"$changes_prefix"* ]]; then
      field_name=$(echo "$line" | cut -d '=' -f 2 | sed "s/^${changes_prefix}//")
      field_value=$(echo "$line" | cut -d '=' -f 3-)
      # Evaluate environment variables
      field_value=$(eval "echo $field_value")
    else
      continue
    fi

    if [[ -n "$field_value" ]]; then
      echo "Updating $field_name to $field_value in $file_path"
      docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e -i "${field_name} = \"${field_value}\"" "$file_path"
      changes_made=true
      change_summary+="\n- Updated $field_name in $(basename "$file_path")"
    fi
  done < <(env | grep "^${changes_prefix}" || true)
}

# Function to merge Helm values with existing chart values
function merge_helm_values() {
  local source="$1"
  local target="$2"
  local description="$3"

  if [[ ! -f "$source" ]]; then
    echo "⚠️  Source values file not found: $source"
    return 1
  fi

  echo "🔄 Merging $description"

  # Create backup of existing values
  if [[ -f "$target" ]]; then
    cp "$target" "${target}.backup"
  fi

  # Create temporary merged file
  local temp_merged="${target}.merged"

  if [[ -f "$target" ]]; then
    # Merge strategy: Use yq to merge the generated core values with existing values
    # The 'core' section from our generated file will be merged/replaced
    # All other sections in the target will be preserved

    echo "  📋 Extracting core configuration from generated values..."
    docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e '.core' "$source" > /tmp/core-values.yaml

    echo "  🔀 Merging with existing chart values..."
    # Start with existing values, then merge in new core section
    docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e '. as $target | load("/tmp/core-values.yaml") as $core | $target | .core = $core' "$target" > "$temp_merged"

    # Also merge any externalSecrets configuration if it exists in generated file
    if docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e '.externalSecrets' "$source" | grep -v "null" > /dev/null 2>&1; then
      echo "  🔐 Merging external secrets configuration..."
      docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e '.externalSecrets' "$source" > /tmp/external-secrets.yaml
      docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e '. as $target | load("/tmp/external-secrets.yaml") as $secrets | $target | .externalSecrets = $secrets' "$temp_merged" > "${temp_merged}.tmp"
      mv "${temp_merged}.tmp" "$temp_merged"
    fi

  else
    echo "  ✨ Creating new values file..."
    cp "$source" "$temp_merged"
  fi

  # Check if there are actual differences
  if [[ -f "$target" ]] && diff -q "$target" "$temp_merged" > /dev/null 2>&1; then
    echo "  ℹ️  No changes detected in $description"
    rm -f "$temp_merged" "${target}.backup" /tmp/core-values.yaml /tmp/external-secrets.yaml
    return 1
  fi

  # Calculate detailed changes for changelog
  local changes_detail=""
  if [[ -f "$target" ]]; then
    echo "  📊 Analyzing changes..."

    # Compare core section changes
    if docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e '.core' "$target" > /tmp/old-core.yaml 2>/dev/null; then
      local core_changes=$(docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e 'diff("/tmp/old-core.yaml")' /tmp/core-values.yaml 2>/dev/null | grep -E '^\+\+\+|^---' | wc -l || echo "0")
      if [[ "$core_changes" -gt 0 ]]; then
        changes_detail+="\n    • Core configuration updated ($core_changes changes)"
      fi
    fi

    # Check for new/modified external secrets
    if [[ -f /tmp/external-secrets.yaml ]]; then
      changes_detail+="\n    • External secrets configuration updated"
    fi
  else
    changes_detail+="\n    • Initial values file created"
  fi

  # Apply the merged changes
  mv "$temp_merged" "$target"
  git add "$target"
  changes_made=true
  change_summary+="\n- 🔄 Merged $description$changes_detail"

  # Cleanup
  rm -f "${target}.backup" /tmp/core-values.yaml /tmp/external-secrets.yaml /tmp/old-core.yaml

  return 0
}

# Function to copy files and detect changes (for non-values files)
function copy_and_track() {
  local source="$1"
  local target="$2"
  local description="$3"

  if [[ -f "$source" ]]; then
    # Check if target exists and has differences
    if [[ -f "$target" ]]; then
      if ! diff -q "$source" "$target" > /dev/null 2>&1; then
        echo "Updating $description"
        cp "$source" "$target"
        git add "$target"
        changes_made=true
        change_summary+="\n- ✅ Updated $description"
        return 0
      fi
    else
      echo "Creating $description"
      mkdir -p "$(dirname "$target")"
      cp "$source" "$target"
      git add "$target"
      changes_made=true
      change_summary+="\n- ✨ Created $description"
      return 0
    fi
  fi
  return 1
}

function copy_directory_and_track() {
  local source="$1"
  local target="$2"
  local description="$3"

  if [[ -d "$source" ]]; then
    # Check if target exists and has differences
    if [[ -d "$target" ]]; then
      if ! diff -r "$source" "$target" > /dev/null 2>&1; then
        echo "Updating $description"
        rm -rf "$target"
        mkdir -p "$(dirname "$target")"
        cp -r "$source" "$target"
        git add "$target"
        changes_made=true
        change_summary+="\n- 🔐 Updated $description"
        return 0
      fi
    else
      echo "Creating $description"
      mkdir -p "$(dirname "$target")"
      cp -r "$source" "$target"
      git add "$target"
      changes_made=true
      change_summary+="\n- 🆕 Created $description"
      return 0
    fi
  fi
  return 1
}

# Update Helm values.yaml (intelligent merging approach)
merge_helm_values \
  "$BUILDKITE_BUILD_CHECKOUT_PATH/config/helm-values.yaml" \
  "$chart_dir/values.yaml" \
  "Helm values.yaml"

# Update external secrets directory
copy_directory_and_track \
  "$BUILDKITE_BUILD_CHECKOUT_PATH/config/external-secrets" \
  "$chart_dir/templates/external-secrets" \
  "External Secrets templates"

# Legacy configmap support (for backward compatibility)
if [[ -f "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" ]]; then
  copy_and_track \
    "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" \
    "$chart_dir/templates/core-configmap.yaml" \
    "ConfigMap template"
fi

# Apply any YAML value updates if configured
if [[ -n "${BUILDKITE_PLUGIN_YAML_UPDATE_FILE:-}" ]]; then
  target_file="$chart_dir/${BUILDKITE_PLUGIN_YAML_UPDATE_FILE}"
  if [[ -f "$target_file" ]]; then
    update_yaml_values "$target_file"
  fi
fi

# Check if we have any changes to commit
if git diff --staged --quiet; then
  echo "ℹ️  No changes detected, skipping PR creation"
  exit 0
fi

echo "📝 Changes detected, proceeding with PR creation"
echo -e "Summary:$change_summary"

# Configure git
git config user.email "bender@theopenlane.io"
git config user.name "theopenlane-bender"

# Increment chart version and generate changelog
chart_file="$chart_dir/Chart.yaml"
changelog_entry=""
if [[ -f "$chart_file" ]]; then
  current=$(grep '^version:' "$chart_file" | awk '{print $2}')
  IFS='.' read -r major minor patch <<< "$current"
  new_patch=$((patch+1))
  new_version="$major.$minor.$new_patch"

  echo "📈 Bumping chart version: $current -> $new_version"
  sed -i -E "s/^version:.*/version: $new_version/" "$chart_file"
  git add "$chart_file"

  # Generate detailed changelog entry
  changelog_entry="## [$new_version] - $(date +%Y-%m-%d)

### Changed$change_summary

### Build Information
- Build Number: ${BUILDKITE_BUILD_NUMBER}
- Source Commit: ${BUILDKITE_COMMIT:0:8}
- Source Branch: ${BUILDKITE_BRANCH:-unknown}
- Generated: $(date +'%Y-%m-%d %H:%M:%S UTC')

---
"

  # Update CHANGELOG.md if it exists, or create it
  changelog_file="$chart_dir/CHANGELOG.md"
  if [[ -f "$changelog_file" ]]; then
    # Insert new entry at the top after the header
    temp_changelog=$(mktemp)
    head -n 2 "$changelog_file" > "$temp_changelog"
    echo "" >> "$temp_changelog"
    echo "$changelog_entry" >> "$temp_changelog"
    tail -n +3 "$changelog_file" >> "$temp_changelog"
    mv "$temp_changelog" "$changelog_file"
  else
    # Create new changelog
    cat > "$changelog_file" << EOF
# Changelog

All notable changes to this Helm chart will be documented in this file.

$changelog_entry
EOF
  fi

  git add "$changelog_file"
  change_summary+="\n- 📈 Bumped chart version to $new_version\n- 📝 Updated changelog with detailed changes"
fi

# Create comprehensive commit message
commit_message="chore: update Helm chart from core config changes

This is an automated update from the core repository.

Changes made:$change_summary

Build Information:
- Build Number: ${BUILDKITE_BUILD_NUMBER}
- Source Commit: ${BUILDKITE_COMMIT:0:8}
- Branch: ${BUILDKITE_BRANCH:-unknown}"

git commit -m "$commit_message"

# Push and create PR
echo "🚀 Pushing branch and creating PR..."
if git push origin "$branch"; then
  pr_body="## 🤖 Automated Helm Chart Update

This PR updates the Helm chart based on changes in the core configuration structure.

### 📋 Changes Made:$change_summary

### 🔧 Build Information:
- **Build Number**: ${BUILDKITE_BUILD_NUMBER}
- **Source Commit**: [\`${BUILDKITE_COMMIT:0:8}\`](https://github.com/theopenlane/core/commit/${BUILDKITE_COMMIT})
- **Source Branch**: \`${BUILDKITE_BRANCH:-unknown}\`

### 🔍 What This Updates:
- **Helm Values**: Configuration schema and default values
- **External Secrets**: Secret management templates for sensitive configuration
- **ConfigMaps**: Non-sensitive configuration templates
- **Chart Version**: Automatically incremented patch version

This PR was automatically generated by the Buildkite pipeline and should be safe to merge after review."

  echo "Creating pull request..."
  if pr_url=$(gh pr create \
    --repo "$repo" \
    --head "$branch" \
    --title "🤖 Update Helm chart from core config (Build #${BUILDKITE_BUILD_NUMBER})" \
    --body "$pr_body" \
    --json url \
    --jq '.url'); then
    echo "✅ Pull request created successfully: $pr_url"

    # Send slack notification if configured
    send_helm_update_notification "$pr_url" "$change_summary"
  else
    echo "❌ Failed to create pull request"
    exit 1
  fi
else
  echo "❌ Failed to push branch"
  exit 1
fi

echo "🎉 Helm automation completed successfully"