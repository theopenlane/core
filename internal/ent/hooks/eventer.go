package hooks

import (
	"fmt"

	"entgo.io/ent"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/events/soiree"
)

type Eventer struct {
	Emitter   *soiree.EventPool
	listeners []mutationListener
	entities  map[string]struct{}
}

type mutationListener struct {
	entity  string
	binding soiree.ListenerBinding
}

type EventerOpts func(*Eventer)

func NewEventer(opts ...EventerOpts) *Eventer {
	e := &Eventer{entities: make(map[string]struct{})}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func WithEventerEmitter(emitter *soiree.EventPool) EventerOpts {
	return func(e *Eventer) {
		e.Emitter = emitter
	}
}

type MutationHandler func(*soiree.EventContext, *MutationPayload) error

type MutationPayload struct {
	Mutation  ent.Mutation
	Operation string
	EntityID  string
	Client    *entgen.Client
}

func mutationTopic(entity string) soiree.TypedTopic[*MutationPayload] {
	return soiree.NewTypedTopic(
		entity,
		func(payload *MutationPayload) soiree.Event { return soiree.NewBaseEvent(entity, payload) },
		func(event soiree.Event) (*MutationPayload, error) {
			payload, ok := event.Payload().(*MutationPayload)
			if !ok {
				return nil, fmt.Errorf("soiree: mutation payload unavailable for topic %s", entity)
			}

			return payload, nil
		},
	)
}

func (e *Eventer) AddMutationListener(entity string, handler MutationHandler, opts ...soiree.ListenerOption) {
	if e == nil || handler == nil || entity == "" {
		return
	}

	bound := soiree.BindListener(
		mutationTopic(entity),
		func(ctx *soiree.EventContext, payload *MutationPayload) error {
			return handler(ctx, payload)
		},
		opts...,
	)

	e.listeners = append(e.listeners, mutationListener{entity: entity, binding: bound})
	e.entities[entity] = struct{}{}
}

func NewEventerPool(client interface{}) *Eventer {
	pool := soiree.NewEventPool(
		soiree.WithPool(
			soiree.NewPondPool(
				soiree.WithMaxWorkers(100), // nolint:mnd
				soiree.WithName("ent_event_pool"))),
		soiree.WithClient(client))

	eventer := NewEventer(
		WithEventerEmitter(pool),
	)

	registerDefaultMutationListeners(eventer)

	return eventer
}

func registerDefaultMutationListeners(e *Eventer) {
	if e == nil {
		return
	}

	e.AddMutationListener(entgen.TypeOrganization, handleOrganizationMutation)
	e.AddMutationListener(entgen.TypeOrganizationSetting, handleOrganizationSettingMutation)
	e.AddMutationListener(entgen.TypeSubscriber, handleSubscriberMutation)
	e.AddMutationListener(entgen.TypeUser, handleUserMutation)
}
