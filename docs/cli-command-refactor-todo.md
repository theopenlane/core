# CLI Command Consolidation TODO

## Phase 0 – Discovery & Scoping
- [ ] Catalogue the generated CLI packages under `cmd/cli/cmd/*` and mark which follow the CRUD pattern vs bespoke flows (`login`, `register`, `search`, etc.) that will stay manual.
- [ ] Capture per-resource deviations: extra filters (e.g. `control` `--ref-code`), enum coercion, multi-id operations, CSV uploads, and bespoke output handling to ensure the spec model can represent them.
- [ ] Map current table output columns and JSON behaviors to understand how much variability the shared renderer must support.
- [ ] Document how `pkg/gencmd` is invoked today (Taskfile targets, flags, build tags) and the expectations contributors have when generating a new command.

## Phase 1 – Spec & Factory Design
- [ ] Define the `CommandSpec` schema covering: resource name, Cobra metadata, supported operations (list/get/create/update/delete), GraphQL operation identifiers, pagination support, and per-operation field metadata (type, required flag, CLI flag alias, description, enum mapping, multivalue).
- [ ] Decide on spec storage (generated Go structs vs. emitted JSON/YAML consumed at runtime) and outline how overrides will be layered for edge cases.
- [ ] Model output rendering requirements: table column definitions (label, field path, formatter), JSON passthrough, and delete confirmation rows.
- [ ] Specify extension hooks for bespoke validation or post-processing so legacy one-off logic can be reattached without forking the factory.
- [ ] Produce an architecture note describing how specs load, how commands register with Cobra, and how openlaneclient method names are resolved (direct call vs reflection/dynamic dispatch).

## Phase 2 – Runtime Implementation
- [x] Create a new package (e.g. `cmd/cli/internal/speccli`) owning spec loading, command construction, and shared helpers while keeping the `//go:build cli` guard.
- [x] Implement a spec loader that merges generated specs with optional handwritten overrides and caches the results.
- [x] Build a command factory that turns a spec into `*cobra.Command`s, wiring shared auth setup, pagination handling, and consistent error surfacing.
- [x] Implement flag binding + validation that derives CLI flags from field metadata (supporting bool/int/string/enums/slices/JSON blobs) and reports missing required inputs using existing error helpers.
- [x] Implement mutation payload marshalling: convert parsed flag values into the generated GraphQL input structs (pointer handling, optional fields, enum conversion).
- [x] Centralize list/get output rendering with reusable helpers that honour spec-defined table schemas and seamlessly fall back to JSON.
- [x] Provide opt-in hooks in the factory for pre-execution request shaping (e.g. building filtered `WhereInput`s) and post-execution formatting needed by outliers.
- [x] Add unit tests for flag binding, input marshalling, and output rendering to prevent regressions as specs expand.

## Phase 3 – Code Generation Updates
- [x] Extend `pkg/gencmd` so the generator emits spec artifacts instead of full Cobra implementations (e.g. `cmd/cli/specs/<resource>.yaml` or Go definitions), keeping backwards compatibility with current CLI tags.
- [x] Enhance the generator to introspect the GraphQL schema (via gqlparser or ent metadata) to derive Create/Update input fields, required flags, enums, and list queries automatically.
- [x] Allow the generator to seed sensible defaults for table columns by reading query selection sets (e.g. `internal/graphapi/query/<resource>.graphql`) while permitting manual override.
- [x] Provide a path for generator-managed overrides (e.g. an adjacent `overrides.go` stub) when automatic metadata is insufficient.
- [x] Update `pkg/gencmd/README.md` and Taskfile targets to explain the new workflow, including how to regenerate specs and where to wire custom logic.

- [x] Implement a temporary compatibility layer so legacy per-resource packages can coexist while new spec-driven commands roll out incrementally.
  - [x] Not needed—CLI runs exclusively on spec-driven commands, so no layer will be maintained.
- [x] Convert one representative resource (e.g. `contact`) to the new spec-driven flow to validate end-to-end behavior, feature parity, and developer ergonomics.
- [x] Convert a read-only/resource variant (e.g. `contacthistory`) to exercise list filtering and non-mutation flows.
- [x] Convert a mutation-heavy resource (e.g. `program`) to validate enum parsing, duration handling, and delete/list outputs.
- [x] Convert a token resource (e.g. `personalaccesstokens`) to exercise custom table formatters and duration parsing.
- [x] Convert a standards/resource (e.g. `standard`) to validate enum parsing, multi-field specs, and generator defaults.
- [x] Convert high-touch commands (`organization`, `user`, `apitokens`, `organizationsetting`, `login`) to spec-driven flow with the necessary hooks for file uploads and interactive authentication.
- [x] Migrate remaining generated resources in manageable batches, deleting obsolete `create/update/delete/get/root.go` files per directory and pruning the imports in `cmd/cli/main.go`.
  - [x] Migrate `group` to spec-driven flow (spec + hooks + register wiring)
  - [x] Migrate `file` to spec-driven flow
  - [x] Migrate `documentdata` to spec-driven flow
  - [x] Migrate `subcontrol` to spec-driven flow
  - [x] Migrate `trustcenter` to spec-driven flow (with settings subcommands retained via helpers)
  - [x] Migrate `trustcenter-domain` to spec-driven flow
  - [x] Migrate `trustcentersubprocessors` to spec-driven flow
  - [x] Migrate `control` to spec-driven flow (preserving role/program editors through hooks)
  - [x] Migrate `groupmembers` and `orgmembers` to spec-driven flow
  - [x] Migrate `programmembers` to spec-driven flow
  - [x] Migrate `template` to spec-driven flow with JSON file parsing helpers
  - [x] Migrate `subscriber` to spec-driven flow (bulk creation + CSV uploads)
  - [x] Migrate `invite` to spec-driven flow while retaining `accept` as an auxiliary subcommand
  - [x] Migrate `trustcenternda` (create/update + submit/send-email extras) to spec-driven flow
- [x] Port shared helpers (`consoleOutput`, enum coercion, validation errors) into the new package and remove duplicate copies from resource folders.
  - [x] Centralize role/invite enum coercion in `speccli` and update membership hooks.
  - [x] Consolidate JSON printing helpers into `speccli` and drop `cmd/output.go`.
  - [x] Re-export validation error helpers from `speccli` so commands avoid bespoke duplicates.
- [ ] Ensure history-only commands (read-only) and other variants (bulk, CSV, multi-id) are either spec-supported or called out for manual retention.
  - [x] Remove generator support for history commands to avoid reintroducing them.
  - [x] Document remaining manual commands (`login`, `register`, `reset`, `invite accept`) as intentional outliers due to bespoke REST flows.
- [x] Update documentation and internal runbooks referencing the old directory structure or hand-edit expectations.

## Phase 5 – QA & Rollout
- [ ] Add regression tests or golden-file snapshots for selected commands using the new factory to guarantee CLI output remains stable.
- [ ] Wire the spec generation into CI (lint/go test) so drift between schema and specs is caught automatically.
- [ ] Validate that `go build -tags cli ./cmd/cli` succeeds after removals and that `task cli:generate` continues to behave for new resources.
- [ ] Communicate the migration plan to the team (changelog, Slack, internal docs) and schedule a deprecation window for legacy command contributions.
- [ ] After successful rollout, remove transitional shims, archive the old templates, and close the loop with post-mortem notes on adoption pain points.
