package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// WorkflowObjectRef is a through table linking workflow instances to workflow-addressable objects.
type WorkflowObjectRef struct {
	SchemaFuncs
	ent.Schema
}

const schemaWorkflowObjectRef = "workflow_object_ref"

// Name returns the name of the WorkflowObjectRef schema
func (WorkflowObjectRef) Name() string {
	return schemaWorkflowObjectRef
}

// GetType returns the type of the WorkflowObjectRef schema
func (WorkflowObjectRef) GetType() any {
	return WorkflowObjectRef.Type
}

// PluralName returns the plural name of the WorkflowObjectRef schema
func (WorkflowObjectRef) PluralName() string {
	return pluralize.NewClient().Plural(schemaWorkflowObjectRef)
}

// Fields of the WorkflowObjectRef
func (WorkflowObjectRef) Fields() []ent.Field {
	return []ent.Field{
		field.String("workflow_instance_id").
			Immutable().
			Comment("Workflow instance this object is associated with").
			NotEmpty(),
		field.String("control_id").
			Immutable().
			Comment("Control referenced by this workflow instance").
			Optional(),
		field.String("task_id").
			Immutable().
			Comment("Task referenced by this workflow instance").
			Optional(),
		field.String("internal_policy_id").
			Immutable().
			Comment("Policy referenced by this workflow instance").
			Optional(),
		field.String("finding_id").
			Immutable().
			Comment("Finding referenced by this workflow instance").
			Optional(),
		field.String("directory_account_id").
			Immutable().
			Comment("Directory account referenced by this workflow instance").
			Optional(),
		field.String("directory_group_id").
			Immutable().
			Comment("Directory group referenced by this workflow instance").
			Optional(),
		field.String("directory_membership_id").
			Immutable().
			Comment("Directory membership referenced by this workflow instance").
			Optional(),
		field.String("evidence_id").
			Immutable().
			Comment("Evidence referenced by this workflow instance").
			Optional(),
		field.String("subcontrol_id").
			Immutable().
			Comment("Subcontrol referenced by this workflow instance").
			Optional(),
		field.String("action_plan_id").
			Immutable().
			Comment("ActionPlan referenced by this workflow instance").
			Optional(),
		field.String("procedure_id").
			Immutable().
			Comment("Procedure referenced by this workflow instance").
			Optional(),
		field.String("campaign_id").
			Immutable().
			Comment("Campaign referenced by this workflow instance").
			Optional(),
		field.String("campaign_target_id").
			Immutable().
			Comment("Campaign target referenced by this workflow instance").
			Optional(),
		field.String("identity_holder_id").
			Immutable().
			Comment("Identity holder referenced by this workflow instance").
			Optional(),
		field.String("platform_id").
			Immutable().
			Comment("Platform referenced by this workflow instance").
			Optional(),
	}
}

// Edges of the WorkflowObjectRef
func (w WorkflowObjectRef) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowInstance{},
			field:      "workflow_instance_id",
			comment:    "Workflow instance this object is associated with",
			required:   true,
			immutable:  true,
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowProposal{},
			ref:        "workflow_object_ref",
			comment:    "Workflow proposals targeting this object reference",
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
				entx.CascadeAnnotationField("WorkflowObjectRefID"),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Control{},
			field:      "control_id",
			comment:    "Control referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Task{},
			field:      "task_id",
			comment:    "Task referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: InternalPolicy{},
			field:      "internal_policy_id",
			comment:    "Policy referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Finding{},
			field:      "finding_id",
			comment:    "Finding referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: DirectoryAccount{},
			field:      "directory_account_id",
			comment:    "Directory account referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: DirectoryGroup{},
			field:      "directory_group_id",
			comment:    "Directory group referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: DirectoryMembership{},
			field:      "directory_membership_id",
			comment:    "Directory membership referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Evidence{},
			field:      "evidence_id",
			comment:    "Evidence referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Subcontrol{},
			field:      "subcontrol_id",
			comment:    "Subcontrol referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: ActionPlan{},
			field:      "action_plan_id",
			comment:    "ActionPlan referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Procedure{},
			field:      "procedure_id",
			comment:    "Procedure referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Campaign{},
			field:      "campaign_id",
			comment:    "Campaign referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: CampaignTarget{},
			field:      "campaign_target_id",
			comment:    "Campaign target referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: IdentityHolder{},
			field:      "identity_holder_id",
			comment:    "Identity holder referenced by this workflow instance",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Platform{},
			field:      "platform_id",
			comment:    "Platform referenced by this workflow instance",
			immutable:  true,
		}),
	}
}

// Indexes of the WorkflowObjectRef
func (WorkflowObjectRef) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workflow_instance_id", "control_id").
			Unique(),
		index.Fields("workflow_instance_id", "task_id").
			Unique(),
		index.Fields("workflow_instance_id", "internal_policy_id").
			Unique(),
		index.Fields("workflow_instance_id", "finding_id").
			Unique(),
		index.Fields("workflow_instance_id", "directory_account_id").
			Unique(),
		index.Fields("workflow_instance_id", "directory_group_id").
			Unique(),
		index.Fields("workflow_instance_id", "directory_membership_id").
			Unique(),
		index.Fields("workflow_instance_id", "evidence_id").
			Unique(),
		index.Fields("workflow_instance_id", "subcontrol_id").
			Unique(),
		index.Fields("workflow_instance_id", "action_plan_id").
			Unique(),
		index.Fields("workflow_instance_id", "procedure_id").
			Unique(),
		index.Fields("workflow_instance_id", "campaign_id").
			Unique(),
		index.Fields("workflow_instance_id", "campaign_target_id").
			Unique(),
		index.Fields("workflow_instance_id", "identity_holder_id").
			Unique(),
		index.Fields("workflow_instance_id", "platform_id").
			Unique(),
	}
}

// Mixin of the WorkflowObjectRef
func (w WorkflowObjectRef) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:            "WFO",
		excludeTags:       true,
		excludeSoftDelete: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.WorkflowObjectRef](w,
				withParents(WorkflowInstance{}, Control{}, InternalPolicy{}, Evidence{}, Subcontrol{}, ActionPlan{}, Procedure{}, Campaign{}, CampaignTarget{}, IdentityHolder{}, Platform{}),
				withOrganizationOwnerServiceOnly(true),
			),
		},
	}.getMixins(w)
}

// Modules this schema has access to
func (WorkflowObjectRef) Modules() []models.OrgModule {
	return []models.OrgModule{models.CatalogBaseModule}
}

// Annotations of the WorkflowObjectRef
func (WorkflowObjectRef) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.SchemaSearchable(false),
		entfga.SelfAccessChecks(),
		entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
	}
}

// Policy of the WorkflowObjectRef
func (WorkflowObjectRef) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.WorkflowObjectRefMutation](),
			entfga.CheckDeleteAccess[*generated.WorkflowObjectRefMutation](),
		),
	)
}
