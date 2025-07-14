#!/bin/bash
set -euo pipefail

# Draft PR automation for config changes
# Creates draft PRs in openlane-infra when config changes are detected in core repo

YQ_VERSION=${YQ_VERSION:-4.45.4}
repo="${HELM_CHART_REPO}"
chart_dir="${HELM_CHART_PATH:-charts/openlane}"

# Install gh if not available
if ! command -v gh >/dev/null 2>&1; then
    echo "Installing gh..."
    apk add --no-cache github-cli
fi

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "=== Draft PR Automation ===="
echo "Repository: $repo"
echo "Chart directory: $chart_dir"
echo "Core PR: ${BUILDKITE_PULL_REQUEST:-none}"
echo "Core Branch: ${BUILDKITE_BRANCH}"

# Check if we're in a PR context
if [[ -z "${BUILDKITE_PULL_REQUEST:-}" || "${BUILDKITE_PULL_REQUEST}" == "false" ]]; then
  echo "ℹ️  Not running in PR context, skipping draft PR creation"
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

# Check if draft branch already exists (from previous builds)
echo "Checking for existing draft branch: $draft_branch"
if git ls-remote --heads origin "$draft_branch" | grep -q "$draft_branch"; then
  echo "📄 Existing draft branch found, updating it..."
  git fetch origin "$draft_branch"
  git checkout "$draft_branch"
  git reset --hard "origin/$draft_branch"
else
  echo "🆕 Creating new draft branch: $draft_branch"
  git checkout -b "$draft_branch"
fi

# Track what changes we make
changes_made=false
change_summary=""

# Import only the merge function from helm-automation.sh, not the full script
merge_helm_values() {
  local source="$1"
  local target="$2"
  local description="$3"

  if [[ ! -f "$source" ]]; then
    echo "  ⚠️  Source file not found: $source"
    return 1
  fi

  if [[ ! -f "$target" ]]; then
    echo "  ⚠️  Target file not found: $target"
    return 1
  fi

  local temp_merged=$(mktemp)

  echo "  🔀 Merging $description..."
  # Copy target file as base
  cp "$target" "$temp_merged"

  # Perform deep merge of all sections from source into target
  echo "  📋 Performing deep merge of all configuration sections..."
  yq e -i '. *= load("'"$source"'")' "$temp_merged"

  # Replace target with merged content
  mv "$temp_merged" "$target"
  return 0
}

echo "🔍 Checking for configuration changes..."

# Update Helm values.yaml using the function from helm-automation.sh
if merge_helm_values \
  "$BUILDKITE_BUILD_CHECKOUT_PATH/config/helm-values.yaml" \
  "$chart_dir/values.yaml" \
  "Helm values.yaml"; then
  changes_made=true
  change_summary+="<br/>- 🔄 Merged Helm values.yaml"
fi

# Update external secrets directory
if [[ -d "$BUILDKITE_BUILD_CHECKOUT_PATH/config/external-secrets" ]]; then
  if [[ -d "$chart_dir/templates/external-secrets" ]]; then
    if ! diff -r "$BUILDKITE_BUILD_CHECKOUT_PATH/config/external-secrets" "$chart_dir/templates/external-secrets" > /dev/null 2>&1; then
      echo "Updating External Secrets templates"
      rm -rf "$chart_dir/templates/external-secrets"
      mkdir -p "$(dirname "$chart_dir/templates/external-secrets")"
      cp -r "$BUILDKITE_BUILD_CHECKOUT_PATH/config/external-secrets" "$chart_dir/templates/external-secrets"
      git add "$chart_dir/templates/external-secrets"
      changes_made=true
      change_summary+="<br/>- 🔐 Updated External Secrets templates"
    fi
  else
    echo "Creating External Secrets templates"
    mkdir -p "$(dirname "$chart_dir/templates/external-secrets")"
    cp -r "$BUILDKITE_BUILD_CHECKOUT_PATH/config/external-secrets" "$chart_dir/templates/external-secrets"
    git add "$chart_dir/templates/external-secrets"
    changes_made=true
    change_summary+="<br/>- 🆕 Created External Secrets templates"
  fi
fi

# Legacy configmap support (for backward compatibility)
if [[ -f "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" ]]; then
  target="$chart_dir/templates/core-configmap.yaml"
  if [[ -f "$target" ]]; then
    if ! diff -q "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" "$target" > /dev/null 2>&1; then
      echo "Updating ConfigMap template"
      cp "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" "$target"
      git add "$target"
      changes_made=true
      change_summary+="<br/>- ✅ Updated ConfigMap template"
    fi
  else
    echo "Creating ConfigMap template"
    mkdir -p "$(dirname "$target")"
    cp "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" "$target"
    git add "$target"
    changes_made=true
    change_summary+="<br/>- ✨ Created ConfigMap template"
  fi
fi

# Check if we have any changes to commit
if [[ "$changes_made" == "false" ]]; then
  echo "ℹ️  No configuration changes detected, skipping draft PR creation"
  exit 0
fi

echo "📝 Configuration changes detected, creating draft PR"
echo -e "Summary:$change_summary"

# Configure git
git config --local user.email "bender@theopenlane.io"
git config --local user.name "theopenlane-bender"

# Configure git to use GitHub token for authentication
git config --local url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# Create comprehensive commit message
commit_message="draft: preview config changes from core PR #${core_pr_number}

This is a DRAFT PR showing proposed changes from core repository PR #${core_pr_number}.

⚠️  DO NOT MERGE until the core PR is merged first.

Changes made:$change_summary

Build Information:
- Core PR: #${core_pr_number}
- Core Branch: ${BUILDKITE_BRANCH}
- Build Number: ${BUILDKITE_BUILD_NUMBER}
- Source Commit: ${BUILDKITE_COMMIT:0:8}"

git commit -m "$commit_message"

# Push and create/update draft PR
echo "🚀 Pushing draft branch..."
if git push -f origin "$draft_branch"; then
  pr_body="## 🚧 DRAFT: Config Changes from Core PR #${core_pr_number}

⚠️ **This is a DRAFT PR** - automatically converts to ready for review once [core PR #${core_pr_number}](https://github.com/theopenlane/core/pull/${core_pr_number}) is merged.

### Changes:
$change_summary

### Source:
- **Core PR**: [#${core_pr_number}](https://github.com/theopenlane/core/pull/${core_pr_number})
- **Branch**: [\`${BUILDKITE_BRANCH}\`](https://github.com/theopenlane/core/tree/${BUILDKITE_BRANCH})
- **Commit**: [\`${BUILDKITE_COMMIT:0:8}\`](https://github.com/theopenlane/core/commit/${BUILDKITE_COMMIT})"

  # Check if PR already exists for this branch
  echo "Checking for existing draft PR..."
  existing_pr_url=$(gh pr view "$draft_branch" --repo "$repo" --json url --jq '.url' 2>/dev/null || echo "")

  if [[ -n "$existing_pr_url" ]]; then
    echo "📝 Updating existing draft PR: $existing_pr_url"
    # Update the existing PR's body
    if gh pr edit "$draft_branch" \
      --repo "$repo" \
      --title "🚧 DRAFT: Config changes from core PR #${core_pr_number}" \
      --body "$pr_body"; then
      pr_url="$existing_pr_url"
      echo "✅ Draft pull request updated successfully: $pr_url"
    else
      echo "⚠️  Failed to update existing PR, but push succeeded"
      pr_url="$existing_pr_url"
    fi
  else
    echo "Creating new draft pull request..."
    if gh pr create \
      --repo "$repo" \
      --head "$draft_branch" \
      --title "🚧 DRAFT: Config changes from core PR #${core_pr_number}" \
      --body "$pr_body" \
      --draft; then
      pr_url=$(gh pr view "$draft_branch" --repo "$repo" --json url --jq '.url' 2>/dev/null || echo "https://github.com/$repo/pull/new/$draft_branch")
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

The infrastructure PR will automatically convert from draft to ready for review once this core PR is merged."

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