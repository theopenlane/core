#!/bin/bash
# Template loading and substitution utility functions

# Function to load template and substitute variables
load_template() {
    local template_file="$1"
    shift

    if [[ ! -f "$template_file" ]]; then
        echo "⚠️  Template file not found: $template_file" >&2
        return 1
    fi

    local content
    content=$(cat "$template_file")

    # Perform substitutions (portable for Bash 3.x)
    for arg in "$@"; do
        if [[ "$arg" == *"="* ]]; then
            key="${arg%%=*}"
            value="${arg#*=}"
            # Escape special characters in value and use printf for safe substitution
            escaped_value=$(printf '%s\n' "$value" | sed 's/[[\.*^$()+?{|]/\\&/g')
            content=$(echo "$content" | sed "s/{{${key}}}/$escaped_value/g")
        fi
    done

    echo "$content"
}

# Function to get template directory path
get_template_dir() {
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    echo "${script_dir}/../templates"
}
