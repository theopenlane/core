# GenCMD

`pkg/gencmd` is the scaffolding tool that bootstraps CLI resources for the
spec-driven factory (`cmd/cli/internal/speccli`). Instead of emitting fully
wired Cobra commands, the generator now focuses on producing the data and glue
needed by the runtime factory:

- `spec.json` – declarative command definition (operations, flags, columns).
- `register.go` – registers the spec with `speccli` and seeds type metadata.
- `doc.go` – package doc stub to appease `go vet`.

Legacy Cobra templates are still available, but new work should prefer the
spec-driven output.

## Generating a New Resource

The Taskfile wraps the generator; pass `--spec` (or `-s`) to opt into the new
flow:

```bash
task cli:generate -- --spec
```

You can also invoke the generator directly:

```bash
go run pkg/gencmd/generate/main.go generate --spec \
  --name OrganizationSetting \
  --dir cmd/cli/cmd
```

During generation you will be prompted for the resource name (singular). The
tool will:

1. Inspect GraphQL queries under `internal/graphapi/query` to infer sensible
   table column defaults (including nested field paths).
2. Read `pkg/openlaneclient` types to derive create/update flag metadata,
   required/optional status, and enum parsers.
3. Emit `spec.json`, `register.go`, and `doc.go` inside
   `cmd/cli/cmd/<resource>/`.

After generation:

1. Add any bespoke hooks to `<resource>/hooks.go` (create/update/get/delete).
   The generator leaves the hook map empty; see existing resources for
   examples.
2. If the command exposes additional flags or needs custom behaviour, update
   the spec by hand – the factory will merge those changes on the next run.
3. Import the package in `cmd/cli/main.go` if it is not already referenced.

## Read-Only Commands

Use `--read-only` when a resource only supports list/get operations:

```bash
task cli:generate:ro -- --spec
```

This produces the same files, but skips mutation metadata.

## Customising Output

- **Columns:** Edit `spec.json` directly or adjust the underlying GraphQL query
  so the generator picks up the desired fields. Nested selections (e.g.
  `organization { name }`) are honoured automatically.
- **Flags:** The generator reads struct field tags from `pkg/openlaneclient`.
  Add appropriate JSON tags and rerun the generator to refresh `spec.json`.
- **Hooks:** Use the hook factories in `register.go` to bind to helpers. Hooks
  return an `OperationOutput`, enabling custom record/JSON rendering while still
  benefiting from the shared factory wiring.
- **Overrides:** Each generated package now includes an `overrides.go` stub.
  Implement the callback registered there to tweak metadata post-load without
  forking the spec file (e.g. remove columns, adjust defaults).

## Future Improvements

- Generator-managed override stubs for edge cases.
- Taskfile/`go generate` integration for bulk regeneration.
- Test coverage for spec inference and flag binding.

Contributions toward these items are welcome; see
`docs/cli-command-refactor-todo.md` for the broader migration roadmap.
