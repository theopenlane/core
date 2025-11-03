package hooks

import (
	"context"

	"entgo.io/ent"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// Eventer is a wrapper struct for having a soiree as well as a list of listeners
type Eventer struct {
	Emitter *soiree.EventPool

	listenerBindings   []soiree.ListenerBinding
	propertyExtractors map[string][]PropertyExtractor
}

// PropertyExtractor allows callers to augment event properties produced by the hook.
type PropertyExtractor func(context.Context, ent.Mutation) soiree.Properties

const globalPropertyScope = "*"

// EventerOpts is a functional options wrapper
type EventerOpts func(*Eventer)

// NewEventer creates a new Eventer with the provided options
func NewEventer(opts ...EventerOpts) *Eventer {
	e := &Eventer{
		propertyExtractors: make(map[string][]PropertyExtractor),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// WithEventerEmitter sets the emitter for the Eventer if there's an existing soiree pool that needs to be passed in
func WithEventerEmitter(emitter *soiree.EventPool) EventerOpts {
	return func(e *Eventer) {
		e.Emitter = emitter
	}
}

// WithEventerListenerBindings allows callers to provide pre-built listener bindings that will be registered alongside the defaults.
func WithEventerListenerBindings(bindings ...soiree.ListenerBinding) EventerOpts {
	return func(e *Eventer) {
		e.listenerBindings = append(e.listenerBindings, bindings...)
	}
}

// WithEventerPropertyExtractor registers a property extractor for a specific topic.
func WithEventerPropertyExtractor(topic soiree.TypedTopic[soiree.Event], extractor PropertyExtractor) EventerOpts {
	return func(e *Eventer) {
		if extractor == nil {
			return
		}

		e.propertyExtractors[topic.Name()] = append(e.propertyExtractors[topic.Name()], extractor)
	}
}

// WithEventerGlobalPropertyExtractor registers a property extractor that runs for every topic.
func WithEventerGlobalPropertyExtractor(extractor PropertyExtractor) EventerOpts {
	return func(e *Eventer) {
		if extractor == nil {
			return
		}

		e.propertyExtractors[globalPropertyScope] = append(e.propertyExtractors[globalPropertyScope], extractor)
	}
}

func (e *Eventer) applyPropertyExtractors(ctx context.Context, topic string, mutation ent.Mutation, target soiree.Properties) {
	if e == nil || target == nil {
		return
	}

	run := func(extractors []PropertyExtractor) {
		for _, extractor := range extractors {
			if extractor == nil {
				continue
			}

			props := extractor(ctx, mutation)
			for key, value := range props {
				target.Set(key, value)
			}
		}
	}

	run(e.propertyExtractors[globalPropertyScope])
	run(e.propertyExtractors[topic])
}

func mutationFieldExtractor(_ context.Context, mutation ent.Mutation) soiree.Properties {
	props := soiree.NewProperties()

	for _, field := range mutation.Fields() {
		value, exists := mutation.Field(field)
		if exists {
			props.Set(field, value)
		}
	}

	return props
}

// NewEventerPool initializes a new Eventer and takes a client to be used as the client for the soiree pool
func NewEventerPool(client interface{}) *Eventer {
	pool := soiree.NewEventPool(
		soiree.WithPool(
			soiree.NewPondPool(
				soiree.WithMaxWorkers(100), // nolint:mnd
				soiree.WithName("ent_event_pool"))),
		soiree.WithClient(client))

	return NewEventer(
		WithEventerEmitter(pool),
		WithEventerGlobalPropertyExtractor(mutationFieldExtractor),
		WithEventerPropertyExtractor(soiree.MutationTopic(entgen.TypeOrganization, ent.OpCreate.String()), organizationCreatePropertyExtractor),
		WithEventerPropertyExtractor(soiree.MutationTopic(entgen.TypeOrganization, ent.OpDelete.String()), organizationDeletePropertyExtractor),
		WithEventerPropertyExtractor(soiree.MutationTopic(entgen.TypeOrganization, ent.OpDeleteOne.String()), organizationDeletePropertyExtractor),
		WithEventerPropertyExtractor(soiree.MutationTopic(entgen.TypeOrganization, SoftDeleteOne), organizationDeletePropertyExtractor),
		WithEventerPropertyExtractor(soiree.MutationTopic(entgen.TypeOrganizationSetting, ent.OpUpdate.String()), organizationSettingPropertyExtractor),
		WithEventerPropertyExtractor(soiree.MutationTopic(entgen.TypeOrganizationSetting, ent.OpUpdateOne.String()), organizationSettingPropertyExtractor),
	)
}
