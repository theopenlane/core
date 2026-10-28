#!/bin/bash
# Disk and cache cleanup shared by the pre-command and post-command hooks

# cache locations come only from the pipeline env, which the pipeline upload job does not get
GO_BUILD_CACHE="${GOCACHE:-}"
GO_MOD_CACHE="${GOMODCACHE:-}"
GOLANGCI_LINT_CACHE_DIR="${GOLANGCI_LINT_CACHE:-}"

# disk usage percent a job must start under
CACHE_CLEANUP_THRESHOLD="${CACHE_CLEANUP_THRESHOLD:-65}"
# disk usage percent at which a job gives up and exits with DISK_RETRY_EXIT_CODE for the pipeline to retry
DISK_RETRY_THRESHOLD="${DISK_RETRY_THRESHOLD:-90}"
DISK_RETRY_EXIT_CODE=75
# cache files modified within this many minutes may be in use by another job on the runner
CACHE_IN_USE_MINUTES=30

# caches_configured reports whether the pipeline env set every cache path, skipping cache work for jobs like the pipeline upload
caches_configured() {
    if [[ -n "$GO_BUILD_CACHE" && -n "$GO_MOD_CACHE" && -n "$GOLANGCI_LINT_CACHE_DIR" ]]; then
        return 0
    fi

    echo "⏭️  Cache paths are not set for this job, skipping cache work"
    return 1
}

# disk_usage_percent prints the used percentage of the current filesystem, failing when it cannot be read
disk_usage_percent() {
    local usage
    usage=$(df -h . 2>/dev/null | tail -1 | awk '{print $5}' | sed 's/%//') || return 1
    [[ "$usage" =~ ^[0-9]+$ ]] || return 1
    echo "$usage"
}

# ensure_cache_dirs creates the cache directories the build tools write to
ensure_cache_dirs() {
    caches_configured || return 0
    mkdir -p "$GO_BUILD_CACHE" "$GO_MOD_CACHE" "$GOLANGCI_LINT_CACHE_DIR" 2>/dev/null || true
}

# is_hosted_runner reports whether the agent lacks the self-hosted tag, where the hosted cache volume manages disk
is_hosted_runner() {
    [[ "${BUILDKITE_AGENT_META_DATA_SELF_HOSTED:-}" != "true" ]]
}

# disk_below_threshold reports whether disk usage is under the cleanup threshold, treating unreadable usage as under so nothing is deleted blindly
disk_below_threshold() {
    local usage
    usage=$(disk_usage_percent) || return 0
    [[ $usage -lt $CACHE_CLEANUP_THRESHOLD ]]
}

# prune_cache_until_below deletes a cache's files oldest first until disk usage is under the threshold, returning 1 if it never gets there
prune_cache_until_below() {
    local label="$1" dir="$2" pattern="${3:-*}" minutes age

    for minutes in 20160 10080 4320 1440 360 "$CACHE_IN_USE_MINUTES"; do
        disk_below_threshold && return 0
        [[ -d "$dir" ]] || return 1

        age="$((minutes / 60))h"
        (( minutes % 60 )) && age="${minutes}m"

        echo "$label cache: removing files unused for over ${age}..."
        find "$dir" -type f -name "$pattern" -mmin +"$minutes" -delete 2>/dev/null || true
    done

    disk_below_threshold
}

# free_disk_to_threshold prunes the go build, golangci-lint, and go module download caches in that order until disk usage is under the threshold
free_disk_to_threshold() {
    local locked=false

    caches_configured || return 0

    # agents on the same runner share the caches, so only one prunes at a time
    if command -v flock >/dev/null 2>&1 && exec 9>"$(dirname "$GO_BUILD_CACHE")/openlane-cache-cleanup.lock"; then
        echo "🔒 Waiting for the runner cache cleanup lock..."
        flock -w 600 9 && locked=true
    fi

    # go only reads module zips when extracting, the extracted module trees stay since running builds read them
    prune_cache_until_below "🔍 Go build" "$GO_BUILD_CACHE" ||
        prune_cache_until_below "🔍 Golangci-lint" "$GOLANGCI_LINT_CACHE_DIR" ||
        prune_cache_until_below "📦 Go module download" "$GO_MOD_CACHE/cache/download" "*.zip" ||
        prune_docker_unused

    if [[ "$locked" == true ]]; then
        flock -u 9
    fi
    exec 9>&- 2>/dev/null || true
}

# prune_build_caches trims the go build, golangci-lint, and go module download caches by age
prune_build_caches() {
    caches_configured || return 0

    prune_cache_dir "🔍 Go build" "$GO_BUILD_CACHE" 5
    prune_cache_dir "🔍 Golangci-lint" "$GOLANGCI_LINT_CACHE_DIR" 5
    # go only reads module zips when extracting, the extracted module trees stay since running builds read them
    prune_cache_dir "📦 Go module download" "$GO_MOD_CACHE/cache/download" 5 "*.zip"
}

# prune_cache_dir reports a cache directory's size and deletes files matching the optional pattern older than the given number of days
prune_cache_dir() {
    local label="$1" dir="$2" days="$3" pattern="${4:-*}"

    [[ -d "$dir" ]] || return 0

    echo "$label cache size: $(du -sh "$dir" 2>/dev/null | cut -f1 || echo unknown)"
    find "$dir" -type f -name "$pattern" -mtime +"$days" -delete 2>/dev/null || true
}

# prune_docker_unused removes stopped containers older than an hour and dangling images, leaving a concurrent job's
# just-exited containers alone
prune_docker_unused() {
    command -v docker >/dev/null 2>&1 || return 0

    echo "🗑️  Removing stopped/dead containers older than 1 hour..."
    docker container prune -f --filter "until=1h" >/dev/null 2>&1 || true

    echo "🖼️  Removing dangling images..."
    docker image prune -f >/dev/null 2>&1 || true
}

# prune_docker_stale removes unused images, anonymous volumes, and build cache older than a week
prune_docker_stale() {
    command -v docker >/dev/null 2>&1 || return 0

    echo "🖼️  Removing unused images older than 7 days..."
    docker image prune -a -f --filter "until=168h" >/dev/null 2>&1 || true

    # Skip network cleanup to avoid interfering with concurrent builds
    echo "🌐 Skipping network cleanup"

    # volume prune does not support an age filter
    echo "💾 Removing unused anonymous volumes..."
    docker volume prune -f >/dev/null 2>&1 || true

    echo "🏗️  Removing build cache older than 7 days..."
    docker builder prune -f --filter "until=168h" >/dev/null 2>&1 || true
}

# report_disk_after prints disk usage after a cleanup and how much it freed, warning when it is still high
report_disk_after() {
    local before="$1" after

    if ! after=$(disk_usage_percent); then
        echo "⚠️  Could not determine disk usage after cleanup"
        return 0
    fi

    echo "📊 Disk usage after cleanup: ${after}%"

    if [[ $before -gt 0 && $((before - after)) -gt 0 ]]; then
        echo "✨ Freed $((before - after))% disk space"
    fi

    if [[ $after -gt 80 ]]; then
        echo "⚠️  WARNING: Disk usage is critically high (${after}%), build may fail due to insufficient disk space"
        echo "💡 Recommend running docker-cleanup.sh during maintenance"
    elif [[ $after -gt 75 ]]; then
        echo "⚠️  WARNING: Disk usage is high (${after}%)"
        echo "💡 Consider running docker-cleanup.sh during off-peak hours"
    fi
}
