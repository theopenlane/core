#!/bin/bash
# Helm-specific utility functions for Buildkite automation scripts

# Source common utilities
HELM_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${HELM_LIB_DIR}/common.sh"

# Function to merge Helm values with existing chart values
merge_helm_values() {
    local source="$1"
    local target="$2"
    local description="$3"

    if [[ ! -f "$source" ]]; then
        echo "⚠️  Source values file not found: $source"
        return 1
    fi

    # Create backup of existing values
    if [[ -f "$target" ]]; then
        cp "$target" "${target}.backup"
    fi

    # Create temporary merged file
    local temp_merged="${target}.merged"

    if [[ -f "$target" ]]; then
        # Merge strategy: Use yq to merge the generated core values with existing values
        # The 'openlane.coreConfiguration' section from our generated file will be merged/replaced
        # All other sections in the target will be preserved

        cp "$target" "$temp_merged"

        # Extract core section from source and merge into target's openlane.coreConfiguration
        # First try .openlane.coreConfiguration, then fall back to .coreConfiguration
        core_section=$(yq e '.openlane.coreConfiguration' "$source" 2>/dev/null || echo "")
        if [[ -z "$core_section" ]] || [[ "$core_section" == "null" ]]; then
            core_section=$(yq e '.coreConfiguration' "$source" 2>/dev/null || echo "")
        fi

        # Always merge into openlane.coreConfiguration regardless of source structure
        if [[ -n "$core_section" ]] && [[ "$core_section" != "null" ]]; then
            # Use yq to write the section, which will preserve the exact formatting from source
            echo "$core_section" > /tmp/core-section.yaml
            yq e -i '.openlane.coreConfiguration = load("/tmp/core-section.yaml")' "$temp_merged"
        fi

        # Also merge any externalSecrets configuration if it exists in generated file
        # Always merge into openlane.coreConfiguration.externalSecrets and remove any root-level externalSecrets
        external_secrets_section=$(yq e '.externalSecrets' "$source" 2>/dev/null || echo "")
        if [[ -n "$external_secrets_section" ]] && [[ "$external_secrets_section" != "null" ]]; then
            echo "$external_secrets_section" > /tmp/external-secrets-section.yaml
            yq e -i '.openlane.coreConfiguration.externalSecrets = load("/tmp/external-secrets-section.yaml")' "$temp_merged"
            # Remove any root-level externalSecrets to avoid duplication
            yq e -i 'del(.externalSecrets)' "$temp_merged"
        fi

    else
        # If target doesn't exist, create a new structure with source content
        # Always structure as openlane.coreConfiguration regardless of source structure

        # Extract coreConfiguration from source
        core_section=$(yq e '.openlane.coreConfiguration' "$source" 2>/dev/null || echo "")
        if [[ -z "$core_section" ]] || [[ "$core_section" == "null" ]]; then
            core_section=$(yq e '.coreConfiguration' "$source" 2>/dev/null || echo "")
        fi

        # Start with minimal structure
        echo 'openlane:' > "$temp_merged"

        # Add coreConfiguration if it exists
        if [[ -n "$core_section" ]] && [[ "$core_section" != "null" ]]; then
            echo "$core_section" > /tmp/core-section.yaml
            yq e -i '.openlane.coreConfiguration = load("/tmp/core-section.yaml")' "$temp_merged"
        else
            yq e -i '.openlane.coreConfiguration = {}' "$temp_merged"
        fi

        # Add externalSecrets if it exists
        external_secrets_section=$(yq e '.externalSecrets' "$source" 2>/dev/null || echo "")
        if [[ -n "$external_secrets_section" ]] && [[ "$external_secrets_section" != "null" ]]; then
            echo "$external_secrets_section" > /tmp/external-secrets-section.yaml
            yq e -i '.openlane.coreConfiguration.externalSecrets = load("/tmp/external-secrets-section.yaml")' "$temp_merged"
        fi
    fi

    # Check if there are actual differences
    if [[ -f "$target" ]] && diff -q "$target" "$temp_merged" > /dev/null 2>&1; then
        echo "  ℹ️  No changes detected in $description" >&2  # Send status to stderr, not stdout
        rm -f "$temp_merged" "${target}.backup" /tmp/core-section.yaml /tmp/external-secrets-section.yaml
        return 1
    fi

    # Calculate detailed changes for changelog
    local changes_detail=""
    if [[ -f "$target" ]]; then

        # Compare core section changes
        if yq e '.openlane.coreConfiguration' "$target" > /tmp/old-core.yaml 2>/dev/null; then
            local core_changes=$(yq e 'diff("/tmp/old-core.yaml")' /tmp/core-section.yaml 2>/dev/null | grep -E '^\+\+\+|^---' | wc -l | tr -d ' \n' || echo "0")
            if [[ "$core_changes" -gt 0 ]]; then
                changes_detail+="\n- ✅ Core configuration updated ($core_changes changes)"
            fi
        fi

        # Check for new/modified external secrets
        if [[ -f /tmp/external-secrets-section.yaml ]]; then
            changes_detail+="\n- 🔐 External secrets configuration updated"
        fi
    else
        changes_detail+="\n- ✅ Initial values file created"
    fi

    # Apply the merged changes
    mv "$temp_merged" "$target"
    git add "$target"

    # Cleanup
    rm -f "${target}.backup" /tmp/core-section.yaml /tmp/external-secrets-section.yaml /tmp/old-core.yaml

    echo "$changes_detail"
    return 0
}

# Function to copy files and detect changes (for non-values files)
copy_and_track() {
    local source="$1"
    local target="$2"
    local description="$3"

    if [[ -f "$source" ]]; then
        # Check if target exists and has differences
        if [[ -f "$target" ]]; then
            if ! diff -q "$source" "$target" > /dev/null 2>&1; then
                cp "$source" "$target"
                git add "$target"
                echo "✅ Updated $description" >&2  # Send status to stderr, not stdout
                return 0
            fi
        else
            mkdir -p "$(dirname "$target")"
            cp "$source" "$target"
            git add "$target"
            echo "✨ Created $description" >&2  # Send status to stderr, not stdout
            return 0
        fi
    fi
    return 1
}

# Function to copy directories and detect changes
copy_directory_and_track() {
    local source="$1"
    local target="$2"
    local description="$3"

    if [[ -d "$source" ]]; then
        # Check if target exists and has differences
        if [[ -d "$target" ]]; then
            if ! diff -r "$source" "$target" > /dev/null 2>&1; then
                rm -rf "$target"
                mkdir -p "$(dirname "$target")"
                cp -r "$source" "$target"
                git add "$target"
                echo "🔐 Updated $description" >&2  # Send status to stderr, not stdout
                return 0
            fi
        else
            mkdir -p "$(dirname "$target")"
            cp -r "$source" "$target"
            git add "$target"
            echo "🆕 Created $description" >&2  # Send status to stderr, not stdout
            return 0
        fi
    fi
    return 1
}

# Function to increment chart version
increment_chart_version() {
    local chart_file="$1"
    local version_type="${2:-patch}"  # major, minor, patch

    if [[ ! -f "$chart_file" ]]; then
        echo "⚠️  Chart file not found: $chart_file"
        return 1
    fi

    local current=$(grep '^version:' "$chart_file" | awk '{print $2}')
    if [[ -z "$current" ]]; then
        echo "⚠️  Could not find version in chart file"
        return 1
    fi

    IFS='.' read -r major minor patch <<< "$current"

    case "$version_type" in
        major)
            new_version="$((major+1)).0.0"
            ;;
        minor)
            new_version="$major.$((minor+1)).0"
            ;;
        patch|*)
            new_version="$major.$minor.$((patch+1))"
            ;;
    esac

    sed -i -E "s/^version:.*/version: $new_version/" "$chart_file"
    git add "$chart_file"

    echo "$new_version"
    return 0
}

# Function to generate changelog entry
generate_changelog_entry() {
    local version="$1"
    local changes="$2"
    local build_number="${3:-${BUILDKITE_BUILD_NUMBER:-unknown}}"
    local commit="${4:-${BUILDKITE_COMMIT:-unknown}}"
    local branch="${5:-${BUILDKITE_BRANCH:-unknown}}"

    cat << EOF
## [$version] - $(date +%Y-%m-%d)

### Changed$changes

### Build Information
- Build Number: $build_number
- Source Commit: ${commit:0:8}
- Source Branch: $branch
- Generated: $(date +'%Y-%m-%d %H:%M:%S UTC')

---
EOF
}

# Function to update changelog
update_changelog() {
    local changelog_file="$1"
    local changelog_entry="$2"

    if [[ -f "$changelog_file" ]]; then
        # Insert new entry at the top after the header
        local temp_changelog=$(mktemp)
        head -n 2 "$changelog_file" > "$temp_changelog"
        echo "" >> "$temp_changelog"
        echo "$changelog_entry" >> "$temp_changelog"
        tail -n +3 "$changelog_file" >> "$temp_changelog"
        mv "$temp_changelog" "$changelog_file"
    else
        # Create new changelog
        cat > "$changelog_file" << EOF
# Changelog

All notable changes to this Helm chart will be documented in this file.

$changelog_entry
EOF
    fi

    git add "$changelog_file"
    echo "📝 Updated changelog" >&2  # Send status to stderr, not stdout
}

# Function to apply config changes to a helm chart
apply_helm_config_changes() {
    local source_dir="$1"
    local chart_dir="$2"
    local changes_made=false
    local change_summary=""

    # Update Helm values.yaml (intelligent merging approach)
    local values_changes=""
    if values_changes=$(merge_helm_values \
        "$source_dir/helm-values.yaml" \
        "$chart_dir/values.yaml" \
        "Helm values.yaml" 2>&1); then
        changes_made=true
        change_summary+="\n- 🔄 Merged Helm values.yaml$values_changes"
    fi

    # Update external secrets directory
    if copy_directory_and_track \
        "$source_dir/external-secrets" \
        "$chart_dir/templates/external-secrets" \
        "External Secrets templates" 2>&1; then
        changes_made=true
        change_summary+="\n- 🔐 Updated External Secrets templates"
    fi

    # ConfigMap that embeds config.yaml for runtime mounting
    if [[ -f "$source_dir/configmap-config-file.yaml" ]]; then
        if copy_and_track \
            "$source_dir/configmap-config-file.yaml" \
            "$chart_dir/templates/core-configmap-file.yaml" \
            "ConfigMap template" 2>&1; then
            changes_made=true
            change_summary+="\n- ✅ Updated ConfigMap template"
        fi
    fi

    # Return results
    if [[ "$changes_made" == "true" ]]; then
        echo "$change_summary"
        return 0
    else
        echo ""
        return 0
    fi
}
