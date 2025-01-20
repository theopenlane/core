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
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
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
			).
			Comment("the name of the control objective"),
		field.Text("description").
			Optional().
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("description of the control objective"),
		field.String("status").
			Optional().
			Comment("status of the control objective"),
		field.String("control_objective_type").
			Optional().
			Comment("type of the control objective"),
		field.String("version").
			Optional().
			Comment("version of the control objective"),
		field.String("control_number").
			Optional().
			Comment("number of the control objective"),
		field.Text("family").
			Optional().
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("family of the control objective"),
		field.String("class").
			Optional().
			Comment("class associated with the control objective"),
		field.String("source").
			Optional().
			Comment("source of the control objective, e.g. framework, template, user-defined, etc."),
		field.Text("mapped_frameworks").
			Optional().
			Comment("mapped frameworks"),
		field.JSON("details", map[string]any{}).
			Optional().
			Comment("json data including details of the control objective"),
	}
}

// Edges of the ControlObjective
func (ControlObjective) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("internal_policies", InternalPolicy.Type).
			Ref("control_objectives"),
		edge.To("controls", Control.Type),
		edge.To("procedures", Procedure.Type),
		edge.To("risks", Risk.Type),
		edge.To("subcontrols", Subcontrol.Type),
		edge.From("standard", Standard.Type).
			Ref("control_objectives"),
		edge.To("narratives", Narrative.Type),
		edge.To("tasks", Task.Type),
		edge.From("programs", Program.Type).
			Ref("control_objectives"),
	}
}

// Mixin of the ControlObjective
func (ControlObjective) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.NewIDMixinWithPrefixedID("CLO"),
		emixin.TagMixin{},
		// control objectives inherit permissions from the associated programs, but must have an organization as well
		// this mixin will add the owner_id field using the OrgHook but not organization tuples are created
		// it will also create program parent tuples for the control objective when a program is associated to the control objectives
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:            []string{"program_id"},
			WithOrganizationOwner: true,
			Ref:                   "control_objectives",
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
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.SelfAccessChecks(),
	}
}

// Policy of the ControlObjective
func (ControlObjective) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			entfga.CheckReadAccess[*generated.ControlObjectiveQuery](),
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.ControlObjectiveMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ControlObjectiveMutation](),
		),
	)
}
