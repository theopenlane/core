#!/bin/bash
# Disk and cache cleanup shared by the pre-command and post-command hooks

# cache locations come from the pipeline env, HOME is /tmp so the fallbacks only apply outside the pipeline
GO_BUILD_CACHE="${GOCACHE:-$HOME/.cache/go-build}"
GO_MOD_CACHE="${GOMODCACHE:-$HOME/.cache/go/pkg/mod}"
GOLANGCI_LINT_CACHE_DIR="${GOLANGCI_LINT_CACHE:-$HOME/.cache/golangci-lint}"

# disk_usage_percent prints the used percentage of the current filesystem, failing when it cannot be read
disk_usage_percent() {
    local usage
    usage=$(df -h . 2>/dev/null | tail -1 | awk '{print $5}' | sed 's/%//') || return 1
    [[ "$usage" =~ ^[0-9]+$ ]] || return 1
    echo "$usage"
}

# ensure_cache_dirs creates the cache directories the build tools write to
ensure_cache_dirs() {
    mkdir -p "$GO_BUILD_CACHE" "$GO_MOD_CACHE" "$GOLANGCI_LINT_CACHE_DIR" 2>/dev/null || true
}

# clear_build_caches removes the go build, go module, and golangci-lint caches entirely
clear_build_caches() {
    if command -v go >/dev/null 2>&1; then
        echo "🧹 Clearing Go build and module caches..."
        GOCACHE="$GO_BUILD_CACHE" GOMODCACHE="$GO_MOD_CACHE" go clean -cache -modcache >/dev/null 2>&1 || true
    fi

    echo "🧹 Clearing golangci-lint cache..."
    if command -v golangci-lint >/dev/null 2>&1; then
        GOLANGCI_LINT_CACHE="$GOLANGCI_LINT_CACHE_DIR" golangci-lint cache clean >/dev/null 2>&1 || true
    else
        rm -rf "$GOLANGCI_LINT_CACHE_DIR" 2>/dev/null || true
    fi

    ensure_cache_dirs
}

# prune_build_caches trims the build and lint caches by age and clears the module cache once it passes its size limit
prune_build_caches() {
    # modules are read-only and only safe to remove whole
    if [[ -d "$GO_MOD_CACHE" ]]; then
        local cache_kb
        cache_kb=$(du -sk "$GO_MOD_CACHE" 2>/dev/null | cut -f1 || echo 0)
        echo "📦 Go module cache size: $((cache_kb / 1024 / 1024))G"

        if [[ $cache_kb -gt $(( ${GOMODCACHE_MAX_GB:-20} * 1024 * 1024 )) ]] && command -v go >/dev/null 2>&1; then
            echo "🧹 Go module cache is over ${GOMODCACHE_MAX_GB:-20}G, clearing it..."
            GOMODCACHE="$GO_MOD_CACHE" go clean -modcache >/dev/null 2>&1 || true
        fi
    fi

    prune_cache_dir "🔍 Golangci-lint" "$GOLANGCI_LINT_CACHE_DIR" 5
    prune_cache_dir "🔍 Go build" "$GO_BUILD_CACHE" 5
}

# prune_cache_dir reports a cache directory's size and deletes files older than the given number of days
prune_cache_dir() {
    local label="$1" dir="$2" days="$3"

    [[ -d "$dir" ]] || return 0

    echo "$label cache size: $(du -sh "$dir" 2>/dev/null | cut -f1 || echo unknown)"
    find "$dir" -type f -mtime +"$days" -delete 2>/dev/null || true
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
