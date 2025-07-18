#!/bin/bash
set -euo pipefail

# Source shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"
source "${SCRIPT_DIR}/lib/helm.sh"
source "${SCRIPT_DIR}/lib/github.sh"
source "${SCRIPT_DIR}/lib/templates.sh"

# Configuration
repo="${HELM_CHART_REPO}"
chart_dir="${HELM_CHART_PATH:-charts/openlane}"

# Install dependencies
install_dependencies

# Create workspace
work=$(create_temp_workspace)

# Log execution context
log_execution_context "Helm Chart Automation"

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
      yq e -i "${field_name} = \"${field_value}\"" "$file_path"
      changes_made=true
      change_summary+="\n- Updated $field_name in $(basename "$file_path")"
    fi
  done < <(env | grep "^${changes_prefix}" || true)
}

# Apply configuration changes using library functions
config_changes=$(apply_helm_config_changes \
  "$BUILDKITE_BUILD_CHECKOUT_PATH/config" \
  "$chart_dir")

if [[ -n "$config_changes" ]]; then
  changes_made=true
  change_summary="$config_changes"
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
  echo "ℹ️  No configuration changes detected, skipping PR creation"
  echo "ℹ️  This prevents unnecessary PRs when only code changes without config changes"
  exit 0
fi

echo "📝 Configuration changes detected, proceeding with PR creation"
echo -e "Summary:$change_summary"

# Source helm documentation utilities
source "${SCRIPT_DIR}/helm-docs-utils.sh"

# Generate documentation before committing
generate_docs_and_commit

# Setup git configuration
setup_git_user

# Increment chart version and generate changelog
chart_file="$chart_dir/Chart.yaml"
if [[ -f "$chart_file" ]]; then
  new_version=$(increment_chart_version "$chart_file")

  # Generate and update changelog
  changelog_entry=$(generate_changelog_entry "$new_version" "$change_summary")
  update_changelog "$chart_dir/CHANGELOG.md" "$changelog_entry"

  change_summary+="\n- 📈 Bumped chart version to $new_version\n- 📝 Updated changelog with detailed changes"
fi

# Create and push commit
build_info="- Build Number: ${BUILDKITE_BUILD_NUMBER}
- Source Commit: ${BUILDKITE_COMMIT:0:8}
- Branch: ${BUILDKITE_BRANCH:-unknown}"

create_commit "chore" "update Helm chart from core config changes" "Changes made:$change_summary" "$build_info"

# Push and create PR
echo "🚀 Pushing branch and creating PR..."
if safe_push_branch "$branch"; then
  pr_body=$(generate_core_config_pr_body "" "$change_summary")

  echo "Creating pull request..."
  if create_pr "$repo" "$branch" "🤖 Update Helm chart from core config (Build #${BUILDKITE_BUILD_NUMBER})" "$pr_body"; then
    pr_url=$(gh pr view "$branch" --repo "$repo" --json url --jq '.url' 2>/dev/null || echo "https://github.com/$repo/pull/new/$branch")
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
