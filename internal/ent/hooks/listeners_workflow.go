package hooks

import (
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// RegisterWorkflowListeners registers workflow event listeners and mutation triggers.
func RegisterWorkflowListeners(eventer *Eventer) {
	getListeners := func(ctx *soiree.EventContext) (*engine.WorkflowListeners, bool) {
		client, ok := soiree.ClientAs[*generated.Client](ctx)
		if !ok {
			return nil, false
		}

		wfEngine, ok := client.WorkflowEngine.(*engine.WorkflowEngine)
		if !ok || wfEngine == nil {
			return nil, false
		}

		return engine.NewWorkflowListeners(client, wfEngine, eventer.Emitter), true
	}

	eventer.AddListenerBinding(bindWorkflowListener(soiree.WorkflowTriggeredTopic, getListeners, (*engine.WorkflowListeners).HandleWorkflowTriggered))
	eventer.AddListenerBinding(bindWorkflowListener(soiree.WorkflowActionStartedTopic, getListeners, (*engine.WorkflowListeners).HandleActionStarted))
	eventer.AddListenerBinding(bindWorkflowListener(soiree.WorkflowActionCompletedTopic, getListeners, (*engine.WorkflowListeners).HandleActionCompleted))
	eventer.AddListenerBinding(bindWorkflowListener(soiree.WorkflowAssignmentCreatedTopic, getListeners, (*engine.WorkflowListeners).HandleAssignmentCreated))
	eventer.AddListenerBinding(bindWorkflowListener(soiree.WorkflowAssignmentCompletedTopic, getListeners, (*engine.WorkflowListeners).HandleAssignmentCompleted))
	eventer.AddListenerBinding(bindWorkflowListener(soiree.WorkflowInstanceCompletedTopic, getListeners, (*engine.WorkflowListeners).HandleInstanceCompleted))

	// small adapter to match MutationHandler signature
	wrapMutationHandler := func(handler func(*engine.WorkflowListeners, *soiree.EventContext, *events.MutationPayload) error) MutationHandler {
		return func(ctx *soiree.EventContext, payload *events.MutationPayload) error {
			listeners, ok := getListeners(ctx)
			if !ok {
				return nil
			}
			return handler(listeners, ctx, payload)
		}
	}

	for _, entity := range enums.WorkflowObjectTypes {
		eventer.AddMutationListener(entity, wrapMutationHandler((*engine.WorkflowListeners).HandleWorkflowMutation))
	}

	eventer.AddMutationListener(generated.TypeWorkflowAssignment, wrapMutationHandler((*engine.WorkflowListeners).HandleWorkflowAssignmentMutation))
}

// bindWorkflowListener is a generics helper to reduce boilerplate when binding typed listeners
func bindWorkflowListener[T any](topic soiree.TypedTopic[T], getListeners func(*soiree.EventContext) (*engine.WorkflowListeners, bool), handler func(*engine.WorkflowListeners, *soiree.EventContext, T) error) soiree.ListenerBinding {
	return soiree.BindListener(topic, func(ctx *soiree.EventContext, payload T) error {
		listeners, ok := getListeners(ctx)
		if !ok {
			return nil
		}

		return handler(listeners, ctx, payload)
	})
}
