#!/bin/bash
set -euo pipefail

#
# EXECUTION CONTEXT: Docker container (with git, gh, docker, jq, buildkite-agent)
# REQUIRED TOOLS: git, gh, jq, buildkite-agent
# ASSUMPTIONS: GitHub token available, Docker daemon accessible
#

# Post-merge PR automation
# Closes draft infra PRs after the core PR is merged.
# Final infra PRs are opened separately from main.

repo="${HELM_CHART_REPO}"

# Source shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"
source "${SCRIPT_DIR}/lib/github.sh"

# Install dependencies using shared library function
install_dependencies

log_execution_context "Post-Merge Draft PR Cleanup"

echo "=== Post-Merge Draft PR Cleanup ==="
echo "Repository: $repo"
echo "Branch: ${BUILDKITE_BRANCH}"
echo "Tools verified: ✅"

work=$(create_temp_workspace)

# Clone the target repository (needed for branch deletion)
echo "Cloning repository..."
if ! git clone "$repo" "$work"; then
  echo "❌ Failed to clone $repo" >&2
  exit 1
fi

cd "$work"

# Look for any existing open draft PRs to close
echo "🔍 Looking for existing draft PRs to close..."

draft_prs=$(find_draft_prs "$repo" "^🚧 DRAFT: Config changes from core PR #[0-9]+")

if [[ -z "$draft_prs" ]]; then
  echo "ℹ️  No existing draft PRs found to close"
  exit 0
fi

echo "📋 Found existing draft PR(s):"
echo "$draft_prs"

# Process each draft PR
while IFS=':' read -r pr_number branch_name title; do
  echo ""
  echo "🗑️  Evaluating draft PR #$pr_number (branch: $branch_name)"
  echo "📝 Title: $title"

  core_pr_number=$(extract_core_pr_number "$title")
  if [[ -z "$core_pr_number" ]]; then
    core_pr_number=$(extract_core_pr_number "$branch_name")
  fi

  if [[ -z "$core_pr_number" ]]; then
    echo "⚠️  Could not extract core PR number, skipping"
    continue
  fi

  echo "🔍 Checking status of core PR #$core_pr_number..."
  core_pr_info=$(check_core_pr_status "$core_pr_number")

  if [[ -z "$core_pr_info" ]]; then
    echo "⚠️  Core PR #$core_pr_number not found, closing draft PR #$pr_number"
    closing_comment=$(generate_closure_comment "$core_pr_number" "closed")
  else
    core_pr_state=$(echo "$core_pr_info" | jq -r '.state')
    core_pr_title=$(echo "$core_pr_info" | jq -r '.title')
    core_pr_updated=$(echo "$core_pr_info" | jq -r '.updatedAt')

    echo "📋 Core PR #$core_pr_number: '$core_pr_title' (State: $core_pr_state, Updated: $core_pr_updated)"

    if [[ "$core_pr_state" == "OPEN" ]]; then
      echo "ℹ️  Core PR #$core_pr_number is still open, keeping draft PR #$pr_number"
      continue
    fi

    if [[ "$core_pr_state" == "MERGED" ]]; then
      echo "✅ Core PR #$core_pr_number was merged ($core_pr_updated), closing draft PR #$pr_number"
      closing_comment=$(generate_closure_comment "$core_pr_number" "merged")
    else
      echo "🗑️  Core PR #$core_pr_number was closed ($core_pr_updated), closing draft PR #$pr_number"
      closing_comment=$(generate_closure_comment "$core_pr_number" "closed")
    fi
  fi

  if close_pr "$pr_number" "$repo" "$closing_comment"; then
    echo "🗑️  Deleting unused branch $branch_name"
    safe_delete_branch "$branch_name" || echo "⚠️  Failed to delete branch $branch_name"
  fi

done <<< "$draft_prs"

echo "🎉 Post-merge draft PR cleanup completed successfully"
