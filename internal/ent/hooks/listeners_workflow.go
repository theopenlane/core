package hooks

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/samber/do/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/gala"
)

// RegisterGalaWorkflowListeners registers workflow mutation and command listeners on Gala.
func RegisterGalaWorkflowListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	mutationIDs, err := RegisterGalaWorkflowMutationListeners(registry)
	if err != nil {
		return nil, err
	}

	commandIDs, err := gala.RegisterListeners(registry,
		gala.Definition[gala.WorkflowTriggeredPayload]{
			Topic:  gala.WorkflowTriggeredEventTopic,
			Name:   "workflows.command.triggered",
			Handle: handleWorkflowTriggeredGala,
		},
	)
	if err != nil {
		return nil, err
	}

	ids, err := gala.RegisterListeners(registry,
		gala.Definition[gala.WorkflowActionStartedPayload]{
			Topic:  gala.WorkflowActionStartedEventTopic,
			Name:   "workflows.command.action_started",
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
			Name:   "workflows.command.action_completed",
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
			Name:   "workflows.command.assignment_created",
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
			Name:   "workflows.command.assignment_completed",
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
			Name:   "workflows.command.instance_completed",
			Handle: handleWorkflowInstanceCompletedGala,
		},
	)
	if err != nil {
		return nil, err
	}
	commandIDs = append(commandIDs, ids...)

	return append(mutationIDs, commandIDs...), nil
}

// RegisterGalaWorkflowMutationListeners registers workflow mutation listeners on Gala.
func RegisterGalaWorkflowMutationListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	definitions := make([]gala.Definition[eventqueue.MutationGalaPayload], 0, len(enums.WorkflowObjectTypes)+1)

	for _, entity := range enums.WorkflowObjectTypes {
		definitions = append(definitions, gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: workflowMutationGalaTopic(entity),
			Name:  fmt.Sprintf("workflows.mutation.%s", strings.ToLower(entity)),
			Operations: []string{
				ent.OpCreate.String(),
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Handle: handleWorkflowMutationGala,
		})
	}

	definitions = append(definitions, gala.Definition[eventqueue.MutationGalaPayload]{
		Topic: workflowMutationGalaTopic(generated.TypeWorkflowAssignment),
		Name:  "workflows.mutation.workflow_assignment",
		Operations: []string{
			ent.OpUpdate.String(),
			ent.OpUpdateOne.String(),
		},
		Handle: handleWorkflowAssignmentMutationGala,
	})

	return gala.RegisterListeners(registry, definitions...)
}

func handleWorkflowMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleWorkflowMutationGala(ctx, payload)
}

func handleWorkflowAssignmentMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleWorkflowAssignmentMutationGala(ctx, payload)
}

func handleWorkflowTriggeredGala(ctx gala.HandlerContext, payload gala.WorkflowTriggeredPayload) error {
	listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleWorkflowTriggered(ctx, payload)
}

func handleWorkflowActionStartedGala(ctx gala.HandlerContext, payload gala.WorkflowActionStartedPayload) error {
	listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleActionStarted(ctx, payload)
}

func handleWorkflowActionCompletedGala(ctx gala.HandlerContext, payload gala.WorkflowActionCompletedPayload) error {
	listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleActionCompleted(ctx, payload)
}

func handleWorkflowAssignmentCreatedGala(ctx gala.HandlerContext, payload gala.WorkflowAssignmentCreatedPayload) error {
	listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleAssignmentCreated(ctx, payload)
}

func handleWorkflowAssignmentCompletedGala(ctx gala.HandlerContext, payload gala.WorkflowAssignmentCompletedPayload) error {
	listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleAssignmentCompleted(ctx, payload)
}

func handleWorkflowInstanceCompletedGala(ctx gala.HandlerContext, payload gala.WorkflowInstanceCompletedPayload) error {
	listeners, ok := workflowListenersFromGala(ctx)
	if !ok {
		return nil
	}

	return listeners.HandleInstanceCompleted(ctx, payload)
}

func workflowListenersFromGala(handlerCtx gala.HandlerContext) (*engine.WorkflowListeners, bool) {
	client := mutationClientFromGala(handlerCtx)
	if client == nil {
		return nil, false
	}

	wfEngine, ok := client.WorkflowEngine.(*engine.WorkflowEngine)
	if !ok || wfEngine == nil {
		return nil, false
	}

	runtime, _ := do.Invoke[*gala.Gala](handlerCtx.Injector)

	return engine.NewWorkflowListeners(client, wfEngine, runtime), true
}
