package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
)

// ControlObjective defines the controlobjective schema.
type ControlObjective struct {
	ent.Schema
}

// Fields returns controlobjective fields.
func (ControlObjective) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			).
			Comment("the name of the control objective"),
		field.Text("desired_outcome").
			Optional().
			Comment("the desired outcome or target of the control objective"),
		field.String("status").
			Optional().
			Annotations(
				entgql.OrderField("status"),
			).
			Comment("status of the control objective"),
		field.Enum("source").
			GoType(enums.ControlSource("")).
			Optional().
			Annotations(
				entgql.OrderField("SOURCE"),
			).
			Default(enums.ControlSourceUserDefined.String()).
			Comment("source of the control, e.g. framework, template, custom, etc."),
		field.String("control_objective_type").
			Optional().
			Annotations(
				entgql.OrderField("control_objective_type"),
			).
			Comment("type of the control objective e.g. compliance, financial, operational, etc."),
		field.String("category").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("category"),
			).
			Comment("category of the control"),
		field.String("subcategory").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("subcategory"),
			).
			Comment("subcategory of the control"),
	}
}

// Edges of the ControlObjective
func (ControlObjective) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("programs", Program.Type).
			Annotations(entgql.RelayConnection()).
			Ref("control_objectives"),
		edge.From("evidence", Evidence.Type).
			Annotations(entgql.RelayConnection()).
			Ref("control_objectives"),

		// control objectives can map to multiple controls and subcontrols
		edge.From("controls", Control.Type).
			Annotations(entgql.RelayConnection()).
			Ref("control_objectives"),
		edge.From("subcontrols", Subcontrol.Type).
			Annotations(entgql.RelayConnection()).
			Ref("control_objectives"),

		edge.From("internal_policies", InternalPolicy.Type).
			Annotations(entgql.RelayConnection()).
			Ref("control_objectives"),

		edge.To("procedures", Procedure.Type).Annotations(entgql.RelayConnection()),

		edge.To("risks", Risk.Type).Annotations(entgql.RelayConnection()),
		edge.To("narratives", Narrative.Type).Annotations(entgql.RelayConnection()),

		edge.To("tasks", Task.Type).Annotations(entgql.RelayConnection()),
	}
}

// Mixin of the ControlObjective
func (ControlObjective) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		mixin.RevisionMixin{},
		emixin.NewIDMixinWithPrefixedID("CLO"),
		emixin.TagMixin{},
		// control objectives inherit permissions from the associated programs, but must have an organization as well
		// this mixin will add the owner_id field using the OrgHook but not organization tuples are created
		// it will also create program parent tuples for the control objective when a program is associated to the control objectives
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:               []string{"program_id", "control_id", "subcontrol_id"},
			WithOrganizationOwner:    true,
			AllowEmptyForSystemAdmin: true, // allow empty organization owner for system owned
			Ref:                      "control_objectives",
		}),
		// add groups permissions with viewer, editor, and blocked groups
		NewGroupPermissionsMixin(true),
	}
}

// Annotations of the ControlObjective
func (ControlObjective) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.MultiOrder(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.SelfAccessChecks(),
	}
}

// Policy of the ControlObjective
func (ControlObjective) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.ControlObjectiveMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ControlObjectiveMutation](),
		),
	)
}
