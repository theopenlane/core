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
# Source shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"
# Install dependencies using shared library function
install_dependencies

echo "=== Post-Merge PR Automation ==="
echo "Repository: $repo"
echo "Chart directory: $chart_dir"
echo "Branch: ${BUILDKITE_BRANCH}"
echo "Tools verified: ✅"

source "${SCRIPT_DIR}/slack-utils.sh"
source "${SCRIPT_DIR}/lib/templates.sh"

# This runs on main branch after merge - find any draft PRs to convert
work=$(create_temp_workspace)

# Clone the target repository
echo "Cloning repository..."
if ! git clone "$repo" "$work"; then
  echo "❌ Failed to clone $repo" >&2
  exit 1
fi

cd "$work"

# Look for any existing open draft PRs to update with latest config
echo "🔍 Looking for existing draft PRs to update with latest configuration..."

# Get list of open draft PRs with our naming pattern
draft_prs=$(gh pr list --repo "$repo" --state open --draft --json number,title,headRefName \
  | jq -r '.[] | select(.title | test("^🚧 DRAFT: Config changes from core PR")) | "\(.number):\(.headRefName):\(.title)"')

if [[ -z "$draft_prs" ]]; then
  echo "ℹ️  No existing draft PRs found to update"
  echo "ℹ️  This means no previous core PRs created config changes that need infrastructure updates"
  exit 0
fi

echo "📋 Found existing draft PR(s) to update:"
echo "$draft_prs"
echo ""
echo "🔄 Updating with current end-state configuration from core repository..."

# Process each draft PR (update them all with latest config)
while IFS=':' read -r pr_number branch_name title; do
  echo ""
  echo "📝 Updating draft PR #$pr_number (branch: $branch_name) with latest configuration"
  echo "📝 Original title: $title"

  # Check out the draft branch
  if git checkout "$branch_name"; then
    echo "✅ Checked out branch $branch_name"

    # Update the branch with latest changes from the merged core repo
    echo "🔄 Updating draft PR with latest config changes..."

    # Source shared libraries to use the correct functions
    source "${SCRIPT_DIR}/lib/helm.sh"

    # Apply latest config changes using the shared library functions
    changes_made=false
    change_summary=""

    # Apply configuration changes using library functions
    # Note: this function returns 0 even when no changes are detected
    if config_changes=$(apply_helm_config_changes \
      "$BUILDKITE_BUILD_CHECKOUT_PATH/config" \
      "$chart_dir"); then

      if [[ -n "$config_changes" ]]; then
        changes_made=true
        change_summary="$config_changes"
        echo "✅ Configuration changes applied" >&2
      else
        echo "ℹ️  No configuration changes detected between source and target"
      fi
    else
      echo "⚠️  Failed to apply configuration changes"
      # Don't exit, continue to check for version increment
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

    # Configure git using shared library function
    setup_git_user

    # Commit any additional changes
    if [[ "$changes_made" == "true" ]]; then
      template_dir=$(get_template_dir)
      commit_message=$(load_template "${template_dir}/github/post-merge-commit.md" \
          "CHANGE_SUMMARY=${change_summary}" \
          "BUILD_NUMBER=${BUILDKITE_BUILD_NUMBER}" \
          "SOURCE_COMMIT_SHORT=${BUILDKITE_COMMIT:0:8}")

      git commit -m "$commit_message"

      # Push using shared function
      if safe_push_branch "$branch_name"; then
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
        new_title=$(gh pr view "$pr_number" --repo "$repo" --json title --jq '.title' | sed -E 's|^🚧 DRAFT: (.*)|🔄 \1|')

        # Add a comment about the conversion
        core_pr_number=$(echo "$branch_name" | grep -o 'core-pr-[0-9]*' | cut -d'-' -f3)
        conversion_comment=$(load_template "${template_dir}/github/pr-ready-comment.md" \
            "CORE_PR_NUMBER=${core_pr_number}")

        gh pr comment "$pr_number" --repo "$repo" --body "$conversion_comment"

        # Send Slack notification that PR is ready for review
        infra_pr_url=$(gh pr view "$pr_number" --repo "$repo" --json url --jq '.url')
        core_pr_url="https://github.com/theopenlane/core/pull/$(echo "$branch_name" | grep -o 'core-pr-[0-9]*' | cut -d'-' -f3)"

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
      closing_comment=$(load_template "${template_dir}/github/pr-close-comment.md" \
          "CORE_PR_NUMBER=${core_pr_number}")

      gh pr comment "$pr_number" --repo "$repo" --body "$closing_comment"

      if gh pr close "$pr_number" --repo "$repo"; then
        echo "✅ Draft PR #$pr_number closed successfully"

        # Delete the branch using shared function since it's no longer needed
        echo "🗑️  Deleting unused branch $branch_name"
        safe_delete_branch "$branch_name" || echo "⚠️  Failed to delete branch $branch_name"

      else
        echo "⚠️  Failed to close draft PR #$pr_number"
      fi
    fi

  else
    echo "⚠️  Failed to checkout branch $branch_name"
  fi

done <<< "$draft_prs"

echo "🎉 Post-merge PR automation completed successfully"
