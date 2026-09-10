package hooks

import (
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// workflowMutationListenerLabel identifies the workflow mutation listener family in
// dependency-resolution skip logs
const workflowMutationListenerLabel = "workflow.mutation"

func internalOperationBypassCaller(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
	return restored.WithCapabilities(auth.CapInternalOperation | auth.CapBypassOrgFilter)
}

// WorkflowListeners wires workflow mutations and command events to the workflow engine
func WorkflowListeners() []gala.Registration {
	return append(WorkflowMutationListeners(),
		gala.Definition[engine.WorkflowTriggeredPayload]{
			Topic:  engine.WorkflowTriggeredEventTopic,
			Name:   string(engine.WorkflowTriggeredEventTopic.Name),
			Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleWorkflowTriggered),
		},
		gala.Definition[engine.WorkflowActionStartedPayload]{
			Topic:  engine.WorkflowActionStartedEventTopic,
			Name:   string(engine.WorkflowActionStartedEventTopic.Name),
			Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleActionStarted),
		},
		gala.Definition[engine.WorkflowActionCompletedPayload]{
			Topic:  engine.WorkflowActionCompletedEventTopic,
			Name:   string(engine.WorkflowActionCompletedEventTopic.Name),
			Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleActionCompleted),
		},
		gala.Definition[engine.WorkflowAssignmentCompletedPayload]{
			Topic:  engine.WorkflowAssignmentCompletedEventTopic,
			Name:   string(engine.WorkflowAssignmentCompletedEventTopic.Name),
			Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleAssignmentCompleted),
		},
		gala.Definition[engine.WorkflowInstanceCompletedPayload]{
			Topic:  engine.WorkflowInstanceCompletedEventTopic,
			Name:   string(engine.WorkflowInstanceCompletedEventTopic.Name),
			Handle: forwardToWorkflowListeners((*engine.WorkflowListeners).HandleInstanceCompleted),
		},
	)
}

// WorkflowMutationListeners forwards workflow-eligible mutations to the workflow engine
func WorkflowMutationListeners() []gala.Registration {
	schemas := lo.FilterMap(enums.WorkflowObjectTypes, func(entity string, _ int) (*entityops.Schema, bool) {
		return entityops.LookupSchema(entity)
	})

	return append(entityops.ForSchemas(schemas, entityops.MutationListener{
		Concern:    entityops.MutationConcernWorkflow,
		Operations: []string{entityops.OpCreate, entityops.OpUpdate, entityops.OpUpdateOne},
		Caller:     internalOperationBypassCaller,
		Handle:     forwardToWorkflowMutation((*engine.WorkflowListeners).HandleWorkflowMutationGala),
	}), entityops.MutationListener{
		Concern:    entityops.MutationConcernWorkflow,
		Schema:     entityops.SchemaWorkflowAssignment,
		Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
		Fields:     []string{workflowassignment.FieldStatus},
		Match: []entityops.FieldMatch{{
			Field:  workflowassignment.FieldStatus,
			In:     []string{enums.WorkflowAssignmentStatusPending.String()},
			Negate: true,
		}},
		Handle: forwardToWorkflowMutation((*engine.WorkflowListeners).HandleWorkflowAssignmentMutationGala),
	})
}

// forwardToWorkflowMutation adapts one engine.WorkflowListeners mutation method into a
// mutation listener handler, resolving the workflow engine and event runtime per event
func forwardToWorkflowMutation(method func(*engine.WorkflowListeners, entityops.Invocation, entityops.MutationPayload) error) func(entityops.Invocation, entityops.MutationPayload) error {
	return func(inv entityops.Invocation, payload entityops.MutationPayload) error {
		wfEngine, ok := gala.Resolve[*engine.WorkflowEngine](inv.Context, inv.Injector, workflowMutationListenerLabel)
		if !ok {
			return nil
		}

		runtime, ok := gala.Resolve[*gala.Gala](inv.Context, inv.Injector, workflowMutationListenerLabel)
		if !ok {
			return nil
		}

		return method(engine.NewWorkflowListeners(inv.Client, wfEngine, runtime), inv, payload)
	}
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
	client, ok := gala.Resolve[*generated.Client](handlerCtx.Context, handlerCtx.Injector, workflowMutationListenerLabel)
	if !ok || client == nil {
		return handlerCtx, nil, false
	}

	handlerCtx.Context = generated.NewContext(handlerCtx.Context, client)

	wfEngine, ok := gala.Resolve[*engine.WorkflowEngine](handlerCtx.Context, handlerCtx.Injector, workflowMutationListenerLabel)
	if !ok || wfEngine == nil {
		return handlerCtx, nil, false
	}

	runtime, ok := gala.Resolve[*gala.Gala](handlerCtx.Context, handlerCtx.Injector, workflowMutationListenerLabel)
	if !ok {
		return handlerCtx, nil, false
	}

	return handlerCtx, engine.NewWorkflowListeners(client, wfEngine, runtime), true
}
