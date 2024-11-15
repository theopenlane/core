<div align="center">

[![Go Report Card](https://goreportcard.com/badge/github.com/theopenlane/core)](https://goreportcard.com/report/github.com/theopenlane/core)
[![Build status](https://badge.buildkite.com/a3a38b934ca2bb7fc771e19bc5a986a1452fa2962e4e1c63bf.svg?branch=main)](https://buildkite.com/theopenlane/core)
[![Go Reference](https://pkg.go.dev/badge/github.com/theopenlane/core.svg)](https://pkg.go.dev/github.com/theopenlane/core)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=theopenlane_core&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=theopenlane_core)

</div>

# openlane

This repository houses the core server and orchestration elements which are at
the heart of the [openlane](https://theopenlane.io) cloud service. 

## Features

At it's core, this repo is a collection of services built on top of an entity
framework which allows us to:

- Model database schemas as graph structures
- Define schemas as programmatic go code
- Execute complex database queries and graph traversals easily
- Extend and customize using templates and code generation utilities
- Type-safe resolvers and GraphQL schema stitching
- Code generated audit / history tables for defined schemas

On top of this powerful core we also have an incredible amount of pluggable,
extensible services:

- Authentication: we today support password, OAuth2 / Social login providers
  (Github, Google), Passkeys as well as standard OIDC Discovery flows
- Multi-factor: built-in 2FA mechanisms, TOTP
- Authorization: extensible and flexible permissions constructs via openFGA
  based on Google Zanzibar
- Session Management: built-in session management with JWKS key validation,
  encrypted cookies and sessions
- Robust Middleware: cache control, CORS, Rate Limiting, transaction rollbacks,
  and more
- Queuing and Scheduling: Task management and scheduling with
  [riverqueue](https://github.com/riverqueue/river)
- External Storage Providers: store data in AWS S3, Google GCS, or locally
- External Database Providers: Leverage NeonDB, or other PostgreSQL compatible
  vendors and libraries
- Data Isolation and Management: Hierarchal organizations and granular
  permissions controls

## Development

### Dependencies

The vast majority of behaviors of the system can be turned on or off by updating
the configuration parameters found in `config`; in some instances, we've made
features or integrations with third party systems which are "always on", but
we're happy to receive PR's wrapping those dependencies if you are interested in
running the software without them!

### Installing Dependencies

Setup [Taskfile](https://taskfile.dev/installation/) by following the
instructions and using one of the various convenient package managers or
installation scripts. After installation, you can then simply run `task install`
to load the associated dependencies. Nearly everything in this repository
assumes you already have a local golang environment setup so this is not
included. Please see the associated documentation.

### Updating Configuration Settings

See the [README](/config/README.md) in the `config` directory.

### Starting the Server

1. Copy the config, this is in .gitignore so you do not have to worry about
   accidentally committing secrets

   ```bash
   cp ./config/config-dev.example.yaml ./config/.config.yaml
   ```

1. Update the configuration with whatever respective settings you desire; the
   defaults inside should allow you to run the server without a problem

1. Use the task commands to start the server

   Run the core server in development mode with dependencies in docker

   ```bash
   task run-dev
   ```

   Run fully in docker

   ```bash
   task docker:all:up
   ```

1. In a separate terminal, with the server running, you can create a verified
   test user by running:

   ```bash
   task cli:user:all
   ```

1. Once this command has finished ^, you can login and perform actions as user
   `mitb@theopenlane.io` with password `mattisthebest1234`

### Creating Queries in GraphQL

The best method of forming / testing queries against the server is to run
`task docker:rover` which will launch an interactive query UI.

If you are running the queries against your local repo, you will have CORS
issues using the local running apollo. Instead, its recommended to use the
[apollo sandbox](https://studio.apollographql.com/sandbox/explorer) and ensure
the following origin is allowed in your `config/.config.yaml`

```
server:
  cors:
    allowOrigins:
      - https://studio.apollographql.com
```

### OpenFGA Playground

You can load up a local openFGA environment with the compose setup in this
repository; `task fga:up` - this will launch an interactive playground where you
can model permissions model(s) or changes to the models

### Creating a new Schema

To ease the effort required to add additional schemas into the system a
template + task function has been created. This isn't doing anything terribly
complex, but it's attempting to ensure you have the _minimum_ set of required
things needed to create a schema - most notably: you need to ensure the IDMixin
is present (otherwise you will get ID type conflicts) and a standard set of
schema annotations.

NOTE: you still have to make intelligent decisions around things like the
presence / integration of hooks, interceptors, policies, etc. This is saving you
about 10 seconds of copy-paste, so don't over estimate the automation, here.

To generate a new schema, you can run `task newschema -- [yourschemaname]` where
you replace the name within `[]`. Please be sure to note that this isn't a
command line flag so there's a space between `--` and the name.

### Migrations

We use [atlas](https://atlasgo.io/) and
[goose](https://github.com/pressly/goose) to create and manage our DB
migrations - you can trigger one via `task atlas:create` and that will generate
the necessary migrations. There should be a new migration file created in
`db/migrations` and `db/migrations-goose-postgres`. On every PR, the Atlas
integration also creates comments with any issues related to the schema changes
/ migrations.

## Deploying

The only "supported" method of deploying today is locally, but we have a WIP
Helm chart which can be found [here](https://github.com/theopenlane/helm-charts)

## Contributing

See the [contributing](.github/CONTRIBUTING.md) guide for more information
