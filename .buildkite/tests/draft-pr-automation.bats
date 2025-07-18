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
create_draft_pr() { echo "create_draft_pr \$@" >> "$TEST_TEMP_DIR/log"; }
update_pr() { echo "update_pr \$@" >> "$TEST_TEMP_DIR/log"; }
get_pr_url() { echo "https://example.com/pr/\$1"; }
find_existing_draft_pr() { echo "\${EXISTING_PR:-}"; }
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
