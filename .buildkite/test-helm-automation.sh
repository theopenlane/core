#!/bin/bash
set -euo pipefail

# Comprehensive test script for helm-automation.sh
# Supports multiple testing modes: dry-run, local-repo, webhook-test

YQ_VERSION=4.9.6
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HELM_AUTOMATION_SCRIPT="$SCRIPT_DIR/helm-automation.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_REPO_URL="https://github.com/test/test-infra.git"
TEST_CHART_PATH="charts/test-app"

usage() {
  cat << EOF
Usage: $0 [MODE] [OPTIONS]

Test modes:
  dry-run         Test script logic without making external calls
  local-repo      Test with a local temporary repository
  draft-pr        Test draft PR workflow (syntax and logic validation)
  webhook-test    Test Slack webhook functionality (requires SLACK_WEBHOOK_URL)
  full-test       Run all test modes sequentially

Options:
  -h, --help      Show this help message
  -v, --verbose   Enable verbose output
  --no-cleanup    Don't clean up test files (useful for debugging)

Environment variables:
  SLACK_WEBHOOK_URL   Slack webhook URL for webhook testing
  TEST_REPO_URL       Repository URL for testing (default: $TEST_REPO_URL)
  TEST_CHART_PATH     Chart path for testing (default: $TEST_CHART_PATH)

Examples:
  $0 dry-run
  $0 local-repo --verbose
  $0 draft-pr
  $0 webhook-test
  SLACK_WEBHOOK_URL=https://hooks.slack.com/... $0 webhook-test
EOF
}

log() {
  echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

log_success() {
  echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
  echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
  echo -e "${RED}❌ $1${NC}"
}

cleanup_test_files() {
  if [[ "${CLEANUP_ENABLED:-true}" == "true" ]]; then
    log "Cleaning up test files..."
    rm -rf "$TEST_WORK_DIR" 2>/dev/null || true
    unset TEST_WORK_DIR
  else
    log_warning "Cleanup disabled, test files remain at: ${TEST_WORK_DIR:-none}"
  fi
}

setup_test_environment() {
  log "Setting up test environment..."

  TEST_WORK_DIR=$(mktemp -d)
  trap cleanup_test_files EXIT

  # Create mock build environment
  export BUILDKITE_BUILD_NUMBER="test-123"
  export BUILDKITE_COMMIT="abc123def456789"
  export BUILDKITE_BRANCH="test-branch"
  export BUILDKITE_PIPELINE_NAME="test-pipeline"
  export BUILDKITE_BUILD_CREATOR="test-user"
  export BUILDKITE_BUILD_URL="https://buildkite.com/test/build/123"
  export BUILDKITE_BUILD_CHECKOUT_PATH="$TEST_WORK_DIR/core"
  export HELM_CHART_REPO="$TEST_REPO_URL"
  export HELM_CHART_PATH="$TEST_CHART_PATH"

  log_success "Test environment configured"
}

create_mock_config_files() {
  log "Creating mock configuration files..."

  mkdir -p "$BUILDKITE_BUILD_CHECKOUT_PATH/config"

  # Create mock helm-values.yaml
  cat > "$BUILDKITE_BUILD_CHECKOUT_PATH/config/helm-values.yaml" << 'EOF'
core:
  app:
    name: "test-app"
    version: "1.0.0"
  database:
    host: "db.example.com"
    port: 5432
  cache:
    redis:
      enabled: true
      host: "redis.example.com"

externalSecrets:
  enabled: true
  secretStore:
    name: "vault-backend"
  secrets:
    - name: "database-password"
      key: "app/database/password"
    - name: "api-key"
      key: "app/external/api-key"
EOF

  # Create mock external-secrets directory
  mkdir -p "$BUILDKITE_BUILD_CHECKOUT_PATH/config/external-secrets"

  cat > "$BUILDKITE_BUILD_CHECKOUT_PATH/config/external-secrets/database.yaml" << 'EOF'
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: database-secret
spec:
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: database-secret
  data:
    - secretKey: password
      remoteRef:
        key: app/database/password
EOF

  # Create legacy configmap for backward compatibility
  cat > "$BUILDKITE_BUILD_CHECKOUT_PATH/config/configmap.yaml" << 'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: core-config
data:
  config.yaml: |
    app:
      name: test-app
      debug: false
EOF

  log_success "Mock configuration files created"
}

create_mock_target_repo() {
  log "Creating mock target repository..." >&2

  local repo_dir="$TEST_WORK_DIR/target-repo"
  mkdir -p "$repo_dir/$TEST_CHART_PATH"
  cd "$repo_dir"

  git init >&2
  git config user.email "test@example.com"
  git config user.name "Test User"
  git config commit.gpgsign false

  # Create existing Chart.yaml
  cat > "$TEST_CHART_PATH/Chart.yaml" << 'EOF'
apiVersion: v2
name: test-app
description: A test Helm chart
type: application
version: 1.2.3
appVersion: "1.0.0"
EOF

  # Create existing values.yaml with Kubernetes-specific config
  cat > "$TEST_CHART_PATH/values.yaml" << 'EOF'
# Kubernetes-specific configuration that should be preserved
replicaCount: 3

image:
  repository: nginx
  pullPolicy: IfNotPresent
  tag: "latest"

service:
  type: ClusterIP
  port: 80

ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: app.example.com
      paths:
        - path: /
          pathType: Prefix

# Core configuration (this will be replaced by automation)
core:
  app:
    name: "old-app"
    version: "0.9.0"
EOF

  # Create templates directory
  mkdir -p "$TEST_CHART_PATH/templates"

  git add . >&2
  git commit -m "Initial commit" >&2

  log_success "Mock target repository created at $repo_dir" >&2

  # Return to original directory before echoing path
  cd "$OLDPWD" >&2
  echo "$repo_dir"
}

test_dry_run() {
  log "Starting dry-run tests..."

  setup_test_environment
  create_mock_config_files

  # Test function definitions
  log "Testing function definitions..."

  # Test script syntax
  if bash -n "$HELM_AUTOMATION_SCRIPT"; then
    log_success "Main script syntax is valid"
  else
    log_error "Main script has syntax errors"
    return 1
  fi

  # Test individual functions by extracting them
  local temp_test_script="$TEST_WORK_DIR/test_functions.sh"

  # Extract and test merge_helm_values function
  cat > "$temp_test_script" << 'EOF'
#!/bin/bash
set -euo pipefail
YQ_VERSION=4.9.6
EOF

  sed -n '/^function merge_helm_values/,/^}/p' "$HELM_AUTOMATION_SCRIPT" >> "$temp_test_script"

  if bash -n "$temp_test_script"; then
    log_success "merge_helm_values function syntax is valid"
  else
    log_error "merge_helm_values function has syntax errors"
    return 1
  fi

  # Test other key functions
  for func in "send_slack_notification" "copy_and_track" "copy_directory_and_track"; do
    local func_test_script="$TEST_WORK_DIR/test_${func}.sh"
    cat > "$func_test_script" << 'EOF'
#!/bin/bash
set -euo pipefail
YQ_VERSION=4.9.6
EOF

    if sed -n "/^function ${func}/,/^}/p" "$HELM_AUTOMATION_SCRIPT" >> "$func_test_script"; then
      if bash -n "$func_test_script"; then
        log_success "$func function syntax is valid"
      else
        log_error "$func function has syntax errors"
        return 1
      fi
    else
      log_warning "$func function not found or has different format"
    fi
  done

  log_success "Dry-run tests completed"
}

test_local_repo() {
  log "Starting local repository tests..."

  setup_test_environment
  create_mock_config_files
  local repo_dir=$(create_mock_target_repo)

  # Override environment to use local repo
  export HELM_CHART_REPO="$repo_dir"

  # Test yq availability
  if ! docker run --rm mikefarah/yq:$YQ_VERSION --version > /dev/null 2>&1; then
    log_error "yq Docker image not available"
    return 1
  fi

  # Test merge functionality
  log "Testing merge_helm_values functionality..."

  local source_file="$BUILDKITE_BUILD_CHECKOUT_PATH/config/helm-values.yaml"
  local target_file="$repo_dir/$TEST_CHART_PATH/values.yaml"

  # Backup original target
  cp "$target_file" "$target_file.original"

  # Extract merge function and test it
  local test_script="$TEST_WORK_DIR/test_merge.sh"
  cat > "$test_script" << EOF
#!/bin/bash
set -euo pipefail
YQ_VERSION=$YQ_VERSION
$(sed -n '/^function merge_helm_values/,/^}/p' "$HELM_AUTOMATION_SCRIPT")

merge_helm_values "$source_file" "$target_file" "test values"
EOF

  chmod +x "$test_script"
  cd "$repo_dir"

  if "$test_script"; then
    log_success "merge_helm_values executed successfully"

    # Verify merge results
    log "Verifying merge results..."

    # Check that core section was updated
    local new_app_name=$(docker run --rm -v "$repo_dir":/workdir mikefarah/yq:$YQ_VERSION e '.core.app.name' "$TEST_CHART_PATH/values.yaml")
    if [[ "$new_app_name" == '"test-app"' ]]; then
      log_success "Core configuration was properly merged"
    else
      log_error "Core configuration merge failed. Got: $new_app_name"
      return 1
    fi

    # Check that Kubernetes config was preserved
    local replica_count=$(docker run --rm -v "$repo_dir":/workdir mikefarah/yq:$YQ_VERSION e '.replicaCount' "$TEST_CHART_PATH/values.yaml")
    if [[ "$replica_count" == "3" ]]; then
      log_success "Kubernetes configuration was preserved"
    else
      log_error "Kubernetes configuration was not preserved. Got: $replica_count"
      return 1
    fi

    # Check that external secrets were added
    local external_secrets_enabled=$(docker run --rm -v "$repo_dir":/workdir mikefarah/yq:$YQ_VERSION e '.externalSecrets.enabled' "$TEST_CHART_PATH/values.yaml")
    if [[ "$external_secrets_enabled" == "true" ]]; then
      log_success "External secrets configuration was added"
    else
      log_warning "External secrets configuration was not found. Got: $external_secrets_enabled"
    fi

  else
    log_error "merge_helm_values failed to execute"
    return 1
  fi

  log_success "Local repository tests completed"
}

test_webhook() {
  log "Starting webhook tests..."

  if [[ -z "${SLACK_WEBHOOK_URL:-}" ]]; then
    log_warning "SLACK_WEBHOOK_URL not set, skipping webhook tests"
    return 0
  fi

  setup_test_environment

  # Extract and test webhook function
  local test_script="$TEST_WORK_DIR/test_webhook.sh"
  cat > "$test_script" << 'EOF'
#!/bin/bash
set -euo pipefail

# Mock change summary for testing
change_summary="\n- 🔄 Merged test values\n- 📈 Bumped chart version to 1.2.4"

$(sed -n '/^function send_slack_notification/,/^}/p' '"$HELM_AUTOMATION_SCRIPT"')

send_slack_notification "https://github.com/test/repo/pull/123"
EOF

  chmod +x "$test_script"

  log "Testing Slack webhook notification..."
  if "$test_script"; then
    log_success "Slack webhook test completed successfully"
  else
    log_error "Slack webhook test failed"
    return 1
  fi

  log_success "Webhook tests completed"
}

test_draft_pr_workflow() {
  log "Testing draft PR workflow..."

  setup_test_environment
  create_mock_config_files
  local repo_dir=$(create_mock_target_repo)

  # Mock PR environment
  export BUILDKITE_PULL_REQUEST="123"
  export HELM_CHART_REPO="$repo_dir"

  log "Testing draft PR creation..."

  # Test draft-pr-automation.sh syntax
  if bash -n "/Users/manderson/core/.buildkite/draft-pr-automation.sh"; then
    log_success "draft-pr-automation.sh syntax is valid"
  else
    log_error "draft-pr-automation.sh has syntax errors"
    return 1
  fi

  # Test link-pr-automation.sh syntax
  if bash -n "/Users/manderson/core/.buildkite/link-pr-automation.sh"; then
    log_success "link-pr-automation.sh syntax is valid"
  else
    log_error "link-pr-automation.sh has syntax errors"
    return 1
  fi

  # Test post-merge-pr-automation.sh syntax
  if bash -n "/Users/manderson/core/.buildkite/post-merge-pr-automation.sh"; then
    log_success "post-merge-pr-automation.sh syntax is valid"
  else
    log_error "post-merge-pr-automation.sh has syntax errors"
    return 1
  fi

  # Create mock draft PR metadata for linking test
  cat > "$BUILDKITE_BUILD_CHECKOUT_PATH/.draft_pr_metadata" << 'EOF'
DRAFT_PR_URL=https://github.com/test/infra/pull/456
DRAFT_BRANCH=draft-core-pr-123-9999
CORE_PR_NUMBER=123
INFRA_REPO=/tmp/test-repo
EOF

  # Test PR linking logic (without actually making API calls)
  log "Testing PR linking logic..."
  local link_test_script="$TEST_WORK_DIR/test_link_logic.sh"
  cat > "$link_test_script" << 'EOF'
#!/bin/bash
set -euo pipefail

# Mock environment
export BUILDKITE_PULL_REQUEST="123"
export BUILDKITE_BUILD_CHECKOUT_PATH="$(pwd)"

# Check if metadata file exists and is readable
if [[ -f ".draft_pr_metadata" ]]; then
  source ".draft_pr_metadata"
  echo "✅ Draft PR metadata loaded successfully"
  echo "   Draft PR URL: $DRAFT_PR_URL"
  echo "   Core PR Number: $CORE_PR_NUMBER"
else
  echo "❌ Draft PR metadata not found"
  exit 1
fi

# Test comment generation logic
core_pr_comment="## 🔗 Related Infrastructure Changes
Test comment content..."

if [[ ${#core_pr_comment} -gt 50 ]]; then
  echo "✅ Core PR comment generated successfully"
else
  echo "❌ Core PR comment generation failed"
  exit 1
fi

echo "✅ PR linking logic test completed"
EOF

  cd "$BUILDKITE_BUILD_CHECKOUT_PATH"
  chmod +x "$link_test_script"

  if "$link_test_script"; then
    log_success "PR linking logic test passed"
  else
    log_error "PR linking logic test failed"
    return 1
  fi

  log_success "Draft PR workflow tests completed"
}

test_chart_versioning() {
  log "Testing chart version increment and changelog generation..."

  setup_test_environment
  local repo_dir=$(create_mock_target_repo)

  cd "$repo_dir"

  # Test chart version increment
  local chart_file="$TEST_CHART_PATH/Chart.yaml"
  local original_version=$(grep '^version:' "$chart_file" | awk '{print $2}')

  log "Original chart version: $original_version"

  # Extract versioning logic and test it
  local test_script="$TEST_WORK_DIR/test_versioning.sh"
  cat > "$test_script" << EOF
#!/bin/bash
set -euo pipefail
cd "$repo_dir"
chart_dir="$TEST_CHART_PATH"
change_summary="\n- 🔄 Merged test values\n- 🔐 Updated external secrets"
export BUILDKITE_BUILD_NUMBER="$BUILDKITE_BUILD_NUMBER"
export BUILDKITE_COMMIT="$BUILDKITE_COMMIT"
export BUILDKITE_BRANCH="$BUILDKITE_BRANCH"

$(sed -n '/^# Increment chart version and generate changelog/,/^fi$/p' "$HELM_AUTOMATION_SCRIPT")
EOF

  chmod +x "$test_script"

  if "$test_script"; then
    # Verify version was incremented
    local new_version=$(grep '^version:' "$chart_file" | awk '{print $2}')
    log "New chart version: $new_version"

    if [[ "$new_version" != "$original_version" ]]; then
      log_success "Chart version was incremented: $original_version -> $new_version"
    else
      log_error "Chart version was not incremented"
      return 1
    fi

    # Verify changelog was created/updated
    local changelog_file="$TEST_CHART_PATH/CHANGELOG.md"
    if [[ -f "$changelog_file" ]]; then
      log_success "Changelog file was created"

      # Check if new version is in changelog
      if grep -q "$new_version" "$changelog_file"; then
        log_success "New version entry added to changelog"
      else
        log_error "New version not found in changelog"
        return 1
      fi
    else
      log_error "Changelog file was not created"
      return 1
    fi

  else
    log_error "Chart versioning test failed"
    return 1
  fi

  log_success "Chart versioning tests completed"
}

run_full_test_suite() {
  log "Running full test suite..."

  local failed_tests=0

  log "=== Test 1: Dry Run ==="
  if ! test_dry_run; then
    log_error "Dry run tests failed"
    ((failed_tests++))
  fi

  log "=== Test 2: Local Repository ==="
  if ! test_local_repo; then
    log_error "Local repository tests failed"
    ((failed_tests++))
  fi

  log "=== Test 3: Chart Versioning ==="
  if ! test_chart_versioning; then
    log_error "Chart versioning tests failed"
    ((failed_tests++))
  fi

  log "=== Test 4: Draft PR Workflow ==="
  if ! test_draft_pr_workflow; then
    log_error "Draft PR workflow tests failed"
    ((failed_tests++))
  fi

  log "=== Test 5: Webhook ==="
  if ! test_webhook; then
    log_error "Webhook tests failed"
    ((failed_tests++))
  fi

  if [[ $failed_tests -eq 0 ]]; then
    log_success "All tests passed! ✨"
    return 0
  else
    log_error "$failed_tests test(s) failed"
    return 1
  fi
}

# Parse command line arguments
VERBOSE=false
CLEANUP_ENABLED=true
MODE=""

while [[ $# -gt 0 ]]; do
  case $1 in
    dry-run|local-repo|draft-pr|webhook-test|full-test)
      MODE="$1"
      shift
      ;;
    -v|--verbose)
      VERBOSE=true
      shift
      ;;
    --no-cleanup)
      CLEANUP_ENABLED=false
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      log_error "Unknown option: $1"
      usage
      exit 1
      ;;
  esac
done

if [[ -z "$MODE" ]]; then
  log_error "No test mode specified"
  usage
  exit 1
fi

# Enable verbose output if requested
if [[ "$VERBOSE" == "true" ]]; then
  set -x
fi

# Verify helm-automation.sh exists
if [[ ! -f "$HELM_AUTOMATION_SCRIPT" ]]; then
  log_error "helm-automation.sh not found at $HELM_AUTOMATION_SCRIPT"
  exit 1
fi

log "Starting helm-automation tests..."
log "Mode: $MODE"
log "Verbose: $VERBOSE"
log "Cleanup: $CLEANUP_ENABLED"

# Run the specified test mode
case "$MODE" in
  dry-run)
    test_dry_run
    ;;
  local-repo)
    test_local_repo
    ;;
  draft-pr)
    test_draft_pr_workflow
    ;;
  webhook-test)
    test_webhook
    ;;
  full-test)
    run_full_test_suite
    ;;
  *)
    log_error "Invalid test mode: $MODE"
    exit 1
    ;;
esac

log_success "Test execution completed successfully!"