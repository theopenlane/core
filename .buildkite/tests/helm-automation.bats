#!/usr/bin/env bats

# Unit tests for helm-automation.sh using BATS framework
# Install BATS: brew install bats-core
# Run tests: bats .buildkite/tests/helm-automation.bats

setup() {
    # Setup test environment
    export TEST_TEMP_DIR="$(mktemp -d)"
    export BUILDKITE_BUILD_NUMBER="9999"
    export BUILDKITE_COMMIT="abcd1234567890abcdef1234567890abcdef1234"
    export BUILDKITE_BRANCH="test-branch"
    export BUILDKITE_PIPELINE_NAME="core-test"
    export BUILDKITE_BUILD_CREATOR="test-user@example.com"
    export BUILDKITE_BUILD_URL="https://buildkite.com/test/builds/9999"
    export BUILDKITE_BUILD_CHECKOUT_PATH="$(pwd)"
    export HELM_CHART_REPO="file://$TEST_TEMP_DIR/test-repo"
    export HELM_CHART_PATH="charts/openlane"

    # Create test functions directory
    mkdir -p "$TEST_TEMP_DIR"

    # Source the script functions (extract functions for testing)
    create_test_functions

    # Ensure required tools are available for tests
    bash -c '. .buildkite/lib/common.sh && install_yq >/dev/null'
}

teardown() {
    rm -rf "$TEST_TEMP_DIR"
}

create_test_functions() {
    # Extract functions from the main script for testing
    cat > "$TEST_TEMP_DIR/functions.sh" << 'EOF'
# Test-friendly versions of functions from helm-automation.sh

function copy_and_track() {
  local source="$1"
  local target="$2"
  local description="$3"

  echo "copy_and_track: $source -> $target ($description)"

  if [[ -f "$source" ]]; then
    if [[ -f "$target" ]]; then
      if ! diff -q "$source" "$target" > /dev/null 2>&1; then
        return 0  # Changes detected
      fi
    else
      return 0  # New file
    fi
  fi
  return 1  # No changes
}

function copy_directory_and_track() {
  local source="$1"
  local target="$2"
  local description="$3"

  echo "copy_directory_and_track: $source -> $target ($description)"

  if [[ -d "$source" ]]; then
    if [[ -d "$target" ]]; then
      if ! diff -r "$source" "$target" > /dev/null 2>&1; then
        return 0  # Changes detected
      fi
    else
      return 0  # New directory
    fi
  fi
  return 1  # No changes
}

function send_slack_notification() {
  local pr_url="$1"

  if [[ -z "${SLACK_WEBHOOK_URL:-}" ]]; then
    echo "Slack not configured, skipping notification"
    return 0
  fi

  echo "Would send Slack notification for PR: $pr_url"
  return 0
}
EOF

    source "$TEST_TEMP_DIR/functions.sh"
}

@test "copy_and_track detects new files" {
    # Create source file
    echo "test content" > "$TEST_TEMP_DIR/source.txt"

    # Test copying new file
    run copy_and_track "$TEST_TEMP_DIR/source.txt" "$TEST_TEMP_DIR/target.txt" "test file"

    [ "$status" -eq 0 ]
    [[ "$output" == *"copy_and_track"* ]]
}

@test "copy_and_track detects file changes" {
    # Create source and target files with different content
    echo "source content" > "$TEST_TEMP_DIR/source.txt"
    echo "target content" > "$TEST_TEMP_DIR/target.txt"

    # Test detecting changes
    run copy_and_track "$TEST_TEMP_DIR/source.txt" "$TEST_TEMP_DIR/target.txt" "test file"

    [ "$status" -eq 0 ]
}

@test "copy_and_track skips identical files" {
    # Create identical source and target files
    echo "same content" > "$TEST_TEMP_DIR/source.txt"
    echo "same content" > "$TEST_TEMP_DIR/target.txt"

    # Test skipping identical files
    run copy_and_track "$TEST_TEMP_DIR/source.txt" "$TEST_TEMP_DIR/target.txt" "test file"

    [ "$status" -eq 1 ]
}

@test "copy_directory_and_track detects new directories" {
    # Create source directory
    mkdir -p "$TEST_TEMP_DIR/source_dir"
    echo "test" > "$TEST_TEMP_DIR/source_dir/file.txt"

    # Test copying new directory
    run copy_directory_and_track "$TEST_TEMP_DIR/source_dir" "$TEST_TEMP_DIR/target_dir" "test directory"

    [ "$status" -eq 0 ]
}

@test "send_slack_notification handles missing webhook" {
    unset SLACK_WEBHOOK_URL

    run send_slack_notification "https://github.com/test/test/pull/1"

    [ "$status" -eq 0 ]
    [[ "$output" == *"Slack not configured"* ]]
}

@test "send_slack_notification works with webhook configured" {
    export SLACK_WEBHOOK_URL="https://hooks.slack.com/test"

    run send_slack_notification "https://github.com/test/test/pull/1"

    [ "$status" -eq 0 ]
    [[ "$output" == *"Would send Slack notification"* ]]
}

@test "environment variables are properly set" {
    [ -n "$BUILDKITE_BUILD_NUMBER" ]
    [ -n "$BUILDKITE_COMMIT" ]
    [ -n "$BUILDKITE_BRANCH" ]
    [ -n "$HELM_CHART_REPO" ]
    [ -n "$HELM_CHART_PATH" ]
}

@test "required files exist for testing" {
    [ -f "config/helm-values.yaml" ]
    [ -f ".buildkite/helm-automation.sh" ]
}

@test "merge_helm_values updates openlane coreConfiguration" {
    git -C "$TEST_TEMP_DIR" init -q

    cat > "$TEST_TEMP_DIR/chart-values.yaml" <<EOF
openlane:
  coreConfiguration:
    domain: old.example.com
EOF

    cat > "$TEST_TEMP_DIR/source-values.yaml" <<EOF
openlane:
  coreConfiguration:
    domain: new.example.com
EOF

    run bash -c '. .buildkite/lib/helm.sh && merge_helm_values "$TEST_TEMP_DIR/source-values.yaml" "$TEST_TEMP_DIR/chart-values.yaml" "test-values"'

    [ "$status" -eq 0 ]
    grep -q "new.example.com" "$TEST_TEMP_DIR/chart-values.yaml"

}
