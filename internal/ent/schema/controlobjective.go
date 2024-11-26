package schema

import (
	"context"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
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
			Comment("the name of the control objective"),
		field.Text("description").
			Optional().
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
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("json data including details of the control objective"),
	}
}

// Edges of the ControlObjective
func (ControlObjective) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("policy", InternalPolicy.Type).
			Ref("controlobjectives"),
		edge.To("controls", Control.Type),
		edge.To("procedures", Procedure.Type),
		edge.To("risks", Risk.Type),
		edge.To("subcontrols", Subcontrol.Type),
		edge.From("standard", Standard.Type).
			Ref("controlobjectives"),
		edge.To("narratives", Narrative.Type),
		edge.To("tasks", Task.Type),
		edge.From("programs", Program.Type).
			Ref("controlobjectives"),
	}
}

// Mixin of the ControlObjective
func (ControlObjective) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		// control objectives inherit permissions from the associated programs, but must have an organization as well
		// this mixin will add the owner_id field using the OrgHook but not organization tuples are created
		// it will also create program parent tuples for the control objective when a program is associated to the control objectives
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:            []string{"program_id"},
			WithOrganizationOwner: true,
			Ref:                   "controlobjectives",
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
		entfga.Annotations{
			ObjectType: "controlobjective", // check access to the controlobjective for update/delete
		},
	}
}

// Policy of the ControlObjective
func (ControlObjective) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			rule.CanCreateObjectsInProgram(), // if mutation contains program_id, check access
			privacy.OnMutationOperation( // if there is no program_id, check access for create in org
				rule.CanCreateObjectsInOrg(),
				ent.OpCreate,
			),
			privacy.ControlObjectiveMutationRuleFunc(func(ctx context.Context, m *generated.ControlObjectiveMutation) error {
				return m.CheckAccessForEdit(ctx) // check access for edit
			}),
			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.ControlObjectiveQueryRuleFunc(func(ctx context.Context, q *generated.ControlObjectiveQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
