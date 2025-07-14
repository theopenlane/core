#!/bin/bash
set -euo pipefail

# Draft PR automation for config changes
# Creates draft PRs in openlane-infra when config changes are detected in core repo
# Links PRs together with comments for better visibility

YQ_VERSION=4.9.6
repo="${HELM_CHART_REPO}"
chart_dir="${HELM_CHART_PATH:-charts/openlane}"

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

# Import functions from helm-automation.sh
source "${BUILDKITE_BUILD_CHECKOUT_PATH}/.buildkite/helm-automation.sh" 2>/dev/null || {
  # If sourcing fails, define the functions we need
  function merge_helm_values() {
    local source="$1"
    local target="$2"
    local description="$3"

    if [[ ! -f "$source" ]]; then
      echo "⚠️  Source values file not found: $source"
      return 1
    fi

    echo "🔄 Merging $description"

    # Create backup of existing values
    if [[ -f "$target" ]]; then
      cp "$target" "${target}.backup"
    fi

    # Create temporary merged file
    local temp_merged="${target}.merged"

    if [[ -f "$target" ]]; then
      echo "  📋 Extracting core configuration from generated values..."
      docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e '.core' "$source" > /tmp/core-values.yaml

      echo "  🔀 Merging with existing chart values..."
      docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e '. as $target | load("/tmp/core-values.yaml") as $core | $target | .core = $core' "$target" > "$temp_merged"

      # Also merge any externalSecrets configuration if it exists
      if docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e '.externalSecrets' "$source" | grep -v "null" > /dev/null 2>&1; then
        echo "  🔐 Merging external secrets configuration..."
        docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e '.externalSecrets' "$source" > /tmp/external-secrets.yaml
        docker run --rm -v "${PWD}":/workdir mikefarah/yq:"${YQ_VERSION}" e '. as $target | load("/tmp/external-secrets.yaml") as $secrets | $target | .externalSecrets = $secrets' "$temp_merged" > "${temp_merged}.tmp"
        mv "${temp_merged}.tmp" "$temp_merged"
      fi
    else
      echo "  ✨ Creating new values file..."
      cp "$source" "$temp_merged"
    fi

    # Check if there are actual differences
    if [[ -f "$target" ]] && diff -q "$target" "$temp_merged" > /dev/null 2>&1; then
      echo "  ℹ️  No changes detected in $description"
      rm -f "$temp_merged" "${target}.backup" /tmp/core-values.yaml /tmp/external-secrets.yaml
      return 1
    fi

    # Apply the merged changes
    mv "$temp_merged" "$target"
    git add "$target"
    changes_made=true
    change_summary+="\\n- 🔄 Merged $description"

    # Cleanup
    rm -f "${target}.backup" /tmp/core-values.yaml /tmp/external-secrets.yaml

    return 0
  }

  function copy_and_track() {
    local source="$1"
    local target="$2"
    local description="$3"

    if [[ -f "$source" ]]; then
      # Check if target exists and has differences
      if [[ -f "$target" ]]; then
        if ! diff -q "$source" "$target" > /dev/null 2>&1; then
          echo "Updating $description"
          cp "$source" "$target"
          git add "$target"
          changes_made=true
          change_summary+="\\n- ✅ Updated $description"
          return 0
        fi
      else
        echo "Creating $description"
        mkdir -p "$(dirname "$target")"
        cp "$source" "$target"
        git add "$target"
        changes_made=true
        change_summary+="\\n- ✨ Created $description"
        return 0
      fi
    fi
    return 1
  }

  function copy_directory_and_track() {
    local source="$1"
    local target="$2"
    local description="$3"

    if [[ -d "$source" ]]; then
      # Check if target exists and has differences
      if [[ -d "$target" ]]; then
        if ! diff -r "$source" "$target" > /dev/null 2>&1; then
          echo "Updating $description"
          rm -rf "$target"
          mkdir -p "$(dirname "$target")"
          cp -r "$source" "$target"
          git add "$target"
          changes_made=true
          change_summary+="\\n- 🔐 Updated $description"
          return 0
        fi
      else
        echo "Creating $description"
        mkdir -p "$(dirname "$target")"
        cp -r "$source" "$target"
        git add "$target"
        changes_made=true
        change_summary+="\\n- 🆕 Created $description"
        return 0
      fi
    fi
    return 1
  }
}

# Apply the same changes as helm-automation.sh but for draft PR
echo "🔍 Checking for configuration changes..."

# Update Helm values.yaml (intelligent merging approach)
merge_helm_values \
  "$BUILDKITE_BUILD_CHECKOUT_PATH/config/helm-values.yaml" \
  "$chart_dir/values.yaml" \
  "Helm values.yaml"

# Update external secrets directory
copy_directory_and_track \
  "$BUILDKITE_BUILD_CHECKOUT_PATH/config/external-secrets" \
  "$chart_dir/templates/external-secrets" \
  "External Secrets templates"

# Legacy configmap support (for backward compatibility)
if [[ -f "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" ]]; then
  copy_and_track \
    "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" \
    "$chart_dir/templates/core-configmap.yaml" \
    "ConfigMap template"
fi

# Check if we have any changes to commit
if git diff --staged --quiet; then
  echo "ℹ️  No configuration changes detected, skipping draft PR creation"
  exit 0
fi

echo "📝 Configuration changes detected, creating draft PR"
echo -e "Summary:$change_summary"

# Configure git
git config user.email "bender@theopenlane.io"
git config user.name "theopenlane-bender"

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