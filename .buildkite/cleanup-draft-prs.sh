#!/bin/bash
set -euo pipefail

# Cleanup script to close draft PRs that correspond to merged core PRs
# This runs independently to handle edge cases where the main post-merge script doesn't run

# Source shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"
source "${SCRIPT_DIR}/lib/github.sh"

# Configuration
repo="${HELM_CHART_REPO}"

# Install dependencies
install_dependencies

# Log execution context
log_execution_context "Draft PR Cleanup"

# Validate build context
if ! validate_build_context "main" true; then
    exit 0
fi

# Find draft PRs that match our pattern and are still open
echo "🔍 Looking for draft PRs that need cleanup..."

# Get list of open draft PRs with our naming pattern
draft_prs=$(find_draft_prs "$repo" "^🚧 DRAFT: Config changes from core PR #[0-9]+")

if [[ -z "$draft_prs" ]]; then
  echo "ℹ️  No draft PRs found that need cleanup"
  exit 0
fi

echo "Found draft PRs to evaluate:"
echo "$draft_prs"

# Process each draft PR
while IFS=':' read -r pr_number branch_name title; do
  echo ""
  echo "🔍 Evaluating draft PR #$pr_number (branch: $branch_name)"

  # Extract core PR number from title
  core_pr_number=$(extract_core_pr_number "$title")

  if [[ -z "$core_pr_number" ]]; then
    echo "⚠️  Could not extract core PR number from title, skipping"
    continue
  fi

  echo "🔍 Checking status of core PR #$core_pr_number..."

  # Check if the corresponding core PR is merged or closed
  core_pr_info=$(check_core_pr_status "$core_pr_number")

  if [[ -z "$core_pr_info" ]]; then
    echo "⚠️  Could not fetch core PR #$core_pr_number information, skipping"
    continue
  fi

  core_pr_state=$(echo "$core_pr_info" | jq -r '.state')
  core_pr_title=$(echo "$core_pr_info" | jq -r '.title')
  core_pr_updated=$(echo "$core_pr_info" | jq -r '.updatedAt')

  echo "📋 Core PR #$core_pr_number: '$core_pr_title' (State: $core_pr_state, Updated: $core_pr_updated)"

  if [[ "$core_pr_state" == "MERGED" ]]; then
    # Additional safety check: only close if the core PR was merged relatively recently
    if is_recent_pr_activity "$core_pr_updated"; then
      echo "✅ Core PR #$core_pr_number was recently merged ($core_pr_updated), closing draft PR #$pr_number"

      closing_comment=$(generate_closure_comment "$core_pr_number" "merged")
      if close_pr "$pr_number" "$repo" "$closing_comment"; then
        safe_delete_branch "$branch_name"
      fi
    else
      echo "⚠️  Core PR #$core_pr_number was merged too long ago ($core_pr_updated), skipping cleanup for safety"
      continue
    fi

  elif [[ "$core_pr_state" == "CLOSED" ]]; then
    # Additional safety check: only close if the core PR was closed relatively recently
    if is_recent_pr_activity "$core_pr_updated"; then
      echo "🗑️  Core PR #$core_pr_number was recently closed ($core_pr_updated), closing draft PR #$pr_number"

      closing_comment=$(generate_closure_comment "$core_pr_number" "closed")
      if close_pr "$pr_number" "$repo" "$closing_comment"; then
        safe_delete_branch "$branch_name"
      fi
    else
      echo "⚠️  Core PR #$core_pr_number was closed too long ago ($core_pr_updated), skipping cleanup for safety"
      continue
    fi

  else
    echo "ℹ️  Core PR #$core_pr_number is still open (state: $core_pr_state), keeping draft PR #$pr_number"
  fi

done <<< "$draft_prs"

echo "🎉 Draft PR cleanup completed successfully"