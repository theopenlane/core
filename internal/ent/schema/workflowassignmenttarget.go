package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// WorkflowAssignmentTarget links an assignment to specific targets (user/group/resolver)
type WorkflowAssignmentTarget struct {
	SchemaFuncs
	ent.Schema
}

const schemaWorkflowAssignmentTarget = "workflow_assignment_target"

// Name returns the name of the WorkflowAssignmentTarget schema
func (WorkflowAssignmentTarget) Name() string {
	return schemaWorkflowAssignmentTarget
}

// GetType returns the type of the WorkflowAssignmentTarget schema
func (WorkflowAssignmentTarget) GetType() any {
	return WorkflowAssignmentTarget.Type
}

// PluralName returns the plural name of the WorkflowAssignmentTarget schema
func (WorkflowAssignmentTarget) PluralName() string {
	return pluralize.NewClient().Plural(schemaWorkflowAssignmentTarget)
}

// Fields of the WorkflowAssignmentTarget
func (WorkflowAssignmentTarget) Fields() []ent.Field {
	return []ent.Field{
		field.String("workflow_assignment_id").
			Comment("Assignment this target belongs to").
			NotEmpty(),
		field.Enum("target_type").
			Comment("Type of the target (USER, GROUP, ROLE, RESOLVER)").
			GoType(enums.WorkflowTargetType("")),
		field.String("target_user_id").
			Comment("User target when target_type is USER").
			Optional(),
		field.String("target_group_id").
			Comment("Group target when target_type is GROUP").
			Optional(),
		field.String("resolver_key").
			Comment("Resolver key when target_type is RESOLVER").
			Optional(),
	}
}

// Edges of the WorkflowAssignmentTarget
func (w WorkflowAssignmentTarget) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowAssignment{},
			field:      "workflow_assignment_id",
			comment:    "Assignment this target belongs to",
			required:   true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: User{},
			name:       "user_target",
			field:      "target_user_id",
			comment:    "User target when target_type is USER",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Group{},
			name:       "group_target",
			field:      "target_group_id",
			comment:    "Group target when target_type is GROUP",
		}),
	}
}

// Indexes of the WorkflowAssignmentTarget
func (WorkflowAssignmentTarget) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workflow_assignment_id").
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("workflow_assignment_id", "target_type", "target_user_id", "target_group_id", "resolver_key").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

// Mixin of the WorkflowAssignmentTarget
func (WorkflowAssignmentTarget) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "WFT",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.WorkflowAssignmentTarget](WorkflowAssignmentTarget{},
				withParents(WorkflowAssignment{}),
				withOrganizationOwnerServiceOnly(true),
			),
		},
	}.getMixins(WorkflowAssignmentTarget{})
}

// Modules this schema has access to
func (WorkflowAssignmentTarget) Modules() []models.OrgModule {
	return []models.OrgModule{models.CatalogBaseModule}
}

// Policy of the WorkflowAssignmentTarget
func (WorkflowAssignmentTarget) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckOrgReadAccess(),
		),
		policy.WithMutationRules(
			rule.AllowIfInternalRequest(),
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the WorkflowAssignmentTarget
func (WorkflowAssignmentTarget) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
	}
}
