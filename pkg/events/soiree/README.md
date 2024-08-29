# Soiree

Soiree, a fancy event affair, or, event - is a library indendied to simplify event management inside a golang codebase but more generically is a 2-tier channel system, one for queuing jobs and another to control how many workers operate on that job queue concurrently. The goal is a dead-simple interface for event subscription and handling, using [pond](https://github.com/alitto/pond) for performance management, goroutine pooling, and wrapping an event management interface with thread-safe interactions.

## Overview

Functionally, Soiree is intended to provide:

- **In-Memory management**: Host and manage events internally without external dependencies or libraries
- **Listener prioritization**: Controls for invocation order
- **Concurrent Processing**: Utilize pooled goroutines for handling events in parallel and re-releasing resources, with thread safety
- **Configurable Subscriptions**: Leverages basic pattern matching for event subscriptions
- **General Re-use**: Configure with custom handlers for errors, IDs, and panics panics

### But why?

In modern software architectures, microservices vs. monoliths is a false dichotomy; the optimum is usually somewhere in the middle. We're pretty intentionally sticking with a "monolith" in the sense we are producing a single docker image from this codebase, but the often overlooked aspect of these architectures is the context under which the service is started, and run. If you assume that the connectivity from your client is created in a homogeneous fashion, ex:
```
┌──────────────┐        .───────.         ┌─────────────────┐
│              │       ╱         ╲        │                 │
│    Client    │──────▶   proxy   ────────▶     Service     │
│              │       `.       ,'        │                 │
└──────────────┘         `─────'          └─────────────────┘
```
Then all instances of the service will be required to perform things like authorizations validation, session issuance, etc. The validity of these actions is managed with external state machines, such as Redis.
```
                                                                   ┌────────────────┐
┌──────────────┐        .───────.         ┌─────────────────┐      │                │
│              │       ╱         ╲        │                 │      │     Redis,     │
│    Client    │──────▶   proxy   ────────▶    Service      ├─────▶│   PostgreSQL   │
│              │       `.       ,'        │                 │      │                │
└──────────────┘         `─────'          └─────────────────┘      └────────────────┘
```
We do this because we want to be able to run many instances of the service, for things such as canary, rollouts, etc.
```
                                          ┌─────────────────┐
                                          │                 │
                                     ┌───▶│    Service      ├──┐
                                     │    │                 │  │
                                     │    └─────────────────┘  │
                                     │                         │   ┌────────────────┐
┌──────────────┐        .───────.    │    ┌─────────────────┐  │   │                │
│              │       ╱         ╲   │    │                 │  │   │     Redis,     │
│    Client    │──────▶   proxy   ───┼────▶    Service      ├──┴┬─▶▶   PostgreSQL   │
│              │       `.       ,'   │    │                 │   │  │                │
└──────────────┘         `─────'     │    └─────────────────┘   │  └────────────────┘
                                     │                          │
                                     │    ┌─────────────────┐   │
                                     │    │                 │   │
                                     └───▶│    Service      │───┘
                                          │                 │
                                          └─────────────────┘
```
Now, where things start to get more fun is when you layer in the desire to perform I/O operations either managed by us, or externally (e.g. S3), as well as connect to external data stores (e.g. Turso).
```
                                             ┌──────────────┐
                                             │              │
                                             │      S3      │
                                             │              │           ┌───────────────┐
                                             └───────▲──────┘    ┌─────▶│ Outbound HTTP │
                                                     │           │      │(e.g. webhooks)│
                                                     │           │      └───────────────┘
                                                     │           │
                                                   ┌─┘           │
                                                   │             │
                   ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┼ ─ ─ ─ ─ ─ ─ ┼ ─ ─ ─ ┬─────────────┐
                                                   │             │       │ Stuff under │
                   │                               │             │       │ our control │
                                          ┌────────┴────────┬────┘       └─────────────┘
                   │                      │                 │                          │
                                     ┌───▶│    Service      ├──┐
                   │                 │    │                 │  │                       │
                                     │    └─────────────────┘  │
                   │                 │                         │   ┌────────────────┐  │
┌──────────────┐        .───────.    │    ┌─────────────────┐  │   │                │
│              │   │   ╱         ╲   │    │                 │  │   │     Redis,     │  │
│    Client    │──────▶   proxy   ───┼────▶    Service      ├──┴┬─▶▶   PostgreSQL   │
│              │   │   `.       ,'   │    │                 │   │  │                │  │
└──────────────┘         `─────'     │    └─────────────────┘   │  └────────────────┘
                   │                 │                          │                      │
                                     │    ┌─────────────────┐   │
                   │                 │    │                 │   │                      │
                                     └───▶│    Service      │───┘
                   │                      │                 │                          │
                                          └────────┬────────┴─────────────┐
                   │                               │                      │            │
                    ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┼ ─ ─ ─ ─ ─ ─
                                                   │                      │
                                                   └─────┐                │
                                                         │                │
                                                         │                │               ┌──────────────┐
                                                         │                │               │ Other future │
                                                         ▼                └──────────────▶│  ridiculous  │
                                                     ┌───────────┐                        │    ideas     │
                                                     │   Turso   │                        └──────────────┘
                                                     └───────────┘
```
Given we need to be able to perform all kinds of workload actions such as writing a file to a bucket, committing a SQL transaction, sending an http request, we need bounded patterns and degrees of resource control. Which is to say, we need to control resource contention in our runtime as we don't want someone's regular HTTP request to be blocked by the fact someone else requested a bulk upload to S3. This means creating rough groupings of workload `types` and bounding them so that you can monitor and control the behaviors and lumpiness of the variances with the workload types.

Check out [this blog](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/) (there are many on this topic) for some real world examples on how systems with these "lumpy" workload types can become easily bottlenecked with volume. Since we are intending to open the flood gates around event ingestion from other sources (similar to how Posthog, Segment, etc., work) we need to anticipate a very high load of unstructured data which needs to be written efficiently to a myriad of external sources, as well as bulk routines which may be long running such as file imports, uploads, exports, etc.

## How many goroutines can / should I have?

A single go-routine currently uses a minimum stack size of 2KB. It is likely that your actual code will also allocate some additional memory on the heap (for e.g. JSON serialization or similar) for each goroutine. This would mean 1M go-routines could easily require 2-4 GB or RAM (should be ok for an average environment)

Most OS will limit the number of open connections in various ways. For TCP/IP there is usually a limit of open ports per interface. This is about 28K on many modern systems. There is usually an additional limit per process (e.g. ulimit for number of open file-descriptors) which will by default be around 1000. So without changing the OS configuration you will have a maximum of 1000 concurrent connections on Linux.

So depending on the system you should probably not create more than 1000 goroutines, because they might start failing with “maximum num of file descriptors reached” or even drop packets. If you increase the limits you are still bound by the 28K connections from a single IP address.

## Quick Start

```go
package main

import (
	"fmt"
	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	e := soiree.NewEventPool()
	e.On("user.created", func(evt soiree.Event) error {
		fmt.Println("Event received:", evt.Topic())
		return nil
	})
	e.Emit("user.created", "Matty Ice")
}
```

## Configuration

Your Soiree can come with a few options if you wish:

```go
e := soiree.NewEventPool(
	soiree.WithErrorHandler(customErrorHandler),
	soiree.WithIDGenerator(customIDGenerator),
)
```

## Subscribing to events using basic pattern matching

Per guidance of many pubsubs such as Kafka, operating a multi-tenant cluster typically requires you to define user spaces for each tenant. For the purpose of this section, "user spaces" are a collection of topics, which are grouped together under the management of a single entity or user.

In Kafka and many other pubsub systems, the main unit of data is the `topic`. Users can create and name each topic. They can also delete them, but it is not possible to rename a topic directly. Instead, to rename a topic, the user must create a new topic, move the messages from the original topic to the new, and then delete the original. With this in mind, it is recommended to define logical spaces, based on an hierarchical topic naming structure. This setup can then be combined with security features, such as prefixed ACLs, to isolate different spaces and tenants, while also minimizing the administrative overhead for securing the data in the cluster.

These logical user spaces can be grouped in different ways, by team or organizational unit: here, the `organization` is the main aggregator.

Example topic naming structure:

<org_id>.<user_id>.<object>.<event-name>
(e.g., ULID.ULID.user.login, or "ULID.ULID.organization.created")
By organization or product: their credentials will be different for each organization, so all the controls and settings will always be organization related which is a good high level topic for performing  a broad set of actions, but you can also be more granular.

Example topic naming structure:

<organization>.<product>.<event-name>
(e.g., "openlane.invoices.received")
Certain information should normally not be put in a topic name, such as information that is likely to change over time (e.g., the name of the intended consumer) or that is a technical detail or metadata that is available elsewhere (e.g., the topic's partition count and other configuration settings).

### Kafka specifics

To enforce a topic naming structure in Kafka, several options are available:

Use prefix ACLs (cf. KIP-290) to enforce a common prefix for topic names. For example, team A may only be permitted to create topics whose names start with payments.teamA..
Define a custom CreateTopicPolicy (cf. KIP-108 and the setting create.topic.policy.class.name) to enforce strict naming patterns. These policies provide the most flexibility and can cover complex patterns and rules to match an organization's needs.
Disable topic creation for normal users by denying it with an ACL, and then rely on an external process to create topics on behalf of users (e.g., scripting or your favorite automation toolkit).
It may also be useful to disable the Kafka feature to auto-create topics on demand by setting auto.create.topics.enable=false in the broker configuration. Note that you should not rely solely on this option.


### Soiree topic matching

- `*` - Matches a single segment
- `**` - Matches multiple segments

### Example:

```go
e := soiree.NewEventPool()
e.On("user.*", userEventListener)
e.On("invoice.**", orderEventListener)
e.On("**.completed", completionEventListener)
```

or:

```go
e := soiree.NewEventPool()
e.On("user.*", func(evt soiree.Event) error {
	fmt.Printf("Event: %s, Payload: %+v\n", evt.Topic(), evt.Payload())
	return nil
})
e.Emit("user.signup", "Funky Sarah")
```

## Aborting Event Propagation

Stop event propagation using `SetAborted`:

```go
e := soiree.NewEventPool()
e.On("invoice.processed", func(evt soiree.Event) error {
	if /* condition fails */ false {
		evt.SetAborted(true)
	}
	return nil
}, soiree.WithPriority(soiree.High))
e.On("invoice.processed", func(evt soiree.Event) error {
	// This will not run if the event is aborted
	return nil
}, soiree.WithPriority(soiree.Low))
e.Emit("invoice.processed", "Order data")
```

## Examples

### Concurrency

Delegate concurrency management to a goroutine pool using the `WithPool` option:

```go
package main

import (
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/alitto/pond"
)

func main() {
	// Initialize a goroutine pool
	pool := soiree.NewPondPool(10, 1000) // 10 workers, queue size 1000

	// Start your soiree and invite your friends (add your pool :))
	e := soiree.NewEventPool(soiree.WithPool(pool))

	// Your soiree is now ready to handle whatever events (or in general, processing) using the pool
}
```

### Error Handling

Change your error handling depending on your needs by passing in a custom error handler:

```go
package main

import (
	"log"
	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	// Define a custom error handler that logs the event and the error
	customErrorHandler := func(event soiree.Event, err error) error {
		log.Printf("Error encountered during event '%s': %v, with payload: %v", event.Topic(), err, event.Payload())
		return nil  // Returning nil to indicate that the error has been swallowed
	}

	// Apply the custom error handler to the soiree
	e := soiree.NewEventPool(soiree.WithErrorHandler(customErrorHandler))

	// Your soiree will now log errors encountered during event handling
}
```

### Prioritizing Listeners

Control the invocation order of event listeners by prescribing priorities:

```go
package main

import (
	"fmt"
	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	// Set up the swanky soiree
	e := soiree.NewEventPool()

	// Define listeners with varying priorities
	normalPriorityListener := func(e soiree.Event) error {
		fmt.Println("Normal priority: Received", e.Topic())
		return nil
	}

	highPriorityListener := func(e soiree.Event) error {
		fmt.Println("High priority: Received", e.Topic())
		return nil
	}

	// Subscribe listeners with specified priorities
	e.On("user.created", normalPriorityListener) // Default is normal priority
	e.On("user.created", highPriorityListener, soiree.WithPriority(soiree.High))

	// Emit an event and observe the order of listener notification
	e.Emit("user.created", "User signup event")
}
```

Listeners with higher priority are notified first when an event occurs - *NOTE* while the listeners may be _notified_ first, there's no guarantees beyond the notification, meaning, if the Listener which receives the high priority event has a long-running action, and the lower priority listener has a quick action, which one finishes first depends on the action(s) themselves. If you need _blocking_ logic (not ideal), use synchronous event handlers.

### Generating Unique IDs

Implement custom ID generation for listener tracking:

```go
package main

import (
	"github.com/google/uuid"
	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	// Custom ID generator using UUID v4
	uuidGenerator := func() string {
		return uuid.NewString()
	}

	// Initialize the soiree with the UUID generator
	e := soiree.NewEventPool(soiree.WithIDGenerator(uuidGenerator))

	// Listeners will now be registered with a UUID
}
```

Listeners are now identified by a UUID vs. the standard ULID generated by Openlane. You can also create listeners with a similar naming convention as the topic so you can identify them more easily. Generally speaking, though: listeners are acting on the events immediately. This is an in-memory implementation so the most appropriate use of the goroutine pools are either writing to a persistent store immediately when the event occurs to perform other future actions if required, handling an event immediately based on a criteria (e.g. webhooks, outbound http methods) or other idempotent actions.

### Handling Panics

Safeguard the overall runtime from unexpected panics during event handling by using a panic handler:

```go
package main

import (
	"log"
	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	// Define a panic handler that logs the occurrence
	logPanicHandler := func(p interface{}) {
		log.Printf("Panic recovered: %v", p)
		// Insert additional logic for panic recovery here
	}

	// Equip the soiree with the panic handler
	e := soiree.NewEventPool(soiree.WithPanicHandler(logPanicHandler))

	// Your soiree is now more resilient to panics!
}
```

This handler ensures that panics are logged and managed without creating a service interruption.

