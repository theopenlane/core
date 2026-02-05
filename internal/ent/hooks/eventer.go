package hooks

import (
	"errors"
	"fmt"

	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/notifications"
	"github.com/theopenlane/core/pkg/events/soiree"
)

var errMutationPayloadUnavailable = errors.New("soiree: mutation payload unavailable for topic")

const eventerPoolWorkers = 200

// Eventer coordinates the mutation listeners that will be registered against the ent client and
// underpins the hook emission predicate
type Eventer struct {
	Emitter   *soiree.EventBus
	listeners map[string][]soiree.ListenerBinding
	bindings  []soiree.ListenerBinding
	// workflowListenersEnabled controls registration of workflow listeners/mutation handlers
	workflowListenersEnabled bool
}

// EventerOpts configures an Eventer instance via the functional-options pattern
type EventerOpts func(*Eventer)

// NewEventer constructs an Eventer and applies the provided option set; callers typically use this
// when they have an existing event bus that needs to be reused
func NewEventer(opts ...EventerOpts) *Eventer {
	e := &Eventer{
		listeners:                make(map[string][]soiree.ListenerBinding),
		workflowListenersEnabled: true,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// WithEventerEmitter injects an existing soiree.EventBus into an Eventer
func WithEventerEmitter(emitter *soiree.EventBus) EventerOpts {
	return func(e *Eventer) {
		e.Emitter = emitter
	}
}

// WithWorkflowListenersEnabled toggles workflow listener registration
func WithWorkflowListenersEnabled(enabled bool) EventerOpts {
	return func(e *Eventer) {
		e.workflowListenersEnabled = enabled
	}
}

// MutationHandler is the signature listener implementations expose for mutation events
type MutationHandler func(*soiree.EventContext, *events.MutationPayload) error

// mutationTopic constructs a typed topic for the supplied entity name
func mutationTopic(entity string) soiree.TypedTopic[*events.MutationPayload] {
	return soiree.NewTypedTopic(
		entity,
		soiree.WithUnwrap(func(event soiree.Event) (*events.MutationPayload, error) {
			payload, ok := event.Payload().(*events.MutationPayload)
			if !ok {
				return nil, fmt.Errorf("%w: %s", errMutationPayloadUnavailable, entity)
			}

			return payload, nil
		}),
	)
}

// AddMutationListener registers a handler for the supplied entity; registration automatically
// opts the entity into event emission
func (e *Eventer) AddMutationListener(entity string, handler MutationHandler) {
	if e == nil || handler == nil || entity == "" {
		return
	}

	if e.listeners == nil {
		e.listeners = make(map[string][]soiree.ListenerBinding)
	}

	bound := soiree.BindListener(
		mutationTopic(entity),
		func(ctx *soiree.EventContext, payload *events.MutationPayload) error {
			return handler(ctx, payload)
		},
	)

	e.listeners[entity] = append(e.listeners[entity], bound)
}

// AddListenerBinding registers a non-mutation listener binding for later registration.
func (e *Eventer) AddListenerBinding(binding soiree.ListenerBinding) {
	if e == nil {
		return
	}

	e.bindings = append(e.bindings, binding)
}

// Initialize configures the Eventer with an event bus bound to the provided client and registers
// the default mutation listeners; use this when you need to pass the same Eventer to multiple
// consumers (e.g., ent hooks and workflow engine)
func (e *Eventer) Initialize(client any) {
	bus := soiree.New(
		soiree.Workers(eventerPoolWorkers),
		soiree.Client(client))
	soiree.WithPoolName("ent_event_pool")

	e.Emitter = bus

	registerDefaultMutationListeners(e)
}

// NewEventerPool builds a fresh event bus, associates it with an Eventer, and wires the default
// mutation listeners
func NewEventerPool(client any) *Eventer {
	eventer := NewEventer()
	eventer.Initialize(client)

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

	e.AddMutationListener(entgen.TypeTrustCenterDoc, handleTrustCenterDocMutation)
	e.AddMutationListener(entgen.TypeTrustCenterSetting, handleTrustCenterSettingMutation)
	e.AddMutationListener(entgen.TypeTrustCenter, handleTrustCenterMutation)
	e.AddMutationListener(entgen.TypeNote, handleNoteMutation)
	e.AddMutationListener(entgen.TypeTrustCenterEntity, handleTrustCenterEntityMutation)
	e.AddMutationListener(entgen.TypeTrustCenterSubprocessor, handleTrustCenterSubprocessorMutation)
	e.AddMutationListener(entgen.TypeTrustCenterCompliance, handleTrustCenterComplianceMutation)
	e.AddMutationListener(entgen.TypeSubprocessor, handleSubprocessorMutation)
	e.AddMutationListener(entgen.TypeStandard, handleStandardMutation)

	notifications.RegisterListeners(func(entityType string, handler func(*soiree.EventContext, *events.MutationPayload) error) {
		e.AddMutationListener(entityType, handler)
	})

	RegisterIntegrationOperationListeners(e)

	if e.workflowListenersEnabled {
		RegisterWorkflowListeners(e)
	}
}
