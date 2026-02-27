#!/bin/bash
set -euo pipefail

# Cleanup script to close draft PRs that correspond to merged core PRs
# This runs independently to handle edge cases where the main post-merge script doesn't run

# Source shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"
source "${SCRIPT_DIR}/lib/github.sh"

# Allow tests to override functions
if [[ -n "${CLEANUP_PR_STUB:-}" && -f "${CLEANUP_PR_STUB}" ]]; then
    source "${CLEANUP_PR_STUB}"
fi

# Configuration
repo="${HELM_CHART_REPO}"

# Install dependencies
install_dependencies

# Log execution context
log_execution_context "Draft PR Cleanup"

# Validate build context
if ! validate_build_context "main" false; then
    exit 0
fi

# Find draft PRs that match our pattern and are still open
echo "🔍 Looking for draft PRs that need cleanup..."

# Match drafts by title (old and current formats) and deterministic branch naming.
draft_pr_pattern='(^🚧 DRAFT: Config (changes from core|from Core) PR #[0-9]+)|(^draft-core-pr-[0-9]+$)'
draft_prs=$(find_draft_prs "$repo" "$draft_pr_pattern")

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

  # Extract core PR number from title, then fall back to branch naming.
  core_pr_number=$(extract_core_pr_number "$title")
  if [[ -z "$core_pr_number" ]]; then
    core_pr_number=$(extract_core_pr_number "$branch_name")
  fi

  if [[ -z "$core_pr_number" ]]; then
    echo "⚠️  Could not extract core PR number from title or branch, skipping"
    continue
  fi

  echo "🔍 Checking status of core PR #$core_pr_number..."

  # Check if the corresponding core PR is merged or closed
  core_pr_info=$(check_core_pr_status "$core_pr_number")

  if [[ -z "$core_pr_info" ]]; then
    echo "⚠️  Core PR #$core_pr_number not found, closing draft PR #$pr_number"

    closing_comment=$(generate_closure_comment "$core_pr_number" "closed")
    if close_pr "$pr_number" "$repo" "$closing_comment"; then
      safe_delete_branch "$branch_name" "$repo"
    fi
    continue
  fi

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
  elif [[ "$core_pr_state" == "CLOSED" ]]; then
      echo "🗑️  Core PR #$core_pr_number was closed ($core_pr_updated), closing draft PR #$pr_number"
      closing_comment=$(generate_closure_comment "$core_pr_number" "closed")
  else
    echo "⚠️  Core PR #$core_pr_number has unexpected state $core_pr_state, closing draft PR #$pr_number"
    closing_comment=$(generate_closure_comment "$core_pr_number" "closed")
  fi

  if close_pr "$pr_number" "$repo" "$closing_comment"; then
    safe_delete_branch "$branch_name" "$repo"
  fi

done <<< "$draft_prs"

echo "🎉 Draft PR cleanup completed successfully"
