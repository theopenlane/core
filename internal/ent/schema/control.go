package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Control defines the control schema.
type Control struct {
	ent.Schema
}

// Fields returns control fields.
func (Control) Fields() []ent.Field {
	// add any fields that are specific to the parent control here
	additionalFields := []ent.Field{
		field.String("standard_id").
			Comment("the id of the standard that the control belongs to, if applicable").
			Optional(),
	}

	return append(controlFields, additionalFields...)
}

// Edges of the Control
func (Control) Edges() []ent.Edge {
	return []ent.Edge{
		// parents of the control (standard, program)
		edge.From("standard", Standard.Type).
			Field("standard_id").
			Unique().
			Ref("controls"),
		edge.From("programs", Program.Type).
			Ref("controls"),

		// evidence can be associated with the control
		edge.From("evidence", Evidence.Type).
			Ref("controls"),

		edge.To("implementation", ControlImplementation.Type).
			Unique().
			Comment("the implementation of the control"),

		edge.To("mapped_controls", MappedControl.Type).
			Unique().
			Comment("controls that are mapped to this control"),

		// controls have control objectives and subcontrols
		edge.To("control_objectives", ControlObjective.Type),
		edge.To("subcontrols", Subcontrol.Type),

		// controls can have associated task, narratives, risks, and action plans
		edge.To("tasks", Task.Type),
		edge.To("narratives", Narrative.Type),
		edge.To("risks", Risk.Type),
		edge.To("action_plans", ActionPlan.Type),

		// policies and procedures are used to implement the control
		edge.To("procedures", Procedure.Type),
		edge.To("internal_policies", InternalPolicy.Type),

		// owner is the user who is responsible for the control
		edge.To("control_owner", User.Type).
			Unique().
			Comment("the user who is responsible for the control"),
		edge.To("delegate", User.Type).
			Unique().
			Comment("temporary delegate for the control, used for temporary control ownership"),
	}
}

func (Control) Indexes() []ent.Index {
	return []ent.Index{
		// ref_code should be unique within the standard
		index.Fields("standard_id", "ref_code").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL"),
		),
	}
}

// Mixin of the Control
func (Control) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.NewIDMixinWithPrefixedID("CTL"),
		emixin.TagMixin{},
		// controls must be associated with an organization but do not inherit permissions from the organization
		// controls can inherit permissions from the associated programs
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:            []string{"program_id", "standard_id"},
			WithOrganizationOwner: true,
			Ref:                   "controls",
		}),
		// add groups permissions with viewer, editor, and blocked groups
		NewGroupPermissionsMixin(true),
	}
}

// Annotations of the Control
func (Control) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.SelfAccessChecks(),
	}
}

// Policy of the Control
func (Control) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.ControlMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ControlMutation](),
		),
	)
}
