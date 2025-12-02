package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/shared/models"
	"github.com/theopenlane/ent/privacy/policy"
)

// WorkflowInstance tracks execution of a workflow definition for a specific object
type WorkflowInstance struct {
	SchemaFuncs
	ent.Schema
}

const schemaWorkflowInstance = "workflow_instance"

// Name returns the name of the WorkflowInstance schema
func (WorkflowInstance) Name() string {
	return schemaWorkflowInstance
}

// GetType returns the type of the WorkflowInstance schema
func (WorkflowInstance) GetType() any {
	return WorkflowInstance.Type
}

// PluralName returns the plural name of the WorkflowInstance schema
func (WorkflowInstance) PluralName() string {
	return pluralize.NewClient().Plural(schemaWorkflowInstance)
}

// Fields of the WorkflowInstance
func (WorkflowInstance) Fields() []ent.Field {
	return []ent.Field{
		field.String("workflow_definition_id").
			Comment("ID of the workflow definition this instance is based on").
			NotEmpty(),
		field.Enum("state").
			Comment("Current state of the workflow instance").
			GoType(enums.WorkflowInstanceState("")).
			Default(string(enums.WorkflowInstanceStateRunning)),
		field.JSON("context", models.WorkflowInstanceContext{}).
			Comment("Optional context for the workflow instance").
			Optional(),
		field.Time("last_evaluated_at").
			Comment("Timestamp when the workflow was last evaluated").
			Optional().Nillable(),
		field.JSON("definition_snapshot", models.WorkflowDefinitionDocument{}).
			Comment("Copy of definition JSON used for this instance").
			Optional(),
	}
}

// Edges of the WorkflowInstance
func (w WorkflowInstance) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowDefinition{},
			field:      "workflow_definition_id",
			comment:    "Definition driving this instance",
			required:   true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowAssignment{},
			name:       "workflow_assignments",
			comment:    "Assignments associated with this workflow instance",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowEvent{},
			name:       "workflow_events",
			comment:    "Events recorded for this instance",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowObjectRef{},
			name:       "workflow_object_refs",
			comment:    "Object references for this workflow instance",
		}),
	}
}

// Indexes of the WorkflowInstance
func (WorkflowInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workflow_definition_id").
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

// Mixin of the WorkflowInstance
func (WorkflowInstance) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "WFI",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(WorkflowInstance{}),
		},
	}.getMixins(WorkflowInstance{})
}

// Modules this schema has access to
func (WorkflowInstance) Modules() []models.OrgModule {
	return []models.OrgModule{models.CatalogBaseModule}
}

// Policy of the WorkflowInstance
func (WorkflowInstance) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}
