# CLI Spec Architecture

The CLI is now fully spec-driven: every resource under `cmd/cli/cmd/*`
registers a JSON spec that the shared factory (`cmd/cli/internal/speccli`)
turns into Cobra commands. This guide explains how the pieces fit together and
documents the few intentional exceptions that still call REST endpoints.

## Directory Layout

- `cmd/cli/cmd/<resource>/spec.json` – declarative command definition
  describing list/get/create/update/delete metadata, table columns, and hooks.
- `cmd/cli/cmd/<resource>/register.go` – loads the embedded spec and registers it with
  `speccli`. Use the hook maps in `LoaderOptions` when you need bespoke behavior.
- `cmd/cli/cmd/<resource>/hooks.go` – optional per-resource logic (file uploads,
  REST fallbacks, additional output shaping, etc.).
- `cmd/cli/internal/speccli` – shared runtime: spec loader, command factory,
  JSON/table rendering, enum parsing, validation errors, file uploads, etc.
- `pkg/gencmd` – generator that emits `doc.go`, `register.go`, `spec.json`,
  and an `overrides.go` stub. Legacy Cobra templates have been removed; run the
  generator with `--spec`.

## Generator Workflow

1. `task cli:generate -- --spec --name MyResource`
2. The generator introspects `pkg/openlaneclient` to derive flag metadata and
   `internal/graphapi/query/*.graphql` to seed table columns.
3. Add any bespoke logic in `hooks.go` (file uploads, preHooks, etc.) and
   register them in `register.go`.
4. Import the package in `cmd/cli/main.go` (spec imports are kept in alpha
   order).

## Manual Commands / Exceptions

Most commands are spec-backed, but four remain bespoke because they drive
REST/OAuth flows that do not map cleanly to GraphQL or the spec pipeline:

1. `login` – handles OAuth/token exchange and session cookie storage.
2. `register` – posts to the registration API.
3. `reset` – drives the password reset API.
4. `invite accept` – calls the invite-accept REST endpoint and stores tokens.

These stay as manual Cobra commands until we redesign those flows; everything
else runs through `speccli`.

## Hooks vs. Pure Specs

Keep domain-specific behavior inside resource-local hooks:

- Pre-hooks for REST fallbacks (`switchcontext`, invite create/delete).
- Custom file upload logic (`organization`, `user`, `trustcenternda`).
- Special output formatting (`trustcenter` settings, search results).

Only move helpers into `speccli` when they are genuinely shared (JSON
rendering, validation errors, enum parsing, upload helpers, etc.).

## Testing

- `cmd/cli/internal/speccli/factory_test.go` covers flag binding, input
  marshalling, and render paths.
- Run `go test -tags cli ./cmd/cli/internal/speccli` and
  `go test ./pkg/gencmd/...` with temporary `GOCACHE/GOTMPDIR` directories if
  SIP prevents writes to the default cache.

## CI / Follow-ups

- Add regression/golden tests for high-touch commands.
- Wire spec generation checks into CI to catch drift between specs and GraphQL.
- Update runbooks and release notes once the manual commands are redesigned.

