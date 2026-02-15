# Gala

Gala is a durable event dispatch system built on [River](https://riverqueue.com). It replaces the in-memory `pkg/soiree` with PostgreSQL-backed job persistence, automatic retries, and multi-instance scaling.

Events emitted through Gala survive process restarts, pod evictions, and deployment rollouts. When a listener fails, River retries the job with backoff until it succeeds or exhausts configured attempts.

## How It Works

```mermaid
sequenceDiagram
    participant Hook as Ent Mutation Hook
    participant Gala as Gala Runtime
    participant PG as PostgreSQL (river_job)
    participant Worker as River Worker
    participant Listener as Listener Handler

    Hook->>Gala: enqueueGalaMutation()
    Gala->>Gala: Encode payload + capture context
    Gala->>PG: INSERT river_job (gala_dispatch_v1)
    Note over PG: Job persisted durably

    Worker->>PG: Poll for available jobs
    PG-->>Worker: Dequeue job
    Worker->>Gala: DispatchEnvelope()
    Gala->>Gala: Decode payload + restore context
    Gala->>Listener: handler(ctx, payload)

    alt Listener succeeds
        Worker->>PG: Mark job complete
    else Listener fails
        Worker->>PG: Schedule retry with backoff
    end
```

The entire envelope (payload, headers, context snapshot) is JSON-serialized into a single `river_job` row. This means your event data lives in the same PostgreSQL database as your application data, simplifying operational concerns.

## Defining Topics and Listeners

A topic is a named channel for events of a specific payload type. The generic type parameter enforces compile-time type safety between emitters and listeners.

```go
// Define your payload type
type InvoiceCreated struct {
    InvoiceID  string `json:"invoice_id"`
    CustomerID string `json:"customer_id"`
    Amount     int64  `json:"amount_cents"`
}

// Define the topic with its payload type
var invoiceCreatedTopic = gala.Topic[InvoiceCreated]{
    Name: "billing.invoice.created",
}
```

Register listeners using `RegisterListeners`, which handles topic registration automatically:

```go
func RegisterBillingListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
    return gala.RegisterListeners(registry,
        gala.Definition[InvoiceCreated]{
            Topic: invoiceCreatedTopic,
            Name:  "billing.invoice.send-receipt",
            Handle: func(ctx gala.HandlerContext, payload InvoiceCreated) error {
                // Access dependencies via the injector
                mailer, err := do.Invoke[*email.Client](ctx.Injector)
                if err != nil {
                    return err
                }
                return mailer.SendReceipt(ctx.Context, payload.CustomerID, payload.InvoiceID)
            },
        },
        gala.Definition[InvoiceCreated]{
            Topic: invoiceCreatedTopic,
            Name:  "billing.invoice.update-metrics",
            Handle: func(ctx gala.HandlerContext, payload InvoiceCreated) error {
                metrics.InvoicesCreated.Inc()
                return nil
            },
        },
    )
}
```

### Operation Filtering

For mutation-style payloads with an `Operation` field, listeners can filter by operation type:

```go
gala.Definition[eventqueue.MutationGalaPayload]{
    Topic: gala.Topic[eventqueue.MutationGalaPayload]{
        Name: "Organization",
    },
    Name:       "entitlements.organization.create",
    Operations: []string{ent.OpCreate.String()},  // Only handle creates
    Handle:     handleOrganizationCreated,
}
```

## Emitting Events

For direct emission (outside the mutation hook flow):

```go
receipt := galaApp.EmitWithHeaders(ctx, invoiceCreatedTopic.Name, InvoiceCreated{
    InvoiceID:  "inv_123",
    CustomerID: "cus_456",
    Amount:     9900,
}, gala.Headers{
    IdempotencyKey: "inv_123", // Enables replay-safe consumption
})

if receipt.Err != nil {
    return receipt.Err
}
```

Most events in the codebase flow through the ent mutation hook (`EmitGalaEventHook`), which automatically builds and emits `MutationGalaPayload` envelopes after successful commits.

## Codecs: Type-Safe Serialization

Codecs handle the serialization boundary between your Go types and the JSON stored in River jobs. Every topic registration requires a codec:

```go
type Codec[T any] interface {
    Encode(T) ([]byte, error)
    Decode([]byte) (T, error)
}
```

The built-in `JSONCodec[T]` handles standard JSON marshaling and is the default for most use cases. Custom codecs are useful when you need:

- Non-JSON formats (protobuf, msgpack)
- Encryption at rest for sensitive payloads
- Schema migration or backwards compatibility handling

```go
gala.RegisterTopic(registry, gala.Registration[InvoiceCreated]{
    Topic: invoiceCreatedTopic,
    Codec: gala.JSONCodec[InvoiceCreated]{}, // Default JSON codec
})
```

When using `RegisterListeners`, topics are auto-registered with `JSONCodec`.

## Context Propagation

Gala snapshots context values at emit time and restores them when the listener executes. This happens transparently for authenticated user context (`auth.AuthenticatedUser`), so listeners can access the original caller's identity even when processing asynchronously.

```mermaid
flowchart LR
    subgraph "Emit Time"
        A[HTTP Request Context] --> B[ContextManager.Capture]
        B --> C[ContextSnapshot JSON]
    end

    subgraph "Stored in River Job"
        C --> D[(Envelope.ContextSnapshot)]
    end

    subgraph "Listener Execution"
        D --> E[ContextManager.Restore]
        E --> F[Listener receives original auth context]
    end
```

### Context Flags

For boolean signals that need to propagate (like workflow bypass flags):

```go
// At emit time
ctx = gala.WithFlag(ctx, gala.ContextFlagWorkflowBypass)

// In listener
if gala.HasFlag(ctx.Context, gala.ContextFlagWorkflowBypass) {
    // Skip workflow processing
}
```

## Dependency Injection

Listeners receive a `HandlerContext` with a `do.Injector` for resolving dependencies. Dependencies are registered at startup:

```go
// During server initialization
do.ProvideValue(galaApp.Injector(), dbClient)
do.ProvideValue(galaApp.Injector(), workflowEngine)

// In listener
func handleInvoiceCreated(ctx gala.HandlerContext, payload InvoiceCreated) error {
    client, err := do.Invoke[*ent.Client](ctx.Injector)
    if err != nil {
        return err
    }
    // Use client...
}
```

This avoids global state and makes listeners testable with mock dependencies.

## Durability and Retries

Gala's durability comes from River, which stores jobs in PostgreSQL. Key characteristics:

| Aspect | Behavior |
|--------|----------|
| **Storage** | Jobs stored in `river_job` table alongside application data |
| **Delivery** | At-least-once (use `IdempotencyKey` for exactly-once semantics) |
| **Retries** | Configurable via `Config.MaxRetries`, exponential backoff |
| **Scaling** | Multiple workers poll the same queue; work distributed automatically |
| **Ordering** | No ordering guarantees across events; order preserved within single event's listeners |

When a listener returns an error, the job is rescheduled for retry. Panics are recovered and converted to errors, triggering the same retry behavior.

## Server Integration

The standard setup in `serveropts`:

```go
galaApp, err := gala.NewGala(ctx, gala.Config{
    ConnectionURI: jobQueueConnectionURI,
    QueueName:     "events",     // River queue name
    WorkerCount:   10,           // Concurrent workers polling this queue
    MaxRetries:    5,            // Retry attempts before marking failed
})
if err != nil {
    return err
}

// Register dependencies for listeners
do.ProvideValue(galaApp.Injector(), dbClient)

// Register listeners
hooks.RegisterGalaSlackListeners(galaApp.Registry())

// Start workers
if err := galaApp.StartWorkers(ctx); err != nil {
    return err
}

// On shutdown
defer galaApp.StopWorkers(ctx)
defer galaApp.Close()
```

## Concurrency Model

```
Event A ──► Worker 1 ──► Listener executes
Event B ──► Worker 2 ──► Listener executes
Event C ──► Worker 3 ──► Listener executes
   ...         ...
```

With `WorkerCount: 100`, up to 100 events process concurrently. Each event's listener(s) execute sequentially within their worker, providing:

- Predictable execution order within a single event
- Simple error/retry semantics per event
- No intra-event race conditions

For higher throughput, increase `WorkerCount` or scale horizontally (multiple server instances share the same queue).

## The Envelope

Every emitted event becomes an `Envelope`:

```go
type Envelope struct {
    ID              EventID         // ULID for tracing and idempotency
    Topic           TopicName       // Routing key for listener dispatch
    OccurredAt      time.Time       // Emit timestamp (UTC)
    Headers         Headers         // IdempotencyKey + arbitrary properties
    Payload         json.RawMessage // Codec-encoded payload bytes
    ContextSnapshot ContextSnapshot // Captured auth context + flags
}
```

The entire envelope is JSON-serialized into `RiverDispatchArgs.Envelope` and stored as the job's arguments. On worker pickup, it's deserialized and dispatched to matching listeners.

## Pre-built Envelopes

For migration adapters or replay scenarios where you already have a fully constructed envelope:

```go
err := galaApp.EmitEnvelope(ctx, gala.Envelope{
    ID:              gala.EventID("evt_replay_123"),
    Topic:           invoiceCreatedTopic.Name,
    Payload:         preEncodedPayload,
    Headers:         gala.Headers{IdempotencyKey: "evt_replay_123"},
    ContextSnapshot: capturedSnapshot,
})
```

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Unregistered topic | `EmitWithHeaders` returns error immediately |
| Codec encode failure | `EmitWithHeaders` returns error immediately |
| River insert failure | `EmitWithHeaders` returns error (event not queued) |
| Listener returns error | Job scheduled for retry |
| Listener panics | Recovered, wrapped as error, job scheduled for retry |
| Max retries exhausted | Job marked as discarded in River |

Listener errors are wrapped in `ListenerError` which includes the listener name and whether a panic occurred, useful for debugging and metrics.
