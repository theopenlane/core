package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/ent/privacy/policy"
)

// WorkflowEvent stores events executed within a workflow instance
type WorkflowEvent struct {
	SchemaFuncs
	ent.Schema
}

const schemaWorkflowEvent = "workflow_event"

// Name returns the name of the WorkflowEvent schema
func (WorkflowEvent) Name() string {
	return schemaWorkflowEvent
}

// GetType returns the type of the WorkflowEvent schema
func (WorkflowEvent) GetType() any {
	return WorkflowEvent.Type
}

// PluralName returns the plural name of the WorkflowEvent schema
func (WorkflowEvent) PluralName() string {
	return pluralize.NewClient().Plural(schemaWorkflowEvent)
}

// Fields of the WorkflowEvent
func (WorkflowEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("workflow_instance_id").
			Comment("ID of the workflow instance that generated the event").
			NotEmpty(),
		field.Enum("event_type").
			Comment("Type of event, typically the action kind").
			GoType(enums.WorkflowEventType("")),
		field.JSON("payload", models.WorkflowEventPayload{}).
			Comment("Payload for the event; stored raw").
			Optional(),
	}
}

// Edges of the WorkflowEvent
func (WorkflowEvent) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: WorkflowEvent{},
			edgeSchema: WorkflowInstance{},
			field:      "workflow_instance_id",
			required:   true,
		}),
	}
}

// Mixin of the WorkflowEvent
func (WorkflowEvent) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "WFE",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(WorkflowEvent{}),
		},
	}.getMixins(WorkflowEvent{})
}

// Modules this schema has access to
func (WorkflowEvent) Modules() []models.OrgModule {
	return []models.OrgModule{models.CatalogBaseModule}
}

// Policy of the WorkflowEvent
func (WorkflowEvent) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}
