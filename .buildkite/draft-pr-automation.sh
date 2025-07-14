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

# Install yq if not available
if ! command -v yq >/dev/null 2>&1; then
    echo "Installing yq version ${YQ_VERSION}..."
    wget -q "https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_linux_amd64" -O /usr/local/bin/yq
    chmod +x /usr/local/bin/yq
    echo "✅ yq installed successfully"
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

# Check if there's already a draft PR for this core PR number
echo "Checking for existing draft PR for core PR #${core_pr_number}..."
existing_pr_number=$(gh pr list --repo "$repo" --state open --draft --json title,number --jq ".[] | select(.title | contains(\"core PR #${core_pr_number}\")) | .number" | head -1)

if [[ -n "$existing_pr_number" ]]; then
  echo "📄 Found existing draft PR #${existing_pr_number} for core PR #${core_pr_number}"
  # Get the branch name from the existing PR
  existing_branch=$(gh pr view "$existing_pr_number" --repo "$repo" --json headRefName --jq '.headRefName')
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

  # Merge coreConfiguration section specifically into openlane.coreConfiguration
  echo "  📋 Merging coreConfiguration section..."
  if yq e '.coreConfiguration' "$source" | grep -v "null" > /dev/null 2>&1; then
    core_config_section=$(yq e '.coreConfiguration' "$source")
    echo "$core_config_section" > /tmp/core-config-section.yaml
    yq e -i '.openlane.coreConfiguration = load("/tmp/core-config-section.yaml")' "$temp_merged"
    echo "  ✅ coreConfiguration merged successfully"
  else
    echo "  ⚠️  No coreConfiguration section found in source"
  fi

  # Merge externalSecrets section if it exists
  if yq e '.externalSecrets' "$source" | grep -v "null" > /dev/null 2>&1; then
    echo "  🔐 Merging external secrets configuration..."
    external_secrets_section=$(yq e '.externalSecrets' "$source")
    echo "$external_secrets_section" > /tmp/external-secrets-section.yaml
    yq e -i '.externalSecrets = load("/tmp/external-secrets-section.yaml")' "$temp_merged"
    echo "  ✅ externalSecrets merged successfully"
  fi

  # Replace target with merged content
  mv "$temp_merged" "$target"
  git add "$target"
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

# Source helm documentation utilities from core repo
source "${BUILDKITE_BUILD_CHECKOUT_PATH}/.buildkite/helm-docs-utils.sh"

# Generate documentation before committing
generate_docs_and_commit

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

  # Create or update the PR
  if [[ -n "$existing_pr_number" ]]; then
    echo "📝 Updating existing draft PR #${existing_pr_number}..."
    pr_url=$(gh pr view "$existing_pr_number" --repo "$repo" --json url --jq '.url')
    if gh pr edit "$existing_pr_number" \
      --repo "$repo" \
      --title "🚧 DRAFT: Config changes from core PR #${core_pr_number}" \
      --body "$pr_body"; then
      echo "✅ Draft pull request updated successfully: $pr_url"
    else
      echo "⚠️  Failed to update existing PR, but push succeeded"
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