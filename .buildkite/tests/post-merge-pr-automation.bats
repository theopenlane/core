#!/usr/bin/env bats

setup() {
    export TEST_TEMP_DIR
    TEST_TEMP_DIR="$(mktemp -d)"
    export HELM_CHART_REPO="theopenlane/openlane-infra"
}

teardown() {
    rm -rf "$TEST_TEMP_DIR"
}

@test "closes draft PR when core PR is merged" {
    cat > "$TEST_TEMP_DIR/stubs.sh" <<'STUB'
install_dependencies() { :; }
validate_build_context() { return 0; }
find_draft_prs() { echo "5:draft-core-pr-42:🚧 DRAFT: Config changes from core PR #42"; }
check_core_pr_status() { echo '{"state":"MERGED","updatedAt":"2026-01-01T00:00:00Z"}'; }
generate_closure_comment() { echo "comment"; }
close_pr() { echo "close_pr $1" >> "$TEST_TEMP_DIR/log"; return 0; }
safe_delete_branch() { echo "delete_branch $1" >> "$TEST_TEMP_DIR/log"; return 0; }
STUB

    POST_MERGE_PR_STUB="$TEST_TEMP_DIR/stubs.sh" run bash .buildkite/post-merge-pr-automation.sh
    [ "$status" -eq 0 ]
    grep -q "close_pr 5" "$TEST_TEMP_DIR/log"
    grep -q "delete_branch draft-core-pr-42" "$TEST_TEMP_DIR/log"
}

@test "keeps draft PR open when core PR is still open" {
    cat > "$TEST_TEMP_DIR/stubs.sh" <<'STUB'
install_dependencies() { :; }
validate_build_context() { return 0; }
find_draft_prs() { echo "6:draft-core-pr-43:🚧 DRAFT: Config changes from core PR #43"; }
check_core_pr_status() { echo '{"state":"OPEN","updatedAt":"2026-01-01T00:00:00Z"}'; }
generate_closure_comment() { echo "comment"; }
close_pr() { echo "close_pr $1" >> "$TEST_TEMP_DIR/log"; return 0; }
safe_delete_branch() { echo "delete_branch $1" >> "$TEST_TEMP_DIR/log"; return 0; }
STUB

    POST_MERGE_PR_STUB="$TEST_TEMP_DIR/stubs.sh" run bash .buildkite/post-merge-pr-automation.sh
    [ "$status" -eq 0 ]
    if [ -f "$TEST_TEMP_DIR/log" ]; then
        ! grep -q "close_pr" "$TEST_TEMP_DIR/log"
    fi
}
