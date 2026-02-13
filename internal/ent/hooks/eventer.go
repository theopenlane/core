package hooks

import (
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/notifications"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/gala"
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
	// mutationOutboxEnabled routes mutation events through River when enabled
	mutationOutboxEnabled bool
	// mutationOutboxFailOnEnqueueError upgrades enqueue fallback logging to error-level
	mutationOutboxFailOnEnqueueError bool
	// mutationOutboxTopics optionally scopes outbox routing to specific mutation topics
	mutationOutboxTopics map[string]struct{}
	// galaRuntimeProvider resolves the active gala runtime at emit time
	galaRuntimeProvider func() *gala.Runtime
	// galaDualEmitEnabled toggles mutation dual emit into gala
	galaDualEmitEnabled bool
	// galaFailOnEnqueueError upgrades gala dual-emit enqueue fallback logging to error-level
	galaFailOnEnqueueError bool
	// galaTopics optionally scopes gala dual emit to specific mutation topics
	galaTopics map[string]struct{}
	// galaTopicModes optionally overrides migration behavior per mutation topic
	galaTopicModes map[string]workflows.GalaTopicMode
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

	lo.ForEach(opts, func(opt EventerOpts, _ int) { opt(e) })

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

// WithMutationOutboxEnabled toggles River-backed mutation dispatch.
func WithMutationOutboxEnabled(enabled bool) EventerOpts {
	return func(e *Eventer) {
		e.mutationOutboxEnabled = enabled
	}
}

// WithMutationOutboxFailOnEnqueueError controls strict-mode logging for outbox enqueue failures.
func WithMutationOutboxFailOnEnqueueError(enabled bool) EventerOpts {
	return func(e *Eventer) {
		e.mutationOutboxFailOnEnqueueError = enabled
	}
}

// WithMutationOutboxTopics scopes River-backed mutation dispatch to the provided topic names.
// Empty list means all mutation topics are eligible for outbox dispatch.
func WithMutationOutboxTopics(topics []string) EventerOpts {
	return func(e *Eventer) {
		e.mutationOutboxTopics = lo.SliceToMap(topics, func(topic string) (string, struct{}) {
			return topic, struct{}{}
		})
	}
}

// WithGalaRuntimeProvider injects a gala runtime provider used during mutation dual emit.
func WithGalaRuntimeProvider(provider func() *gala.Runtime) EventerOpts {
	return func(e *Eventer) {
		e.galaRuntimeProvider = provider
	}
}

// WithGalaDualEmitEnabled toggles mutation dual emit into gala.
func WithGalaDualEmitEnabled(enabled bool) EventerOpts {
	return func(e *Eventer) {
		e.galaDualEmitEnabled = enabled
	}
}

// WithGalaFailOnEnqueueError controls strict-mode logging for gala dual-emit enqueue failures.
func WithGalaFailOnEnqueueError(enabled bool) EventerOpts {
	return func(e *Eventer) {
		e.galaFailOnEnqueueError = enabled
	}
}

// WithGalaTopics scopes gala dual emit to provided mutation topics.
// Empty list means all mutation topics are eligible for dual emit.
func WithGalaTopics(topics []string) EventerOpts {
	return func(e *Eventer) {
		e.galaTopics = lo.SliceToMap(topics, func(topic string) (string, struct{}) {
			return topic, struct{}{}
		})
	}
}

// WithGalaTopicModes overrides migration behavior by mutation topic.
func WithGalaTopicModes(topicModes map[string]workflows.GalaTopicMode) EventerOpts {
	return func(e *Eventer) {
		validModes := lo.PickBy(topicModes, func(topic string, mode workflows.GalaTopicMode) bool {
			return strings.TrimSpace(topic) != "" && mode.IsValid()
		})
		if len(validModes) == 0 {
			e.galaTopicModes = nil
			return
		}

		e.galaTopicModes = lo.Assign(map[string]workflows.GalaTopicMode{}, validModes)
	}
}

// shouldUseMutationOutbox reports whether the topic should be dispatched through the mutation outbox.
func (e *Eventer) shouldUseMutationOutbox(topic string) bool {
	if !e.mutationOutboxEnabled {
		return false
	}

	if len(e.mutationOutboxTopics) == 0 {
		return true
	}

	if _, ok := e.mutationOutboxTopics["*"]; ok {
		return true
	}

	_, ok := e.mutationOutboxTopics[topic]

	return ok
}

// galaRuntime resolves the configured gala runtime if one is available.
func (e *Eventer) galaRuntime() *gala.Runtime {
	if e.galaRuntimeProvider == nil {
		return nil
	}

	return e.galaRuntimeProvider()
}

// shouldUseGalaDispatch reports whether the topic should be emitted into gala.
func (e *Eventer) shouldUseGalaDispatch(topic string) bool {
	if mode, ok := e.topicMode(topic); ok {
		return mode == workflows.GalaTopicModeDualEmit || mode == workflows.GalaTopicModeV2Only
	}

	if !e.galaDualEmitEnabled {
		return false
	}

	if len(e.galaTopics) == 0 {
		return true
	}

	if _, ok := e.galaTopics["*"]; ok {
		return true
	}

	_, ok := e.galaTopics[topic]

	return ok
}

// shouldUseLegacyEmit reports whether the topic should continue through legacy soiree paths.
func (e *Eventer) shouldUseLegacyEmit(topic string) bool {
	mode, ok := e.topicMode(topic)
	if !ok {
		return true
	}

	return mode != workflows.GalaTopicModeV2Only
}

// topicMode resolves explicit per-topic mode overrides using exact-match, then wildcard.
func (e *Eventer) topicMode(topic string) (workflows.GalaTopicMode, bool) {
	if len(e.galaTopicModes) == 0 {
		return "", false
	}

	if mode, ok := e.galaTopicModes[topic]; ok {
		return mode, true
	}

	if mode, ok := e.galaTopicModes["*"]; ok {
		return mode, true
	}

	return "", false
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
	if handler == nil || entity == "" {
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

	if e.workflowListenersEnabled {
		RegisterWorkflowListeners(e)
	}
}
