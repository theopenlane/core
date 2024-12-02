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

// Control defines the control schema.
type Control struct {
	ent.Schema
}

// Fields returns control fields.
func (Control) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("the name of the control"),
		field.Text("description").
			Optional().
			Comment("description of the control"),
		field.String("status").
			Optional().
			Comment("status of the control"),
		field.String("control_type").
			Optional().
			Comment("type of the control"),
		field.String("version").
			Optional().
			Comment("version of the control"),
		field.String("control_number").
			Optional().
			Comment("control number or identifier"),
		field.Text("family").
			Optional().
			Comment("family associated with the control"),
		field.String("class").
			Optional().
			Comment("class associated with the control"),
		field.String("source").
			Optional().
			Comment("source of the control, e.g. framework, template, custom, etc."),
		field.Text("satisfies").
			Optional().
			Comment("which control objectives are satisfied by the control"),
		field.Text("mapped_frameworks").
			Optional().
			Comment("mapped frameworks"),
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("json data including details of the control"),
	}
}

// Edges of the Control
func (Control) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("procedures", Procedure.Type),
		edge.To("subcontrols", Subcontrol.Type),
		edge.To("controlobjectives", ControlObjective.Type),
		edge.From("standard", Standard.Type).
			Ref("controls"),
		edge.To("narratives", Narrative.Type),
		edge.To("risks", Risk.Type),
		edge.To("actionplans", ActionPlan.Type),
		edge.To("tasks", Task.Type),
		edge.From("programs", Program.Type).
			Ref("controls"),
	}
}

// Mixin of the Control
func (Control) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		// controls must be associated with an organization but do not inherit permissions from the organization
		// controls can inherit permissions from the associated programs
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:            []string{"program_id"},
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
		entfga.Annotations{
			ObjectType: "control", // check access to the control for update/delete
		},
	}
}

// Policy of the Control
func (Control) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			rule.CanCreateObjectsInProgram(), // if mutation contains program_id, check access
			privacy.OnMutationOperation( // if there is no program_id, check access for create in org
				rule.CanCreateObjectsInOrg(), // TODO (sfunk): update to allow create access based on groups
				ent.OpCreate,
			),
			privacy.ControlMutationRuleFunc(func(ctx context.Context, m *generated.ControlMutation) error {
				return m.CheckAccessForEdit(ctx) // check access for edit
			}),
			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.ControlQueryRuleFunc(func(ctx context.Context, q *generated.ControlQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
