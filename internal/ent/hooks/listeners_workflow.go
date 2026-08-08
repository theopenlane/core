package hooks

import (
	"entgo.io/ent"
	"github.com/samber/do/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/gala"
)

// RegisterGalaWorkflowListeners registers workflow mutation and command listeners on Gala
func RegisterGalaWorkflowListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	var ids []gala.ListenerID

	collect := func(registered []gala.ListenerID, err error) error {
		if err != nil {
			return err
		}

		ids = append(ids, registered...)

		return nil
	}

	if err := collect(RegisterGalaWorkflowMutationListeners(g)); err != nil {
		return nil, err
	}

	if err := collect(gala.Register(g, gala.Definition[gala.WorkflowTriggeredPayload]{
		Topic:  gala.WorkflowTriggeredEventTopic,
		Name:   string(gala.TopicWorkflowTriggered),
		Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleWorkflowTriggered),
	})); err != nil {
		return nil, err
	}

	if err := collect(gala.Register(g, gala.Definition[gala.WorkflowActionStartedPayload]{
		Topic:  gala.WorkflowActionStartedEventTopic,
		Name:   string(gala.TopicWorkflowActionStarted),
		Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleActionStarted),
	})); err != nil {
		return nil, err
	}

	if err := collect(gala.Register(g, gala.Definition[gala.WorkflowActionCompletedPayload]{
		Topic:  gala.WorkflowActionCompletedEventTopic,
		Name:   string(gala.TopicWorkflowActionCompleted),
		Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleActionCompleted),
	})); err != nil {
		return nil, err
	}

	if err := collect(gala.Register(g, gala.Definition[gala.WorkflowAssignmentCompletedPayload]{
		Topic:  gala.WorkflowAssignmentCompletedEventTopic,
		Name:   string(gala.TopicWorkflowAssignmentCompleted),
		Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleAssignmentCompleted),
	})); err != nil {
		return nil, err
	}

	if err := collect(gala.Register(g, gala.Definition[gala.WorkflowInstanceCompletedPayload]{
		Topic:  gala.WorkflowInstanceCompletedEventTopic,
		Name:   string(gala.TopicWorkflowInstanceCompleted),
		Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleInstanceCompleted),
	})); err != nil {
		return nil, err
	}

	return ids, nil
}

// RegisterGalaWorkflowMutationListeners registers workflow mutation listeners on Gala
func RegisterGalaWorkflowMutationListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	definitions := make([]gala.Definition[entityops.MutationPayload], 0, len(enums.WorkflowObjectTypes)+1)

	for _, entity := range enums.WorkflowObjectTypes {
		topicName := gala.MutationTopicName(gala.MutationConcernWorkflow, entity)
		definitions = append(definitions, gala.Definition[entityops.MutationPayload]{
			Topic: entityops.MutationTopic(gala.MutationConcernWorkflow, entity),
			Name:  string(topicName),
			Operations: []string{
				ent.OpCreate.String(),
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleWorkflowMutationGala),
		})
	}

	assignmentTopicName := gala.MutationTopicName(gala.MutationConcernWorkflow, generated.TypeWorkflowAssignment)
	definitions = append(definitions, gala.Definition[entityops.MutationPayload]{
		Topic: entityops.MutationTopic(gala.MutationConcernWorkflow, generated.TypeWorkflowAssignment),
		Name:  string(assignmentTopicName),
		Operations: []string{
			ent.OpUpdate.String(),
			ent.OpUpdateOne.String(),
		},
		Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleWorkflowAssignmentMutationGala),
	})

	return gala.Register(g, definitions...)
}

// forwardToWorkflowListeners adapts one engine.WorkflowListeners method into a gala handler,
// resolving the listener dependencies from the handler context per event
func forwardToWorkflowListeners[T any](method func(*engine.WorkflowListeners, gala.HandlerContext, T) error) gala.Handler[T] {
	return func(ctx gala.HandlerContext, payload T) error {
		ctx, listeners, ok := workflowListenersFromGala(ctx)
		if !ok {
			return nil
		}

		return method(listeners, ctx, payload)
	}
}

// workflowListenersFromGala resolves workflow listener dependencies from the gala injector
// and enriches the handler context so the ent client is available to interceptors
func workflowListenersFromGala(handlerCtx gala.HandlerContext) (gala.HandlerContext, *engine.WorkflowListeners, bool) {
	client, err := do.Invoke[*generated.Client](handlerCtx.Injector)
	if err != nil || client == nil {
		return handlerCtx, nil, false
	}

	handlerCtx.Context = generated.NewContext(handlerCtx.Context, client)

	wfEngine, err := do.Invoke[*engine.WorkflowEngine](handlerCtx.Injector)
	if err != nil || wfEngine == nil {
		return handlerCtx, nil, false
	}

	runtime, err := do.Invoke[*gala.Gala](handlerCtx.Injector)
	if err != nil {
		return handlerCtx, nil, false
	}

	return handlerCtx, engine.NewWorkflowListeners(client, wfEngine, runtime), true
}
