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

// Narrative defines the narrative schema
type Narrative struct {
	ent.Schema
}

// Fields returns narrative fields
func (Narrative) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("the name of the narrative"),
		field.Text("description").
			Optional().
			Comment("the description of the narrative"),
		field.Text("satisfies").
			Optional().
			Comment("which controls are satisfied by the narrative"),
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("json data for the narrative document"),
	}
}

// Edges of the Narrative
func (Narrative) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("policy", InternalPolicy.Type).
			Ref("narratives"),
		edge.From("control", Control.Type).
			Ref("narratives"),
		edge.From("procedure", Procedure.Type).
			Ref("narratives"),
		edge.From("controlobjective", ControlObjective.Type).
			Ref("narratives"),
		edge.From("programs", Program.Type).
			Ref("narratives"),
	}
}

// Mixin of the Narrative
func (Narrative) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		// narratives inherit permissions from the associated programs, but must have an organization as well
		// this mixin will add the owner_id field using the OrgHook but not organization tuples are created
		// it will also create program parent tuples for the narrative when a program is associated to the narrative
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:            []string{"program_id"},
			WithOrganizationOwner: true,
			Ref:                   "narratives",
		}),
		// add groups permissions with viewer, editor, and blocked groups
		NewGroupPermissionsMixin(true),
	}
}

// Annotations of the Narrative
func (Narrative) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.Annotations{
			ObjectType: "narrative", // check access to the narrative for update/delete
		},
	}
}

// Policy of the Narrative
func (Narrative) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			rule.CanCreateObjectsInProgram(), // if mutation contains program_id, check access
			privacy.OnMutationOperation( // if there is no program_id, check access for create in org
				rule.CanCreateObjectsInOrg(),
				ent.OpCreate,
			),
			privacy.NarrativeMutationRuleFunc(func(ctx context.Context, m *generated.NarrativeMutation) error {
				return m.CheckAccessForEdit(ctx) // check access for edit
			}),
			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.NarrativeQueryRuleFunc(func(ctx context.Context, q *generated.NarrativeQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
