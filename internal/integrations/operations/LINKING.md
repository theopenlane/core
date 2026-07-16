# Ingest record linking

Cross-object linking sets edges from an ingested record to existing records, **at create time**, as part of the ingest operation. Example: an ingested `Finding` is created already linked to existing `Control`s whose `ref_code` matches the finding's `category`/`categories`.

Linking configuration lives in `internal/integrations`, supplied as input the same way `FilterExpr`/`MapExpr` are — a definition ships defaults and an installation may override them. There are no ingest-link annotations on the ent schema, and no global registry: every default is pushed down into the definition that ingests the schema.

The data model only supplies the **capability catalog** (`entityops`): which object types a schema can be linked to, the create-input key per edge, the match-able fields, and the resolution engine (`SelectTargets`).

## How a link is configured

A mapping declares `Links []types.LinkRule`. Each rule names the target object type and how to match candidates — either a **field match** (indexed query push-down) or a **CEL expression**:

```go
// internal/integrations/definitions/awssecurityhub/mappings.go
Spec: types.MappingOverride{
    MapExpr: mapExprFinding,
    Links: []types.LinkRule{
        // field match: Control.ref_code IN (finding.category, finding.categories[*])
        {TargetSchema: entityops.SchemaControl.Name, TargetField: "ref_code", SourceField: "category", SourceList: "categories"},
    },
},
```

```go
// CEL expression: non-equality / multi-field conditions. "target" is the candidate, "source" is the ingested record.
{TargetSchema: entityops.SchemaAsset.Name, Expression: `target.name == source.resource_name || source.resource_name.startsWith(target.name + "/")`},
```

`LinkRule`:

| field | meaning |
|---|---|
| `TargetSchema` | entityops object type to link to (e.g. `"Control"`); resolves to the source schema's edge to that type |
| `TargetField` | target field for a field match (e.g. `"ref_code"`) — must be an indexed string field |
| `SourceField` / `SourceList` | source scalar / list field whose value(s) the target field must equal |
| `Expression` | CEL match used instead of a field match |

A rule uses a field match **or** an expression. Field match is the fast path (compiled to `WHERE owner_id = <org> AND <TargetField> IN (<source values>)` via the generated `QueryByKey`); reach for an expression when the match isn't a single-field `IN`.

### Per-installation override

The definition's `Links` are defaults. An installation overrides them per operation through the operation's `UserInput` config (`links`), resolved by `resolveInstallationLinkRules` exactly like `FilterExpr`. Effective rules = installation override **or** definition default.

## How it works

Resolve-then-create: targets are resolved from the mapped record and their ids are written into the **create input**, so the record is created with its edges already set rather than linked in a second step. This is atomic and works for immutable edges for free (the create builder has their setter; the update builder does not).

1. Ingest maps the provider payload to the source create input (`ingest.go`). Before persisting, `applyPayloadSets` resolves the effective rules and calls `injectLinks` (`ingest_link.go`).
2. Per rule, `injectLinks` finds the source schema's edge whose target is `rule.TargetSchema` (in the `entityops` `Edges` catalog), builds a `TargetSelector` (key match when `TargetField` is set, else the expression), and calls `entityops.SelectTargets`.
3. Matched ids are written into the payload under the edge's create-input key — `edge.CreateField` (`controlIDs`, or `<edge>ID` for a unique edge) — via `jsonx.SetObjectKey`.
4. The augmented payload is persisted (sync) or emitted for the consumer to persist (async); `SetInput` on the create builder sets the edges.

Injection happens once in the shared `applyPayloadSets`, so both paths behave identically:
- **Sync** (`ProcessPayloadSets`): persists the augmented payload directly.
- **Async** (`EmitPayloadSets` → `handleIngestRequested`): emits the augmented payload; the consumer just persists it, no separate link step.

**Source context.** A rule's `source.<field>` and `SourceField`/`SourceList` read the create-input payload re-keyed into the schema's snake_case field names by the generated `Schema.SourceContext` (the same `normalizeFieldKeys` machinery the update path uses), so source criteria are snake_case, consistent with `target.<field>`. For a field match, the scalar source field plus every element of the source list are deduped and empties dropped — no source values ⇒ no link.

## UI surfacing

The integration-definition REST endpoint (`ListIntegrationProviders` → `[]types.Definition`) carries everything the config UI needs, so it queries one endpoint and gets the extra data:
- **Defaults**: each mapping's `Spec.Links`.
- **Inventory**: `MappingRegistration.LinkTargets`, populated once at registration (`populateMappingLinkTargets`) from the `entityops` catalog — per source schema, the linkable `targetType`s and the match-able `targetFields`/`sourceFields` (the `MatchKey` fields). The UI renders the target dropdown and field pickers from this and pre-fills the rows with the defaults.

## Requirements / validation

- `TargetSchema` must be a real edge target on the source schema — `validateMappingLinks` fails registration on a typo.
- A field-match `TargetField` must be a plain-string indexed field on the target (those are the `MatchKey` fields in the catalog; custom Go types and enums are excluded from `QueryByKey`).
- The target schema must be entityops-eligible (carries `entx.IntegrationMappingSchema()` or a workflow-eligible annotation) and `owner_id`-scoped.

## Relationship to workflows

The workflow `CREATE_OBJECT` action and integration ingest now share the same create-time linking path: both resolve targets through the one `entityops` catalog + `SelectTargets` engine and apply them via `entityops.InjectCreateLinks`, which writes the matched ids into the create input under each edge's `CreateField` so the object is created with its edges already set in a single mutation. This is atomic (no orphaned object on a link failure) and works for immutable edges, which can only be set at create time. The two surfaces differ only in config shape — ingest authors `types.LinkRule` (target-type driven, source = the mapped payload), the workflow authors `entityops.LinkSpec{Edge, Target}` (edge-name driven, with `ExcludeIDs` templated from the triggering object) — but both lower to the same `[]LinkSpec` and the same injection primitive, so a workflow definition can set edges identically to an ingest mapping.
