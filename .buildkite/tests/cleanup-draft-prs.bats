#!/usr/bin/env bats

# Tests for cleanup-draft-prs.sh flow

setup() {
    export TEST_TEMP_DIR="$(mktemp -d)"
    export HELM_CHART_REPO="theopenlane/openlane-infra"
    export PR_ACTIVITY_THRESHOLD_SECONDS=2592000

    mkdir -p "$TEST_TEMP_DIR"

    create_stub_functions
}

teardown() {
    rm -rf "$TEST_TEMP_DIR"
}

create_stub_functions() {
    cat > "$TEST_TEMP_DIR/stubs.sh" <<'STUB'
install_dependencies() { :; }
validate_build_context() { return 0; }
find_draft_prs() { echo "1:test-branch:\ud83d\udea7 DRAFT: Config changes from core PR #10"; }
check_core_pr_status() { echo '{"state":"MERGED","title":"Test","updatedAt":"2025-07-15T00:00:00Z"}'; }
is_recent_pr_activity() { return 0; }
close_pr() { echo "close_pr $1" >> "$TEST_TEMP_DIR/log"; }
safe_delete_branch() { echo "delete_branch $1 $2" >> "$TEST_TEMP_DIR/log"; }
extract_core_pr_number() { echo 10; }
generate_closure_comment() { echo "comment"; }
STUB
    source "$TEST_TEMP_DIR/stubs.sh"
}

@test "draft PR is closed when core PR merged" {
    CLEANUP_PR_STUB="$TEST_TEMP_DIR/stubs.sh" run bash .buildkite/cleanup-draft-prs.sh
    [ "$status" -eq 0 ]
    grep -q "close_pr 1" "$TEST_TEMP_DIR/log"
    grep -q "delete_branch test-branch theopenlane/openlane-infra" "$TEST_TEMP_DIR/log"
}

@test "draft PR kept when core PR open" {
    cat > "$TEST_TEMP_DIR/stubs.sh" <<'STUB'
install_dependencies() { :; }
validate_build_context() { return 0; }
find_draft_prs() { echo "2:test-branch:\ud83d\udea7 DRAFT: Config changes from core PR #20"; }
check_core_pr_status() { echo '{"state":"OPEN","title":"Test","updatedAt":"2025-07-15T00:00:00Z"}'; }
close_pr() { echo "close_pr $1" >> "$TEST_TEMP_DIR/log"; }
safe_delete_branch() { echo "delete_branch $1 $2" >> "$TEST_TEMP_DIR/log"; }
extract_core_pr_number() { echo 20; }
STUB
    source "$TEST_TEMP_DIR/stubs.sh"

    CLEANUP_PR_STUB="$TEST_TEMP_DIR/stubs.sh" run bash .buildkite/cleanup-draft-prs.sh
    [ "$status" -eq 0 ]
    if [ -f "$TEST_TEMP_DIR/log" ]; then
        ! grep -q "close_pr" "$TEST_TEMP_DIR/log"
    fi
}

@test "draft PR closed when core PR missing" {
    cat > "$TEST_TEMP_DIR/stubs.sh" <<'STUB'
install_dependencies() { :; }
validate_build_context() { return 0; }
find_draft_prs() { echo "3:test-branch:\ud83d\udea7 DRAFT: Config changes from core PR #30"; }
check_core_pr_status() { echo ""; }
close_pr() { echo "close_pr $1" >> "$TEST_TEMP_DIR/log"; }
safe_delete_branch() { echo "delete_branch $1 $2" >> "$TEST_TEMP_DIR/log"; }
extract_core_pr_number() { echo 30; }
generate_closure_comment() { echo "comment"; }
STUB
    source "$TEST_TEMP_DIR/stubs.sh"

    CLEANUP_PR_STUB="$TEST_TEMP_DIR/stubs.sh" run bash .buildkite/cleanup-draft-prs.sh
    [ "$status" -eq 0 ]
    grep -q "close_pr 3" "$TEST_TEMP_DIR/log"
}
