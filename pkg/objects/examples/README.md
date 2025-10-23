# Object Storage Examples

This directory contains examples demonstrating the object storage capabilities of Openlane.

## Quick Start

TL;DR: run `task run-dev` to bring an instance of the server up, and then `go run pkg/objects/examples/main.go setup --force` to seed all the relevant objects and configs, and then `go run pkg/objects/examples/main.go openlane` to demonstrate full use of the storage package.

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
- Save the configuration

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

See the output of the CLI for all the commands available to you.
