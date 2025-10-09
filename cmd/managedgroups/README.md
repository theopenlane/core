# Managed Groups CLI Tool

This directory contains a command line utility used for managing user groups in
organizations. This ensures that each organization member has a managed group named
after them in their org.

Ideally this will only be run once if you had the
openlane project running before `v0.37`.

## Overview

## Usage

```bash
go run cmd/managedgroups/main.go [options]
```

### Flags

- `--config` – Path to the config YAML file. Default is `./config/.config.yaml`. Can also be set via `CORE_CONFIG` environment variable.
- `--dry-run` – Preview changes without making them (default: true). Set to `false` to actually create groups and memberships.
- `--debug` – Enable debug logging.

### Examples

```bash
# dry-run by default (preview changes without making them)
go run cmd/managedgroups/main.go

# make changes
go run cmd/managedgroups/main.go --dry-run=false

# use custom config
go run cmd/managedgroups/main.go --config /path/to/config.yaml
```

## What It Does

1. **Finds Organizations**: Queries all non-personal organizations in the database
2. **Processes Members**: For each organization, finds all members
3. **Creates Groups**: For each member, creates a managed group named after their display name (if it doesn't already exist)
4. **Adds Memberships**: Ensures each user is a member of their own managed group

