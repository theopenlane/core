# Entityops Consolidation and Branch Remediation Plan

Status: proposed, implementation not started

Audit date: 2026-08-12

Core branch: `feat-integrationsandgalauplifts` at `8d58adcb18ae`

Comparison: 9 commits ahead of local `origin/main` at `2e21c6666e6a`

Tandem generator: `/Users/manderson/entx`, branch `feat-bindandattach`

## Executive decision

Use one complete `entityops.Schema` catalog as the canonical source for entity identity,
fields, edges, metadata, intrinsic operations, and optional capabilities.

Do not merge `SchemaHandler` and `MutationListener`, and do not put Gala `Topic`, `Wrap`,
or `Unwrap` on `Schema`:

- `MutationListener` is a live outbound mutation subscription.
- `SchemaHandler` was an inbound asynchronous ingest command adapter.
- This branch removed the only production ingest emitter. `SchemaHandler.Register`,
  `SchemaHandler.Emit`, the 13 generated ingest events/topics, and their consumers now form
  an idle transport plane. The correct consolidation is to delete that plane after handling
  queued-job compatibility, not to make it part of the entity descriptor.
- The surviving synchronous ingest preparation and persistence behavior should become a
  type-erased `IngestCapability` reached through the canonical schema. Preparation must be
  inside that operation so a caller cannot accidentally bypass it.

The target is one entity catalog and one ingest entry point, not one type that owns every
transport and application behavior.

## Audit basis

The investigation covered the generated core surface, its tandem `entx/entityops` generator
and templates, and production/test callers in hooks, integrations, workflows, GraphQL, and
Gala.

Key inventory results:

- The generated Ent client exposes 107 entity clients. The entityops registry contains only
  48 schemas.
- There are 38 literal `entityops.MutationListener` registration sites.
- `SchemaHandler.Persist` is used only by synchronous integration ingestion.
- `SchemaHandler.Register` is called during executor setup, but no branch production path
  calls `SchemaHandler.Emit`.
- All 13 generated `Prepare*Input` functions are live through the operations-side
  `ingestHandlers` map.
- Workflows still use separate generated catalogs/switches for eligible fields, metadata,
  loading, owner lookup, field updates, and object-reference mechanics.
- The tandem `entx` worktree was inspected but not modified. Its pre-existing staged and
  unstaged changes must be preserved.

Representative evidence:

- Selective generation: `/Users/manderson/entx/entityops/generator.go:317-353`
- Current schema shape: `internal/ent/entityops/entity_registry.go:80-105`
- Ingest handler abstraction: `internal/ent/entityops/entity_handlers.go:17-134`
- Live synchronous caller: `internal/integrations/operations/ingest_handlers.go:147-156`
- Idle listener registration: `internal/integrations/operations/ingest_handlers.go:127-145`
- Mutation listener resolution: `internal/ent/entityops/entity_listener.go:175-213`
- Mutation snapshot path: `internal/ent/hooks/event_hook.go:118-175`
- Parallel workflow registry: `internal/workflows/fields.go:13-136`
- Workflow-only edge extraction: `internal/ent/entityops/entity_workflow.go:135-173`

Verification performed during this audit:

- production and test caller tracing with `rg` across core and tandem generator sources;
- comparison of the branch ingest path with `origin/main`, confirming that main still had the
  asynchronous producer that this branch removed;
- the focused package suite passed: `go test -timeout 120s ./internal/ent/entityops ./internal/integrations/registry ./internal/integrations/operations ./internal/workflows/...`;
- `internal/ent/entityops` currently reports `[no test files]`, which is why the generator,
  listener, veto, and snapshot regression tests below are required;
- `git diff --no-index --check /dev/null docs/entityops-consolidation-plan.md` reported no
  whitespace errors for this new document.

## Merge blockers found or reconfirmed

### 1. The canonical registry is not canonical

`collectEntityData` includes only schemas with workflow, integration mapping, or task-rule
annotations. GraphQL generation skip rules also influence eligibility. That is valid for
optional capabilities but invalid for base entity identity and mutation loading.

`MutationListener.Schema` accepts an arbitrary string. At dispatch,
`mutationHandler` ignores the result of `LookupSchema(payload.MutationType)` and can pass
`Invocation.Schema == nil` to the handler. It also does not verify that the payload type
matches the schema used to construct the listener topic.

Concrete consequences:

- Trust-center listeners register `TrustCenterDoc`, `Note`, `TrustCenterEntity`,
  `TrustCenterSubprocessor`, `TrustCenterCompliance`, `TrustCenterFAQ`,
  `TrustCenterSetting`, `Standard`, and `TrustCenter`; several are absent from the catalog.
- `trustCenterIDForMutation` dereferences `inv.Schema.Load`, and
  `refreshResolvedTrustCenter` dereferences `inv.Schema.Snake`
  (`listeners_trustcenter_cache.go:123-159`). Normal update/delete events can panic, retry,
  and eventually discard.
- `snapshotMutation` skips old-value capture for uncataloged schemas. Trust-center moves can
  fail to invalidate the old trust center, and Note updates can lose the old rich-text/linkage
  data needed to prevent duplicate work.
- `LegacyTopicRenames` iterates the same incomplete catalog, so queued mutation jobs for
  uncataloged listener schemas are omitted from the migration map.

Required fix:

1. Generate a base descriptor for every primary, eventable Ent schema. Do not reuse
   `SchemaGen`/`QueryGen` skip flags for entityops. Add a dedicated `EntityOpsSkip`
   annotation only for genuine exclusions such as history/internal graph types.
2. Give every base descriptor identity, full field/edge metadata, `LoadRaw`, `LoadManyRaw`,
   and `LoadNode` where the generated type is a `generated.Noder`.
3. Make mutation subscriptions hold a canonical `*Schema`, not a string.
4. Bind `Invocation.Schema` from the registered subscription and reject a mismatched
   `payload.MutationType` as a poison/misrouted envelope.
5. Validate all subscription fields, matches, operations, and required capabilities during
   attachment. Startup must fail on an invalid subscription.
6. Return a transient dependency error when the Ent client cannot be resolved. Do not log and
   acknowledge the event as successful.

### 2. Nested mutations share one mutable emission veto

`WithEmissionVeto` reuses an inherited holder. A nested soft-delete pass vetoes that shared
holder, so the outer, unrelated parent mutation suppresses its own envelope. This reaches
Standard, TrustCenter, JobRunnerRegistrationToken, User, and organization membership flows.

Required fix:

- Split two concepts that the current API conflates:
  - inherited subtree suppression, for workflow/cascade writes that intentionally emit no
    mutation events;
  - a hook-local emission frame, for the outer delete and its rewritten soft-delete pass.
- Install a fresh frame for every mutation hook invocation.
- Reuse a frame only through an explicit soft-delete rewrite token for the same mutation.
  A nested mutation must receive a new frame.
- Replace `WithEmissionVetoed` call sites with an explicitly named inherited API such as
  `SuppressMutationEvents`.
- Add nested delete/create/update, transaction, and soft-delete double-pass regression tests.

### 3. Mutation ID and snapshot failures are silent

`hooks.getMutationIDs` discards `m.IDs(ctx)` errors. `snapshotMutation` also silently skips
schema loads and JSON decode failures. A mutation can commit while producing no envelopes or
incomplete old values.

Required fix:

- Move a minimal, transport-neutral `MutationIDs(ctx, mutation) ([]string, error)` helper to
  entityops and migrate hook callers to explicit error handling.
- Return a snapshot result and error from one entityops constructor. Use `LoadManyRaw` to
  avoid an N-query bulk snapshot.
- Treat ID resolution as required for an interested update/delete. The default should fail
  the mutation before commit; if a future path chooses degraded emission, it must emit a
  typed id-less failure signal and an error metric rather than silently returning.
- Define an explicit policy for per-row load/marshal failures. For delete old values and
  listener routing guarantees, fail closed unless the listener declares old values optional.

### 4. Event identity aliases entity identity

`EmitMutation` passes `payload.EntityID` as Gala `EventID`. Two distinct mutations of the
same entity therefore receive the same event identity. Domain-scan dispatch derives its
unique key from that identity, so a later legitimate scan can be skipped for River's terminal
retention window: approximately 24 hours for completed/cancelled rows and up to 7 days for
discarded rows under the current configuration.

Required fix:

- Allocate a fresh event ID for each logical per-row mutation envelope.
- Reuse that ID only across the concern fan-out of the same logical mutation.
- Keep entity ID as payload/property data, never as event identity.
- Test two updates to the same entity, concern fan-out correlation, and scan dispatch after a
  completed and discarded prior job.

### 5. Synchronous SCIM failures are acknowledged as success

`ProcessPayloadSets` logs record failures, returns nil, and is called directly by SCIM.
That best-effort River policy is not valid for a synchronous IdP push: the caller can receive
2xx while no record was persisted, and there is no next polling cycle to retry it.

Required fix:

```go
type IngestMode uint8

const (
    IngestBestEffort IngestMode = iota
    IngestStrict
)

type IngestResult struct {
    Attempted int
    Succeeded int
    Failed    int
    Failures  []RecordFailure
}
```

- Return `(IngestResult, error)` from the batch processor.
- River polling/reconcile callers may select best-effort behavior, but must record accurate
  failed counts.
- SCIM selects strict mode, executes the write set transactionally where its atomic contract
  requires it, and returns the existing typed errors so 400/409 mappings are live.
- Panics in one record must become a captured record failure or job error, not terminate the
  batch process unpredictably.

### 6. Timeout detachment permits concurrent complete snapshots

River enforces job timeout through context cancellation. Gala's dispatcher detaches from that
context before handler dispatch. A long directory sync can continue after River rescues the
job and starts another attempt. Concurrent complete-snapshot run IDs can overwrite one
another's `LastConfirmedRunID` and falsely remove memberships.

Required fix:

- Preserve the River job context through handler execution. Detach only the short enqueue
  step used by post-commit mutation emission.
- Set an explicit timeout appropriate to the actual operation and honor cancellation before
  finalization/removal inference.
- Add integration+operation single-flight and a fencing token for complete snapshots. Only
  the current, successful run may infer removals.
- Never infer removals after cancellation, partial failure, or a superseding run.

### 7. Snapshot completeness is global and Authentik is incremental

Authentik fetches users with `LastUpdatedGt` and emits memberships only for changed users,
while runtime execution marks the run `CompleteDirectorySnapshot=true`. A later successful
incremental run can therefore mark memberships for unchanged users as removed. Record-failure
skipping makes the risk larger.

Required fix:

- Replace the batch-wide bool with per-contract/per-schema completeness, for example
  `SnapshotCompleteness map[SchemaID]Completeness`.
- Authentik membership output must be marked partial for incremental runs. It may claim a
  complete membership snapshot only after a full-fetch contract actually returns all active
  memberships.
- Authorize removal inference only when membership completeness is full, failed count is zero,
  the run still owns the fencing token, and finalization succeeds.

## Structural findings

### Dead asynchronous ingest control plane

On `origin/main`, `EmitPayloadSets` could call `emitMappedRecord`; this branch removed that
producer and made batch persistence synchronous. The following remains generated/registered
without a producer:

- `SchemaHandler.Register` and `SchemaHandler.Emit`
- `SchemaHandlerConfig.Topic`, `Wrap`, and `Unwrap`
- `BuildSchemaHandler` and 13 per-schema handler factories
- 13 `*IngestRequested` payload types
- 13 `Topic*` values and the `IngestTopics` namespace
- `RegisterIngestListeners` and `ingestSchemaOrder`
- async fallback in `resolveIngestIntegration`
- `buildIngestOperationContext` and its provenance-only option fields
- `gala.JobKindIntegrationIngest`, subject to legacy queue compatibility

Deletion gate: a deployment based on `origin/main` may still have queued per-record ingest
jobs. Before deleting registrations/job-kind routing, inspect the production queue. Either
drain/migrate those rows or retain terminal tombstone consumers for at least the maximum River
retention/attempt window. If this code has never been deployed, record that fact and delete
directly.

### Preparation is live but bypassable

The 13 generated `Prepare*Input` functions stamp owner, integration, and platform defaults.
Directory wrappers additionally stamp `DirectorySyncRunID`. They are all live, but only because
the operations package maintains a second map that correctly pairs prepare and persist
functions. A new schema can be added to one registry and omitted from the other.

The current ordering is also wrong for source-aware links:

1. map raw payload;
2. resolve/inject links using the unprepared payload as source context;
3. prepare integration defaults inside persistence;
4. persist.

Link CEL/key matching cannot see implicit owner/integration/platform/run values. The canonical
order should be:

1. map raw payload;
2. resolve canonical schema and compiled mapping once;
3. apply schema ingest preparation;
4. resolve/inject links using the prepared payload;
5. persist/upsert;
6. apply through-edge rows exactly once in the same transaction;
7. return a structured result.

Do not put integration preparation in general `Schema.Create`. Non-integration creates must not
inherit integration authority or defaults.

### Integration definitions are only partially compiled

`populateMappingLinkTargets` and `validateMappingLinks` continue when `LookupSchema` fails.
Runtime ingestion then discovers the unsupported schema after startup. Each batch also repeats
schema lookup, contract scans, and a linear mapping lookup.

At integration registration:

- resolve every contract and mapping to canonical `*Schema`;
- reject unknown schemas;
- require a bound ingest capability;
- validate mapped fields, variants, edges, and target schemas;
- reject duplicate `(schema, variant)` mappings;
- precompute contract and mapping indexes;
- retain the compiled canonical schema pointer in the definition entry.

### Workflow still has parallel sources of truth

Entityops already has `Schema.Fields`, `Schema.Edges`, `Schema.Load`, `Schema.Update`, and
workflow eligibility flags. Workflows still maintain:

- the standalone generated eligible-field registry and `RegisterEligibleFields`;
- `generated.GetWorkflowMetadata`;
- `generated.LoadWorkflowObject` plus 18 per-type loaders;
- `generated.GetObjectOwnerID`;
- `generated.ApplyObjectFieldUpdates`;
- object-instance and object-ref setter/predicate switches;
- 18 declaration-only workflow instance helpers;
- 51 declaration-only query helpers;
- 18 generated webhook enrichment functions, most of which are per-type wrappers.

These sources have already diverged. The older workflow catalogs expose all 17
`workflow_eligible_marker` fields as real editable workflow fields. Entityops correctly stores
the marker with `WorkflowEligible=false`.

Required consolidation:

- derive workflow metadata and domain validation from `Schema.WorkflowFields()` and eligible
  edge descriptors;
- remove the mutable eligible-fields registry and the standalone workflowgen output;
- add `LoadNode` and a narrow `WorkflowCapability` to canonical schemas;
- put owner lookup and eligibility-checked updates behind that capability;
- preserve typed builder closures only where Go's concrete builder types make a generic edge
  setter unsafe;
- derive WorkflowInstance/WorkflowObjectRef predicates and setters from edge descriptors where
  possible;
- keep notification target resolution, webhook enrichment policy, and workflow service logic
  workflow-owned unless they are truly schema mechanics.

Before replacing typed workflow updates with `Schema.Update`, prove equivalent coercion for
enums, custom types, timestamps, explicit null/clear, and eager-loaded object requirements. If
equivalence is not possible, store the generated typed setter behind
`Schema.Workflow.ApplyFields`.

### ChangeSet is generic in name but workflow-filtered in behavior

`ChangeSetFromMutation` claims to capture catalog-known edges, but it calls
`ExtractChangedEdges`, which includes only workflow-eligible edges. Non-workflow mutation
consumers therefore receive an incomplete generic delta.

Workflows also duplicate field normalization and proposed-change construction. The event hook
builds the `ChangeSet` once during snapshot interest work and again during emission.
`ProposedMap` decodes the same JSON for every accessor/gate.

Required consolidation:

- make one error-returning entityops snapshot builder capture all changed/cleared fields and all
  catalog-known edge add/remove/clear deltas;
- build it once per event-hook invocation and thread it through snapshot and emission;
- make JSON marshal/decode failure explicit instead of returning an empty delta;
- add `ChangeSet.SelectFields`, `Schema.PartitionFields`, or a `WorkflowChangeSet` view so
  workflow filtering happens at the consumer boundary;
- replace workflow `CollectChangedFields`, `CollectAllChangedFields`,
  `BuildProposedChanges`, and their normalization helper with the entityops snapshot/view;
- decode proposed values once into a non-wire `ChangeSetView` used by gates and headers;
- preserve existing queued `MutationPayload` and workflow context JSON tags.

## Target model

The exact names can change during implementation, but the ownership boundaries should not.

```go
type Schema struct {
    SchemaDescriptor

    Operations EntityOperations
    Fields     []FieldDescriptor
    Edges      []EdgeDescriptor
    TaskRules  []TaskRuleDescriptor

    ProjectionType reflect.Type
    Ingest         *IngestCapability
    Workflow       *WorkflowCapability
}

type EntityOperations struct {
    LoadRaw    func(context.Context, *generated.Client, string) (json.RawMessage, error)
    LoadManyRaw func(context.Context, *generated.Client, []string) (map[string]json.RawMessage, error)
    LoadNode   func(context.Context, *generated.Client, string) (generated.Noder, error)

    CreateRaw  func(context.Context, *generated.Client, json.RawMessage) (string, error)
    UpdateRaw  func(context.Context, *generated.Client, string, json.RawMessage) error
    QueryRaw   func(context.Context, *generated.Client, string) ([]json.RawMessage, error)
    QueryByKeyRaw func(context.Context, *generated.Client, string, string, []string) ([]json.RawMessage, error)
}

type IngestScope struct {
    Integration        *generated.Integration
    DirectorySyncRunID string
}

type IngestRequest struct {
    Scope   IngestScope
    Payload json.RawMessage
    Links   []LinkSpec
}

type IngestCapability struct {
    // write is assembled and frozen at startup. Callers use Schema.WriteIngest,
    // which enforces prepare -> links -> persist -> through-edge ordering.
    write func(context.Context, *generated.Client, IngestRequest) (string, error)
}

type WorkflowCapability struct {
    OwnerID     func(context.Context, *generated.Client, string) (string, error)
    ApplyFields func(context.Context, *generated.Client, string, map[string]any) error
    Enrich      func(context.Context, *generated.Client, string) (map[string]any, error)
}
```

Catalog invariants:

1. Every supported primary schema has exactly one canonical `*Schema`.
2. Base identity, fields, edges, and load operations are complete.
3. Optional behavior is represented by a nil capability group, not by excluding the schema.
4. The catalog is assembled, validated, and frozen before listeners/registries/workers start.
5. Callers pass canonical handles after startup compilation; string lookup is an input-boundary
   operation only.
6. A capability method returns a typed unsupported error; callers do not infer support from a
   scattered set of nil closures.

Binding custom ingest persistence requires an application-owned startup step because entityops
cannot import the operations package without a cycle. Use a builder/freeze API that binds a
custom persistence closure to the canonical schema. The final capability lives on the schema;
there is no parallel runtime map.

`StockPersist` must either select a real generated `Schema.Upsert` binding or be deleted. It is
currently copied by the generator but not consumed by templates.

## Surface disposition

| Surface | Decision | Reason |
| --- | --- | --- |
| `Schema`, `SchemaDescriptor` | Keep and reshape | Canonical identity/capability root |
| `Fields`, `Edges`, `TaskRules`, projections | Keep | Live shared metadata; generate for all base schemas |
| `Create/Update/Query/QueryByKey/Load` | Keep, group under operations | Live intrinsic mechanics; add `LoadManyRaw`/`LoadNode` |
| `AllSchemas` | Replace with catalog iteration | Validation needs enumeration, but not a mutable package global copy |
| `AllowedKey` | Delete | No caller; `ResolveInputKey` is the canonical resolver |
| `DisplayField` | Unexport | Only `DisplayValue` needs it outside implementation |
| `SchemaHandler`, `SchemaHandlerConfig`, factories | Delete after queue gate | Async transport is idle; persistence moves to ingest capability |
| Generated ingest events/topics and `IngestTopics` | Delete after queue gate | No branch producer |
| Exported `Prepare*Input` functions | Fold into schema ingest capability | Live behavior becomes mandatory, not caller convention |
| `MutationListener` | Keep, schema-type, optionally rename `MutationSubscription` | Live outbound subscription, distinct from ingest |
| `MutationListener.LogFields`, `.Cancel` | Delete | No registration uses them; raw Gala definitions remain available for custom behavior |
| `LoadEntity` | Keep | Live standard not-found/retry helper |
| `ChangeSet` and value accessors | Keep and harden | Durable shared delta contract |
| `OldValueSource`, `BuildOldValues` | Delete | No live caller; database snapshot path supersedes them |
| `EdgeChange`, `ExtractChangedEdges` | Make private/generalize | Implementation detail of canonical ChangeSet |
| `WorkflowFieldNames` | Delete | Names-only duplicate of descriptor view |
| `WorkflowFields` and eligible edge descriptor view | Keep | Canonical workflow metadata/filter input |
| `LinkTargets`, `UnlinkTargets` | Delete unless reconciliation gets an immediate consumer | No callers |
| `InjectCreateLinks` | Keep behind ingest pipeline | Live generic entity mechanic |
| `MutationHeaders`, `MutationConcernTopics` | Unexport | Internal mutation transport helpers |
| Current veto APIs | Replace | Separate inherited suppression from per-mutation frames |
| Console routes, mention specs, target selection | Keep | Live shared metadata/selection behavior |
| `ValueAsString`, `ParseEnum` | Keep | Live mutation listener helpers |
| workflowgen eligible-field/domain output | Delete after migration | Parallel divergent catalog |
| `workflow_instance_helpers.go` | Delete | 18 declaration-only functions |
| `workflow_query_helpers.go` | Delete | 51 declaration-only functions |
| per-type workflow loaders | Replace with `Schema.Operations.LoadNode` | Duplicate generated switch |
| workflow metadata/owner/update switches | Replace with schema workflow capability | Duplicate source of truth |

Tandem generator cleanup after the transport deletion:

- remove `EntitySchema.IngestTopic`, `IngestRequestType`, and `IngestTopicVar`;
- remove the copied `EntityField.FromIntegration` after runtime defaults are compiled directly;
- activate or remove `EntitySchema.StockPersist`;
- stop generating `entity_handlers.go` and the event/topic half of
  `entity_integration.go`;
- keep input-key constants and integration annotation parsing that feed the canonical field
  descriptors and prepare capability.

## Remaining confirmed defect plan

These fixes are not all entityops-owned, but they must be sequenced with the consolidation.

| Finding | Disposition and proposed update |
| --- | --- |
| Same-kind job rename migration | Valid latent defect. Prefer a transactional in-place args/topic rewrite. If reinsertion is required, use a migration identity that cannot dedupe against the source row, lock/limit eligible states, insert first, verify, then cancel and propagate cancellation failure. |
| Schedule-loop rescue fork | Valid when the successor has reached a terminal state before rescue; definite for short domain-poll cycles and conditional for long reconcile intervals. Persist a cycle token and use terminal-retaining (`UniqueOnce`) keys. A failed successor enqueue must leave the predecessor retryable; do not cancel it. |
| Schedule predecessor cancellation | Additional defect. Current execution can cancel the predecessor when both execution and successor emission fail. Couple cancellation to verified successor creation and return cancel errors. |
| Migration pickup race | Additional defect. List/dispatch/cancel is not pickup-race-free. Restrict to non-running states under lock/transaction or update in place. |
| Trust-center transient errors | Return database/enqueue/network transients so River retries. Skip only classified permanent/not-found conditions. Do not bury them under a successful listener result. |
| Missing Ent client in listener | Return `ErrClientResolveFailed`; do not acknowledge success. |
| WorkflowAssignmentCreated has no registration | Pre-existing on `origin/main`, not a branch regression. Add the intended handler/definition or remove the emitter after an ownership decision; do not leave a permanently failing event. |
| Raw workflow `do.Invoke` sites | Route through the shared workflow mutation forwarder/preamble so nil runtime and skip logging semantics are consistent. |
| Inline privacy elevation | Audit each site against interceptor behavior. Replace only after proving the declared caller/capability is sufficient; do not blindly delete the `DecisionContext` calls. |
| Listener labels | Reconcile the design record with code. Require labels only for same-topic collisions, or enforce labels universally in registration validation. Update the design record to the selected invariant. |
| Stale Gala README examples | Regenerate examples against real `MutationPayload`, runtime emit APIs, and header options; compile examples in tests. |
| `payloadOperation` reflection | Replace with an optional typed `interface { PayloadOperation() string }` contract plus the generic fallback. |
| Unused `ConfigureGala` context | Remove the parameter or use it for actual startup validation; do not retain a ceremonial API argument. |
| ChangeSet built twice and decoded repeatedly | Build once, return errors, use a decoded `ChangeSetView`, and filter consumer-specific views. |
| Double topic registry lookup | Return registration+kind in one lookup. Cache startup-fixed listener interest where measurements justify it. |
| Dead workflow-timeout topic family | Delete unless an owner documents and implements the producer/consumer contract now. |
| Mutation listener cancel hatch | Delete with the other unused listener fields; Gala definitions still support custom cancellation. |
| Ingest provenance plumbing | Delete async-only `RunID/Webhook/Event/DeliveryID/WorkflowMeta` propagation unless it is wired directly into live run accounting/observability. |
| Shared helper candidates | Extract after correctness phases: internal-operation caller, reconcile metadata/filter helpers, River page iteration, retry policy, and test registration helpers. |
| HTTP retry response-body leak claim | Not a defect with pinned `httpsling v0.3.0`; it consumes/closes the network body and returns an in-memory body. Retry-loop consolidation remains optional refactoring. |

Other report items such as notification narrow selects, registry lookup reduction, naming/godoc
cleanup, and test scaffolding are valid low-risk follow-ups. They should not be mixed into the
merge-blocker commits.

## Implementation sequence

### Phase 0: Lock behavior with failing tests

- Reproduce uncataloged trust-center/Note listener handling.
- Reproduce nested soft-delete veto leakage.
- Reproduce mutation ID query failure.
- Reproduce same-entity event ID/scan collision.
- Reproduce SCIM strict failure acknowledgment.
- Reproduce Authentik incremental false removal and failed-record removal inference.
- Reproduce schedule successor failure and same-kind migration collision.
- Record the legacy ingest queue compatibility decision.

### Phase 1: Complete and validate the catalog

- Change tandem generator collection to produce base descriptors for all supported primary Ent
  schemas and optional capability flags separately.
- Add `LoadRaw`, `LoadManyRaw`, and `LoadNode` generation.
- Convert mutation subscriptions to canonical schema handles.
- Fail attachment on invalid schema, operation, field, match, or missing capability.
- Bind invocation schema from the subscription and reject payload mismatch.
- Make integration definition registration reject/compile schema references and mapping indexes.
- Regenerate core and prove byte-clean core/entx parity.

### Phase 2: Repair mutation emission

- Introduce frame-scoped emission control plus inherited explicit suppression.
- Centralize error-returning mutation ID/snapshot construction.
- Capture one complete ChangeSet and bulk old-value snapshot.
- Generate fresh logical event IDs and correlate concern fan-out.
- Preserve post-commit-only dispatch and add transaction/nesting tests.

### Phase 3: Replace the ingest handler plane

- Add `IngestScope`, the schema ingest binding builder, validation, and catalog freeze.
- Move generated defaults behind the capability.
- Bind generic stock upsert or schema-specific persistence once at startup.
- Make prepare -> link -> persist -> through-edge ordering transactional and exactly once.
- Remove the operations `ingestHandlers` map and context-carried integration fallback.
- Return structured strict/best-effort results.
- Add per-schema completeness, fencing, cancellation, and correct removal inference.
- Drain/migrate/tombstone legacy async jobs, then delete events/topics/listeners/job kind.

### Phase 4: Consolidate workflows

- Capture all edge deltas in ChangeSet and filter through schema capability views.
- Replace workflow changed/proposed field helpers with entityops views.
- Move load-node, owner, and eligibility-checked update mechanics behind schema workflow
  capability.
- Replace GraphQL metadata and workflow domain generation with canonical descriptors.
- Remove workflowgen and declaration-only generated helper families.
- Verify marker exclusion and typed update behavior.

### Phase 5: Repair Gala scheduling and migration

- Preserve worker cancellation and configure real operation timeouts.
- Add complete-snapshot single-flight/fencing.
- Make cycle identity survive rescue and terminal rows.
- Make successor creation/cancellation failure-atomic.
- Make topic migration same-kind-safe and pickup-race-safe.

### Phase 6: Prune and document

- Delete the dead exports and generator data fields in the disposition table.
- Consolidate repeated helpers only after their semantics match.
- Repair README/examples/godocs and add compile checks.
- Re-run unused/dead-topic searches after regeneration.

## Verification gates

Generator and catalog:

- golden tests for the tandem generator;
- a clean regeneration with no unexplained core diff;
- every registered mutation listener resolves to exactly one schema with required operations;
- every mapping/contract resolves at startup;
- every workflow object type maps exactly once;
- `workflow_eligible_marker` is absent from public workflow metadata and update eligibility.

Mutation events:

- field set/clear and edge add/remove/clear coverage;
- bulk update/delete ID query error coverage;
- hard-delete old values and bulk-load behavior;
- nested mutation/soft-delete/transaction coverage;
- unique event identity with shared concern correlation;
- malformed schema/payload mismatch is rejected, not delivered with a nil/wrong schema.

Ingest:

- preparation default/no-override behavior for all 13 mapped schemas;
- directory run ID handling without hidden context dependence;
- prepared fields visible to source-aware link resolution;
- through edges applied once and transactionally;
- stock and special persistence parity;
- strict SCIM rollback/error mapping and best-effort River result accounting;
- no removal after partial, failed, cancelled, incremental, or superseded runs;
- complete Authentik/full-fetch membership reconciliation;
- no remaining async ingest producer or listener registration after the compatibility gate.

Workflows:

- metadata/domain parity except intentional marker removal;
- enum, custom type, time, null/clear, and ineligible-field update tests;
- eager-load behavior for workflow object resolution;
- object-ref/instance predicate and setter coverage for every workflow schema;
- approval and post-commit paths observe equivalent filtered ChangeSets.

Gala/River:

- same-kind and different-kind rename migration;
- cancel failure and pickup race tests;
- rescued schedule after a terminal successor does not fork the chain;
- successor enqueue failure does not terminate the only live loop;
- timeout cancellation stops downstream work and prevents stale finalization.

## Definition of done

The consolidation is complete only when:

1. Core and tandem entx generation are reproducible and byte-clean.
2. Every eventable primary Ent type has a canonical schema descriptor.
3. No listener or mapping can start with an unresolved string schema.
4. There is one live synchronous ingest entry point and preparation cannot be bypassed.
5. Workflow metadata, eligible fields/edges, loading, and updates derive from entityops.
6. Mutation deltas are constructed once per path, errors are loud, and all catalog-known edges
   are represented before consumer filtering.
7. SCIM, polling, incremental, and complete-snapshot semantics are explicit and tested.
8. Nested emission, job identity, scheduling, timeout, and migration regressions are closed.
9. Legacy queue compatibility is handled explicitly before dead topic/job-kind deletion.
10. The dead surfaces listed above have no generated copies or stale documentation remaining.
