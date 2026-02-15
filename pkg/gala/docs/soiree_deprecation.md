# Workflow Engine: Soiree to Gala Refactor Plan

This plan details what to replace related to existing soiree usage with gala across the workflow engine, hook infrastructure, and all call sites. Although the packages execute work in very different fashions, the overall primitives are quite similar (given gala was designed to easily take the existing shape of soiree listeners to reduce complexity and risk) so the refactor should hopefully be straightforward.

## Define Workflow Command Topics in `pkg/gala`

Create `pkg/gala/topics_workflow.go` with typed `Topic[T]` definitions and JSON codecs for each workflow command:

- `workflow.command.trigger` -- `WorkflowTriggeredPayload` (move from `soiree/workflow.go`, fields: InstanceID, DefinitionID, ObjectID, ObjectType, TriggerEventType, TriggerChangedFields)
- `workflow.command.advance` -- `WorkflowActionStartedPayload` (InstanceID, ActionIndex, ActionType, ObjectID, ObjectType)
- `workflow.command.action_completed` -- `WorkflowActionCompletedPayload` (InstanceID, ActionIndex, ActionType, ObjectID, ObjectType, Success, Skipped, ErrorMessage)
- `workflow.command.assignment_created` -- `WorkflowAssignmentCreatedPayload` (AssignmentID, InstanceID, TargetType, TargetIDs, ObjectID, ObjectType)
- `workflow.command.assignment_completed` -- `WorkflowAssignmentCompletedPayload` (AssignmentID, InstanceID, Status, CompletedBy, ObjectID, ObjectType)
- `workflow.command.instance_completed` -- `WorkflowInstanceCompletedPayload` (InstanceID, State, ObjectID, ObjectType)
- `workflow.command.timeout_expire` -- `WorkflowTimeoutExpiredPayload` (InstanceID, AssignmentID, ObjectID, ObjectType)
- `integration.command.execute` -- `IntegrationOperationRequestedPayload` (move from `internal/integrations/events.go`)

Each topic gets a `gala.Registration[T]` with `gala.JSONCodec[T]{}` and explicit `gala.TopicName`. Delete `soiree/workflow.go` entirely after migration.

## Rewrite `internal/workflows/engine` Emit Layer

### `engine/eventhandlers.go`

- Change `WorkflowListeners` struct: replace `emitter soiree.Emitter` field with `gala *gala.Gala`
- Change `NewWorkflowListeners` signature: accept `*gala.Gala` instead of `soiree.Emitter`
- All handler methods (`HandleWorkflowTriggered`, `HandleActionStarted`, `HandleActionCompleted`, `HandleAssignmentCompleted`, `HandleInstanceCompleted`, `HandleAssignmentCreated`): change signature from `(ctx *soiree.EventContext, payload soiree.XxxPayload)` to `(handlerCtx gala.HandlerContext, payload gala.XxxPayload)` -- extract `ctx` from `handlerCtx.Context`, resolve `*generated.Client` and `*WorkflowEngine` from `handlerCtx.Injector` via `do.Invoke`
- `HandleWorkflowMutation` and `HandleWorkflowAssignmentMutation`: change from `(*soiree.EventContext, *events.MutationPayload)` to `(gala.HandlerContext, eventqueue.MutationGalaPayload)` -- consolidate with the existing `listeners_workflow_gala.go` handlers (they are the target pattern)
- Remove `mutationSchemaType` soiree helper (replaced by gala envelope topic + payload.MutationType)

### `engine/emit.go`

- Replace all `soiree.WorkflowXxxPayload` references with the new gala topic payload types
- Replace `workflows.EmitWorkflowEvent(ctx, engine.emitter, topic, payload, client)` calls with `galaApp.EmitWithHeaders(ctx, topicName, payload, headers)` returning `gala.EmitReceipt`
- `emitEngineEvent` / `emitListenerEvent` generics: change type constraint from `soiree.TypedTopic[T]` to `gala.TopicName` + payload, use `gala.EmitReceipt` for error tracking
- Update `recordEmitFailure` to use gala receipt types (same shape, different source type)

### `engine/engine.go`

- `WorkflowEngine` struct: replace `emitter soiree.Emitter` with `gala *gala.Gala`, remove `integrationEmitter *soiree.EventBus`
- `NewWorkflowEngine` / `NewWorkflowEngineWithConfig`: accept `*gala.Gala` instead of `soiree.Emitter`
- Integration operations: emit via gala using `integration.command.execute` topic instead of separate `soiree.EventBus`

## Rewrite `internal/workflows/emit.go`

- Replace `EmitWorkflowEvent[T]` generic: change from soiree wrapping (`topic.Wrap(payload)` + `emitter.Emit()` + `drainEmitErrors()`) to `galaApp.EmitWithHeaders(ctx, topicName, payload, headers)` returning `gala.EmitReceipt`
- Replace `EmitWorkflowEventWithEvent` (pre-built soiree.Event) with direct gala envelope emission via `galaApp.EmitEnvelope()`
- Remove `drainEmitErrors()` entirely (gala returns synchronous `EmitReceipt`)
- Keep `EmitReceipt`, `EmitFailureMeta`, `EmitFailureDetails` types (used for reconciler) but update the underlying implementation to use gala types

## Rewrite `internal/workflows/observability/emit.go`

- Replace all soiree references: `emitTyped[T]`, `emitTypedFromScope[T]`, `EmitFromTopic[T]`, `EmitEngine[T]` -- these all use `soiree.Emitter`, `soiree.TypedTopic[T]`, `soiree.Event`
- Change to accept `*gala.Gala` and `gala.TopicName`, emit via `galaApp.EmitWithHeaders()`
- `BeginListenerTopic` currently takes `*soiree.EventContext` for scope initialization -- change to accept `gala.HandlerContext` and extract context, observer, and operation from it
- The `Scope` type itself stays; only the constructors and emit helpers change

## Rewrite `internal/workflows/reconciler/reconciler.go`

- `Reconciler` struct: replace `emitter soiree.Emitter` with `gala *gala.Gala`
- `reemit()`: currently builds `soiree.NewBaseEvent()` and calls `EmitWorkflowEventWithEvent` -- replace with building a `gala.Envelope` from stored `EmitFailureDetails` and calling `galaApp.EmitEnvelope()`
- Remove soiree import entirely

## Delete Legacy Soiree Hook Infrastructure

### Delete `internal/ent/hooks/eventer.go`

The entire `Eventer` struct, `EventerOpts`, `NewEventer`, `Initialize`, `NewEventerPool`, `MutationHandler`, `AddMutationListener`, `AddListenerBinding`, `RegisterListeners`, `registerDefaultMutationListeners` -- all soiree-backed. Gala handles listener registration via its `Registry` and the `EmitGalaEventHook` pattern already in place.

### Delete `internal/ent/hooks/event.go`

The `EmitEventHook` function that builds `soiree.NewBaseEvent()` and emits through `Eventer.Emitter` -- replaced by `EmitGalaEventHook` in `event_gala_hook.go`. Keep the `EventID` parsing helpers (`parseEventID`, `parseSoftDeleteEventID`, `getOperation`), `mutationChangedAndClearedFields`, `mutationProposedChanges`, and `uniqueStrings` as they are reused by the gala hook. Move these into `event_gala_hook.go` or a shared file.

### Delete `internal/ent/hooks/listeners_workflow.go`

The soiree-based `RegisterWorkflowListeners` and `bindWorkflowListener` helper -- replaced by `RegisterGalaWorkflowListeners` in `listeners_workflow_gala.go`.

### Delete `internal/ent/hooks/listeners_entitlements.go`

The soiree-based entitlement mutation handlers -- replaced by `RegisterGalaEntitlementListeners` in `listeners_entitlements_gala.go`.

### Delete `internal/ent/hooks/listeners_slack.go`

The soiree-based Slack notification handlers -- replaced by `RegisterGalaSlackListeners` in `listeners_slack_gala.go`.

### Rename gala listener files

- `listeners_workflow_gala.go` -> `listeners_workflow.go`
- `listeners_entitlements_gala.go` -> `listeners_entitlements.go`
- `listeners_slack_gala.go` -> `listeners_slack.go`
- `event_gala.go` -> `event_emit.go`
- `event_gala_hook.go` -> `event_hook.go`

## Update `internal/ent/hooks/listeners_workflow_gala.go` (Future `listeners_workflow.go`)

### Register Workflow Command Listeners

Currently only registers mutation listeners (`handleWorkflowMutationGala`, `handleWorkflowAssignmentMutationGala`). Add registration for all workflow command topics:

```
gala.Definition[WorkflowTriggeredPayload]{
    Topic: workflow.command.trigger,
    Name:  "workflows.triggered",
    Handle: handleWorkflowTriggered,
}
```

Same pattern for `workflow.command.advance`, `workflow.command.action_completed`, `workflow.command.assignment_created`, `workflow.command.assignment_completed`, `workflow.command.instance_completed`.

Each handler resolves `*generated.Client` and `*engine.WorkflowEngine` from `handlerCtx.Injector` via `do.Invoke` (pattern already established in the existing gala listeners).

### Remove Duplicate Code

The existing `workflowEventTypeFromEntOperation` in `listeners_workflow_gala.go` duplicates `engine/eventhandlers.go`. Consolidate into one location (keep in engine, remove from hooks).

## Rewrite `internal/httpserve/serveropts/` Wiring

### `serveropts/events.go`

- Delete `WithEventEmitter` (soiree-based). The gala hook is already wired via `serveropts/gala.go`.

### `serveropts/gala.go`

- Remove the conditional workflow-enabled check around `RegisterGalaWorkflowListeners` -- workflows are always gala-backed now
- Register workflow command topic codecs (`gala.RegisterTopic`) before listener registration
- Register the `integration.command.execute` topic and listener for integration operations
- Remove the `galaCfg.FailOnEnqueueError` parameter from `EmitGalaEventHook` (no dual-emit fallback)
- Wire `*gala.Gala` into the `WorkflowEngine` constructor instead of `soiree.Emitter`

### `serveropts/integration_ingest.go`

- Replace the dedicated `soiree.EventBus` for integration ingest with a gala topic + listener registration on the shared gala instance
- Delete `WithIntegrationIngestEvents`; register integration ingest listeners in `WithGala` or a new `WithIntegrationListeners` serveropt

## Update `internal/workflows/options.go`

- Remove `GalaConfig` struct entirely (no dual-emit, no topic modes, no migration controls)
- Remove `GalaTopicMode`, `GalaTopicModeSoireeOnly`, `GalaTopicModeDualEmit`, `GalaTopicModeV2Only`
- Fold remaining gala configuration (worker count, max retries, queue name) into the main `Config` struct or keep a slim `QueueConfig` if separation is clearer
- Remove `Config.Gala` field

## Update `internal/workflows/context.go`

- Remove soiree-based workflow context flags (`WorkflowBypassContextKey`, `WorkflowAllowEventEmissionKey` if implemented via raw context)
- Use `gala.ContextFlagWorkflowBypass` and `gala.ContextFlagWorkflowAllowEventEmission` exclusively via `gala.WithFlag()` / `gala.HasFlag()`
- Consolidate `shouldSkipWorkflowMutationForBypass()` to only check gala flags (remove dual soiree+gala check)

## Replace `soiree.Pool` Usage in `internal/graphapi/`

`soiree.Pool` is a standalone worker pool (wraps `alitto/pond`), unrelated to event dispatch. Two options:

- **Option A (recommended):** Replace with direct `pond` usage since `soiree.Pool` is a thin wrapper. Create a small internal pool helper or use `pond` directly in `internal/graphapi/pool.go`.
- **Option B:** Extract `soiree.Pool` into its own utility package if broader reuse is needed.

Files affected:
- `internal/graphapi/pool.go` -- `queryResolver.withPool()`, `mutationResolver.withPool()`
- `internal/graphapi/resolveropts.go` -- pool initialization
- `internal/graphapi/history/resolver.go` -- history pool
- `internal/graphapi/tools_test.go` -- test pool setup

## Integration Workloads: Separate Queue + Typed Envelopes

### Separate Queue Class

Register a dedicated River queue for integration operations in gala config:
- Queue name: `integrations` (separate from default `events` queue)
- Configure per-queue concurrency and retries via River's queue config
- Integration listeners dispatch to the `integrations` queue via `gala.Headers` with queue override

### Typed Operation Envelope

Define `IntegrationOperationEnvelope` in `internal/integrations/`:
- Operation type enum (sync, import, export, webhook, etc.)
- Explicit timeout policy per operation type
- Retry policy per operation type (use River's `MaxAttempts` + backoff)
- Idempotency key policy keyed by `(provider, operation, org_id, run_id)`

### Remove Bespoke Integration Bus

- Delete `internal/httpserve/serveropts/integration_ingest.go` standalone soiree bus
- Register integration listeners on the shared gala instance with queue routing to `integrations`
- Move `internal/integrations/events.go` typed topic to gala topic registration

## Update `cmd/serve.go`

- Remove soiree imports
- Remove any soiree-related shutdown/cleanup
- Gala shutdown is already handled (`galaApp.StopWorkers`, `galaApp.Close`)
- Remove dual-emit configuration references

## Remove `internal/ent/eventqueue/gala_adapter.go` Dual-Mode Code

- Remove `projectGalaFlagsFromWorkflowContext()` (consolidated into gala context codecs)
- Simplify `NewMutationGalaEnvelope` -- context flags are handled by `gala.ContextManager.Capture()` directly, no manual projection needed
- Remove any references to soiree-style properties fallbacks

## Delete or Deprecate Soiree Package

After all call sites are migrated:
- Remove `pkg/events/soiree/workflow.go` (payloads moved to gala)
- Remove soiree from `go.mod` if no other consumers remain
- If `soiree.Pool` is extracted (option B above), that lives separately
- Run `go mod tidy` to clean up

## Update `internal/ent/hooks/event_gala_hook.go`

- Rename to `event_hook.go`
- Absorb the `parseEventID`, `parseSoftDeleteEventID`, `getOperation`, `mutationChangedAndClearedFields`, `mutationProposedChanges` helpers from the deleted `event.go`
- Remove `FailOnEnqueueError` parameter (no dual-emit, errors are terminal)
- Remove the `emitEventOn` predicate pattern (gala registry's `InterestedIn` is sufficient)

## Update Tests

- `internal/ent/hooks/listeners_workflow_parity_test.go` -- rewrite to test gala-only flow
- `internal/ent/hooks/listeners_slack_gala_test.go` -> `listeners_slack_test.go` -- remove soiree references
- `internal/ent/eventqueue/gala_adapter_test.go` -- remove dual-mode assertions, simplify
- `internal/workflows/engine/*_test.go` -- update all engine tests to construct gala runtime instead of soiree emitter
- `internal/workflows/reconciler/*_test.go` -- update reconciler to use gala
- `internal/graphapi/tools_test.go` -- replace `soiree.New()` / `soiree.NewPool()` with gala or pond equivalents
- All test files that create `NewEventerPool` or `NewEventer` -- replace with gala setup
