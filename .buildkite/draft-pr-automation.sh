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

# Create a draft branch name based on the core PR
core_pr_number="${BUILDKITE_PULL_REQUEST}"
draft_branch="draft-core-pr-${core_pr_number}-${BUILDKITE_BUILD_NUMBER}"

echo "Creating draft branch: $draft_branch"
git checkout -b "$draft_branch"

# Track what changes we make
changes_made=false
change_summary=""

# Use the helm-automation.sh functions by sourcing it
source "${BUILDKITE_BUILD_CHECKOUT_PATH}/.buildkite/helm-automation.sh"

echo "🔍 Checking for configuration changes..."

# Update Helm values.yaml using the function from helm-automation.sh
if merge_helm_values \
  "$BUILDKITE_BUILD_CHECKOUT_PATH/config/helm-values.yaml" \
  "$chart_dir/values.yaml" \
  "Helm values.yaml"; then
  changes_made=true
  change_summary+="\n- 🔄 Merged Helm values.yaml"
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
      change_summary+="\n- 🔐 Updated External Secrets templates"
    fi
  else
    echo "Creating External Secrets templates"
    mkdir -p "$(dirname "$chart_dir/templates/external-secrets")"
    cp -r "$BUILDKITE_BUILD_CHECKOUT_PATH/config/external-secrets" "$chart_dir/templates/external-secrets"
    git add "$chart_dir/templates/external-secrets"
    changes_made=true
    change_summary+="\n- 🆕 Created External Secrets templates"
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
      change_summary+="\n- ✅ Updated ConfigMap template"
    fi
  else
    echo "Creating ConfigMap template"
    mkdir -p "$(dirname "$target")"
    cp "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" "$target"
    git add "$target"
    changes_made=true
    change_summary+="\n- ✨ Created ConfigMap template"
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

# Push and create draft PR
echo "🚀 Pushing draft branch and creating draft PR..."
if git push origin "$draft_branch"; then
  pr_body="## 🚧 DRAFT: Preview Config Changes from Core PR

⚠️ **This is a DRAFT PR** - DO NOT MERGE until core PR #${core_pr_number} is merged first.

### 🔗 Related Core PR
- **Core PR**: [#${core_pr_number}](https://github.com/theopenlane/core/pull/${core_pr_number})
- **Core Branch**: \`${BUILDKITE_BRANCH}\`

### 📋 Proposed Changes:$change_summary

### 🔧 Build Information:
- **Build Number**: ${BUILDKITE_BUILD_NUMBER}
- **Source Commit**: [\`${BUILDKITE_COMMIT:0:8}\`](https://github.com/theopenlane/core/commit/${BUILDKITE_COMMIT})

### 🔍 What This Shows:
- **Helm Values**: How configuration schema and defaults will change
- **External Secrets**: How secret management templates will be updated
- **ConfigMaps**: How non-sensitive configuration templates will change

### 📋 Next Steps:
1. **Review**: Review the proposed changes in this PR
2. **Core PR Review**: Complete review and merge of core PR #${core_pr_number}
3. **Auto-Update**: This PR will automatically convert from draft to ready after core merge
4. **Final Review**: Conduct final review and merge this PR

This PR was automatically generated to provide visibility into configuration changes before they're finalized."

  echo "Creating draft pull request..."
  if pr_url=$(gh pr create \
    --repo "$repo" \
    --head "$draft_branch" \
    --title "🚧 DRAFT: Config changes from core PR #${core_pr_number}" \
    --body "$pr_body" \
    --draft \
    --json url \
    --jq '.url'); then
    echo "✅ Draft pull request created successfully: $pr_url"

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

  else
    echo "❌ Failed to create draft pull request"
    exit 1
  fi
else
  echo "❌ Failed to push draft branch"
  exit 1
fi

echo "🎉 Draft PR automation completed successfully"