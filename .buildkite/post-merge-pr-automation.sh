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

# Try to determine which core PR was just merged by checking recent commits
echo "🔍 Attempting to identify recently merged core PR..."
merged_core_pr=""

# Check if we can find the merged PR from the commit message or branch name
recent_commit_msg=$(git log --oneline -1 2>/dev/null || echo "")
if [[ "$recent_commit_msg" =~ \#([0-9]+) ]]; then
  potential_pr="${BASH_REMATCH[1]}"
  echo "🔍 Found potential PR number in commit message: #$potential_pr"

  # Verify this PR is actually merged
  pr_state=$(gh pr view "$potential_pr" --repo "theopenlane/core" --json state --jq '.state' 2>/dev/null || echo "NOT_FOUND")
  if [[ "$pr_state" == "MERGED" ]]; then
    merged_core_pr="$potential_pr"
    echo "✅ Confirmed recently merged core PR: #$merged_core_pr"
  fi
fi

# Get list of open draft PRs with our naming pattern
draft_prs=$(gh pr list --repo "$repo" --state open --draft --json number,title,headRefName \
  | jq -r '.[] | select(.title | test("^🚧 DRAFT: Config changes from core PR")) | "\(.number):\(.headRefName):\(.title)"')

# If we identified a specific merged core PR, filter to only process that one
if [[ -n "$merged_core_pr" ]]; then
  echo "🎯 Filtering draft PRs to only process core PR #$merged_core_pr"
  draft_prs=$(echo "$draft_prs" | grep "core PR #$merged_core_pr" || echo "")
fi

if [[ -z "$draft_prs" ]]; then
  echo "ℹ️  No draft PRs found to convert"
  exit 0
fi

echo "Found draft PRs to process:"
echo "$draft_prs"

# Process each draft PR
while IFS=':' read -r pr_number branch_name title; do
  echo ""
  echo "📝 Processing draft PR #$pr_number (branch: $branch_name)"

  # Extract core PR number from title for verification
  core_pr_from_title=$(echo "$title" | grep -o 'core PR #[0-9]*' | grep -o '[0-9]*' || echo "")

  # If we have a specific merged core PR, double-check this PR matches
  if [[ -n "$merged_core_pr" && "$core_pr_from_title" != "$merged_core_pr" ]]; then
    echo "⚠️  Draft PR #$pr_number is for core PR #$core_pr_from_title, not the recently merged #$merged_core_pr, skipping"
    continue
  fi

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

      # Create backup of existing values
      if [[ -f "$target" ]]; then
        cp "$target" "${target}.backup"
      fi

      # Create temporary merged file
      local temp_merged="${target}.merged"

      if [[ -f "$target" ]]; then
        yq e '.core' "$source" > /tmp/core-values.yaml

        yq e '. as $target | load("/tmp/core-values.yaml") as $core | $target | .core = $core' "$target" > "$temp_merged"

        # Also merge any externalSecrets configuration if it exists
        if yq e '.externalSecrets' "$source" | grep -v "null" > /dev/null 2>&1; then
          yq e '.externalSecrets' "$source" > /tmp/external-secrets.yaml
          yq e '. as $target | load("/tmp/external-secrets.yaml") as $secrets | $target | .externalSecrets = $secrets' "$temp_merged" > "${temp_merged}.tmp"
          mv "${temp_merged}.tmp" "$temp_merged"
        fi
      else
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
            cp "$source" "$target"
            git add "$target"
            changes_made=true
            change_summary+="\\n- ✅ Updated $description"
            return 0
          fi
        else
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
            rm -rf "$target"
            mkdir -p "$(dirname "$target")"
            cp -r "$source" "$target"
            git add "$target"
            changes_made=true
            change_summary+="\\n- 🔐 Updated $description"
            return 0
          fi
        else
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

      sed -i -E "s/^version:.*/version: $new_version/" "$chart_file"
      git add "$chart_file"
      changes_made=true
      change_summary+="\\n- 📈 Bumped chart version to $new_version"
    fi

    # Source helm documentation utilities from core repo
    source "${BUILDKITE_BUILD_CHECKOUT_PATH}/.buildkite/helm-docs-utils.sh"

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
      if git push origin "$branch_name"; then
        echo "✅ Branch updated successfully"
      else
        echo "⚠️  Failed to push updated branch"
      fi
    else
      echo "ℹ️  No additional changes needed"
    fi

    # Decide whether to convert to ready or close based on changes
    if [[ "$changes_made" == "true" ]]; then
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
      # Close the draft PR since there are no meaningful changes
      echo "🗑️  Closing draft PR #$pr_number - no configuration changes needed"

      # Extract core PR number from branch name
      core_pr_number=$(echo "$branch_name" | grep -o 'core-pr-[0-9]*' | cut -d'-' -f3)

      # Add closing comment
      closing_comment="## 🗑️ Closing Draft PR - No Configuration Changes Needed

This draft PR is being automatically closed because the related core repository changes did not require any infrastructure configuration updates.

### 📋 Summary:
- ✅ Core PR #${core_pr_number} has been merged successfully
- ℹ️  No Helm chart configuration changes were needed
- ℹ️  No external secrets updates were required
- ℹ️  No infrastructure deployment changes necessary

### 🔄 What This Means:
The core repository changes were code-only modifications that don't affect the infrastructure configuration. Therefore, no infrastructure PR is needed.

---
*This PR was automatically closed after the core repository merge.*"

      gh pr comment "$pr_number" --repo "$repo" --body "$closing_comment"

      if gh pr close "$pr_number" --repo "$repo"; then
        echo "✅ Draft PR #$pr_number closed successfully"

        # Optionally delete the branch since it's no longer needed
        echo "🗑️  Deleting unused branch $branch_name"
        git push origin --delete "$branch_name" || echo "⚠️  Failed to delete branch $branch_name"

      else
        echo "⚠️  Failed to close draft PR #$pr_number"
      fi
    fi

  else
    echo "⚠️  Failed to checkout branch $branch_name"
  fi

done <<< "$draft_prs"

echo "🎉 Post-merge PR automation completed successfully"
