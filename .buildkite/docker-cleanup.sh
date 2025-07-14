#!/bin/bash
#
# Docker Cleanup Script for Buildkite Runners
#
# This script can be run periodically (e.g., via cron) to maintain disk health
# on Buildkite runners by aggressively cleaning up Docker resources.
#
# Recommended cron schedule:
# # Run every 6 hours
# 0 */6 * * * /path/to/docker-cleanup.sh
#
# # Or run daily at 2 AM
# 0 2 * * * /path/to/docker-cleanup.sh

set -euo pipefail

echo "🐳 Starting scheduled Docker cleanup for Buildkite runners..."
echo "⏰ Started at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"

# Check if Docker is available
if ! command -v docker >/dev/null 2>&1; then
    echo "❌ Docker not found - skipping cleanup"
    exit 0
fi

# Get initial disk usage and Docker space usage
initial_usage=$(df -h / | tail -1 | awk '{print $5}' | sed 's/%//')
echo "📊 Initial disk usage: ${initial_usage}%"

echo "🐳 Initial Docker space usage:"
docker system df 2>/dev/null || true

# Function to get current disk usage
get_disk_usage() {
    df -h / | tail -1 | awk '{print $5}' | sed 's/%//'
}

# Function to report space freed
report_space_freed() {
    local initial=$1
    local current=$(get_disk_usage)
    local freed=$((initial - current))

    if [[ $freed -gt 0 ]]; then
        echo "✨ Freed ${freed}% disk space (${initial}% → ${current}%)"
    else
        echo "📊 Disk usage: ${current}%"
    fi
}

echo ""
echo "🧹 Phase 1: Standard cleanup..."

# Stop containers that have been running for more than 1 hour
echo "🛑 Stopping long-running containers (>1 hour)..."
long_running=$(docker ps --filter "status=running" --format "table {{.ID}}\t{{.RunningFor}}" | grep -E "(hour|day|week|month)" | awk '{print $1}' | grep -v CONTAINER || true)
if [[ -n "$long_running" ]]; then
    echo "$long_running" | xargs docker stop >/dev/null 2>&1 || true
    echo "   Stopped $(echo "$long_running" | wc -l) long-running containers"
else
    echo "   No long-running containers found"
fi

# Remove exited containers
echo "🗑️  Removing exited containers..."
exited_containers=$(docker ps -aq --filter "status=exited")
if [[ -n "$exited_containers" ]]; then
    echo "$exited_containers" | xargs docker rm >/dev/null 2>&1 || true
    echo "   Removed $(echo "$exited_containers" | wc -l) exited containers"
else
    echo "   No exited containers found"
fi

# Remove dangling images
echo "🖼️  Removing dangling images..."
dangling_count=$(docker images -f "dangling=true" -q | wc -l)
if [[ $dangling_count -gt 0 ]]; then
    docker image prune -f >/dev/null 2>&1 || true
    echo "   Removed $dangling_count dangling images"
else
    echo "   No dangling images found"
fi

# Remove unused networks
echo "🌐 Removing unused networks..."
docker network prune -f >/dev/null 2>&1 || true

# Remove unused volumes
echo "💾 Removing unused volumes..."
volume_count=$(docker volume ls -f "dangling=true" -q | wc -l)
if [[ $volume_count -gt 0 ]]; then
    docker volume prune -f >/dev/null 2>&1 || true
    echo "   Removed $volume_count unused volumes"
else
    echo "   No unused volumes found"
fi

# Remove build cache older than 24 hours
echo "🏗️  Removing build cache older than 24 hours..."
docker builder prune -f --filter "until=24h" >/dev/null 2>&1 || true

report_space_freed $initial_usage

echo ""
echo "🧹 Phase 2: Aggressive cleanup (images older than 24 hours)..."

# Remove images older than 24 hours that aren't being used
echo "🖼️  Removing unused images older than 24 hours..."
docker image prune -f --filter "until=24h" >/dev/null 2>&1 || true

report_space_freed $initial_usage

# Check if we need even more aggressive cleanup
current_usage=$(get_disk_usage)
if [[ $current_usage -gt 85 ]]; then
    echo ""
    echo "🚨 Phase 3: Emergency cleanup (disk usage ${current_usage}% > 85%)..."

    # Remove ALL unused images
    echo "🖼️  Removing ALL unused images..."
    docker image prune -a -f >/dev/null 2>&1 || true

    # Remove ALL build cache
    echo "🏗️  Removing ALL build cache..."
    docker builder prune -a -f >/dev/null 2>&1 || true

    # Full system prune
    echo "🧹 Running full system prune..."
    docker system prune -a -f --volumes >/dev/null 2>&1 || true

    report_space_freed $initial_usage
fi

# If still critically high, remove more aggressively
current_usage=$(get_disk_usage)
if [[ $current_usage -gt 90 ]]; then
    echo ""
    echo "🆘 Phase 4: Critical cleanup (disk usage ${current_usage}% > 90%)..."

    # Stop ALL containers except those with specific labels
    echo "🛑 Stopping ALL containers..."
    all_containers=$(docker ps -q)
    if [[ -n "$all_containers" ]]; then
        echo "$all_containers" | xargs docker stop >/dev/null 2>&1 || true
        echo "   Stopped $(echo "$all_containers" | wc -l) containers"
    fi

    # Remove ALL containers
    echo "🗑️  Removing ALL containers..."
    all_containers=$(docker ps -aq)
    if [[ -n "$all_containers" ]]; then
        echo "$all_containers" | xargs docker rm -f >/dev/null 2>&1 || true
        echo "   Removed $(echo "$all_containers" | wc -l) containers"
    fi

    # Remove ALL images except base images
    echo "🖼️  Removing ALL images except base images..."
    # Keep common base images that are frequently used
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}" | \
        grep -v -E "(ubuntu|alpine|node|golang|python|postgres|redis|nginx):(latest|[0-9])" | \
        tail -n +2 | awk '{print $3}' | \
        xargs docker rmi -f >/dev/null 2>&1 || true

    report_space_freed $initial_usage
fi

echo ""
echo "📊 Final cleanup summary:"
final_usage=$(get_disk_usage)
total_freed=$((initial_usage - final_usage))

echo "🐳 Final Docker space usage:"
docker system df 2>/dev/null || true

echo ""
echo "📈 Cleanup Results:"
echo "   Initial disk usage: ${initial_usage}%"
echo "   Final disk usage:   ${final_usage}%"
if [[ $total_freed -gt 0 ]]; then
    echo "   Total space freed:  ${total_freed}%"
    echo "   ✅ Cleanup successful!"
else
    echo "   ℹ️  No significant space freed"
fi

if [[ $final_usage -gt 90 ]]; then
    echo "   🚨 WARNING: Disk usage still critically high!"
    echo "   🔧 Consider manual intervention or expanding disk space"
elif [[ $final_usage -gt 80 ]]; then
    echo "   ⚠️  WARNING: Disk usage still high - monitor closely"
else
    echo "   ✅ Disk usage is healthy"
fi

echo ""
echo "⏰ Completed at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
echo "🐳 Docker cleanup complete!"