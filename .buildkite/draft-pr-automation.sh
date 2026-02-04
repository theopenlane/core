#!/bin/bash
set -euo pipefail

# Draft PR automation for config changes
# Creates draft PRs in openlane-infra when config changes are detected in core repo

# Source shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"
source "${SCRIPT_DIR}/lib/helm.sh"
source "${SCRIPT_DIR}/lib/github.sh"

# Configuration
repo="${HELM_CHART_REPO}"
chart_dir="${HELM_CHART_PATH:-charts/openlane}"

# Install dependencies
install_dependencies

# Create workspace
work=$(create_temp_workspace)

# Log execution context
log_execution_context "Draft PR Automation"

# Check if we're in a PR context
if [[ -z "${BUILDKITE_PULL_REQUEST:-}" || "${BUILDKITE_PULL_REQUEST}" == "false" ]]; then
  echo "ℹ️  Not running in PR context, skipping draft PR creation"
  exit 0
fi

# Exit early if this PR does not modify configuration
base_branch="${BUILDKITE_PULL_REQUEST_BASE_BRANCH:-main}"
if ! pr_has_config_changes "$base_branch" "$BUILDKITE_BUILD_CHECKOUT_PATH" "config"; then
  echo "ℹ️  No configuration changes detected in this PR compared to $base_branch"
  exit 0
fi

# Clone the target repository
echo "Cloning repository..."
if ! git clone "$repo" "$work"; then
  echo "❌ Failed to clone $repo" >&2
  exit 1
fi

cd "$work"

# Create a draft branch name based on the core PR (consistent naming)
core_pr_number="${BUILDKITE_PULL_REQUEST}"
draft_branch="draft-core-pr-${core_pr_number}"

# Check if there's already a draft PR for this core PR number
echo "Checking for existing draft PR for core PR #${core_pr_number}..."
existing_pr_number=$(find_existing_draft_pr "$repo" "$core_pr_number")

if [[ -n "$existing_pr_number" ]]; then
  echo "📄 Found existing draft PR #${existing_pr_number} for core PR #${core_pr_number}"
  # Get the branch name from the existing PR
  existing_branch=$(get_pr_branch "$existing_pr_number" "$repo")
  echo "📄 Using existing branch: $existing_branch"
  git fetch origin "$existing_branch"
  git checkout "$existing_branch"
  draft_branch="$existing_branch"  # Use the existing branch name
else
  echo "🆕 No existing draft PR found, creating new branch: $draft_branch"
  git checkout -b "$draft_branch"
fi

# Track what changes we make
changes_made=false
change_summary=""

# Apply configuration changes using library functions
config_changes=$(apply_helm_config_changes \
  "$BUILDKITE_BUILD_CHECKOUT_PATH/config" \
  "$chart_dir")

if [[ -n "$config_changes" ]]; then
  changes_made=true
  # Convert newlines to <br/> for HTML formatting in PR body
  change_summary=$(echo "$config_changes" | sed 's/$/\\n/g' | tr '\n' ' ' | sed 's/\\n/<br\/>/g')
fi

# Check if we have any changes to commit
if [[ "$changes_made" == "false" ]]; then
  echo "ℹ️  No configuration changes detected, skipping draft PR creation"
  exit 0
fi

echo "📝 Configuration changes detected, creating draft PR"
echo -e "Summary:$change_summary"

# Source helm documentation utilities from core repo
source "${BUILDKITE_BUILD_CHECKOUT_PATH}/.buildkite/helm-docs-utils.sh"

# Generate documentation before committing
generate_docs_and_commit

# Setup git configuration
setup_git_user

# Create comprehensive commit message
build_info="- Core PR: #${core_pr_number}
- Core Branch: ${BUILDKITE_BRANCH}
- Build Number: ${BUILDKITE_BUILD_NUMBER}
- Source Commit: ${BUILDKITE_COMMIT:0:8}"

change_details="Changes made:$change_summary

⚠️  DO NOT MERGE until the core PR is merged first."

create_commit "draft" "preview config changes from core PR #${core_pr_number}" "$change_details" "$build_info"

# Push and create/update draft PR
echo "🚀 Pushing draft branch..."
if safe_push_branch "$draft_branch" true; then
  pr_body=$(generate_draft_pr_body "$core_pr_number" "$change_summary")

  # Create or update the PR
  if [[ -n "$existing_pr_number" ]]; then
    echo "📝 Updating existing draft PR #${existing_pr_number}..."
    pr_url=$(get_pr_url "$existing_pr_number" "$repo")
    if update_pr "$existing_pr_number" "$repo" "🚧 DRAFT: Config changes from core PR #${core_pr_number}" "$pr_body"; then
      echo "✅ Draft pull request updated successfully: $pr_url"
    else
      echo "⚠️  Failed to update existing PR, but push succeeded"
    fi
  else
    echo "Creating new draft pull request..."
    if create_draft_pr "$repo" "$draft_branch" "🚧 DRAFT: Config changes from core PR #${core_pr_number}" "$pr_body"; then
      pr_url=$(get_pr_url "$draft_branch" "$repo")
      echo "✅ Draft pull request created successfully: $pr_url"
    else
      echo "❌ Failed to create draft pull request"
      exit 1
    fi
  fi

  # Store the draft PR URL for linking
  echo "$pr_url" > "${BUILDKITE_BUILD_CHECKOUT_PATH}/.draft_pr_url"

  # Store metadata for later use
  cat > "${BUILDKITE_BUILD_CHECKOUT_PATH}/.draft_pr_metadata" << EOF
DRAFT_PR_URL=$pr_url
DRAFT_BRANCH=$draft_branch
CORE_PR_NUMBER=$core_pr_number
INFRA_REPO=$repo
EOF

  echo "📝 Draft PR metadata saved for linking and post-merge processing"

  # Add comment to core PR linking to the draft infrastructure PR
  echo "💬 Adding comment to core PR #${core_pr_number}..."
  comment_body="## 🔧 Configuration Changes Detected

This PR contains changes that will affect the Helm chart configuration. A draft infrastructure PR has been automatically created to preview these changes:

**📋 Draft PR**: $pr_url

### Changes Preview:
$change_summary

After this core PR merges, the draft will be closed and a new infrastructure PR will be opened from \`main\` for review."

  if gh pr comment "${core_pr_number}" \
    --repo "theopenlane/core" \
    --body "$comment_body"; then
    echo "✅ Comment added to core PR successfully"
  else
    echo "⚠️  Failed to add comment to core PR (this won't affect the automation)"
  fi
else
  echo "❌ Failed to push draft branch"
  exit 1
fi

echo "🎉 Draft PR automation completed successfully"
