# Gala

`pkg/gala` is a typed event runtime for emitting domain events with:

- Compile-time topic/payload contracts.
- Per-topic dispatch policy (`inline`, `durable`, `dual`).
- Envelope-based durable delivery (River integration included).
- Context snapshot/rehydration across async boundaries.
- Listener dependency injection via `samber/do`.

It is intended as a durable, River-native replacement for ad-hoc in-memory dispatch.

## Architecture

### Core types

- `Topic[T]`: typed topic contract (`Name`, optional `SchemaVersion`).
- `Registration[T]`: binds a topic to a `Codec[T]` and `TopicPolicy`.
- `Definition[T]`: typed listener (`Name`, `Handle`).
- `Runtime`: emit/dispatch engine.
- `Envelope`: durable event shape (`ID`, `Topic`, `Headers`, `Payload`, `ContextSnapshot`).

### Runtime flow

1. Register topic + codec (+ policy).
1. Attach listeners for that topic.
1. Emit payload (`Emit`, `EmitWithHeaders`, or `EmitTyped`).
1. Runtime encodes payload, captures context, builds `Envelope`.
1. Runtime dispatches by policy (`inline` = in-process, `durable` = enqueue only, `dual` = both; any path failure returns `ErrDualDispatchFailed`).

### Registration helpers

You can register in one step instead of manual topic+listener wiring:

- `RegisterListener`: topic + listener (idempotent topic registration).
- `RegisterListeners`: batch version.
- `RegisterDurableListeners`: batch durable registration with one shared `QueueClass`.

## Quick start (inline mode)

```go
package main

import (
	"context"
	"log"

	"github.com/theopenlane/core/pkg/gala"
)

type UserCreated struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

var userCreatedTopic = gala.Topic[UserCreated]{
	Name:          gala.TopicName("user.created"),
	SchemaVersion: 1,
}

func main() {
	runtime, err := gala.NewRuntime(gala.RuntimeOptions{})
	if err != nil {
		log.Fatal(err)
	}

	err = (gala.Registration[UserCreated]{
		Topic: userCreatedTopic,
		Codec: gala.JSONCodec[UserCreated]{},
		Policy: gala.TopicPolicy{
			EmitMode: gala.EmitModeInline,
		},
	}).Register(runtime.Registry())
	if err != nil {
		log.Fatal(err)
	}

	_, err = (gala.Definition[UserCreated]{
		Topic: userCreatedTopic,
		Name:  "welcome-email",
		Handle: func(ctx gala.HandlerContext, payload UserCreated) error {
			log.Printf("send welcome email to %s (%s)", payload.UserID, payload.Email)
			return nil
		},
	}).Register(runtime.Registry())
	if err != nil {
		log.Fatal(err)
	}

	receipt := gala.EmitTyped(context.Background(), runtime, userCreatedTopic, UserCreated{
		UserID: "usr_123",
		Email:  "user@example.com",
	})
	if receipt.Err != nil {
		log.Fatal(receipt.Err)
	}
	log.Printf("event accepted: id=%s", receipt.EventID)
}
```

## Payload codecs

Every topic registration requires a `Codec[T]`:

```go
type Codec[T any] interface {
	Encode(T) ([]byte, error)
	Decode([]byte) (T, error)
}
```

`JSONCodec[T]` is the default and recommended starting point. Use a custom codec when you need non-JSON formats, encryption, or strict compatibility decoding.

## Dispatch policy

`TopicPolicy` controls how each topic is emitted:

- `EmitModeInline`: listeners run in the emitter call path.
- `EmitModeDurable`: envelope is sent to durable transport only.
- `EmitModeDual`: durable + inline.
- `Durable: true`: compatibility shortcut; treated as durable when `EmitMode` is empty.
- `QueueClass` / `QueueName`: durable queue routing hints.

If policy is omitted, effective mode is inline.

## Execution semantics

- Listeners run sequentially in registration order for a topic.
- Inline dispatch is fail-fast: dispatch stops on the first listener error.
- Fail-fast keeps ordering deterministic and surfaces listener failure back to the emit caller.
- If you want best-effort fanout, handle/log errors inside each listener and return `nil`, or route work through durable listeners/queues.
- Emitting to a topic with no listeners is a successful no-op after decode/restore.
- `Emit` creates a new `EventID`; `EmitEnvelope` uses the supplied `Envelope.ID`.
- `Topic.SchemaVersion` defaults to `1` when unset.

## Context propagation

Gala snapshots selected context values at emit time and restores them for listeners.

### Default behavior

`NewRuntime` always wires an auth context codec (`auth_user`), so authenticated user context is propagated automatically.

### Custom typed context codec

```go
type Actor struct {
	ID string `json:"id"`
}

contextManager, err := gala.NewContextManager(
	gala.NewTypedContextCodec[Actor](gala.ContextKey("actor")),
)
if err != nil {
	panic(err)
}

runtime, err := gala.NewRuntime(gala.RuntimeOptions{
	ContextManager: contextManager,
})
if err != nil {
	panic(err)
}
```

Set and read the typed value in listener context:

```go
import "github.com/theopenlane/utils/contextx"

emitCtx := contextx.With(context.Background(), Actor{ID: "actor-1"})
emitCtx = gala.WithFlag(emitCtx, gala.ContextFlagWorkflowBypass)

// ... emit using emitCtx ...

// inside listener:
actor, ok := contextx.From[Actor](handlerCtx.Context)
if ok {
	_ = actor.ID
}

if gala.HasFlag(handlerCtx.Context, gala.ContextFlagWorkflowBypass) {
	// flag propagated from emit context
}
```

Built-in flags:

- `ContextFlagWorkflowBypass`
- `ContextFlagWorkflowAllowEventEmission`

## Listener dependency injection

Each listener receives a `HandlerContext` containing a `do.Injector`.

```go
type Formatter struct {
	Prefix string
}

injector := gala.NewInjector()
gala.ProvideValue(injector, &Formatter{Prefix: "fmt"})

runtime, err := gala.NewRuntime(gala.RuntimeOptions{
	Injector: injector,
})
if err != nil {
	panic(err)
}

// inside listener:
formatter, err := gala.ResolveFromContext[*Formatter](handlerCtx)
if err != nil {
	return err
}
_ = formatter.Prefix
```

### Accessing core runtime dependencies in listeners

Preferred pattern: inject concrete dependencies directly. Inject `*generated.Client` only when the listener needs DB query/builders.

```go
// runtime bootstrap
gala.ProvideValue(runtime.Injector(), dbClient.EntitlementManager) // *entitlements.StripeClient
gala.ProvideValue(runtime.Injector(), dbClient.TokenManager)       // *tokens.TokenManager
gala.ProvideValue(runtime.Injector(), wfEngine)                    // *engine.WorkflowEngine
gala.ProvideValue(runtime.Injector(), dbClient)                    // optional: *generated.Client for data access
gala.ProvideNamedValue(runtime.Injector(), "stripe_client", stripeClient)

// listener
func handle(ctx gala.HandlerContext, payload MyPayload) error {
	entitlementsClient, err := gala.ResolveFromContext[*entitlements.StripeClient](ctx)
	if err != nil {
		return err
	}
	tokenManager, err := gala.ResolveFromContext[*tokens.TokenManager](ctx)
	if err != nil {
		return err
	}

	// explicitly injected named dependency
	stripeClient, err := gala.ResolveNamedFromContext[*stripe.Client](ctx, "stripe_client")
	if err != nil {
		return err
	}

	client, err := gala.ResolveFromContext[*generated.Client](ctx) // optional DB access dependency
	if err != nil {
		return err
	}

	_ = entitlementsClient
	_ = tokenManager
	_ = stripeClient
	_ = client
	return nil
}
```

Helper API:

- Register: `Provide`, `ProvideValue`, `ProvideNamedValue`
- Resolve: `Resolve`, `ResolveNamed`, `ResolveOption`, `MustResolve`
- Listener-context resolve: `ResolveFromContext`, `ResolveNamedFromContext`, `ResolveOptionFromContext`, `MustResolveFromContext`, `MustResolveNamedFromContext`
- Missing listener injector in context resolve helpers returns `ErrInjectorRequired`.

## Durable delivery with River

Preferred startup path is `NewRiverRuntime`, which creates and owns:

- a dedicated Gala River client (separate from your app's other River plumbing)
- queue/class routing for Gala durable dispatch
- optional Gala worker registration + lifecycle helpers

```go
type DurablePayload struct {
	ID string `json:"id"`
}

var durableTopic = gala.Topic[DurablePayload]{
	Name: gala.TopicName("example.durable"),
}

riverRuntime, err := gala.NewRiverRuntime(ctx, gala.RiverRuntimeOptions{
	ConnectionURI:  jobQueueURI,
	QueueName:      "events", // defaults to gala.DefaultQueueName
	WorkerCount:    10,
	MaxRetries:     5,
	WorkersEnabled: true,
	ConfigureRuntime: func(runtime *gala.Runtime) error {
		// app-specific DI + listener registration
		gala.ProvideValue(runtime.Injector(), dbClient)

		_, err := gala.RegisterDurableListeners(
			runtime.Registry(),
			gala.QueueClassWorkflow,
			gala.Definition[DurablePayload]{
				Topic: durableTopic,
				Name:  "example.durable.listener",
				Handle: func(gala.HandlerContext, DurablePayload) error {
					return nil
				},
			},
		)
		return err
	},
})
if err != nil {
	panic(err)
}
defer riverRuntime.Close()

runtime := riverRuntime.Runtime()

if err := riverRuntime.StartWorkers(ctx); err != nil {
	panic(err)
}
// stop with your app shutdown context
```

### Advanced manual wiring

`RiverDispatcher` implements `DurableDispatcher` and enqueues `RiverDispatchArgs` jobs when you need low-level control:

```go
dispatcher, err := gala.NewRiverDispatcher(gala.RiverDispatcherOptions{
	JobClient: riverClient, // must support Insert; InsertTx is optional
	QueueByClass: map[gala.QueueClass]string{
		gala.QueueClassWorkflow:    "workflow",
		gala.QueueClassIntegration: "integration",
		gala.QueueClassGeneral:     "general",
	},
	DefaultQueue: gala.DefaultQueueName,
})
if err != nil {
	panic(err)
}

runtime, err := gala.NewRuntime(gala.RuntimeOptions{
	DurableDispatcher: dispatcher,
})
if err != nil {
	panic(err)
}
```

For transactional enqueue, place `pgx.Tx` in context:

```go
ctxWithTx := gala.WithPGXTx(ctx, tx)
receipt := runtime.Emit(ctxWithTx, topicName, payload)
```

If `JobClient` also implements `InsertTx`, Gala uses transactional insert automatically.

### Manual River worker wiring

```go
workers := river.NewWorkers()
if err := gala.AddRiverDispatchWorker(workers, func() *gala.Runtime {
	return runtime
}); err != nil {
	panic(err)
}
```

`RiverDispatchWorker` decodes the envelope and calls `runtime.DispatchEnvelope(...)`.

## Pre-built envelopes

Use `EmitEnvelope` when the caller already has a fully built envelope (for example, migration adapters or outbox replay):

```go
err := runtime.EmitEnvelope(ctx, gala.Envelope{
	ID:            gala.EventID("evt_prebuilt_123"),
	Topic:         userCreatedTopic.Name,
	SchemaVersion: userCreatedTopic.EffectiveSchemaVersion(),
	Payload:       encodedPayload,
	Headers: gala.Headers{
		IdempotencyKey: "evt_prebuilt_123",
	},
	ContextSnapshot: snapshot,
})
```

`EmitEnvelope` preserves the supplied envelope ID.

## Error behavior and semantics

- Topic/listener registration validates required fields and duplicate topic names.
- Emitting to an unregistered topic fails (`ErrTopicNotRegistered`).
- Payload type mismatch is detected before encode/decode (`ErrPayloadTypeMismatch`).
- Inline dispatch validates envelope topic/payload and returns `ErrListenerExecutionFailed` on listener failure.
- Durable mode without a dispatcher fails (`ErrDurableDispatcherRequired`).
- Dual mode returns `ErrDualDispatchFailed` when either durable or inline path fails.
- Context snapshot capture/restore failures are surfaced on emit/dispatch.
- Gala-owned runtime/registry validation and dispatch paths return static sentinel errors from `errors.go`.
- Integration/library calls (for example River internals or DI providers) may still surface non-Gala errors.

## API reference

### Runtime

- Build: `NewRuntime(RuntimeOptions)`
- River-backed build: `NewRiverRuntime(ctx, RiverRuntimeOptions)`
- Emit: `Emit`, `EmitWithHeaders`, `EmitTyped`, `EmitEnvelope`
- Direct dispatch: `DispatchEnvelope`
- Accessors: `Registry`, `Injector`, `ContextManager`, `DurableDispatcher`

### Registry and listeners

- Topic registration: `Registration[T].Register`, `RegisterTopic`
- Listener registration: `Definition[T].Register`, `AttachListener`, `RegisterListener`, `RegisterListeners`, `RegisterDurableListeners`
- Introspection/processing: `TopicPolicy`, `EncodePayload(topic, payload)`, `DecodePayload(topic, payload)`, `Listeners`

### Context

- Manager: `NewContextManager`, `ContextManager.Register`, `Capture`, `Restore`
- Typed codec: `NewTypedContextCodec[T](ContextKey)`
- Auth codec: `NewAuthContextCodec`
- Flags: `WithFlag`, `HasFlag`

### Durable (River)

- Runtime lifecycle: `RiverRuntime.Runtime`, `RiverRuntime.JobClient`, `StartWorkers`, `StopWorkers`, `Close`
- Dispatcher: `NewRiverDispatcher`, `DispatchDurable`
- Worker: `NewRiverDispatchWorker`, `AddRiverDispatchWorker`
- Job args: `NewRiverDispatchArgs`, `RiverDispatchArgs.DecodeEnvelope`
- Transactional context helpers: `WithPGXTx`, `PGXTxFromContext`

## Practical recommendations

- Keep topic names stable and schema versions explicit for long-lived contracts.
- Always set listener names to stable identifiers (`service.feature.action` style).
- Use `Headers.IdempotencyKey` for replay-safe consumers.
- Prefer `EmitTyped` for compile-time payload safety.
- Use `EmitEnvelope` for adapter paths that need deterministic event IDs.
- For durable workloads, prefer `NewRiverRuntime` so Gala setup/lifecycle stays inside the package.
