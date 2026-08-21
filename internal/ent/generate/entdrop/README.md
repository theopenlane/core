# entdrop

Removes ent schemas without hand-editing generated code.

```bash
task drop:schema -- ScheduledJob JobRunner
```

That runs three steps: `entdrop --phase pre`, then `task regenerate`, then
`entdrop --phase post`.

## Why two phases

`entc` type-checks everything `internal/ent/schema` imports before it generates, so
a stale reference in that closure blocks the regeneration that would clean up the
rest. **pre** clears those. **post** handles what only becomes removable once codegen
has dropped the generated types.

Orphans matter for the same reason. ent's `cleanOldNodes` only deletes five per-type
templates under an `entc.Generate` target; graphql fragments, resolvers, csv samples
and cli commands are left behind. An orphaned `graphapi/schema/<entity>.graphql` keeps
declaring the type, so gqlgen rebuilds it into `ent.generated.go`. Delete the orphan
and that file comes out clean on its own.

## What each phase does

**pre** — writes the `DROP TABLE` migration (ent cannot: its Atlas differ filters
changes down to `AddTable`/`ModifyTable`), prunes edges from sibling schemas, deletes
orphans and dedicated files, trims `preTrims`.

**post** — trims `postTrims`, strips the block gqlgen parks removed resolvers in,
deletes files that are only now unreferenced (enums).

**Still manual:** `fga/model/roles/roles.fga` `@crud` lists, `fga/tests/tests.yaml`,
`fga/model/fga.mod`, then `task fga:generate && task fga:test`.

## Report mode

No flags = report only.

```bash
go run ./internal/ent/generate/entdrop ScheduledJob
```

| bucket | action |
|---|---|
| `BLOCKING` | fix first, or codegen can't run |
| `ORPHAN` | delete; they resurrect the type |
| `AUTOCLEAN` | leave; codegen removes them |
| `HANDWRITTEN` | any time before the final build |
| `MIGRATION` | never edit; add a `DROP TABLE` |

## Gotchas

- **Enums come out in post.** `ent/generated/<entity>/` references
  `enums.<Entity>Status` and must compile for entc to load.
- **`historygenerated` is an entc target too**, so its per-type files are
  `AUTOCLEAN`. Deleting them breaks `historygenerated/client.go`.
- **`atlas migrate hash` needs a valid login.** If it fails the migration files are
  still correct; run `task db:resethash` afterwards.
- **Review the migration.** `DROP TABLE ... CASCADE` clears dependent FK constraints
  but leaves now-meaningless columns on surviving tables; the next `task db:create`
  emits those column drops, since ent can do `DropColumn`.

`--trim` is blunt: it cuts any declaration, field, case arm, import, or literal
element mentioning a target. Fine for files that are regenerated or deleted right
after; check the diff when it touches hand-written code.

## Notes

AST byte-range excision then `gofmt`, so diffs are pure deletions with no
reformatting. Only edge-helper calls are pruned; anything else is reported as
`MANUAL` with file:line. Naming uses `stoewer/go-strcase` and `gertd/go-pluralize`;
stem matching is exact, so `ScheduledJob` never claims `ScheduledJobRun`'s files.

Detection is textual, so concatenation-built references are invisible.
