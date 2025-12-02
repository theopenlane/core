# Database Support

## Dependencies

1. [ent](https://entgo.io/) - insane entity mapping tool, definitely not an ORM
   but kind of an ORM
1. [atlas](https://atlasgo.io/) - Schema generation and migration
1. [entx](https://github.com/theopenlane/entx) - Wrapper to interact with the
   ent

## Supported Drivers

1. [postgres](https://github.com/lib/pq)
1. [pgx](https://github.com/jackc/pgx)

## Local Development

### Config Examples

#### Postgres

1. Postgres is included in `docker/docker-compose-fga.yml` and the same instance
   can be used for development. The following connection string should work when
   using `task docker:all:up`

```yaml
db:
  debug: true
  driver_name: pgx # or `postgres` to use lib/pg instead
  primary_db_source: "postgres://postgres:password@postgres:5432?sslmode=disable"
  run_migrations: true
```
