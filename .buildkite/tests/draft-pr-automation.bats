#!/usr/bin/env bats

setup() {
    export TEST_TEMP_DIR="$(mktemp -d)"
    export PATH="$TEST_TEMP_DIR:$PATH"
    export BUILDKITE_BUILD_CHECKOUT_PATH="$TEST_TEMP_DIR/checkout"
    mkdir -p "$BUILDKITE_BUILD_CHECKOUT_PATH/config"
    mkdir -p "$BUILDKITE_BUILD_CHECKOUT_PATH/.buildkite"
    touch "$BUILDKITE_BUILD_CHECKOUT_PATH/.buildkite/helm-docs-utils.sh"
    export HELM_CHART_REPO="theopenlane/openlane-infra"
    export HELM_CHART_PATH="charts/openlane"

    create_stubs
}

teardown() {
    rm -rf "$TEST_TEMP_DIR"
}

create_stubs() {
    cat > "$TEST_TEMP_DIR/git" <<'SCRIPT'
#!/bin/bash
echo "git $@" >> "$TEST_TEMP_DIR/git.log"
if [[ "$1" == "clone" ]]; then
  mkdir -p "$3"
fi
SCRIPT
    chmod +x "$TEST_TEMP_DIR/git"

    cat > "$TEST_TEMP_DIR/gh" <<'SCRIPT'
#!/bin/bash
echo "gh $@" >> "$TEST_TEMP_DIR/gh.log"
SCRIPT
    chmod +x "$TEST_TEMP_DIR/gh"

    cat > "$TEST_TEMP_DIR/stubs.sh" <<STUB
temp_workspace="$TEST_TEMP_DIR/workspace"
install_dependencies() { :; }
create_temp_workspace() { mkdir -p "\$temp_workspace"; echo "\$temp_workspace"; }
log_execution_context() { :; }
pr_has_config_changes() { if [ "\${PR_HAS_CHANGES:-0}" -eq 0 ]; then return 0; else return 1; fi }
apply_helm_config_changes() { echo "\${APPLY_CHANGES_OUTPUT:-}"; }
generate_docs_and_commit() { :; }
setup_git_user() { :; }
create_commit() { echo "create_commit \$@" >> "$TEST_TEMP_DIR/log"; }
safe_push_branch() { echo "safe_push_branch \$1 \$2" >> "$TEST_TEMP_DIR/log"; return 0; }
create_draft_pr() {
  echo "create_draft_pr \$@" >> "$TEST_TEMP_DIR/log"
  if [ "\${CREATE_DRAFT_FAIL:-0}" -eq 1 ]; then
    return 1
  fi
}
update_pr() { echo "update_pr \$@" >> "$TEST_TEMP_DIR/log"; }
get_pr_url() { echo "https://example.com/pr/\$1"; }
find_existing_draft_pr() {
  if [ -n "\${EXISTING_PR_SEQUENCE:-}" ]; then
    state_file="$TEST_TEMP_DIR/find_existing_pr_count"
    count=0
    if [ -f "\$state_file" ]; then
      count=\$(cat "\$state_file")
    fi
    count=\$((count + 1))
    echo "\$count" > "\$state_file"

    IFS=',' read -r -a sequence <<< "\$EXISTING_PR_SEQUENCE"
    index=\$((count - 1))
    echo "\${sequence[\$index]:-}"
    return 0
  fi

  echo "\${EXISTING_PR:-}"
}
get_pr_branch() { echo "\${EXISTING_BRANCH:-}"; }
STUB
}

@test "skips when not in PR context" {
    unset BUILDKITE_PULL_REQUEST
    DRAFT_PR_STUB="$TEST_TEMP_DIR/stubs.sh" run bash .buildkite/draft-pr-automation.sh
    [ "$status" -eq 0 ]
    [[ "$output" == *"Not running in PR context"* ]]
}

@test "skips when no config changes" {
    export BUILDKITE_PULL_REQUEST=5
    export PR_HAS_CHANGES=1
    DRAFT_PR_STUB="$TEST_TEMP_DIR/stubs.sh" run bash .buildkite/draft-pr-automation.sh
    [ "$status" -eq 0 ]
    [[ "$output" == *"No configuration changes detected"* ]]
}

@test "reuses existing PR when create fails because one is already open" {
    export BUILDKITE_PULL_REQUEST=9
    export BUILDKITE_BRANCH="feature/test-draft-reuse"
    export BUILDKITE_BUILD_NUMBER=123
    export BUILDKITE_COMMIT="0123456789abcdef0123456789abcdef01234567"
    export PR_HAS_CHANGES=0
    export APPLY_CHANGES_OUTPUT="- changed values"
    export CREATE_DRAFT_FAIL=1
    # First lookup: none, second lookup (after create failure): existing PR #42.
    export EXISTING_PR_SEQUENCE=",42"

    DRAFT_PR_STUB="$TEST_TEMP_DIR/stubs.sh" run bash .buildkite/draft-pr-automation.sh
    [ "$status" -eq 0 ]
    [[ "$output" == *"already exists"* ]]
    grep -q "create_draft_pr" "$TEST_TEMP_DIR/log"
    [ ! -f "$BUILDKITE_BUILD_CHECKOUT_PATH/.draft_pr_metadata" ]
    [ ! -f "$BUILDKITE_BUILD_CHECKOUT_PATH/.draft_pr_url" ]
}

@test "reuses existing draft PR without emitting link metadata or comments" {
    export BUILDKITE_PULL_REQUEST=11
    export BUILDKITE_BRANCH="feature/test-draft-existing"
    export BUILDKITE_BUILD_NUMBER=124
    export BUILDKITE_COMMIT="fedcba9876543210fedcba9876543210fedcba98"
    export PR_HAS_CHANGES=0
    export APPLY_CHANGES_OUTPUT="- changed values"
    export EXISTING_PR=88
    export EXISTING_BRANCH="draft-core-pr-11"

    DRAFT_PR_STUB="$TEST_TEMP_DIR/stubs.sh" run bash .buildkite/draft-pr-automation.sh
    [ "$status" -eq 0 ]
    [[ "$output" == *"already exists"* ]]
    [ ! -f "$TEST_TEMP_DIR/git.log" ]
    if [ -f "$TEST_TEMP_DIR/log" ]; then
        ! grep -q "update_pr" "$TEST_TEMP_DIR/log"
        ! grep -q "create_draft_pr" "$TEST_TEMP_DIR/log"
        ! grep -q "safe_push_branch" "$TEST_TEMP_DIR/log"
    fi
    [ ! -f "$BUILDKITE_BUILD_CHECKOUT_PATH/.draft_pr_metadata" ]
    [ ! -f "$BUILDKITE_BUILD_CHECKOUT_PATH/.draft_pr_url" ]
}

@test "writes link metadata when a new draft PR is created" {
    export BUILDKITE_PULL_REQUEST=12
    export BUILDKITE_BRANCH="feature/test-draft-new"
    export BUILDKITE_BUILD_NUMBER=125
    export BUILDKITE_COMMIT="00112233445566778899aabbccddeeff00112233"
    export PR_HAS_CHANGES=0
    export APPLY_CHANGES_OUTPUT="- changed values"
    unset EXISTING_PR
    unset EXISTING_PR_SEQUENCE

    DRAFT_PR_STUB="$TEST_TEMP_DIR/stubs.sh" run bash .buildkite/draft-pr-automation.sh
    [ "$status" -eq 0 ]
    [ -f "$BUILDKITE_BUILD_CHECKOUT_PATH/.draft_pr_metadata" ]
    [ -f "$BUILDKITE_BUILD_CHECKOUT_PATH/.draft_pr_url" ]
    grep -q "DRAFT_PR_URL=https://example.com/pr/" "$BUILDKITE_BUILD_CHECKOUT_PATH/.draft_pr_metadata"
}
