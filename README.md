<div align="center">

[![Go Report Card](https://goreportcard.com/badge/github.com/theopenlane/core)](https://goreportcard.com/report/github.com/theopenlane/core)
[![Build status](https://badge.buildkite.com/a3a38b934ca2bb7fc771e19bc5a986a1452fa2962e4e1c63bf.svg?branch=main)](https://buildkite.com/theopenlane/core)
[![Go Reference](https://pkg.go.dev/badge/github.com/theopenlane/core.svg)](https://pkg.go.dev/github.com/theopenlane/core)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)

</div>

This repository houses the core server and orchestration elements which are at the heart of the [OpenLane](https://theopenlane.io) cloud service. We have no plans to ever gate / silo elements of the code that may fall under our "enterprise licensing" (or any other commercial license we offer) and intend to keep the code Apache 2.0 licensed and free for use, forever. Given that, if you find value in anything we're doing here, our cloud services, or use this software yourself (for any purpose) - don't be afraid to become a contributor! If you have any questions please reach out to `contribute@theopenlane.io`.

## Features

At it's core, this repo is a collection of services built on top of an entity framework which allows us to:
- Model database schemas as graph structures
- Define schemas as programmatic go code
- Execute complex database queries and graph traversals easily
- Extend and customize using templates and code generation utilities
- Type-safe resolvers and GraphQL schema stitching
- Code generated audit / history tables for defined schemas

On top of this powerful core we also have an incredible amount of pluggable, extensible services:
- Authentication: we today support password, OAuth2 / Social login providers (Github, Google), Passkeys as well as standard OIDC Discovery flows
- Multi-factor: built-in 2FA mechanisms, TOTP
- Authorization: extensible and flexible permissions constructs via openFGA based on Google Zanzibar
- Session Management: built-in session management with JWKS key validation, encrypted cookies and sessions
- Robust Middleware: cache control, CORS, Rate Limiting, transaction rollbacks, and more
- Queuing and Scheduling: Task management and scheduling with Marionette
- External Storage Providers: store data in AWS S3, Google GCS, or locally
- External Database Providers: Leverage Turso, or other PostgreSQL / SQLite compatible vendors and libraries
- Data Isolation and Management: Hierarchal organizations and granular permissions controls

## Development

Developing against this repo involves a few mandatory tools; please read up on these and familiarize yourself if you're interested in making additions or changes!

1. [ent](https://entgo.io/) - insane entity mapping tool, definitely not an ORM but kind of an ORM (handles our relational data storage, mappings, codegen processes)
1. [atlas](https://atlasgo.io/) - Schema generation and migrations (can be disabled in lieu of migrations on disk)
1. [goose](https://github.com/pressly/goose) - Secondary database migration utility we also use for seeding data
1. [gqlgen](https://gqlgen.com/) - Code generation + GraphQL server building from from `ent` schema definitions
1. [gqlgenc](https://github.com/Yamashou/gqlgenc) - Client building utilities with GraphQL
1. [openfga](https://openfga.dev/) - Flexible authorization/permission engine inspired by Google Zanzibar
1. [echo](https://echo.labstack.com/) - High performance, extensible, minimalist Go web framework
1. [koanf](https://github.com/knadh/koanf) - Configuration management library which parses command line arguments, Go structs + creates our main configuration files

We also leverage many secondary technologies in use, including (but not limited to!):

1. [taskfile](https://taskfile.dev/usage/) - So much better than Make zomg
1. [redis](https://redis.io/) - in-memory datastore used for sessions, caching
1. databases:
    * [postgres](https://www.postgresql.org/)
    * [libsql](https://turso.tech/libsql)
    * [sqlite](https://www.sqlite.org/)
1. [golangci-lint](https://github.com/golangci/golangci-lint) - an annoyingly opinionated linter
1. [buildkite](https://buildkite.com/theopenlane) - our CI system of choice (with github actions providing some intermediary support)

Lastly we're already ourselves using (and plan to support our customers usage in our cloud service) these third party integrations:

1. [turso/libsql](https://github.com/tursodatabase/libsql) - Turso is an edge-hosted, distributed database that's based on libSQL , an open-source and open-contribution fork of SQLite
1. [posthog](https://posthog.com/) - Product analytics
1. [sendgrid](https://sendgrid.com/en-us) - Transactional email send provider

All of these components are bundled into our respective Docker images; for additional information / instructions, see the [contributing guide](.github/CONTRIBUTING.md) in this repository. We're constantly adding and changing things, but have tried to list all the great open source tools and projects we rely on; if you see your project (or one you use) in here and wish to list it, feel free to open a PR!

## Dependencies

The vast majority of behaviors of the system can be turned on or off by updating the configuration parameters found in `config`; in some instances, we've made features or integrations with third party systems which are "always on", but we're happy to receive PR's wrapping those dependencies if you are interested in running the software without them!

### Installing Dependencies

Setup [Taskfile](https://taskfile.dev/installation/) by following the instructions and using one of the various convenient package managers or installation scripts. After installation, you can then simply run `task install` to load the associated dependencies. Nearly everything in this repository assumes you already have a local golang environment setup so this is not included. Please see the associated documentation.

### Updating Configuration Settings

See the [README](/config/README.md) in the `config` directory.

## Deploying

The only "supported" method of deploying today is locally, but we have a WIP Helm chart which can be found [here](https://github.com/theopenlane/helm-charts)

## Contributing

Please read the [contributing](.github/CONTRIBUTING.md) guide as well as the [Developer Certificate of Origin](https://developercertificate.org/). You will be required to sign all commits to the The OpenLane project, so if you're unfamiliar with how to set that up, see [github's documentation](https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification).

## Licensing

This repository contains `core` which is open source software under [Apache 2.0](LICENSE). The OpenLane is a product produced from this open source software exclusively by The Open Lane, Inc. This product is produced under our published commercial terms (which are subject to change), and any logos or trademarks in this repository or the broader [theopenlane](https://github.com/theopenlane) organization are not covered under the Apache License.

Others are allowed to make their own distribution of this software or include this software in other commercial offerings, but cannot use any of the The OpenLane logos, trademarks, cloud services, etc.

## Security

We take the security of our software products and services seriously, including all of the open source code repositories managed through our Github Organizations, such as [theopenlane](https://github.com/theopenlane). If you believe you have found a security vulnerability in any of our repositories, please report it to us through coordinated disclosure.

**Please do NOT report security vulnerabilities through public github issues, discussions, or pull requests!**

Instead, please send an email to `security@theopenlane.io` with as much information as possible to best help us understand and resolve the issues. See the security policy attached to this repository for more details.

## Questions?

You can email us at `info@theopenlane.io`, open a github issue in this repository, or reach out to [matoszz](https://github.com/matoszz) directly.
