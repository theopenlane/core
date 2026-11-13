#!/bin/bash
set -euo pipefail

# Commits go work sync changes back to renovate branches
# Renovate only updates the go.mod it targets, so check:clean fails on the other workspace modules

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

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

git add -u
commit_staged_to_branch "chore: sync workspace modules"

# Fail on purpose so the wait step in pipeline.yaml stops this superseded build
echo "🔁 Committed go work sync changes, the new commit will run its own build"
exit 1
