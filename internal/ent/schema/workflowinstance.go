package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
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
		field.String("workflow_proposal_id").
			Comment("ID of the workflow proposal this instance is associated with (when approval-before-commit is used)").
			Optional(),
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
		field.Int("current_action_index").
			Comment("Index of the current action being executed (used for recovery and resumption)").
			Default(0).
			NonNegative(),
		field.String("control_id").
			Comment("ID of the control this workflow instance is associated with").
			Optional(),
		field.String("internal_policy_id").
			Comment("ID of the internal policy this workflow instance is associated with").
			Optional(),
		field.String("evidence_id").
			Comment("ID of the evidence this workflow instance is associated with").
			Optional(),
		field.String("subcontrol_id").
			Comment("ID of the subcontrol this workflow instance is associated with").
			Optional(),
		field.String("action_plan_id").
			Comment("ID of the actionplan this workflow instance is associated with").
			Optional(),
		field.String("procedure_id").
			Comment("ID of the procedure this workflow instance is associated with").
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
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(WorkflowDefinition{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Control{},
			field:      "control_id",
			comment:    "Control this workflow instance is associated with",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: InternalPolicy{},
			field:      "internal_policy_id",
			comment:    "Internal policy this workflow instance is associated with",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Evidence{},
			field:      "evidence_id",
			comment:    "Evidence this workflow instance is associated with",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Subcontrol{},
			field:      "subcontrol_id",
			comment:    "Subcontrol this workflow instance is associated with",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: ActionPlan{},
			field:      "action_plan_id",
			comment:    "ActionPlan this workflow instance is associated with",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Procedure{},
			field:      "procedure_id",
			comment:    "Procedure this workflow instance is associated with",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowProposal{},
			field:      "workflow_proposal_id",
			comment:    "Proposal this workflow instance is associated with",
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
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
			newObjectOwnedMixin[generated.WorkflowInstance](WorkflowInstance{},
				withParents(Control{}, InternalPolicy{}, Evidence{}, Subcontrol{}, ActionPlan{}, Procedure{}),
				withOrganizationOwner(true),
			),
		},
	}.getMixins(WorkflowInstance{})
}

// Modules this schema has access to
func (WorkflowInstance) Modules() []models.OrgModule {
	return []models.OrgModule{models.CatalogBaseModule}
}

// Annotations of the WorkflowInstance
func (WorkflowInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
	}
}

// Policy of the WorkflowInstance
func (WorkflowInstance) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.WorkflowInstanceMutation](),
		),
	)
}
