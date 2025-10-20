# Object Storage Examples

This directory contains examples demonstrating the object storage capabilities of Openlane.

## Quick Start

### Initial Setup

Before running any Openlane-integrated examples, initialize your configuration:

```bash
go run main.go setup
```

This command will:
- Register a new user account
- Verify the user
- Create an organization
- Generate a Personal Access Token (PAT)
- Save the configuration to `~/.openlane-examples.json`

### Running Examples

Once setup is complete, you can run any of the available examples:

```bash
# Simple disk-based storage
go run main.go simple

# S3/MinIO storage
go run main.go simple-s3

# Multi-provider example (disk + multiple S3 backends)
go run main.go multi-provider

# High-throughput benchmark
go run main.go multi-provider benchmark

# Openlane integration example
go run main.go openlane
```

## Available Commands

### `setup`

Initialize Openlane configuration for examples.

**Flags:**
- `--api` - Openlane API base URL (default: http://localhost:17608)
- `--force` - Force re-initialization even if configuration exists

**Example:**
```bash
go run main.go setup --api http://localhost:17608
```

### `simple`

Run local disk object storage example.

**Flags:**
- `--dir` - Directory to use for disk storage (default: ./tmp/storage)
- `--local-url` - Local URL for presigned links (default: http://localhost:17608/v1/files)
- `--keep` - Keep storage directory after completion

### `simple-s3`

Run S3/MinIO object storage example.

**Flags:**
- `--endpoint` - S3 or MinIO endpoint URL (default: http://127.0.0.1:9000)
- `--access-key` - Access key ID (default: minioadmin)
- `--secret-key` - Secret access key (default: minioadmin)
- `--region` - AWS region (default: us-east-1)
- `--bucket` - Bucket name (default: core-simple-s3)
- `--source` - Path to file to upload (default: assets/sample-data.txt)
- `--object` - Object key in bucket (default: examples/simple-s3/sample-data.txt)
- `--download` - Destination path for downloaded file (default: output/downloaded-sample.txt)
- `--path-style` - Use path-style addressing (default: true)

### `multi-provider`

Demonstrate provider resolution across disk and S3 backends with optional benchmarking.

**Flags:**
- `--skip-setup` - Assume infrastructure is already running
- `--skip-teardown` - Leave services running after example

**Subcommands:**

#### `setup`
Start docker services and seed credentials.

```bash
go run main.go multi-provider setup
```

#### `teardown`
Stop docker services and remove containers.

```bash
go run main.go multi-provider teardown
```

#### `benchmark`
Run high-throughput benchmark across multiple isolated storage contexts.

**Flags:**
- `--ops` - Number of operations per tenant (default: 100)
- `--concurrent` - Number of concurrent workers (default: 10)
- `--tenants` - Number of tenants to provision (default: 10)
- `--parallel` - Number of parallel provisioning workers (default: 5)
- `--config` - Path to tenant configuration file (default: tenants.json)

**Benchmark Subcommands:**
- `setup` - Provision tenants for benchmarking
  - `--tenants` - Number of tenants (default: 10)
  - `--parallel` - Parallel workers (default: 5)
- `setup-1000` - Provision 1000 tenants for large-scale testing
- `teardown` - Remove tenants and services

**Examples:**
```bash
# Run basic multi-provider demo
go run main.go multi-provider

# Setup benchmark tenants
go run main.go multi-provider benchmark setup --tenants 10

# Run benchmark
go run main.go multi-provider benchmark --ops 100 --concurrent 10

# Run large-scale benchmark
go run main.go multi-provider benchmark setup-1000
go run main.go multi-provider benchmark --ops 50 --concurrent 20

# Cleanup
go run main.go multi-provider benchmark teardown
```

### `openlane`

Run Openlane end-to-end integration example.

**Flags:**
- `--api` - Openlane API base URL (default: http://localhost:17608)
- `--token` - Authentication token (uses saved config if not provided)
- `--organization-id` - Organization ID (uses saved config if not provided)
- `--name` - Evidence name (default: Security Compliance Evidence)
- `--description` - Evidence description (default: Security compliance validation evidence)
- `--file` - Path to evidence file
- `--verbose` - Enable verbose logging

**Example:**
```bash
# Uses saved configuration from setup
go run main.go openlane

# With custom file
go run main.go openlane --file path/to/document.pdf

# With verbose output
go run main.go openlane --verbose
```

## Architecture

### Common Package

The `common` package provides shared infrastructure utilities:
- Docker command execution
- MinIO setup and configuration
- GCS fake server setup
- S3 client creation
- Bucket management

### Openlane Package

The `openlane` package handles Openlane-specific functionality:
- User registration and verification
- Organization management
- Personal Access Token (PAT) creation
- Configuration persistence
- Client initialization
- Evidence and file upload

### Configuration

The setup command stores configuration in `~/.openlane-examples.json`:

```json
{
  "email": "example@openlane.io",
  "password": "...",
  "token": "...",
  "organization_id": "...",
  "pat": "...",
  "base_url": "http://localhost:17608"
}
```

Override the configuration file location with the `OPENLANE_EXAMPLES_CONFIG` environment variable.

## Multi-Provider and Benchmarking

The `multi-provider` command demonstrates Openlane's provider resolution capabilities:

1. **Basic Mode** (default): Demonstrates switching between different storage providers (disk, S3 provider 1, 2, 3) with concurrent operations
2. **Benchmark Mode** (`multi-provider benchmark`): High-throughput performance testing across many isolated storage contexts

Both modes use the same underlying resolver pattern to switch between storage backends based on context. The benchmark mode provisions multiple MinIO users/buckets to test throughput and caching behavior under load.

## Multi-Tenancy

Openlane is implicitly multi-tenant. All Openlane-integrated examples automatically include organization context via headers, providing complete tenant isolation without requiring separate code paths or examples.

## Docker Services

The `multi-provider` examples require Docker services (MinIO, fake GCS). These are automatically started unless `--skip-setup` is specified.
