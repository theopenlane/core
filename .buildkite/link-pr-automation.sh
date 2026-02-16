#!/bin/bash
set -euo pipefail

# Source shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"
source "${SCRIPT_DIR}/lib/templates.sh"

# Install dependencies using shared library function
install_dependencies

echo "=== PR Linking Automation ==="

# Check if we're in a PR context
if [[ -z "${BUILDKITE_PULL_REQUEST:-}" || "${BUILDKITE_PULL_REQUEST}" == "false" ]]; then
  echo "ℹ️  Not running in PR context, skipping PR linking"
  exit 0
fi

# Check if draft PR metadata exists
if [[ ! -f "${BUILDKITE_BUILD_CHECKOUT_PATH}/.draft_pr_metadata" ]]; then
  echo "ℹ️  No draft PR metadata found, skipping PR linking"
  exit 0
fi

# Load draft PR metadata
source "${BUILDKITE_BUILD_CHECKOUT_PATH}/.draft_pr_metadata"

echo "Core PR: #${BUILDKITE_PULL_REQUEST}"
echo "Draft Infra PR: ${DRAFT_PR_URL}"

# Note: We don't send Slack notifications for draft PRs to avoid being too chatty.
# Notifications are sent for the final post-merge infrastructure PR workflow.

# Add comment to core PR linking to draft infra PR
echo "💬 Adding comment to core PR #${BUILDKITE_PULL_REQUEST}..."

template_dir=$(get_template_dir)
core_pr_comment=$(load_template "${template_dir}/github/core-pr-link-comment.md" \
    "DRAFT_PR_URL=${DRAFT_PR_URL}")

# Add comment to core PR
if gh pr comment "${BUILDKITE_PULL_REQUEST}" --body "$core_pr_comment"; then
  echo "✅ Comment added to core PR #${BUILDKITE_PULL_REQUEST}"
else
  echo "⚠️  Failed to add comment to core PR #${BUILDKITE_PULL_REQUEST}"
fi

# Extract draft PR number from URL for commenting
draft_pr_number=$(echo "$DRAFT_PR_URL" | grep -o '[0-9]*$')

if [[ -n "$draft_pr_number" ]]; then
  echo "💬 Adding comment to draft infra PR #${draft_pr_number}..."

  # Get core PR URL
  core_pr_url="https://github.com/theopenlane/core/pull/${BUILDKITE_PULL_REQUEST}"

  infra_pr_comment=$(load_template "${template_dir}/github/infra-pr-link-comment.md" \
      "CORE_PR_URL=${core_pr_url}")

  # Add comment to draft infra PR (need to specify the repo)
  if gh pr comment "$draft_pr_number" --repo "$(echo "$DRAFT_PR_URL" | sed 's|/pull/.*||' | sed 's|.*github.com/||')" --body "$infra_pr_comment"; then
    echo "✅ Comment added to draft infra PR #${draft_pr_number}"
  else
    echo "⚠️  Failed to add comment to draft infra PR #${draft_pr_number}"
  fi
else
  echo "⚠️  Could not extract PR number from draft PR URL: $DRAFT_PR_URL"
fi

# Note: No Slack notification sent for draft PR creation to reduce noise.
# Notifications are sent as part of the final post-merge workflow.

echo "🎉 PR linking completed successfully"
