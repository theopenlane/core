#!/bin/bash
set -euo pipefail

# Commits the patches that check:clean saved from the generate steps back to the branch
# Generate runs branch code, so it never sees GITHUB_TOKEN and only this step pushes

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

HEADLINE="chore: run generate"
PATCH_DIR="generated-patches"

git config --global --add safe.directory "${BUILDKITE_BUILD_CHECKOUT_PATH:-/workdir}"

if ! buildkite-agent artifact download "${PATCH_DIR}/*.patch" .; then
  echo "✅ Generate left no changes"
  exit 0
fi

# a generator that is not deterministic would otherwise commit on every build
if [[ "$(git log -1 --format=%s)" == "$HEADLINE" ]]; then
  echo "🚨 Generate still changes files after a generated commit, run task generate locally"
  exit 1
fi

for patch in "${PATCH_DIR}"/*.patch; do
  echo "--- applying ${patch}"
  git apply --3way "$patch"
done

if git diff --cached --quiet; then
  echo "✅ Patches were already applied"
  exit 0
fi

git --no-pager diff --cached --stat

commit_staged_to_branch "$HEADLINE"

# Fail on purpose so this superseded build stops, the new commit runs its own build
echo "🔁 Committed generated changes, the new commit will run its own build"
exit 1
