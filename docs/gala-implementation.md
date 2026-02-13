# Eventing V2 Implementation Plan (River-Native, Workflow-First)

## Purpose
This document defines the end-state architecture and phased implementation plan for a new gala package that is independent from `pkg/events/soiree`, uses River natively for durability, and is designed to simplify workflow engine adoption before broad production usage.

This plan intentionally reflects lessons learned from prior Soiree + mutation outbox iterations.

## Context Snapshot
- Current Soiree path is in-memory by default and not sufficient for durable multi-pod processing.
- Redis-backed Soiree durability is not viable for our current contracts.
- Workflow engine depends on event semantics but is not yet broadly active; this is the best point to establish final architecture.
- Existing listeners remain valuable and should be migrated topic-by-topic with controlled risk.
- We can dual-emit during migration (`soiree` + `gala`) while keeping ownership clear per topic.

## Core Goals
- Formal, testable delivery semantics (durable, retryable, idempotent).
- Strongly typed contracts (no loose payload/metadata handling).
- Minimal listener boilerplate (define listener once, registration is seamless).
- First-class context reconstruction (auth + context flags + operational metadata).
- Workflow-first design using River primitives rather than bespoke queue mechanics.
- Extensible for long-running integration tasks.
- Preserve global domain semantics: same topic behavior regardless of emit origin.

## Non-Goals (Initial)
- Rewrite every legacy Soiree listener in one PR.
- Introduce broad DI framework usage into runtime code paths.
- Rebuild every queueing pattern in the codebase immediately.

## Engineering Constraints (Required)
- Avoid unnecessary defensive boilerplate and redundant guards.
- Every function (exported and unexported) must have comments.
- Every exported type/field must have comments.
- Use `samber/lo` and `samber/mo` where they improve clarity and reduce repetitive code.
- Use `samber/do` only where appropriate for composition/setup (not core runtime hot paths).
- Use Go generics for reusable typed constructs.
- Static errors must live in `errors.go`.
- Every new package/subpackage must include `doc.go`.
- Avoid stringly-typed logic and ad-hoc string comparisons where typed enums/constants are possible.
- Use `contextx` for context key set/get patterns.
- Reuse existing helpers before introducing new map/slice/string utilities.

## Delivery Semantics (V2)
- Durable dispatch through River jobs.
- At-least-once delivery.
- Idempotent consumer contract required for all durable listeners.
- Transactional enqueue where available (`InsertTx`) for mutation/state-change coupling.
- Queue partitioning by domain and workload profile.
- Structured retries + terminal handling.

## Proposed Package Layout

```text
pkg/gala/
  doc.go
  errors.go

  types.go                # Topic, EventID, headers, strongly typed envelope metadata
  envelope.go             # Durable JSON-safe envelope model
  codec.go                # Codec interfaces + generic helpers
  registry.go             # Topic/listener/codec registration
  listener.go             # Listener definition + attachment helpers

  bus/
    doc.go
    errors.go
    emitter.go            # Emitter interface + dispatch API
    dispatcher.go         # Routes to inline or durable path based on policy

  river/
    doc.go
    errors.go
    args.go               # River job args (durable envelope)
    worker.go             # Generic dispatch worker (rehydrate + invoke)
    client.go             # River client abstraction (wrap riverqueue prior art)
    policy.go             # Per-topic mode + queue policy

  context/
    doc.go
    keys.go               # Typed context keys for gala metadata
    snapshot.go           # Capture context metadata into durable snapshot
    restore.go            # Rehydrate metadata back into context

  runtime/
    doc.go
    runtime.go            # Bootstrap, attach listeners, start/stop workers
    options.go            # Runtime options and defaults

  compat/
    doc.go
    soiree_bridge.go      # Optional helper during dual-emit migration
```

## Listener Ergonomics (Fixing Current Clunkiness)

### Requirement
Defining a listener should be enough to make registration straightforward without repeated wrappers.

### Proposed Shape

```go
// Package gala demonstrates listener-first registration.
package gala

import "context"

// ListenerID identifies a registered listener.
type ListenerID string

// Handler processes typed events.
type Handler[T any] func(context.Context, T) error

// Definition describes a single listener binding.
type Definition[T any] struct {
	// Topic is the strongly typed topic handled by this listener.
	Topic Topic[T]
	// Name is a stable listener name used for metrics/idempotency.
	Name string
	// Handle contains the typed callback.
	Handle Handler[T]
}

// Register attaches the definition to a registry/runtime.
func (d Definition[T]) Register(r *Registry) (ListenerID, error) {
	return r.register(d)
}
```

```go
// Example: listener defines itself and is attached directly.
var HandleWorkflowMutation = gala.Definition[workflows.MutationDetected]{
	Topic: workflowtopics.MutationDetected,
	Name:  "workflow.handle_mutation",
	Handle: func(ctx context.Context, evt workflows.MutationDetected) error {
		return workflowService.HandleMutation(ctx, evt)
	},
}
```

```go
// Bootstrap usage.
ids, err := runtime.Attach(
	workflowlisteners.HandleWorkflowMutation,
	workflowlisteners.HandleAssignmentCompleted,
	integrationlisteners.HandleOperationRequested,
)
```

No per-listener wrapper + separate bind + separate pool registration steps.

## Emit API Ergonomics (Replacing Err Channel Pattern)
The v2 API should not require call sites to manually drain channels for enqueue outcomes.

```go
package gala

import "context"

// EmitReceipt captures dispatch outcome.
type EmitReceipt struct {
	// EventID is the stable dispatch event ID.
	EventID EventID
	// Accepted reports whether dispatch was accepted for processing.
	Accepted bool
	// Err contains a terminal dispatch error.
	Err error
}

// Emitter dispatches typed topics.
type Emitter interface {
	// Emit dispatches a payload and returns a synchronous receipt.
	Emit(context.Context, TopicName, any) EmitReceipt
}
```

Call sites should use receipt helpers + centralized metrics/logging, not per-call ad-hoc drain loops.

## Strongly Typed Topics and Envelope

```go
package gala

import "github.com/theopenlane/utils/ulids"

// EventID is the stable idempotency identifier.
type EventID string

// NewEventID creates a new event ID.
func NewEventID() EventID {
	return EventID(ulids.New().String())
}

// TopicName is a stable typed topic key.
type TopicName string

// Topic is a strongly typed event topic.
type Topic[T any] struct {
	// Name is the stable topic identifier.
	Name TopicName
}
```

```go
package gala

import "time"

// Envelope carries JSON-safe durable event data.
type Envelope struct {
	// ID is the unique durable event identifier.
	ID EventID `json:"id"`
	// Topic is the destination topic.
	Topic TopicName `json:"topic"`
	// SchemaVersion is the payload schema version.
	SchemaVersion int `json:"schema_version"`
	// OccurredAt captures emit time in UTC.
	OccurredAt time.Time `json:"occurred_at"`
	// Headers carries typed operational metadata.
	Headers Headers `json:"headers"`
	// Payload contains encoded event payload bytes.
	Payload []byte `json:"payload"`
	// ContextSnapshot captures restorable context metadata.
	ContextSnapshot ContextSnapshot `json:"context_snapshot"`
}
```

## Typed Metadata (No Stringly Contracts)

```go
package gala

// EmitMode controls per-topic dispatch behavior.
type EmitMode string

const (
	// EmitModeSoireeOnly keeps legacy-only emission.
	EmitModeSoireeOnly EmitMode = "soiree_only"
	// EmitModeDualEmit emits to both legacy and v2.
	EmitModeDualEmit EmitMode = "dual_emit"
	// EmitModeV2Only emits only through v2.
	EmitModeV2Only EmitMode = "v2_only"
)

// QueueClass identifies River queue classes for workload isolation.
type QueueClass string

const (
	// QueueClassWorkflow is for workflow state transitions.
	QueueClassWorkflow QueueClass = "workflow"
	// QueueClassIntegration is for integration operations.
	QueueClassIntegration QueueClass = "integration"
	// QueueClassNotification is for notification fanout.
	QueueClassNotification QueueClass = "notification"
)
```

## Codec Registry (Generic + Reusable)

```go
package gala

import "context"

// Codec encodes and decodes a topic payload.
type Codec[T any] interface {
	// Encode converts payload into bytes.
	Encode(context.Context, T) ([]byte, error)
	// Decode converts bytes into payload.
	Decode(context.Context, []byte) (T, error)
}

// Registration ties topic, codec, and policy together.
type Registration[T any] struct {
	// Topic is the typed event topic.
	Topic Topic[T]
	// Codec is the serializer for the topic.
	Codec Codec[T]
	// Policy defines durable behavior and queue routing.
	Policy TopicPolicy
}
```

Use generics to avoid N switch/case enum or payload mappers.

## Context Capture and Restore (contextx-first)

```go
package galactx

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// Snapshot captures context values needed for durable rehydration.
type Snapshot struct {
	// OrgID is the organization context.
	OrgID string `json:"org_id,omitempty"`
	// SubjectID is the authenticated subject.
	SubjectID string `json:"subject_id,omitempty"`
	// Flags carries typed boolean context flags.
	Flags map[Flag]bool `json:"flags,omitempty"`
}

// Flag identifies known context flags.
type Flag string

const (
	// FlagWorkflowBypass indicates workflow bypass behavior.
	FlagWorkflowBypass Flag = "workflow_bypass"
)

// WithFlag sets a typed flag in context.
func WithFlag(ctx context.Context, flag Flag) context.Context {
	return contextx.With(ctx, flag)
}
```

Context restoration must be explicit and centrally tested.

## River Integration (Prior Art: riverboat)
Use existing `riverqueue` client patterns from `/Users/manderson/riverboat/pkg/riverqueue`:
- shared client abstraction with `Insert`, `InsertTx`, `GetRiverClient`, lifecycle.
- avoid bespoke DB plumbing.

### Durable Job Args

```go
package galariver

import (
	"github.com/riverqueue/river"
)

// DispatchArgs is the generic River durable event job.
type DispatchArgs struct {
	// Envelope is the encoded durable event.
	Envelope []byte `json:"envelope"`
}

// Kind returns River job kind.
func (DispatchArgs) Kind() string {
	return "gala_dispatch_v1"
}

// InsertOpts returns queue defaults.
func (DispatchArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: "events_workflow"}
}
```

### Worker

```go
package galariver

import (
	"context"

	"github.com/riverqueue/river"
)

// DispatchWorker rehydrates and dispatches event envelopes.
type DispatchWorker struct {
	river.WorkerDefaults[DispatchArgs]

	runtime Runtime
}

// Work processes one durable event.
func (w *DispatchWorker) Work(ctx context.Context, job *river.Job[DispatchArgs]) error {
	return w.runtime.DispatchEnvelope(ctx, job.Args.Envelope)
}
```

## Runtime DI Approach

### Principle
- Runtime paths use explicit constructor wiring.
- `samber/do` may be used at bootstrap/composition root only.

### Suggested
- Keep core gala runtime independent of container APIs.
- Build container module for app startup if helpful.

```go
// Composition root example (optional do usage only here).
func BuildEventingRuntime(i *do.Injector) (*galaruntime.Runtime, error) {
	client := do.MustInvoke[riverqueue.JobClient](i)
	workflow := do.MustInvoke[*engine.WorkflowEngine](i)
	entitlements := do.MustInvoke[*entitlements.Manager](i)

	return galaruntime.New(galaruntime.Options{
		JobClient: client,
		Workflow: workflow,
		Entitlements: entitlements,
	})
}
```

## Typed Dependency Access for Handlers
Handlers should receive typed dependency accessors to avoid repetitive pointer extraction and type assertions.

```go
package galaruntime

import "github.com/samber/mo"

// Services holds runtime-scoped dependencies.
type Services struct {
	// Workflow is the workflow engine service.
	Workflow *engine.WorkflowEngine
	// Entitlements is the entitlements manager.
	Entitlements *entitlements.Manager
	// TokenManager is the token manager service.
	TokenManager *authmanager.TokenManager
}

// ResolveWorkflow resolves workflow service with Option semantics.
func (s Services) ResolveWorkflow() mo.Option[*engine.WorkflowEngine] {
	return mo.TupleToOption(s.Workflow, s.Workflow != nil)
}
```

Use explicit structs first. Optional `do` usage stays at bootstrap/composition root only.

## Workflow-First Domain Contracts
Prefer durable commands over broad generic mutation payloads for workflow internals.

### Example command topics
- `workflow.command.trigger`
- `workflow.command.advance`
- `workflow.command.timeout_expire`
- `integration.command.execute`

Keep external mutation emits available for cross-domain listeners, but workflow engine internals should move toward command contracts.

## Integration Extensibility (Long Running Workloads)
- Separate queue classes and worker pools for integration workloads.
- Add typed operation envelopes with explicit timeout/retry policy.
- Use River per-queue concurrency and retries rather than bespoke goroutine orchestration.
- Add idempotency key policy per integration operation type.

## Dual Emit Migration Strategy

### Per-topic modes
- `soiree_only`
- `dual_emit`
- `v2_only`

### Rules
- In `dual_emit`, both emits share canonical `event_id`.
- Listener ownership must be single-system per topic at any point.
- Use parity tests + metrics before flipping to `v2_only`.

### Suggested progression
1. Workflow-relevant mutation topics -> `dual_emit`.
2. Workflow consumers read from v2 only.
3. Migrate legacy listeners topic-by-topic.
4. Flip topic to `v2_only`.

## Progress Snapshot (Updated 2026-02-13)

### Completed in this branch
- `pkg/gala` foundation exists (`doc.go`, `errors.go`, typed primitives, runtime, registry, codecs, context snapshot, River dispatcher/worker, tests).
- Mutation dual-emit path is wired in ent hooks to build Gala envelopes and dispatch durably while preserving legacy Soiree emit path.
- `cmd/serve.go` now wires Gala runtime + River worker startup behind workflow Gala flags.
- Gala entitlement listener scaffold exists for initial topics (`organization`, `organization_setting`) with DI-backed dependency resolution.
- Gala topic registration for those entitlement listeners is now ensured during registration.
- `pkg/gala` durable dispatcher now supports `InsertTx` when a `pgx.Tx` is provided via context (`WithPGXTx` / `PGXTxFromContext`).
- Per-topic migration mode config is wired (`soiree_only`, `dual_emit`, `v2_only`) and applied by the mutation hook dispatch path.
- Initial workflow-facing Gala listeners are registered for workflow object mutation topics and workflow assignment mutation handling.
- Hook tests cover topic-mode behavior (`v2_only`, `dual_emit`, v2 fail-open fallback).

### Not completed yet
- Per-topic migration matrix and `v2_only` cutover implementation.
- Workflow command-topic migration and replay/recovery hardening.
- Broader parity suite covering all high-risk listener behavior changes.

## TODO Tracker

## Phase 0: Foundation
- [x] EVT2-001 Create `pkg/gala/doc.go` and package scope documentation.
- [x] EVT2-002 Create `pkg/gala/errors.go` and central static errors.
- [x] EVT2-003 Create typed core primitives (`Topic`, `EventID`, `Headers`, `EmitMode`, `QueueClass`).
- [ ] EVT2-004 Add lint/test guard to require comments on all functions and exported fields in new package paths.
- [ ] EVT2-005 Add architecture README sections to `README.md` or `docs/` index linking this plan.

### Phase 0 Acceptance
- [x] Foundation package compiles.
- [ ] Comments present for all functions/types/fields in new package.
- [ ] No ad-hoc errors outside `errors.go` in `pkg/gala/**`.

## Phase 1: Durable Envelope + Codec
- [x] EVT2-101 Implement envelope model + schema versioning.
- [x] EVT2-102 Implement generic codec interfaces and registration.
- [x] EVT2-103 Implement JSON codec helpers with typed decode.
- [x] EVT2-104 Implement context snapshot capture/restore (`contextx` usage only).
- [x] EVT2-105 Add unit tests for envelope encode/decode and context restore.

### Phase 1 Acceptance
- [x] Envelope round-trip tests pass.
- [x] Context snapshot round-trip tests pass.
- [ ] No stringly metadata maps for core controls.

## Phase 2: River Runtime
- [x] EVT2-201 Implement River dispatch args and worker.
- [x] EVT2-202 Implement dispatcher with per-topic mode and queue policy.
- [x] EVT2-203 Support `InsertTx` path for transactional enqueue where caller has tx.
- [ ] EVT2-204 Add retries/terminal state policy and structured metrics.
- [x] EVT2-205 Add graceful start/stop wiring with River client lifecycle.

### Phase 2 Acceptance
- [ ] Worker integration tests pass with River test DB.
- [ ] Dispatch is durable and retry behavior is deterministic.
- [ ] Runtime startup/shutdown hooks are verified.

## Phase 3: Listener UX + Runtime Attach
- [x] EVT2-301 Implement listener `Definition[T]` with direct registration helpers.
- [ ] EVT2-302 Implement registry attach APIs to remove wrapper boilerplate.
- [x] EVT2-303 Add typed dependency resolver for listener handlers (explicit first, optional container at setup).
- [ ] EVT2-304 Add compile-time helper for common client extraction and context metadata retrieval.
- [ ] EVT2-305 Add docs and examples replacing old err-channel drain patterns.

### Phase 3 Acceptance
- [ ] New listener requires one definition and one attach call only.
- [ ] No repetitive error-channel drain code in migrated call sites.
- [ ] Dependency access pattern documented and tested.

## Phase 4: Workflow-First Integration
- [ ] EVT2-401 Add workflow command topics and payload types.
- [ ] EVT2-402 Route workflow engine emits through v2 runtime.
- [ ] EVT2-403 Keep mutation-driven workflow triggers in dual emit mode initially.
- [ ] EVT2-404 Add idempotency tests for workflow command handling.
- [ ] EVT2-405 Add replay and recovery tests for workflow command queues.

### Phase 4 Acceptance
- [ ] Workflow events/commands durable and recoverable.
- [ ] Workflow semantics are not dependent on in-memory Soiree behavior.

## Phase 5: Integration Workloads
- [ ] EVT2-501 Add integration command topics and queue classes.
- [ ] EVT2-502 Add per-operation retry and timeout policy configuration.
- [ ] EVT2-503 Add idempotency strategy for integration operations.
- [ ] EVT2-504 Add long-running task observability and backpressure controls.
- [ ] EVT2-505 Add failure reconciliation handlers for integration tasks.

### Phase 5 Acceptance
- [ ] Integration tasks run on dedicated queues.
- [ ] Retry and timeout behavior configurable per operation type.

## Phase 6: Migration and Cutover
- [x] EVT2-601 Add per-topic mode config (`soiree_only`, `dual_emit`, `v2_only`).
- [x] EVT2-602 Wire mutation emit path for dual emit with canonical event ID.
- [ ] EVT2-603 Migrate first target topics (workflow-related) to `v2_only`.
- [x] EVT2-604 Migrate selected legacy listeners topic-by-topic.
- [ ] EVT2-605 Remove legacy emission for migrated topics.

### Phase 6 Acceptance
- [ ] Topic ownership matrix is explicit and conflict-free.
- [ ] Parity tests pass for migrated topics.
- [ ] Production metrics confirm equivalent/expected behavior.

## Phase 7: Hardening and Cleanup
- [ ] EVT2-701 Add runbook for incident handling and replay.
- [ ] EVT2-702 Add queue pause/resume procedures and operator docs.
- [ ] EVT2-703 Add final deprecation plan for unused Soiree durable paths.
- [ ] EVT2-704 Remove dead migration code.
- [ ] EVT2-705 Final architecture review + ADR.

## Topic Ownership Matrix (to fill during migration)

| Topic | Current Owner | Target Owner | Mode | Notes |
| --- | --- | --- | --- | --- |
| `<mutation.workflow_object.*>` | soiree | gala | dual_emit | first migration tranche |
| `<workflow.command.*>` | n/a | gala | v2_only | new contracts |
| `<integration.command.execute>` | mixed | gala | v2_only | long-running |

## Testing Plan
- Unit tests:
  - codec correctness
  - context snapshot restore
  - policy routing
  - listener registration ergonomics
- Integration tests:
  - River worker dispatch
  - transactional enqueue (`InsertTx`)
  - retry + terminal handling
- Parity tests:
  - selected mutation topics in dual emit
  - workflow trigger equivalence
- Failure tests:
  - worker crash/restart replay
  - poison payload terminal behavior

## Open Decisions
- [ ] OPN-001 Should workflow internals fully move to command topics immediately, or phase commands behind mutation emits?
- [ ] OPN-002 Should `samber/do` be adopted only in `cmd/serve` composition root for gala runtime setup?
- [ ] OPN-003 Queue naming and concurrency defaults per domain.
- [ ] OPN-004 Retention and replay window policy for River jobs/events.

## Initial Execution Order (Recommended)
1. Phase 0 + Phase 1
2. Phase 2
3. Phase 3
4. Phase 4 (workflow-first)
5. Phase 6 (dual emit + migration)
6. Phase 5 and Phase 7 in parallel as needed

## Notes for Implementation Sessions
- Prefer existing helpers for map/slice transforms (`lo.Map`, `lo.Filter`, `lo.PickBy`, `lo.Assign`, `lo.Uniq`).
- Prefer typed enum parsing helpers over local switch/case parsers.
- Keep context restore centralized; avoid per-listener ad-hoc context mutation.
- Keep idempotency keys stable across dual emits.
- Keep logging consistent and avoid mode-specific message churn unless behavior differs.
