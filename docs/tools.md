# Tools

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

All of these components are bundled into our respective Docker images; for additional information / instructions, see the [contributing guide](.github/CONTRIBUTING.md) in this repository. We're constantly adding and changing things, but have tried to list all the great open source tools and projects we rely on; if you see your project (or one you use) in here and wish to list it, feel free to open a PR!