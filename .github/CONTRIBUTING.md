# Contributing

Given external users will not have write to the branches in this repository, you'll need to follow the forking process to open a PR - [here](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request-from-a-fork) is a guide from github on how to do so.

Please also read our main [contributing guide](https://github.com/theopenlane/.github/blob/main/CONTRIBUTING.md) in addition to this one; the main guide mostly says that we'd like for you to open an issue first but it's not hard-required, and that we accept all forms of proposed changes given the state of this code base (in it's infancy, still!)

## Pre-requisites to a PR

This repository contains a number of code generating functions / utilities which take schema modifications and scaffold out resolvers, graphql API schemas, openAPI specifications, among other things. To ensure you've generated all the necessary dependencies run `task pr`; this will run the entirety of the commands required to safely generate a PR. If for some reason one of the commands fails / encounters an error, you will need to debug the individual steps. It should be decently easy to follow the `Taskfile` in the root of this repository.

### Pre-Commit Hooks

We have several `pre-commit` hooks that should be run before pushing a commit. Make sure this is installed:

```bash
brew install pre-commit
pre-commit install
```

You can optionally run against all files:

```bash
pre-commit run --all-files
```

## Starting the Server

1. Copy the config, this is in .gitignore so you do not have to worry about accidentally committing secrets

   ```bash
   cp ./config/config-dev.example.yaml ./config/.config.yaml
   ```

1. Update the configuration with whatever respective settings you desire; the defaults inside should allow you to run the server without a problem

1. Use the task commands to start the server

   Run the core server in development mode with dependencies in docker

   ```bash
   task run-dev
   ```

   Run fully in docker

   ```bash
   task docker:all:up
   ```

1. In a separate terminal, with the server running, you can create a verified test user by running:

   ```bash
   task cli:user:all
   ```

1. Once this command has finished ^, you can login and perform actions as user `mitb@theopenlane.io` with password `mattisthebest1234!`

## Creating Queries in GraphQL

The best method of forming / testing queries against the server is to run `task docker:rover` which will launch an interactive query UI.

If you are running the queries against your local repo, you will have CORS issues using the local running apollo. Instead, its recommended to use the [apollo sandbox](https://studio.apollographql.com/sandbox/explorer) and ensure the following origin is allowed in your `config/.config.yaml`

```
server:
  cors:
    allowOrigins:
      - https://studio.apollographql.com
```

## OpenFGA Playground

You can load up a local openFGA environment with the compose setup in this repository; `task fga:up` - this will launch an interactive playground where you can model permissions model(s) or changes to the models

## Creating a new Schema

To ease the effort required to add additional schemas into the system a template + task function has been created. This isn't doing anything terribly complex, but it's attempting to ensure you have the _minimum_ set of required things needed to create a schema - most notably: you need to ensure the IDMixin is present (otherwise you will get ID type conflicts) and a standard set of schema annotations.

NOTE: you still have to make intelligent decisions around things like the presence / integration of hooks, interceptors, policies, etc. This is saving you about 10 seconds of copy-paste, so don't over estimate the automation, here.

To generate a new schema, you can run `task newschema -- [yourschemaname]` where you replace the name within `[]`. Please be sure to note that this isn't a command line flag so there's a space between `--` and the name.

### Migrations

We use [atlas](https://atlasgo.io/) and [goose](https://github.com/pressly/goose) to create and manage our DB migrations - you can trigger one via `task atlas:create` and that will generate the necessary migrations. There should be a new migration file created in `db/migrations`, `db/migrations-goose-postgre` and `db/migrations-goose-sqlite`. On every PR, the Atlas integration also creates comments with any issues related to the schema changes / migrations.
