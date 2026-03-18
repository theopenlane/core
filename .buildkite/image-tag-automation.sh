#!/bin/bash
set -euo pipefail

# Source shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"
source "${SCRIPT_DIR}/lib/templates.sh"

YQ_VERSION=${YQ_VERSION:-4.45.4}
repo="${HELM_CHART_REPO}"
chart_dir="${HELM_CHART_PATH:-charts/openlane}"

# Install dependencies using shared library function
install_dependencies

echo "=== Image Tag Automation ==="
echo "Repository: $repo"
echo "Chart directory: $chart_dir"
echo "Release Tag: ${BUILDKITE_TAG}"
echo "Build: ${BUILDKITE_BUILD_NUMBER}"

# Verify this is a release build
if [[ -z "${BUILDKITE_TAG:-}" ]]; then
  echo "❌ No release tag found - this automation only runs for tagged releases"
  exit 1
fi

# Import slack utility functions
source "${SCRIPT_DIR}/slack-utils.sh"

work=$(create_temp_workspace)

# Clone the target repository
echo "Cloning repository..."
if ! git clone "$repo" "$work"; then
  echo "❌ Failed to clone $repo" >&2
  exit 1
fi

cd "$work"

# Create a release branch name
release_branch="release-${BUILDKITE_TAG}"

echo "🚀 Creating release branch: $release_branch"
git checkout -b "$release_branch"

# Track changes
changes_made=false
change_summary=""

# Update the image tag in values.yaml
values_file="$chart_dir/values.yaml"
if [[ -f "$values_file" ]]; then
  current_tag=$(yq e '.openlane.image.tag' "$values_file")
  new_tag="${BUILDKITE_TAG}"

  echo "📦 Updating image tag: $current_tag -> $new_tag"

  # Update the image tag
  yq e -i ".openlane.image.tag = \"$new_tag\"" "$values_file"
  git add "$values_file"
  changes_made=true
  change_summary+="\n- 📦 Updated image tag from $current_tag to $new_tag"

  echo "✅ Image tag updated successfully"
else
  echo "❌ Values file not found: $values_file"
  exit 1
fi

# Update chart version for release
chart_file="$chart_dir/Chart.yaml"
if [[ -f "$chart_file" ]]; then
  current_version=$(grep '^version:' "$chart_file" | awk '{print $2}')

  # For releases, we increment minor version (major.minor.0)
  # Parse the tag to determine new chart version
  if [[ "$new_tag" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    major="${BASH_REMATCH[1]}"
    minor="${BASH_REMATCH[2]}"
    patch="${BASH_REMATCH[3]}"
    new_chart_version="$major.$minor.$patch"
  else
    # Fallback: increment current chart version
    IFS='.' read -r maj min pat <<< "$current_version"
    new_patch=$((pat+1))
    new_chart_version="$maj.$min.$new_patch"
  fi

  echo "📈 Updating chart version: $current_version -> $new_chart_version"
  sed -i -E "s/^version:.*/version: $new_chart_version/" "$chart_file"
  git add "$chart_file"
  changes_made=true
  change_summary+="\n- 📈 Updated chart version to $new_chart_version"
fi

# Update app version to match the release tag
if [[ -f "$chart_file" ]]; then
  current_app_version=$(grep '^appVersion:' "$chart_file" | awk '{print $2}' | tr -d '"' || echo "")
  new_app_version="${BUILDKITE_TAG}"

  if [[ "$current_app_version" != "$new_app_version" ]]; then
    echo "🔄 Updating app version: $current_app_version -> $new_app_version"

    if grep -q '^appVersion:' "$chart_file"; then
      sed -i -E "s/^appVersion:.*/appVersion: \"$new_app_version\"/" "$chart_file"
    else
      # Add appVersion if it doesn't exist
      echo "appVersion: \"$new_app_version\"" >> "$chart_file"
    fi

    git add "$chart_file"
    change_summary+="\n- 🔄 Updated app version to $new_app_version"
  fi
fi

# Check if we have any changes to commit
if [[ "$changes_made" == "false" ]]; then
  echo "ℹ️  No image tag changes needed (already up to date)"
  exit 0
fi

# Normalize escaped newlines for markdown/notification output.
formatted_change_summary=$(printf '%b' "$change_summary")

echo "📝 Release changes detected, creating PR"
echo "Summary:${formatted_change_summary}"

# Source helm documentation utilities from core repo
source "${BUILDKITE_BUILD_CHECKOUT_PATH}/.buildkite/helm-docs-utils.sh"

# Generate documentation before committing
generate_docs_and_commit

# Configure git using shared library function
setup_git_user

# Create comprehensive commit message
template_dir=$(get_template_dir)
commit_message=$(load_template "${template_dir}/github/release-commit.md" \
    "RELEASE_TAG=${BUILDKITE_TAG}" \
    "CHANGE_SUMMARY=${formatted_change_summary}" \
    "BUILD_NUMBER=${BUILDKITE_BUILD_NUMBER}" \
    "SOURCE_COMMIT_SHORT=${BUILDKITE_COMMIT:0:8}")

# Use shared function to create commit and push
create_commit "release" "update to ${BUILDKITE_TAG}" "Release update for ${BUILDKITE_TAG}" "- Source Commit: ${BUILDKITE_COMMIT:0:8}
- Generated: $(date +'%Y-%m-%d %H:%M:%S UTC')"

# Push using shared function
echo "🚀 Pushing release branch..."
if safe_push_branch "$release_branch"; then
  pr_body=$(load_template "${template_dir}/github/release-pr.md" \
      "RELEASE_TAG=${BUILDKITE_TAG}" \
      "CHANGE_SUMMARY=${formatted_change_summary}" \
      "BUILD_NUMBER=${BUILDKITE_BUILD_NUMBER}" \
      "SOURCE_COMMIT_SHORT=${BUILDKITE_COMMIT:0:8}" \
      "SOURCE_COMMIT_FULL=${BUILDKITE_COMMIT}" \
      "SOURCE_BRANCH=${BUILDKITE_BRANCH:-main}")

  echo "Creating release pull request..."
  if gh pr create \
    --repo "$repo" \
    --head "$release_branch" \
    --title "🚀 Release ${BUILDKITE_TAG}" \
    --body "$pr_body"; then
    pr_url=$(gh pr view "$release_branch" --repo "$repo" --json url --jq '.url' 2>/dev/null || echo "https://github.com/$repo/pull/new/$release_branch")
    echo "✅ Release pull request created successfully: $pr_url"

    # Send Slack notification for release deployment
    send_release_notification "$pr_url" "${BUILDKITE_TAG}" "$formatted_change_summary"
  else
    echo "❌ Failed to create release pull request"
    exit 1
  fi
else
  echo "❌ Failed to push release branch"
  exit 1
fi

echo "🎉 Image tag automation completed successfully"
