#!/bin/bash
# Slack notification utilities using external templates

# Function to send slack notification using template file
function send_slack_notification_from_template() {
  local template_file="$1"
  shift

  # Check if slack webhook is configured
  if [[ -z "${SLACK_WEBHOOK_URL:-}" ]]; then
    echo "ℹ️  Slack not configured (SLACK_WEBHOOK_URL missing), skipping notification"
    return 0
  fi

  # Check if template file exists
  if [[ ! -f "$template_file" ]]; then
    echo "⚠️  Slack template file not found: $template_file"
    return 1
  fi

  echo "📨 Sending slack notification from template: $(basename "$template_file")"

  # Read template
  local message_content
  message_content=$(cat "$template_file")

  # Perform substitutions (portable for Bash 3.x)
  for arg in "$@"; do
    if [[ "$arg" == *"="* ]]; then
      key="${arg%%=*}"
      value="${arg#*=}"
      # Escape special characters for JSON
      # First escape backslashes, then quotes, then convert newlines to \n for JSON
      value=$(echo "$value" | sed 's/\\/\\\\/g' | sed 's/"/\\"/g' | sed ':a;N;$!ba;s/\n/\\n/g')
      # Use a delimiter that won't appear in the content
      message_content=$(echo "$message_content" | sed "s|{{${key}}}|$value|g")
    fi
  done

  # Send to slack using webhook
  local json_response
  json_response=$(curl -sL -X POST \
    -H "Content-Type: application/json" \
    -d "$message_content" \
    "${SLACK_WEBHOOK_URL}")

  # Check if the webhook call was successful
  if [[ $? -eq 0 ]]; then
    echo "✅ Slack notification sent successfully"
    return 0
  else
    echo "⚠️  Failed to send slack notification"
    echo "Response: $json_response"
    return 1
  fi
}

format_summary() {
  local summary="$1"

  # Replace literal \n with actual newlines for slack formatting
  # Use printf to properly interpret the escape sequences
  printf "%b" "$summary" | sed 's/\\n/\n/g'
}

# Function to send helm update notification
function send_helm_update_notification() {
  local pr_url="$1"
  local change_summary="$2"

  local template_file="${BASH_SOURCE[0]%/*}/templates/slack/helm-update-notification.json"

  # Format change summary for Slack (convert <br/> or \n to actual newlines)
  local formatted_summary=$(format_summary "$change_summary")

  send_slack_notification_from_template "$template_file" \
    "PR_URL=$pr_url" \
    "CHANGE_SUMMARY=$formatted_summary" \
    "BUILD_NUMBER=${BUILDKITE_BUILD_NUMBER:-unknown}" \
    "SOURCE_BRANCH=${BUILDKITE_BRANCH:-unknown}" \
    "PIPELINE_NAME=${BUILDKITE_PIPELINE_NAME:-unknown}" \
    "BUILD_CREATOR=${BUILDKITE_BUILD_CREATOR:-unknown}" \
    "BUILD_URL=${BUILDKITE_BUILD_URL:-unknown}"
}

# Function to send PR ready notification (for post-merge)
function send_pr_ready_notification() {
  local infra_pr_url="$1"
  local core_pr_url="$2"
  local core_pr_number="$3"
  local change_summary="$4"

  local template_file="${BASH_SOURCE[0]%/*}/templates/slack/pr-ready-notification.json"

  local formatted_summary=$(format_summary "$change_summary")

  send_slack_notification_from_template "$template_file" \
    "INFRA_PR_URL=$infra_pr_url" \
    "CORE_PR_URL=$core_pr_url" \
    "CORE_PR_NUMBER=$core_pr_number" \
    "CHANGE_SUMMARY=$formatted_summary" \
    "BUILD_NUMBER=${BUILDKITE_BUILD_NUMBER:-unknown}" \
    "BUILD_URL=${BUILDKITE_BUILD_URL:-unknown}"
}

# Function to send release notification (for tagged releases)
function send_release_notification() {
  local pr_url="$1"
  local release_tag="$2"
  local change_summary="$3"

  local template_file="${BASH_SOURCE[0]%/*}/templates/slack/release-notification.json"

  local formatted_summary=$(format_summary "$change_summary")

  send_slack_notification_from_template "$template_file" \
    "PR_URL=$pr_url" \
    "RELEASE_TAG=$release_tag" \
    "CHANGE_SUMMARY=$formatted_summary" \
    "BUILD_NUMBER=${BUILDKITE_BUILD_NUMBER:-unknown}" \
    "BUILD_URL=${BUILDKITE_BUILD_URL:-unknown}" \
    "RELEASE_URL=https://github.com/theopenlane/core/releases/tag/$release_tag"
}
