#!/bin/bash
# GitHub-specific utility functions for Buildkite automation scripts

# Source common utilities
GITHUB_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${GITHUB_LIB_DIR}/common.sh"
source "${GITHUB_LIB_DIR}/templates.sh"

# Function to create a PR comment
create_pr_comment() {
    local pr_number="$1"
    local repo="$2"
    local comment_body="$3"

    if gh pr comment "$pr_number" --repo "$repo" --body "$comment_body"; then
        echo "✅ Comment added to PR #$pr_number"
        return 0
    else
        echo "⚠️  Failed to add comment to PR #$pr_number"
        return 1
    fi
}

# Function to create a draft PR
create_draft_pr() {
    local repo="$1"
    local head_branch="$2"
    local title="$3"
    local body="$4"

    if gh pr create \
        --repo "$repo" \
        --head "$head_branch" \
        --title "$title" \
        --body "$body" \
        --draft; then
        echo "✅ Draft PR created successfully"
        return 0
    else
        echo "❌ Failed to create draft PR"
        return 1
    fi
}

# Function to create a regular PR
create_pr() {
    local repo="$1"
    local head_branch="$2"
    local title="$3"
    local body="$4"

    if gh pr create \
        --repo "$repo" \
        --head "$head_branch" \
        --title "$title" \
        --body "$body"; then
        echo "✅ PR created successfully"
        return 0
    else
        echo "❌ Failed to create PR"
        return 1
    fi
}

# Function to convert draft PR to ready
convert_draft_to_ready() {
    local pr_number="$1"
    local repo="$2"

    if gh pr ready "$pr_number" --repo "$repo"; then
        echo "✅ PR #$pr_number converted from draft to ready"
        return 0
    else
        echo "⚠️  Failed to convert PR #$pr_number from draft to ready"
        return 1
    fi
}

# Function to close a PR
close_pr() {
    local pr_number="$1"
    local repo="$2"
    local comment="${3:-}"

    # Add comment before closing if provided
    if [[ -n "$comment" ]]; then
        create_pr_comment "$pr_number" "$repo" "$comment"
    fi

    if gh pr close "$pr_number" --repo "$repo"; then
        echo "✅ PR #$pr_number closed successfully"
        return 0
    else
        echo "⚠️  Failed to close PR #$pr_number"
        return 1
    fi
}

# Function to update PR title and body
update_pr() {
    local pr_number="$1"
    local repo="$2"
    local title="$3"
    local body="$4"

    if gh pr edit "$pr_number" --repo "$repo" --title "$title" --body "$body"; then
        echo "✅ PR #$pr_number updated successfully"
        return 0
    else
        echo "⚠️  Failed to update PR #$pr_number"
        return 1
    fi
}

# Function to get PR URL
get_pr_url() {
    local pr_number="$1"
    local repo="$2"

    gh pr view "$pr_number" --repo "$repo" --json url --jq '.url' 2>/dev/null || echo ""
}

# Function to find draft PRs matching a pattern
find_draft_prs() {
    local repo="$1"
    local pattern="$2"
    local output_format="${3:-number:headRefName:title}"

    gh pr list --repo "$repo" --state open --draft --json number,title,headRefName \
        | jq -r ".[] | select((.title | test(\"$pattern\"; \"i\")) or (.headRefName | test(\"$pattern\"; \"i\"))) | \"\\(.number):\\(.headRefName):\\(.title)\""
}

# Function to check if PR exists and get its info
get_pr_info() {
    local pr_number="$1"
    local repo="$2"
    local fields="${3:-state,title,updatedAt}"

    gh pr view "$pr_number" --repo "$repo" --json "$fields" 2>/dev/null || echo ""
}

# Function to find existing draft PR for a core PR
find_existing_draft_pr() {
    local repo="$1"
    local core_pr_number="$2"
    local draft_branch="${3:-draft-core-pr-${core_pr_number}}"

    # Prefer matching by branch so we can reuse PRs even if the title changes.
    local existing_pr
    existing_pr=$(gh pr list --repo "$repo" --state open --json number,headRefName,title --jq ".[] | select(.headRefName == \"${draft_branch}\") | .number" | head -1)

    if [[ -n "$existing_pr" ]]; then
        echo "$existing_pr"
        return 0
    fi

    existing_pr=$(gh pr list --repo "$repo" --state open --json title,number --jq ".[] | select(.title | test(\"core PR #${core_pr_number}\"; \"i\")) | .number" | head -1)
    echo "$existing_pr"
}

# Function to get the branch name from a PR
get_pr_branch() {
    local pr_number="$1"
    local repo="$2"

    gh pr view "$pr_number" --repo "$repo" --json headRefName --jq '.headRefName' 2>/dev/null || echo ""
}

# Function to generate PR body for core config changes
generate_core_config_pr_body() {
    local core_pr_number="$1"
    local changes="$2"
    local build_number="${3:-${BUILDKITE_BUILD_NUMBER:-unknown}}"
    local commit="${4:-${BUILDKITE_COMMIT:-unknown}}"
    local branch="${5:-${BUILDKITE_BRANCH:-unknown}}"

    # Format changes to convert literal \n to actual newlines
    local formatted_changes=$(printf "%b" "$changes")
    local template_dir=$(get_template_dir)

    load_template "${template_dir}/github/helm-update-pr.md" \
        "CHANGE_SUMMARY=${formatted_changes}" \
        "BUILD_NUMBER=${build_number}" \
        "SOURCE_COMMIT_SHORT=${commit:0:8}" \
        "SOURCE_COMMIT_FULL=${commit}" \
        "SOURCE_BRANCH=${branch}"
}

# Function to generate draft PR body
generate_draft_pr_body() {
    local core_pr_number="$1"
    local changes="$2"
    local commit="${3:-${BUILDKITE_COMMIT:-unknown}}"
    local branch="${4:-${BUILDKITE_BRANCH:-unknown}}"

    # Format changes to convert literal \n to actual newlines
    local formatted_changes=$(printf "%b" "$changes")
    local template_dir=$(get_template_dir)

    load_template "${template_dir}/github/draft-pr.md" \
        "CORE_PR_NUMBER=${core_pr_number}" \
        "CHANGE_SUMMARY=${formatted_changes}" \
        "BUILD_NUMBER=${BUILDKITE_BUILD_NUMBER:-unknown}" \
        "SOURCE_COMMIT_SHORT=${commit:0:8}" \
        "SOURCE_COMMIT_FULL=${commit}" \
        "SOURCE_BRANCH=${branch}"
}

# Function to generate PR closure comment
generate_closure_comment() {
    local core_pr_number="$1"
    local reason="${2:-merged}"
    local template_dir=$(get_template_dir)

    case "$reason" in
        "merged")
            load_template "${template_dir}/github/pr-close-comment.md" \
                "CORE_PR_NUMBER=${core_pr_number}"
            ;;
        "closed")
            load_template "${template_dir}/github/pr-close-without-merge-comment.md" \
                "CORE_PR_NUMBER=${core_pr_number}"
            ;;
        *)
            echo "Unknown reason: $reason" >&2
            return 1
            ;;
    esac
}

# Function to generate PR ready comment
generate_ready_comment() {
    local core_pr_number="$1"
    local changes="$2"

    local template_dir=$(get_template_dir)

    load_template "${template_dir}/github/pr-ready-comment.md" \
        "CORE_PR_NUMBER=${core_pr_number}"
}
