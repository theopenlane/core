package hooks

import (
	"context"
	"errors"
	"strings"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaWorkflowListeners registers workflow-facing Gala mutation listeners.
func RegisterGalaWorkflowListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	topicNames := append(append([]string(nil), enums.WorkflowObjectTypes...), generated.TypeWorkflowAssignment)
	topicNames = lo.Uniq(topicNames)

	registrations := lo.Map(topicNames, func(topicName string, _ int) gala.Registration[eventqueue.MutationGalaPayload] {
		return gala.Registration[eventqueue.MutationGalaPayload]{
			Topic: gala.Topic[eventqueue.MutationGalaPayload]{
				Name: gala.TopicName(topicName),
			},
			Codec: gala.JSONCodec[eventqueue.MutationGalaPayload]{},
			Policy: gala.TopicPolicy{
				EmitMode:   gala.EmitModeDurable,
				QueueClass: gala.QueueClassWorkflow,
			},
		}
	})

	for _, registration := range registrations {
		err := registration.Register(registry)
		if err == nil || errors.Is(err, gala.ErrTopicAlreadyRegistered) {
			continue
		}

		return nil, err
	}

	definitions := lo.Map(enums.WorkflowObjectTypes, func(topicName string, _ int) gala.Definition[eventqueue.MutationGalaPayload] {
		return gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: gala.Topic[eventqueue.MutationGalaPayload]{
				Name: gala.TopicName(topicName),
			},
			Name:   "workflows.mutation." + strings.ToLower(topicName),
			Handle: handleWorkflowMutationGala,
		}
	})

	definitions = append(definitions, gala.Definition[eventqueue.MutationGalaPayload]{
		Topic: gala.Topic[eventqueue.MutationGalaPayload]{
			Name: gala.TopicName(generated.TypeWorkflowAssignment),
		},
		Name:   "workflows.assignment.mutation",
		Handle: handleWorkflowAssignmentMutationGala,
	})

	ids := make([]gala.ListenerID, 0, len(definitions))
	for _, definition := range definitions {
		id, err := definition.Register(registry)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

// handleWorkflowMutationGala evaluates and triggers matching workflows for workflow-eligible mutations.
func handleWorkflowMutationGala(handlerContext gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	client, wfEngine, ok := galaWorkflowDeps(handlerContext)
	if !ok {
		return nil
	}

	ctx := handlerContext.Context
	if workflows.IsWorkflowBypass(ctx) && !workflows.AllowWorkflowEventEmission(ctx) {
		return nil
	}

	if workflows.ShouldSkipEventEmission(ctx) {
		return nil
	}

	if payload.Operation != ent.OpUpdate.String() && payload.Operation != ent.OpUpdateOne.String() && payload.Operation != ent.OpCreate.String() {
		return nil
	}

	schemaType := strings.TrimSpace(payload.MutationType)
	if schemaType == "" {
		schemaType = strings.TrimSpace(string(handlerContext.Envelope.Topic))
	}
	if schemaType == "" {
		return nil
	}

	eventType := workflowEventTypeFromEntOperation(payload.Operation)
	changedFields := lo.Uniq(append([]string(nil), payload.ChangedFields...))
	if eventType != "CREATE" {
		if objectType := enums.ToWorkflowObjectType(schemaType); objectType != nil {
			eligible := workflows.EligibleWorkflowFields(*objectType)
			if len(eligible) > 0 {
				changedFields = lo.Filter(changedFields, func(field string, _ int) bool {
					_, exists := eligible[field]

					return exists
				})
			}
		}
	}

	changedEdges := lo.Uniq(append([]string(nil), payload.ChangedEdges...))
	addedIDs := events.CloneStringSliceMap(payload.AddedIDs)
	removedIDs := events.CloneStringSliceMap(payload.RemovedIDs)
	proposedChanges := events.CloneAnyMap(payload.ProposedChanges)

	if len(changedFields) == 0 && len(changedEdges) == 0 && eventType != "CREATE" {
		return nil
	}

	entityID := galaMutationEntityID(handlerContext, payload)
	if entityID == "" {
		return nil
	}

	obj, err := loadGalaWorkflowObject(workflows.AllowContext(ctx), client, schemaType, entityID)
	if err != nil {
		return nil
	}

	definitions, err := wfEngine.FindMatchingDefinitions(
		workflows.AllowContext(ctx),
		schemaType,
		eventType,
		changedFields,
		changedEdges,
		addedIDs,
		removedIDs,
		proposedChanges,
		obj,
	)
	if err != nil || len(definitions) == 0 {
		return nil
	}

	for _, definition := range definitions {
		if workflows.DefinitionUsesPreCommitApprovals(definition.DefinitionJSON) {
			continue
		}

		_, triggerErr := wfEngine.TriggerWorkflow(ctx, definition, obj, engine.TriggerInput{
			EventType:       eventType,
			ChangedFields:   changedFields,
			ChangedEdges:    changedEdges,
			AddedIDs:        addedIDs,
			RemovedIDs:      removedIDs,
			ProposedChanges: proposedChanges,
		})
		if triggerErr != nil && !errors.Is(triggerErr, workflows.ErrWorkflowAlreadyActive) {
			logx.FromContext(ctx).Error().Err(triggerErr).Str("definition_id", definition.ID).Msg("failed to trigger workflow")
		}
	}

	return nil
}

// handleWorkflowAssignmentMutationGala completes workflow assignments when status transitions.
func handleWorkflowAssignmentMutationGala(handlerContext gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	_, wfEngine, ok := galaWorkflowDeps(handlerContext)
	if !ok {
		return nil
	}

	if payload.Operation != ent.OpUpdate.String() && payload.Operation != ent.OpUpdateOne.String() {
		return nil
	}

	assignmentID := galaMutationEntityID(handlerContext, payload)
	if assignmentID == "" {
		return nil
	}

	if !lo.Contains(payload.ChangedFields, workflowassignment.FieldStatus) {
		return nil
	}

	rawStatus, found := payload.ProposedChanges[workflowassignment.FieldStatus]
	if !found {
		status := strings.TrimSpace(handlerContext.Envelope.Headers.Properties[workflowassignment.FieldStatus])
		if status != "" {
			rawStatus = status
			found = true
		}
	}
	if !found {
		return nil
	}

	nextStatus := events.ParseEnumPtr(rawStatus, enums.ToWorkflowAssignmentStatus)
	if nextStatus == nil || *nextStatus == enums.WorkflowAssignmentStatusPending {
		return nil
	}

	return wfEngine.CompleteAssignment(handlerContext.Context, assignmentID, *nextStatus, nil, nil)
}

// galaWorkflowDeps resolves workflow mutation listener dependencies from Gala handler context.
func galaWorkflowDeps(handlerContext gala.HandlerContext) (*generated.Client, *engine.WorkflowEngine, bool) {
	client, err := gala.ResolveFromContext[*generated.Client](handlerContext)
	if err != nil || client == nil {
		return nil, nil, false
	}

	wfEngine, ok := client.WorkflowEngine.(*engine.WorkflowEngine)
	if !ok || wfEngine == nil {
		return nil, nil, false
	}

	return client, wfEngine, true
}

// galaMutationEntityID resolves the mutation entity identifier from payload metadata and headers.
func galaMutationEntityID(handlerContext gala.HandlerContext, payload eventqueue.MutationGalaPayload) string {
	if id := strings.TrimSpace(payload.EntityID); id != "" {
		return id
	}

	return strings.TrimSpace(handlerContext.Envelope.Headers.Properties["ID"])
}

// loadGalaWorkflowObject loads a workflow object from the generated object registry.
func loadGalaWorkflowObject(ctx context.Context, client *generated.Client, schemaType, entityID string) (*workflows.Object, error) {
	entity, err := workflows.LoadWorkflowObject(ctx, client, schemaType, entityID)
	if err != nil {
		return nil, err
	}

	objectType := enums.ToWorkflowObjectType(schemaType)
	if objectType == nil {
		return nil, workflows.ErrUnsupportedObjectType
	}

	return &workflows.Object{
		ID:   entityID,
		Type: *objectType,
		Node: entity,
	}, nil
}

// workflowEventTypeFromEntOperation maps ent mutation operations to workflow event type names.
func workflowEventTypeFromEntOperation(operation string) string {
	switch operation {
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		return "UPDATE"
	case ent.OpCreate.String():
		return "CREATE"
	case ent.OpDelete.String(), ent.OpDeleteOne.String():
		return "DELETE"
	default:
		return operation
	}
}
