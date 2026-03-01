# Helper Reuse Guide

## Purpose
Define the existing helper surface we should reuse for integrations and workflow-event plumbing so we do not add overlapping bespoke utilities

## Decision Table
| Need | Reuse | Avoid adding |
|---|---|---|
| Decode arbitrary JSON-ish input into typed output | `jsonx.RoundTrip` | Local marshal+unmarshal wrappers |
| Convert arbitrary value to `map[string]any` | `jsonx.ToMap` | Local `toMap` helpers |
| Deep clone nested `map[string]any` | `mapx.DeepCloneMapAny` | Ad-hoc recursive clone funcs |
| Deep merge nested `map[string]any` | `mapx.DeepMergeMapAny` | Bespoke merge logic |
| Remove zero-value leaves before merge/overlay | `mapx.PruneMapZeroAny` | One-off prune logic |
| Build set semantics from slices | `mapx.MapSetFromSlice` | Repeated inline map-set builders |
| Slice intersection with uniqueness | `mapx.MapIntersectionUnique` | Custom loops for intersection/dedupe |
| Register typed event topic + listeners | `gala.RegisterListeners` | Custom event registries |
| Manual topic registration | `gala.RegisterTopic` + `gala.AttachListener` | Local topic/listener maps |
| Durable/in-memory event runtime | `gala.NewGala` / `gala.NewInMemory` | Parallel event runtime abstractions |
| Emit events with transport headers | `gala.EmitWithHeaders` / `gala.EmitEnvelope` | Custom envelope dispatch code |
| Context snapshot/restore across durable boundaries | `gala.ContextManager`, `gala.NewContextCodec`, `gala.NewTypedContextCodec` | Bespoke context serialization |
| Workflow bypass propagation | `gala.WithFlag`, `gala.HasFlag`, existing context flags | New duplicate flag systems |
| JSON schema generation (core config) | `go run jsonschema/schema_generator.go` | New config schema generators |
| JSON schema generation (workflow definition) | `go run jsonschema/workflow_schema_generator.go` | New workflow schema generators |
| Schema/docs task orchestration | `jsonschema/Taskfile.yaml` tasks | Parallel ad-hoc schema/doc scripts |
| Merge integration provider state | `state.IntegrationProviderState.MergeProviderData` | Provider-specific keystore switch blocks |

## Package Inventory

### `pkg/jsonx`
- `RoundTrip(input, output)` for tolerant decode/encode round-trips
- `ToMap(value)` for object conversion with static `ErrObjectExpected`
- Use this before writing any JSON normalization helper

### `pkg/mapx`
- `DeepCloneMapAny` for nested map payload/config copies
- `CloneMapStringSlice` for `map[string][]string`
- `PruneMapZeroAny` for overlay cleanup
- `DeepMergeMapAny` for nested config overlays
- `MapSetFromSlice` for set-like lookups
- `MapIntersectionUnique` for stable deduped intersections

### `pkg/gala`
- Runtime/bootstrap: `NewGala`, `NewInMemory`, `Config`, `DispatchMode*`
- Registry/topic helpers: `NewRegistry`, `Registration[T]`, `Topic[T]`, `Definition[T]`, `RegisterListeners`, `RegisterTopic`, `AttachListener`
- Payload codec helper: `JSONCodec[T]`
- Emission/dispatch: `EmitWithHeaders`, `EmitEnvelope`, `DispatchEnvelope`
- Envelope/headers: `Envelope`, `Headers`, `EventID`, `NewEventID`
- Context durability: `ContextManager`, `NewContextCodec`, `NewTypedContextCodec`, `WithFlag`, `HasFlag`
- Durable dispatch helpers: `NewRiverDispatcher`, `NewRiverDispatchArgs`, `NewRiverDispatchWorker`
- Worker lifecycle: `StartWorkers`, `StopWorkers`, `WaitIdle`, `Close`
- In-process pool: `NewPool`, `WithWorkers`, `WithPoolName`, `Submit`, `SubmitMultipleAndWait`
- Reuse `pkg/gala/errors.go` sentinel errors rather than redefining equivalent runtime errors

### `jsonschema`
- Canonical generators live at:
- `jsonschema/schema_generator.go`
- `jsonschema/workflow_schema_generator.go`
- Canonical task entrypoints live at:
- `jsonschema/Taskfile.yaml`
- Existing generator extension points to reuse before adding new generators:
- `initializeIntegrationProviders` for provider default bootstrapping
- `buildCollectionIndex` and collection index helpers for env/example expansion
- `workflowTypeMapper` for workflow enum/type schema mapping
- `workflowDefinitionDecorators` for workflow schema description overrides
- Use these generators/tasks when extending config/workflow schema outputs instead of creating parallel schema pipelines

### `common/integrations/state`
- `IntegrationProviderState.ProviderData` for provider-keyed state reads
- `IntegrationProviderState.MergeProviderData` for deep merged provider-keyed writes

## Branch-specific alignment updates applied
- Replaced new shallow `maps.Clone` calls on `map[string]any` payload/config data with `mapx.DeepCloneMapAny` in workflow integration/notification paths

## Guardrails
1. Before adding a helper, grep `pkg/jsonx`, `pkg/mapx`, `pkg/gala`, and `jsonschema` for an existing equivalent
2. If an equivalent exists, use it directly
3. If no equivalent exists, document why existing helpers are insufficient in the PR/commit notes before adding a new helper
