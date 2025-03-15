package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Subcontrol defines the file schema.
type Subcontrol struct {
	ent.Schema
}

// Fields returns file fields.
func (Subcontrol) Fields() []ent.Field {
	// add any fields that are specific to the subcontrol here
	additionalFields := []ent.Field{
		field.String("ref_code").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("ref_code"),
			).
			NotEmpty().
			Comment("the unique reference code for the control"),
		field.String("control_id").
			Unique().
			Comment("the id of the parent control").
			NotEmpty(),
	}

	return append(controlFields, additionalFields...)
}

// Edges of the Subcontrol
func (Subcontrol) Edges() []ent.Edge {
	return []ent.Edge{
		// subcontrols are required to have a parent control
		edge.From("control", Control.Type).
			Unique().
			Field("control_id").
			Required().
			Ref("subcontrols"),

		// controls can be mapped to other controls as a reference
		edge.From("mapped_controls", MappedControl.Type).
			Ref("subcontrols").
			Comment("mapped subcontrols that have a relation to another control or subcontrol"),

		// evidence can be associated with the control
		edge.From("evidence", Evidence.Type).
			Annotations(entgql.RelayConnection()).
			Ref("subcontrols"),

		edge.To("control_objectives", ControlObjective.Type).Annotations(entgql.RelayConnection()),

		// sub controls can have associated task, narratives, risks, and action plans
		edge.To("tasks", Task.Type).Annotations(entgql.RelayConnection()),
		edge.To("narratives", Narrative.Type).Annotations(entgql.RelayConnection()),
		edge.To("risks", Risk.Type).Annotations(entgql.RelayConnection()),
		edge.To("action_plans", ActionPlan.Type).Annotations(entgql.RelayConnection()),

		// policies and procedures are used to implement the subcontrol
		edge.To("procedures", Procedure.Type).Annotations(entgql.RelayConnection()),
		edge.To("internal_policies", InternalPolicy.Type).Annotations(entgql.RelayConnection()),

		// owner is the user who is responsible for the subcontrol
		edge.To("control_owner", Group.Type).
			Unique().
			Comment("the user who is responsible for the subcontrol, defaults to the parent control owner if not set"),
		edge.To("delegate", Group.Type).
			Unique().
			Comment("temporary delegate for the control, used for temporary control ownership"),
	}
}

// Mixin of the Subcontrol
func (Subcontrol) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.NewIDMixinWithPrefixedID("SCL"),
		emixin.TagMixin{},
		// subcontrols can inherit permissions from the parent control
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:               []string{"control_id"},
			WithOrganizationOwner:    true,
			AllowEmptyForSystemAdmin: true, // allow organization owner to be empty
			Ref:                      "subcontrols",
		}),
	}
}

func (Subcontrol) Indexes() []ent.Index {
	return []ent.Index{
		// ref_code should be unique within the parent control
		index.Fields("control_id", "ref_code").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL"),
		),
	}
}

// Annotations of the Subcontrol
func (Subcontrol) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.MultiOrder(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the Subcontrol
func (Subcontrol) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookSubcontrolUpdate(),
	}
}

// Policy of the Subcontrol
func (Subcontrol) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.SubcontrolMutation](rule.ControlParent), // if mutation contains control_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.SubcontrolMutation](),
		),
	)
}
