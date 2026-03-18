#!/bin/bash
# Common utility functions for Buildkite automation scripts

# Global variables
YQ_VERSION=${YQ_VERSION:-4.45.4}
COMMON_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Function to install yq if not available
install_yq() {
    if ! command -v yq >/dev/null 2>&1; then
        echo "Installing yq version ${YQ_VERSION}..."
        local yq_binary="yq_linux_amd64"
        local temp_file="/tmp/yq"

        if wget -q "https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/${yq_binary}" -O "$temp_file"; then
            chmod +x "$temp_file"
            mv "$temp_file" /usr/local/bin/yq
            echo "✅ yq v${YQ_VERSION} installed successfully"
        else
            echo "❌ Failed to install yq" >&2
            return 1
        fi
    else
        echo "✅ yq already available ($(yq --version))"
    fi
}

# Function to install gh (GitHub CLI) if not available
install_gh() {
    if ! command -v gh >/dev/null 2>&1; then
        echo "Installing GitHub CLI..."
        if command -v apk >/dev/null 2>&1; then
            apk add --no-cache github-cli
            echo "✅ GitHub CLI installed successfully"
        else
            echo "❌ Package manager not supported for GitHub CLI installation" >&2
            return 1
        fi
    else
        echo "✅ GitHub CLI already available ($(gh --version | head -1))"
    fi
}

# Function to configure git user settings
setup_git_user() {
    local email="${1:-bender@theopenlane.io}"
    local name="${2:-theopenlane-bender}"

    echo "🔧 Configuring git user settings..."
    git config --local user.email "$email"
    git config --local user.name "$name"

    # Configure git to use GitHub token for authentication
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        git config --local url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
    fi

    echo "✅ Git user configured: $name <$email>"
}

# Function to create a temporary working directory with cleanup
create_temp_workspace() {
    local temp_dir=$(mktemp -d)
    trap "rm -rf \"$temp_dir\"" EXIT
    echo "$temp_dir"
}

# Function to extract core PR number from various sources
extract_core_pr_number() {
    local source="$1"

    # Try different patterns to extract PR number
    if [[ "$source" =~ core\ PR\ \#([0-9]+) ]]; then
        echo "${BASH_REMATCH[1]}"
    elif [[ "$source" =~ core-pr-([0-9]+) ]]; then
        echo "${BASH_REMATCH[1]}"
    elif [[ "$source" =~ \#([0-9]+) ]]; then
        echo "${BASH_REMATCH[1]}"
    fi
}

# Function to check if a core PR exists and get its status
check_core_pr_status() {
    local pr_number="$1"
    local repo="${2:-theopenlane/core}"

    if [[ -z "$pr_number" ]]; then
        echo "ERROR: PR number is required" >&2
        return 1
    fi

    gh pr view "$pr_number" --repo "$repo" --json state,title,updatedAt 2>/dev/null || echo ""
}

# Function to check if a core PR was recently updated (within last 24 hours)
is_recent_pr_activity() {
    local pr_updated_timestamp="$1"

    if [[ -z "$pr_updated_timestamp" ]]; then
        return 1
    fi

    local updated_epoch=$(date -d "$pr_updated_timestamp" +%s 2>/dev/null || echo "0")
    local current_epoch=$(date +%s)
    local time_diff=$((current_epoch - updated_epoch))

    # Return 0 (true) if updated within last 24 hours (86400 seconds)
    [[ $time_diff -lt 86400 ]]
}

# Function to create a git commit with standardized message format
create_commit() {
    local commit_type="$1"
    local summary="$2"
    local changes="$3"
    local build_info="$4"

    local commit_message="$commit_type: $summary

$changes

Build Information:
$build_info"

    git commit -m "$commit_message"
}

# Function to safely push a branch
safe_push_branch() {
    local branch="$1"
    local force="${2:-false}"

    local push_args="origin $branch"
    if [[ "$force" == "true" ]]; then
        push_args="-f $push_args"
    fi

    if git push $push_args; then
        echo "✅ Branch $branch pushed successfully"
        return 0
    else
        echo "❌ Failed to push branch $branch" >&2
        return 1
    fi
}

# Function to normalize GitHub repository references to owner/repo form
normalize_github_repo() {
    local repo="$1"

    repo="${repo#git@github.com:}"
    repo="${repo#ssh://git@github.com/}"
    repo="${repo#https://github.com/}"
    repo="${repo#http://github.com/}"
    repo="${repo%.git}"
    repo="${repo#/}"

    echo "$repo"
}

# Function to safely delete a remote branch (best effort)
safe_delete_branch() {
    local branch="$1"
    local repo="${2:-${HELM_CHART_REPO:-}}"

    if [[ -z "$branch" ]]; then
        echo "WARNING: No branch name provided for deletion"
        return 1
    fi

    local normalized_repo=""
    if [[ -n "$repo" ]]; then
        normalized_repo=$(normalize_github_repo "$repo")
    fi

    local encoded_branch="${branch//%/%25}"
    encoded_branch="${encoded_branch//\//%2F}"

    if [[ -n "$normalized_repo" ]] && command -v gh >/dev/null 2>&1; then
        if gh api --repo "$normalized_repo" -X DELETE "repos/{owner}/{repo}/git/refs/heads/${encoded_branch}" >/dev/null 2>&1; then
            echo "Branch $branch deleted successfully"
        else
            echo "WARNING: Failed to delete branch $branch from $normalized_repo (may not exist)"
        fi
        return 0
    fi

    if GIT_TERMINAL_PROMPT=0 git push origin --delete "$branch" >/dev/null 2>&1; then
        echo "Branch $branch deleted successfully"
    else
        echo "WARNING: Failed to delete branch $branch (may not exist)"
    fi

    return 0
}

# Function to check if running in correct build context
validate_build_context() {
    local required_branch="${1:-main}"
    local require_merge="${2:-false}"

    if [[ "${BUILDKITE_BRANCH:-}" != "$required_branch" ]]; then
        echo "ℹ️  Not running on $required_branch branch (current: ${BUILDKITE_BRANCH:-unknown}), skipping"
        return 1
    fi

    if [[ "$require_merge" == "true" ]] && command -v git >/dev/null 2>&1; then
        local parent_count=$(git rev-list --parents -n 1 HEAD 2>/dev/null | wc -w || echo "1")
        if [[ $parent_count -lt 3 ]]; then
            echo "ℹ️  This appears to be a regular commit, not a merge (parent count: $((parent_count - 1)))"
            return 1
        fi
    fi

    return 0
}

# Function to log script execution context
log_execution_context() {
    local script_name="$1"

    echo "=== $script_name ==="
    echo "Repository: ${HELM_CHART_REPO:-unknown}"
    echo "Branch: ${BUILDKITE_BRANCH:-unknown}"
    echo "Build Number: ${BUILDKITE_BUILD_NUMBER:-unknown}"
    echo "Commit: ${BUILDKITE_COMMIT:-unknown}"
    echo "Script Directory: ${COMMON_LIB_DIR}"
    echo "=========================="
}

# Function to install all common dependencies
install_dependencies() {
    echo "🔧 Installing dependencies..."
    install_yq
    install_gh
    echo "✅ All dependencies installed"
}

# Function to determine if the current PR introduces configuration changes
# Arguments:
#   $1 - base branch to diff against (default: main)
#   $2 - repository directory (default: $BUILDKITE_BUILD_CHECKOUT_PATH)
#   $3 - path containing configuration files (default: config)
pr_has_config_changes() {
    local base_branch="${1:-main}"
    local repo_dir="${2:-${BUILDKITE_BUILD_CHECKOUT_PATH}}"
    local config_path="${3:-config}"

    git -C "$repo_dir" fetch origin "$base_branch" >/dev/null 2>&1 || return 1

    if git -C "$repo_dir" diff --quiet "origin/${base_branch}...HEAD" -- "$config_path"; then
        return 1
    fi

    return 0
}
