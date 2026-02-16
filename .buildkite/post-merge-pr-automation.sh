#!/bin/bash
set -euo pipefail

# Source shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"
source "${SCRIPT_DIR}/lib/github.sh"

# Allow tests to override functions
if [[ -n "${POST_MERGE_PR_STUB:-}" && -f "${POST_MERGE_PR_STUB}" ]]; then
  source "${POST_MERGE_PR_STUB}"
fi

# Configuration
repo="${HELM_CHART_REPO}"

# Install dependencies
install_dependencies

# Log execution context
log_execution_context "Post-Merge Draft PR Cleanup"

# Validate build context
if ! validate_build_context "main" false; then
  exit 0
fi

echo "🔍 Looking for draft infrastructure PRs to close..."

# Match drafts by title (old and current formats) and deterministic branch naming.
draft_pr_pattern='(^🚧 DRAFT: Config (changes from core|from Core) PR #[0-9]+)|(^draft-core-pr-[0-9]+$)'
draft_prs=$(find_draft_prs "$repo" "$draft_pr_pattern")

if [[ -z "$draft_prs" ]]; then
  echo "ℹ️  No draft PRs found to close"
  exit 0
fi

echo "Found draft PRs to evaluate:"
echo "$draft_prs"

while IFS=':' read -r pr_number branch_name title; do
  echo ""
  echo "🔍 Evaluating draft PR #$pr_number (branch: $branch_name)"

  core_pr_number=$(extract_core_pr_number "$title")
  if [[ -z "$core_pr_number" ]]; then
    core_pr_number=$(extract_core_pr_number "$branch_name")
  fi

  if [[ -z "$core_pr_number" ]]; then
    echo "⚠️  Could not extract core PR number from title or branch, skipping"
    continue
  fi

  core_pr_info=$(check_core_pr_status "$core_pr_number")
  if [[ -z "$core_pr_info" ]]; then
    echo "⚠️  Core PR #$core_pr_number not found, closing draft PR #$pr_number"
    closing_comment=$(generate_closure_comment "$core_pr_number" "closed")
    if close_pr "$pr_number" "$repo" "$closing_comment"; then
      safe_delete_branch "$branch_name"
    fi
    continue
  fi

  core_pr_state=$(echo "$core_pr_info" | jq -r '.state')
  core_pr_updated=$(echo "$core_pr_info" | jq -r '.updatedAt')

  if [[ "$core_pr_state" == "OPEN" ]]; then
    echo "ℹ️  Core PR #$core_pr_number is still open, keeping draft PR #$pr_number"
    continue
  fi

  if [[ "$core_pr_state" == "MERGED" ]]; then
    echo "✅ Core PR #$core_pr_number was merged ($core_pr_updated), closing draft PR #$pr_number"
    closing_comment=$(generate_closure_comment "$core_pr_number" "merged")
  else
    echo "🗑️  Core PR #$core_pr_number is $core_pr_state ($core_pr_updated), closing draft PR #$pr_number"
    closing_comment=$(generate_closure_comment "$core_pr_number" "closed")
  fi

  if close_pr "$pr_number" "$repo" "$closing_comment"; then
    safe_delete_branch "$branch_name"
  fi
done <<< "$draft_prs"

echo "🎉 Post-merge draft PR cleanup completed successfully"
