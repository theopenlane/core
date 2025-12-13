package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// WorkflowAssignment stores approval assignment records for workflow instances
type WorkflowAssignment struct {
	SchemaFuncs
	ent.Schema
}

// schemaWorkflowAssignment is the name of the WorkflowAssignment schema in snake case
const schemaWorkflowAssignment = "workflow_assignment"

// Name returns the name of the WorkflowAssignment schema
func (WorkflowAssignment) Name() string {
	return schemaWorkflowAssignment
}

// GetType returns the type of the WorkflowAssignment schema
func (WorkflowAssignment) GetType() any {
	return WorkflowAssignment.Type
}

// PluralName returns the plural name of the WorkflowAssignment schema
func (WorkflowAssignment) PluralName() string {
	return pluralize.NewClient().Plural(schemaWorkflowAssignment)
}

// Fields of the WorkflowAssignment
func (WorkflowAssignment) Fields() []ent.Field {
	return []ent.Field{
		field.String("workflow_instance_id").
			Comment("ID of the workflow instance this assignment belongs to").
			NotEmpty(),
		field.String("assignment_key").
			Comment("Unique key for the assignment within the workflow instance").
			NotEmpty(),
		field.String("role").
			Comment("Role assigned to the target, e.g. APPROVER").
			Default("APPROVER"),
		field.String("label").
			Comment("Optional label for the assignment").
			Optional(),
		field.Bool("required").
			Comment("Whether this assignment is required for workflow progression").
			Default(true),
		field.Enum("status").
			Comment("Current status of the assignment").
			GoType(enums.WorkflowAssignmentStatus("")).
			Default(string(enums.WorkflowAssignmentStatusPending)),
		field.JSON("metadata", map[string]any{}).
			Comment("Optional metadata for the assignment").
			Optional(),
		field.Time("decided_at").
			Comment("Timestamp when the assignment was decided").
			Optional().Nillable(),
		field.String("actor_user_id").
			Comment("User who made the decision").
			Optional(),
		field.String("actor_group_id").
			Comment("Group that acted on the decision (if applicable)").
			Optional(),
		field.Text("notes").
			Comment("Optional notes about the assignment").
			Optional(),
	}
}

// Edges of the WorkflowAssignment
func (w WorkflowAssignment) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowInstance{},
			field:      "workflow_instance_id",
			comment:    "Instance this assignment belongs to",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(WorkflowInstance{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowAssignmentTarget{},
			name:       "targets",
			comment:    "Targets for this assignment (user/group/resolver)",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: User{},
			name:       "actor_user",
			field:      "actor_user_id",
			comment:    "User who acted on this assignment",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Group{},
			name:       "actor_group",
			field:      "actor_group_id",
			comment:    "Group that acted on this assignment",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
	}
}

// Indexes of the WorkflowAssignment
func (WorkflowAssignment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workflow_instance_id", "assignment_key").
			Unique(),
	}
}

// Mixin of the WorkflowAssignment
func (WorkflowAssignment) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "WFA",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.WorkflowAssignment](WorkflowAssignment{},
				withParents(WorkflowInstance{}),
				withOrganizationOwner(true),
			),
		},
	}.getMixins(WorkflowAssignment{})
}

// Modules this schema has access to
func (WorkflowAssignment) Modules() []models.OrgModule {
	return []models.OrgModule{models.CatalogBaseModule}
}

// Annotations of the WorkflowAssignment
func (WorkflowAssignment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the WorkflowAssignment
func (WorkflowAssignment) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			//			entfga.CheckEditAccess[*generated.WorkflowAssignmentMutation](),
		),
	)
}
