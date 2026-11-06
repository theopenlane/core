#!/bin/bash
set -euo pipefail

# Commits go work sync changes back to renovate branches
# Renovate only updates the go.mod it targets, so check:clean fails on the other workspace modules

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# GitHub signs commits made through this mutation, so CI needs no signing key
COMMIT_MUTATION='mutation($input: CreateCommitOnBranchInput!) { createCommitOnBranch(input: $input) { commit { url } } }'

# Prints each changed file as a createCommitOnBranch addition, piped because go.sum exceeds the argument size limit
changed_files_as_additions() {
  local path
  git diff --name-only | while read -r path; do
    base64 -w0 "$path" | jq -Rs --arg path "$path" '{path: $path, contents: .}'
  done
}

if [[ ! "${BUILDKITE_BRANCH:-}" =~ ^renovate/ ]]; then
  echo "ℹ️  Not a renovate branch (current: ${BUILDKITE_BRANCH:-unknown}), skipping"
  exit 0
fi

git config --global --add safe.directory "${BUILDKITE_BUILD_CHECKOUT_PATH:-/workdir}"

task tidy

if git diff --quiet; then
  echo "✅ Workspace modules are in sync"
  exit 0
fi

git status --porcelain

install_gh

# expectedHeadOid rejects the commit if the branch moved after this checkout
changed_files_as_additions | jq -s \
  --arg query "$COMMIT_MUTATION" \
  --arg repo "$(normalize_github_repo "$BUILDKITE_REPO")" \
  --arg branch "$BUILDKITE_BRANCH" \
  --arg head "$(git rev-parse HEAD)" \
  '{query: $query, variables: {input: {
    branch: {repositoryNameWithOwner: $repo, branchName: $branch},
    expectedHeadOid: $head,
    message: {headline: "chore: sync workspace modules"},
    fileChanges: {additions: .}
  }}}' | gh api graphql --input -

# Fail on purpose so the wait step in pipeline.yaml stops this superseded build
echo "🔁 Committed go work sync changes, the new commit will run its own build"
exit 1
