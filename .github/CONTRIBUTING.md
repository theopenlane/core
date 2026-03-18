# Contributing

Please read the [contributing](.github/CONTRIBUTING.md) guide as well as the [Developer Certificate of Origin](https://developercertificate.org/). You will be required to sign all commits to the Openlane project, so if you're unfamiliar with how to set that up, see [github's documentation](https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification).

Given external users will not have write to the branches in this repository, you'll need to follow the forking process to open a PR - [here](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request-from-a-fork) is a guide from github on how to do so.

## Licensing

This repository contains open source software that comprises the Openlane stack which is open source software under [Apache 2.0](LICENSE). Openlane's SaaS / Cloud Services are products produced from this open source software exclusively by theopenlane, Inc. This product is produced under our published commercial terms (which are subject to change). Any logos or trademarks in our repositories in [theopenlane](https://github.com/theopenlane) organization are not covered under the Apache License and are trademarks of theopenlane, Inc.

Others are allowed to make their own distribution of this software or include this software in other commercial offerings, but cannot use any of the Openlane logos, trademarks, cloud services, etc.

## Security

We take the security of our software products and services seriously, including our commercial services and all of the open source code repositories managed through our Github Organizations, such as [theopenlane](https://github.com/theopenlane). If you believe you have found a security vulnerability in any of our repositories or in our SaaS offering(s), please report it to us through coordinated disclosure.

**Please do NOT report security vulnerabilities through public github issues, discussions, or pull requests!**

Instead, please send an email to `security@theopenlane.io` with as much information as possible to best help us understand and resolve the issues. See the security policy attached to this repository for more details.

## Tools

Developing against this repo involves a few mandatory tools; please read up on
these and familiarize yourself if you're interested in making additions or
changes!

1. [ent](https://entgo.io/) - insane entity mapping tool, definitely not an ORM
   but kind of an ORM (handles our relational data storage, mappings, codegen
   processes)
1. [atlas](https://atlasgo.io/) - Schema generation and migrations (can be
   disabled in lieu of migrations on disk)
1. [goose](https://github.com/pressly/goose) - Secondary database migration
   utility we also use for seeding data
1. [gqlgen](https://gqlgen.com/) - Code generation + GraphQL server building
   from from `ent` schema definitions
1. [gqlgenc](https://github.com/gqlgo/gqlgenc) - Client building utilities
   with GraphQL
1. [openfga](https://openfga.dev/) - Flexible authorization/permission engine
   inspired by Google Zanzibar
1. [echo](https://echo.labstack.com/) - High performance, extensible, minimalist
   Go web framework
1. [koanf](https://github.com/knadh/koanf) - Configuration management library
   which parses command line arguments, Go structs + creates our main
   configuration files

We also leverage many secondary technologies in use, including (but not limited
to!):

1. [taskfile](https://taskfile.dev/usage/) - So much better than Make zomg
1. [redis](https://redis.io/) - in-memory datastore used for sessions, caching
1. [postgres](https://www.postgresql.org/)
1. [golangci-lint](https://github.com/golangci/golangci-lint) - an annoyingly
   opinionated linter
1. [buildkite](https://buildkite.com/theopenlane) - our CI system of choice
   (with github actions providing some intermediary support)

All of these components are bundled into our respective Docker images; for
additional information / instructions, see the
[contributing guide](.github/CONTRIBUTING.md) in this repository. We're
constantly adding and changing things, but have tried to list all the great open
source tools and projects we rely on; if you see your project (or one you use)
in here and wish to list it, feel free to open a PR!

## Questions?

You can email us at `info@theopenlane.io`, open a github issue in this repository, or reach out to [matoszz](https://github.com/matoszz) directly.


