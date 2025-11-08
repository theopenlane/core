package hooks

import (
	"errors"
	"fmt"

	"entgo.io/ent"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/events/soiree"
)

var errMutationPayloadUnavailable = errors.New("soiree: mutation payload unavailable for topic")

const eventerPoolWorkers = 100

// Eventer coordinates the mutation listeners that will be registered against the ent client and
// underpins the hook emission predicate
type Eventer struct {
	Emitter   *soiree.EventPool
	listeners map[string][]soiree.ListenerBinding
}

// EventerOpts configures an Eventer instance via the functional-options pattern
type EventerOpts func(*Eventer)

// NewEventer constructs an Eventer and applies the provided option set; callers typically use this
// when they have an existing event pool that needs to be reused
func NewEventer(opts ...EventerOpts) *Eventer {
	// listeners is keyed by ent entity name so emission decisions can be made without maintaining
	// a separate allowlist. Mutations that have nothing registered against them simply never fire
	// events.
	e := &Eventer{listeners: make(map[string][]soiree.ListenerBinding)}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// WithEventerEmitter injects an existing soiree.EventPool into an Eventer
func WithEventerEmitter(emitter *soiree.EventPool) EventerOpts {
	return func(e *Eventer) {
		e.Emitter = emitter
	}
}

// MutationHandler is the signature listener implementations expose for mutation events
type MutationHandler func(*soiree.EventContext, *MutationPayload) error

// MutationPayload carries the raw ent mutation, the resolved operation, the entity ID and the ent
// client so listeners can act without additional lookups
type MutationPayload struct {
	Mutation  ent.Mutation
	Operation string
	EntityID  string
	Client    *entgen.Client
}

func mutationTopic(entity string) soiree.TypedTopic[*MutationPayload] {
	// mutationTopic builds a typed topic for the supplied entity so listeners receive strongly typed payloads
	// Ensure every entity shares the same wrapping/unwrapping logic so listeners can rely on a
	// strongly typed payload instead of re-parsing the soiree.Event in every handler
	return soiree.NewTypedTopic(
		entity,
		func(payload *MutationPayload) soiree.Event { return soiree.NewBaseEvent(entity, payload) },
		func(event soiree.Event) (*MutationPayload, error) {
			payload, ok := event.Payload().(*MutationPayload)
			if !ok {
				return nil, fmt.Errorf("%w: %s", errMutationPayloadUnavailable, entity)
			}

			return payload, nil
		},
	)
}

// AddMutationListener registers a handler for the supplied entity and records any listener
// options; registration automatically opts the entity into event emission
func (e *Eventer) AddMutationListener(entity string, handler MutationHandler, opts ...soiree.ListenerOption) {
	if e == nil || handler == nil || entity == "" {
		return
	}

	if e.listeners == nil {
		// The zero-value Eventer can be embedded into other structs; lazily recreate the map so
		// calls remain safe after JSON/YAML unmarshalling or tests that bypass NewEventer.
		e.listeners = make(map[string][]soiree.ListenerBinding)
	}

	bound := soiree.BindListener(
		mutationTopic(entity),
		func(ctx *soiree.EventContext, payload *MutationPayload) error {
			return handler(ctx, payload)
		},
		opts...,
	)

	e.listeners[entity] = append(e.listeners[entity], bound)
}

// NewEventerPool builds a fresh event pool, associates it with an Eventer, and wires the default
// mutation listeners
func NewEventerPool(client any) *Eventer {
	pool := soiree.NewEventPool(
		soiree.WithPool(
			soiree.NewPondPool(
				soiree.WithMaxWorkers(eventerPoolWorkers),
				soiree.WithName("ent_event_pool"))),
		soiree.WithClient(client))

	eventer := NewEventer(
		WithEventerEmitter(pool),
	)

	registerDefaultMutationListeners(eventer)

	return eventer
}

// registerDefaultMutationListeners wires the listeners we ship by default
func registerDefaultMutationListeners(e *Eventer) {
	if e == nil {
		return
	}

	e.AddMutationListener(entgen.TypeOrganization, handleOrganizationMutation)
	e.AddMutationListener(entgen.TypeOrganizationSetting, handleOrganizationSettingMutation)
	e.AddMutationListener(entgen.TypeSubscriber, handleSubscriberMutation)
	e.AddMutationListener(entgen.TypeUser, handleUserMutation)
}
