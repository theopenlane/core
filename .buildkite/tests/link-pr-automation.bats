#!/usr/bin/env bats

setup() {
    export TEST_TEMP_DIR="$(mktemp -d)"
    export BUILDKITE_BUILD_CHECKOUT_PATH="$TEST_TEMP_DIR"
    export PATH="$TEST_TEMP_DIR:$PATH"
    # Stub gh to avoid installing github-cli
    cat > "$TEST_TEMP_DIR/gh" <<'EOS'
#!/bin/bash
exit 0
EOS
    chmod +x "$TEST_TEMP_DIR/gh"
}

teardown() {
    rm -rf "$TEST_TEMP_DIR"
}

@test "skips when not in PR context" {
    unset BUILDKITE_PULL_REQUEST
    run bash .buildkite/link-pr-automation.sh
    [ "$status" -eq 0 ]
    [[ "$output" == *"Not running in PR context"* ]]
}

@test "skips when metadata missing" {
    export BUILDKITE_PULL_REQUEST=5
    run bash .buildkite/link-pr-automation.sh
    [ "$status" -eq 0 ]
    [[ "$output" == *"No draft PR metadata found"* ]]
}

@test "comments on PRs when metadata present" {
    export BUILDKITE_PULL_REQUEST=7
    cat > "$TEST_TEMP_DIR/.draft_pr_metadata" <<EOM
DRAFT_PR_URL=https://github.com/theopenlane/openlane-infra/pull/42
EOM
    cat > "$TEST_TEMP_DIR/gh" <<'SCRIPT'
#!/bin/bash
 echo "gh $@" >> "$TEST_TEMP_DIR/gh.log"
 exit 0
SCRIPT
    chmod +x "$TEST_TEMP_DIR/gh"

    run bash .buildkite/link-pr-automation.sh
    [ "$status" -eq 0 ]
    grep -q "gh pr comment 7" "$TEST_TEMP_DIR/gh.log"
    grep -q "gh pr comment 42" "$TEST_TEMP_DIR/gh.log"
}
