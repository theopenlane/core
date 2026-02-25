package hooks

import (
	"entgo.io/ent"
	"github.com/samber/do/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/gala"
)

// RegisterGalaWorkflowListeners registers workflow mutation and command listeners on Gala
func RegisterGalaWorkflowListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	mutationIDs, err := RegisterGalaWorkflowMutationListeners(registry)
	if err != nil {
		return nil, err
	}

	commandIDs, err := gala.RegisterListeners(registry,
		gala.Definition[gala.WorkflowTriggeredPayload]{
			Topic:  gala.WorkflowTriggeredEventTopic,
			Name:   string(gala.TopicWorkflowTriggered),
			Handle: handleWorkflowTriggeredGala,
		},
	)
	if err != nil {
		return nil, err
	}

	ids, err := gala.RegisterListeners(registry,
		gala.Definition[gala.WorkflowActionStartedPayload]{
			Topic:  gala.WorkflowActionStartedEventTopic,
			Name:   string(gala.TopicWorkflowActionStarted),
			Handle: handleWorkflowActionStartedGala,
		},
	)
	if err != nil {
		return nil, err
	}
	commandIDs = append(commandIDs, ids...)

	ids, err = gala.RegisterListeners(registry,
		gala.Definition[gala.WorkflowActionCompletedPayload]{
			Topic:  gala.WorkflowActionCompletedEventTopic,
			Name:   string(gala.TopicWorkflowActionCompleted),
			Handle: handleWorkflowActionCompletedGala,
		},
	)
	if err != nil {
		return nil, err
	}
	commandIDs = append(commandIDs, ids...)

	ids, err = gala.RegisterListeners(registry,
		gala.Definition[gala.WorkflowAssignmentCreatedPayload]{
			Topic:  gala.WorkflowAssignmentCreatedEventTopic,
			Name:   string(gala.TopicWorkflowAssignmentCreated),
			Handle: handleWorkflowAssignmentCreatedGala,
		},
	)
	if err != nil {
		return nil, err
	}
	commandIDs = append(commandIDs, ids...)

	ids, err = gala.RegisterListeners(registry,
		gala.Definition[gala.WorkflowAssignmentCompletedPayload]{
			Topic:  gala.WorkflowAssignmentCompletedEventTopic,
			Name:   string(gala.TopicWorkflowAssignmentCompleted),
			Handle: handleWorkflowAssignmentCompletedGala,
		},
	)
	if err != nil {
		return nil, err
	}
	commandIDs = append(commandIDs, ids...)

	ids, err = gala.RegisterListeners(registry,
		gala.Definition[gala.WorkflowInstanceCompletedPayload]{
			Topic:  gala.WorkflowInstanceCompletedEventTopic,
			Name:   string(gala.TopicWorkflowInstanceCompleted),
			Handle: handleWorkflowInstanceCompletedGala,
		},
	)
	if err != nil {
		return nil, err
	}
	commandIDs = append(commandIDs, ids...)

	return append(mutationIDs, commandIDs...), nil
}

// RegisterGalaWorkflowMutationListeners registers workflow mutation listeners on Gala
func RegisterGalaWorkflowMutationListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	definitions := make([]gala.Definition[eventqueue.MutationGalaPayload], 0, len(enums.WorkflowObjectTypes)+1)

	for _, entity := range enums.WorkflowObjectTypes {
		topicName := eventqueue.MutationTopicName(eventqueue.MutationConcernWorkflow, entity)
		definitions = append(definitions, gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: eventqueue.MutationTopic(eventqueue.MutationConcernWorkflow, entity),
			Name:  string(topicName),
			Operations: []string{
				ent.OpCreate.String(),
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Handle: handleWorkflowMutationGala,
		})
	}

	assignmentTopicName := eventqueue.MutationTopicName(eventqueue.MutationConcernWorkflow, generated.TypeWorkflowAssignment)
	definitions = append(definitions, gala.Definition[eventqueue.MutationGalaPayload]{
		Topic: eventqueue.MutationTopic(eventqueue.MutationConcernWorkflow, generated.TypeWorkflowAssignment),
		Name:  string(assignmentTopicName),
		Operations: []string{
			ent.OpUpdate.String(),
			ent.OpUpdateOne.String(),
		},
		Handle: handleWorkflowAssignmentMutationGala,
	})

	return gala.RegisterListeners(registry, definitions...)
}

// handleWorkflowMutationGala forwards workflow mutation envelopes to workflow listeners
func handleWorkflowMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleWorkflowMutationGala(ctx, payload)
}

// handleWorkflowAssignmentMutationGala forwards assignment mutation envelopes to workflow listeners
func handleWorkflowAssignmentMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleWorkflowAssignmentMutationGala(ctx, payload)
}

// handleWorkflowTriggeredGala forwards workflow trigger command envelopes to workflow listeners
func handleWorkflowTriggeredGala(ctx gala.HandlerContext, payload gala.WorkflowTriggeredPayload) error {
	ctx, listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleWorkflowTriggered(ctx, payload)
}

// handleWorkflowActionStartedGala forwards action started command envelopes to workflow listeners
func handleWorkflowActionStartedGala(ctx gala.HandlerContext, payload gala.WorkflowActionStartedPayload) error {
	ctx, listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleActionStarted(ctx, payload)
}

// handleWorkflowActionCompletedGala forwards action completed command envelopes to workflow listeners
func handleWorkflowActionCompletedGala(ctx gala.HandlerContext, payload gala.WorkflowActionCompletedPayload) error {
	ctx, listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleActionCompleted(ctx, payload)
}

// handleWorkflowAssignmentCreatedGala forwards assignment created command envelopes to workflow listeners
func handleWorkflowAssignmentCreatedGala(ctx gala.HandlerContext, payload gala.WorkflowAssignmentCreatedPayload) error {
	ctx, listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleAssignmentCreated(ctx, payload)
}

// handleWorkflowAssignmentCompletedGala forwards assignment completed command envelopes to workflow listeners
func handleWorkflowAssignmentCompletedGala(ctx gala.HandlerContext, payload gala.WorkflowAssignmentCompletedPayload) error {
	ctx, listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleAssignmentCompleted(ctx, payload)
}

// handleWorkflowInstanceCompletedGala forwards instance completed command envelopes to workflow listeners
func handleWorkflowInstanceCompletedGala(ctx gala.HandlerContext, payload gala.WorkflowInstanceCompletedPayload) error {
	ctx, listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleInstanceCompleted(ctx, payload)
}

// workflowListenersFromGala resolves workflow listener dependencies from the gala injector
// and enriches the handler context so the ent client is available to interceptors
func workflowListenersFromGala(handlerCtx gala.HandlerContext) (gala.HandlerContext, *engine.WorkflowListeners, bool) {
	handlerCtx, client, ok := eventqueue.ClientFromHandler(handlerCtx)
	if !ok {
		return handlerCtx, nil, false
	}

	wfEngine, ok := client.WorkflowEngine.(*engine.WorkflowEngine)
	if !ok || wfEngine == nil {
		return handlerCtx, nil, false
	}

	runtime, err := do.Invoke[*gala.Gala](handlerCtx.Injector)
	if err != nil {
		return handlerCtx, nil, false
	}

	return handlerCtx, engine.NewWorkflowListeners(client, wfEngine, runtime), true
}
