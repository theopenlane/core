# Gala

`gala` is a "v2" of the original `pkg/soiree` that was an in-memory event emitter + listener definition library that wrapped the `alitto/pond` library.

Known deficiencies / rationale:

- Current Soiree path is in-memory by default and not sufficient for durable multi-pod processing.
- Redis-backed Soiree durability is not viable for our current contracts.
- Workflow engine depends on event semantics but is not yet broadly active; this is the best point to establish final-state architecture.
- Existing listeners remain valuable and should be migrated topic-by-topic with controlled risk.
- We can dual-emit during migration (`soiree` + `gala`) while keeping ownership clear per topic.

## Core Goals & Directives

Applying lessons learned with Soiree:

- Listener registration....defining a listener should be enough to make registration straightforward without repeated wrappers
- The new package should not require call sites to manually drain channels for enqueue outcomes
- Handlers / Listeners should receive typed dependency accessors to avoid repetitive pointer extraction and type assertions
- We need consistent durability and patterns...use River per-queue concurrency and retries rather than bespoke goroutine orchestration
- Data access...prefer typed enum parsing helpers over local switch/case parsers
- Keep context restore centralized; avoid per-listener ad-hoc context mutation
- First-class context reconstruction (auth + context flags + operational metadata) is priority

## Rough flow

Emit(Topic: "Organization", Op: "Create")
    ──► River Worker
        ──► Find listener for (Organization, Create)
            ──► Execute entitlements.organization.create handler (in-tree execution is performed by river threads)

## Concurrency notes

With WorkerCount: 100, there's up to 100 events being processed concurrently, each with its own sequential listener
chain (if you actually chain them, which we don't really). This scales horizontally - if we need more throughput, increase worker count or add more instances.

The sequential-per-event design gives us:
- Predictable ordering within a single event
- Simple error/retry semantics per event
- No intra-event race conditions

River handles:
- Concurrent processing across events
- Durable job storage
- Retry with backoff on failure
- Multi-instance scaling (multiple server processes can run workers)

## Accessing core runtime dependencies in listeners

```go

do.ProvideValue(runtime.Injector(), dbClient)                    // *generated.Client
do.ProvideValue(runtime.Injector(), dbClient.EntitlementManager) // you could access via the dbclient -> *entitlements.StripeClient
do.ProvideNamedValue(runtime.Injector(), "stripe_client", stripeClient) // you can also access the client directly
do.ProvideValue(runtime.Injector(), dbClient.TokenManager)       // *tokens.TokenManager
do.ProvideValue(runtime.Injector(), wfEngine)                    // *engine.WorkflowEngine

// listener
func handle(ctx gala.HandlerContext, payload MyPayload) error {
	// now you can profit from the injected type
	stripeClient, err := do.InvokeNamed[*stripe.Client](ctx.Injector, "stripe_client")
	if err != nil {
		return err
	}

	client, err := do.Invoke[*generated.Client](ctx.Injector)
	if err != nil {
		return err
	}

	return nil
}
```

## Rollout / Execution plan

- Develop new package addressing the goals / directives
- Confirm event emission, listener registration, etc.
- Double run both packages during stability phase
- Convert non-critical path listeners (e.g. sending slack messages or similar) to gala and run in production + monitor
- Migrate initial listener tranche to Gala (workflow mutation + assignment, entitlements org + org setting).
- Migrate remaining legacy listeners topic-by-topic
- Remove legacy emission for migrated topics / deprecate soiree

## Representative Example

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

// define a stable topic name and your "envelope" which is your JSON data payload
var userCreatedTopic = gala.Topic[UserCreated]{
	Name: gala.TopicName("user.created"),
}

func main() {
	// Initialize Gala application and configure worker pool
	app, err := gala.NewGala(context.Background(), gala.Config{
		ConnectionURI: jobQueueURI,
		QueueName:     gala.DefaultQueueName,
		WorkerCount:   10,
		MaxRetries:    5,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	// Start Gala workers
	if err := app.StartWorkers(context.Background()); err != nil {
		log.Fatal(err)
	}

	// Register event topic and payload codec
	runtime := app.Runtime()
	err = gala.RegisterTopic(runtime.Registry(), gala.Registration[UserCreated]{
		Topic: userCreatedTopic,
		Codec: gala.JSONCodec[UserCreated]{},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Attach event listener (handler) for the topic
	_, err = gala.AttachListener(runtime.Registry(), gala.Definition[UserCreated]{
		Topic: userCreatedTopic,
		Name:  "welcome-email",
		Handle: func(ctx gala.HandlerContext, payload UserCreated) error {
			log.Printf("send welcome email to %s (%s)", payload.UserID, payload.Email)
			return nil
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Emit event to the topic (triggers durable dispatch)
	receipt := runtime.EmitWithHeaders(context.Background(), userCreatedTopic.Name, UserCreated{
		UserID: "usr_123",
		Email:  "user@example.com",
	}, gala.Headers{})
	if receipt.Err != nil {
		log.Fatal(receipt.Err)
	}
	log.Printf("event accepted: id=%s", receipt.EventID)
}
```
## Basic benchmarks

┌──────────────────────────┬────────┬─────────┬────────┬──────────┬───────────────────┬─────────────────────┬───────┐
│         Scenario         │ Events │ Workers │ Topics │ Emitters │ Gala (events/sec) │ Soiree (events/sec) │ Ratio │
├──────────────────────────┼────────┼─────────┼────────┼──────────┼───────────────────┼─────────────────────┼───────┤
│ Sequential, Multi-topic  │ 500    │ 20      │ 5      │ 1        │ 966               │ 414,350             │ 1:429 │
├──────────────────────────┼────────┼─────────┼────────┼──────────┼───────────────────┼─────────────────────┼───────┤
│ Sequential, Single-topic │ 500    │ 20      │ 1      │ 1        │ 915               │ 268,372             │ 1:293 │
├──────────────────────────┼────────┼─────────┼────────┼──────────┼───────────────────┼─────────────────────┼───────┤
│ Concurrent, Multi-topic  │ 1000   │ 50      │ 5      │ 50       │ 2,100             │ 295,610             │ 1:141 │
├──────────────────────────┼────────┼─────────┼────────┼──────────┼───────────────────┼─────────────────────┼───────┤
│ Concurrent, Single-topic │ 1000   │ 50      │ 1      │ 50       │ 3,351             │ 430,339             │ 1:128 │
└──────────────────────────┴────────┴─────────┴────────┴──────────┴───────────────────┴─────────────────────┴───────┘
Observations:

1. Emission pattern matters for Gala: Concurrent emitters (50 goroutines) more than double throughput (966 → 2,100 for
multi-topic, 915 → 3,351 for single-topic) because PostgreSQL INSERTTs parallelize.
1. Topic count has minimal impact on Gala: Single vs multi-topic shows ~5% variance (966 vs 915 sequential). The bottleneck is
River's job fetch cycle, not topic dispatch.
1. Soiree's single-topic variance: Soiree actually performs better with single topic in concurrent scenarios (430k vs 296k).
Less topic lookup overhead, and the pond pool distributes listener work efficiently.
1. Gap narrows with concurrency: The Gala:Soiree ratio improves from 1:429 (sequential) to 1:128 (concurrent) because Gala's parallel INSERT/fetch paths scale with load while soiree's in-memory dispatch has near-zero overhead regardless.
