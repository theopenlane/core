#!/bin/bash

#
# Shared utility functions for Buildkite automation scripts
# Source this script at the beginning of automation scripts that run in containers
#

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