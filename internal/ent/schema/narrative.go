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
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("the name of the narrative"),
		field.Text("description").
			Optional().
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("the description of the narrative"),
		field.Text("details").
			Optional().
			Comment("text data for the narrative document"),
	}
}

// Edges of the Narrative
func (Narrative) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("satisfies", Control.Type).
			Comment("which controls are satisfied by the narrative").
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
		emixin.NewIDMixinWithPrefixedID("NRT"),
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
		entfga.SelfAccessChecks(),
	}
}

// Policy of the Narrative
func (Narrative) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results

		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.NarrativeMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.NarrativeMutation](),
		),
	)
}
