#!/bin/bash

#
# Shared utility functions for Buildkite automation scripts
# Source this script at the beginning of automation scripts that run in containers
#

# Check required tools are available in container
check_required_tools() {
    local missing_tools=()

    # Default tools needed by most automation scripts
    local default_tools=("git" "gh" "docker" "jq" "buildkite-agent")

    # Allow override of required tools via parameter
    local required_tools=("${@:-${default_tools[@]}}")

    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done

    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        echo "❌ Missing required tools: ${missing_tools[*]}"
        echo "This script must run in a container with these tools installed"
        echo "Expected to run via Buildkite pipeline with ghcr.io/theopenlane/build-image:latest"
        exit 1
    fi
}

# Verify environment variables are set
check_environment() {
    local missing_vars=()

    for var in "$@"; do
        if [[ -z "${!var:-}" ]]; then
            missing_vars+=("$var")
        fi
    done

    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        echo "❌ Missing required environment variables: ${missing_vars[*]}"
        exit 1
    fi
}

# Initialize automation script with standard checks
init_automation_script() {
    local script_name="$1"
    shift
    local required_tools=("$@")

    echo "=== $script_name ==="

    # Run standard checks
    if [[ ${#required_tools[@]} -gt 0 ]]; then
        check_required_tools "${required_tools[@]}"
    else
        check_required_tools  # Use defaults
    fi

    echo "Tools verified: ✅"
}

# Common environment variables that automation scripts need
setup_common_environment() {
    export YQ_VERSION=${YQ_VERSION:-4.9.6}
    export HELM_CHART_REPO=${HELM_CHART_REPO}
    export HELM_CHART_PATH=${HELM_CHART_PATH:-charts/openlane}
}