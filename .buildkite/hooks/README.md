# Buildkite Hooks

This directory contains Buildkite hooks that run at different stages of the pipeline execution.

## Hook Execution Order

1. **`environment`** - Sets up dynamic environment variables
2. **`pre-command`** - Runs before each pipeline step
3. **[Pipeline Step]** - The actual command from pipeline.yaml
4. **`post-command`** - Runs after each pipeline step

## Hook Descriptions

### `environment`
**When**: Runs once at the start of each job, before any commands
**Purpose**: Sets up dynamic environment variables based on build context

**Features**:
- ✅ Go environment configuration (GOPROXY, GOSUMDB, cache paths)
- ✅ Build metadata (timestamp, version, commit, branch)
- ✅ CI vs non-CI context detection
- ✅ PR-specific variables and automation flags
- ✅ Test and artifact directory setup
- ✅ Docker environment configuration
- ✅ Tool-specific settings (golangci-lint, security scanners)
- ✅ Resource limits and performance tuning
- ✅ Debug mode support with verbose logging

### `pre-command`
**When**: Runs before each pipeline step command
**Purpose**: Ensures the build environment is ready

**Features**:
- ✅ Go build cache directories
- ✅ Docker layer caching setup
- ✅ Build tool availability checks (Go, Git)
- ✅ Environment validation
- ✅ Temporary directory creation
- ✅ Disk space monitoring and emergency cleanup
- ✅ Clear status reporting

### `post-command`
**When**: Runs after each pipeline step command
**Purpose**: Cleanup and artifact management

**Features**:
- ✅ Temporary file cleanup
- ✅ Cache pruning (Go modules, golangci-lint)
- ✅ Build artifact archival
- ✅ Test result collection
- ✅ Comprehensive Docker cleanup (containers, images, volumes, networks)
- ✅ Aggressive disk space management
- ✅ Disk usage monitoring and reporting
- ✅ Failure status reporting

## Environment Variables Set

### Build Context
- `BUILD_TIMESTAMP` - ISO 8601 timestamp of build start
- `BUILD_VERSION` - Build number (from Buildkite or 'dev')
- `BUILD_COMMIT` - Git commit hash
- `BUILD_BRANCH` - Git branch name
- `IS_CI` - Whether running in CI context (true/false)

### Automation Flags
- `HELM_AUTOMATION_ENABLED` - Whether to run Helm automation (main branch only)
- `DRAFT_PR_AUTOMATION_ENABLED` - Whether to create draft PRs (PR context only)
- `IS_PULL_REQUEST` - Whether this is a pull request build
- `PR_NUMBER` - Pull request number (if applicable)

### Tool Configuration
- `GOPROXY`, `GOSUMDB`, `GOMODCACHE`, `GOCACHE` - Go toolchain
- `DOCKER_BUILDKIT`, `COMPOSE_DOCKER_CLI_BUILD` - Docker optimization
- `GOMAXPROCS` - Go runtime parallelism
- `GOSEC_SEVERITY`, `GOSEC_CONFIDENCE` - Security scanning

### Directory Paths
- `TEST_RESULTS_DIR` - Where to store test results
- `ARTIFACTS_DIR` - Where to store build artifacts
- `COVERAGE_DIR` - Where to store coverage reports

### Integration Flags
- `SLACK_NOTIFICATIONS_ENABLED` - Whether Slack notifications are configured
- `GITHUB_API_ENABLED` - Whether GitHub API access is configured
- `DEBUG_MODE` - Whether debug mode is enabled
- `VERBOSE_LOGGING` - Whether verbose logging is enabled

## Testing Hooks

To test the hooks locally:

```bash
# Test environment hook
./.buildkite/hooks/environment

# Test pre-command hook
./.buildkite/hooks/pre-command

# Test post-command hook
BUILDKITE_COMMAND_EXIT_STATUS=0 ./.buildkite/hooks/post-command
```

## Debugging

Enable debug mode by setting:
```bash
export BUILDKITE_AGENT_DEBUG=true
```

This will:
- Enable verbose logging
- Show all environment variables set by hooks
- Enable bash tracing (`set -x`)

## Hook Development

When modifying hooks:

1. **Keep them fast** - Hooks run before/after every command
2. **Handle failures gracefully** - Use `|| true` for non-critical operations
3. **Use proper error handling** - Always use `set -euo pipefail`
4. **Test thoroughly** - Test hooks with different build contexts
5. **Document changes** - Update this README for new features

## Integration with Automation Scripts

The hooks set up environment variables that the automation scripts use:

- `draft-pr-automation.sh` - Uses `DRAFT_PR_AUTOMATION_ENABLED`
- `helm-automation.sh` - Uses `HELM_AUTOMATION_ENABLED`
- `slack-utils.sh` - Uses `SLACK_NOTIFICATIONS_ENABLED`
- All scripts - Use `BUILD_*` variables for context

This ensures consistent behavior across the entire CI/CD pipeline.

## Disk Space Management

The hooks include comprehensive disk space management to prevent build failures due to full disks:

### Automatic Cleanup in Hooks

#### **`pre-command` Disk Monitoring**
- Checks disk usage before each build step
- **>90% usage**: Triggers emergency Docker cleanup
- **>80% usage**: Shows warning but continues
- **<80% usage**: Normal operation

#### **`post-command` Docker Cleanup**
Runs after every build step with progressive cleanup:

1. **Standard Cleanup**:
   - Stop running containers
   - Remove stopped containers
   - Remove dangling images
   - Remove unused images >24h old
   - Remove unused networks
   - Remove ALL unused volumes
   - Remove build cache >24h old

2. **Aggressive Cleanup** (if disk >80%):
   - Remove ALL unused images
   - Remove ALL build cache
   - Full system prune with volumes

3. **Reporting**:
   - Shows before/after disk usage
   - Reports space freed
   - Docker space usage summary

### Scheduled Cleanup Script

**`docker-cleanup.sh`** - Standalone script for periodic maintenance:

```bash
# Make executable
chmod +x .buildkite/docker-cleanup.sh

# Run manually
./.buildkite/docker-cleanup.sh

# Add to cron for regular cleanup
# Every 6 hours:
0 */6 * * * /path/to/buildkite/docker-cleanup.sh

# Daily at 2 AM:
0 2 * * * /path/to/buildkite/docker-cleanup.sh
```

#### **Cleanup Phases**:

1. **Phase 1 - Standard**: Remove exited containers, dangling images, unused volumes
2. **Phase 2 - Aggressive**: Remove unused images >24h old
3. **Phase 3 - Emergency** (disk >85%): Remove ALL unused Docker resources
4. **Phase 4 - Critical** (disk >90%): Stop/remove ALL containers and images (except base images)

#### **Base Image Protection**:
The script preserves commonly used base images:
- `ubuntu`, `alpine`, `node`, `golang`, `python`
- `postgres`, `redis`, `nginx`

### Monitoring and Alerts

The cleanup system provides detailed logging:

```bash
📊 Initial disk usage: 78%
🐳 Running Docker cleanup to prevent disk space issues...
🛑 Stopping running containers...
🗑️  Removing stopped containers...
🖼️  Removing dangling images...
💾 Removing ALL unused volumes...
📊 Final disk usage: 65%
✨ Freed 13% disk space
```

### Troubleshooting Disk Issues

If builds are still failing due to disk space:

1. **Check runner disk capacity**: May need larger disks
2. **Run manual cleanup**: `./buildkite/docker-cleanup.sh`
3. **Check for large files**: `du -sh /var/lib/docker/*`
4. **Monitor build logs**: Look for disk usage warnings
5. **Consider runner maintenance**: Regular restart/cleanup schedule

### Configuration

You can customize disk thresholds by modifying the hooks:

```bash
# In pre-command and post-command hooks
WARNING_THRESHOLD=80    # Show warnings
CRITICAL_THRESHOLD=90   # Emergency cleanup
AGGRESSIVE_THRESHOLD=85 # Aggressive cleanup in post-command
```