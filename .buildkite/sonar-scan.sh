#!/bin/bash
#
# Runs the SonarScanner inside the sonarsource/sonar-scanner-cli image.
#
# Before scanning, strips coverage for files that sonar-project.properties
# excludes. Go coverage profiles include generated code, mocks, test utilities
# and cmd/, and Sonar logs a WARN for every one of those files otherwise.
#
# Override SONAR_COVERAGE_EXCLUDE_REGEX to change which files are dropped.

set -euo pipefail

MODULE_PREFIX='github\.com/theopenlane/core/v2'
DEFAULT_EXCLUDE="^${MODULE_PREFIX}/(cmd|internal/testutils|internal/ent/(generated|historygenerated|schema|entityops|csvgenerated|authzgenerated)|internal/graphapi/(generated|historygenerated|model|historymodel|testclient)|pkg/(objects|entitlements)/mocks)/"
EXCLUDE_REGEX="${SONAR_COVERAGE_EXCLUDE_REGEX:-$DEFAULT_EXCLUDE}"

shopt -s nullglob
for f in coverage-*.out; do
  before=$(wc -l < "$f")
  # grep exits 1 when every line is filtered out; that's not an error here
  grep -vE "$EXCLUDE_REGEX" "$f" > "$f.filtered" || true
  mv "$f.filtered" "$f"
  echo "filtered $f: $before -> $(wc -l < "$f") lines"
done

# hand off to the image's entrypoint so it still injects SONAR_TOKEN etc.
exec /usr/bin/entrypoint.sh sonar-scanner "$@"
