#!/bin/bash
set -euo pipefail

#
# EXECUTION CONTEXT: Docker container (with git, gh, docker, jq, buildkite-agent)
# REQUIRED TOOLS: git, gh, docker, jq, buildkite-agent
# ASSUMPTIONS: GitHub token available, Docker daemon accessible
#

# Post-merge PR automation
# Converts draft infra PRs to ready for review after core PR is merged
# Updates the final infra PR with any additional changes from the merge


YQ_VERSION=${YQ_VERSION:-4.45.4}
repo="${HELM_CHART_REPO}"
chart_dir="${HELM_CHART_PATH:-charts/openlane}"

# Install yq if not available (since we're in a container, can't use docker run)
if ! command -v yq >/dev/null 2>&1; then
    echo "Installing yq..."
    YQ_BINARY="yq_linux_amd64"
    wget -qO /tmp/yq https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/${YQ_BINARY}
    chmod +x /tmp/yq
    mv /tmp/yq /usr/local/bin/yq
fi

# Install gh if not available
if ! command -v gh >/dev/null 2>&1; then
    echo "Installing gh..."
    apk add --no-cache github-cli
fi

echo "=== Post-Merge PR Automation ==="
echo "Repository: $repo"
echo "Chart directory: $chart_dir"
echo "Branch: ${BUILDKITE_BRANCH}"
echo "Tools verified: ✅"

# Import slack utility functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/slack-utils.sh"

# This runs on main branch after merge - find any draft PRs to convert
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# Clone the target repository
echo "Cloning repository..."
if ! git clone "$repo" "$work"; then
  echo "❌ Failed to clone $repo" >&2
  exit 1
fi

cd "$work"

# Find draft PRs that match our pattern and are still open
echo "🔍 Looking for draft PRs to convert..."

# Get list of open draft PRs with our naming pattern
draft_prs=$(gh pr list --repo "$repo" --state open --draft --json number,title,headRefName \
  | jq -r '.[] | select(.title | test("^🚧 DRAFT: Config changes from core PR")) | "\(.number):\(.headRefName)"')

if [[ -z "$draft_prs" ]]; then
  echo "ℹ️  No draft PRs found to convert"
  exit 0
fi

echo "Found draft PRs to process:"
echo "$draft_prs"

# Process each draft PR
while IFS=':' read -r pr_number branch_name; do
  echo ""
  echo "📝 Processing draft PR #$pr_number (branch: $branch_name)"

  # Check out the draft branch
  if git checkout "$branch_name"; then
    echo "✅ Checked out branch $branch_name"

    # Update the branch with latest changes from the merged core repo
    echo "🔄 Updating draft PR with latest config changes..."

    # Import the same functions and logic as helm-automation.sh
    changes_made=false
    change_summary=""

    # Source functions from helm-automation.sh or define them inline
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
        echo "  📋 Extracting core configuration from generated values..."
        yq e '.core' "$source" > /tmp/core-values.yaml

        echo "  🔀 Merging with existing chart values..."
        yq e '. as $target | load("/tmp/core-values.yaml") as $core | $target | .core = $core' "$target" > "$temp_merged"

        # Also merge any externalSecrets configuration if it exists
        if yq e '.externalSecrets' "$source" | grep -v "null" > /dev/null 2>&1; then
          echo "  🔐 Merging external secrets configuration..."
          yq e '.externalSecrets' "$source" > /tmp/external-secrets.yaml
          yq e '. as $target | load("/tmp/external-secrets.yaml") as $secrets | $target | .externalSecrets = $secrets' "$temp_merged" > "${temp_merged}.tmp"
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

      # Apply the merged changes
      mv "$temp_merged" "$target"
      git add "$target"
      changes_made=true
      change_summary+="\\n- 🔄 Updated $description"

      # Cleanup
      rm -f "${target}.backup" /tmp/core-values.yaml /tmp/external-secrets.yaml

      return 0
    }

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
            change_summary+="\\n- ✅ Updated $description"
            return 0
          fi
        else
          echo "Creating $description"
          mkdir -p "$(dirname "$target")"
          cp "$source" "$target"
          git add "$target"
          changes_made=true
          change_summary+="\\n- ✨ Created $description"
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
            change_summary+="\\n- 🔐 Updated $description"
            return 0
          fi
        else
          echo "Creating $description"
          mkdir -p "$(dirname "$target")"
          cp -r "$source" "$target"
          git add "$target"
          changes_made=true
          change_summary+="\\n- 🆕 Created $description"
          return 0
        fi
      fi
      return 1
    }

    # Apply latest config changes
    merge_helm_values \
      "$BUILDKITE_BUILD_CHECKOUT_PATH/config/helm-values.yaml" \
      "$chart_dir/values.yaml" \
      "Helm values.yaml"

    copy_directory_and_track \
      "$BUILDKITE_BUILD_CHECKOUT_PATH/config/external-secrets" \
      "$chart_dir/templates/external-secrets" \
      "External Secrets templates"

    if [[ -f "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" ]]; then
      copy_and_track \
        "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" \
        "$chart_dir/templates/core-configmap.yaml" \
        "ConfigMap template"
    fi

    # Increment chart version for final release
    chart_file="$chart_dir/Chart.yaml"
    if [[ -f "$chart_file" ]]; then
      current=$(grep '^version:' "$chart_file" | awk '{print $2}')
      IFS='.' read -r major minor patch <<< "$current"
      new_patch=$((patch+1))
      new_version="$major.$minor.$new_patch"

      echo "📈 Bumping chart version: $current -> $new_version"
      sed -i -E "s/^version:.*/version: $new_version/" "$chart_file"
      git add "$chart_file"
      changes_made=true
      change_summary+="\\n- 📈 Bumped chart version to $new_version"
    fi

    # Source helm documentation utilities
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    source "${SCRIPT_DIR}/helm-docs-utils.sh"

    # Generate documentation before committing
    generate_docs_and_commit

    # Configure git
    git config user.email "bender@theopenlane.io"
    git config user.name "theopenlane-bender"

    # Commit any additional changes
    if [[ "$changes_made" == "true" ]]; then
      commit_message="chore: finalize config changes after core merge

Updated configuration after core PR merge with final changes.

Changes made:$change_summary

Build Information:
- Build Number: ${BUILDKITE_BUILD_NUMBER}
- Source Commit: ${BUILDKITE_COMMIT:0:8}
- Branch: ${BUILDKITE_BRANCH}"

      git commit -m "$commit_message"

      # Push the updated branch
      echo "🚀 Pushing updated branch..."
      if git push origin "$branch_name"; then
        echo "✅ Branch updated successfully"
      else
        echo "⚠️  Failed to push updated branch"
      fi
    else
      echo "ℹ️  No additional changes needed"
    fi

    # Convert from draft to ready for review
    echo "🔄 Converting draft PR #$pr_number to ready for review..."

    if gh pr ready "$pr_number" --repo "$repo"; then
      echo "✅ PR #$pr_number converted from draft to ready"

      # Update the PR title to remove draft indicator
      new_title=$(gh pr view "$pr_number" --repo "$repo" --json title --jq '.title' | sed 's/^🚧 DRAFT: /🔄 /')
      gh pr edit "$pr_number" --repo "$repo" --title "$new_title"

      # Add a comment about the conversion
      conversion_comment="## ✅ PR Ready for Review

This PR has been automatically converted from draft to ready for review because the related core repository changes have been merged.

### 🔄 Status Update
- ✅ Core PR merged to main branch
- ✅ Configuration changes finalized
- ✅ Chart version incremented
- ✅ Ready for infrastructure deployment

### 📋 Final Review Checklist
- [ ] Review all configuration changes
- [ ] Verify chart version increment is appropriate
- [ ] Confirm external secrets configuration is correct
- [ ] Approve and merge to deploy changes

---
*This PR was automatically updated after the core repository merge.*"

      gh pr comment "$pr_number" --repo "$repo" --body "$conversion_comment"

      # Send Slack notification that PR is ready for review
      infra_pr_url=$(gh pr view "$pr_number" --repo "$repo" --json url --jq '.url')
      core_pr_url="https://github.com/theopenlane/core/pull/$(echo "$branch_name" | grep -o 'core-pr-[0-9]*' | cut -d'-' -f3)"
      core_pr_number=$(echo "$branch_name" | grep -o 'core-pr-[0-9]*' | cut -d'-' -f3)

      send_pr_ready_notification "$infra_pr_url" "$core_pr_url" "$core_pr_number" "$change_summary"

      echo "✅ PR #$pr_number updated and ready for final review"

    else
      echo "⚠️  Failed to convert PR #$pr_number from draft to ready"
    fi

  else
    echo "⚠️  Failed to checkout branch $branch_name"
  fi

done <<< "$draft_prs"

echo "🎉 Post-merge PR automation completed successfully"